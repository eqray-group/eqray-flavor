<!-- 营销活动管理页面：配置活动规则、互斥与名额 -->
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
          <ElButton v-auth="'promotion:activity:create'" type="primary" @click="openDialog('add')">
            新增活动
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

    <ActivityDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :activity-id="currentActivityId"
      @success="handleSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteActivity, fetchGetActivityList } from '@/api/promotion/promotion'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import ActivityDialog from './modules/activity-dialog.vue'

  defineOptions({ name: 'PromotionActivity' })

  type ActivityItem = Api.Promotion.ActivityItem

  const { hasAuth } = useAuth()

  const typeOptions = [
    { label: '满减', value: 1 },
    { label: '第二份半价', value: 2 },
    { label: '限时折扣', value: 3 },
    { label: '套餐特惠', value: 4 },
    { label: '加价购', value: 5 }
  ]

  const statusOptions = [
    { label: '已启用', value: 1 },
    { label: '未启用', value: 0 }
  ]

  const searchForm = ref<Api.Promotion.ActivitySearchParams>({
    keyword: '',
    type: undefined,
    status: undefined
  })

  const searchItems = computed(() => [
    {
      label: '活动名称',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '活动名称',
        clearable: true
      }
    },
    {
      label: '活动玩法',
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
  const currentActivityId = ref('')

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
      apiFn: fetchGetActivityList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'name',
          label: '活动名称',
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'typeText',
          label: '玩法',
          width: 110
        },
        {
          prop: 'levelText',
          label: '层级',
          width: 90
        },
        {
          prop: 'scopeText',
          label: '门店范围',
          width: 100
        },
        {
          prop: 'priority',
          label: '优先级',
          width: 90
        },
        {
          prop: 'mutexGroup',
          label: '互斥组',
          width: 110,
          formatter: (row: ActivityItem) => row.mutexGroup || '-'
        },
        {
          prop: 'exclusive',
          label: '独占',
          width: 80,
          formatter: (row: ActivityItem) => (row.exclusive === 1 ? '是' : '否')
        },
        {
          prop: 'stackable',
          label: '可叠加',
          width: 90,
          formatter: (row: ActivityItem) => (row.stackable === 1 ? '是' : '否')
        },
        {
          prop: 'quotaLimit',
          label: '名额',
          width: 90,
          formatter: (row: ActivityItem) => (row.quotaLimit > 0 ? row.quotaLimit : '不限')
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: ActivityItem) =>
            h(
              ElTag,
              { type: row.status === 1 ? 'success' : 'info', effect: 'plain' },
              () => row.statusText || '未知'
            )
        },
        { prop: 'startTime', label: '开始时间', width: 170 },
        { prop: 'endTime', label: '结束时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 160,
          fixed: 'right',
          formatter: (row: ActivityItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('promotion:activity:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => openDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('promotion:activity:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => removeActivity(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载活动列表
  const handleSearch = (params: Api.Promotion.ActivitySearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗
  const openDialog = (type: 'add' | 'edit', row?: ActivityItem): void => {
    dialogType.value = type
    currentActivityId.value = row?.id || ''
    dialogVisible.value = true
  }

  // 弹窗提交成功后刷新列表
  const handleSuccess = (): void => {
    refreshData()
  }

  // 删除营销活动
  const removeActivity = async (row: ActivityItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除活动「${row.name}」吗？`, '删除营销活动', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchDeleteActivity(row.id)
      ElMessage.success('已删除')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
