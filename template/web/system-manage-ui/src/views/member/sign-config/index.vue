<!-- 会员签到配置页面 -->
<template>
  <div class="member-sign-config-page art-full-height">
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
            <ElButton v-auth="'member:sign-config:create'" @click="showDialog('add')" v-ripple>
              新增档位
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

      <SignConfigDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :config-data="currentConfig"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteSignConfig, fetchGetSignConfigList } from '@/api/member/sign-config'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import SignConfigDialog from './modules/sign-config-dialog.vue'

  defineOptions({ name: 'MemberSignConfig' })

  type SignConfigItem = Api.Member.SignConfigItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Member.SignConfigSearchParams>({
    day: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '签到天数',
      key: 'day',
      type: 'number',
      span: 6,
      props: {
        min: 1,
        max: 31,
        controlsPosition: 'right',
        placeholder: '全部'
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
  const currentConfig = ref<SignConfigItem | null>(null)

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
      apiFn: fetchGetSignConfigList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'day',
          label: '签到天数',
          width: 120,
          formatter: (row: SignConfigItem) =>
            h(ElTag, { type: 'warning', effect: 'plain' }, () => `第 ${row.day} 天`)
        },
        { prop: 'point', label: '奖励积分', width: 120 },
        { prop: 'experience', label: '奖励经验', width: 120 },
        {
          prop: 'description',
          label: '说明',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: SignConfigItem) =>
            `连续签到第 ${row.day} 天发放 ${row.point} 积分、${row.experience} 经验`
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: SignConfigItem) =>
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
          formatter: (row: SignConfigItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('member:sign-config:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('member:sign-config:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteConfig(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Member.SignConfigSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: SignConfigItem): void => {
    dialogType.value = type
    currentConfig.value = type === 'edit' ? (row ?? null) : null
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

  const deleteConfig = async (row: SignConfigItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除「第 ${row.day} 天」的签到档位吗？`, '删除签到档位', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteSignConfig(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
