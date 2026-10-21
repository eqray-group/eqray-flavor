<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑角色' : '新增角色'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
      <ElFormItem label="角色名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入角色名称" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="角色标识" prop="code">
        <ElInput
          v-model="form.code"
          :disabled="isBuiltin"
          placeholder="字母开头，如 editor、SUPER_ADMIN"
          maxlength="64"
        />
      </ElFormItem>
      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          :disabled="isBuiltin"
          active-text="启用"
          inactive-text="禁用"
        />
      </ElFormItem>
      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="选填"
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
  import { fetchCreateRole, fetchUpdateRole } from '@/api/system/role'

  type RoleItem = Api.SystemManage.RoleItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    roleData?: RoleItem | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), { roleData: null })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')
  // 内置角色不允许改标识与禁用（后端同样会校验）
  const isBuiltin = computed(() => props.roleData?.type === 1)
  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    code: '',
    sort: 0,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入角色名称', trigger: 'blur' },
      { max: 64, message: '角色名称不能超过 64 个字符', trigger: 'blur' }
    ],
    code: [
      { required: true, message: '请输入角色标识', trigger: 'blur' },
      { max: 64, message: '角色标识不能超过 64 个字符', trigger: 'blur' },
      {
        pattern: /^[A-Za-z][A-Za-z0-9_-]*$/,
        message: '只能包含字母、数字、下划线与中划线，且以字母开头',
        trigger: 'blur'
      }
    ],
    remark: [{ max: 500, message: '备注不能超过 500 个字符', trigger: 'blur' }]
  }

  const initForm = (): void => {
    const data = props.roleData
    Object.assign(form, {
      name: data?.name ?? '',
      code: data?.code ?? '',
      sort: data?.sort ?? 0,
      status: data?.status ?? 1,
      remark: data?.remark ?? ''
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

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.SystemManage.RoleSaveParams = {
      name: form.name.trim(),
      code: form.code.trim(),
      sort: form.sort,
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.roleData) {
        await fetchUpdateRole(props.roleData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateRole(payload)
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
