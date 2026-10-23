<!-- 桌台图：按区域分组看台位，空台直接进点单，用完清台 -->
<template>
  <div class="page">
    <div class="toolbar">
      <ElRadioGroup v-model="areaId" @change="loadData">
        <ElRadioButton :value="''">全部</ElRadioButton>
        <ElRadioButton v-for="area in areas" :key="area.id" :value="area.id">
          {{ area.name }}
        </ElRadioButton>
      </ElRadioGroup>

      <ElButton :loading="loading" @click="loadData">刷新</ElButton>
      <ElSwitch v-model="autoRefresh" active-text="每 10 秒自动刷新" />

      <div class="legend">
        <span class="dot idle" />空闲
        <span class="dot dining" />使用中
        <span class="dot pending" />待清台
      </div>
    </div>

    <ElEmpty v-if="!loading && !tables.length" description="该门店还没有配置桌台" />

    <ElDialog v-model="mergeVisible" title="并台" width="400px">
      <div class="transfer-from">
        保留：{{ mergeFrom?.name }}（{{ mergeFrom?.currentOrderNo }}）
      </div>
      <ElSelect v-model="mergeTo" placeholder="选择要并入的桌台" filterable class="w-full">
        <ElOption
          v-for="item in mergeCandidates"
          :key="item.id"
          :label="`${item.name}（${item.currentOrderNo}）`"
          :value="item.id"
        />
      </ElSelect>
      <div class="transfer-tip">
        合并后对方桌台的菜会并入当前台，对方订单作废并自动清台，不可撤销
      </div>

      <template #footer>
        <ElButton @click="mergeVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="merging" @click="confirmMerge">确认并台</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="transferVisible" title="换台" width="380px">
      <div class="transfer-from">当前台位：{{ transferFrom?.name }}</div>
      <ElSelect v-model="transferTo" placeholder="选择目标空台" filterable class="w-full">
        <ElOption
          v-for="item in idleTables"
          :key="item.id"
          :label="`${item.areaName || ''} ${item.name}（${item.seats}座）`"
          :value="item.id"
        />
      </ElSelect>
      <div class="transfer-tip">只列出当前筛选下已加载的空闲桌台</div>

      <template #footer>
        <ElButton @click="transferVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="transferring" @click="confirmTransfer">确认换台</ElButton>
      </template>
    </ElDialog>

    <div class="table-grid">
      <div
        v-for="table in tables"
        :key="table.id"
        class="table-card"
        :class="statusClass(table.status)"
      >
        <div class="table-head">
          <span class="table-name">{{ table.name }}</span>
          <span class="table-seats">{{ table.seats }} 座</span>
        </div>
        <div class="table-area">{{ table.areaName || '未分区' }}</div>
        <div v-if="table.currentOrderNo" class="table-order">
          {{ table.currentOrderNo }} · ¥{{ yuan(table.currentOrderAmount ?? 0) }}
        </div>
        <div v-else class="table-order">{{ statusText(table.status) }}</div>
        <div v-if="(table.openedMinutes ?? 0) > 0" class="table-duration">
          已开 {{ table.openedMinutes }} 分钟
        </div>

        <div class="table-actions">
          <ElButton v-if="table.status === 1" type="primary" size="small" @click="toPos(table)">
            {{ canOperate ? '点单开台' : '去点单' }}
          </ElButton>
          <template v-else>
            <ElButton size="small" :disabled="!table.currentOrderId" @click="toPos(table)">
              加菜
            </ElButton>
            <ElButton
              v-if="canOperate && table.status === 2 && table.currentOrderId"
              size="small"
              @click="openMerge(table)"
            >
              并台
            </ElButton>
            <ElButton
              v-if="canOperate && table.status === 2"
              size="small"
              @click="openTransfer(table)"
            >
              换台
            </ElButton>
            <ElButton v-if="canOperate" type="danger" plain size="small" @click="clearTable(table)">
              清台
            </ElButton>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'

  import { fetchMergeOrders } from '@/api/order'
  import { fetchClearTable, fetchTablePage, fetchTransferTable } from '@/api/store'
  import { useAuthStore } from '@/stores/auth'

  const router = useRouter()
  const auth = useAuthStore()

  const loading = ref(false)
  const tables = ref<Api.Store.TableItem[]>([])
  const areaId = ref('')
  const autoRefresh = ref(true)
  let timer: ReturnType<typeof setInterval> | null = null

  // 清台会改桌台状态，需要单独的操作权限
  const canOperate = computed(() => auth.hasAuth('store:table:operate'))

  // 区域从桌台数据里聚合，少调一个接口，也保证和展示口径一致
  const areas = computed(() => {
    const map = new Map<string, string>()
    for (const table of tables.value) {
      if (table.areaId && table.areaName) {
        map.set(table.areaId, table.areaName)
      }
    }

    return [...map.entries()].map(([id, name]) => ({ id, name }))
  })

  const yuan = (cent: number): string => ((cent || 0) / 100).toFixed(2)

  const statusText = (status: number): string =>
    status === 1 ? '空闲' : status === 2 ? '使用中' : '待清台'

  const statusClass = (status: number): string =>
    status === 1 ? 'idle' : status === 2 ? 'dining' : 'pending'

  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      const params: Record<string, unknown> = { pageNo: 1, pageSize: 200 }
      if (areaId.value) {
        params.areaId = areaId.value
      }
      const res = await fetchTablePage(params)
      tables.value = res.list ?? []
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  // 跳点单页：带桌台；已有台位单时再带订单，POS 进入加菜模式
  const toPos = (table: Api.Store.TableItem): void => {
    const query: Record<string, string> = { tableId: table.id }
    if (table.currentOrderId) {
      query.orderId = table.currentOrderId
      if (table.currentOrderNo) {
        query.orderNo = table.currentOrderNo
      }
    }
    router.push({ path: '/pos', query })
  }

  // 并台：把对方台的订单并入当前台
  const mergeVisible = ref(false)
  const merging = ref(false)
  const mergeFrom = ref<Api.Store.TableItem | null>(null)
  const mergeTo = ref('')

  const mergeCandidates = computed(() =>
    tables.value.filter(
      (item) =>
        item.status === 2 && !!item.currentOrderId && item.id !== mergeFrom.value?.id
    )
  )

  const openMerge = (table: Api.Store.TableItem): void => {
    mergeFrom.value = table
    mergeTo.value = ''
    mergeVisible.value = true
  }

  const confirmMerge = async (): Promise<void> => {
    const from = mergeFrom.value
    const target = mergeCandidates.value.find((item) => item.id === mergeTo.value)
    if (!from || !from.currentOrderId || !target || !target.currentOrderId) {
      ElMessage.warning('请选择要并入的桌台')

      return
    }

    merging.value = true
    try {
      const res = await fetchMergeOrders(from.currentOrderId, target.currentOrderId)
      // 对方台已无单，顺手清台，避免留个「使用中但没单」的僵尸台
      if (res.sourceTableId) {
        try {
          await fetchClearTable(res.sourceTableId)
        } catch {
          // 清台失败不影响并台结果，店长可手动清
        }
      }
      ElMessage.success(`并台成功，${from.name} 应收 ¥${yuan(res.payAmount)}`)
      mergeVisible.value = false
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      merging.value = false
    }
  }

  const clearTable = async (table: Api.Store.TableItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定清台「${table.name}」吗？清台后该台恢复空闲，台位单关联会被解除。`,
        '清台',
        { confirmButtonText: '确认清台', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchClearTable(table.id)
      ElMessage.success('已清台')
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    }
  }

  // 换台弹窗：源台与目标台都在同一门店，目标只列空闲台
  const transferVisible = ref(false)
  const transferring = ref(false)
  const transferFrom = ref<Api.Store.TableItem | null>(null)
  const transferTo = ref('')

  const idleTables = computed(() =>
    tables.value.filter((item) => item.status === 1 && item.id !== transferFrom.value?.id)
  )

  const openTransfer = (table: Api.Store.TableItem): void => {
    transferFrom.value = table
    transferTo.value = ''
    transferVisible.value = true
  }

  const confirmTransfer = async (): Promise<void> => {
    const from = transferFrom.value
    if (!from || !transferTo.value) {
      ElMessage.warning('请选择目标桌台')

      return
    }

    transferring.value = true
    try {
      await fetchTransferTable(from.id, transferTo.value)
      ElMessage.success('换台成功')
      transferVisible.value = false
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      transferring.value = false
    }
  }

  const stopTimer = (): void => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  watch(autoRefresh, (on) => {
    stopTimer()
    if (on) {
      timer = setInterval(() => void loadData(), 10000)
    }
  })

  onMounted(() => {
    void loadData()
    if (autoRefresh.value) {
      timer = setInterval(() => void loadData(), 10000)
    }
  })

  onBeforeUnmount(stopTimer)
</script>

<style scoped>
  .table-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 12px;
  }

  .table-card {
    padding: 14px;
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
    border-left-width: 5px;
  }

  .table-card.idle {
    border-left-color: #67c23a;
  }

  .table-card.dining {
    border-left-color: #e6a23c;
    background: #fdf6ec;
  }

  .table-card.pending {
    border-left-color: #909399;
    background: #f4f4f5;
  }

  .table-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .table-name {
    font-size: 18px;
    font-weight: 600;
  }

  .table-seats {
    font-size: 12px;
    color: #6b7280;
  }

  .table-area {
    margin-top: 4px;
    font-size: 12px;
    color: #909399;
  }

  .table-order {
    margin-top: 8px;
    font-size: 13px;
    color: #606266;
  }

  .table-duration {
    margin-top: 2px;
    font-size: 12px;
    color: #e6a23c;
  }

  .table-actions {
    display: flex;
    gap: 6px;
    margin-top: 12px;
  }

  .transfer-from {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 500;
  }

  .transfer-tip {
    margin-top: 8px;
    font-size: 12px;
    color: #909399;
  }

  .w-full {
    width: 100%;
  }

  .legend {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-left: auto;
    font-size: 12px;
    color: #6b7280;
  }

  .dot {
    width: 8px;
    height: 8px;
    margin-left: 8px;
    border-radius: 50%;
  }

  .dot.idle {
    background: #67c23a;
  }

  .dot.dining {
    background: #e6a23c;
  }

  .dot.pending {
    background: #909399;
  }
</style>
