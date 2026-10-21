<!-- 会员分组页面 -->
<template>
  <div class="member-group-page art-full-height">
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
            <ElButton v-auth="'member:group:create'" @click="showDialog('add')" v-ripple>
              新增分组
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

      <GroupDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :group-data="currentGroup"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteGroup, fetchGetGroupList } from '@/api/member/group'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import GroupDialog from './modules/group-dialog.vue'

  defineOptions({ name: 'MemberGroup' })

  type GroupItem = Api.Member.GroupItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Member.GroupSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '分组名',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入分组名',
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
  const currentGroup = ref<GroupItem | null>(null)

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
      apiFn: fetchGetGroupList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '分组名', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: GroupItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '禁用'
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: GroupItem) => row.remark || '-'
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: GroupItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('member:group:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('member:group:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteGroup(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Member.GroupSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: GroupItem): void => {
    dialogType.value = type
    currentGroup.value = type === 'edit' ? (row ?? null) : null
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

  const deleteGroup = async (row: GroupItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除分组「${row.name}」吗？`, '删除分组', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteGroup(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
