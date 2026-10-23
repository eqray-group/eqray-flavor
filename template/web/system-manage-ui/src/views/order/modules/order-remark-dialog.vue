<!-- 订单备注弹窗：顾客备注与商家备注分列保存，店员改动不覆盖顾客原始诉求 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="修改备注"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="mb-3 text-xs text-g-600">订单号：{{ props.order?.no || '-' }}</div>

    <ElForm ref="formRef" :model="form" label-width="96px">
      <ElFormItem label="顾客备注">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="3"
          maxlength="255"
          show-word-limit
          placeholder="顾客下单时填写的备注"
        />
      </ElFormItem>

      <ElFormItem label="商家备注">
        <ElInput
          v-model="form.merchantRemark"
          type="textarea"
          :rows="3"
          maxlength="255"
          show-word-limit
          placeholder="如：已电话确认、加急出餐"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance } from 'element-plus'
  import { fetchUpdateOrderRemark } from '@/api/order/order'

  defineOptions({ name: 'OrderRemarkDialog' })

  interface Props {
    visible: boolean
    order?: Api.Order.OrderItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    order: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    remark: '',
    merchantRemark: ''
  })

  // 打开时回填当前备注，便于在原文上修改
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.remark = props.order?.remark ?? ''
      form.merchantRemark = props.order?.merchantRemark ?? ''
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交备注：允许清空，后端按空串处理
  const handleSubmit = async (): Promise<void> => {
    if (!props.order) return

    submitting.value = true
    try {
      await fetchUpdateOrderRemark(props.order.id, {
        remark: form.remark,
        merchantRemark: form.merchantRemark
      })
      ElMessage.success('备注已保存')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
