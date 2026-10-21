<!-- 会员标签新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑标签' : '新增标签'"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="80px">
      <ElFormItem label="标签名" prop="name">
        <ElInput v-model="form.name" placeholder="如：高价值客户" maxlength="30" />
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
          :rows="2"
          maxlength="255"
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
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreateTag, fetchUpdateTag } from '@/api/member/tag'

  defineOptions({ name: 'TagDialog' })

  type TagItem = Api.Member.TagItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    tagData?: TagItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    tagData: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入标签名', trigger: 'blur' },
      { max: 30, message: '标签名不能超过 30 个字符', trigger: 'blur' }
    ],
    remark: [{ max: 255, message: '备注不能超过 255 个字符', trigger: 'blur' }]
  }

  watch(
    () => [props.visible, props.tagData] as const,
    ([visible]) => {
      if (!visible) return

      Object.assign(form, {
        name: props.tagData?.name ?? '',
        status: props.tagData?.status ?? 1,
        remark: props.tagData?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.TagSaveParams = {
      name: form.name.trim(),
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.tagData) {
        await fetchUpdateTag(props.tagData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateTag(payload)
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
