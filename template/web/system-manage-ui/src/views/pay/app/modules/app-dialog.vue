<!-- 支付应用新增 / 编辑弹窗：维护应用名、回调地址与可选的应用密钥 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑支付应用' : '新增支付应用'"
    width="620px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="应用名" prop="name">
        <ElInput v-model="form.name" placeholder="如品牌小程序、门店 POS" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="支付回调地址" prop="orderNotifyUrl">
        <ElInput
          v-model="form.orderNotifyUrl"
          placeholder="https://your-domain/api/notify/pay"
          maxlength="255"
        />
        <div class="mt-1 text-xs text-g-600">
          接收本系统推送的支付结果；渠道回调由系统统一接收后再转发到这里。
        </div>
      </ElFormItem>

      <ElFormItem label="退款回调地址" prop="refundNotifyUrl">
        <ElInput
          v-model="form.refundNotifyUrl"
          placeholder="https://your-domain/api/notify/refund"
          maxlength="255"
        />
      </ElFormItem>

      <ElFormItem label="转账回调地址" prop="transferNotifyUrl">
        <ElInput
          v-model="form.transferNotifyUrl"
          placeholder="选填，未开通转账时留空"
          maxlength="255"
        />
      </ElFormItem>

      <ElFormItem v-if="isEdit" label="应用标识">
        <ElInput :model-value="props.appData?.appKey || ''" disabled />
        <div class="mt-1 text-xs text-g-600">应用标识创建后不可修改。</div>
      </ElFormItem>

      <ElFormItem label="应用密钥" prop="appSecret">
        <ElInput
          v-model="form.appSecret"
          type="password"
          show-password
          autocomplete="new-password"
          :placeholder="isEdit ? '留空表示不修改原密钥' : '留空由系统生成'"
          maxlength="64"
        />
        <div class="mt-1 text-xs text-g-600">
          参与回调签名校验；新增时留空由系统生成，明文只在创建成功时展示一次。
        </div>
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="form.status">
          <ElRadio :value="0">开启</ElRadio>
          <ElRadio :value="1">关闭</ElRadio>
        </ElRadioGroup>
        <div class="mt-1 text-xs text-g-600">关闭后该应用下的渠道无法发起支付。</div>
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

  <ElDialog v-model="secretVisible" title="应用密钥" width="520px" align-center>
    <ElAlert
      type="warning"
      :closable="false"
      title="密钥只展示这一次，请立即保存到安全的位置"
      show-icon
    />
    <ElInput class="mt-4" :model-value="createdSecret" readonly />
    <template #footer>
      <ElButton type="primary" @click="secretVisible = false">我已保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreatePayApp, fetchUpdatePayApp } from '@/api/pay/app'

  defineOptions({ name: 'PayAppDialog' })

  type AppItem = Api.Pay.AppItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    appData?: AppItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    appData: null
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
  const secretVisible = ref(false)
  const createdSecret = ref('')

  const form = reactive({
    name: '',
    orderNotifyUrl: '',
    refundNotifyUrl: '',
    transferNotifyUrl: '',
    appSecret: '',
    status: 0,
    remark: ''
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入应用名', trigger: 'blur' }],
    orderNotifyUrl: [{ required: true, message: '请输入支付回调地址', trigger: 'blur' }],
    refundNotifyUrl: [{ required: true, message: '请输入退款回调地址', trigger: 'blur' }],
    status: [{ required: true, message: '请选择应用状态', trigger: 'change' }]
  }

  // 每次打开弹窗重置表单，编辑时回填当前应用
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      const data = props.appData
      Object.assign(form, {
        name: data?.name ?? '',
        orderNotifyUrl: data?.orderNotifyUrl ?? '',
        refundNotifyUrl: data?.refundNotifyUrl ?? '',
        transferNotifyUrl: data?.transferNotifyUrl ?? '',
        appSecret: '',
        status: data?.status ?? 0,
        remark: data?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交支付应用；新增成功后弹一次明文密钥
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Pay.AppSaveParams = {
      name: form.name,
      orderNotifyUrl: form.orderNotifyUrl,
      refundNotifyUrl: form.refundNotifyUrl,
      transferNotifyUrl: form.transferNotifyUrl,
      appSecret: form.appSecret || undefined,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.appData) {
        await fetchUpdatePayApp(props.appData.id, payload)
        ElMessage.success('保存成功')
      } else {
        const created = await fetchCreatePayApp(payload)
        ElMessage.success('新增成功')
        if (created?.appSecret) {
          createdSecret.value = created.appSecret
          secretVisible.value = true
        }
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
