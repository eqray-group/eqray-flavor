<!-- 打印任务页面：查看支付成功后自动生成的单据，正文预览与重打 -->
<template>
  <div class="store-print-task-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <ElDrawer v-model="previewVisible" title="打印内容预览" size="520px">
        <div class="mb-3 text-sm text-g-600">
          任务 {{ currentTask?.orderNo || '-' }} · {{ currentTask?.printerName || '-' }} ·
          {{ currentTask?.statusName || '' }}
        </div>
        <ElAlert
          v-if="currentTask?.errorMsg"
          type="error"
          :closable="false"
          show-icon
          :title="`失败原因：${currentTask.errorMsg}`"
        />
        <pre class="print-preview">{{ currentTask?.content || '（无内容）' }}</pre>
      </ElDrawer>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchGetPrintTaskDetail,
    fetchGetPrintTaskList,
    fetchReprintPrintTask
  } from '@/api/store/print-task'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'StorePrintTask' })

  type TaskItem = Api.Store.PrintTaskItem

  // 打印类型取值对齐 store 模块 enums.PrintTypeEnum
  const PRINT_TYPE_OPTIONS = [
    { label: '结账单', value: 1 },
    { label: '预结单', value: 2 },
    { label: '厨打单', value: 3 },
    { label: '退菜单', value: 4 },
    { label: '标签', value: 5 },
    { label: '退款单', value: 6 },
    { label: '交班单', value: 7 },
    { label: '日结单', value: 8 }
  ]

  // 任务状态：0 待打印 / 1 已打印 / 2 打印失败 / 3 已取消
  const STATUS_OPTIONS = [
    { label: '待打印', value: 0 },
    { label: '已打印', value: 1 },
    { label: '打印失败', value: 2 },
    { label: '已取消', value: 3 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'danger' | 'info'> = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    3: 'info'
  }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Store.PrintTaskSearchParams>({
    orderNo: '',
    printType: undefined,
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '订单号',
      key: 'orderNo',
      type: 'input',
      span: 8,
      props: { placeholder: '输入订单号搜索', clearable: true }
    },
    {
      label: '打印类型',
      key: 'printType',
      type: 'select',
      span: 6,
      props: { placeholder: '全部', clearable: true, options: PRINT_TYPE_OPTIONS }
    },
    {
      label: '任务状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: { placeholder: '全部', clearable: true, options: STATUS_OPTIONS }
    }
  ])

  const previewVisible = ref(false)
  const currentTask = ref<TaskItem | null>(null)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetPrintTaskList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'storeName', label: '门店', width: 140, showOverflowTooltip: true },
        { prop: 'orderNo', label: '订单号', minWidth: 190, showOverflowTooltip: true },
        { prop: 'printTypeName', label: '打印类型', width: 110 },
        { prop: 'printerName', label: '打印设备', width: 140, showOverflowTooltip: true },
        { prop: 'copies', label: '份数', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 110,
          formatter: (row: TaskItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || '未知'
            )
        },
        { prop: 'retryCount', label: '重打次数', width: 100 },
        {
          prop: 'errorMsg',
          label: '失败原因',
          minWidth: 160,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '生成时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 180,
          fixed: 'right',
          formatter: (row: TaskItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '预览',
                onClick: () => openPreview(row)
              }),
              // 待打印的还没出纸，不需要重打
              ...(row.status !== 0 && hasAuth('store:print-task:reprint')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '重打',
                      onClick: () => reprint(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载
  const handleSearch = (params: Api.Store.PrintTaskSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 预览正文：重新拉详情，避免列表里不带大字段
  const openPreview = async (row: TaskItem): Promise<void> => {
    currentTask.value = row
    previewVisible.value = true
    try {
      currentTask.value = await fetchGetPrintTaskDetail(row.id)
    } catch {
      // 错误提示已由请求层统一弹出，保留列表里的摘要信息
    }
  }

  // 重打：任务重置为待打印，设备下次取任务时会重新出纸
  const reprint = async (row: TaskItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定重打「${row.orderNo}」的${row.printTypeName}吗？任务会重新进入待打印队列。`,
        '重打小票',
        { confirmButtonText: '确认重打', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchReprintPrintTask(row.id)
      ElMessage.success('已加入待打印队列')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>

<style scoped>
  .print-preview {
    margin: 0;
    padding: 16px;
    border: 1px dashed var(--el-border-color);
    border-radius: 8px;
    font-family: 'Microsoft YaHei', monospace;
    font-size: 13px;
    line-height: 1.8;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--el-fill-color-light);
  }
</style>
