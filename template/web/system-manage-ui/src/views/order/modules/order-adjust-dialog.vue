<!-- 订单改价弹窗：正数加价、负数减免，仅未支付订单可改 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="订单改价"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      show-icon
      :title="`当前应付 ${formatYuan(props.order?.payAmount ?? 0)} 元`"
    />
    <div class="mt-1 text-xs text-g-600">
      正数加价、负数减免；已支付订单不能改价，改价差额会累计记录便于对账。
    </div>

    <ElForm ref="formRef" :model="form" :rules="rules" label-width="96px" class="mt-4">
      <ElFormItem label="改价差额" prop="deltaYuan">
        <ElInputNumber
          v-model="form.deltaYuan"
          :precision="2"
          :step="1"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="改价后">
        <span class="text-base font-medium">{{ formatYuan(previewAmount) }} 元</span>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确认改价</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchAdjustOrderPrice } from '@/api/order/order'

  defineOptions({ name: 'OrderAdjustDialog' })

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
    deltaYuan: 0 as number | null
  })

  const rules: FormRules = {
    deltaYuan: [
      {
        validator: (_rule, value: number | null, callback) => {
          if (!value) {
            callback(new Error('改价差额不能为 0'))
            return
          }
          callback()
        },
        trigger: 'change'
      }
    ]
  }

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 改价后的应付金额，实时预览
  const previewAmount = computed(() => {
    const current = props.order?.payAmount ?? 0
    return current + Math.round((form.deltaYuan || 0) * 100)
  })

  // 打开弹窗时重置输入
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.deltaYuan = 0
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交改价：元转分后调用接口
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.order) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const delta = Math.round((form.deltaYuan || 0) * 100)
    if (delta === 0) {
      ElMessage.warning('改价差额不能为 0')
      return
    }
    if (previewAmount.value < 0) {
      ElMessage.warning('改价后应付金额不能小于 0')
      return
    }

    submitting.value = true
    try {
      await fetchAdjustOrderPrice(props.order.id, { delta })
      ElMessage.success('改价成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
