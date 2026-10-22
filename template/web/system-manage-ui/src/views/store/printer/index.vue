<!-- 打印设备管理页面：按当前门店上下文维护小票机、后厨机、标签机与叫号设备 -->
<template>
  <div class="store-printer-page art-full-height">
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
            <ElButton v-auth="'store:printer:create'" @click="showDialog('add')" v-ripple>
              新增打印设备
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

      <PrinterDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :printer-data="currentPrinter"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeletePrinter, fetchGetPrinterList } from '@/api/store/printer'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import PrinterDialog from './modules/printer-dialog.vue'

  defineOptions({ name: 'StorePrinter' })

  type PrinterItem = Api.Store.PrinterItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const printerTypeOptions = [
    { label: '前台小票', value: 1 },
    { label: '后厨', value: 2 },
    { label: '标签', value: 3 },
    { label: '叫号', value: 4 }
  ]

  const accessModeOptions = [
    { label: '云打印机', value: 1 },
    { label: '局域网', value: 2 },
    { label: '本地客户端', value: 3 }
  ]

  const statusOptions = [
    { label: '启用', value: 1 },
    { label: '禁用', value: 0 }
  ]

  const statusTagTypes: Record<number, 'success' | 'info'> = {
    1: 'success',
    0: 'info'
  }

  const searchForm = ref<Api.Store.PrinterSearchParams>({
    keyword: '',
    type: undefined,
    accessMode: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '设备名称 / 设备号',
        clearable: true
      }
    },
    {
      label: '设备类型',
      key: 'type',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: printerTypeOptions
      }
    },
    {
      label: '接入方式',
      key: 'accessMode',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: accessModeOptions
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
  const currentPrinter = ref<PrinterItem | null>(null)

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
      apiFn: fetchGetPrinterList,
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
          label: '设备名称',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'deviceNo',
          label: '设备号',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'typeName',
          label: '设备类型',
          width: 110
        },
        {
          prop: 'accessModeName',
          label: '接入方式',
          width: 120
        },
        {
          prop: 'secretMask',
          label: '密钥',
          width: 110,
          formatter: (row: PrinterItem) => (row.hasSecret ? row.secretMask || '已配置' : '未配置')
        },
        { prop: 'copies', label: '份数', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PrinterItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || (row.status === 1 ? '启用' : '禁用')
            )
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
          formatter: (row: PrinterItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('store:printer:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('store:printer:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deletePrinter(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后回到第一页重新加载打印设备列表
  const handleSearch = (params: Api.Store.PrinterSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: PrinterItem): void => {
    dialogType.value = type
    currentPrinter.value = type === 'edit' ? (row ?? null) : null
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

  // 删除打印设备，确认后调用接口并刷新列表
  const deletePrinter = async (row: PrinterItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除打印设备「${row.name}」吗？删除后该门店可重新绑定设备号 ${row.deviceNo}。`,
        '删除打印设备',
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
      await fetchDeletePrinter(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
