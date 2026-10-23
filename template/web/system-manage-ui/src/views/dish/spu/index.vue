<!-- 菜品管理页面：维护品牌级菜品与规格，并提供按门店配置可售与门店价的入口 -->
<template>
  <div class="dish-spu-page art-full-height">
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
            <ElButton v-auth="'dish:spu:create'" @click="showDialog('add')" v-ripple>
              新增菜品
            </ElButton>
            <ElButton v-auth="'dish:store-sku:list'" @click="showStorePanel"> 门店配置 </ElButton>
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

      <SpuDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :spu-data="currentSpu"
        :category-options="categoryOptions"
        :addon-options="addonOptions"
        @success="handleDialogSuccess"
      />

      <StoreSkuPanel v-model:visible="storePanelVisible" :category-options="categoryOptions" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteSpu, fetchGetSpuList, fetchUpdateSpuStatus } from '@/api/dish/spu'
  import { fetchGetCategoryTree } from '@/api/dish/category'
  import { fetchGetAddonGroupList } from '@/api/dish/addon'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import SpuDialog from './modules/spu-dialog.vue'
  import StoreSkuPanel from './modules/store-sku-panel.vue'

  defineOptions({ name: 'DishSpu' })

  type SpuItem = Api.Dish.SpuItem
  type DialogType = 'add' | 'edit'
  type SelectOption = { label: string; value: string }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Dish.SpuSearchParams>({
    keyword: '',
    categoryId: '',
    status: ''
  })

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentSpu = ref<SpuItem | null>(null)
  const storePanelVisible = ref(false)

  const categoryOptions = ref<Api.Dish.CategoryNode[]>([])
  const addonOptions = ref<SelectOption[]>([])

  // 把分类树扁平化成分组下拉项，用缩进宽度体现层级
  const flattenCategories = (
    nodes: Api.Dish.CategoryNode[],
    depth = 0
  ): Array<SelectOption & { depth: number }> => {
    return nodes.reduce<Array<SelectOption & { depth: number }>>((acc, node) => {
      acc.push({ label: node.name, value: node.id, depth })
      acc.push(...flattenCategories(node.children || [], depth + 1))
      return acc
    }, [])
  }

  const flatCategories = computed(() => flattenCategories(categoryOptions.value))

  // 搜索栏：分类下拉直接使用扁平化后的分类树
  const searchItems = computed(() => [
    {
      label: '菜品名称 / 编码',
      key: 'keyword',
      type: 'input',
      span: 7,
      props: {
        placeholder: '请输入菜品名称或编码',
        clearable: true
      }
    },
    {
      label: '所属分类',
      key: 'categoryId',
      type: 'select',
      span: 7,
      props: {
        placeholder: '全部分类',
        clearable: true,
        filterable: true,
        options: flatCategories.value.map((item) => ({
          label: `${'　'.repeat(item.depth)}${item.label}`,
          value: item.value
        }))
      }
    },
    {
      label: '上架状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '上架', value: 1 },
          { label: '下架', value: 0 }
        ]
      }
    }
  ])

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
      apiFn: fetchGetSpuList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'image',
          label: '主图',
          width: 90,
          formatter: (row: SpuItem) =>
            row.image
              ? h('img', {
                  src: row.image,
                  class: 'h-10 w-10 rounded object-cover',
                  alt: row.name
                })
              : h('span', { class: 'text-xs text-g-600' }, '—')
        },
        {
          prop: 'name',
          label: '菜品名称',
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'categoryName',
          label: '所属分类',
          minWidth: 130,
          showOverflowTooltip: true
        },
        {
          prop: 'skuCount',
          label: '规格数',
          width: 100
        },
        {
          prop: 'unit',
          label: '单位',
          width: 90
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: SpuItem) =>
            h(
              ElTag,
              { type: row.status === 1 ? 'success' : 'info', effect: 'plain', size: 'small' },
              () => (row.status === 1 ? '上架' : '下架')
            )
        },
        {
          prop: 'sort',
          label: '排序',
          width: 90
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 250,
          fixed: 'right',
          formatter: (row: SpuItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('dish:spu:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    }),
                    h(ArtButtonTable, {
                      type: 'action',
                      text: row.status === 1 ? '下架' : '上架',
                      icon: row.status === 1 ? 'ri:arrow-down-line' : 'ri:arrow-up-line',
                      onClick: () => toggleStatus(row)
                    })
                  ]
                : []),
              ...(hasAuth('dish:spu:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteSpu(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后回到第一页重新加载菜品列表
  const handleSearch = (params: Api.Dish.SpuSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 加载分类树与加料组下拉：菜品弹窗与门店配置抽屉共用
  const loadOptions = async (): Promise<void> => {
    try {
      const [categories, addons] = await Promise.all([
        fetchGetCategoryTree(),
        fetchGetAddonGroupList({ pageNo: 1, pageSize: 200 })
      ])
      categoryOptions.value = categories
      addonOptions.value = addons.list.map((item) => ({ label: item.name, value: item.id }))
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: SpuItem): void => {
    dialogType.value = type
    currentSpu.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 打开门店配置抽屉，门店上下文由顶部门店切换器决定
  const showStorePanel = (): void => {
    storePanelVisible.value = true
  }

  // 弹窗提交成功后按新增 / 编辑分别刷新列表
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 切换上下架：直接调用状态接口并刷新当前页
  const toggleStatus = async (row: SpuItem): Promise<void> => {
    const next = row.status === 1 ? 0 : 1
    try {
      await fetchUpdateSpuStatus(row.id, next)
      ElMessage.success(next === 1 ? '已上架' : '已下架')
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 删除菜品：规格与门店配置一并清理，确认后调用接口
  const deleteSpu = async (row: SpuItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除菜品「${row.name}」吗？其规格与各门店配置会一并清除。`,
        '删除菜品',
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
      await fetchDeleteSpu(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(() => {
    loadOptions()
  })
</script>
