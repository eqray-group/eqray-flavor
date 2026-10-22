<!-- 桌台批量生成弹窗：按前缀与序号区间生成桌号，桌码由后端统一生成 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="批量生成桌台"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="92px">
      <ElFormItem label="所属门店" prop="storeId">
        <ElSelect
          v-model="form.storeId"
          class="w-full"
          filterable
          :disabled="!!storeContext.currentStoreId"
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

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="桌号前缀" prop="prefix">
            <ElInput v-model="form.prefix" placeholder="如 A" maxlength="16" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="座位数" prop="seats">
            <ElInputNumber v-model="form.seats" :min="1" :max="99" style="width: 100%" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="起始序号" prop="start">
            <ElInputNumber v-model="form.start" :min="1" :max="9999" style="width: 100%" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="生成数量" prop="count">
            <ElInputNumber v-model="form.count" :min="1" :max="100" style="width: 100%" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElAlert :closable="false" type="info" show-icon>
        <template #title>
          预计生成：<span class="font-medium">{{ previewText }}</span>
        </template>
        <div class="mt-1 text-xs text-g-600">
          桌号按前缀补零，单次最多 100
          张；桌码按「门店编码-桌号」生成，已存在的桌号或桌码会整批回滚。
        </div>
      </ElAlert>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">生成</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchBatchCreateTables, fetchGetTableAreaOptions } from '@/api/store/table'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'TableBatchDialog' })

  type TableAreaOption = Api.Store.TableAreaOption

  interface Props {
    visible: boolean
  }

  const props = defineProps<Props>()

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success', count: number): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

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
    prefix: 'A',
    start: 1,
    count: 10,
    seats: 4
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    areaId: [{ required: true, message: '请选择所属区域', trigger: 'change' }],
    prefix: [{ required: true, message: '请输入桌号前缀', trigger: 'blur' }],
    start: [{ required: true, message: '请输入起始序号', trigger: 'change' }],
    count: [{ required: true, message: '请输入生成数量', trigger: 'change' }],
    seats: [{ required: true, message: '请输入座位数', trigger: 'change' }]
  }

  // 当前门店下可选区域，切换门店时自动收窄
  const availableAreas = computed(() =>
    areaOptions.value.filter((area) => area.storeId === form.storeId)
  )

  // 与后端保持一致：按整批最大序号决定补零宽度，至少两位
  const previewText = computed(() => {
    const start = Math.max(1, Number(form.start) || 1)
    const count = Math.min(100, Math.max(1, Number(form.count) || 1))
    const end = start + count - 1
    const width = Math.max(2, String(end).length)
    const firstName = `${form.prefix}${String(start).padStart(width, '0')}`
    const lastName = `${form.prefix}${String(end).padStart(width, '0')}`
    return count === 1 ? firstName : `${firstName} ~ ${lastName}（共 ${count} 张）`
  })

  // 切换门店后，若原区域不属于新门店则回退到该门店第一个区域
  const syncAreaSelection = (): void => {
    if (!availableAreas.value.some((area) => area.id === form.areaId)) {
      form.areaId = availableAreas.value[0]?.id ?? ''
    }
  }

  // 拉取可见门店；选定具体门店时锁定门店下拉
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      if (storeContext.currentStoreId) {
        list = list.filter((store) => store.id === storeContext.currentStoreId)
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

  // 门店变化时同步区域选择
  watch(
    () => form.storeId,
    () => {
      if (!props.visible) return
      syncAreaSelection()
    }
  )

  // 每次打开弹窗重置为默认批量参数并加载门店 / 区域
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      Object.assign(form, {
        storeId: storeContext.currentStoreId ?? '',
        areaId: '',
        prefix: 'A',
        start: 1,
        count: 10,
        seats: 4
      })
      storeOptions.value = []
      areaOptions.value = []

      await Promise.all([loadStores(), loadAreas()])
      syncAreaSelection()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交批量生成；成功后关闭弹窗并按生成数量提示
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      const result = await fetchBatchCreateTables({
        storeId: form.storeId,
        areaId: form.areaId,
        prefix: form.prefix,
        start: form.start,
        count: form.count,
        seats: form.seats
      })
      ElMessage.success(`已生成 ${result.count} 张桌台`)
      dialogVisible.value = false
      emit('success', result.count)
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
