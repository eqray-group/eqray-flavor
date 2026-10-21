<!-- 第三方登录渠道配置新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑配置' : '新增配置'"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="应用名" prop="name">
        <ElInput v-model="form.name" placeholder="如：小程序会员端" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="渠道" prop="socialType">
        <ElSelect
          v-model="form.socialType"
          placeholder="请选择渠道"
          :disabled="isEdit"
          style="width: 100%"
        >
          <ElOption label="微信小程序" value="wechat_mini" />
          <ElOption label="微信公众号" value="wechat_mp" />
          <ElOption label="微信开放平台" value="wechat_open" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="ClientID" prop="clientId">
        <ElInput v-model="form.clientId" placeholder="微信的 appid" maxlength="128" />
      </ElFormItem>
      <ElFormItem label="ClientSecret" prop="clientSecret">
        <ElInput
          v-model="form.clientSecret"
          type="password"
          show-password
          :placeholder="isEdit ? '留空表示不修改' : '微信的 appsecret'"
          maxlength="256"
        />
      </ElFormItem>
      <ElFormItem label="AgentID" prop="agentId">
        <ElInput v-model="form.agentId" placeholder="钉钉 / 企微等渠道使用，选填" maxlength="64" />
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
  import { fetchCreateSocialClient, fetchUpdateSocialClient } from '@/api/member/social-client'

  defineOptions({ name: 'SocialClientDialog' })

  type SocialClientItem = Api.Member.SocialClientItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    clientData?: SocialClientItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    clientData: null
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
    socialType: '',
    clientId: '',
    clientSecret: '',
    agentId: '',
    status: 1
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入应用名', trigger: 'blur' }],
    socialType: [{ required: true, message: '请选择渠道', trigger: 'change' }],
    clientId: [{ required: true, message: '请输入 ClientID', trigger: 'blur' }],
    clientSecret: [
      {
        validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
          if (!isEdit.value && !value) {
            callback(new Error('请输入 ClientSecret'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  watch(
    () => [props.visible, props.clientData] as const,
    ([visible]) => {
      if (!visible) return

      Object.assign(form, {
        name: props.clientData?.name ?? '',
        socialType: props.clientData?.socialType ?? '',
        clientId: props.clientData?.clientId ?? '',
        clientSecret: '',
        agentId: props.clientData?.agentId ?? '',
        status: props.clientData?.status ?? 1
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.SocialClientSaveParams = {
      name: form.name.trim(),
      socialType: form.socialType,
      clientId: form.clientId.trim(),
      clientSecret: form.clientSecret.trim(),
      agentId: form.agentId.trim(),
      status: form.status
    }

    submitting.value = true
    try {
      if (isEdit.value && props.clientData) {
        await fetchUpdateSocialClient(props.clientData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateSocialClient(payload)
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
