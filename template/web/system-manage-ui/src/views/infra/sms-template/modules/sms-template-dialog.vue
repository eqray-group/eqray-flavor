<!-- 短信模板新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑模板' : '新增模板'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="模板名" prop="name">
        <ElInput v-model="form.name" placeholder="如：会员登录验证码" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="模板代码" prop="code">
        <ElInput
          v-model="form.code"
          placeholder="如：member_sms_login"
          maxlength="64"
          :disabled="isEdit"
        />
      </ElFormItem>
      <ElFormItem label="短信渠道" prop="channelId">
        <ElSelect v-model="form.channelId" placeholder="请选择短信渠道" style="width: 100%">
          <ElOption
            v-for="item in channelOptions"
            :key="item.id"
            :label="`${item.name}（${item.code}）`"
            :value="item.id"
          />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="模板类型" prop="type">
        <ElRadioGroup v-model="form.type">
          <ElRadio :value="1">验证码</ElRadio>
          <ElRadio :value="2">通知</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="第三方模板 ID" prop="apiTemplateId">
        <ElInput
          v-model="form.apiTemplateId"
          placeholder="阿里云模板 Code / 腾讯云模板 ID"
          maxlength="128"
        />
      </ElFormItem>
      <ElFormItem label="模板内容" prop="content">
        <ElInput
          v-model="form.content"
          type="textarea"
          :rows="3"
          maxlength="512"
          show-word-limit
          placeholder="如：您的验证码是 {code}，5 分钟内有效"
        />
      </ElFormItem>
      <ElFormItem label="模板变量" prop="params">
        <ElInput v-model="form.params" placeholder="变量名，逗号分隔，如：code" maxlength="128" />
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
  import { fetchCreateSmsTemplate, fetchUpdateSmsTemplate } from '@/api/infra/sms-template'
  import { fetchGetSmsChannelOptions } from '@/api/infra/sms-channel'

  defineOptions({ name: 'SmsTemplateDialog' })

  type SmsTemplateItem = Api.Infra.SmsTemplateItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    templateData?: SmsTemplateItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    templateData: null
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
  const channelOptions = ref<Api.Infra.SmsChannelItem[]>([])

  const form = reactive({
    channelId: '',
    code: '',
    name: '',
    type: 1,
    apiTemplateId: '',
    content: '',
    params: '',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入模板名', trigger: 'blur' }],
    code: [{ required: true, message: '请输入模板代码', trigger: 'blur' }],
    channelId: [{ required: true, message: '请选择短信渠道', trigger: 'change' }],
    content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }]
  }

  const loadChannels = async (): Promise<void> => {
    try {
      channelOptions.value = await fetchGetSmsChannelOptions()
    } catch {
      channelOptions.value = []
    }
  }

  watch(
    () => [props.visible, props.templateData] as const,
    ([visible]) => {
      if (!visible) return

      loadChannels()
      Object.assign(form, {
        channelId: props.templateData?.channelId ?? '',
        code: props.templateData?.code ?? '',
        name: props.templateData?.name ?? '',
        type: props.templateData?.type ?? 1,
        apiTemplateId: props.templateData?.apiTemplateId ?? '',
        content: props.templateData?.content ?? '',
        params: props.templateData?.params ?? '',
        status: props.templateData?.status ?? 1,
        remark: props.templateData?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Infra.SmsTemplateSaveParams = {
      channelId: form.channelId,
      code: form.code.trim(),
      name: form.name.trim(),
      type: form.type,
      apiTemplateId: form.apiTemplateId.trim(),
      content: form.content.trim(),
      params: form.params.trim(),
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.templateData) {
        await fetchUpdateSmsTemplate(props.templateData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateSmsTemplate(payload)
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
