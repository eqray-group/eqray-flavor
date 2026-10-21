<!-- 用户分配角色弹窗：搜索、全选、已选计数，覆盖式保存 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="`分配角色 - ${user?.username || ''}`"
    width="600px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="role-toolbar">
      <ElInput v-model="keyword" class="role-search" placeholder="搜索角色名称或编码" clearable />
      <ElButton link type="primary" :disabled="filteredRoles.length === 0" @click="toggleSelectAll">
        {{ allFilteredSelected ? '取消全选' : '全选' }}
      </ElButton>
      <ElButton link type="danger" :disabled="checkedRoleIDs.length === 0" @click="clearSelected">
        清空
      </ElButton>
    </div>

    <div class="role-summary">
      已选 <span class="role-summary__count">{{ checkedRoleIDs.length }}</span> /
      {{ roleOptions.length }} 个角色
    </div>

    <ElAlert
      v-if="disabledGrantedIDs.length"
      class="role-alert"
      type="warning"
      :closable="false"
      show-icon
      :title="`有 ${disabledGrantedIDs.length} 个已分配角色已被停用，保存后会继续保留`"
    />

    <div v-loading="loading" class="role-list">
      <ElCheckboxGroup v-model="checkedRoleIDs" class="role-grid">
        <ElCheckbox v-for="role in filteredRoles" :key="role.id" :value="role.id" class="role-item">
          <div class="role-item__title">
            <span class="role-item__name">{{ role.name }}</span>
            <ElTag v-if="role.type === 1" size="small" type="info" effect="plain">内置</ElTag>
          </div>
          <span class="role-item__code">{{ role.code }}</span>
        </ElCheckbox>
      </ElCheckboxGroup>

      <ElEmpty
        v-if="!loading && filteredRoles.length === 0"
        :description="roleOptions.length > 0 ? '没有匹配的角色' : '暂无可分配的角色'"
      />
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessageBox } from 'element-plus'
  import { fetchAssignUserRoles, fetchGetUserRoles } from '@/api/system/user'
  import { fetchGetRoleOptions } from '@/api/system/role'

  defineOptions({ name: 'UserRoleDialog' })

  type UserItem = Api.SystemManage.UserListItem
  type RoleItem = Api.SystemManage.RoleItem

  interface Props {
    visible: boolean
    user?: Partial<UserItem> | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), { user: null })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const loading = ref(false)
  const submitting = ref(false)
  const keyword = ref('')
  const roleOptions = ref<RoleItem[]>([])
  const checkedRoleIDs = ref<string[]>([])

  // 按名称或编码过滤后的角色，全选只作用于当前结果
  const filteredRoles = computed<RoleItem[]>(() => {
    const kw = keyword.value.trim().toLowerCase()
    if (!kw) return roleOptions.value

    return roleOptions.value.filter(
      (role) => role.name.toLowerCase().includes(kw) || role.code.toLowerCase().includes(kw)
    )
  })

  // 当前筛选结果是否已全部勾选
  const allFilteredSelected = computed<boolean>(() => {
    if (filteredRoles.value.length === 0) return false
    return filteredRoles.value.every((role) => checkedRoleIDs.value.includes(role.id))
  })

  // 已分配但不在可选列表里的角色（角色被停用），保存时需要原样带回
  const disabledGrantedIDs = computed<string[]>(() =>
    checkedRoleIDs.value.filter((id) => !roleOptions.value.some((role) => role.id === id))
  )

  /**
   * 加载可选角色与用户已分配的角色，打开弹窗时调用
   */
  const loadData = async (): Promise<void> => {
    if (!props.user?.id) return

    keyword.value = ''
    loading.value = true
    try {
      const [roles, granted] = await Promise.all([
        fetchGetRoleOptions(),
        fetchGetUserRoles(props.user.id)
      ])
      roleOptions.value = roles
      checkedRoleIDs.value = granted
    } finally {
      loading.value = false
    }
  }

  // 每次打开弹窗都重新取数，避免上一个用户的勾选状态残留
  watch(
    () => props.visible,
    (visible) => {
      if (visible) loadData()
    },
    { immediate: true }
  )

  /**
   * 全选 / 取消全选当前筛选结果
   */
  const toggleSelectAll = (): void => {
    const ids = filteredRoles.value.map((role) => role.id)
    if (allFilteredSelected.value) {
      checkedRoleIDs.value = checkedRoleIDs.value.filter((id) => !ids.includes(id))
      return
    }

    checkedRoleIDs.value = Array.from(new Set([...checkedRoleIDs.value, ...ids]))
  }

  /**
   * 清空已选角色
   */
  const clearSelected = (): void => {
    checkedRoleIDs.value = []
  }

  /**
   * 保存角色分配（覆盖式）；清空角色前二次确认
   */
  const handleSubmit = async (): Promise<void> => {
    if (!props.user?.id) return

    if (checkedRoleIDs.value.length === 0) {
      try {
        await ElMessageBox.confirm(
          '未勾选任何角色，保存后该用户将没有任何权限，确定继续吗？',
          '取消全部角色',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
      } catch {
        return
      }
    }

    submitting.value = true
    try {
      await fetchAssignUserRoles(props.user.id, checkedRoleIDs.value)
      ElMessage.success('分配成功')
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
  .role-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  .role-search {
    flex: 1;
  }

  .role-summary {
    margin-bottom: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .role-summary__count {
    color: var(--el-color-primary);
    font-size: 13px;
  }

  .role-alert {
    margin-bottom: 8px;
  }

  .role-list {
    max-height: 360px;
    overflow: auto;
  }

  .role-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .role-item {
    height: auto;
    margin-right: 0;
    padding: 8px 12px;
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    transition: border-color 0.2s;
  }

  .role-item:hover {
    border-color: var(--el-color-primary);
  }

  .role-item :deep(.el-checkbox__label) {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
    overflow: hidden;
  }

  .role-item__title {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .role-item__name {
    overflow: hidden;
    color: var(--el-text-color-primary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .role-item__code {
    overflow: hidden;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
