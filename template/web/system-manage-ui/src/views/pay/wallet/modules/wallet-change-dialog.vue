<!-- 钱包调账弹窗：正数加款、负数扣款，后端写一条 manual 流水 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="调整余额"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      :title="`当前可用余额 ${formatYuan(props.wallet?.balance ?? 0)} 元`"
      show-icon
    />
    <div class="mt-1 text-xs text-g-600">
      正数加款、负数扣款；每次调整都会写入钱包流水，备注会一并保存。
    </div>

    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px" class="mt-4">
      <ElFormItem label="变动金额" prop="priceYuan">
        <ElInputNumber v-model="form.priceYuan" :precision="2" :step="1" style="width: 100%" />
      </ElFormItem>

      <ElFormItem label="调整原因" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="如：活动补偿、线下收款冲正"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确认调整</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchChangePayWalletBalance } from '@/api/pay/wallet'

  defineOptions({ name: 'PayWalletChangeDialog' })

  interface Props {
    visible: boolean
    wallet?: Api.Pay.WalletItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    wallet: null
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
    priceYuan: 0 as number | null,
    remark: ''
  })

  const rules: FormRules = {
    priceYuan: [
      {
        validator: (_rule, value: number | null, callback) => {
          if (!value) {
            callback(new Error('变动金额不能为 0'))
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

  // 打开弹窗时清空上一次输入
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.priceYuan = 0
      form.remark = ''
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交调账：元转分后调用接口，成功后由父页面刷新列表
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.wallet) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const price = Math.round((form.priceYuan || 0) * 100)
    if (price === 0) {
      ElMessage.warning('变动金额不能为 0')
      return
    }

    submitting.value = true
    try {
      await fetchChangePayWalletBalance(props.wallet.id, {
        price,
        remark: form.remark || undefined
      })
      ElMessage.success('调整成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
