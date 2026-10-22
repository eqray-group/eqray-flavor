<!-- 桌台新增 / 编辑弹窗：选择门店与区域，维护桌号、桌码、座位数与排序 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑桌台' : '新增桌台'"
    width="600px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="92px">
      <ElRow :gutter="16">
        <ElCol :span="12">
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
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="所属区域" prop="areaId">
            <ElSelect
              v-model="form.areaId"
              class="w-full"
              filterable
              :loading="areaLoading"
              placeholder="请选择区域"
              no-data-text="该门店暂无区域，请先创建区域"
            >
              <ElOption
                v-for="area in availableAreas"
                :key="area.id"
                :label="area.name"
                :value="area.id"
              />
            </ElSelect>
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="桌号" prop="name">
            <ElInput v-model="form.name" placeholder="如 A01、包间1" maxlength="32" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="桌码" prop="code">
            <ElInput
              v-model="form.code"
              :disabled="isEdit"
              placeholder="留空自动生成"
              maxlength="64"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>
      <div class="-mt-3 mb-3 pl-26 text-xs text-g-600">
        桌码留空时按「门店编码-桌号」生成；创建后不可修改，保证二维码长期有效。
      </div>

      <!-- 座位数与排序各占半行：三列布局时输入框被 label 挤压，数字几乎不可见 -->
      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="座位数" prop="seats">
            <ElInputNumber
              v-model="form.seats"
              :min="1"
              :max="99"
              controls-position="right"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="排序" prop="sort">
            <ElInputNumber
              v-model="form.sort"
              :min="0"
              :max="9999"
              controls-position="right"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>

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
  import { fetchCreateTable, fetchGetTableAreaOptions, fetchUpdateTable } from '@/api/store/table'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'TableDialog' })

  type TableItem = Api.Store.TableItem
  type TableAreaOption = Api.Store.TableAreaOption
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    tableData?: TableItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    tableData: null
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

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const storeLoading = ref(false)
  const areaLoading = ref(false)
  const storeOptions = ref<Api.Store.VisibleStore[]>([])
  const areaOptions = ref<TableAreaOption[]>([])

  const form = reactive({
    storeId: '',
    areaId: '',
    name: '',
    code: '',
    seats: 4,
    sort: 0,
    remark: ''
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    areaId: [{ required: true, message: '请选择所属区域', trigger: 'change' }],
    name: [{ required: true, message: '请输入桌号', trigger: 'blur' }],
    seats: [{ required: true, message: '请输入座位数', trigger: 'change' }]
  }

  // 当前门店下可选区域，切换门店时自动收窄
  const availableAreas = computed(() =>
    areaOptions.value.filter((area) => area.storeId === form.storeId)
  )

  // 切换门店后，若原区域不属于新门店则回退到该门店第一个区域
  const syncAreaSelection = (): void => {
    if (!availableAreas.value.some((area) => area.id === form.areaId)) {
      form.areaId = availableAreas.value[0]?.id ?? ''
    }
  }

  // 拉取可见门店；选定具体门店或编辑时锁定门店下拉
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      if (storeContext.currentStoreId) {
        list = list.filter((store) => store.id === storeContext.currentStoreId)
      } else if (
        isEdit.value &&
        props.tableData &&
        !list.some((store) => store.id === props.tableData?.storeId)
      ) {
        list.unshift({
          id: props.tableData.storeId,
          name: props.tableData.storeName || '当前门店',
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

  // 拉取当前门店上下文或可见范围内的区域选项
  const loadAreas = async (): Promise<void> => {
    areaLoading.value = true
    try {
      areaOptions.value = await fetchGetTableAreaOptions()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      areaLoading.value = false
    }
  }

  // 门店或区域变化时，保证已选区域属于当前门店
  watch(
    () => form.storeId,
    () => {
      if (!props.visible) return
      syncAreaSelection()
    }
  )

  // 每次打开弹窗重置表单并并行加载门店与区域
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.tableData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        areaId: data?.areaId ?? '',
        name: data?.name ?? '',
        code: data?.code ?? '',
        seats: data?.seats ?? 4,
        sort: data?.sort ?? 0,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []
      areaOptions.value = []

      await Promise.all([loadStores(), loadAreas()])
      syncAreaSelection()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交桌台；成功后关闭弹窗并通知父页面刷新
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Store.TableSaveParams = {
      storeId: form.storeId,
      areaId: form.areaId,
      name: form.name,
      code: form.code,
      seats: form.seats,
      sort: form.sort,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.tableData) {
        await fetchUpdateTable(props.tableData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateTable(payload)
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
