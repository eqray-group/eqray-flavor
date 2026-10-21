<!-- 会员签到档位新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑签到档位' : '新增签到档位'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="签到天数" prop="day">
        <ElInputNumber v-model="form.day" :min="1" :max="31" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">连续签到第几天，1 ~ 31，同一天只能配置一条</span>
      </ElFormItem>

      <ElFormItem label="奖励积分" prop="point">
        <ElInputNumber v-model="form.point" :min="0" :max="99999999" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">该档位发放的积分</span>
      </ElFormItem>

      <ElFormItem label="奖励经验" prop="experience">
        <ElInputNumber
          v-model="form.experience"
          :min="0"
          :max="99999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">该档位发放的成长经验</span>
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
  import { fetchCreateSignConfig, fetchUpdateSignConfig } from '@/api/member/sign-config'

  defineOptions({ name: 'SignConfigDialog' })

  type SignConfigItem = Api.Member.SignConfigItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    configData?: SignConfigItem | null
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

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    day: 1,
    point: 0,
    experience: 0,
    status: 1
  })

  const rules: FormRules = {
    day: [{ required: true, message: '请输入签到天数', trigger: 'change' }],
    point: [{ required: true, message: '请输入奖励积分', trigger: 'change' }],
    experience: [{ required: true, message: '请输入奖励经验', trigger: 'change' }]
  }

  // 每次打开弹窗时按当前数据重置表单
  watch(
    () => [props.visible, props.configData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.configData
      Object.assign(form, {
        day: data?.day ?? 1,
        point: data?.point ?? 0,
        experience: data?.experience ?? 0,
        status: data?.status ?? 1
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.SignConfigSaveParams = {
      day: form.day,
      point: form.point,
      experience: form.experience,
      status: form.status
    }

    submitting.value = true
    try {
      if (isEdit.value && props.configData) {
        await fetchUpdateSignConfig(props.configData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateSignConfig(payload)
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
