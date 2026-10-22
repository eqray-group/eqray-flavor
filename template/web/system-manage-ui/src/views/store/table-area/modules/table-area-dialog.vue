<!-- 桌台区域新增 / 编辑弹窗：选择可见门店并维护区域名称与排序 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑桌台区域' : '新增桌台区域'"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="92px">
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

      <ElFormItem label="区域名称" prop="name">
        <ElInput v-model="form.name" placeholder="如大厅、包间、露台" maxlength="32" />
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" style="width: 100%" />
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

      <div class="-mt-2 pl-23 text-xs text-g-600">
        区域名称在同一门店内唯一；区域下仍有桌台时不能删除。
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
  import { fetchCreateTableArea, fetchUpdateTableArea } from '@/api/store/table'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'TableAreaDialog' })

  type TableAreaItem = Api.Store.TableAreaItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    areaData?: TableAreaItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    areaData: null
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
  const storeOptions = ref<Api.Store.VisibleStore[]>([])

  const form = reactive({
    storeId: '',
    name: '',
    sort: 0,
    remark: ''
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    name: [{ required: true, message: '请输入区域名称', trigger: 'blur' }]
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
        props.areaData &&
        !list.some((store) => store.id === props.areaData?.storeId)
      ) {
        list.unshift({
          id: props.areaData.storeId,
          name: props.areaData.storeName || '当前门店',
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

      const data = props.areaData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        name: data?.name ?? '',
        sort: data?.sort ?? 0,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []

      await loadStores()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交区域；成功后关闭弹窗并通知父页面刷新
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Store.TableAreaSaveParams = {
      storeId: form.storeId,
      name: form.name,
      sort: form.sort,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.areaData) {
        await fetchUpdateTableArea(props.areaData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateTableArea(payload)
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
