<!-- 菜品分类管理页面：品牌级分类树，一次加载整棵树，关键字与状态在前端过滤 -->
<template>
  <div class="dish-category-page art-full-height">
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
            <ElButton v-auth="'dish:category:create'" @click="showDialog('add')" v-ripple>
              新增分类
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="displayTree"
        :columns="columns"
        row-key="id"
        :tree-props="{ children: 'children' }"
        :empty-text="hasFilter ? '没有匹配的分类' : '还没有分类，先新增一个根分类'"
      />

      <CategoryDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :category-data="currentCategory"
        :parent-id="presetParentId"
        :tree="data"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteCategory, fetchGetCategoryTree } from '@/api/dish/category'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import CategoryDialog from './modules/category-dialog.vue'

  defineOptions({ name: 'DishCategory' })

  type CategoryNode = Api.Dish.CategoryNode
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Dish.CategorySearchParams>({
    keyword: '',
    status: ''
  })

  const searchItems = [
    {
      label: '分类名称',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '请输入分类名称',
        clearable: true
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
  const currentCategory = ref<CategoryNode | null>(null)
  // 从行内「新增子分类」打开弹窗时预置的上级分类
  const presetParentId = ref('0')

  const {
    columns,
    columnChecks,
    data,
    loading,
    resetSearchParams,
    refreshData,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      // 分类树一次返回全量数据，因此不传查询参数，关键字与状态由前端过滤
      apiFn: fetchGetCategoryTree,
      columnsFactory: () => [
        {
          prop: 'name',
          label: '分类名称',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: CategoryNode) => row.name
        },
        {
          prop: 'image',
          label: '图标',
          width: 90,
          formatter: (row: CategoryNode) =>
            row.image
              ? h('img', {
                  src: row.image,
                  class: 'h-8 w-8 rounded object-cover',
                  alt: row.name
                })
              : h('span', { class: 'text-xs text-g-600' }, '—')
        },
        {
          prop: 'spuCount',
          label: '菜品数',
          width: 100
        },
        {
          prop: 'sort',
          label: '排序',
          width: 90
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: CategoryNode) =>
            h(
              ElTag,
              { type: row.status === 1 ? 'success' : 'info', effect: 'plain', size: 'small' },
              () => (row.status === 1 ? '启用' : '禁用')
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 180,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 230,
          fixed: 'right',
          formatter: (row: CategoryNode) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('dish:category:create')
                ? [
                    h(ArtButtonTable, {
                      type: 'add',
                      text: '子分类',
                      onClick: () => showDialog('add', null, row.id)
                    })
                  ]
                : []),
              ...(hasAuth('dish:category:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('dish:category:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteCategory(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 是否处于筛选状态，用于区分空态文案
  const hasFilter = computed(() => !!searchForm.value.keyword || searchForm.value.status !== '')

  // 递归过滤分类树：命中节点保留，未命中但子孙命中的节点也保留
  const filterNodes = (nodes: CategoryNode[]): CategoryNode[] => {
    const keyword = (searchForm.value.keyword || '').trim().toLowerCase()
    const status = searchForm.value.status

    return nodes.reduce<CategoryNode[]>((acc, node) => {
      const children = filterNodes(node.children || [])
      const matchedKeyword = !keyword || node.name.toLowerCase().includes(keyword)
      const matchedStatus = status === '' || status === undefined || node.status === status

      if ((matchedKeyword && matchedStatus) || children.length > 0) {
        acc.push({ ...node, children })
      }
      return acc
    }, [])
  }

  // 表格实际渲染的数据：带筛选条件时用过滤后的树，否则直接用接口结果
  const displayTree = computed<CategoryNode[]>(() => {
    const list = (data.value || []) as unknown as CategoryNode[]
    return hasFilter.value ? filterNodes(list) : list
  })

  // 搜索条件变化后同步表单并重新计算过滤结果，分类树不分页因此无需回第一页
  const handleSearch = (params: Api.Dish.CategorySearchParams) => {
    Object.assign(searchForm.value, params)
  }

  // 打开新增 / 编辑弹窗；新增子分类时由 parentId 指定上级
  const showDialog = (type: DialogType, row?: CategoryNode | null, parentId = '0'): void => {
    dialogType.value = type
    currentCategory.value = type === 'edit' ? (row ?? null) : null
    presetParentId.value = type === 'edit' ? (row?.parentId ?? '0') : parentId
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 弹窗提交成功后按新增 / 编辑分别刷新分类树
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 删除分类：仍有子分类或菜品时由后端拒绝，这里只做二次确认
  const deleteCategory = async (row: CategoryNode): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除分类「${row.name}」吗？该操作不可恢复。`,
        '删除菜品分类',
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
      await fetchDeleteCategory(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
