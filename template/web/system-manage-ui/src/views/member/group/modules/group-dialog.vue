<!-- 会员分组新增 / 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑分组' : '新增分组'"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="80px">
      <ElFormItem label="分组名" prop="name">
        <ElInput v-model="form.name" placeholder="如：企业客户" maxlength="30" />
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
  import { fetchCreateGroup, fetchUpdateGroup } from '@/api/member/group'

  defineOptions({ name: 'GroupDialog' })

  type GroupItem = Api.Member.GroupItem

  interface Props {
    visible: boolean
    type: 'add' | 'edit'
    groupData?: GroupItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    groupData: null
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
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    name: [
      { required: true, message: '请输入分组名', trigger: 'blur' },
      { max: 30, message: '分组名不能超过 30 个字符', trigger: 'blur' }
    ],
    remark: [{ max: 255, message: '备注不能超过 255 个字符', trigger: 'blur' }]
  }

  watch(
    () => [props.visible, props.groupData] as const,
    ([visible]) => {
      if (!visible) return

      Object.assign(form, {
        name: props.groupData?.name ?? '',
        status: props.groupData?.status ?? 1,
        remark: props.groupData?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.GroupSaveParams = {
      name: form.name.trim(),
      status: form.status,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.groupData) {
        await fetchUpdateGroup(props.groupData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateGroup(payload)
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
