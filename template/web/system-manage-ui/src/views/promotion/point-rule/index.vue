<!-- 积分抵扣规则：全品牌唯一一条规则，保存后即时生效于 C 端下单试算 -->
<template>
  <div v-loading="loading" class="point-rule-page art-full-height">
    <ElRow :gutter="20">
      <ElCol :xs="24" :lg="14">
        <ElCard class="art-table-card">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-base font-medium">积分抵扣规则</h3>
                <p class="mt-1 text-xs text-g-600">
                  会员下单时可使用积分抵扣订单金额，规则全品牌统一，保存后立即生效。
                </p>
              </div>
              <ElTag v-if="form.status === 1" type="success" effect="plain">已启用</ElTag>
              <ElTag v-else type="info" effect="plain">已停用</ElTag>
            </div>
          </template>

          <ElForm
            ref="formRef"
            :model="form"
            :rules="rules"
            label-width="150px"
            :disabled="!canUpdate"
          >
            <ElFormItem label="启用状态" prop="status">
              <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
              <span class="ml-2 text-xs text-g-600">停用后 C 端下单不再支持积分抵扣</span>
            </ElFormItem>

            <ElFormItem label="兑换比例" prop="exchangeRate">
              <ElInputNumber
                v-model="form.exchangeRate"
                :min="1"
                :max="100000"
                :step="1"
                step-strictly
                controls-position="right"
                class="w-45!"
              />
              <span class="ml-2 text-xs text-g-600">积分抵 1 元</span>
            </ElFormItem>

            <ElFormItem label="单笔抵扣上限" prop="maxDeductYuan">
              <ElInputNumber
                v-model="form.maxDeductYuan"
                :min="0"
                :max="100000"
                :precision="2"
                :step="1"
                controls-position="right"
                class="w-45!"
              />
              <span class="ml-2 text-xs text-g-600">元，0 表示不限制</span>
            </ElFormItem>

            <ElFormItem label="最低订单金额" prop="minOrderYuan">
              <ElInputNumber
                v-model="form.minOrderYuan"
                :min="0"
                :max="100000"
                :precision="2"
                :step="1"
                controls-position="right"
                class="w-45!"
              />
              <span class="ml-2 text-xs text-g-600">元，订单金额不足时不可使用积分</span>
            </ElFormItem>

            <ElFormItem>
              <ElButton type="primary" :loading="submitting" @click="handleSubmit"
                >保存规则</ElButton
              >
              <ElButton :disabled="submitting" @click="loadRule">重置</ElButton>
            </ElFormItem>
          </ElForm>
        </ElCard>
      </ElCol>

      <ElCol :xs="24" :lg="10">
        <ElCard class="art-table-card">
          <template #header>
            <h3 class="text-base font-medium">规则说明</h3>
          </template>

          <div class="space-y-3 text-sm text-g-700">
            <p>
              <span class="text-g-900">兑换比例：</span>填 100 表示 100 积分抵扣 1
              元；前端展示的抵扣额度按订单金额与上限取小值。
            </p>
            <p>
              <span class="text-g-900">单笔上限：</span
              >单笔订单最多用积分抵扣的金额，防止大额订单被一次性抵掉；0 表示不限制。
            </p>
            <p>
              <span class="text-g-900">最低订单金额：</span>订单金额未达到该值时不允许使用积分，0
              表示无门槛。
            </p>
            <p>
              <span class="text-g-900">生效方式：</span
              >保存后即时生效，无需重启服务；已生成的历史订单不受影响。
            </p>
          </div>

          <ElAlert
            class="mt-4"
            type="info"
            :closable="false"
            show-icon
            title="提示"
            description="积分由会员模块发放与流水记账，本页只控制「积分怎么抵钱」，不影响积分发放规则。"
          />
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchGetPointRule, fetchUpdatePointRule } from '@/api/promotion/promotion'
  import { useAuth } from '@/hooks/core/useAuth'

  defineOptions({ name: 'PromotionPointRule' })

  const { hasAuth } = useAuth()

  // 保存按钮单独控权，避免只读角色误改规则
  const canUpdate = computed(() => hasAuth('promotion:activity:update'))

  const formRef = ref<FormInstance>()
  const loading = ref(false)
  const submitting = ref(false)

  // 表单金额以元展示，提交时换算为分
  const form = reactive({
    exchangeRate: 100,
    maxDeductYuan: 50,
    minOrderYuan: 0,
    status: 1
  })

  const rules: FormRules = {
    exchangeRate: [
      { required: true, message: '请填写兑换比例', trigger: 'blur' },
      { type: 'integer', min: 1, message: '兑换比例必须是大于 0 的整数', trigger: 'blur' }
    ],
    maxDeductYuan: [
      {
        validator: (_rule, value: number, callback) => {
          if (value !== null && value !== undefined && value < 0) {
            callback(new Error('单笔抵扣上限不能为负数'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    minOrderYuan: [
      {
        validator: (_rule, value: number, callback) => {
          if (value !== null && value !== undefined && value < 0) {
            callback(new Error('最低订单金额不能为负数'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  // 元转分：四舍五入避免浮点误差
  const yuanToCent = (yuan: number): number => Math.round((yuan || 0) * 100)
  // 分转元：保留两位小数便于输入
  const centToYuan = (cent: number): number => Number(((cent || 0) / 100).toFixed(2))

  // 读取当前规则并回填表单
  const loadRule = async (): Promise<void> => {
    loading.value = true
    try {
      const rule = await fetchGetPointRule()
      form.exchangeRate = rule?.exchangeRate ?? 100
      form.maxDeductYuan = centToYuan(rule?.maxDeductAmount ?? 0)
      form.minOrderYuan = centToYuan(rule?.minOrderAmount ?? 0)
      form.status = rule?.status ?? 1
      nextTick(() => formRef.value?.clearValidate())
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  // 保存规则：元转分后提交
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await fetchUpdatePointRule({
        exchangeRate: form.exchangeRate,
        maxDeductAmount: yuanToCent(form.maxDeductYuan),
        minOrderAmount: yuanToCent(form.minOrderYuan),
        status: form.status
      })
      ElMessage.success('积分规则已保存')
      await loadRule()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  onMounted(loadRule)
</script>
