<!-- 充值套餐新增 / 编辑弹窗：实付金额与赠送金额均按元录入，提交时换算成分 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑充值套餐' : '新增充值套餐'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="套餐名" prop="name">
        <ElInput v-model="form.name" placeholder="如「充 100 送 10」" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="实付金额" prop="payYuan">
        <ElInputNumber
          v-model="form.payYuan"
          :min="0.01"
          :precision="2"
          :step="10"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="赠送金额" prop="bonusYuan">
        <ElInputNumber
          v-model="form.bonusYuan"
          :min="0"
          :precision="2"
          :step="10"
          style="width: 100%"
        />
        <div class="mt-1 text-xs text-g-600">
          到账金额 = 实付 + 赠送 = {{ formatYuan((form.payYuan || 0) + (form.bonusYuan || 0)) }} 元
        </div>
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="form.status">
          <ElRadio :value="0">开启</ElRadio>
          <ElRadio :value="1">关闭</ElRadio>
        </ElRadioGroup>
        <div class="mt-1 text-xs text-g-600"
          >关闭后会员端不再展示该档位，已充值的记录不受影响。</div
        >
      </ElFormItem>

      <ElFormItem label="描述" prop="description">
        <ElInput
          v-model="form.description"
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
  import { fetchCreatePayWalletPackage, fetchUpdatePayWalletPackage } from '@/api/pay/wallet'

  defineOptions({ name: 'PayWalletPackageDialog' })

  type PackageItem = Api.Pay.WalletPackageItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    packageData?: PackageItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    packageData: null
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
    payYuan: 100 as number | null,
    bonusYuan: 0 as number | null,
    status: 0,
    description: ''
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入套餐名', trigger: 'blur' }],
    payYuan: [{ required: true, message: '请输入实付金额', trigger: 'change' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  }

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 打开弹窗时重置表单，编辑时回填当前套餐
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      const data = props.packageData
      form.name = data?.name ?? ''
      form.payYuan = data ? data.payPrice / 100 : 100
      form.bonusYuan = data ? data.bonusPrice / 100 : 0
      form.status = data?.status ?? 0
      form.description = data?.description ?? ''

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交套餐：元转分后调用接口
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Pay.WalletPackageSaveParams = {
      name: form.name,
      payPrice: Math.round((form.payYuan || 0) * 100),
      bonusPrice: Math.round((form.bonusYuan || 0) * 100),
      status: form.status,
      description: form.description
    }
    if (payload.payPrice <= 0) {
      ElMessage.warning('实付金额必须大于 0')
      return
    }

    submitting.value = true
    try {
      if (isEdit.value && props.packageData) {
        await fetchUpdatePayWalletPackage(props.packageData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreatePayWalletPackage(payload)
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
