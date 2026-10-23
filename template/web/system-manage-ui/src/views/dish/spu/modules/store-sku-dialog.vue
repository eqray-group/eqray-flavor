<!-- 门店菜品配置弹窗：维护单个规格在当前门店的可售开关、门店价与生效区间 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="门店菜品配置"
    width="640px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="mb-4 rounded-lg bg-[var(--el-fill-color-lighter)] px-4 py-3 text-sm">
      <div class="font-medium">{{ skuData?.spuName }} · {{ skuData?.skuName }}</div>
      <div class="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-g-600">
        <span>品牌标准价：{{ formatYuan(skuData?.standardPrice) }} 元</span>
        <span>品牌会员价：{{ formatYuan(skuData?.memberPrice) }} 元</span>
        <span v-if="skuData?.skuStatus !== 1" class="text-[var(--el-color-warning)]">
          品牌侧已下架，门店配置不生效
        </span>
      </div>
    </div>

    <ElForm ref="formRef" :model="form" label-width="104px">
      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="门店可售">
          <ElSwitch v-model="form.available" :active-value="1" :inactive-value="0" />
        </ElFormItem>

        <ElFormItem label="沽清">
          <ElSwitch v-model="form.soldOut" :active-value="1" :inactive-value="0" />
        </ElFormItem>

        <ElFormItem label="仅线上可售">
          <ElSwitch v-model="form.onlineOnly" :active-value="1" :inactive-value="0" />
        </ElFormItem>

        <ElFormItem label="门店内排序">
          <ElInputNumber v-model="form.sort" :min="0" :max="9999" style="width: 100%" />
        </ElFormItem>

        <ElFormItem label="门店价（元）">
          <ElInputNumber
            v-model="form.storePrice"
            :min="0"
            :precision="2"
            :step="1"
            :controls="false"
            class="w-full"
            placeholder="留空沿用品牌标准价"
          />
        </ElFormItem>

        <ElFormItem label="门店会员价（元）">
          <ElInputNumber
            v-model="form.storeMemberPrice"
            :min="0"
            :precision="2"
            :step="1"
            :controls="false"
            class="w-full"
            placeholder="留空沿用品牌会员价"
          />
        </ElFormItem>

        <ElFormItem label="生效开始">
          <ElDatePicker
            v-model="form.startAt"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="留空表示立即生效"
            class="w-full"
          />
        </ElFormItem>

        <ElFormItem label="生效结束">
          <ElDatePicker
            v-model="form.endAt"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="留空表示长期有效"
            class="w-full"
          />
        </ElFormItem>
      </div>

      <ElFormItem label="备注">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="选填，仅后台可见"
        />
      </ElFormItem>

      <div class="-mt-2 pl-26 text-xs text-g-600">
        门店价与门店会员价都留空且不设区间时，视为不覆盖品牌价；门店会员价不能高于门店价。
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
  import { fetchSaveStoreSku } from '@/api/dish/store-sku'

  defineOptions({ name: 'DishStoreSkuDialog' })

  type StoreSkuItem = Api.Dish.StoreSkuItem

  interface Props {
    visible: boolean
    skuData?: StoreSkuItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    skuData: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const formRef = ref()
  const submitting = ref(false)

  const form = reactive({
    available: 0,
    soldOut: 0,
    onlineOnly: 0,
    sort: 0,
    remark: '',
    storePrice: null as number | null,
    storeMemberPrice: null as number | null,
    startAt: '',
    endAt: ''
  })

  // 分转元并保留两位小数，用于顶部只读展示
  const formatYuan = (cent?: number): string => {
    if (!cent) return '0.00'
    return (cent / 100).toFixed(2)
  }

  // 分转元，空值返回 null 以便输入框显示占位文案
  const centToYuan = (cent: number): number | null =>
    cent > 0 ? Number((cent / 100).toFixed(2)) : null

  // 元转分，空值按 0 处理表示不覆盖
  const yuanToCent = (yuan: number | null): number => {
    if (yuan === null || Number.isNaN(yuan)) return 0
    return Math.round(yuan * 100)
  }

  // 打开弹窗时按当前行回填门店配置
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      const data = props.skuData
      Object.assign(form, {
        available: data?.available ?? 0,
        soldOut: data?.soldOut ?? 0,
        onlineOnly: data?.onlineOnly ?? 0,
        sort: data?.sort ?? 0,
        remark: data?.remark ?? '',
        storePrice: centToYuan(data?.storePrice ?? 0),
        storeMemberPrice: centToYuan(data?.storeMemberPrice ?? 0),
        startAt: data?.startAt || '',
        endAt: data?.endAt || ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 保存门店配置：门店由请求头决定，这里只提交开关、价格与区间
  const handleSubmit = async (): Promise<void> => {
    if (!props.skuData) return

    const storePrice = yuanToCent(form.storePrice)
    const storeMemberPrice = yuanToCent(form.storeMemberPrice)

    if (storePrice > 0 && storeMemberPrice > storePrice) {
      ElMessage.error('门店会员价不能高于门店价')
      return
    }

    if (form.startAt && form.endAt && form.startAt >= form.endAt) {
      ElMessage.error('生效开始时间必须早于结束时间')
      return
    }

    submitting.value = true
    try {
      await fetchSaveStoreSku(props.skuData.skuId, {
        available: form.available,
        soldOut: form.soldOut,
        onlineOnly: form.onlineOnly,
        sort: form.sort,
        remark: form.remark,
        storePrice,
        storeMemberPrice,
        startAt: form.startAt,
        endAt: form.endAt
      })
      ElMessage.success('保存成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
