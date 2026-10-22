<!-- 打印模板管理页面：按当前门店维护各机型与打印类型的模板与版本号 -->
<template>
  <div class="store-print-template-page art-full-height">
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
            <ElButton v-auth="'store:print-template:create'" @click="showDialog('add')" v-ripple>
              新增打印模板
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

      <PrintTemplateDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :template-data="currentTemplate"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeletePrintTemplate, fetchGetPrintTemplateList } from '@/api/store/print-template'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import PrintTemplateDialog from './modules/print-template-dialog.vue'

  defineOptions({ name: 'StorePrintTemplate' })

  type PrintTemplateItem = Api.Store.PrintTemplateItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const printerTypeOptions = [
    { label: '前台小票', value: 1 },
    { label: '后厨', value: 2 },
    { label: '标签', value: 3 },
    { label: '叫号', value: 4 }
  ]

  const printTypeOptions = [
    { label: '结账单', value: 1 },
    { label: '预结单', value: 2 },
    { label: '厨打', value: 3 },
    { label: '退菜', value: 4 },
    { label: '标签', value: 5 },
    { label: '退款单', value: 6 },
    { label: '交班', value: 7 },
    { label: '日结', value: 8 }
  ]

  const statusOptions = [
    { label: '启用', value: 1 },
    { label: '禁用', value: 0 }
  ]

  const statusTagTypes: Record<number, 'success' | 'info'> = {
    1: 'success',
    0: 'info'
  }

  const searchForm = ref<Api.Store.PrintTemplateSearchParams>({
    keyword: '',
    printerType: undefined,
    printType: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '模板名称',
      key: 'keyword',
      type: 'input',
      span: 7,
      props: {
        placeholder: '请输入模板名称',
        clearable: true
      }
    },
    {
      label: '设备类型',
      key: 'printerType',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: printerTypeOptions
      }
    },
    {
      label: '打印类型',
      key: 'printType',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: printTypeOptions
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

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentTemplate = ref<PrintTemplateItem | null>(null)

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
      apiFn: fetchGetPrintTemplateList,
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
          label: '模板名称',
          minWidth: 170,
          showOverflowTooltip: true
        },
        { prop: 'printerTypeName', label: '设备类型', width: 110 },
        { prop: 'printTypeName', label: '打印类型', width: 100 },
        {
          prop: 'paperWidth',
          label: '纸宽',
          width: 90,
          formatter: (row: PrintTemplateItem) => `${row.paperWidth} mm`
        },
        {
          prop: 'version',
          label: '版本',
          width: 90,
          formatter: (row: PrintTemplateItem) => `v${row.version}`
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PrintTemplateItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || (row.status === 1 ? '启用' : '禁用')
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 140,
          showOverflowTooltip: true
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: PrintTemplateItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('store:print-template:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('store:print-template:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteTemplate(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后回到第一页重新加载打印模板列表
  const handleSearch = (params: Api.Store.PrintTemplateSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: PrintTemplateItem): void => {
    dialogType.value = type
    currentTemplate.value = type === 'edit' ? (row ?? null) : null
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

  // 删除打印模板，确认后调用接口并刷新列表
  const deleteTemplate = async (row: PrintTemplateItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除打印模板「${row.name}」吗？删除后该机型与打印类型可以重新创建模板。`,
        '删除打印模板',
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
      await fetchDeletePrintTemplate(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
