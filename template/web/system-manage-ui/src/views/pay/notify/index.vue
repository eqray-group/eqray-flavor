<!-- 业务回调通知页面：查看支付 / 退款结果推送给业务方的任务状态与派发日志 -->
<template>
  <div class="pay-notify-page art-full-height">
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

      <NotifyLogDrawer v-model:visible="logVisible" :task="currentTask" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayAppOptions } from '@/api/pay/app'
  import { fetchGetPayNotifyTaskList } from '@/api/pay/notify'
  import { ElTag } from 'element-plus'
  import NotifyLogDrawer from './modules/notify-log-drawer.vue'

  defineOptions({ name: 'PayNotify' })

  type TaskItem = Api.Pay.NotifyTaskItem

  // 1 支付 / 2 退款 / 3 转账
  const typeOptions = [
    { label: '支付单', value: 1 },
    { label: '退款单', value: 2 },
    { label: '转账单', value: 3 }
  ]

  // 0 等待通知 / 10 通知成功 / 20 通知失败
  // 取值对齐 yudao：1 等待 / 2 成功 / 3 失败 / 4 请求成功但结果失败 / 5 请求失败
  const statusOptions = [
    { label: '等待通知', value: 1 },
    { label: '通知成功', value: 2 },
    { label: '通知失败', value: 3 },
    { label: '请求成功但结果失败', value: 4 },
    { label: '请求失败', value: 5 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'danger'> = {
    1: 'warning',
    2: 'success',
    3: 'danger',
    4: 'warning',
    5: 'danger'
  }

  const appOptions = ref<Api.Pay.AppOption[]>([])

  const searchForm = ref<Api.Pay.NotifyTaskSearchParams>({
    keyword: '',
    appId: '',
    type: undefined,
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '商户订单号',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '业务侧订单号',
        clearable: true
      }
    },
    {
      label: '所属应用',
      key: 'appId',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: appOptions.value.map((item) => ({ label: item.name, value: item.id }))
      }
    },
    {
      label: '类型',
      key: 'type',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: typeOptions
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    }
  ])

  const logVisible = ref(false)
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
      apiFn: fetchGetPayNotifyTaskList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'merchantOrderId',
          label: '商户订单号',
          minWidth: 170,
          showOverflowTooltip: true
        },
        { prop: 'typeName', label: '类型', width: 100 },
        {
          prop: 'notifyUrl',
          label: '回调地址',
          minWidth: 220,
          showOverflowTooltip: true
        },
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
        {
          prop: 'notifyTimes',
          label: '通知次数',
          width: 120,
          formatter: (row: TaskItem) => `${row.notifyTimes} / ${row.maxNotifyTimes}`
        },
        { prop: 'lastExecuteTime', label: '最后执行时间', width: 170 },
        { prop: 'nextNotifyTime', label: '下次通知时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 110,
          fixed: 'right',
          formatter: (row: TaskItem) =>
            h(ArtButtonTable, {
              type: 'view',
              text: '派发日志',
              onClick: () => openLogs(row)
            })
        }
      ]
    }
  })

  // 加载应用下拉，供筛选使用
  const loadApps = async (): Promise<void> => {
    try {
      appOptions.value = (await fetchGetPayAppOptions()) ?? []
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(() => {
    loadApps()
  })

  // 搜索条件变化后重新加载通知任务列表
  const handleSearch = (params: Api.Pay.NotifyTaskSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开派发日志抽屉
  const openLogs = (row: TaskItem): void => {
    currentTask.value = row
    logVisible.value = true
  }
</script>
