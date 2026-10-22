<!-- 门店员工新增 / 编辑弹窗：选择可见门店、后台账号与门店角色 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑门店员工' : '新增门店员工'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="96px">
      <ElFormItem label="所属门店" prop="storeId">
        <ElSelect
          v-model="form.storeId"
          class="w-full"
          filterable
          :disabled="!!storeContext.currentStoreId"
          :loading="storeLoading"
          placeholder="请选择门店"
          no-data-text="暂无可分配门店"
        >
          <ElOption
            v-for="store in storeOptions"
            :key="store.id"
            :label="store.name"
            :value="store.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="后台账号" prop="userId">
        <ElSelect
          v-model="form.userId"
          class="w-full"
          filterable
          remote
          reserve-keyword
          :remote-method="searchUsers"
          :loading="userLoading"
          placeholder="输入账号 / 昵称 / 手机号搜索"
          no-data-text="没有匹配的启用账号"
        >
          <ElOption
            v-for="user in userOptions"
            :key="user.id"
            :label="displayUser(user)"
            :value="user.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="门店角色" prop="role">
        <ElSelect v-model="form.role" class="w-full">
          <ElOption
            v-for="role in roleOptions"
            :key="role.value"
            :label="role.label"
            :value="role.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          active-text="启用"
          inactive-text="禁用"
        />
        <span class="ml-3 text-xs text-g-600">禁用后该账号不再通过此门店获得可见权限</span>
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
  import {
    fetchCreateEmployee,
    fetchSearchEmployeeUsers,
    fetchUpdateEmployee
  } from '@/api/store/employee'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'EmployeeDialog' })

  type EmployeeItem = Api.Store.EmployeeItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    employeeData?: EmployeeItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    employeeData: null
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
  const storeContext = useStoreContextStore()

  const roleOptions = [
    { label: '店长', value: 1 },
    { label: '收银', value: 2 },
    { label: '后厨', value: 3 },
    { label: '服务员', value: 4 },
    { label: '骑手', value: 5 }
  ]

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const storeLoading = ref(false)
  const userLoading = ref(false)
  const storeOptions = ref<Api.Store.VisibleStore[]>([])
  const userOptions = ref<Api.Store.EmployeeUserOption[]>([])

  const form = reactive({
    storeId: '',
    userId: '',
    role: 4,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    userId: [{ required: true, message: '请选择后台账号', trigger: 'change' }],
    role: [{ required: true, message: '请选择门店角色', trigger: 'change' }]
  }

  // 账号下拉展示昵称与登录账号，手机号作为辅助信息
  const displayUser = (user: Api.Store.EmployeeUserOption): string => {
    const name = user.nickname ? `${user.nickname}（${user.username}）` : user.username
    return user.mobile ? `${name} · ${user.mobile}` : name
  }

  // 当前编辑记录使用的账号选项，远程搜索结果没有命中时兜底插入
  const currentUserOption = computed<Api.Store.EmployeeUserOption | null>(() => {
    if (!isEdit.value || !props.employeeData) return null
    return {
      id: props.employeeData.userId,
      username: props.employeeData.username,
      nickname: props.employeeData.nickname,
      mobile: props.employeeData.mobile,
      status: 1
    }
  })

  // 拉取当前用户可见门店；选定具体门店时只允许在该门店下分配
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      const contextStoreID = storeContext.currentStoreId

      if (contextStoreID) {
        // 请求头已锁定门店，弹窗不能再选其他门店，否则后端会报上下文不一致
        list = list.filter((store) => store.id === contextStoreID)
      } else if (
        isEdit.value &&
        props.employeeData &&
        !list.some((store) => store.id === props.employeeData?.storeId)
      ) {
        list.unshift({
          id: props.employeeData.storeId,
          name: props.employeeData.storeName || '当前门店',
          code: '',
          status: 1
        })
      }
      storeOptions.value = list

      const storeExists = list.some((store) => store.id === form.storeId)
      if (!storeExists) {
        form.storeId = list[0]?.id ?? ''
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      storeLoading.value = false
    }
  }

  // 远程搜索启用账号；编辑时保证当前账号始终在选项中
  const searchUsers = async (keyword: string): Promise<void> => {
    userLoading.value = true
    try {
      const list = await fetchSearchEmployeeUsers(keyword)
      const current = currentUserOption.value
      if (current && !list.some((user) => user.id === current.id)) {
        list.unshift(current)
      }
      userOptions.value = list
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      userLoading.value = false
    }
  }

  // 每次打开弹窗重置表单并加载门店、账号选项
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.employeeData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        userId: data?.userId ?? '',
        role: data?.role ?? 4,
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []
      userOptions.value = []

      await Promise.all([loadStores(), searchUsers('')])
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交员工关联；成功后关闭弹窗并通知父页面刷新
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Store.EmployeeSaveParams = {
      storeId: form.storeId,
      userId: form.userId,
      role: form.role,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.employeeData) {
        await fetchUpdateEmployee(props.employeeData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateEmployee(payload)
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
