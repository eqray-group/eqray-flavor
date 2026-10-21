<!-- 参数配置页面：键值对形式的系统参数维护 -->
<template>
  <div class="config-page art-full-height">
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
            <ElButton v-auth="'infra:config:create'" @click="showDialog('add')" v-ripple>
              新增参数
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
    </ElCard>

    <ConfigDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :config-data="currentConfig"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteConfig, fetchGetConfigList } from '@/api/infra/config'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ConfigDialog from './modules/config-dialog.vue'

  defineOptions({ name: 'InfraConfig' })

  type ConfigItem = Api.Infra.ConfigItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.ConfigSearchParams>({
    category: undefined,
    keyword: undefined
  })

  const searchItems = [
    {
      label: '参数分组',
      key: 'category',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '文件', value: '文件' },
          { label: '登录', value: '登录' },
          { label: '通用', value: '通用' },
          { label: '会员', value: '会员' },
          { label: '支付', value: '支付' }
        ]
      }
    },
    {
      label: '参数名称',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入参数名称',
      clearable: true
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentConfig = ref<ConfigItem | null>(null)

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
      apiFn: fetchGetConfigList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'category',
          label: '参数分组',
          width: 110,
          formatter: (row: ConfigItem) => row.category || '-'
        },
        { prop: 'name', label: '参数名称', minWidth: 150, showOverflowTooltip: true },
        {
          prop: 'configKey',
          label: '参数键',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'value',
          label: '参数值',
          minWidth: 160,
          showOverflowTooltip: true,
          formatter: (row: ConfigItem) => row.value || '-'
        },
        {
          prop: 'type',
          label: '类型',
          width: 100,
          formatter: (row: ConfigItem) =>
            h(ElTag, { type: row.type === 1 ? 'warning' : 'info' }, () =>
              row.type === 1 ? '系统内置' : '自定义'
            )
        },
        {
          prop: 'visible',
          label: '公开',
          width: 90,
          formatter: (row: ConfigItem) =>
            h(ElTag, { type: row.visible === 1 ? 'success' : 'info' }, () =>
              row.visible === 1 ? '是' : '否'
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 160,
          showOverflowTooltip: true,
          formatter: (row: ConfigItem) => row.remark || '-'
        },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: ConfigItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('infra:config:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:config:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      disabled: row.type === 1,
                      title: row.type === 1 ? '系统内置参数不允许删除' : '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.ConfigSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: ConfigItem): void => {
    dialogType.value = type
    currentConfig.value = type === 'edit' ? (row ?? null) : null
    dialogVisible.value = true
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'edit') {
      refreshUpdate()
    } else {
      refreshCreate()
    }
  }

  const handleDelete = async (row: ConfigItem): Promise<void> => {
    if (row.type === 1) return

    try {
      await ElMessageBox.confirm(`确定要删除参数「${row.name}」吗？`, '删除参数', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteConfig(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
