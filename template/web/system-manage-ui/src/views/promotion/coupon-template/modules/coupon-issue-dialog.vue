<!-- 发放优惠券弹窗：选择会员与发放张数 -->
<template>
  <ElDialog v-model="visible" title="发放优惠券" width="460px" @close="handleClose">
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="90px">
      <ElFormItem label="券模板">
        <ElInput :model-value="templateName" disabled />
      </ElFormItem>

      <ElFormItem label="会员 ID" prop="memberId">
        <ElInput v-model="formData.memberId" placeholder="会员 ID" clearable />
      </ElFormItem>

      <ElFormItem label="发放张数" prop="count">
        <ElInputNumber v-model="formData.count" :min="1" :max="100" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import { fetchIssueCoupon } from '@/api/promotion/promotion'

  defineOptions({ name: 'CouponIssueDialog' })

  interface CouponIssueData {
    id: string
    name: string
  }

  const visible = defineModel<boolean>({ default: false })

  const props = defineProps<{
    /** 当前发券的券模板 */
    template?: CouponIssueData | null
  }>()

  const emit = defineEmits<{
    (e: 'success'): void
  }>()

  const formRef = ref<FormInstance>()

  const formData = reactive({
    memberId: '',
    count: 1
  })

  const formRules = reactive<FormRules>({
    memberId: [{ required: true, message: '请输入会员 ID', trigger: 'blur' }],
    count: [{ required: true, message: '请输入发放张数', trigger: 'blur' }]
  })

  const submitting = ref(false)

  const templateName = computed(() => props.template?.name || '')

  // 打开弹窗时重置表单
  watch(
    () => visible.value,
    (open) => {
      if (open) {
        formData.memberId = ''
        formData.count = 1
        formRef.value?.clearValidate()
      }
    }
  )

  // 关闭弹窗
  const handleClose = (): void => {
    visible.value = false
  }

  // 提交发券
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await fetchIssueCoupon(props.template!.id, {
        memberId: formData.memberId,
        count: formData.count
      })
      ElMessage.success('发放成功')
      visible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
