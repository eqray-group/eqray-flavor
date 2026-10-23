<!-- 订单管理页面：查询订单、查看明细与加料、执行状态流转与取消 -->
<template>
  <div class="order-page art-full-height">
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

      <OrderDetailDialog v-model:visible="detailVisible" :order-id="currentOrderId" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchCancelOrder, fetchGetOrderList, fetchTransitOrder } from '@/api/order/order'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import OrderDetailDialog from './modules/order-detail-dialog.vue'

  defineOptions({ name: 'OrderManage' })

  type OrderItem = Api.Order.OrderItem

  const { hasAuth } = useAuth()

  // 订单状态文案与后端 order/enums 一致
  const STATUS_TEXT: Record<number, string> = {
    100: '待支付',
    200: '待接单',
    300: '制作中',
    400: '待取餐',
    500: '已完成',
    600: '已取消',
    700: '退款中',
    800: '已退款'
  }

  const statusOptions = Object.entries(STATUS_TEXT).map(([value, label]) => ({
    label,
    value: Number(value)
  }))

  const typeOptions = [
    { label: '堂食', value: 1 },
    { label: '外卖', value: 2 },
    { label: '自提', value: 3 }
  ]

  const payStatusOptions = [
    { label: '未支付', value: 0 },
    { label: '已支付', value: 10 },
    { label: '已退款', value: 20 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'info' | 'danger' | 'primary'> = {
    100: 'warning',
    200: 'warning',
    300: 'primary',
    400: 'primary',
    500: 'success',
    600: 'info',
    700: 'danger',
    800: 'info'
  }

  const searchForm = ref<Api.Order.OrderSearchParams>({
    keyword: '',
    status: undefined,
    payStatus: undefined,
    type: undefined
  })

  const searchItems = computed(() => [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '订单号 / 取餐码 / 收货电话',
        clearable: true
      }
    },
    {
      label: '订单状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    },
    {
      label: '支付状态',
      key: 'payStatus',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: payStatusOptions
      }
    },
    {
      label: '订单类型',
      key: 'type',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: typeOptions
      }
    }
  ])

  const detailVisible = ref(false)
  const currentOrderId = ref('')

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
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetOrderList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'no',
          label: '订单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'pickupCode',
          label: '取餐码',
          width: 90
        },
        {
          prop: 'typeText',
          label: '类型',
          width: 80
        },
        {
          prop: 'channelText',
          label: '渠道',
          width: 100
        },
        {
          prop: 'tableName',
          label: '桌台',
          width: 100
        },
        {
          prop: 'payAmount',
          label: '应付（元）',
          width: 110,
          formatter: (row: OrderItem) => formatYuan(row.payAmount)
        },
        {
          prop: 'status',
          label: '订单状态',
          width: 100,
          formatter: (row: OrderItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusText || '未知'
            )
        },
        {
          prop: 'payStatus',
          label: '支付状态',
          width: 100,
          formatter: (row: OrderItem) =>
            h(
              ElTag,
              {
                type: row.payStatus === 10 ? 'success' : row.payStatus === 20 ? 'info' : 'warning',
                effect: 'plain'
              },
              () => row.payStatusText || '未知'
            )
        },
        {
          prop: 'receiverName',
          label: '收货人',
          width: 110
        },
        { prop: 'createdAt', label: '下单时间', width: 170 },
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
              // 接单：待接单 → 制作中
              ...(row.status === 200 && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '接单',
                      onClick: () => transitOrder(row, 300)
                    })
                  ]
                : []),
              // 出餐：制作中 → 待取餐
              ...(row.status === 300 && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '出餐',
                      onClick: () => transitOrder(row, 400)
                    })
                  ]
                : []),
              // 完成：待取餐 / 制作中 → 已完成
              ...([300, 400].includes(row.status) && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '完成',
                      onClick: () => transitOrder(row, 500)
                    })
                  ]
                : []),
              // 取消：仅待支付可取消，其余状态要走退款
              ...(row.status === 100 && hasAuth('order:order:cancel')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '取消',
                      onClick: () => cancelOrder(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载订单列表
  const handleSearch = (params: Api.Order.OrderSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开订单详情弹窗，按主键拉取明细与日志
  const openDetail = (row: OrderItem): void => {
    currentOrderId.value = row.id
    detailVisible.value = true
  }

  // 执行状态流转：带原因的流转（如拒单）弹窗收集原因，其余直接确认
  const transitOrder = async (row: OrderItem, to: number): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要将订单「${row.no}」流转到「${STATUS_TEXT[to]}」吗？`,
        '流转订单',
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
      await fetchTransitOrder(row.id, { to })
      ElMessage.success('已流转')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 取消订单：只允许待支付状态，原因可为空
  const cancelOrder = async (row: OrderItem): Promise<void> => {
    try {
      const { value } = await ElMessageBox.prompt(
        `确定要取消订单「${row.no}」吗？取消后不可恢复。`,
        '取消订单',
        {
          confirmButtonText: '确定取消',
          cancelButtonText: '返回',
          inputPlaceholder: '取消原因（选填）',
          inputValue: ''
        }
      )
      await fetchCancelOrder(row.id, value || undefined)
      ElMessage.success('已取消')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
