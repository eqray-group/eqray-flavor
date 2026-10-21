<!-- 用户管理页面 -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'system:user:create'" @click="showDialog('add')" v-ripple>
              新增用户
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 用户弹窗 -->
      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @success="handleDialogSuccess"
      />

      <!-- 分配角色 -->
      <UserRoleDialog
        v-model:visible="roleDialogVisible"
        :user="currentRoleUser"
        @success="refreshUpdate"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteUser, fetchGetUserList, fetchResetUserPassword } from '@/api/system/user'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import UserRoleDialog from './modules/user-role-dialog.vue'
  import {
    ElAvatar,
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElTag
  } from 'element-plus'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import type { DialogType } from '@/types'

  defineOptions({ name: 'SystemUser' })

  type UserListItem = Api.SystemManage.UserListItem

  const { hasAuth } = useAuth()

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 分配角色
  const roleDialogVisible = ref(false)
  const currentRoleUser = ref<Partial<UserListItem> | null>(null)

  // 搜索表单（key 与后端查询参数一致）
  const searchForm = ref<Api.SystemManage.UserSearchParams>({
    username: undefined,
    nickname: undefined,
    mobile: undefined,
    status: undefined
  })

  // 性别与状态展示配置，取值与 system_users 表一致
  const SEX_TEXT: Record<number, string> = { 0: '未知', 1: '男', 2: '女' }
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
    refreshUpdate,
    refreshRemove
  } = useTable({
    // 核心配置
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'selection' }, // 勾选列
        {
          prop: 'avatar',
          label: '头像',
          width: 80,
          align: 'center',
          formatter: (row: UserListItem) =>
            h(ElAvatar, {
              size: 36,
              shape: 'circle',
              fit: 'cover',
              src: row.avatar || defaultAvatar
            })
        },
        {
          prop: 'username',
          label: '账号',
          minWidth: 180,
          formatter: (row: UserListItem) =>
            h('div', [
              h('p', row.username),
              h(
                'p',
                { style: 'color: var(--el-text-color-secondary); font-size: 12px' },
                row.nickname
              )
            ])
        },
        {
          prop: 'email',
          label: '邮箱',
          minWidth: 180,
          formatter: (row: UserListItem) => row.email || '-'
        },
        {
          prop: 'mobile',
          label: '手机号',
          width: 130,
          formatter: (row: UserListItem) => row.mobile || '-'
        },
        {
          prop: 'sex',
          label: '性别',
          width: 80,
          formatter: (row: UserListItem) => SEX_TEXT[row.sex] ?? '未知'
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: UserListItem) => {
            const config = STATUS_CONFIG[row.status] ?? { type: 'info' as const, text: '未知' }
            return h(ElTag, { type: config.type }, () => config.text)
          }
        },
        {
          prop: 'createdAt',
          label: '创建时间',
          width: 170
        },
        {
          prop: 'operation',
          label: '操作',
          width: 200,
          fixed: 'right', // 固定列
          // 常用操作直接展示，其余收进「更多」下拉（hover 展开），并按权限过滤
          formatter: (row: UserListItem) => {
            const canUpdate = hasAuth('system:user:update')
            const canDelete = hasAuth('system:user:delete')
            const canAssignRole = hasAuth('system:user:assign-role')

            const moreItems = [
              ...(canAssignRole ? [h(ElDropdownItem, { command: 'role' }, () => '分配角色')] : []),
              ...(canUpdate ? [h(ElDropdownItem, { command: 'reset' }, () => '重置密码')] : [])
            ]

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
                      onClick: () => deleteUser(row)
                    })
                  ]
                : []),
              ...(moreItems.length
                ? [
                    h(
                      ElDropdown,
                      {
                        trigger: 'hover',
                        class: 'ml-3',
                        onCommand: (command: string) => {
                          if (command === 'role') showRoleDialog(row)
                          if (command === 'reset') resetPassword(row)
                        }
                      },
                      {
                        default: () => h(ArtButtonTable, { type: 'more' }),
                        dropdown: () => h(ElDropdownMenu, null, () => moreItems)
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

  /**
   * 搜索处理
   * @param params 参数
   */
  const handleSearch = (params: Api.SystemManage.UserSearchParams) => {
    Object.assign(searchParams, params)
    // useTable 暴露的 getData 本身就是「回到第一页」的取数
    getData()
  }

  /**
   * 显示用户弹窗
   */
  const showDialog = (type: DialogType, row?: UserListItem): void => {
    dialogType.value = type
    currentUserData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 弹窗提交成功后刷新列表
   */
  /** 打开分配角色弹窗 */
  const showRoleDialog = (row: UserListItem): void => {
    currentRoleUser.value = row
    nextTick(() => {
      roleDialogVisible.value = true
    })
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  /**
   * 删除用户（后端为软删除）
   */
  const deleteUser = async (row: UserListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除用户「${row.username}」吗？`, '删除用户', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteUser(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  /**
   * 重置密码
   */
  const resetPassword = async (row: UserListItem): Promise<void> => {
    let password = ''
    try {
      const { value } = await ElMessageBox.prompt(
        `为用户「${row.username}」设置新密码`,
        '重置密码',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          inputType: 'password',
          inputPattern: /^.{6,72}$/,
          inputErrorMessage: '密码长度需在 6 ~ 72 个字符'
        }
      )
      password = value
    } catch {
      return
    }

    try {
      await fetchResetUserPassword(row.id, password)
      ElMessage.success('密码已重置')
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
