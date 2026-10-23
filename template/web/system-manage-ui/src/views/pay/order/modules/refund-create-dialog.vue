<!-- 发起退款弹窗：对已支付的支付单退全款或部分金额，金额上限由后端二次校验 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="发起退款"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      :title="`支付单 ${props.order?.no ?? ''} · 实付 ${formatYuan(props.order?.price ?? 0)} 元 · 已退 ${formatYuan(props.order?.refundPrice ?? 0)} 元`"
      show-icon
    />
    <div class="mt-1 text-xs text-g-600">
      可退余额 {{ formatYuan(refundable) }} 元；等待中的退款也会占用额度，以服务端校验为准。
    </div>

    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px" class="mt-4">
      <ElFormItem label="退款金额" prop="refundYuan">
        <ElInputNumber
          v-model="form.refundYuan"
          :min="0.01"
          :max="refundableYuan"
          :precision="2"
          :step="1"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="业务退款号" prop="merchantRefundId">
        <ElInput
          v-model="form.merchantRefundId"
          placeholder="业务侧退款单号，同一应用下不可重复"
          maxlength="64"
        />
      </ElFormItem>

      <ElFormItem label="退款原因" prop="reason">
        <ElInput
          v-model="form.reason"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="选填，会同步给渠道"
        />
      </ElFormItem>
    </ElForm>
    <div class="text-xs text-g-600">
      渠道同步确认完成的会立即变为「退款成功」；微信通常在处理中，需等退款回调补齐。
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确认退款</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreatePayRefund } from '@/api/pay/refund'

  defineOptions({ name: 'PayRefundCreateDialog' })

  interface Props {
    visible: boolean
    order?: Api.Pay.OrderItem | null
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
    refundYuan: 0.01 as number | null,
    merchantRefundId: '',
    reason: ''
  })

  // 可退金额（分）= 实付 - 已退，等待中的退款由服务端再校验
  const refundable = computed(() => {
    const price = props.order?.price ?? 0
    const refunded = props.order?.refundPrice ?? 0

    return Math.max(price - refunded, 0)
  })

  const refundableYuan = computed(() => refundable.value / 100)

  const rules: FormRules = {
    refundYuan: [{ required: true, message: '请输入退款金额', trigger: 'change' }],
    merchantRefundId: [{ required: true, message: '请输入业务退款单号', trigger: 'blur' }]
  }

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 每次打开弹窗按可退余额预填退款金额
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.refundYuan = Number(formatYuan(refundable.value))
      form.merchantRefundId = ''
      form.reason = ''
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交退款：元转分后调用接口，成功后关闭并通知父页面刷新
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.order) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const refundPrice = Math.round((form.refundYuan || 0) * 100)
    if (refundPrice <= 0) {
      ElMessage.warning('退款金额必须大于 0')
      return
    }

    submitting.value = true
    try {
      const refund = await fetchCreatePayRefund({
        orderId: props.order.id,
        merchantRefundId: form.merchantRefundId,
        refundPrice,
        reason: form.reason || undefined
      })
      ElMessage.success(refund?.status === 1 ? '退款成功' : '退款已提交，等待渠道处理结果')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
