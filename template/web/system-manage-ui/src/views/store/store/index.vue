<!-- 门店管理：左侧组织树 + 右侧门店内容，总部 / 大区负责维护组织与下级门店 -->
<template>
  <div class="store-page art-full-height">
    <div class="store-layout">
      <!-- 左侧：组织架构树；总部全库唯一，只有尚未创建时才出现创建入口 -->
      <ElCard class="art-table-card store-tree-card">
        <div class="mb-3 flex items-center justify-between">
          <span class="font-medium">组织架构</span>
          <ElButton
            v-if="!hasHeadquarter"
            v-auth="'store:store:create'"
            size="small"
            @click="openCreate(1, null)"
          >
            创建总部
          </ElButton>
        </div>

        <ElInput v-model="treeKeyword" placeholder="搜索门店名称" clearable class="mb-3" />

        <ElTree
          v-if="storeTree.length > 0"
          ref="treeRef"
          class="min-h-0 flex-1 overflow-auto"
          :data="storeTree"
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          :expand-on-click-node="false"
          :filter-node-method="filterNode"
          :indent="20"
          default-expand-all
          highlight-current
          @node-click="handleNodeClick"
        >
          <template #default="{ data }">
            <div class="flex items-center gap-2 overflow-hidden">
              <span class="truncate">{{ data.name }}</span>
              <ElTag size="small" :type="typeTag(data.type)" effect="plain">
                {{ typeText(data.type) }}
              </ElTag>
            </div>
          </template>
        </ElTree>

        <div v-else class="flex flex-col items-center gap-3 py-8 text-sm text-g-600">
          <span>还没有总部，先创建组织根节点</span>
          <ElButton v-auth="'store:store:create'" plain @click="openCreate(1, null)">
            创建总部
          </ElButton>
        </div>
      </ElCard>

      <!-- 右侧：节点信息与门店内容同卡展示，避免上下两张卡片割裂 -->
      <ElCard class="art-table-card store-main-card">
        <div v-if="currentNode" class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <span class="truncate text-base font-medium">{{ currentNode.name }}</span>
              <ElTag :type="typeTag(currentNode.type)" effect="plain" size="small">
                {{ typeText(currentNode.type) }}
              </ElTag>
              <ElTag
                v-if="currentNode.type === 3"
                :type="statusTag(currentNode.status)"
                size="small"
              >
                {{ statusText(currentNode.status) }}
              </ElTag>
            </div>
            <div class="mt-1 truncate text-xs text-g-600">
              编码：{{ currentNode.code }}
              <span v-if="currentNode.shortName"> · 简称：{{ currentNode.shortName }}</span>
              <span v-if="currentNode.type !== 1"> · 上级：{{ parentName }}</span>
            </div>
          </div>

          <ElSpace wrap>
            <ElButton
              v-if="currentNode.type !== 3"
              v-auth="'store:store:create'"
              type="primary"
              @click="openCreate(currentNode.type === 1 ? 2 : 3, currentNode)"
            >
              {{ currentNode.type === 1 ? '新增下级' : '新增门店' }}
            </ElButton>
            <ElButton v-auth="'store:store:update'" @click="openEdit(currentNode)">编辑</ElButton>
            <!-- 总部是唯一根节点，后端禁止删除，前端不提供入口 -->
            <ElButton
              v-if="currentNode.type !== 1"
              v-auth="'store:store:delete'"
              type="danger"
              plain
              @click="deleteNode(currentNode)"
            >
              删除
            </ElButton>
          </ElSpace>
        </div>
        <div v-else class="py-6 text-center text-sm text-g-600">从左侧选择一个节点</div>

        <ElDivider v-if="currentNode" class="store-main-divider" />

        <div v-if="currentNode" class="min-h-0 flex-1 overflow-auto">
          <!-- 门店节点：展示门店档案，不铺列表 -->
          <ElDescriptions v-if="currentNode.type === 3" :column="2" border>
            <ElDescriptionsItem label="门店编码">{{ currentNode.code }}</ElDescriptionsItem>
            <ElDescriptionsItem label="门店名称">{{ currentNode.name }}</ElDescriptionsItem>
            <ElDescriptionsItem label="联系人">
              {{ currentNode.contactName || '—' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="联系电话">
              {{ currentNode.contactMobile || '—' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="详细地址" :span="2">
              {{ currentNode.address || '—' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="营业时间">
              {{ currentNode.openTime }} ~ {{ currentNode.closeTime }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="门店坐标">
              {{
                currentNode.longitude || currentNode.latitude
                  ? `${currentNode.longitude}, ${currentNode.latitude}`
                  : '未设置'
              }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="自动接单">
              {{ currentNode.autoAccept === 1 ? '自动' : '人工' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="线上点餐">
              {{ currentNode.onlineEnabled === 1 ? '开放' : '关闭' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="配送规则" :span="2">
              {{ describeDelivery(currentNode) }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="备注" :span="2">
              {{ currentNode.remark || '—' }}
            </ElDescriptionsItem>
          </ElDescriptions>

          <div v-if="currentNode.type === 3 && canManageTables" class="mt-4 flex gap-2">
            <ElButton v-auth="'store:table:list'" @click="goTables">管理桌台</ElButton>
            <ElButton v-auth="'store:employee:list'" @click="goEmployees">管理员工</ElButton>
          </div>

          <!-- 总部 / 大区：展示其下所有门店卡片（含下级大区里的门店） -->
          <template v-if="currentNode.type !== 3">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
              <span class="text-sm text-g-600">
                下级门店
                <span v-if="pagination.total > 0" class="ml-1">共 {{ pagination.total }} 家</span>
              </span>
              <div class="flex items-center gap-2">
                <ElInput
                  v-model="listKeyword"
                  placeholder="搜索门店名称"
                  clearable
                  class="w-52"
                  @keyup.enter="reloadList"
                  @clear="reloadList"
                />
                <ElSelect
                  v-model="listStatus"
                  placeholder="全部状态"
                  clearable
                  class="w-32"
                  @change="reloadList"
                >
                  <ElOption label="营业" :value="1" />
                  <ElOption label="休息" :value="2" />
                  <ElOption label="装修" :value="3" />
                  <ElOption label="停业" :value="4" />
                </ElSelect>
              </div>
            </div>

            <div v-if="loading" class="py-8 text-center text-sm text-g-600">加载中…</div>

            <!-- 空态采用一行文案 + 主按钮，浅色底块内居中，不用大插图撑满 -->
            <div
              v-else-if="data.length === 0"
              class="flex flex-col items-center gap-3 rounded-lg bg-[var(--el-fill-color-lighter)] py-10 text-sm text-g-600"
            >
              <span>{{ emptyHint }}</span>
              <ElButton
                v-auth="'store:store:create'"
                type="primary"
                plain
                @click="openCreate(currentNode.type === 1 ? 2 : 3, currentNode)"
              >
                {{ currentNode.type === 1 ? '新增大区' : '新增门店' }}
              </ElButton>
            </div>

            <div v-else class="grid grid-cols-1 gap-4 lg:grid-cols-2 2xl:grid-cols-3">
              <div
                v-for="item in data"
                :key="item.id"
                class="store-card"
                @click="selectStore(item)"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="min-w-0">
                    <div class="truncate font-medium">{{ item.name }}</div>
                    <div class="text-xs text-g-600">{{ item.code }}</div>
                  </div>
                  <ElTag :type="statusTag(item.status)">{{ statusText(item.status) }}</ElTag>
                </div>

                <div class="mt-3 space-y-1 text-xs text-g-700">
                  <div class="truncate">营业时间：{{ item.openTime }} ~ {{ item.closeTime }}</div>
                  <div class="truncate">地址：{{ item.address || '—' }}</div>
                  <div class="truncate">
                    联系人：{{ item.contactName || '—' }}
                    <span v-if="item.contactMobile"> / {{ item.contactMobile }}</span>
                  </div>
                </div>

                <div class="mt-3 flex justify-end gap-2" @click.stop>
                  <ElButton v-auth="'store:store:update'" size="small" @click="openEdit(item)">
                    编辑
                  </ElButton>
                  <ElButton
                    v-auth="'store:store:delete'"
                    size="small"
                    type="danger"
                    plain
                    @click="deleteNode(item)"
                  >
                    删除
                  </ElButton>
                </div>
              </div>
            </div>

            <div v-if="data.length > 0" class="mt-4 flex justify-end">
              <ElPagination
                :current-page="pagination.current"
                :page-size="pagination.size"
                :total="pagination.total"
                layout="total, prev, pager, next"
                background
                @current-change="handleCurrentChange"
              />
            </div>
          </template>
        </div>
      </ElCard>
    </div>

    <StoreDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :store-data="currentStore"
      :parent-node="parentNode"
      :default-type="createType"
      :store-tree="storeTree"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteStore, fetchGetStoreList, fetchGetStoreTree } from '@/api/store/store'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useRouter } from 'vue-router'
  import {
    ElButton,
    ElCard,
    ElDivider,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPagination,
    ElSelect,
    ElSpace,
    ElTag,
    ElTree
  } from 'element-plus'
  import StoreDialog from './modules/store-dialog.vue'

  defineOptions({ name: 'StoreStore' })

  type StoreItem = Api.Store.StoreItem
  type StoreTreeNode = Api.Store.StoreTreeNode
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()
  const router = useRouter()

  // 节点类型与状态展示配置，与后端枚举一一对应
  const TYPE_TEXT: Record<number, string> = { 1: '总部', 2: '大区', 3: '门店' }
  const TYPE_TAG: Record<number, 'primary' | 'warning' | 'success' | 'info'> = {
    1: 'primary',
    2: 'warning',
    3: 'success'
  }
  const STATUS_TEXT: Record<number, string> = { 1: '营业', 2: '休息', 3: '装修', 4: '停业' }
  const STATUS_TAG: Record<number, 'success' | 'info' | 'warning' | 'danger'> = {
    1: 'success',
    2: 'info',
    3: 'warning',
    4: 'danger'
  }

  // 节点类型文案：未知类型统一降级为「未知」，避免界面出现 undefined
  const typeText = (type: number): string => TYPE_TEXT[type] ?? '未知'
  // 节点类型对应的标签色：总部主色、大区警告色、门店成功色
  const typeTag = (type: number): 'primary' | 'warning' | 'success' | 'info' =>
    TYPE_TAG[type] ?? 'info'
  // 经营状态文案：与后端 StoreStatusEnum 一致
  const statusText = (status: number): string => STATUS_TEXT[status] ?? '未知'
  // 经营状态对应的标签色：营业绿、休息灰、装修黄、停业红
  const statusTag = (status: number): 'success' | 'info' | 'warning' | 'danger' =>
    STATUS_TAG[status] ?? 'info'

  // 只有门店才有桌台、员工等业务入口
  const canManageTables = computed(
    () => hasAuth('store:table:list') || hasAuth('store:employee:list')
  )

  // 空态提示：总部下可以先建大区，也可以直接挂直营门店
  const emptyHint = computed(() =>
    currentNode.value?.type === 1
      ? '总部下还没有门店，可先创建大区，或直接新增直营门店'
      : '该大区下还没有门店'
  )

  const treeRef = ref<InstanceType<typeof ElTree>>()
  const treeKeyword = ref('')
  const storeTree = ref<StoreTreeNode[]>([])
  const currentNode = ref<StoreTreeNode | null>(null)
  const hasHeadquarter = computed(() =>
    storeTree.value.some((node) => node.type === 1 || hasType(node, 1))
  )

  // 递归判断子树中是否存在指定类型节点
  function hasType(node: StoreTreeNode, type: number): boolean {
    return (node.children ?? []).some((child) => child.type === type || hasType(child, type))
  }

  // 弹窗状态：编辑传当前节点，新增按父节点推断默认类型
  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentStore = ref<StoreItem | null>(null)
  const parentNode = ref<StoreTreeNode | null>(null)
  const createType = ref(3)

  const listKeyword = ref('')
  const listStatus = ref<number | undefined>(undefined)

  // 分页查询门店：useTable 按 records 结构推导行类型，而后端返回 { list, total }，
  // 这里只做类型对齐，运行期结构与请求层解包逻辑保持不变。
  const fetchStorePage = async (
    params: Api.Store.StoreSearchParams
  ): Promise<Api.Common.PaginatedResponse<StoreItem>> =>
    (await fetchGetStoreList(params)) as unknown as Api.Common.PaginatedResponse<StoreItem>

  const {
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleCurrentChange,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchStorePage,
      apiParams: {
        pageNo: 1,
        pageSize: 12,
        type: 3
      },
      columnsFactory: () => []
    }
  })

  // 加载组织树，并按首个根节点初始化右侧内容
  const loadTree = async (): Promise<void> => {
    try {
      storeTree.value = await fetchGetStoreTree()
    } catch {
      storeTree.value = []
      return
    }

    if (storeTree.value.length === 0) {
      currentNode.value = null
      return
    }

    const first = storeTree.value[0]
    currentNode.value = first
    nextTick(() => treeRef.value?.setCurrentKey(first.id))
    if (first.type === 3) {
      return
    }
    applyEntity(first.id)
  }

  // 右侧列表按选中节点过滤，门店节点不查列表
  const applyEntity = (entityId: string): void => {
    searchParams.entityId = entityId
    searchParams.keyword = listKeyword.value || undefined
    searchParams.status = listStatus.value
    getData()
  }

  // 重新查询当前节点下的门店
  const reloadList = (): void => {
    if (!currentNode.value || currentNode.value.type === 3) return
    applyEntity(currentNode.value.id)
  }

  // 树节点点击：门店看档案，总部 / 大区看下级门店
  const handleNodeClick = (node: StoreTreeNode): void => {
    currentNode.value = node
    if (node.type === 3) return
    applyEntity(node.id)
  }

  // 树搜索：按名称或编码过滤节点，参数类型跟随 Element Plus 的 filter-node-method
  const filterNode = (value: string, data: Record<string, any>): boolean => {
    if (!value) return true
    return String(data.name ?? '').includes(value) || String(data.code ?? '').includes(value)
  }

  watch(treeKeyword, (value) => {
    treeRef.value?.filter(value)
  })

  onMounted(() => {
    loadTree()
  })

  // 打开新增：父节点决定新建节点的类型
  const openCreate = (type: number, parent: StoreTreeNode | null): void => {
    dialogType.value = 'add'
    createType.value = type
    parentNode.value = parent
    currentStore.value = null
    dialogVisible.value = true
  }

  // 打开编辑：总部 / 大区与门店共用同一个弹窗，字段按类型显隐
  const openEdit = (item: StoreItem): void => {
    dialogType.value = 'edit'
    currentStore.value = item
    parentNode.value = null
    dialogVisible.value = true
  }

  // 从卡片进入门店档案视图
  const selectStore = (item: StoreItem): void => {
    const found = findNode(storeTree.value, item.id)
    if (!found) return
    currentNode.value = found
    nextTick(() => treeRef.value?.setCurrentKey(found.id))
  }

  // 在树中按 ID 查找节点
  const findNode = (nodes: StoreTreeNode[], id: string): StoreTreeNode | null => {
    for (const node of nodes) {
      if (node.id === id) return node
      const hit = findNode(node.children ?? [], id)
      if (hit) return hit
    }

    return null
  }

  // 当前节点的上级名称：总部无上级，大区显示总部名，门店显示所属大区名
  const parentName = computed(() => {
    const parentId = currentNode.value?.parentId
    if (!parentId || parentId === '0') return '—'
    return findNode(storeTree.value, parentId)?.name ?? '—'
  })

  // 弹窗提交成功：新增或编辑后都要刷新树，编辑门店时同步当前节点
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
    } else {
      refreshUpdate()
    }

    loadTree().then(() => {
      if (dialogType.value === 'edit' && currentStore.value) {
        const found = findNode(storeTree.value, currentStore.value.id)
        if (found) currentNode.value = found
      }
    })
  }

  // 删除节点：有下级或关联数据时后端会拒绝，错误提示由请求层统一弹出
  const deleteNode = async (item: StoreItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除「${item.name}」吗？`, '删除节点', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteStore(item.id)
      ElMessage.success('删除成功')
      refreshRemove()
      currentNode.value = null
      loadTree()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 配送规则摘要：未配置时给出明确文案
  const describeDelivery = (item: StoreItem): string => {
    if (!item.deliveryRule) return '未配置（不限范围、费用 0）'
    try {
      const rule = JSON.parse(item.deliveryRule) as Api.Store.DeliveryRule
      const radius = rule.radius ? `${rule.radius} 米` : '不限范围'
      return `${radius} · 起送 ${(rule.minAmount / 100).toFixed(2)} 元 · 配送费 ${(rule.deliveryFee / 100).toFixed(2)} 元`
    } catch {
      return '配置格式异常'
    }
  }

  // 跳转到桌台页；门店上下文由请求头注入，这里只负责导航
  const goTables = (): void => {
    router.push('/store/table')
  }

  // 跳转到员工页；门店上下文同样由请求头注入
  const goEmployees = (): void => {
    router.push('/store/employee')
  }
</script>

<style scoped>
  /* 整页左右分栏：左树固定 280px，右侧自适应，两栏都吃满可视高度 */
  .store-layout {
    display: flex;
    flex: 1;
    gap: 12px;
    min-height: 0;
  }

  .store-layout .store-tree-card {
    flex: none;
    width: 280px;
  }

  .store-layout .store-main-card {
    min-width: 0;
  }

  /* 两张卡片内部都改成纵向弹性布局：树与门店内容各自滚动，不顶开整页 */
  .store-tree-card :deep(.el-card__body),
  .store-main-card :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  /* 节点信息与内容区的分隔线：只留呼吸感，不额外撑高卡片 */
  .store-main-divider {
    margin: 12px 0;
  }

  /* 组织树行：字号放大到 15px、行高 40px，节点之间留 4px 空隙更好点 */
  .store-tree-card :deep(.el-tree-node__content) {
    height: 40px;
    padding-right: 8px;
    margin-bottom: 4px;
    font-size: 15px;
    border-radius: 6px;
  }

  /* 选中节点高亮更明显一些，配合放大后的行高 */
  .store-tree-card :deep(.el-tree-node.is-current > .el-tree-node__content) {
    font-weight: 500;
  }

  /* 门店卡片：hover 时轻微描边高亮，比纯阴影更克制 */
  .store-card {
    padding: 16px;
    cursor: pointer;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
  }

  .store-card:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: var(--el-box-shadow-light);
  }

  /* 窄屏改成上下布局：树固定高度，右侧内容继续自适应 */
  @media (width <= 1100px) {
    .store-layout {
      flex-direction: column;
    }

    .store-layout .store-tree-card {
      flex: none;
      width: 100%;
      height: 280px;
    }
  }
</style>
