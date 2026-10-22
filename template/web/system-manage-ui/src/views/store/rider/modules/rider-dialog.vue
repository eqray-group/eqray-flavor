<!-- 骑手新增 / 编辑弹窗：自动开通账号或绑定已有账号，并维护接单上限与状态 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑骑手' : '新增骑手'"
    width="620px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="所属门店" prop="storeId">
        <ElSelect
          v-model="form.storeId"
          class="w-full"
          filterable
          :disabled="isEdit || !!storeContext.currentStoreId"
          :loading="storeLoading"
          placeholder="请选择门店"
          no-data-text="暂无可操作门店"
        >
          <ElOption
            v-for="store in storeOptions"
            :key="store.id"
            :label="store.name"
            :value="store.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem v-if="isEdit" label="骑手账号">
        <ElInput :model-value="displayRider" disabled />
      </ElFormItem>

      <template v-else>
        <ElFormItem label="账号方式">
          <ElRadioGroup v-model="form.accountMode">
            <ElRadioButton value="auto">自动开通</ElRadioButton>
            <ElRadioButton value="existing">绑定已有账号</ElRadioButton>
          </ElRadioGroup>
        </ElFormItem>

        <template v-if="form.accountMode === 'auto'">
          <ElFormItem label="骑手姓名" prop="name">
            <ElInput v-model="form.name" placeholder="用于账号昵称与列表展示" maxlength="64" />
          </ElFormItem>
          <ElFormItem label="手机号" prop="mobile">
            <ElInput v-model="form.mobile" placeholder="作为骑手登录账号" maxlength="11" />
          </ElFormItem>
        </template>

        <ElFormItem v-else label="后台账号" prop="userId">
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
      </template>

      <ElFormItem label="接单上限" prop="maxOrders">
        <ElInputNumber v-model="form.maxOrders" :min="1" :max="20" style="width: 100%" />
        <span class="ml-3 text-xs text-g-600">骑手手上进行中的配送单不能超过该数量</span>
      </ElFormItem>

      <ElFormItem label="骑手状态" prop="status">
        <ElSelect v-model="form.status" class="w-full">
          <ElOption
            v-for="status in statusOptions"
            :key="status.value"
            :label="status.label"
            :value="status.value"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600">
          「接单中」表示可抢单的在线状态；有进行中订单时不能休息或停用。
        </div>
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
  import { ElMessage, ElMessageBox } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreateRider, fetchSearchRiderUsers, fetchUpdateRider } from '@/api/store/rider'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'RiderDialog' })

  type RiderItem = Api.Store.RiderItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    riderData?: RiderItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    riderData: null
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

  const statusOptions = [
    { label: '空闲', value: 1 },
    { label: '接单中', value: 2 },
    { label: '配送中', value: 3 },
    { label: '休息', value: 4 },
    { label: '停用', value: 5 }
  ]

  interface RiderForm {
    storeId: string
    accountMode: 'auto' | 'existing'
    userId: string
    name: string
    mobile: string
    maxOrders: number
    status: number
    remark: string
  }

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const storeLoading = ref(false)
  const userLoading = ref(false)
  const storeOptions = ref<Api.Store.VisibleStore[]>([])
  const userOptions = ref<Api.Store.RiderUserOption[]>([])

  const form = reactive<RiderForm>({
    storeId: '',
    accountMode: 'auto',
    userId: '',
    name: '',
    mobile: '',
    maxOrders: 3,
    status: 1,
    remark: ''
  })

  const rules = computed<FormRules>(() => {
    const result: FormRules = {
      storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
      maxOrders: [{ required: true, message: '请填写接单上限', trigger: 'change' }],
      status: [{ required: true, message: '请选择骑手状态', trigger: 'change' }]
    }
    if (isEdit.value) return result

    if (form.accountMode === 'auto') {
      result.name = [{ required: true, message: '请填写骑手姓名', trigger: 'blur' }]
      result.mobile = [
        { required: true, message: '请填写手机号', trigger: 'blur' },
        { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
      ]
    } else {
      result.userId = [{ required: true, message: '请选择后台账号', trigger: 'change' }]
    }

    return result
  })

  // 骑手列展示昵称与登录账号，编辑时用于只读展示
  const displayRider = computed(() => {
    const data = props.riderData
    if (!data) return ''
    if (data.nickname && data.username) return `${data.nickname}（${data.username}）`
    return data.nickname || data.username || '-'
  })

  // 账号下拉展示昵称、登录账号与手机号
  const displayUser = (user: Api.Store.RiderUserOption): string => {
    const name = user.nickname ? `${user.nickname}（${user.username}）` : user.username
    return user.mobile ? `${name} · ${user.mobile}` : name
  }

  // 拉取当前用户可见门店；编辑或具体门店上下文下锁定门店
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      if (storeContext.currentStoreId) {
        list = list.filter((store) => store.id === storeContext.currentStoreId)
      } else if (
        isEdit.value &&
        props.riderData &&
        !list.some((store) => store.id === props.riderData?.storeId)
      ) {
        list.unshift({
          id: props.riderData.storeId,
          name: props.riderData.storeName || '当前门店',
          code: '',
          status: 1
        })
      }
      storeOptions.value = list

      if (!list.some((store) => store.id === form.storeId)) {
        form.storeId = list[0]?.id ?? ''
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      storeLoading.value = false
    }
  }

  // 远程搜索启用账号，供绑定已有账号使用
  const searchUsers = async (keyword: string): Promise<void> => {
    userLoading.value = true
    try {
      userOptions.value = await fetchSearchRiderUsers(keyword)
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      userLoading.value = false
    }
  }

  // 切换账号方式时清空另一种模式的输入，避免提交隐藏字段造成误解
  watch(
    () => form.accountMode,
    async (mode) => {
      if (mode === 'auto') {
        form.userId = ''
      } else {
        form.name = ''
        form.mobile = ''
        await searchUsers('')
      }
      nextTick(() => formRef.value?.clearValidate())
    }
  )

  // 每次打开弹窗重置表单并加载门店、账号选项
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.riderData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        accountMode: data ? 'existing' : 'auto',
        userId: data?.userId ?? '',
        name: data?.nickname ?? '',
        mobile: data?.mobile ?? '',
        maxOrders: data?.maxOrders ?? 3,
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []
      userOptions.value = []

      await loadStores()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交骑手档案；自动开通账号时弹出只展示一次的初始密码
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      if (isEdit.value && props.riderData) {
        await fetchUpdateRider(props.riderData.id, {
          maxOrders: form.maxOrders,
          status: form.status,
          remark: form.remark
        })
        ElMessage.success('保存成功')
      } else {
        const payload: Api.Store.RiderCreateParams = {
          storeId: form.storeId,
          userId: form.accountMode === 'existing' ? form.userId : '0',
          maxOrders: form.maxOrders,
          status: form.status,
          remark: form.remark
        }
        if (form.accountMode === 'auto') {
          payload.name = form.name
          payload.mobile = form.mobile
        }

        const result = await fetchCreateRider(payload)
        if (result.initialPassword) {
          await ElMessageBox.alert(
            `<div class="leading-7">
              <p>登录账号：<b>${result.username}</b></p>
              <p>初始密码：<b>${result.initialPassword}</b></p>
              <p class="text-xs text-g-600">初始密码只显示一次，请立即告知骑手并提醒登录后修改。</p>
            </div>`,
            '骑手账号已开通',
            {
              dangerouslyUseHTMLString: true,
              confirmButtonText: '我已保存',
              showClose: false,
              closeOnClickModal: false,
              closeOnPressEscape: false
            }
          )
        }
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
