<!-- 打印任务：查看自动生成的单据、预览正文与重打 -->
<template>
  <div class="page">
    <div class="toolbar">
      <ElSelect v-model="status" placeholder="全部状态" clearable class="w-40" @change="reload">
        <ElOption label="待打印" :value="0" />
        <ElOption label="已打印" :value="1" />
        <ElOption label="打印失败" :value="2" />
        <ElOption label="已取消" :value="3" />
      </ElSelect>
      <ElButton :loading="loading" @click="loadData">刷新</ElButton>
      <span class="total">共 {{ total }} 条</span>
    </div>

    <ElTable v-loading="loading" :data="list" border size="small">
      <ElTableColumn prop="orderNo" label="订单号" min-width="180" show-overflow-tooltip />
      <ElTableColumn prop="printTypeName" label="类型" width="90" />
      <ElTableColumn prop="printerName" label="设备" width="130" show-overflow-tooltip />
      <ElTableColumn label="状态" width="100">
        <template #default="{ row }">
          <ElTag :type="statusTag(row.status)" effect="plain">{{ row.statusName }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="retryCount" label="重打" width="70" />
      <ElTableColumn prop="errorMsg" label="失败原因" min-width="140" show-overflow-tooltip />
      <ElTableColumn prop="createdAt" label="生成时间" width="170" />
      <ElTableColumn label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <ElButton link type="primary" @click="preview(row)">预览</ElButton>
          <ElButton
            v-if="row.status !== 0 && auth.hasAuth('store:print-task:reprint')"
            link
            type="warning"
            @click="reprint(row)"
          >
            重打
          </ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElPagination
      v-model:current-page="pageNo"
      class="pager"
      layout="prev, pager, next"
      :page-size="pageSize"
      :total="total"
      @current-change="loadData"
    />

    <ElDrawer v-model="previewVisible" title="打印内容" size="420px">
      <pre class="preview">{{ current?.content || '（无内容）' }}</pre>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onMounted, ref } from 'vue'

  import { fetchPrintTaskPage, fetchReprintPrintTask } from '@/api/store'
  import { useAuthStore } from '@/stores/auth'

  const auth = useAuthStore()

  const list = ref<Api.Store.PrintTaskItem[]>([])
  const loading = ref(false)
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = 20
  const status = ref<number | undefined>(undefined)

  const previewVisible = ref(false)
  const current = ref<Api.Store.PrintTaskItem | null>(null)

  const statusTag = (value: number): 'warning' | 'success' | 'danger' | 'info' => {
    if (value === 0) return 'warning'
    if (value === 1) return 'success'
    if (value === 2) return 'danger'

    return 'info'
  }

  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      const params: Record<string, unknown> = { pageNo: pageNo.value, pageSize }
      // 状态 0（待打印）是合法筛选值，只有真正清空才表示不限
      if (typeof status.value === 'number') {
        params.status = status.value
      }
      const res = await fetchPrintTaskPage(params)
      list.value = res.list ?? []
      total.value = res.total ?? 0
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  const reload = (): void => {
    pageNo.value = 1
    void loadData()
  }

  const preview = (row: Api.Store.PrintTaskItem): void => {
    current.value = row
    previewVisible.value = true
  }

  const reprint = async (row: Api.Store.PrintTaskItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定重打「${row.orderNo}」的${row.printTypeName}吗？`, '重打', {
        confirmButtonText: '确认重打',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchReprintPrintTask(row.id)
      ElMessage.success('已加入待打印队列')
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .total {
    margin-left: auto;
    font-size: 13px;
    color: #6b7280;
  }

  .pager {
    justify-content: center;
    margin-top: 16px;
  }

  .preview {
    margin: 0;
    padding: 12px;
    font-family: 'Microsoft YaHei', monospace;
    font-size: 13px;
    line-height: 1.8;
    white-space: pre-wrap;
    background: #f5f6f8;
    border-radius: 8px;
  }

  .w-40 {
    width: 160px;
  }
</style>
