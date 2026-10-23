<!-- 门店菜品配置抽屉：按当前门店维护规格可售、沽清、门店价与生效区间 -->
<template>
  <ElDrawer
    v-model="drawerVisible"
    title="门店菜品配置"
    size="76%"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="flex h-full flex-col">
      <!-- 门店上下文由顶部切换器决定，未选择门店时不发起查询 -->
      <div
        v-if="!storeContext.currentStoreId"
        class="flex flex-1 flex-col items-center justify-center gap-3 rounded-lg bg-[var(--el-fill-color-lighter)] py-10 text-sm text-g-600"
      >
        <span>请先在顶部切换到具体门店，再维护该门店的菜品可售与门店价</span>
      </div>

      <template v-else>
        <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-medium">{{
              storeContext.currentStoreName || '当前门店'
            }}</span>
            <span class="text-xs text-g-600">门店价不填则沿用品牌价</span>
          </div>

          <ElSpace wrap>
            <ElInput
              v-model="searchForm.keyword"
              class="!w-56"
              placeholder="菜品名称或编码"
              clearable
              @keyup.enter="handleSearch()"
              @clear="handleSearch()"
            />
            <ElSelect
              v-model="searchForm.categoryId"
              class="!w-44"
              placeholder="全部分类"
              clearable
              filterable
              @change="handleSearch()"
            >
              <ElOption
                v-for="item in flatCategoryOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
            <ElButton :loading="loading" @click="handleSearch()">查询</ElButton>
          </ElSpace>
        </div>

        <div class="mb-3 flex flex-wrap items-center gap-2">
          <ElButton
            v-auth="'dish:store-sku:update'"
            size="small"
            :disabled="!selectedRows.length"
            @click="batchUpdate({ available: 1 })"
          >
            批量上架
          </ElButton>
          <ElButton
            v-auth="'dish:store-sku:update'"
            size="small"
            :disabled="!selectedRows.length"
            @click="batchUpdate({ available: 0 })"
          >
            批量下架
          </ElButton>
          <ElButton
            v-auth="'dish:store-sku:update'"
            size="small"
            :disabled="!selectedRows.length"
            @click="batchUpdate({ soldOut: 1 })"
          >
            批量沽清
          </ElButton>
          <ElButton
            v-auth="'dish:store-sku:update'"
            size="small"
            :disabled="!selectedRows.length"
            @click="batchUpdate({ soldOut: 0 })"
          >
            取消沽清
          </ElButton>
          <span class="text-xs text-g-600">已选 {{ selectedRows.length }} 条</span>
        </div>

        <div class="min-h-0 flex-1">
          <ArtTable
            :loading="loading"
            :data="data"
            :columns="columns"
            :pagination="pagination"
            :show-table-header="false"
            height="calc(100vh - 300px)"
            row-key="skuId"
            @selection-change="handleSelectionChange"
            @pagination:size-change="handleSizeChange"
            @pagination:current-change="handleCurrentChange"
          />
        </div>
      </template>
    </div>

    <StoreSkuDialog
      v-model:visible="dialogVisible"
      :sku-data="currentSku"
      @success="handleDialogSuccess"
    />
  </ElDrawer>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useStoreContextStore } from '@/store/modules/store-context'
  import { fetchBatchStoreSku, fetchGetStoreSkuList } from '@/api/dish/store-sku'
  import { ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import StoreSkuDialog from './store-sku-dialog.vue'

  defineOptions({ name: 'DishStoreSkuPanel' })

  type StoreSkuItem = Api.Dish.StoreSkuItem
  type SelectOption = { label: string; value: string }

  interface Props {
    visible: boolean
    categoryOptions?: Api.Dish.CategoryNode[]
  }

  const props = withDefaults(defineProps<Props>(), {
    categoryOptions: () => []
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const { hasAuth } = useAuth()
  const storeContext = useStoreContextStore()

  const selectedRows = ref<StoreSkuItem[]>([])
  const dialogVisible = ref(false)
  const currentSku = ref<StoreSkuItem | null>(null)

  const searchForm = reactive({
    keyword: '',
    categoryId: ''
  })

  // 分类下拉扁平化，用缩进宽度体现层级
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

  const flatCategoryOptions = computed(() =>
    flattenCategories(props.categoryOptions).map((item) => ({
      label: `${'　'.repeat(item.depth)}${item.label}`,
      value: item.value
    }))
  )

  // 分转元展示，0 表示沿用品牌价
  const formatYuan = (cent: number): string => (cent > 0 ? (cent / 100).toFixed(2) : '—')

  const {
    columns,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    refreshUpdate
  } = useTable({
    core: {
      apiFn: fetchGetStoreSkuList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        keyword: '',
        categoryId: ''
      },
      immediate: !!storeContext.currentStoreId,
      columnsFactory: () => [
        { type: 'selection', width: 48 },
        {
          prop: 'spuName',
          label: '菜品 / 规格',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: StoreSkuItem) =>
            h('div', { class: 'flex flex-col' }, [
              h('span', { class: 'font-medium' }, row.spuName),
              h('span', { class: 'text-xs text-g-600' }, row.skuName)
            ])
        },
        {
          prop: 'categoryName',
          label: '分类',
          minWidth: 120,
          showOverflowTooltip: true
        },
        {
          prop: 'standardPrice',
          label: '品牌价（元）',
          width: 130,
          formatter: (row: StoreSkuItem) => (row.standardPrice / 100).toFixed(2)
        },
        {
          prop: 'storePrice',
          label: '门店价（元）',
          width: 130,
          formatter: (row: StoreSkuItem) => formatYuan(row.storePrice)
        },
        {
          prop: 'storeMemberPrice',
          label: '门店会员价（元）',
          width: 150,
          formatter: (row: StoreSkuItem) => formatYuan(row.storeMemberPrice)
        },
        {
          prop: 'available',
          label: '可售',
          width: 100,
          formatter: (row: StoreSkuItem) =>
            h(
              ElTag,
              {
                type: row.configured && row.available === 1 ? 'success' : 'info',
                effect: 'plain',
                size: 'small'
              },
              () => (row.configured && row.available === 1 ? '可售' : '不可售')
            )
        },
        {
          prop: 'soldOut',
          label: '沽清',
          width: 90,
          formatter: (row: StoreSkuItem) =>
            h(
              ElTag,
              { type: row.soldOut === 1 ? 'warning' : 'info', effect: 'plain', size: 'small' },
              () => (row.soldOut === 1 ? '已沽清' : '正常')
            )
        },
        {
          prop: 'onlineOnly',
          label: '仅线上',
          width: 90,
          formatter: (row: StoreSkuItem) => (row.onlineOnly === 1 ? '是' : '否')
        },
        {
          prop: 'startAt',
          label: '生效区间',
          minWidth: 190,
          formatter: (row: StoreSkuItem) => {
            const start = row.startAt || ''
            const end = row.endAt || ''
            if (!start && !end) return '长期有效'
            return `${start || '立即'} ~ ${end || '长期'}`
          }
        },
        {
          prop: 'operation',
          label: '操作',
          width: 110,
          fixed: 'right',
          formatter: (row: StoreSkuItem) =>
            hasAuth('dish:store-sku:update')
              ? h(ArtButtonTable, {
                  type: 'edit',
                  text: '配置',
                  onClick: () => showDialog(row)
                })
              : h('span', { class: 'text-xs text-g-600' }, '—')
        }
      ]
    }
  })

  // 未配置过的规格也算一条记录，用于区分「未配置」与「显式下架」
  watch(
    () => storeContext.currentStoreId,
    () => {
      selectedRows.value = []
      if (drawerVisible.value && storeContext.currentStoreId) {
        refreshData()
      }
    }
  )

  // 抽屉打开时按当前门店加载，避免门店切换后仍展示旧数据
  watch(drawerVisible, (visible) => {
    if (!visible) return
    if (!storeContext.currentStoreId) return
    refreshData()
  })

  // 记录表格勾选行，供批量操作使用
  const handleSelectionChange = (rows: StoreSkuItem[]): void => {
    selectedRows.value = rows || []
  }

  // 应用筛选条件后重新查询：查询参数写回 useTable 内部状态再取数
  const handleSearch = (): void => {
    Object.assign(searchParams, {
      keyword: searchForm.keyword,
      categoryId: searchForm.categoryId
    })
    getData()
  }

  // 打开单规格配置弹窗
  const showDialog = (row: StoreSkuItem): void => {
    currentSku.value = row
    dialogVisible.value = true
  }

  // 保存成功后清空勾选并刷新列表
  const handleDialogSuccess = (): void => {
    selectedRows.value = []
    refreshUpdate()
  }

  // 批量调整门店开关：未勾选时提示，成功后刷新列表
  const batchUpdate = async (payload: {
    available?: number
    soldOut?: number
    onlineOnly?: number
  }): Promise<void> => {
    if (selectedRows.value.length === 0) {
      ElMessage.warning('请先勾选要调整的规格')
      return
    }

    try {
      await ElMessageBox.confirm(
        `确定要调整已选 ${selectedRows.value.length} 个规格吗？`,
        '批量调整门店菜品',
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
      await fetchBatchStoreSku({
        skuIds: selectedRows.value.map((row) => row.skuId),
        ...payload
      })
      ElMessage.success('调整成功')
      selectedRows.value = []
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
