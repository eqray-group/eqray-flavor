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
        <div class="table-order">
          {{ table.currentOrderId ? '有台位单' : statusText(table.status) }}
        </div>

        <div class="table-actions">
          <ElButton v-if="table.status === 1" type="primary" size="small" @click="toPos(table)">
            {{ canOperate ? '点单开台' : '去点单' }}
          </ElButton>
          <template v-else>
            <ElButton size="small" @click="toPos(table)">加菜</ElButton>
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

  import { fetchClearTable, fetchTablePage } from '@/api/store'
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

  // 跳点单页并带上桌台，POS 会预选堂食与该台
  const toPos = (table: Api.Store.TableItem): void => {
    router.push({ path: '/pos', query: { tableId: table.id } })
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

  .table-actions {
    display: flex;
    gap: 6px;
    margin-top: 12px;
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
