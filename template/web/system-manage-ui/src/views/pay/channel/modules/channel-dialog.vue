<!-- 支付渠道新增 / 编辑弹窗：维护所属应用、渠道编码、费率与渠道参数 JSON -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑支付渠道' : '新增支付渠道'"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="所属应用" prop="appId">
        <ElSelect
          v-model="form.appId"
          class="w-full"
          filterable
          :disabled="isEdit"
          placeholder="请选择支付应用"
          no-data-text="请先创建支付应用"
        >
          <ElOption
            v-for="app in props.appOptions"
            :key="app.id"
            :label="app.name"
            :value="app.id"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600">渠道创建后不能跨应用移动。</div>
      </ElFormItem>

      <ElFormItem label="渠道编码" prop="code">
        <ElSelect
          v-model="form.code"
          class="w-full"
          filterable
          :disabled="isEdit"
          placeholder="请选择渠道"
          @change="handleCodeChange"
        >
          <ElOption
            v-for="item in channelOptions"
            :key="item.code"
            :label="`${item.name}（${item.code}）`"
            :value="item.code"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600">同一应用下同一渠道只能配置一次。</div>
      </ElFormItem>

      <ElFormItem label="渠道费率" prop="feeRate">
        <ElInputNumber
          v-model="form.feeRate"
          :min="0"
          :max="100"
          :precision="3"
          :step="0.1"
          style="width: 100%"
        />
        <div class="mt-1 text-xs text-g-600"
          >单位百分比，用于计算渠道手续费，不影响用户支付金额。</div
        >
      </ElFormItem>

      <ElFormItem label="渠道配置" prop="config">
        <ElInput
          v-model="form.config"
          type="textarea"
          :rows="8"
          maxlength="8192"
          show-word-limit
          placeholder="JSON 对象，字段随渠道不同"
        />
        <div class="mt-1 text-xs text-g-600">{{ configHint }}</div>
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="form.status">
          <ElRadio :value="0">开启</ElRadio>
          <ElRadio :value="1">关闭</ElRadio>
        </ElRadioGroup>
        <div class="mt-1 text-xs text-g-600">关闭后该渠道无法发起支付，历史支付单不受影响。</div>
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
  import {
    fetchCreatePayChannel,
    fetchGetPayChannelOptions,
    fetchUpdatePayChannel
  } from '@/api/pay/channel'

  defineOptions({ name: 'PayChannelDialog' })

  type ChannelItem = Api.Pay.ChannelItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    channelData?: ChannelItem | null
    appOptions?: Api.Pay.AppOption[]
  }

  const props = withDefaults(defineProps<Props>(), {
    channelData: null,
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

  const isEdit = computed(() => props.type === 'edit')

  const channelOptions = ref<Api.Pay.ChannelOption[]>([])
  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    appId: '',
    code: '',
    feeRate: 0.6,
    config: '',
    status: 0,
    remark: ''
  })

  // 各渠道配置项的填写提示，避免后台填错键名导致下单失败
  const configHint = computed(() => {
    if (form.code.startsWith('wx_')) {
      return '微信必填：appId、mchId、apiV3Key、serialNo、privateKey；回调验签另需 publicKey 与 publicKeyId。'
    }
    if (form.code.startsWith('alipay_')) {
      return '支付宝必填：appId、privateKey；回调验签必填 alipayPublicKey；isProd 为 true 时走正式网关。'
    }
    if (form.code === 'wallet') {
      return '钱包余额是系统内置渠道，无需配置。'
    }

    return '模拟支付无需配置，留 {} 即可，仅用于本地联调。'
  })

  const rules: FormRules = {
    appId: [{ required: true, message: '请选择所属应用', trigger: 'change' }],
    code: [{ required: true, message: '请选择渠道编码', trigger: 'change' }],
    feeRate: [{ required: true, message: '请填写渠道费率', trigger: 'change' }],
    config: [
      { required: true, message: '请填写渠道配置', trigger: 'blur' },
      {
        validator: (_rule, value: string, callback) => {
          if (!value) {
            callback()
            return
          }
          try {
            const parsed = JSON.parse(value)
            if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
              callback(new Error('渠道配置必须是 JSON 对象'))
              return
            }
          } catch {
            callback(new Error('渠道配置不是合法的 JSON'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  // 渠道编码切换时，若是新建且配置为空，填入该渠道的最小配置骨架
  const handleCodeChange = (code: string): void => {
    if (isEdit.value || form.config.trim() !== '') return

    // 钱包余额是系统内置渠道，后端不允许配置，提前拦住并说明
    if (code === 'wallet') {
      form.config = ''
      ElMessage.warning('钱包余额是系统内置渠道，无需配置，请选择其它渠道')
      return
    }

    if (code.startsWith('wx_')) {
      form.config = JSON.stringify(
        {
          appId: '',
          mchId: '',
          apiV3Key: '',
          serialNo: '',
          privateKey: '',
          publicKey: '',
          publicKeyId: ''
        },
        null,
        2
      )
      return
    }
    if (code.startsWith('alipay_')) {
      form.config = JSON.stringify(
        { appId: '', privateKey: '', alipayPublicKey: '', isProd: true },
        null,
        2
      )
      return
    }
    form.config = '{}'
  }

  // 加载后端支持的渠道编码清单；钱包余额为内置渠道，不出现在可配置列表里
  const loadChannelOptions = async (): Promise<void> => {
    try {
      const options = (await fetchGetPayChannelOptions()) ?? []
      channelOptions.value = options.filter((item) => item.code !== 'wallet')
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 每次打开弹窗重置表单；编辑时回填渠道配置（密钥字段为掩码）
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.channelData
      Object.assign(form, {
        appId: data?.appId ?? props.appOptions[0]?.id ?? '',
        code: data?.code ?? '',
        feeRate: data?.feeRate ?? 0.6,
        config: data?.config ?? '',
        status: data?.status ?? 0,
        remark: data?.remark ?? ''
      })

      await loadChannelOptions()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交支付渠道；编辑时应用与编码由后端沿用原值
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Pay.ChannelSaveParams = {
      appId: form.appId,
      code: form.code,
      status: form.status,
      remark: form.remark,
      feeRate: form.feeRate,
      config: form.config
    }

    submitting.value = true
    try {
      if (isEdit.value && props.channelData) {
        await fetchUpdatePayChannel(props.channelData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreatePayChannel(payload)
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
