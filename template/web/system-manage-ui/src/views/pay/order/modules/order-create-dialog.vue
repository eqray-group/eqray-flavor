<!-- 创建支付单弹窗：选应用与渠道后调起渠道下单，成功展示二维码 / 支付链接 / 调起参数 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="创建支付单"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px">
      <ElFormItem label="所属应用" prop="appId">
        <ElSelect
          v-model="form.appId"
          class="w-full"
          filterable
          placeholder="请选择支付应用"
          no-data-text="请先创建支付应用"
          @change="handleAppChange"
        >
          <ElOption
            v-for="app in props.appOptions"
            :key="app.id"
            :label="app.name"
            :value="app.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="支付渠道" prop="channelCode">
        <ElSelect
          v-model="form.channelCode"
          class="w-full"
          filterable
          :loading="channelLoading"
          placeholder="请选择已开启的渠道"
          no-data-text="该应用下暂无开启的渠道"
        >
          <ElOption
            v-for="channel in channels"
            :key="channel.id"
            :label="`${channel.channelName}（${channel.code}）`"
            :value="channel.code"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem v-if="isWallet" label="付款用户" prop="userId">
        <ElInput
          v-model="form.userId"
          placeholder="钱包余额支付必须指定付款用户编号"
          maxlength="32"
        />
        <div class="mt-1 text-xs text-g-600">余额从该用户的钱包扣减，请填写会员的用户编号。</div>
      </ElFormItem>

      <ElFormItem label="业务订单号" prop="merchantOrderId">
        <ElInput
          v-model="form.merchantOrderId"
          placeholder="业务侧唯一单号，用于下单幂等"
          maxlength="64"
        />
      </ElFormItem>

      <ElFormItem label="商品标题" prop="subject">
        <ElInput v-model="form.subject" placeholder="如「堂食订单 20261022-001」" maxlength="128" />
      </ElFormItem>

      <ElFormItem label="支付金额" prop="priceYuan">
        <ElInputNumber
          v-model="form.priceYuan"
          :min="0.01"
          :precision="2"
          :step="1"
          style="width: 100%"
        />
        <div class="mt-1 text-xs text-g-600">单位元，提交时换算成分；金额只认服务端试算结果。</div>
      </ElFormItem>

      <ElFormItem v-if="isWechatJsapi" label="用户 openid" prop="openId">
        <ElInput v-model="form.openId" placeholder="微信 JSAPI / 小程序支付必填" maxlength="128" />
      </ElFormItem>

      <ElFormItem v-if="isAlipayBar" label="用户付款码" prop="authCode">
        <ElInput
          v-model="form.authCode"
          placeholder="支付宝条码支付必填，收银员扫码得到"
          maxlength="64"
        />
      </ElFormItem>

      <ElFormItem label="回调地址" prop="notifyUrl">
        <ElInput
          v-model="form.notifyUrl"
          placeholder="选填，留空使用应用配置的支付回调地址"
          maxlength="255"
        />
        <div class="mt-1 text-xs text-g-600">
          渠道回调由系统接收；这里填的是业务侧接收支付结果通知的地址。
        </div>
      </ElFormItem>

      <ElFormItem label="有效期(分钟)" prop="expireMinutes">
        <ElInputNumber v-model="form.expireMinutes" :min="0" :max="1440" style="width: 100%" />
        <div class="mt-1 text-xs text-g-600">0 表示默认 30 分钟，超时后需重新下单。</div>
      </ElFormItem>

      <ElFormItem label="商品描述" prop="body">
        <ElInput
          v-model="form.body"
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
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">下单并调起</ElButton>
    </template>
  </ElDialog>

  <ElDialog v-model="resultVisible" title="支付调起参数" width="560px" align-center>
    <ElAlert type="info" :closable="false" :title="resultTip" show-icon />
    <div class="mt-4">
      <div class="mb-2 text-sm">支付单号：{{ created?.no }}</div>
      <ElInput :model-value="paramsText" type="textarea" :rows="8" readonly resize="none" />
    </div>
    <template #footer>
      <ElButton @click="copyParams">复制参数</ElButton>
      <ElButton type="primary" @click="resultVisible = false">完成</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchGetPayChannelList } from '@/api/pay/channel'
  import { fetchCreatePayOrder } from '@/api/pay/order'

  defineOptions({ name: 'PayOrderCreateDialog' })

  interface Props {
    visible: boolean
    appOptions?: Api.Pay.AppOption[]
  }

  const props = withDefaults(defineProps<Props>(), {
    appOptions: () => []
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
  const channelLoading = ref(false)
  const channels = ref<Api.Pay.ChannelItem[]>([])

  const resultVisible = ref(false)
  const created = ref<Api.Pay.OrderSubmitResult | null>(null)

  const form = reactive({
    appId: '',
    channelCode: '',
    userId: '',
    merchantOrderId: '',
    subject: '',
    priceYuan: 0.01 as number | null,
    openId: '',
    authCode: '',
    notifyUrl: '',
    expireMinutes: 30,
    body: ''
  })

  // 微信小程序 / 公众号支付需要 openid
  const isWechatJsapi = computed(
    () => form.channelCode === 'wx_lite' || form.channelCode === 'wx_pub'
  )

  // 支付宝条码支付需要用户付款码
  const isAlipayBar = computed(() => form.channelCode === 'alipay_bar')

  // 钱包余额是系统内置渠道，不需要渠道配置，但必须指定付款用户
  const isWallet = computed(() => form.channelCode === 'wallet')

  const rules: FormRules = {
    appId: [{ required: true, message: '请选择支付应用', trigger: 'change' }],
    channelCode: [{ required: true, message: '请选择支付渠道', trigger: 'change' }],
    userId: [
      {
        validator: (_rule, value: string, callback) => {
          if (isWallet.value && !value) {
            callback(new Error('钱包余额支付必须填写付款用户编号'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    merchantOrderId: [{ required: true, message: '请输入业务订单号', trigger: 'blur' }],
    subject: [{ required: true, message: '请输入商品标题', trigger: 'blur' }],
    priceYuan: [{ required: true, message: '请输入支付金额', trigger: 'change' }],
    openId: [
      {
        validator: (_rule, value: string, callback) => {
          if (isWechatJsapi.value && !value) {
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
          if (isAlipayBar.value && !value) {
            callback(new Error('支付宝条码支付必须填写付款码'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  // 渠道编码决定调起方式，提示信息随之变化
  const resultTip = computed(() => {
    const code = created.value?.channelCode ?? ''
    if (code === 'wx_native' || code === 'alipay_qr') {
      return '把二维码链接渲染成二维码给顾客扫码支付。'
    }
    if (code === 'wx_lite' || code === 'wx_pub') {
      return '把参数原样交给前端 wx.requestPayment 调起支付。'
    }
    if (code === 'alipay_app') {
      return '把 orderStr 交给支付宝客户端 SDK 调起支付。'
    }
    if (code === 'alipay_pc' || code === 'alipay_wap' || code === 'wx_wap') {
      return '在浏览器中打开支付链接完成付款。'
    }
    if (code === 'wallet') {
      return '已使用钱包余额支付完成，无需再调起收银台。'
    }

    return '按渠道要求把参数交给客户端调起支付。'
  })

  // 调起参数格式化展示
  const paramsText = computed(() => {
    if (!created.value) return ''

    return JSON.stringify(created.value.payParams ?? {}, null, 2)
  })

  // 拉取所选应用下已开启的渠道；钱包余额是内置渠道，不在 pay_channel 里，手动补一个选项
  const loadChannels = async (appId: string): Promise<void> => {
    channels.value = []
    form.channelCode = ''
    if (!appId) return

    channelLoading.value = true
    try {
      const res = await fetchGetPayChannelList({ appId, pageNo: 1, pageSize: 100 })
      channels.value = (res?.list ?? []).filter((channel) => channel.status === 0)
      channels.value.push({
        id: '0',
        appId,
        code: 'wallet',
        channelName: '钱包余额',
        status: 0,
        statusName: '开启',
        remark: '系统内置渠道',
        feeRate: 0,
        config: '{}'
      })
      form.channelCode = channels.value[0]?.code ?? ''
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      channelLoading.value = false
    }
  }

  // 切换应用后重新加载该应用的渠道
  const handleAppChange = (appId: string): void => {
    loadChannels(appId)
  }

  // 每次打开弹窗重置表单并加载默认应用的渠道
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      Object.assign(form, {
        appId: props.appOptions[0]?.id ?? '',
        channelCode: '',
        userId: '',
        merchantOrderId: '',
        subject: '',
        priceYuan: 0.01,
        openId: '',
        authCode: '',
        notifyUrl: '',
        expireMinutes: 30,
        body: ''
      })

      await loadChannels(form.appId)
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交下单：元转分后调用接口，成功后展示调起参数
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const price = Math.round((form.priceYuan || 0) * 100)
    if (price <= 0) {
      ElMessage.warning('支付金额必须大于 0')
      return
    }

    submitting.value = true
    try {
      const submitted = await fetchCreatePayOrder({
        appId: form.appId,
        channelCode: form.channelCode,
        userId: form.userId || undefined,
        merchantOrderId: form.merchantOrderId,
        subject: form.subject,
        body: form.body,
        price,
        notifyUrl: form.notifyUrl || undefined,
        openId: form.openId || undefined,
        authCode: form.authCode || undefined,
        expireMinutes: form.expireMinutes
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
