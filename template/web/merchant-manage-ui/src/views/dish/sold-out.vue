<!-- 沽清：门店当天卖完的菜一键售罄，顾客端立即置灰不可点 -->
<template>
  <div class="page">
    <div class="toolbar">
      <ElSelect v-model="categoryId" placeholder="全部分类" clearable class="w-44" @change="reload">
        <ElOption
          v-for="item in categoryOptions"
          :key="item.id"
          :label="item.name"
          :value="item.id"
        />
      </ElSelect>
      <ElInput
        v-model="keyword"
        placeholder="搜索菜品名称 / 编码"
        clearable
        class="w-64"
        @keyup.enter="reload"
        @clear="reload"
      />
      <ElButton :loading="loading" @click="reload">刷新</ElButton>

      <div class="bulk">
        <span v-if="selected.length" class="hint">已选 {{ selected.length }} 项</span>
        <ElButton
          :disabled="!canEdit || !selected.length || batchLoading"
          :loading="batchLoading"
          type="warning"
          plain
          @click="batchToggle(1)"
        >
          批量沽清
        </ElButton>
        <ElButton
          :disabled="!canEdit || !selected.length || batchLoading"
          :loading="batchLoading"
          type="success"
          plain
          @click="batchToggle(0)"
        >
          批量恢复
        </ElButton>
      </div>
    </div>

    <ElTable
      v-loading="loading"
      :data="list"
      border
      size="small"
      row-key="skuId"
      @selection-change="handleSelectionChange"
    >
      <ElTableColumn type="selection" width="46" reserve-selection />
      <ElTableColumn label="菜品" min-width="200">
        <template #default="{ row }">
          <div class="dish-name">{{ row.spuName }}</div>
          <div v-if="row.skuName" class="dish-sku">{{ row.skuName }}</div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="categoryName" label="分类" width="110" show-overflow-tooltip />
      <ElTableColumn label="售价" width="100">
        <template #default="{ row }">
          ¥{{ yuan(row.storePrice > 0 ? row.storePrice : row.standardPrice) }}
        </template>
      </ElTableColumn>
      <ElTableColumn label="可售" width="90">
        <template #default="{ row }">
          <ElTag :type="row.available === 1 ? 'success' : 'info'" effect="plain" size="small">
            {{ row.available === 1 ? '可售' : '下架' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="沽清" width="120">
        <template #default="{ row }">
          <ElSwitch
            :model-value="row.soldOut === 1"
            :loading="pending.has(row.skuId)"
            :disabled="!canEdit"
            inline-prompt
            active-text="已沽清"
            inactive-text="正常"
            @update:model-value="(value: string | number | boolean) => toggle(row, value ? 1 : 0)"
          />
        </template>
      </ElTableColumn>
      <ElTableColumn prop="remark" label="备注" min-width="120" show-overflow-tooltip />
    </ElTable>

    <ElPagination
      v-model:current-page="pageNo"
      class="pager"
      layout="prev, pager, next, total"
      :page-size="pageSize"
      :total="total"
      @current-change="loadData"
    />
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { computed, onMounted, reactive, ref } from 'vue'

  import { fetchCategoryTree, fetchSaveStoreSku, fetchStoreSkuPage } from '@/api/dish'
  import { useAuthStore } from '@/stores/auth'

  const auth = useAuthStore()

  // 沽清本质是改门店菜品配置，权限点沿用管理后台的同一个
  const canEdit = computed(() => auth.hasAuth('dish:store-sku:update'))

  const list = ref<Api.Dish.StoreSkuItem[]>([])
  const loading = ref(false)
  const batchLoading = ref(false)
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = 50

  const keyword = ref('')
  const categoryId = ref<string>('')
  const categoryOptions = ref<Api.Dish.CategoryNode[]>([])
  const selected = ref<Api.Dish.StoreSkuItem[]>([])

  // 正在切换的行，用于开关自身的 loading，避免连点
  const pending = reactive(new Set<string>())

  const yuan = (cent: number): string => ((cent || 0) / 100).toFixed(2)

  // 分类树拍平成下拉项：门店端只用来筛，不需要层级缩进
  const flatten = (nodes: Api.Dish.CategoryNode[], out: Api.Dish.CategoryNode[] = []) => {
    for (const node of nodes) {
      out.push(node)
      if (node.children?.length) {
        flatten(node.children, out)
      }
    }

    return out
  }

  const loadCategories = async (): Promise<void> => {
    try {
      const tree = await fetchCategoryTree()
      categoryOptions.value = flatten(tree ?? [])
    } catch {
      // 错误提示由请求层统一弹出，分类为空不影响主列表
    }
  }

  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      const params: Record<string, unknown> = { pageNo: pageNo.value, pageSize }
      if (keyword.value.trim()) {
        params.keyword = keyword.value.trim()
      }
      if (categoryId.value) {
        params.categoryId = categoryId.value
      }
      const res = await fetchStoreSkuPage(params)
      list.value = res.list ?? []
      total.value = res.total ?? 0
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  const reload = (): void => {
    pageNo.value = 1
    void loadData()
  }

  const handleSelectionChange = (rows: Api.Dish.StoreSkuItem[]): void => {
    selected.value = rows
  }

  /**
   * 保存沽清状态。
   * 只改 soldOut，其余字段原样回传：后端的保存是整行覆盖，漏字段会把门店价冲掉。
   */
  const saveSoldOut = (row: Api.Dish.StoreSkuItem, soldOut: number): Promise<unknown> =>
    fetchSaveStoreSku(row.skuId, {
      available: row.available,
      soldOut,
      onlineOnly: row.onlineOnly,
      sort: row.sort,
      remark: row.remark,
      storePrice: row.storePrice,
      storeMemberPrice: row.storeMemberPrice,
      startAt: row.startAt ?? '',
      endAt: row.endAt ?? ''
    })

  // 单条切换：先乐观更新，失败回滚，避免整页刷新打断操作
  const toggle = async (row: Api.Dish.StoreSkuItem, soldOut: number): Promise<void> => {
    const previous = row.soldOut
    pending.add(row.skuId)
    row.soldOut = soldOut
    try {
      await saveSoldOut(row, soldOut)
      ElMessage.success(soldOut === 1 ? `${row.spuName} 已沽清` : `${row.spuName} 已恢复可售`)
    } catch {
      row.soldOut = previous
    } finally {
      pending.delete(row.skuId)
    }
  }

  // 批量切换：逐条保存，结果按成功数汇总，失败的保持原状态
  const batchToggle = async (soldOut: number): Promise<void> => {
    const rows = selected.value
    if (!rows.length) return

    try {
      await ElMessageBox.confirm(
        `确定把选中的 ${rows.length} 个菜品${soldOut === 1 ? '全部沽清' : '全部恢复可售'}吗？`,
        soldOut === 1 ? '批量沽清' : '批量恢复',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    batchLoading.value = true
    let success = 0
    let failed = 0
    await Promise.all(
      rows.map(async (row) => {
        try {
          await saveSoldOut(row, soldOut)
          row.soldOut = soldOut
          success++
        } catch {
          failed++
        }
      })
    )
    batchLoading.value = false

    if (failed === 0) {
      ElMessage.success(`已处理 ${success} 个菜品`)
    } else {
      ElMessage.warning(`成功 ${success} 个，失败 ${failed} 个，请重试失败项`)
    }
    void loadData()
  }

  onMounted(async () => {
    await loadCategories()
    await loadData()
  })
</script>

<style scoped>
  .w-44 {
    width: 180px;
  }

  .w-64 {
    width: 240px;
  }

  .bulk {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-left: auto;
  }

  .hint {
    font-size: 13px;
    color: #6b7280;
  }

  .dish-name {
    font-size: 14px;
    font-weight: 500;
  }

  .dish-sku {
    margin-top: 2px;
    font-size: 12px;
    color: #6b7280;
  }

  .pager {
    justify-content: flex-end;
    margin-top: 12px;
  }
</style>
