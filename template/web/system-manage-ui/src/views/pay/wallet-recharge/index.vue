<!-- 充值记录页面：查看每笔钱包充值的实付、赠送、到账与支付状态 -->
<template>
  <div class="pay-wallet-recharge-page art-full-height">
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

      <RechargeDetailDialog v-model:visible="detailVisible" :recharge-id="currentRechargeId" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayWalletRechargeList } from '@/api/pay/wallet'
  import { ElTag } from 'element-plus'
  import RechargeDetailDialog from './modules/recharge-detail-dialog.vue'

  defineOptions({ name: 'PayWalletRecharge' })

  type RechargeItem = Api.Pay.WalletRechargeItem

  // 0 未支付 / 10 已支付 / 20 已退款
  const statusOptions = [
    { label: '未支付', value: 0 },
    { label: '已支付', value: 10 },
    { label: '已退款', value: 20 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'info'> = {
    0: 'warning',
    10: 'success',
    20: 'info'
  }

  const searchForm = ref<Api.Pay.WalletRechargeSearchParams>({
    walletId: '',
    status: undefined
  })

  const searchItems = [
    {
      label: '钱包 ID',
      key: 'walletId',
      type: 'input',
      span: 8,
      props: {
        placeholder: '钱包 ID，可留空',
        clearable: true
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
  ]

  const detailVisible = ref(false)
  const currentRechargeId = ref('')

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
      apiFn: fetchGetPayWalletRechargeList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'walletId',
          label: '钱包 ID',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'payPrice',
          label: '实付（元）',
          width: 120,
          formatter: (row: RechargeItem) => formatYuan(row.payPrice)
        },
        {
          prop: 'bonusPrice',
          label: '赠送（元）',
          width: 120,
          formatter: (row: RechargeItem) => formatYuan(row.bonusPrice)
        },
        {
          prop: 'totalPrice',
          label: '到账（元）',
          width: 120,
          formatter: (row: RechargeItem) => formatYuan(row.totalPrice)
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: RechargeItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || '未知'
            )
        },
        {
          prop: 'payChannelCode',
          label: '支付渠道',
          width: 130
        },
        { prop: 'payTime', label: '支付时间', width: 170 },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 100,
          fixed: 'right',
          formatter: (row: RechargeItem) =>
            h(ArtButtonTable, {
              type: 'view',
              text: '详情',
              onClick: () => openDetail(row)
            })
        }
      ]
    }
  })

  // 搜索条件变化后重新加载充值记录
  const handleSearch = (params: Api.Pay.WalletRechargeSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开充值记录详情弹窗
  const openDetail = (row: RechargeItem): void => {
    currentRechargeId.value = row.id
    detailVisible.value = true
  }
</script>
