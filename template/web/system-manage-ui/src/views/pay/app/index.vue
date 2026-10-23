<!-- 支付应用管理页面：维护接入支付能力的应用与支付 / 退款回调地址 -->
<template>
  <div class="pay-app-page art-full-height">
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
            <ElButton v-auth="'pay:app:create'" @click="showDialog('add')" v-ripple>
              新增支付应用
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

      <AppDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :app-data="currentApp"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeletePayApp, fetchGetPayAppList } from '@/api/pay/app'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import AppDialog from './modules/app-dialog.vue'

  defineOptions({ name: 'PayApp' })

  type AppItem = Api.Pay.AppItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  // 支付模块的开启状态沿用 yudao 口径：0 开启 / 1 关闭
  const statusOptions = [
    { label: '开启', value: 0 },
    { label: '关闭', value: 1 }
  ]

  const statusTagTypes: Record<number, 'success' | 'info'> = {
    0: 'success',
    1: 'info'
  }

  const searchForm = ref<Api.Pay.AppSearchParams>({
    keyword: '',
    status: undefined
  })

  const searchItems = [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '应用名 / 应用标识',
        clearable: true
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentApp = ref<AppItem | null>(null)

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
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetPayAppList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'name',
          label: '应用名',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'appKey',
          label: '应用标识',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: AppItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || (row.status === 0 ? '开启' : '关闭')
            )
        },
        { prop: 'channelCount', label: '渠道数', width: 90 },
        {
          prop: 'orderNotifyUrl',
          label: '支付回调地址',
          minWidth: 220,
          showOverflowTooltip: true
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 150,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: AppItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('pay:app:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('pay:app:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteApp(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载支付应用列表
  const handleSearch = (params: Api.Pay.AppSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: AppItem): void => {
    dialogType.value = type
    currentApp.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 弹窗提交成功后按新增 / 编辑分别刷新列表
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 删除支付应用，确认后调用接口并刷新列表
  const deleteApp = async (row: AppItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除支付应用「${row.name}」吗？应用下仍有渠道配置时会被拒绝。`,
        '删除支付应用',
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
      await fetchDeletePayApp(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
