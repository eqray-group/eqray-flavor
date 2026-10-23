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
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'pay:refund:list'" :loading="exporting" @click="exportRefunds">
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

      <RefundDetailDialog v-model:visible="detailVisible" :refund-id="currentRefundId" />
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
    fetchApprovePayRefund,
    fetchExportPayRefunds,
    fetchGetPayRefundList,
    fetchRejectPayRefund
  } from '@/api/pay/refund'
  import { exportFilename, saveBlob } from '@/utils/download'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
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

  // 审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回
  const auditStatusOptions = [
    { label: '无需审核', value: 0 },
    { label: '待审核', value: 1 },
    { label: '审核通过', value: 2 },
    { label: '审核驳回', value: 3 }
  ]

  const auditTagTypes: Record<number, 'warning' | 'success' | 'danger' | 'info'> = {
    0: 'info',
    1: 'warning',
    2: 'success',
    3: 'danger'
  }

  const { hasAuth } = useAuth()

  const appOptions = ref<Api.Pay.AppOption[]>([])
  const channelOptions = ref<Api.Pay.ChannelOption[]>([])

  const searchForm = ref<Api.Pay.RefundSearchParams>({
    keyword: '',
    appId: '',
    channelCode: '',
    status: undefined,
    auditStatus: undefined
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
    },
    {
      label: '审核状态',
      key: 'auditStatus',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: auditStatusOptions
      }
    }
  ])

  const detailVisible = ref(false)
  const exporting = ref(false)
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
          prop: 'auditStatus',
          label: '审核状态',
          width: 110,
          formatter: (row: RefundItem) =>
            h(
              ElTag,
              { type: auditTagTypes[row.auditStatus] ?? 'info', effect: 'plain' },
              () => row.auditStatusName || '无需审核'
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
          width: 220,
          fixed: 'right',
          formatter: (row: RefundItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '详情',
                onClick: () => openDetail(row)
              }),
              // 只有待审核的单才出现审核动作，且需要单独的审核权限
              ...(row.auditStatus === 1 && hasAuth('pay:refund:audit')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '通过',
                      onClick: () => approveRefund(row)
                    }),
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '驳回',
                      onClick: () => rejectRefund(row)
                    })
                  ]
                : [])
            ])
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

  // 导出当前筛选条件下的退款单，含审核状态便于对账溯源
  const exportRefunds = async (): Promise<void> => {
    exporting.value = true
    try {
      const blob = await fetchExportPayRefunds({ ...searchParams })
      saveBlob(blob, exportFilename('pay-refunds'))
      ElMessage.success('导出成功')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      exporting.value = false
    }
  }

  // 打开退款单详情弹窗
  const openDetail = (row: RefundItem): void => {
    currentRefundId.value = row.id
    detailVisible.value = true
  }

  // 审核通过：二次确认后提交渠道（钱包单直接回冲余额）
  const approveRefund = async (row: RefundItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定通过退款单「${row.no}」吗？通过后将立即向用户退款 ${formatYuan(row.refundPrice)} 元，不可撤回。`,
        '审核通过',
        {
          confirmButtonText: '确认放款',
          cancelButtonText: '再想想',
          type: 'warning'
        }
      )
    } catch {
      return
    }

    let remark = ''
    try {
      const { value } = await ElMessageBox.prompt('审核意见（选填）', '审核通过', {
        confirmButtonText: '提交',
        cancelButtonText: '取消',
        inputPlaceholder: '如：已核实、客户确认',
        inputValue: ''
      })
      remark = value || ''
    } catch {
      return
    }

    try {
      await fetchApprovePayRefund(row.id, { remark })
      ElMessage.success('已审核通过，退款已提交')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 审核驳回：原因必填，驳回后退款单关闭且不动钱
  const rejectRefund = async (row: RefundItem): Promise<void> => {
    let remark = ''
    try {
      const { value } = await ElMessageBox.prompt(
        `确定驳回退款单「${row.no}」吗？驳回后该单关闭，不会向用户退款。`,
        '审核驳回',
        {
          confirmButtonText: '确认驳回',
          cancelButtonText: '取消',
          inputPlaceholder: '请填写驳回原因（必填）',
          inputValue: '',
          inputValidator: (value: string) => (value?.trim() ? true : '驳回原因必填')
        }
      )
      remark = value || ''
    } catch {
      return
    }

    try {
      await fetchRejectPayRefund(row.id, { remark })
      ElMessage.success('已驳回')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
