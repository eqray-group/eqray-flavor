<!-- 字典管理页面：左侧字典类型 + 右侧字典数据 -->
<template>
  <div class="dict-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <div class="dict-layout">
      <!-- 左侧：字典类型 -->
      <ElCard class="art-table-card dict-type-card">
        <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
          <template #left>
            <ElSpace wrap>
              <ElButton v-auth="'system:dict:create'" @click="showDialog('add')" v-ripple>
                新增字典
              </ElButton>
            </ElSpace>
          </template>
        </ArtTableHeader>

        <ArtTable
          :loading="loading"
          :data="data"
          :columns="columns"
          :pagination="pagination"
          highlight-current-row
          row-key="id"
          :current-row-key="currentType?.id"
          @row-click="handleRowClick"
          @pagination:size-change="handleSizeChange"
          @pagination:current-change="handleCurrentChange"
        />
      </ElCard>

      <!-- 右侧：字典数据 -->
      <ElCard class="art-table-card dict-data-card">
        <DictDataPanel :dict-type="currentType" />
      </ElCard>
    </div>

    <DictTypeDialog
      v-model:visible="editVisible"
      :type="editType"
      :dict-type="currentType"
      @success="handleSaveSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteDictType, fetchGetDictTypeList } from '@/api/system/dict'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import DictTypeDialog from './modules/dict-type-dialog.vue'
  import DictDataPanel from './modules/dict-data-panel.vue'

  defineOptions({ name: 'SystemDict' })

  type DictTypeItem = Api.SystemManage.DictTypeItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const editVisible = ref(false)
  const editType = ref<DialogType>('add')
  const currentType = ref<DictTypeItem | null>(null)

  const searchForm = ref<Api.SystemManage.DictTypeSearchParams>({
    name: undefined,
    type: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '字典名称',
      key: 'name',
      type: 'input',
      placeholder: '请输入字典名称',
      clearable: true
    },
    {
      label: '字典类型',
      key: 'type',
      type: 'input',
      placeholder: '请输入字典类型',
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 }
        ]
      }
    }
  ]

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
    refreshUpdate
  } = useTable({
    core: {
      apiFn: fetchGetDictTypeList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '字典名称', minWidth: 120 },
        { prop: 'type', label: '字典类型', minWidth: 150, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 80,
          formatter: (row: DictTypeItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '禁用'
            )
        },
        {
          prop: 'operation',
          label: '操作',
          width: 160,
          formatter: (row: DictTypeItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('system:dict:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('system:dict:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.SystemManage.DictTypeSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: DictTypeItem): void => {
    editType.value = type
    currentType.value = row || null
    nextTick(() => {
      editVisible.value = true
    })
  }

  /** 点击左侧行切换右侧字典数据 */
  const handleRowClick = (row: DictTypeItem): void => {
    currentType.value = row
  }

  // 列表刷新后保持选中项；若选中项已不存在则默认选中第一行
  watch(
    () => data.value,
    (list: DictTypeItem[]) => {
      if (!list?.length) {
        currentType.value = null
        return
      }
      const exists = list.some((item) => item.id === currentType.value?.id)
      if (!exists) currentType.value = list[0]
    },
    { immediate: true }
  )

  const handleSaveSuccess = (): void => {
    if (editType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  const handleDelete = async (row: DictTypeItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除字典「${row.name}」及其全部字典数据吗？`, '删除字典', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteDictType(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>

<style scoped>
  .dict-layout {
    display: flex;
    flex: 1;
    gap: 12px;
    min-height: 0;
  }

  /* 左侧字典类型：固定宽度 */
  .dict-layout .dict-type-card {
    width: 560px;
    flex: none;
  }

  /* 右侧字典数据：自适应剩余宽度 */
  .dict-layout .dict-data-card {
    min-width: 0;
  }

  @media (max-width: 1100px) {
    .dict-layout {
      flex-direction: column;
    }

    .dict-layout .dict-type-card {
      width: 100%;
      flex: none;
      height: 320px;
    }

    .dict-layout .dict-data-card {
      flex: 1;
      min-height: 420px;
    }
  }
</style>
