<!-- 第三方登录渠道配置页面（对齐 yudao 社交客户端） -->
<template>
  <div class="member-social-client-page art-full-height">
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
            <ElButton v-auth="'member:social-client:create'" @click="showDialog('add')" v-ripple>
              新增配置
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

      <SocialClientDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :client-data="currentClient"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteSocialClient, fetchGetSocialClientList } from '@/api/member/social-client'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import SocialClientDialog from './modules/social-client-dialog.vue'

  defineOptions({ name: 'MemberSocialClient' })

  type SocialClientItem = Api.Member.SocialClientItem
  type DialogType = 'add' | 'edit'

  const SOCIAL_TYPE_TEXT: Record<string, string> = {
    wechat_mini: '微信小程序',
    wechat_mp: '微信公众号',
    wechat_open: '微信开放平台'
  }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Member.SocialClientSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '应用名',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入应用名',
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
  const currentClient = ref<SocialClientItem | null>(null)

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
      apiFn: fetchGetSocialClientList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '应用名', minWidth: 140, showOverflowTooltip: true },
        {
          prop: 'socialType',
          label: '渠道',
          width: 130,
          formatter: (row: SocialClientItem) =>
            h(
              ElTag,
              { type: 'success', effect: 'plain' },
              () => SOCIAL_TYPE_TEXT[row.socialType] ?? row.socialType
            )
        },
        { prop: 'clientId', label: 'ClientID（appid）', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: SocialClientItem) =>
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
          formatter: (row: SocialClientItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('member:social-client:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('member:social-client:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteClient(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Member.SocialClientSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: SocialClientItem): void => {
    dialogType.value = type
    currentClient.value = type === 'edit' ? (row ?? null) : null
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

  const deleteClient = async (row: SocialClientItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除配置「${row.name}」吗？`, '删除配置', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteSocialClient(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
