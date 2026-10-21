<!-- 菜单管理页面 -->
<template>
  <div class="menu-page art-full-height">
    <!-- 搜索栏 -->
    <ArtSearchBar
      v-model="formFilters"
      :items="formItems"
      :show-expand="false"
      @reset="handleReset"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader
        v-model:columns="columnChecks"
        :show-zebra="false"
        :loading="loading"
        @refresh="getMenuList"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton
              v-auth="'system:menu:create'"
              type="primary"
              @click="handleAddMenu()"
              v-ripple
            >
              新增菜单
            </ElButton>
            <ElButton @click="toggleExpand" v-ripple>
              {{ isExpanded ? '收起' : '展开' }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data="filteredData"
        :stripe="false"
        :tree-props="{ children: 'children' }"
        :default-expand-all="false"
      />

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :menu-data="editData"
        :parent-id="presetParentId"
        :menu-tree="tableData"
        @success="getMenuList"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteMenu, fetchGetMenuTree } from '@/api/system/menu'
  import {
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElTag
  } from 'element-plus'
  import MenuDialog from './modules/menu-dialog.vue'

  defineOptions({ name: 'SystemMenu' })

  type MenuItem = Api.SystemManage.MenuItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  // 列表与弹窗状态
  const loading = ref(false)
  const tableData = ref<MenuItem[]>([])
  const isExpanded = ref(false)
  const tableRef = ref()

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const editData = ref<MenuItem | null>(null)
  const presetParentId = ref('')

  // 搜索条件：名称与路由，前端过滤（菜单总量很小，不需要分页）
  const initialSearchState = { name: '', path: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const formItems = computed(() => [
    { label: '菜单名称', key: 'name', type: 'input', props: { clearable: true } },
    { label: '路由地址', key: 'path', type: 'input', props: { clearable: true } }
  ])

  // 1 目录 / 2 菜单 / 3 按钮
  const TYPE_CONFIG: Record<
    number,
    { type: 'primary' | 'success' | 'info' | 'danger'; text: string }
  > = {
    1: { type: 'info', text: '目录' },
    2: { type: 'primary', text: '菜单' },
    3: { type: 'danger', text: '按钮' }
  }

  onMounted(() => {
    getMenuList()
  })

  /** 获取菜单树 */
  const getMenuList = async (): Promise<void> => {
    loading.value = true
    try {
      tableData.value = await fetchGetMenuTree()
      if (isExpanded.value) {
        nextTick(() => toggleRows(true))
      }
    } finally {
      loading.value = false
    }
  }

  // 表格列配置
  const { columnChecks, columns } = useTableColumns(() => [
    {
      prop: 'name',
      label: '菜单名称'
    },
    {
      prop: 'type',
      label: '类型',
      formatter: (row: MenuItem) => {
        const config = TYPE_CONFIG[row.type] ?? { type: 'info' as const, text: '未知' }
        return h(ElTag, { type: config.type }, () => config.text)
      }
    },
    {
      prop: 'path',
      label: '路由 / 权限标识',
      formatter: (row: MenuItem) => (row.type === 3 ? row.permission : row.path) || '-'
    },
    {
      prop: 'component',
      label: '组件路径',
      formatter: (row: MenuItem) => row.component || '-'
    },
    {
      prop: 'permission',
      label: '权限标识',
      formatter: (row: MenuItem) => row.permission || '-'
    },
    {
      prop: 'sort',
      label: '排序'
    },
    {
      prop: 'status',
      label: '状态',
      formatter: (row: MenuItem) =>
        h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
          row.status === 1 ? '启用' : '禁用'
        )
    },
    {
      prop: 'createdAt',
      label: '创建时间'
    },
    {
      prop: 'operation',
      label: '操作',
      width: 200,
      align: 'right',
      formatter: (row: MenuItem) => {
        // 常用操作直接展示，其余收进「更多」下拉（hover 展开），并按权限过滤
        const buttons = [
          ...(hasAuth('system:menu:update')
            ? [
                h(ArtButtonTable, {
                  type: 'edit',
                  text: '编辑',
                  onClick: () => handleEditMenu(row)
                })
              ]
            : []),
          ...(hasAuth('system:menu:delete')
            ? [
                h(ArtButtonTable, {
                  type: 'delete',
                  text: '删除',
                  onClick: () => handleDeleteMenu(row)
                })
              ]
            : [])
        ]

        // 只有目录 / 菜单可以挂子节点
        if (hasAuth('system:menu:create') && row.type !== 3) {
          buttons.push(
            h(
              ElDropdown,
              {
                trigger: 'hover',
                class: 'ml-3',
                onCommand: () => handleAddMenu(row)
              },
              {
                default: () => h(ArtButtonTable, { type: 'more' }),
                dropdown: () =>
                  h(ElDropdownMenu, null, () => [
                    h(ElDropdownItem, { command: 'add' }, () => '新增子项')
                  ])
              }
            )
          )
        }

        return h('div', { class: 'flex items-center justify-end' }, buttons)
      }
    }
  ])

  /** 按名称 / 路由递归过滤，命中子节点时保留父级 */
  const filterTree = (nodes: MenuItem[]): MenuItem[] => {
    const name = appliedFilters.name.trim().toLowerCase()
    const path = appliedFilters.path.trim().toLowerCase()
    if (!name && !path) return nodes

    const result: MenuItem[] = []
    for (const node of nodes) {
      const children = filterTree(node.children || [])
      const nameMatch = !name || node.name.toLowerCase().includes(name)
      const pathMatch = !path || node.path.toLowerCase().includes(path)
      if ((nameMatch && pathMatch) || children.length > 0) {
        result.push({ ...node, children })
      }
    }

    return result
  }

  const filteredData = computed(() => filterTree(tableData.value))

  /** 重置搜索条件 */
  const handleReset = (): void => {
    Object.assign(formFilters, { ...initialSearchState })
    Object.assign(appliedFilters, { ...initialSearchState })
  }

  /** 执行搜索 */
  const handleSearch = (): void => {
    Object.assign(appliedFilters, { ...formFilters })
  }

  /**
   * 打开新增弹窗
   * @param parent 传入时表示在该节点下新增子项
   */
  const handleAddMenu = (parent?: MenuItem): void => {
    dialogType.value = 'add'
    editData.value = null
    presetParentId.value = parent?.id ?? ''
    dialogVisible.value = true
  }

  /** 打开编辑弹窗 */
  const handleEditMenu = (row: MenuItem): void => {
    dialogType.value = 'edit'
    editData.value = row
    presetParentId.value = ''
    dialogVisible.value = true
  }

  /** 删除菜单（后端会校验是否存在子节点） */
  const handleDeleteMenu = async (row: MenuItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除「${row.name}」吗？删除后不可恢复`, '删除菜单', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteMenu(row.id)
      ElMessage.success('删除成功')
      getMenuList()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  /** 展开 / 收起全部节点 */
  const toggleExpand = (): void => {
    isExpanded.value = !isExpanded.value
    toggleRows(isExpanded.value)
  }

  const toggleRows = (expanded: boolean): void => {
    if (!tableRef.value?.elTableRef) return

    const processRows = (rows: MenuItem[]) => {
      rows.forEach((row) => {
        if (row.children?.length) {
          tableRef.value.elTableRef.toggleRowExpansion(row, expanded)
          processRows(row.children)
        }
      })
    }
    processRows(filteredData.value)
  }
</script>
