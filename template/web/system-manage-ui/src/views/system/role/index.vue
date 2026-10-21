<!-- 角色管理页面 -->
<template>
  <div class="role-page art-full-height">
    <!-- 搜索栏 -->
    <RoleSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></RoleSearch>

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'system:role:create'" @click="showDialog('add')" v-ripple>
              新增角色
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
      >
      </ArtTable>

      <!-- 新增 / 编辑 -->
      <RoleEditDialog
        v-model:visible="editVisible"
        :type="editType"
        :role-data="currentRole"
        @success="handleSaveSuccess"
      />

      <!-- 角色授权 -->
      <RolePermissionDialog
        v-model:visible="permissionVisible"
        :role="currentRole"
        @success="refreshUpdate"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteRole, fetchGetRoleList } from '@/api/system/role'
  import {
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElTag
  } from 'element-plus'
  import RoleSearch from './modules/role-search.vue'
  import RoleEditDialog from './modules/role-edit-dialog.vue'
  import RolePermissionDialog from './modules/role-permission-dialog.vue'

  defineOptions({ name: 'SystemRole' })

  type RoleItem = Api.SystemManage.RoleItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const editVisible = ref(false)
  const editType = ref<DialogType>('add')
  const currentRole = ref<RoleItem | null>(null)
  const permissionVisible = ref(false)

  const searchForm = ref<Api.SystemManage.RoleSearchParams>({
    name: undefined,
    code: undefined,
    status: undefined
  })

  // 角色状态：0 禁用 / 1 启用
  const STATUS_CONFIG: Record<number, { type: 'success' | 'info'; text: string }> = {
    0: { type: 'info', text: '禁用' },
    1: { type: 'success', text: '启用' }
  }

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
      apiFn: fetchGetRoleList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'selection' },
        { prop: 'name', label: '角色名称', minWidth: 140 },
        { prop: 'code', label: '角色标识', minWidth: 140 },
        {
          prop: 'type',
          label: '类型',
          width: 100,
          formatter: (row: RoleItem) =>
            h(ElTag, { type: row.type === 1 ? 'warning' : 'info' }, () =>
              row.type === 1 ? '内置角色' : '自定义'
            )
        },
        { prop: 'sort', label: '排序', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: RoleItem) => {
            const config = STATUS_CONFIG[row.status] ?? { type: 'info' as const, text: '未知' }
            return h(ElTag, { type: config.type }, () => config.text)
          }
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 160,
          formatter: (row: RoleItem) => row.remark || '-'
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 200,
          fixed: 'right',
          // 常用操作直接展示，其余收进「更多」下拉（hover 展开），并按权限过滤
          formatter: (row: RoleItem) => {
            const canUpdate = hasAuth('system:role:update')
            const canDelete = hasAuth('system:role:delete')
            const canAssignMenu = hasAuth('system:role:assign-menu')

            return h('div', { class: 'flex items-center' }, [
              ...(canUpdate
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(canDelete
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : []),
              ...(canAssignMenu
                ? [
                    h(
                      ElDropdown,
                      {
                        trigger: 'hover',
                        class: 'ml-3',
                        onCommand: () => showPermissionDialog(row)
                      },
                      {
                        default: () => h(ArtButtonTable, { type: 'more' }),
                        dropdown: () =>
                          h(ElDropdownMenu, null, () => [
                            h(ElDropdownItem, { command: 'permission' }, () => '授权')
                          ])
                      }
                    )
                  ]
                : [])
            ])
          }
        }
      ]
    }
  })

  const handleSearch = (params: Api.SystemManage.RoleSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: RoleItem): void => {
    editType.value = type
    currentRole.value = row || null
    nextTick(() => {
      editVisible.value = true
    })
  }

  const showPermissionDialog = (row: RoleItem): void => {
    currentRole.value = row
    nextTick(() => {
      permissionVisible.value = true
    })
  }

  const handleSaveSuccess = (): void => {
    if (editType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  /** 删除角色：内置角色与被引用角色由后端拦截 */
  const handleDelete = async (row: RoleItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除角色「${row.name}」吗？`, '删除角色', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteRole(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
