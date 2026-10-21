<!-- 短信渠道配置页面（对齐 yudao system_sms_channel） -->
<template>
  <div class="member-sms-channel-page art-full-height">
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
            <ElButton v-auth="'infra:sms-channel:create'" @click="showDialog('add')" v-ripple>
              新增渠道
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

      <SmsChannelDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :channel-data="currentChannel"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteSmsChannel, fetchGetSmsChannelList } from '@/api/infra/sms-channel'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import SmsChannelDialog from './modules/sms-channel-dialog.vue'

  defineOptions({ name: 'InfraSmsChannel' })

  type SmsChannelItem = Api.Infra.SmsChannelItem
  type DialogType = 'add' | 'edit'

  const CHANNEL_TEXT: Record<string, string> = {
    DebugLog: '本地日志',
    AliyunSms: '阿里云',
    TencentSms: '腾讯云'
  }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.SmsChannelSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '渠道名',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入渠道名',
      clearable: true
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
  const currentChannel = ref<SmsChannelItem | null>(null)

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
      apiFn: fetchGetSmsChannelList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '渠道名', minWidth: 140, showOverflowTooltip: true },
        {
          prop: 'code',
          label: '渠道代码',
          width: 140,
          formatter: (row: SmsChannelItem) =>
            h(ElTag, { type: 'info', effect: 'plain' }, () => CHANNEL_TEXT[row.code] ?? row.code)
        },
        { prop: 'apiKey', label: 'APIKey', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: SmsChannelItem) =>
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
          formatter: (row: SmsChannelItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('infra:sms-channel:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:sms-channel:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteChannel(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.SmsChannelSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: SmsChannelItem): void => {
    dialogType.value = type
    currentChannel.value = type === 'edit' ? (row ?? null) : null
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

  const deleteChannel = async (row: SmsChannelItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除渠道「${row.name}」吗？`, '删除渠道', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteSmsChannel(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
