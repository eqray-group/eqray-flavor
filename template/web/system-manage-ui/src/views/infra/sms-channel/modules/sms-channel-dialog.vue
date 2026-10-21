<!-- 短信渠道新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑渠道' : '新增渠道'"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="渠道名" prop="name">
        <ElInput v-model="form.name" placeholder="如：腾讯云正式环境" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="渠道代码" prop="code">
        <ElSelect
          v-model="form.code"
          placeholder="请选择渠道代码"
          :disabled="isEdit"
          style="width: 100%"
        >
          <ElOption label="本地日志（开发）" value="DebugLog" />
          <ElOption label="阿里云" value="AliyunSms" />
          <ElOption label="腾讯云" value="TencentSms" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem v-if="form.code !== 'DebugLog'" label="APIKey" prop="apiKey">
        <ElInput
          v-model="form.apiKey"
          placeholder="阿里云 AccessKeyId / 腾讯云 SecretId"
          maxlength="128"
        />
      </ElFormItem>
      <ElFormItem v-if="form.code !== 'DebugLog'" label="APISecret" prop="apiSecret">
        <ElInput
          v-model="form.apiSecret"
          type="password"
          show-password
          :placeholder="isEdit ? '留空表示不修改' : '阿里云 AccessKeySecret / 腾讯云 SecretKey'"
          maxlength="256"
        />
      </ElFormItem>
      <ElFormItem v-if="form.code !== 'DebugLog'" label="短信签名" prop="signature">
        <ElInput v-model="form.signature" placeholder="已在短信平台申请的签名" maxlength="64" />
      </ElFormItem>
      <ElFormItem v-if="form.code === 'TencentSms'" label="扩展配置" prop="extra">
        <ElInput
          v-model="form.extra"
          type="textarea"
          :rows="2"
          maxlength="512"
          show-word-limit
          placeholder='{"region":"ap-guangzhou","sdkAppId":"1400000000"}'
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
  import { fetchCreateSmsChannel, fetchUpdateSmsChannel } from '@/api/infra/sms-channel'

  defineOptions({ name: 'SmsChannelDialog' })

  type SmsChannelItem = Api.Infra.SmsChannelItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    channelData?: SmsChannelItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    channelData: null
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
  const isDebugLog = computed(() => form.code === 'DebugLog')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    code: 'DebugLog',
    apiKey: '',
    apiSecret: '',
    signature: '',
    extra: '',
    status: 1,
    remark: ''
  })

  const rules = computed<FormRules>(() => ({
    name: [{ required: true, message: '请输入渠道名', trigger: 'blur' }],
    code: [{ required: true, message: '请选择渠道代码', trigger: 'change' }],
    apiKey: isDebugLog.value ? [] : [{ required: true, message: '请输入 APIKey', trigger: 'blur' }],
    apiSecret: isDebugLog.value
      ? []
      : [
          {
            validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
              if (!isEdit.value && !value) {
                callback(new Error('请输入 APISecret'))
                return
              }
              callback()
            },
            trigger: 'blur'
          }
        ],
    signature: isDebugLog.value
      ? []
      : [{ required: true, message: '请输入短信签名', trigger: 'blur' }]
  }))

  watch(
    () => [props.visible, props.channelData] as const,
    ([visible]) => {
      if (!visible) return

      Object.assign(form, {
        name: props.channelData?.name ?? '',
        code: props.channelData?.code ?? 'DebugLog',
        apiKey: props.channelData?.apiKey ?? '',
        apiSecret: '',
        signature: props.channelData?.signature ?? '',
        extra: props.channelData?.extra ?? '',
        status: props.channelData?.status ?? 1,
        remark: props.channelData?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Infra.SmsChannelSaveParams = {
      name: form.name.trim(),
      code: form.code,
      apiKey: form.apiKey.trim(),
      apiSecret: form.apiSecret.trim(),
      signature: form.signature.trim(),
      extra: form.extra.trim(),
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.channelData) {
        await fetchUpdateSmsChannel(props.channelData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateSmsChannel(payload)
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
