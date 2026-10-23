<!-- 加料组新增 / 编辑弹窗：维护名称、必选规则、选择数量与启停状态 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑加料组' : '新增加料组'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="104px">
      <ElFormItem label="加料组名称" prop="name">
        <ElInput v-model="form.name" placeholder="如加料、甜度、温度" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="是否必选" prop="required">
        <ElRadioGroup v-model="form.required">
          <ElRadio :value="0">可选</ElRadio>
          <ElRadio :value="1">必选</ElRadio>
        </ElRadioGroup>
      </ElFormItem>

      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="最少选择" prop="minSelect">
          <ElInputNumber v-model="form.minSelect" :min="0" :max="99" style="width: 100%" />
        </ElFormItem>

        <ElFormItem label="最多选择" prop="maxSelect">
          <ElInputNumber v-model="form.maxSelect" :min="0" :max="99" style="width: 100%" />
        </ElFormItem>
      </div>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" style="width: 100%" />
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="form.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
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

      <div class="-mt-2 pl-27 text-xs text-g-600">
        加料组名称品牌级唯一；必选时最少选择数至少为 1，最多选择数填 0 表示不限。
      </div>
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
  import { fetchCreateAddonGroup, fetchUpdateAddonGroup } from '@/api/dish/addon'

  defineOptions({ name: 'DishAddonDialog' })

  type AddonGroupItem = Api.Dish.AddonGroupItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    addonData?: AddonGroupItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    addonData: null
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
    required: 0,
    minSelect: 0,
    maxSelect: 0,
    sort: 0,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入加料组名称', trigger: 'blur' }]
  }

  // 每次打开弹窗回填数据并清除校验状态
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      const data = props.addonData
      Object.assign(form, {
        name: data?.name ?? '',
        required: data?.required ?? 0,
        minSelect: data?.minSelect ?? 0,
        maxSelect: data?.maxSelect ?? 0,
        sort: data?.sort ?? 0,
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交加料组：必选时最少选择数至少为 1，最多选择数不能小于最少选择数
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    if (form.required === 1 && form.minSelect < 1) {
      ElMessage.error('必选加料组的最少选择数至少为 1')
      return
    }

    if (form.maxSelect > 0 && form.maxSelect < form.minSelect) {
      ElMessage.error('最多选择数不能小于最少选择数')
      return
    }

    const payload: Api.Dish.AddonGroupSaveParams = {
      name: form.name.trim(),
      required: form.required,
      minSelect: form.minSelect,
      maxSelect: form.maxSelect,
      sort: form.sort,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.addonData) {
        await fetchUpdateAddonGroup(props.addonData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateAddonGroup(payload)
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
