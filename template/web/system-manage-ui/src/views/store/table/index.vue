<!-- 桌台管理页面：以台位图按区域展示桌台，后台只维护桌台档案，开台 / 清台属前厅业务 -->
<template>
  <div class="store-table-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <!-- 工具条：左侧状态图例，右侧维护入口 -->
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-4">
          <span class="font-medium">台位图</span>
          <div class="flex items-center gap-3 text-xs text-g-600">
            <span
              v-for="option in statusOptions"
              :key="option.value"
              class="flex items-center gap-1"
            >
              <span class="status-dot" :class="`is-dot-${option.value}`"></span>
              {{ option.label }}
            </span>
          </div>
        </div>

        <ElSpace wrap>
          <ElButton v-auth="'store:table:create'" type="primary" @click="showTableDialog('add')">
            新增桌台
          </ElButton>
          <ElButton v-auth="'store:table:create'" @click="showBatchDialog">批量生成</ElButton>
          <ElButton :loading="loading" @click="refreshData">刷新</ElButton>
        </ElSpace>
      </div>

      <div class="min-h-0 flex-1 overflow-auto">
        <div v-if="loading && data.length === 0" class="py-10 text-center text-sm text-g-600">
          加载中…
        </div>

        <!-- 空态：浅色底块内一行文案 + 主按钮 -->
        <div
          v-else-if="data.length === 0"
          class="flex flex-col items-center gap-3 rounded-lg bg-[var(--el-fill-color-lighter)] py-10 text-sm text-g-600"
        >
          <span>当前范围内还没有桌台，可单个新增或按前缀批量生成</span>
          <ElButton
            v-auth="'store:table:create'"
            type="primary"
            plain
            @click="showTableDialog('add')"
          >
            新增桌台
          </ElButton>
        </div>

        <div v-else class="space-y-6">
          <section v-for="group in areaGroups" :key="group.key">
            <div class="mb-3 flex items-center gap-2">
              <span class="text-sm font-medium">{{ group.name }}</span>
              <span class="text-xs text-g-600">{{ group.items.length }} 张</span>
            </div>

            <div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
              <div
                v-for="item in group.items"
                :key="item.id"
                class="table-card"
                :class="statusCardClass(item)"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="min-w-0">
                    <div class="truncate text-lg font-medium leading-tight">{{ item.name }}</div>
                    <div class="mt-1 truncate text-xs text-g-600">{{ item.code }}</div>
                  </div>
                  <ElTag :type="statusTagTypes[item.status] ?? 'info'" effect="plain" size="small">
                    {{ statusLabel(item) }}
                  </ElTag>
                </div>

                <div class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-g-600">
                  <span>{{ item.seats }} 座</span>
                  <span v-if="!storeContext.currentStoreId" class="truncate">
                    {{ item.storeName }}
                  </span>
                  <span v-if="item.currentOrderId !== '0'">台位单 {{ item.currentOrderId }}</span>
                  <span v-if="item.openedAt">{{ item.openedAt }}</span>
                </div>

                <div class="mt-3 flex flex-wrap justify-end gap-2">
                  <ElTooltip
                    v-if="hasAuth('store:table:update')"
                    content="编辑桌台"
                    placement="top"
                    :show-after="300"
                  >
                    <ElButton
                      class="card-action"
                      size="small"
                      circle
                      @click.stop="showTableDialog('edit', item)"
                    >
                      <ArtSvgIcon icon="ri:pencil-line" />
                    </ElButton>
                  </ElTooltip>

                  <ElTooltip
                    v-if="
                      hasAuth('store:table:delete') &&
                      item.status === 1 &&
                      item.currentOrderId === '0'
                    "
                    content="删除桌台"
                    placement="top"
                    :show-after="300"
                  >
                    <ElButton
                      class="card-action is-danger"
                      size="small"
                      circle
                      @click.stop="deleteTable(item)"
                    >
                      <ArtSvgIcon icon="ri:delete-bin-5-line" />
                    </ElButton>
                  </ElTooltip>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>

      <div v-if="pagination.total > 0" class="mt-4 flex items-center justify-between">
        <span class="text-xs text-g-600">共 {{ pagination.total }} 张桌台</span>
        <ElPagination
          :current-page="pagination.current"
          :page-size="pagination.size"
          :page-sizes="[20, 50, 100]"
          :total="pagination.total"
          layout="sizes, prev, pager, next"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>

      <TableDialog
        v-model:visible="tableDialogVisible"
        :type="tableDialogType"
        :table-data="currentTable"
        @success="handleTableDialogSuccess"
      />
      <TableBatchDialog v-model:visible="batchDialogVisible" @success="handleBatchSuccess" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteTable, fetchGetTableAreaOptions, fetchGetTableList } from '@/api/store/table'
  import { ElButton, ElMessage, ElMessageBox, ElPagination, ElSpace, ElTag } from 'element-plus'
  import { useStoreContextStore } from '@/store/modules/store-context'
  import TableBatchDialog from './modules/table-batch-dialog.vue'
  import TableDialog from './modules/table-dialog.vue'

  defineOptions({ name: 'StoreTable' })

  type TableItem = Api.Store.TableItem
  type TableAreaOption = Api.Store.TableAreaOption
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()
  const storeContext = useStoreContextStore()

  // 桌台状态枚举，与后端 table_status 字典一致
  const statusOptions = [
    { label: '空闲', value: 1 },
    { label: '使用中', value: 2 },
    { label: '待清台', value: 3 }
  ]

  // 状态对应的标签色：空闲灰、使用中绿、待清台黄
  const statusTagTypes: Record<number, 'success' | 'info' | 'warning'> = {
    1: 'info',
    2: 'success',
    3: 'warning'
  }

  // 状态对应的台位卡样式类，用左侧色条区分占用情况
  const statusCardClasses: Record<number, string> = {
    1: 'is-free',
    2: 'is-dining',
    3: 'is-cleaning'
  }

  const searchForm = ref<Api.Store.TableSearchParams>({
    keyword: '',
    areaId: undefined,
    status: undefined
  })

  const areaOptions = ref<TableAreaOption[]>([])

  // 全部门店视角下区域可能重名，下拉展示带上门店名称
  const displayArea = (area: TableAreaOption): string => {
    if (!storeContext.currentStoreId && area.storeName) {
      return `${area.storeName} / ${area.name}`
    }
    return area.name
  }

  const searchItems = computed(() => [
    {
      label: '桌号 / 桌码',
      key: 'keyword',
      type: 'input',
      span: 7,
      props: {
        placeholder: '请输入桌号或桌码',
        clearable: true
      }
    },
    {
      label: '所属区域',
      key: 'areaId',
      type: 'select',
      span: 7,
      props: {
        placeholder: '全部',
        clearable: true,
        filterable: true,
        options: areaOptions.value.map((area) => ({
          label: displayArea(area),
          value: area.id
        }))
      }
    },
    {
      label: '桌台状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    }
  ])

  const tableDialogVisible = ref(false)
  const tableDialogType = ref<DialogType>('add')
  const currentTable = ref<TableItem | null>(null)
  const batchDialogVisible = ref(false)

  // 分页查询桌台：useTable 按 records 结构推导行类型，后端返回 { list, total }，
  // 这里只做类型对齐，运行期结构与请求层解包逻辑保持不变。
  const fetchTablePage = async (
    params: Api.Store.TableSearchParams
  ): Promise<Api.Common.PaginatedResponse<TableItem>> =>
    (await fetchGetTableList(params)) as unknown as Api.Common.PaginatedResponse<TableItem>

  const {
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
      apiFn: fetchTablePage,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      }
    }
  })

  // 台位图按区域分组渲染：顺序沿用接口返回，未关联区域时归到「未分区」
  const areaGroups = computed(() => {
    const groups = new Map<string, { key: string; name: string; items: TableItem[] }>()
    for (const item of data.value) {
      const key = item.areaId || '0'
      const existed = groups.get(key)
      if (existed) {
        existed.items.push(item)
        continue
      }
      groups.set(key, { key, name: item.areaName || '未分区', items: [item] })
    }

    return Array.from(groups.values())
  })

  // 台位卡样式：按状态给出左侧色条与浅色底
  const statusCardClass = (item: TableItem): string => statusCardClasses[item.status] ?? ''

  // 状态文案：优先后端下发的 statusName，缺失时回落到本地枚举
  const statusLabel = (item: TableItem): string =>
    item.statusName || statusOptions.find((option) => option.value === item.status)?.label || '未知'

  // 加载当前门店或可见范围内的区域下拉选项
  const loadAreaOptions = async (): Promise<void> => {
    try {
      areaOptions.value = await fetchGetTableAreaOptions()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(loadAreaOptions)

  // 搜索条件变化后回到第一页重新加载桌台列表
  const handleSearch = (params: Api.Store.TableSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑桌台弹窗，编辑时携带当前行用于回填
  const showTableDialog = (type: DialogType, row?: TableItem): void => {
    tableDialogType.value = type
    currentTable.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      tableDialogVisible.value = true
    })
  }

  // 打开批量生成弹窗
  const showBatchDialog = (): void => {
    batchDialogVisible.value = true
  }

  // 桌台弹窗提交成功后按新增 / 编辑分别刷新列表
  const handleTableDialogSuccess = (): void => {
    if (tableDialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 批量生成成功后刷新列表与区域选项
  const handleBatchSuccess = (): void => {
    refreshCreate()
    loadAreaOptions()
  }

  // 删除空闲桌台，确认后调用接口并刷新列表
  const deleteTable = async (row: TableItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除桌台「${row.name}」吗？该操作不可恢复。`, '删除桌台', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteTable(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>

<style scoped>
  /* 卡片内容改为纵向弹性布局：台位图区域独立滚动，工具条与分页固定 */
  .store-table-page :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  /* 台位卡：左侧色条表示占用状态，hover 轻微上浮 */
  .table-card {
    padding: 14px 16px;
    cursor: default;
    background-color: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-left: 3px solid var(--el-border-color);
    border-radius: 10px;
    transition:
      border-color 0.2s,
      box-shadow 0.2s,
      transform 0.2s;
  }

  .table-card:hover {
    box-shadow: var(--el-box-shadow-light);
    transform: translateY(-1px);
  }

  /* 空闲：绿色色条 */
  .table-card.is-free {
    border-left-color: var(--el-color-success);
  }

  /* 使用中：橙色色条 + 浅橙底，一眼能看出占用 */
  .table-card.is-dining {
    background-color: var(--el-color-warning-light-9);
    border-left-color: var(--el-color-warning);
  }

  /* 待清台：灰色色条 + 浅灰底 */
  .table-card.is-cleaning {
    background-color: var(--el-color-info-light-9);
    border-left-color: var(--el-color-info);
  }

  /* 卡片操作按钮：默认中性描边，hover 才显色，避免按钮比桌号更抢眼 */
  .card-action {
    width: 28px;
    height: 28px;
    padding: 0;
    color: var(--el-text-color-secondary);
    background-color: transparent;
    border: 1px solid var(--el-border-color);
  }

  .card-action:hover {
    color: var(--el-color-primary);
    background-color: var(--el-color-primary-light-9);
    border-color: var(--el-color-primary-light-5);
  }

  .card-action.is-danger:hover {
    color: var(--el-color-danger);
    background-color: var(--el-color-danger-light-9);
    border-color: var(--el-color-danger-light-5);
  }

  /* 状态图例圆点：与台位卡色条保持一致 */
  .status-dot {
    width: 8px;
    height: 8px;
    background-color: var(--el-color-info);
    border-radius: 50%;
  }

  .status-dot.is-dot-1 {
    background-color: var(--el-color-success);
  }

  .status-dot.is-dot-2 {
    background-color: var(--el-color-warning);
  }

  .status-dot.is-dot-3 {
    background-color: var(--el-color-info);
  }
</style>
