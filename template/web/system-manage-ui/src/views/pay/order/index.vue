<!-- 支付单管理页面：查询支付单、代下单并调起渠道、查看渠道调用记录与关单 -->
<template>
  <div class="pay-order-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'pay:order:create'" @click="showCreateDialog" v-ripple>
              创建支付单
            </ElButton>
            <ElButton v-auth="'pay:order:list'" :loading="exporting" @click="exportOrders">
              导出 CSV
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <OrderCreateDialog
        v-model:visible="createVisible"
        :app-options="appOptions"
        @success="handleCreateSuccess"
      />

      <OrderDetailDialog v-model:visible="detailVisible" :order-id="currentOrderId" />

      <OrderSubmitDialog
        v-model:visible="submitVisible"
        :order="currentOrder"
        @success="handleSubmitSuccess"
      />

      <RefundCreateDialog
        v-model:visible="refundVisible"
        :order="currentRefundOrder"
        @success="handleRefundSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetPayAppOptions } from '@/api/pay/app'
  import { fetchGetPayChannelOptions } from '@/api/pay/channel'
  import {
    fetchClosePayOrder,
    fetchExportPayOrders,
    fetchGetPayOrderList,
    fetchSyncPayOrder
  } from '@/api/pay/order'
  import { exportFilename, saveBlob } from '@/utils/download'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import OrderCreateDialog from './modules/order-create-dialog.vue'
  import OrderDetailDialog from './modules/order-detail-dialog.vue'
  import OrderSubmitDialog from './modules/order-submit-dialog.vue'
  import RefundCreateDialog from './modules/refund-create-dialog.vue'

  defineOptions({ name: 'PayOrder' })

  type OrderItem = Api.Pay.OrderItem

  const { hasAuth } = useAuth()

  // 取值对齐 yudao：0 未支付 / 10 支付成功 / 20 支付关闭
  // 全额退款不单列状态，看 refundPrice 是否等于 price
  const statusOptions = [
    { label: '待支付', value: 0 },
    { label: '支付成功', value: 10 },
    { label: '支付关闭', value: 20 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'info' | 'danger'> = {
    0: 'warning',
    10: 'success',
    20: 'info'
  }

  const appOptions = ref<Api.Pay.AppOption[]>([])
  const channelOptions = ref<Api.Pay.ChannelOption[]>([])

  const searchForm = ref<Api.Pay.OrderSearchParams>({
    keyword: '',
    appId: '',
    channelCode: '',
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '支付单号 / 业务订单号 / 渠道订单号',
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
      label: '渠道',
      key: 'channelCode',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: channelOptions.value.map((item) => ({ label: item.name, value: item.code }))
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

  const createVisible = ref(false)
  const exporting = ref(false)
  const detailVisible = ref(false)
  const submitVisible = ref(false)
  const refundVisible = ref(false)
  const currentOrderId = ref('')
  const currentOrder = ref<OrderItem | null>(null)
  const currentRefundOrder = ref<OrderItem | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

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
    refreshData,
    refreshCreate
  } = useTable({
    core: {
      apiFn: fetchGetPayOrderList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'no',
          label: '支付单号',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'merchantOrderId',
          label: '业务订单号',
          minWidth: 170,
          showOverflowTooltip: true
        },
        {
          prop: 'channelName',
          label: '渠道',
          width: 130
        },
        {
          prop: 'price',
          label: '支付金额（元）',
          width: 130,
          formatter: (row: OrderItem) => formatYuan(row.price)
        },
        {
          prop: 'refundPrice',
          label: '已退（元）',
          width: 120,
          formatter: (row: OrderItem) => formatYuan(row.refundPrice)
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: OrderItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || '未知'
            )
        },
        {
          prop: 'channelOrderNo',
          label: '渠道订单号',
          minWidth: 200,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 260,
          fixed: 'right',
          formatter: (row: OrderItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '详情',
                onClick: () => openDetail(row)
              }),
              // 已支付的单可以发起退款（部分退款由弹窗内金额控制）
              ...(row.status === 10 && hasAuth('pay:refund:create')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '退款',
                      onClick: () => openRefund(row)
                    })
                  ]
                : []),
              // 只有待支付的单可以重新调起或关闭
              ...(row.status === 0 && hasAuth('pay:order:submit')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '调起',
                      onClick: () => openSubmit(row)
                    })
                  ]
                : []),
              // 补单：渠道已收款但回调丢失时，主动查单把状态补成已支付
              ...(row.status === 0 && hasAuth('pay:order:sync')
                ? [
                    h(ArtButtonTable, {
                      type: 'view',
                      text: '同步',
                      onClick: () => syncOrder(row)
                    })
                  ]
                : []),
              ...(row.status === 0 && hasAuth('pay:order:close')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '关闭',
                      onClick: () => closeOrder(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 加载应用与渠道编码下拉，供筛选与代下单弹窗使用
  const loadOptions = async (): Promise<void> => {
    try {
      const [apps, codes] = await Promise.all([
        fetchGetPayAppOptions(),
        fetchGetPayChannelOptions()
      ])
      appOptions.value = apps ?? []
      channelOptions.value = codes ?? []
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(() => {
    loadOptions()
  })

  // 搜索条件变化后重新加载支付单列表
  const handleSearch = (params: Api.Pay.OrderSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 导出当前筛选条件下的支付单，行数上限由后端截断
  const exportOrders = async (): Promise<void> => {
    exporting.value = true
    try {
      const blob = await fetchExportPayOrders({ ...searchParams })
      saveBlob(blob, exportFilename('pay-orders'))
      ElMessage.success('导出成功')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      exporting.value = false
    }
  }

  // 打开创建支付单弹窗
  const showCreateDialog = (): void => {
    createVisible.value = true
  }

  // 下单成功后刷新列表：新建的支付单通常排在第一页首位
  const handleCreateSuccess = (): void => {
    refreshCreate()
  }

  // 打开支付单详情弹窗，按主键拉取详情与渠道调用记录
  const openDetail = (row: OrderItem): void => {
    currentOrderId.value = row.id
    detailVisible.value = true
  }

  // 打开重新调起弹窗：待支付的单补 openid / 付款码后可再调一次渠道
  const openSubmit = (row: OrderItem): void => {
    currentOrder.value = row
    submitVisible.value = true
  }

  // 重新调起成功后刷新列表，让渠道调用结果与状态同步到最新
  const handleSubmitSuccess = (): void => {
    refreshData()
  }

  // 打开退款弹窗：只有已支付的单可以退款
  const openRefund = (row: OrderItem): void => {
    currentRefundOrder.value = row
    refundVisible.value = true
  }

  // 退款提交成功后刷新列表，同步已退金额与状态
  const handleRefundSuccess = (): void => {
    refreshData()
  }

  // 主动查单补单：渠道已收款时把本地状态补成已支付，未收款则提示仍待支付
  const syncOrder = async (row: OrderItem): Promise<void> => {
    try {
      const latest = await fetchSyncPayOrder(row.id)
      if (latest?.status === 10) {
        ElMessage.success('渠道已收款，支付单已同步为已支付')
      } else {
        ElMessage.info('渠道仍未收款，支付单保持待支付')
      }
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 关闭待支付的支付单，确认后调用接口并刷新列表
  const closeOrder = async (row: OrderItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要关闭支付单「${row.no}」吗？关闭后渠道将无法再收款。`,
        '关闭支付单',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }

    try {
      await fetchClosePayOrder(row.id)
      ElMessage.success('已关闭')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
