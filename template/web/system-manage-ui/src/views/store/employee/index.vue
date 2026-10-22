<!-- 门店员工管理页面：按当前门店上下文维护账号与门店的角色关联 -->
<template>
  <div class="store-employee-page art-full-height">
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
            <ElButton v-auth="'store:employee:create'" @click="showDialog('add')" v-ripple>
              新增员工
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

      <EmployeeDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :employee-data="currentEmployee"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteEmployee, fetchGetEmployeeList } from '@/api/store/employee'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import EmployeeDialog from './modules/employee-dialog.vue'

  defineOptions({ name: 'StoreEmployee' })

  type EmployeeItem = Api.Store.EmployeeItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const roleOptions = [
    { label: '店长', value: 1 },
    { label: '收银', value: 2 },
    { label: '后厨', value: 3 },
    { label: '服务员', value: 4 },
    { label: '骑手', value: 5 }
  ]

  const roleTagTypes: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    1: 'primary',
    2: 'success',
    3: 'warning',
    4: 'info',
    5: 'danger'
  }

  const searchForm = ref<Api.Store.EmployeeSearchParams>({
    keyword: '',
    role: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '账号 / 昵称 / 手机号',
        clearable: true
      }
    },
    {
      label: '门店角色',
      key: 'role',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: roleOptions
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      span: 6,
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
  const currentEmployee = ref<EmployeeItem | null>(null)

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
      apiFn: fetchGetEmployeeList,
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
          prop: 'username',
          label: '后台账号',
          minWidth: 180,
          formatter: (row: EmployeeItem) => formatUser(row)
        },
        { prop: 'mobile', label: '手机号', width: 130 },
        {
          prop: 'role',
          label: '门店角色',
          width: 110,
          formatter: (row: EmployeeItem) =>
            h(
              ElTag,
              { type: roleTagTypes[row.role] ?? 'info', effect: 'plain' },
              () =>
                row.roleName || roleOptions.find((item) => item.value === row.role)?.label || '未知'
            )
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: EmployeeItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '禁用'
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 160,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: EmployeeItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('store:employee:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('store:employee:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteEmployee(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 账号列优先展示昵称，没有昵称时回退到登录账号
  function formatUser(row: EmployeeItem): string {
    if (row.nickname && row.username) return `${row.nickname}（${row.username}）`
    return row.nickname || row.username || '-'
  }

  // 搜索条件变化后回到第一页重新加载员工列表
  const handleSearch = (params: Api.Store.EmployeeSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: EmployeeItem): void => {
    dialogType.value = type
    currentEmployee.value = type === 'edit' ? (row ?? null) : null
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

  // 删除员工关联，确认后调用接口并刷新列表
  const deleteEmployee = async (row: EmployeeItem): Promise<void> => {
    const userLabel = row.nickname || row.username || '该账号'
    try {
      await ElMessageBox.confirm(
        `确定要删除「${userLabel}」在「${row.storeName || '当前门店'}」的员工关联吗？`,
        '删除门店员工',
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
      await fetchDeleteEmployee(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
