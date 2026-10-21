<!-- 角色授权弹窗：搜索菜单、全选/清空、已选计数，保存时自动带上父级目录 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="`角色授权 - ${role?.name || ''}`"
    width="600px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      v-if="isAdmin"
      type="info"
      :closable="false"
      show-icon
      title="超级管理员拥有全部权限，无需单独授权"
    />

    <template v-else>
      <div class="perm-toolbar">
        <ElInput
          v-model="keyword"
          class="perm-search"
          placeholder="搜索菜单名称"
          clearable
          :disabled="loading"
        />
        <ElButton link type="primary" :disabled="loading" @click="checkAll">全选</ElButton>
        <ElButton link type="danger" :disabled="loading || checkedTotal === 0" @click="clearAll">
          清空
        </ElButton>
      </div>

      <div class="perm-summary">
        已选 <span class="perm-summary__count">{{ checkedTotal }}</span> 项
        <span class="perm-summary__tip">勾选子菜单会自动包含所属目录</span>
      </div>

      <div v-loading="loading" class="permission-tree">
        <ElTree
          ref="treeRef"
          :data="treeData"
          show-checkbox
          node-key="id"
          default-expand-all
          :props="{ label: 'label', children: 'children' }"
          :filter-node-method="filterNode"
          :default-checked-keys="checkedKeys"
          @check="refreshCheckedCount"
        />
      </div>
    </template>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton
        v-if="!isAdmin"
        type="primary"
        :loading="submitting"
        :disabled="loading"
        @click="handleSubmit"
      >
        保存
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessageBox } from 'element-plus'
  import type { ElTree } from 'element-plus'
  import { fetchAssignRoleMenus, fetchGetRoleMenus } from '@/api/system/role'
  import { fetchGetMenuOptions } from '@/api/system/menu'

  defineOptions({ name: 'RolePermissionDialog' })

  type RoleItem = Api.SystemManage.RoleItem
  type MenuItem = Api.SystemManage.MenuItem

  interface Props {
    visible: boolean
    role?: RoleItem | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  interface TreeOption {
    id: string
    label: string
    children: TreeOption[]
  }

  const props = withDefaults(defineProps<Props>(), { role: null })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const isAdmin = computed(() => props.role?.code === 'SUPER_ADMIN')
  const loading = ref(false)
  const submitting = ref(false)
  const keyword = ref('')
  const treeRef = ref<InstanceType<typeof ElTree>>()
  const treeData = ref<TreeOption[]>([])
  const checkedKeys = ref<string[]>([])
  const leafIDs = ref<Set<string>>(new Set())
  const checkedTotal = ref(0)

  /**
   * 后端菜单树 → ElTree 结构，同时记录叶子节点用于回显
   */
  const buildTree = (nodes: MenuItem[]): TreeOption[] =>
    nodes.map((node) => {
      const children = buildTree(node.children || [])
      if (children.length === 0) {
        leafIDs.value.add(node.id)
      }
      return { id: node.id, label: node.name, children }
    })

  /**
   * 统计已勾选 + 半选节点数，半选代表被自动带入的父级目录
   */
  const refreshCheckedCount = (): void => {
    if (!treeRef.value) {
      checkedTotal.value = 0
      return
    }

    const checked = treeRef.value.getCheckedKeys(false) as string[]
    const halfChecked = treeRef.value.getHalfCheckedKeys() as string[]
    checkedTotal.value = checked.length + halfChecked.length
  }

  /**
   * 加载菜单树与角色已有授权，打开弹窗时调用
   */
  const loadData = async (): Promise<void> => {
    if (!props.role || isAdmin.value) return

    keyword.value = ''
    loading.value = true
    try {
      const [menus, granted] = await Promise.all([
        fetchGetMenuOptions(),
        fetchGetRoleMenus(props.role.id)
      ])

      leafIDs.value = new Set()
      treeData.value = buildTree(menus)
      // 只回显叶子节点：父节点由 ElTree 根据子节点自动勾选 / 半选，避免整棵子树被误勾
      checkedKeys.value = granted.filter((id) => leafIDs.value.has(id))

      await nextTick()
      // 同一个弹窗切换角色时 default-checked-keys 不会重新生效，这里显式回填一次
      treeRef.value?.setCheckedKeys(checkedKeys.value, false)
      refreshCheckedCount()
    } finally {
      loading.value = false
    }
  }

  // 每次打开弹窗都重新取数，避免上一个角色的勾选状态残留
  watch(
    () => props.visible,
    (visible) => {
      if (visible) loadData()
    },
    { immediate: true }
  )

  // 搜索词变化时联动过滤树节点，匹配到的节点会保留父级路径
  watch(keyword, (value) => {
    treeRef.value?.filter(value)
  })

  /**
   * ElTree 的过滤方法：按节点名称包含匹配
   */
  const filterNode = (value: string, data: Record<string, unknown>): boolean => {
    const kw = value?.trim()
    if (!kw) return true

    return String(data.label ?? '').includes(kw)
  }

  /**
   * 全选全部菜单（父节点勾选会覆盖所有子节点）
   */
  const checkAll = (): void => {
    const allIDs: string[] = []
    const walk = (nodes: TreeOption[]): void => {
      nodes.forEach((node) => {
        allIDs.push(node.id)
        walk(node.children)
      })
    }
    walk(treeData.value)

    treeRef.value?.setCheckedKeys(allIDs, false)
    refreshCheckedCount()
  }

  /**
   * 清空全部勾选
   */
  const clearAll = (): void => {
    treeRef.value?.setCheckedKeys([], false)
    refreshCheckedCount()
  }

  /**
   * 保存授权：已勾选 + 半选（父级目录）一起提交，保证父级节点也被授权
   */
  const handleSubmit = async (): Promise<void> => {
    if (!props.role || !treeRef.value) return

    const checked = treeRef.value.getCheckedKeys(false) as string[]
    const halfChecked = treeRef.value.getHalfCheckedKeys() as string[]
    const menuIDs = Array.from(new Set([...checked, ...halfChecked]))

    if (menuIDs.length === 0) {
      try {
        await ElMessageBox.confirm(
          '未勾选任何菜单，保存后该角色将没有任何权限，确定继续吗？',
          '取消全部授权',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
      } catch {
        return
      }
    }

    submitting.value = true
    try {
      await fetchAssignRoleMenus(props.role.id, menuIDs)
      ElMessage.success('授权成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .perm-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  .perm-search {
    flex: 1;
  }

  .perm-summary {
    margin-bottom: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .perm-summary__count {
    color: var(--el-color-primary);
    font-size: 13px;
  }

  .perm-summary__tip {
    margin-left: 8px;
    color: var(--el-text-color-placeholder);
  }

  .permission-tree {
    max-height: 420px;
    overflow: auto;
    padding: 4px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
  }
</style>
