<!-- 优惠券模板新增 / 编辑弹窗：维护券规则、适用范围与发放限制 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑优惠券模板' : '新增优惠券模板'"
    width="640px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="券名称" prop="name">
        <ElInput v-model="form.name" placeholder="如：满 50 减 10 元券" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="券类型" prop="type">
        <ElSelect v-model="form.type" class="w-full">
          <ElOption
            v-for="item in typeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="券说明">
        <ElInput v-model="form.description" placeholder="展示给会员的说明" maxlength="255" />
      </ElFormItem>

      <ElFormItem label="使用门槛" prop="thresholdAmount">
        <ElInputNumber
          v-model="form.thresholdAmount"
          :min="0"
          :max="99999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">单位：分，0 表示无门槛</span>
      </ElFormItem>

      <!-- 折扣券按折扣率计算，其他券按固定面额 -->
      <ElFormItem v-if="form.type === 2" label="折扣率" prop="discountRate">
        <ElInputNumber v-model="form.discountRate" :min="1" :max="9999" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">万分比，7500 表示 75 折</span>
      </ElFormItem>

      <ElFormItem v-if="form.type === 2" label="折扣封顶">
        <ElInputNumber
          v-model="form.maxDiscountAmount"
          :min="0"
          :max="99999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">单位：分，0 表示不封顶</span>
      </ElFormItem>

      <ElFormItem v-else label="抵扣面额" prop="discountAmount">
        <ElInputNumber
          v-model="form.discountAmount"
          :min="0"
          :max="99999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">单位：分</span>
      </ElFormItem>

      <ElFormItem label="适用范围" prop="scope">
        <ElSelect v-model="form.scope" class="w-full">
          <ElOption
            v-for="item in scopeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem v-if="form.scope !== 1" label="适用对象" prop="scopeIds">
        <ElInput v-model="scopeIdsText" placeholder="多个 ID 用英文逗号分隔" />
        <span class="ml-3 text-xs text-g-600">
          {{ form.scope === 2 ? '填门店 ID' : '填菜品 SPU ID' }}
        </span>
      </ElFormItem>

      <ElFormItem label="渠道限制" prop="channelLimit">
        <ElSelect v-model="form.channelLimit" class="w-full">
          <ElOption
            v-for="item in channelOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="发放方式" prop="issueType">
        <ElSelect v-model="form.issueType" class="w-full">
          <ElOption
            v-for="item in issueTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="发行总量">
        <ElInputNumber
          v-model="form.totalCount"
          :min="0"
          :max="9999999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">0 表示不限量</span>
      </ElFormItem>

      <ElFormItem label="每人限领">
        <ElInputNumber
          v-model="form.perMemberLimit"
          :min="0"
          :max="999"
          controls-position="right"
        />
        <span class="ml-3 text-xs text-g-600">0 表示不限</span>
      </ElFormItem>

      <ElFormItem label="领取后有效" prop="validDays">
        <ElInputNumber v-model="form.validDays" :min="0" :max="3650" controls-position="right" />
        <span class="ml-3 text-xs text-g-600">天，0 表示长期有效至结束时间</span>
      </ElFormItem>

      <ElFormItem label="生效开始" prop="startTime">
        <ElDatePicker
          v-model="form.startTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          placeholder="选择生效开始时间"
          class="w-full"
        />
      </ElFormItem>

      <ElFormItem label="生效结束" prop="endTime">
        <ElDatePicker
          v-model="form.endTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          placeholder="选择生效结束时间"
          class="w-full"
        />
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          active-text="启用"
          inactive-text="停用"
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
    fetchCreateCouponTemplate,
    fetchGetCouponTemplateDetail,
    fetchUpdateCouponTemplate
  } from '@/api/promotion/promotion'

  defineOptions({ name: 'CouponTemplateDialog' })

  type TemplateItem = Api.Promotion.CouponTemplateItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    templateId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    templateId: ''
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
    { label: '满减券', value: 1 },
    { label: '折扣券', value: 2 },
    { label: '单品券', value: 3 },
    { label: '兑换券', value: 4 },
    { label: '代金券', value: 5 }
  ]

  const scopeOptions = [
    { label: '全品牌', value: 1 },
    { label: '指定门店', value: 2 },
    { label: '指定商品', value: 3 }
  ]

  const channelOptions = [
    { label: '不限渠道', value: 0 },
    { label: '仅线上', value: 1 },
    { label: '仅线下', value: 2 }
  ]

  const issueTypeOptions = [
    { label: '主动领取', value: 1 },
    { label: '消费返券', value: 2 },
    { label: '新人礼', value: 3 },
    { label: '定向发券', value: 4 }
  ]

  const form = reactive({
    name: '',
    type: 1 as Api.Promotion.CouponType,
    description: '',
    thresholdAmount: 0,
    discountAmount: 0,
    discountRate: 7500,
    maxDiscountAmount: 0,
    scope: 1 as Api.Promotion.CouponScope,
    channelLimit: 0 as Api.Promotion.CouponChannelLimit,
    issueType: 1 as Api.Promotion.CouponIssueType,
    totalCount: 0,
    perMemberLimit: 0,
    validDays: 0,
    startTime: '',
    endTime: '',
    status: 1
  })

  // 适用对象 ID 的文本形态，提交时按逗号拆成数组
  const scopeIdsText = ref('')

  const rules = reactive<FormRules>({
    name: [{ required: true, message: '请输入券名称', trigger: 'blur' }],
    type: [{ required: true, message: '请选择券类型', trigger: 'change' }],
    thresholdAmount: [{ required: true, message: '请输入使用门槛', trigger: 'blur' }],
    discountRate: [{ required: true, message: '请输入折扣率', trigger: 'blur' }],
    discountAmount: [{ required: true, message: '请输入抵扣面额', trigger: 'blur' }],
    scope: [{ required: true, message: '请选择适用范围', trigger: 'change' }],
    scopeIds: [{ required: true, message: '请输入适用对象', trigger: 'blur' }],
    issueType: [{ required: true, message: '请选择发放方式', trigger: 'change' }],
    startTime: [{ required: true, message: '请选择生效开始时间', trigger: 'change' }],
    endTime: [{ required: true, message: '请选择生效结束时间', trigger: 'change' }],
    validDays: [{ required: true, message: '请输入领取后有效天数', trigger: 'blur' }]
  })

  // 打开弹窗时：编辑回填详情，新增用默认值
  watch(
    () => dialogVisible.value,
    async (open) => {
      if (!open) return
      formRef.value?.clearValidate()
      scopeIdsText.value = ''

      if (!isEdit.value || !props.templateId) {
        form.name = ''
        form.type = 1
        form.description = ''
        form.thresholdAmount = 0
        form.discountAmount = 0
        form.discountRate = 7500
        form.maxDiscountAmount = 0
        form.scope = 1
        form.channelLimit = 0
        form.issueType = 1
        form.totalCount = 0
        form.perMemberLimit = 0
        form.validDays = 0
        form.startTime = ''
        form.endTime = ''
        form.status = 1
        return
      }

      try {
        const detail = await fetchGetCouponTemplateDetail(props.templateId)
        applyDetail(detail)
      } catch {
        // 错误提示已由请求层统一弹出
      }
    }
  )

  // 把详情回填到表单
  const applyDetail = (detail: TemplateItem): void => {
    form.name = detail.name
    form.type = detail.type
    form.description = detail.description
    form.thresholdAmount = detail.thresholdAmount
    form.discountAmount = detail.discountAmount
    form.discountRate = detail.discountRate
    form.maxDiscountAmount = detail.maxDiscountAmount
    form.scope = detail.scope
    form.channelLimit = detail.channelLimit
    form.issueType = detail.issueType
    form.totalCount = detail.totalCount
    form.perMemberLimit = detail.perMemberLimit
    form.validDays = detail.validDays
    form.startTime = detail.startTime
    form.endTime = detail.endTime
    form.status = detail.status
    scopeIdsText.value = detail.scopeIds
  }

  // 提交新增 / 编辑
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const scopeIds = scopeIdsText.value
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item !== '')

    // 指定范围必须选对象，否则券永远不可用
    if (form.scope !== 1 && scopeIds.length === 0) {
      ElMessage.warning('请填写适用对象')
      return
    }

    const payload: Api.Promotion.CouponTemplateSaveParams = {
      id: isEdit.value ? props.templateId : undefined,
      name: form.name,
      type: form.type,
      description: form.description,
      thresholdAmount: form.thresholdAmount,
      discountAmount: form.discountAmount,
      discountRate: form.discountRate,
      maxDiscountAmount: form.maxDiscountAmount,
      scope: form.scope,
      scopeIds,
      channelLimit: form.channelLimit,
      issueType: form.issueType,
      totalCount: form.totalCount,
      perMemberLimit: form.perMemberLimit,
      startTime: form.startTime,
      endTime: form.endTime,
      validDays: form.validDays,
      status: form.status
    }

    submitting.value = true
    try {
      if (isEdit.value) {
        await fetchUpdateCouponTemplate(props.templateId, payload)
      } else {
        await fetchCreateCouponTemplate(payload)
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
</script>
