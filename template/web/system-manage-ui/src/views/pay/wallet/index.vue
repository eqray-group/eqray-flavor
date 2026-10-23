<!-- 会员钱包页面：查看余额与累计收支，并支持后台调账与流水分页查询 -->
<template>
  <div class="pay-wallet-page art-full-height">
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

      <WalletTransactionDrawer v-model:visible="drawerVisible" :wallet="currentWallet" />
      <WalletChangeDialog
        v-model:visible="changeVisible"
        :wallet="currentWallet"
        @success="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetPayWalletList } from '@/api/pay/wallet'
  import WalletChangeDialog from './modules/wallet-change-dialog.vue'
  import WalletTransactionDrawer from './modules/wallet-transaction-drawer.vue'

  defineOptions({ name: 'PayWallet' })

  type WalletItem = Api.Pay.WalletItem

  const { hasAuth } = useAuth()

  // 用户类型：1 会员
  const userTypeOptions = [{ label: '会员', value: 1 }]

  const searchForm = ref<Api.Pay.WalletSearchParams>({
    userId: '',
    userType: undefined
  })

  const searchItems = [
    {
      label: '用户 ID',
      key: 'userId',
      type: 'input',
      span: 8,
      props: {
        placeholder: '会员用户 ID',
        clearable: true
      }
    },
    {
      label: '用户类型',
      key: 'userType',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: userTypeOptions
      }
    }
  ]

  const drawerVisible = ref(false)
  const changeVisible = ref(false)
  const currentWallet = ref<WalletItem | null>(null)

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
      apiFn: fetchGetPayWalletList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'userId',
          label: '用户 ID',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'userType',
          label: '用户类型',
          width: 100,
          formatter: (row: WalletItem) => (row.userType === 1 ? '会员' : row.userType)
        },
        {
          prop: 'balance',
          label: '可用余额（元）',
          width: 130,
          formatter: (row: WalletItem) => formatYuan(row.balance)
        },
        {
          prop: 'freezePrice',
          label: '冻结（元）',
          width: 110,
          formatter: (row: WalletItem) => formatYuan(row.freezePrice)
        },
        {
          prop: 'totalRecharge',
          label: '累计充值（元）',
          width: 130,
          formatter: (row: WalletItem) => formatYuan(row.totalRecharge)
        },
        {
          prop: 'totalExpense',
          label: '累计支出（元）',
          width: 130,
          formatter: (row: WalletItem) => formatYuan(row.totalExpense)
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 180,
          fixed: 'right',
          formatter: (row: WalletItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '流水',
                onClick: () => openDrawer(row)
              }),
              ...(hasAuth('pay:wallet:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '调账',
                      onClick: () => openChange(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载钱包列表
  const handleSearch = (params: Api.Pay.WalletSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开流水分页抽屉
  const openDrawer = (row: WalletItem): void => {
    currentWallet.value = row
    drawerVisible.value = true
  }

  // 打开调账弹窗
  const openChange = (row: WalletItem): void => {
    currentWallet.value = row
    changeVisible.value = true
  }
</script>
