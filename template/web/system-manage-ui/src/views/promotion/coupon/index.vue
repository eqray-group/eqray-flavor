<!-- 会员券管理页面：查看已发放的券并按需作废 -->
<template>
  <div class="promotion-page art-full-height">
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
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetCouponList, fetchRevokeCoupon } from '@/api/promotion/promotion'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'PromotionCoupon' })

  type CouponItem = Api.Promotion.CouponItem

  const { hasAuth } = useAuth()

  const statusOptions = [
    { label: '未使用', value: 1 },
    { label: '锁定', value: 2 },
    { label: '已使用', value: 3 },
    { label: '已过期', value: 4 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'info' | 'danger' | 'primary'> = {
    1: 'success',
    2: 'warning',
    3: 'info',
    4: 'info'
  }

  const searchForm = ref<Api.Promotion.CouponSearchParams>({
    templateId: '',
    memberId: '',
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '会员 ID',
      key: 'memberId',
      type: 'input',
      span: 8,
      props: {
        placeholder: '会员 ID',
        clearable: true
      }
    },
    {
      label: '券状态',
      key: 'status',
      type: 'select',
      span: 8,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    }
  ])

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
      apiFn: fetchGetCouponList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'templateName',
          label: '券名称',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'memberId',
          label: '会员 ID',
          width: 160
        },
        {
          prop: 'status',
          label: '券状态',
          width: 100,
          formatter: (row: CouponItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusText || '未知'
            )
        },
        {
          prop: 'orderId',
          label: '关联订单',
          width: 170,
          formatter: (row: CouponItem) => (row.orderId && row.orderId !== '0' ? row.orderId : '-')
        },
        { prop: 'getTime', label: '领取时间', width: 170 },
        { prop: 'expireTime', label: '过期时间', width: 170 },
        {
          prop: 'useTime',
          label: '核销时间',
          width: 170,
          formatter: (row: CouponItem) => row.useTime || '-'
        },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row: CouponItem) =>
            h('div', { class: 'flex items-center' }, [
              // 作废只对未使用的券有意义
              ...(row.status === 1 && hasAuth('promotion:coupon:revoke')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '作废',
                      onClick: () => revokeCoupon(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载券列表
  const handleSearch = (params: Api.Promotion.CouponSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 作废会员券
  const revokeCoupon = async (row: CouponItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要作废「${row.templateName}」吗？`, '作废优惠券', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchRevokeCoupon(row.id)
      ElMessage.success('已作废')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
