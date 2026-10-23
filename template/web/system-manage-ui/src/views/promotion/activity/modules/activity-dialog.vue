<!-- 营销活动新增 / 编辑弹窗：按玩法配置规则、互斥关系与名额 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑营销活动' : '新增营销活动'"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="活动名称" prop="name">
        <ElInput v-model="form.name" placeholder="如：周末满 100 减 20" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="活动玩法" prop="type">
        <ElSelect v-model="form.type" class="w-full">
          <ElOption
            v-for="item in typeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="作用层级" prop="level">
        <ElSelect v-model="form.level" class="w-full">
          <ElOption
            v-for="item in levelOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <span class="ml-3 text-xs text-g-600">商品级优惠落在参与商品上，订单级作用于整单</span>
      </ElFormItem>

      <ElFormItem label="活动说明">
        <ElInput v-model="form.description" placeholder="展示给会员的说明" maxlength="255" />
      </ElFormItem>

      <!-- 满减：单门槛或阶梯档位 -->
      <template v-if="form.type === 1">
        <ElFormItem label="满减门槛">
          <ElInputNumber
            v-model="form.thresholdAmount"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分，未启用阶梯时生效</span>
        </ElFormItem>

        <ElFormItem label="减免金额">
          <ElInputNumber
            v-model="form.discountAmount"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分，未启用阶梯时生效</span>
        </ElFormItem>

        <ElFormItem label="阶梯档位">
          <div class="w-full">
            <div
              v-for="(tier, index) in form.tiers"
              :key="index"
              class="mb-2 flex items-center gap-2"
            >
              <ElInputNumber
                v-model="tier.thresholdAmount"
                :min="0"
                :max="99999999"
                controls-position="right"
                placeholder="门槛（分）"
              />
              <span class="text-xs text-g-600">减</span>
              <ElInputNumber
                v-model="tier.discountAmount"
                :min="0"
                :max="99999999"
                controls-position="right"
                placeholder="减免（分）"
              />
              <ElButton link type="danger" @click="removeTier(index)">删除</ElButton>
            </div>
            <ElButton link type="primary" @click="addTier">添加档位</ElButton>
            <span class="ml-3 text-xs text-g-600">填了档位后按档位判定，取满足门槛的最高一档</span>
          </div>
        </ElFormItem>
      </template>

      <!-- 第二份半价 -->
      <ElFormItem v-if="form.type === 2" label="第二份折扣">
        <ElInputNumber
          v-model="form.secondHalfRate"
          :min="1"
          :max="9999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">万分比，5000 表示半价</span>
      </ElFormItem>

      <!-- 限时折扣 -->
      <template v-if="form.type === 3">
        <ElFormItem label="折扣率" prop="discountRate">
          <ElInputNumber
            v-model="form.discountRate"
            :min="1"
            :max="9999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">万分比，8000 表示 80 折</span>
        </ElFormItem>

        <ElFormItem label="折扣封顶">
          <ElInputNumber
            v-model="form.maxDiscountAmount"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分，0 表示不封顶</span>
        </ElFormItem>
      </template>

      <!-- 套餐特惠 -->
      <template v-if="form.type === 4">
        <ElFormItem label="套餐商品">
          <div class="w-full">
            <div
              v-for="(combo, index) in form.comboItems"
              :key="index"
              class="mb-2 flex items-center gap-2"
            >
              <ElInput v-model="combo.spuId" placeholder="菜品 SPU ID" />
              <ElInputNumber
                v-model="combo.quantity"
                :min="1"
                :max="99"
                controls-position="right"
              />
              <ElButton link type="danger" @click="removeCombo(index)">删除</ElButton>
            </div>
            <ElButton link type="primary" @click="addCombo">添加商品</ElButton>
          </div>
        </ElFormItem>

        <ElFormItem label="套餐价" prop="comboPrice">
          <ElInputNumber
            v-model="form.comboPrice"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分，低于单品之和才产生优惠</span>
        </ElFormItem>
      </template>

      <!-- 加价购 -->
      <template v-if="form.type === 5">
        <ElFormItem label="满额门槛">
          <ElInputNumber
            v-model="form.thresholdAmount"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分</span>
        </ElFormItem>

        <ElFormItem label="加价购商品" prop="addonSpuId">
          <ElInput v-model="form.addonSpuId" placeholder="菜品 SPU ID" />
        </ElFormItem>

        <ElFormItem label="加价购价" prop="addonPrice">
          <ElInputNumber
            v-model="form.addonPrice"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">单位：分</span>
        </ElFormItem>
      </template>

      <ElFormItem label="门店范围" prop="scope">
        <ElSelect v-model="form.scope" class="w-full">
          <ElOption
            v-for="item in scopeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem v-if="form.scope === 2" label="适用门店" prop="scopeIds">
        <ElInput v-model="scopeIdsText" placeholder="多个门店 ID 用英文逗号分隔" />
      </ElFormItem>

      <ElFormItem v-if="form.level === 1" label="参与商品" prop="productSpuIds">
        <ElInput v-model="productSpuIdsText" placeholder="多个菜品 SPU ID 用英文逗号分隔" />
      </ElFormItem>

      <ElFormItem label="优先级">
        <ElInputNumber v-model="form.priority" :min="0" :max="9999" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">数值大的先判定</span>
      </ElFormItem>

      <ElFormItem label="互斥组">
        <ElInput
          v-model="form.mutexGroup"
          placeholder="同组内只生效一个，留空表示不互斥"
          maxlength="64"
        />
      </ElFormItem>

      <ElFormItem label="独占">
        <ElSwitch v-model="form.exclusive" active-text="是" inactive-text="否" />
        <span class="ml-3 text-xs text-g-600">独占活动生效后其余活动终止判定</span>
      </ElFormItem>

      <ElFormItem label="可叠加">
        <ElSwitch v-model="form.stackable" active-text="是" inactive-text="否" />
        <span class="ml-3 text-xs text-g-600">可叠加活动的优惠金额累加</span>
      </ElFormItem>

      <ElFormItem label="名额上限">
        <ElInputNumber
          v-model="form.quotaLimit"
          :min="0"
          :max="9999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">0 表示不限</span>
      </ElFormItem>

      <ElFormItem label="名额周期">
        <ElRadioGroup v-model="form.quotaCycle">
          <ElRadio :value="1">整个活动期</ElRadio>
          <ElRadio :value="2">每日重置</ElRadio>
        </ElRadioGroup>
      </ElFormItem>

      <ElFormItem label="开始时间" prop="startTime">
        <ElDatePicker
          v-model="form.startTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          placeholder="选择开始时间"
          class="w-full"
        />
      </ElFormItem>

      <ElFormItem label="结束时间" prop="endTime">
        <ElDatePicker
          v-model="form.endTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          placeholder="选择结束时间"
          class="w-full"
        />
      </ElFormItem>

      <ElFormItem label="启用状态">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          active-text="启用"
          inactive-text="未启用"
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
    fetchCreateActivity,
    fetchGetActivityDetail,
    fetchUpdateActivity
  } from '@/api/promotion/promotion'

  defineOptions({ name: 'ActivityDialog' })

  type ActivityItem = Api.Promotion.ActivityItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    activityId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    activityId: ''
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success', type: DialogType): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const typeOptions = [
    { label: '满减', value: 1 },
    { label: '第二份半价', value: 2 },
    { label: '限时折扣', value: 3 },
    { label: '套餐特惠', value: 4 },
    { label: '加价购', value: 5 }
  ]

  const levelOptions = [
    { label: '商品级', value: 1 },
    { label: '订单级', value: 2 }
  ]

  const scopeOptions = [
    { label: '全品牌', value: 1 },
    { label: '指定门店', value: 2 }
  ]

  const form = reactive({
    name: '',
    type: 1 as Api.Promotion.ActivityType,
    level: 2 as Api.Promotion.ActivityLevel,
    description: '',
    thresholdAmount: 0,
    discountAmount: 0,
    discountRate: 8000,
    maxDiscountAmount: 0,
    secondHalfRate: 5000,
    comboPrice: 0,
    addonSpuId: '',
    addonPrice: 0,
    scope: 1 as Api.Promotion.ActivityScope,
    priority: 0,
    mutexGroup: '',
    exclusive: false,
    stackable: true,
    quotaLimit: 0,
    quotaCycle: 1 as Api.Promotion.ActivityQuotaCycle,
    startTime: '',
    endTime: '',
    status: 0,
    tiers: [] as Api.Promotion.ActivityTier[],
    comboItems: [] as Api.Promotion.ActivityComboItem[]
  })

  const scopeIdsText = ref('')
  const productSpuIdsText = ref('')

  const rules = reactive<FormRules>({
    name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
    type: [{ required: true, message: '请选择活动玩法', trigger: 'change' }],
    level: [{ required: true, message: '请选择作用层级', trigger: 'change' }],
    discountRate: [{ required: true, message: '请输入折扣率', trigger: 'blur' }],
    comboPrice: [{ required: true, message: '请输入套餐价', trigger: 'blur' }],
    addonSpuId: [{ required: true, message: '请输入加价购商品 SPU ID', trigger: 'blur' }],
    addonPrice: [{ required: true, message: '请输入加价购价', trigger: 'blur' }],
    scope: [{ required: true, message: '请选择门店范围', trigger: 'change' }],
    scopeIds: [{ required: true, message: '请输入适用门店', trigger: 'blur' }],
    productSpuIds: [{ required: true, message: '请输入参与商品', trigger: 'blur' }],
    startTime: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
    endTime: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
  })

  // 添加阶梯档位
  const addTier = (): void => {
    form.tiers.push({ thresholdAmount: 0, discountAmount: 0 })
  }

  // 删除阶梯档位
  const removeTier = (index: number): void => {
    form.tiers.splice(index, 1)
  }

  // 添加套餐商品
  const addCombo = (): void => {
    form.comboItems.push({ spuId: '', quantity: 1 })
  }

  // 删除套餐商品
  const removeCombo = (index: number): void => {
    form.comboItems.splice(index, 1)
  }

  // 打开弹窗时：编辑回填详情，新增用默认值
  watch(
    () => dialogVisible.value,
    async (open) => {
      if (!open) return
      formRef.value?.clearValidate()
      scopeIdsText.value = ''
      productSpuIdsText.value = ''
      form.tiers = []
      form.comboItems = []

      if (!isEdit.value || !props.activityId) {
        form.name = ''
        form.type = 1
        form.level = 2
        form.description = ''
        form.thresholdAmount = 0
        form.discountAmount = 0
        form.discountRate = 8000
        form.maxDiscountAmount = 0
        form.secondHalfRate = 5000
        form.comboPrice = 0
        form.addonSpuId = ''
        form.addonPrice = 0
        form.scope = 1
        form.priority = 0
        form.mutexGroup = ''
        form.exclusive = false
        form.stackable = true
        form.quotaLimit = 0
        form.quotaCycle = 1
        form.startTime = ''
        form.endTime = ''
        form.status = 0
        return
      }

      try {
        const detail = await fetchGetActivityDetail(props.activityId)
        applyDetail(detail)
      } catch {
        // 错误提示已由请求层统一弹出
      }
    }
  )

  // 把详情回填到表单；规则字段存在 JSON 里，需解析后按玩法取值
  const applyDetail = (detail: ActivityItem): void => {
    form.name = detail.name
    form.type = detail.type
    form.level = detail.level
    form.description = detail.description
    form.scope = detail.scope
    form.priority = detail.priority
    form.mutexGroup = detail.mutexGroup
    form.exclusive = detail.exclusive === 1
    form.stackable = detail.stackable === 1
    form.quotaLimit = detail.quotaLimit
    form.quotaCycle = detail.quotaCycle
    form.startTime = detail.startTime
    form.endTime = detail.endTime
    form.status = detail.status
    scopeIdsText.value = detail.scopeIds
    productSpuIdsText.value = (detail.productSpuIds || []).join(',')

    if (!detail.rule) return
    try {
      const rule = JSON.parse(detail.rule) as Api.Promotion.ActivitySaveParams
      form.thresholdAmount = rule.thresholdAmount || 0
      form.discountAmount = rule.discountAmount || 0
      form.discountRate = rule.discountRate || 8000
      form.maxDiscountAmount = rule.maxDiscountAmount || 0
      form.secondHalfRate = rule.secondHalfRate || 5000
      form.comboPrice = rule.comboPrice || 0
      form.addonSpuId = rule.addonSpuId || ''
      form.addonPrice = rule.addonPrice || 0
      form.tiers = rule.tiers || []
      form.comboItems = rule.comboItems || []
    } catch {
      // 规则 JSON 解析失败时保留默认值
    }
  }

  // 提交新增 / 编辑
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const scopeIds = splitIds(scopeIdsText.value)
    const productSpuIds = splitIds(productSpuIdsText.value)

    // 指定门店必须选门店；商品级活动必须选参与商品
    if (form.scope === 2 && scopeIds.length === 0) {
      ElMessage.warning('请填写适用门店')
      return
    }
    if (form.level === 1 && productSpuIds.length === 0) {
      ElMessage.warning('请填写参与商品')
      return
    }

    const payload: Api.Promotion.ActivitySaveParams = {
      id: isEdit.value ? props.activityId : undefined,
      name: form.name,
      type: form.type,
      level: form.level,
      description: form.description,
      thresholdAmount: form.thresholdAmount,
      discountAmount: form.discountAmount,
      tiers: form.tiers,
      discountRate: form.discountRate,
      maxDiscountAmount: form.maxDiscountAmount,
      secondHalfRate: form.secondHalfRate,
      comboItems: form.comboItems,
      comboPrice: form.comboPrice,
      addonSpuId: form.addonSpuId,
      addonPrice: form.addonPrice,
      scope: form.scope,
      scopeIds,
      priority: form.priority,
      mutexGroup: form.mutexGroup,
      exclusive: form.exclusive,
      stackable: form.stackable,
      quotaLimit: form.quotaLimit,
      quotaCycle: form.quotaCycle,
      startTime: form.startTime,
      endTime: form.endTime,
      status: form.status,
      productSpuIds
    }

    submitting.value = true
    try {
      if (isEdit.value) {
        await fetchUpdateActivity(props.activityId, payload)
      } else {
        await fetchCreateActivity(payload)
      }
      ElMessage.success(isEdit.value ? '已保存' : '已创建')
      dialogVisible.value = false
      emit('success', props.type)
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  // 把逗号分隔的 ID 文本拆成数组
  const splitIds = (text: string): string[] =>
    text
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item !== '')
</script>
