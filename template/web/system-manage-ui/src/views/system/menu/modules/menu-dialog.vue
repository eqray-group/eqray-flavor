<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑菜单' : '新增菜单'"
    width="600px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="菜单类型" prop="type">
        <ElRadioGroup v-model="form.type" :disabled="isEdit">
          <ElRadioButton :value="1">目录</ElRadioButton>
          <ElRadioButton :value="2">菜单</ElRadioButton>
          <ElRadioButton :value="3">按钮</ElRadioButton>
        </ElRadioGroup>
      </ElFormItem>

      <ElFormItem label="上级菜单" prop="parentId">
        <ElTreeSelect
          v-model="form.parentId"
          :data="parentOptions"
          check-strictly
          clearable
          :render-after-expand="false"
          placeholder="不选表示顶级菜单"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="菜单名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入菜单名称" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
      </ElFormItem>

      <template v-if="form.type !== 3">
        <ElFormItem label="路由地址" prop="path">
          <ElInput
            v-model="form.path"
            :placeholder="form.type === 1 ? '如：/system' : '如：user'"
            maxlength="255"
          />
        </ElFormItem>
        <ElFormItem label="图标" prop="icon">
          <ElInput v-model="form.icon" placeholder="如：ri:user-line" maxlength="64" />
        </ElFormItem>
        <ElFormItem label="是否显示" prop="visible">
          <ElSwitch
            v-model="form.visible"
            :active-value="1"
            :inactive-value="0"
            active-text="显示"
            inactive-text="隐藏"
          />
        </ElFormItem>
      </template>

      <template v-if="form.type === 2">
        <ElFormItem label="组件路径" prop="component">
          <ElInput v-model="form.component" placeholder="如：/system/user" maxlength="255" />
        </ElFormItem>
        <ElFormItem label="组件名称" prop="componentName">
          <ElInput v-model="form.componentName" placeholder="与组件 name 保持一致" maxlength="64" />
        </ElFormItem>
        <ElFormItem label="页面缓存" prop="keepAlive">
          <ElSwitch v-model="form.keepAlive" :active-value="1" :inactive-value="0" />
        </ElFormItem>
      </template>

      <template v-if="form.type === 1">
        <ElFormItem label="始终显示" prop="alwaysShow">
          <ElSwitch v-model="form.alwaysShow" :active-value="1" :inactive-value="0" />
        </ElFormItem>
      </template>

      <template v-if="form.type === 3">
        <ElFormItem label="权限标识" prop="permission">
          <ElInput v-model="form.permission" placeholder="如：system:user:create" maxlength="128" />
        </ElFormItem>
      </template>

      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          active-text="启用"
          inactive-text="禁用"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreateMenu, fetchUpdateMenu } from '@/api/system/menu'

  type MenuItem = Api.SystemManage.MenuItem

  interface Props {
    visible: boolean
    /** add：新增；edit：编辑 */
    type: 'add' | 'edit'
    /** 编辑时传入行数据 */
    menuData?: MenuItem | null
    /** 新增时指定上级菜单 */
    parentId?: string
    /** 完整菜单树，用于选择上级 */
    menuTree: MenuItem[]
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  interface TreeOption {
    value: string
    label: string
    children: TreeOption[]
  }

  const props = withDefaults(defineProps<Props>(), {
    menuData: null,
    parentId: '',
    menuTree: () => []
  })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')
  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    type: 2,
    parentId: '',
    sort: 0,
    path: '',
    icon: '',
    component: '',
    componentName: '',
    permission: '',
    status: 1,
    visible: 1,
    keepAlive: 1,
    alwaysShow: 0
  })

  const rules = computed<FormRules>(() => ({
    name: [
      { required: true, message: '请输入菜单名称', trigger: 'blur' },
      { max: 64, message: '菜单名称不能超过 64 个字符', trigger: 'blur' }
    ],
    path:
      form.type === 3
        ? []
        : [
            { required: true, message: '请输入路由地址', trigger: 'blur' },
            { max: 255, message: '路由地址不能超过 255 个字符', trigger: 'blur' }
          ],
    icon: [{ max: 64, message: '图标不能超过 64 个字符', trigger: 'blur' }],
    component: [{ max: 255, message: '组件路径不能超过 255 个字符', trigger: 'blur' }],
    componentName: [{ max: 64, message: '组件名称不能超过 64 个字符', trigger: 'blur' }],
    permission:
      form.type === 3
        ? [
            { required: true, message: '请输入权限标识', trigger: 'blur' },
            { max: 128, message: '权限标识不能超过 128 个字符', trigger: 'blur' },
            {
              pattern: /^[a-z][a-z0-9]*(:[a-z0-9-]+)+$/,
              message: '格式如 system:user:create',
              trigger: 'blur'
            }
          ]
        : []
  }))

  /** 收集自身与全部子孙 ID，避免把节点挂到自己或自己的子树下 */
  const collectIds = (node: MenuItem, target: Set<string>): void => {
    target.add(node.id)
    ;(node.children || []).forEach((child) => collectIds(child, target))
  }

  /** 上级菜单选项：只保留目录与菜单，编辑时排除自身子树 */
  const parentOptions = computed<TreeOption[]>(() => {
    const excluded = new Set<string>()
    if (isEdit.value && props.menuData) {
      collectIds(props.menuData, excluded)
    }

    const build = (nodes: MenuItem[]): TreeOption[] =>
      nodes
        .filter((node) => node.type !== 3 && !excluded.has(node.id))
        .map((node) => ({
          value: node.id,
          label: node.name,
          children: build(node.children || [])
        }))

    return [{ value: '0', label: '顶级菜单', children: [] }, ...build(props.menuTree)]
  })

  /** 深度查找菜单节点 */
  const findMenu = (nodes: MenuItem[], id: string): MenuItem | null => {
    for (const node of nodes) {
      if (node.id === id) return node
      const found = findMenu(node.children || [], id)
      if (found) return found
    }
    return null
  }

  /** 用行数据或默认值初始化表单 */
  const initForm = (): void => {
    const data = props.menuData
    if (isEdit.value && data) {
      Object.assign(form, {
        name: data.name,
        type: data.type,
        parentId: data.parentId || '0',
        sort: data.sort ?? 0,
        path: data.path,
        icon: data.icon,
        component: data.component,
        componentName: data.componentName,
        permission: data.permission,
        status: data.status,
        visible: data.visible,
        keepAlive: data.keepAlive,
        alwaysShow: data.alwaysShow
      })
      return
    }

    const parent = props.parentId ? findMenu(props.menuTree, props.parentId) : null

    Object.assign(form, {
      name: '',
      // 目录下默认新增菜单，菜单下默认新增按钮
      type: parent ? (parent.type === 1 ? 2 : 3) : 2,
      parentId: props.parentId || '0',
      sort: 0,
      path: '',
      icon: '',
      component: '',
      componentName: '',
      permission: '',
      status: 1,
      visible: 1,
      keepAlive: 1,
      alwaysShow: 0
    })
  }

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      initForm()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  /** 类型切换后清掉不再适用的字段，避免脏数据落库 */
  watch(
    () => form.type,
    (type) => {
      if (type === 3) {
        form.path = ''
        form.icon = ''
        form.component = ''
        form.componentName = ''
      } else {
        form.permission = ''
      }
      if (type !== 1) {
        form.alwaysShow = 0
      }
      if (type !== 2) {
        form.keepAlive = 0
      }
    }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.SystemManage.MenuSaveParams = {
      name: form.name.trim(),
      permission: form.type === 3 ? form.permission.trim() : '',
      type: form.type,
      sort: form.sort,
      parentId: form.parentId || '0',
      path: form.type === 3 ? '' : form.path.trim(),
      icon: form.type === 3 ? '' : form.icon.trim(),
      component: form.type === 2 ? form.component.trim() : '',
      componentName: form.type === 2 ? form.componentName.trim() : '',
      status: form.status,
      visible: form.visible,
      keepAlive: form.keepAlive,
      alwaysShow: form.alwaysShow
    }

    submitting.value = true
    try {
      if (isEdit.value && props.menuData) {
        await fetchUpdateMenu(props.menuData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateMenu(payload)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
