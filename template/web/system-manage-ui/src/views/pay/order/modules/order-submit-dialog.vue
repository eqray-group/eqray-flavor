<!-- 重新调起支付弹窗：对仍待支付的支付单补 openid / 付款码后再次调用渠道 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="重新调起支付"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      :title="`支付单 ${props.order?.no ?? ''} · ${props.order?.channelName ?? ''}`"
      show-icon
    />
    <div class="mt-1 text-xs text-g-600">
      重新调起会新增一条渠道调用记录；已过期或已支付成功的支付单会被拒绝。
    </div>

    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px" class="mt-4">
      <ElFormItem v-if="needOpenId" label="用户 openid" prop="openId">
        <ElInput v-model="form.openId" placeholder="微信 JSAPI / 小程序支付必填" maxlength="128" />
      </ElFormItem>

      <ElFormItem v-if="needAuthCode" label="用户付款码" prop="authCode">
        <ElInput v-model="form.authCode" placeholder="支付宝条码支付必填" maxlength="64" />
      </ElFormItem>

      <ElFormItem v-if="!needOpenId && !needAuthCode" label="调起方式">
        <span class="text-sm text-g-600"> 该渠道重新调起不需要额外的用户标识，直接提交即可。 </span>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">重新调起</ElButton>
    </template>
  </ElDialog>

  <ElDialog v-model="resultVisible" title="支付调起参数" width="520px" align-center>
    <div class="mb-2 text-sm">支付单号：{{ created?.no }}</div>
    <ElInput :model-value="paramsText" type="textarea" :rows="8" readonly resize="none" />
    <template #footer>
      <ElButton @click="copyParams">复制参数</ElButton>
      <ElButton type="primary" @click="resultVisible = false">完成</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchSubmitPayOrder } from '@/api/pay/order'

  defineOptions({ name: 'PayOrderSubmitDialog' })

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
  const resultVisible = ref(false)
  const created = ref<Api.Pay.OrderSubmitResult | null>(null)

  const form = reactive({
    openId: '',
    authCode: ''
  })

  // 微信小程序 / 公众号渠道需要 openid
  const needOpenId = computed(() => {
    const code = props.order?.channelCode ?? ''

    return code === 'wx_lite' || code === 'wx_pub'
  })

  // 支付宝条码渠道需要用户付款码
  const needAuthCode = computed(() => props.order?.channelCode === 'alipay_bar')

  const rules: FormRules = {
    openId: [
      {
        validator: (_rule, value: string, callback) => {
          if (needOpenId.value && !value) {
            callback(new Error('微信 JSAPI / 小程序支付必须填写 openid'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    authCode: [
      {
        validator: (_rule, value: string, callback) => {
          if (needAuthCode.value && !value) {
            callback(new Error('支付宝条码支付必须填写付款码'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  // 调起参数字符串，供复制与展示
  const paramsText = computed(() =>
    created.value ? JSON.stringify(created.value.payParams ?? {}, null, 2) : ''
  )

  // 打开弹窗时清空上一次输入
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.openId = ''
      form.authCode = ''
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交重新调起，成功后展示新的调起参数
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.order) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      const submitted = await fetchSubmitPayOrder(props.order.id, {
        openId: form.openId || undefined,
        authCode: form.authCode || undefined
      })
      created.value = submitted
      resultVisible.value = true
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  // 复制调起参数，便于在联调工具里直接使用
  const copyParams = async (): Promise<void> => {
    try {
      await navigator.clipboard.writeText(paramsText.value)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('当前浏览器不支持自动复制，请手动选择文本')
    }
  }
</script>
