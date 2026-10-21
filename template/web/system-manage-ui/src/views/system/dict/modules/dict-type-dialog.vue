<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑字典类型' : '新增字典类型'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
      <ElFormItem label="字典名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入字典名称，如：通用状态" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="字典类型" prop="type">
        <ElInput
          v-model="form.type"
          :disabled="isEdit"
          placeholder="小写字母开头，如：common_status"
          maxlength="64"
        />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
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
  import { fetchCreateDictType, fetchUpdateDictType } from '@/api/system/dict'

  type DictTypeItem = Api.SystemManage.DictTypeItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    dictType?: DictTypeItem | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), { dictType: null })
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
    type: '',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入字典名称', trigger: 'blur' },
      { max: 64, message: '字典名称不能超过 64 个字符', trigger: 'blur' }
    ],
    type: [
      { required: true, message: '请输入字典类型', trigger: 'blur' },
      { max: 64, message: '字典类型不能超过 64 个字符', trigger: 'blur' },
      {
        pattern: /^[a-z][a-z0-9_]*$/,
        message: '只能包含小写字母、数字与下划线，且以字母开头',
        trigger: 'blur'
      }
    ],
    remark: [{ max: 500, message: '备注不能超过 500 个字符', trigger: 'blur' }]
  }

  const initForm = (): void => {
    const data = props.dictType
    Object.assign(form, {
      name: data?.name ?? '',
      type: data?.type ?? '',
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

    const payload: Api.SystemManage.DictTypeSaveParams = {
      name: form.name.trim(),
      type: form.type.trim(),
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.dictType) {
        await fetchUpdateDictType(props.dictType.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateDictType(payload)
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
