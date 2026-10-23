<!-- 退款单管理页面：查询退款结果、查看渠道退款单号与回调原文 -->
<template>
  <div class="pay-refund-page art-full-height">
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

      <RefundDetailDialog v-model:visible="detailVisible" :refund-id="currentRefundId" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayAppOptions } from '@/api/pay/app'
  import { fetchGetPayChannelOptions } from '@/api/pay/channel'
  import { fetchGetPayRefundList } from '@/api/pay/refund'
  import { ElTag } from 'element-plus'
  import RefundDetailDialog from './modules/refund-detail-dialog.vue'

  defineOptions({ name: 'PayRefund' })

  type RefundItem = Api.Pay.RefundItem

  // 0 等待退款 / 10 退款成功 / 20 退款失败
  // 取值对齐 yudao：0 退款订单生成 / 1 退款成功 / 2 退款失败 / 99 退款关闭
  const statusOptions = [
    { label: '退款订单生成', value: 0 },
    { label: '退款成功', value: 1 },
    { label: '退款失败', value: 2 },
    { label: '退款关闭', value: 99 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'danger' | 'info'> = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    99: 'info'
  }

  const appOptions = ref<Api.Pay.AppOption[]>([])
  const channelOptions = ref<Api.Pay.ChannelOption[]>([])

  const searchForm = ref<Api.Pay.RefundSearchParams>({
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
        placeholder: '退款单号 / 支付单号 / 渠道退款单号',
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

  const detailVisible = ref(false)
  const currentRefundId = ref('')

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
      apiFn: fetchGetPayRefundList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'no',
          label: '退款单号',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'orderNo',
          label: '支付单号',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'merchantRefundId',
          label: '业务退款单号',
          minWidth: 170,
          showOverflowTooltip: true
        },
        {
          prop: 'channelName',
          label: '渠道',
          width: 130
        },
        {
          prop: 'refundPrice',
          label: '退款金额（元）',
          width: 130,
          formatter: (row: RefundItem) => formatYuan(row.refundPrice)
        },
        {
          prop: 'status',
          label: '状态',
          width: 110,
          formatter: (row: RefundItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || '未知'
            )
        },
        {
          prop: 'channelRefundNo',
          label: '渠道退款单号',
          minWidth: 200,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 100,
          fixed: 'right',
          formatter: (row: RefundItem) =>
            h(ArtButtonTable, {
              type: 'view',
              text: '详情',
              onClick: () => openDetail(row)
            })
        }
      ]
    }
  })

  // 加载应用与渠道编码下拉，供筛选使用
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

  // 搜索条件变化后重新加载退款单列表
  const handleSearch = (params: Api.Pay.RefundSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开退款单详情弹窗
  const openDetail = (row: RefundItem): void => {
    currentRefundId.value = row.id
    detailVisible.value = true
  }
</script>
