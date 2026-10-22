<!-- 打印设备新增 / 编辑弹窗：维护设备号、接入方式、非敏感配置与加密密钥 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑打印设备' : '新增打印设备'"
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

      <ElFormItem label="设备名称" prop="name">
        <ElInput v-model="form.name" placeholder="如前台小票机、后厨打印机" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="设备号" prop="deviceNo">
        <ElInput
          v-model="form.deviceNo"
          placeholder="云打印机 SN / 局域网设备标识"
          maxlength="64"
        />
        <div class="mt-1 text-xs text-g-600">同一门店内设备号唯一，重复绑定会被拒绝。</div>
      </ElFormItem>

      <ElFormItem label="设备类型" prop="type">
        <ElSelect v-model="form.type" class="w-full">
          <ElOption
            v-for="item in printerTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="接入方式" prop="accessMode">
        <ElSelect v-model="form.accessMode" class="w-full">
          <ElOption
            v-for="item in accessModeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600">
          一期优先云打印机；局域网与本地客户端作为兜底方案。
        </div>
      </ElFormItem>

      <ElFormItem label="打印份数" prop="copies">
        <ElInputNumber v-model="form.copies" :min="1" :max="9" style="width: 100%" />
      </ElFormItem>

      <ElFormItem label="非敏感配置" prop="config">
        <ElInput
          v-model="form.config"
          type="textarea"
          :rows="3"
          maxlength="2000"
          show-word-limit
          placeholder='选填 JSON，如 {"vendor":"feie","endpoint":"https://..."}'
        />
        <div class="mt-1 text-xs text-g-600">
          仅填写厂商、地址、端口等非敏感信息；密钥请使用下方独立字段。
        </div>
      </ElFormItem>

      <ElFormItem v-if="isEdit && props.printerData?.hasSecret" label="当前密钥">
        <ElInput :model-value="props.printerData?.secretMask || '****'" disabled />
      </ElFormItem>

      <ElFormItem label="设备密钥" prop="secret">
        <ElInput
          v-model="form.secret"
          type="password"
          show-password
          autocomplete="new-password"
          :placeholder="isEdit ? '留空表示不修改原密钥' : '选填，保存后只展示脱敏值'"
          maxlength="512"
        />
        <div class="mt-1 text-xs text-g-600">
          密钥由服务端加密存储，列表与编辑接口都不会返回明文。
        </div>
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
  import { fetchCreatePrinter, fetchUpdatePrinter } from '@/api/store/printer'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'PrinterDialog' })

  type PrinterItem = Api.Store.PrinterItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    printerData?: PrinterItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    printerData: null
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

  const printerTypeOptions = [
    { label: '前台小票', value: 1 },
    { label: '后厨', value: 2 },
    { label: '标签', value: 3 },
    { label: '叫号', value: 4 }
  ]

  const accessModeOptions = [
    { label: '云打印机', value: 1 },
    { label: '局域网', value: 2 },
    { label: '本地客户端', value: 3 }
  ]

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const storeLoading = ref(false)
  const storeOptions = ref<Api.Store.VisibleStore[]>([])

  const form = reactive({
    storeId: '',
    name: '',
    deviceNo: '',
    type: 1,
    accessMode: 1,
    config: '',
    secret: '',
    copies: 1,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
    deviceNo: [{ required: true, message: '请输入设备号', trigger: 'blur' }],
    type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
    accessMode: [{ required: true, message: '请选择接入方式', trigger: 'change' }],
    copies: [{ required: true, message: '请填写打印份数', trigger: 'change' }],
    status: [{ required: true, message: '请选择设备状态', trigger: 'change' }]
  }

  // 拉取当前用户可见门店；具体门店上下文或编辑场景下锁定门店
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      if (storeContext.currentStoreId) {
        list = list.filter((store) => store.id === storeContext.currentStoreId)
      } else if (
        isEdit.value &&
        props.printerData &&
        !list.some((store) => store.id === props.printerData?.storeId)
      ) {
        list.unshift({
          id: props.printerData.storeId,
          name: props.printerData.storeName || '当前门店',
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

  // 每次打开弹窗重置表单并加载门店选项
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.printerData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        name: data?.name ?? '',
        deviceNo: data?.deviceNo ?? '',
        type: data?.type ?? 1,
        accessMode: data?.accessMode ?? 1,
        config: data?.config ?? '',
        secret: '',
        copies: data?.copies ?? 1,
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []

      await loadStores()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交打印设备；编辑时密钥留空表示沿用原值
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Store.PrinterSaveParams = {
      storeId: form.storeId,
      name: form.name,
      deviceNo: form.deviceNo,
      type: form.type,
      accessMode: form.accessMode,
      config: form.config,
      secret: form.secret || undefined,
      copies: form.copies,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.printerData) {
        await fetchUpdatePrinter(props.printerData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreatePrinter(payload)
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
