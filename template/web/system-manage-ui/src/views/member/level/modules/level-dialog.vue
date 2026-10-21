<!-- 会员等级新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑会员等级' : '新增会员等级'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="等级序号" prop="level">
        <ElInputNumber
          v-model="form.level"
          :min="1"
          :max="99"
          :disabled="isEdit"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">
          {{ isEdit ? '创建后不可修改' : '1 ~ 99，数字越大等级越高' }}
        </span>
      </ElFormItem>

      <ElFormItem label="等级名称" prop="name">
        <ElInput v-model="form.name" placeholder="如：金卡会员" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="升级经验" prop="experience">
        <ElInputNumber
          v-model="form.experience"
          :min="0"
          :max="99999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">会员经验达到该值时升级</span>
      </ElFormItem>

      <ElFormItem label="折扣" prop="discountPercent">
        <ElInputNumber
          v-model="form.discountPercent"
          :min="1"
          :max="100"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">百分比，100 表示无折扣</span>
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">越小越靠前</span>
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

      <ElFormItem label="等级说明" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="选填，描述该等级可享受的权益"
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
  import { fetchCreateLevel, fetchUpdateLevel } from '@/api/member/level'

  defineOptions({ name: 'LevelDialog' })

  type LevelItem = Api.Member.LevelItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    levelData?: LevelItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    levelData: null
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
    level: 1,
    name: '',
    experience: 0,
    discountPercent: 100,
    sort: 0,
    status: 1,
    description: ''
  })

  const rules: FormRules = {
    level: [{ required: true, message: '请输入等级序号', trigger: 'blur' }],
    name: [
      { required: true, message: '请输入等级名称', trigger: 'blur' },
      { max: 64, message: '等级名称不能超过 64 个字符', trigger: 'blur' }
    ],
    discountPercent: [{ required: true, message: '请输入折扣百分比', trigger: 'blur' }],
    description: [{ max: 500, message: '等级说明不能超过 500 个字符', trigger: 'blur' }]
  }

  // 每次打开弹窗时按当前数据重置表单
  watch(
    () => [props.visible, props.levelData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.levelData
      Object.assign(form, {
        level: data?.level ?? 1,
        name: data?.name ?? '',
        experience: data?.experience ?? 0,
        discountPercent: data?.discountPercent ?? 100,
        sort: data?.sort ?? 0,
        status: data?.status ?? 1,
        description: data?.description ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.LevelSaveParams = {
      level: form.level,
      name: form.name.trim(),
      experience: form.experience,
      discountPercent: form.discountPercent,
      sort: form.sort,
      status: form.status,
      description: form.description.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.levelData) {
        await fetchUpdateLevel(props.levelData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateLevel(payload)
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
