<!-- 骑手管理页面：按当前门店上下文维护骑手档案与接单并发上限 -->
<template>
  <div class="store-rider-page art-full-height">
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
            <ElButton v-auth="'store:rider:create'" @click="showDialog('add')" v-ripple>
              新增骑手
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

      <RiderDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :rider-data="currentRider"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteRider, fetchGetRiderList } from '@/api/store/rider'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import RiderDialog from './modules/rider-dialog.vue'

  defineOptions({ name: 'StoreRider' })

  type RiderItem = Api.Store.RiderItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const statusOptions = [
    { label: '空闲', value: 1 },
    { label: '接单中', value: 2 },
    { label: '配送中', value: 3 },
    { label: '休息', value: 4 },
    { label: '停用', value: 5 }
  ]

  const statusTagTypes: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    1: 'info',
    2: 'success',
    3: 'primary',
    4: 'warning',
    5: 'danger'
  }

  const searchForm = ref<Api.Store.RiderSearchParams>({
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
        placeholder: '账号 / 昵称 / 手机号',
        clearable: true
      }
    },
    {
      label: '骑手状态',
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
  const currentRider = ref<RiderItem | null>(null)

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
      apiFn: fetchGetRiderList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'storeName',
          label: '门店',
          minWidth: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'username',
          label: '骑手',
          minWidth: 180,
          formatter: (row: RiderItem) => formatRider(row)
        },
        { prop: 'mobile', label: '手机号', width: 130 },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: RiderItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () =>
                row.statusName ||
                statusOptions.find((item) => item.value === row.status)?.label ||
                '未知'
            )
        },
        { prop: 'maxOrders', label: '并发上限', width: 100 },
        { prop: 'currentOrders', label: '进行中', width: 90 },
        { prop: 'todayOrders', label: '今日单量', width: 100 },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 160,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: RiderItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('store:rider:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('store:rider:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteRider(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 骑手列优先展示昵称，没有昵称时回退到登录账号
  function formatRider(row: RiderItem): string {
    if (row.nickname && row.username) return `${row.nickname}（${row.username}）`
    return row.nickname || row.username || '-'
  }

  // 搜索条件变化后回到第一页重新加载骑手列表
  const handleSearch = (params: Api.Store.RiderSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: RiderItem): void => {
    dialogType.value = type
    currentRider.value = type === 'edit' ? (row ?? null) : null
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

  // 删除骑手档案，确认后调用接口并刷新列表
  const deleteRider = async (row: RiderItem): Promise<void> => {
    const riderLabel = row.nickname || row.username || '该骑手'
    try {
      await ElMessageBox.confirm(
        `确定要删除「${riderLabel}」的骑手档案吗？删除后门店员工关联会同步停用。`,
        '删除骑手',
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
      await fetchDeleteRider(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
