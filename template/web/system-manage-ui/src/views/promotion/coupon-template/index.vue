<!-- 优惠券模板管理页面：维护券规则、发放限制与有效期 -->
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
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #title>
          <ElButton
            v-auth="'promotion:coupon-template:create'"
            type="primary"
            @click="openDialog('add')"
          >
            新增券模板
          </ElButton>
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
    </ElCard>

    <CouponTemplateDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :template-id="currentTemplateId"
      @success="handleSuccess"
    />

    <CouponIssueDialog
      v-model:visible="issueVisible"
      :template="currentTemplate"
      @success="handleSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteCouponTemplate, fetchGetCouponTemplateList } from '@/api/promotion/promotion'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import CouponTemplateDialog from './modules/coupon-template-dialog.vue'
  import CouponIssueDialog from './modules/coupon-issue-dialog.vue'

  defineOptions({ name: 'PromotionCouponTemplate' })

  type TemplateItem = Api.Promotion.CouponTemplateItem

  const { hasAuth } = useAuth()

  const typeOptions = [
    { label: '满减券', value: 1 },
    { label: '折扣券', value: 2 },
    { label: '单品券', value: 3 },
    { label: '兑换券', value: 4 },
    { label: '代金券', value: 5 }
  ]

  const statusOptions = [
    { label: '开启', value: 1 },
    { label: '关闭', value: 0 }
  ]

  const searchForm = ref<Api.Promotion.CouponTemplateSearchParams>({
    keyword: '',
    type: undefined,
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '券名称',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '券名称',
        clearable: true
      }
    },
    {
      label: '券类型',
      key: 'type',
      type: 'select',
      span: 8,
      props: {
        placeholder: '全部',
        clearable: true,
        options: typeOptions
      }
    },
    {
      label: '启用状态',
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

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentTemplateId = ref('')

  const issueVisible = ref(false)
  const currentTemplate = ref<Api.Promotion.CouponTemplateItem | null>(null)

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
      apiFn: fetchGetCouponTemplateList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'name',
          label: '券名称',
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'typeText',
          label: '券类型',
          width: 100
        },
        {
          prop: 'thresholdAmount',
          label: '门槛（元）',
          width: 100,
          formatter: (row: TemplateItem) => formatYuan(row.thresholdAmount)
        },
        {
          prop: 'discountAmount',
          label: '面额（元）',
          width: 100,
          formatter: (row: TemplateItem) =>
            row.type === 2 ? formatYuan(row.maxDiscountAmount) : formatYuan(row.discountAmount)
        },
        {
          prop: 'discountRate',
          label: '折扣率',
          width: 90,
          formatter: (row: TemplateItem) =>
            row.type === 2 ? `${(row.discountRate / 1000).toFixed(1)} 折` : '-'
        },
        {
          prop: 'scopeText',
          label: '适用范围',
          width: 100
        },
        {
          prop: 'channelLimitText',
          label: '渠道限制',
          width: 100
        },
        {
          prop: 'issueTypeText',
          label: '发放方式',
          width: 100
        },
        {
          prop: 'issuedCount',
          label: '已发放',
          width: 100,
          formatter: (row: TemplateItem) =>
            row.totalCount > 0 ? `${row.issuedCount} / ${row.totalCount}` : `${row.issuedCount}`
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: TemplateItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info', effect: 'plain' }, () =>
              row.status === 1 ? '开启' : '关闭'
            )
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 240,
          fixed: 'right',
          formatter: (row: TemplateItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('promotion:coupon-template:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => openDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('promotion:coupon-template:issue')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '发券',
                      onClick: () => issueCoupon(row)
                    })
                  ]
                : []),
              ...(hasAuth('promotion:coupon-template:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => removeTemplate(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载模板列表
  const handleSearch = (params: Api.Promotion.CouponTemplateSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗
  const openDialog = (type: 'add' | 'edit', row?: TemplateItem): void => {
    dialogType.value = type
    currentTemplateId.value = row?.id || ''
    dialogVisible.value = true
  }

  // 弹窗提交成功后按类型刷新列表
  const handleSuccess = (): void => {
    refreshData()
  }

  // 打开发券弹窗：会员与张数在弹窗里填写
  const issueCoupon = (row: TemplateItem): void => {
    currentTemplate.value = row
    issueVisible.value = true
  }

  // 删除券模板
  const removeTemplate = async (row: TemplateItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除券模板「${row.name}」吗？`, '删除券模板', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchDeleteCouponTemplate(row.id)
      ElMessage.success('已删除')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
