<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isAdd ? '新增用户' : '编辑用户'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="80px">
      <ElFormItem label="账号" prop="username">
        <ElInput v-model="formData.username" placeholder="请输入账号" maxlength="64" />
      </ElFormItem>
      <ElFormItem v-if="isAdd" label="密码" prop="password">
        <ElInput
          v-model="formData.password"
          type="password"
          show-password
          placeholder="请输入初始密码"
          maxlength="72"
        />
      </ElFormItem>
      <ElFormItem label="昵称" prop="nickname">
        <ElInput v-model="formData.nickname" placeholder="不填默认与账号相同" maxlength="64" />
      </ElFormItem>
      <ElFormItem label="手机号" prop="mobile">
        <ElInput v-model="formData.mobile" placeholder="选填" maxlength="20" />
      </ElFormItem>
      <ElFormItem label="邮箱" prop="email">
        <ElInput v-model="formData.email" placeholder="选填" maxlength="128" />
      </ElFormItem>
      <ElFormItem label="性别" prop="sex">
        <ElRadioGroup v-model="formData.sex">
          <ElRadio :value="0">未知</ElRadio>
          <ElRadio :value="1">男</ElRadio>
          <ElRadio :value="2">女</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="formData.status"
          :active-value="1"
          :inactive-value="0"
          active-text="启用"
          inactive-text="禁用"
        />
      </ElFormItem>
      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="formData.remark"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="选填"
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreateUser, fetchUpdateUser } from '@/api/system/user'
  import type { DialogType } from '@/types'

  interface Props {
    visible: boolean
    type: DialogType
    userData?: Partial<Api.SystemManage.UserListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const isAdd = computed(() => props.type === 'add')
  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  // 表单只暴露可编辑字段；avatar 随行数据回传，避免全量更新时被清空
  const formData = reactive({
    username: '',
    password: '',
    nickname: '',
    mobile: '',
    email: '',
    sex: 0,
    status: 1,
    remark: '',
    avatar: ''
  })

  /** 空值跳过校验的格式规则 */
  const optionalPattern = (pattern: RegExp, message: string) => ({
    validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
      if (!value || pattern.test(value)) {
        callback()
        return
      }
      callback(new Error(message))
    },
    trigger: 'blur'
  })

  const rules = computed<FormRules>(() => ({
    username: [
      { required: true, message: '请输入账号', trigger: 'blur' },
      { max: 64, message: '账号不能超过 64 个字符', trigger: 'blur' }
    ],
    password: isAdd.value
      ? [
          { required: true, message: '请输入初始密码', trigger: 'blur' },
          { min: 6, max: 72, message: '密码长度需在 6 ~ 72 个字符', trigger: 'blur' }
        ]
      : [],
    nickname: [{ max: 64, message: '昵称不能超过 64 个字符', trigger: 'blur' }],
    mobile: [optionalPattern(/^1[3-9]\d{9}$/, '手机号格式不正确')],
    email: [optionalPattern(/^[^\s@]+@[^\s@]+\.[^\s@]+$/, '邮箱格式不正确')],
    remark: [{ max: 500, message: '备注不能超过 500 个字符', trigger: 'blur' }]
  }))

  /** 用行数据回填表单，新增时使用默认值 */
  const initFormData = () => {
    const row = props.type === 'edit' ? props.userData : undefined

    Object.assign(formData, {
      username: row?.username ?? '',
      password: '',
      nickname: row?.nickname ?? '',
      mobile: row?.mobile ?? '',
      email: row?.email ?? '',
      sex: row?.sex ?? 0,
      status: row?.status ?? 1,
      remark: row?.remark ?? '',
      avatar: row?.avatar ?? ''
    })
  }

  watch(
    () => [props.visible, props.type, props.userData],
    ([visible]) => {
      if (visible) {
        initFormData()
        nextTick(() => formRef.value?.clearValidate())
      }
    },
    { immediate: true }
  )

  /** 提交表单：新增走 POST，编辑走 PUT（编辑不改密码） */
  const handleSubmit = async () => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.SystemManage.UserSaveParams = {
      username: formData.username.trim(),
      nickname: formData.nickname.trim(),
      mobile: formData.mobile.trim(),
      email: formData.email.trim(),
      sex: formData.sex,
      status: formData.status,
      remark: formData.remark.trim(),
      avatar: formData.avatar
    }

    submitting.value = true
    try {
      if (isAdd.value) {
        await fetchCreateUser({ ...payload, password: formData.password })
        ElMessage.success('新增成功')
      } else {
        await fetchUpdateUser(String(props.userData?.id ?? ''), payload)
        ElMessage.success('保存成功')
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
