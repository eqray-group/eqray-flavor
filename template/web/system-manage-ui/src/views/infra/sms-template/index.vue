<!-- 短信模板页面（对齐 yudao system_sms_template） -->
<template>
  <div class="member-sms-template-page art-full-height">
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
            <ElButton v-auth="'infra:sms-template:create'" @click="showDialog('add')" v-ripple>
              新增模板
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

      <SmsTemplateDialog
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
  import { fetchDeleteSmsTemplate, fetchGetSmsTemplateList } from '@/api/infra/sms-template'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import SmsTemplateDialog from './modules/sms-template-dialog.vue'

  defineOptions({ name: 'InfraSmsTemplate' })

  type SmsTemplateItem = Api.Infra.SmsTemplateItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.SmsTemplateSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '模板名',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入模板名',
      clearable: true
    },
    {
      label: '类型',
      key: 'type',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '验证码', value: 1 },
          { label: '通知', value: 2 }
        ]
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentTemplate = ref<SmsTemplateItem | null>(null)

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
      apiFn: fetchGetSmsTemplateList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '模板名', minWidth: 140, showOverflowTooltip: true },
        { prop: 'code', label: '模板代码', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'type',
          label: '类型',
          width: 100,
          formatter: (row: SmsTemplateItem) =>
            h(ElTag, { type: 'success', effect: 'plain' }, () =>
              row.type === 1 ? '验证码' : '通知'
            )
        },
        { prop: 'apiTemplateId', label: '第三方模板 ID', minWidth: 160, showOverflowTooltip: true },
        { prop: 'content', label: '模板内容', minWidth: 220, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: SmsTemplateItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '禁用'
            )
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: SmsTemplateItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('infra:sms-template:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:sms-template:delete')
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

  const handleSearch = (params: Api.Infra.SmsTemplateSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: SmsTemplateItem): void => {
    dialogType.value = type
    currentTemplate.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  const deleteTemplate = async (row: SmsTemplateItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除模板「${row.name}」吗？`, '删除模板', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteSmsTemplate(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
