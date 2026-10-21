<!-- 参数配置新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑参数' : '新增参数'"
    width="580px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
      <ElFormItem label="参数分组" prop="category">
        <ElSelect
          v-model="form.category"
          filterable
          allow-create
          default-first-option
          placeholder="选择或输入分组"
        >
          <ElOption v-for="item in categoryOptions" :key="item" :label="item" :value="item" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="参数名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入参数名称" maxlength="128" />
      </ElFormItem>

      <ElFormItem label="参数键" prop="configKey">
        <ElInput
          v-model="form.configKey"
          :disabled="isEdit"
          placeholder="如：infra.upload.max_size"
          maxlength="128"
        />
      </ElFormItem>

      <ElFormItem label="参数值" prop="value">
        <ElInput
          v-model="form.value"
          type="textarea"
          :rows="3"
          maxlength="512"
          show-word-limit
          placeholder="请输入参数值"
        />
      </ElFormItem>

      <ElFormItem label="参数类型" prop="type">
        <ElSelect v-model="form.type" :disabled="isEdit" placeholder="请选择类型">
          <ElOption label="系统内置" :value="1" />
          <ElOption label="自定义" :value="2" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="是否公开" prop="visible">
        <ElSwitch v-model="form.visible" :active-value="1" :inactive-value="0" />
        <span class="ml-2 text-xs text-g-600">公开后允许前端直接读取</span>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="2"
          maxlength="500"
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
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElDialog, ElMessage } from 'element-plus'
  import { fetchCreateConfig, fetchUpdateConfig } from '@/api/infra/config'

  defineOptions({ name: 'ConfigDialog' })

  type ConfigItem = Api.Infra.ConfigItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    configData?: ConfigItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    configData: null
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

  const categoryOptions = ['文件', '登录', '通用', '会员', '支付']

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    category: '通用',
    name: '',
    configKey: '',
    value: '',
    type: 2,
    visible: 0,
    remark: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入参数名称', trigger: 'blur' },
      { max: 128, message: '参数名称不能超过 128 个字符', trigger: 'blur' }
    ],
    configKey: [
      { required: true, message: '请输入参数键', trigger: 'blur' },
      { max: 128, message: '参数键不能超过 128 个字符', trigger: 'blur' }
    ],
    value: [{ max: 512, message: '参数值不能超过 512 个字符', trigger: 'blur' }],
    remark: [{ max: 500, message: '备注不能超过 500 个字符', trigger: 'blur' }]
  }

  // 每次打开弹窗时按当前数据重置表单
  watch(
    () => [props.visible, props.configData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.configData
      Object.assign(form, {
        category: data?.category ?? '通用',
        name: data?.name ?? '',
        configKey: data?.configKey ?? '',
        value: data?.value ?? '',
        type: data?.type ?? 2,
        visible: data?.visible ?? 0,
        remark: data?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      if (isEdit.value && props.configData) {
        await fetchUpdateConfig(props.configData.id, {
          category: form.category,
          name: form.name.trim(),
          value: form.value,
          visible: form.visible,
          remark: form.remark.trim()
        })
        ElMessage.success('保存成功')
      } else {
        await fetchCreateConfig({
          category: form.category,
          name: form.name.trim(),
          configKey: form.configKey.trim(),
          value: form.value,
          type: form.type,
          visible: form.visible,
          remark: form.remark.trim()
        })
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
