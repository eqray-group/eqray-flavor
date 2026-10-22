<!-- 桌台区域管理页面：按当前门店上下文维护大厅、包间等用餐分区 -->
<template>
  <div class="store-table-area-page art-full-height">
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
            <ElButton v-auth="'store:table-area:create'" @click="showDialog('add')" v-ripple>
              新增区域
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

      <TableAreaDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :area-data="currentArea"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteTableArea, fetchGetTableAreaList } from '@/api/store/table'
  import { ElButton, ElMessage, ElMessageBox, ElSpace } from 'element-plus'
  import TableAreaDialog from './modules/table-area-dialog.vue'

  defineOptions({ name: 'StoreTableArea' })

  type TableAreaItem = Api.Store.TableAreaItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Store.TableAreaSearchParams>({
    keyword: ''
  })

  const searchItems = [
    {
      label: '区域名称',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '请输入区域名称',
        clearable: true
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentArea = ref<TableAreaItem | null>(null)

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
      apiFn: fetchGetTableAreaList,
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
          prop: 'name',
          label: '区域名称',
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'tableCount',
          label: '桌台数',
          width: 100
        },
        {
          prop: 'sort',
          label: '排序',
          width: 90
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 180,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: TableAreaItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('store:table-area:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('store:table-area:delete') && row.tableCount === 0
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteArea(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后回到第一页重新加载区域列表
  const handleSearch = (params: Api.Store.TableAreaSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: TableAreaItem): void => {
    dialogType.value = type
    currentArea.value = type === 'edit' ? (row ?? null) : null
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

  // 删除区域，确认后调用接口并刷新列表
  const deleteArea = async (row: TableAreaItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除区域「${row.name}」吗？该操作不可恢复。`,
        '删除桌台区域',
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
      await fetchDeleteTableArea(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
