<!-- 订单看板：按状态筛选，支持接单 / 出餐 / 完成 / 取消与三种收款方式 -->
<template>
  <div class="page">
    <div class="stats">
      <div class="stat">
        <div class="stat-label">今日营收</div>
        <div class="stat-value">¥{{ yuan(overview?.revenue ?? 0) }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">已支付单</div>
        <div class="stat-value">{{ overview?.paidOrders ?? 0 }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">客单价</div>
        <div class="stat-value">¥{{ yuan(overview?.avgPrice ?? 0) }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">退款</div>
        <div class="stat-value">¥{{ yuan(overview?.refundAmount ?? 0) }}</div>
      </div>
      <div class="stat clickable" @click="filterByStatus(200)">
        <div class="stat-label">待接单</div>
        <div class="stat-value warn">{{ overview?.pendingAccept ?? 0 }}</div>
      </div>
      <div class="stat clickable" @click="filterByStatus(300)">
        <div class="stat-label">制作中</div>
        <div class="stat-value">{{ overview?.making ?? 0 }}</div>
      </div>
      <div class="stat clickable" @click="filterByStatus(400)">
        <div class="stat-label">待取餐</div>
        <div class="stat-value">{{ overview?.waitingPickup ?? 0 }}</div>
      </div>
    </div>

    <div class="toolbar">
      <ElRadioGroup v-model="status" @change="reload">
        <ElRadioButton :value="0">全部</ElRadioButton>
        <ElRadioButton :value="100">待支付</ElRadioButton>
        <ElRadioButton :value="200">待接单</ElRadioButton>
        <ElRadioButton :value="300">制作中</ElRadioButton>
        <ElRadioButton :value="400">待取餐</ElRadioButton>
      </ElRadioGroup>

      <ElButton :loading="loading" @click="reload">刷新</ElButton>
      <ElSwitch v-model="autoRefresh" active-text="每 15 秒自动刷新" />
      <span class="total">共 {{ total }} 单</span>
    </div>

    <ElEmpty v-if="!loading && list.length === 0" description="当前没有订单" />

    <div class="card-grid">
      <div
        v-for="item in list"
        :key="item.id"
        class="order-card"
        :class="{ 'order-card--urgent': item.status === 200 }"
      >
        <div class="flex-between">
          <span class="order-no">{{ item.no }}</span>
          <ElTag :type="statusTagType(item.status)" effect="plain">{{ item.statusText }}</ElTag>
        </div>

        <div class="meta">
          <span>{{ item.typeText }} · {{ item.channelText }}</span>
          <span v-if="item.tableName">桌台 {{ item.tableName }}</span>
          <span v-if="item.pickupCode">取餐码 {{ item.pickupCode }}</span>
        </div>

        <div class="flex-between amount-row">
          <span class="money">¥{{ yuan(item.payAmount) }}</span>
          <span class="pay-status">{{ item.payStatusText }}</span>
        </div>

        <div v-if="item.remark" class="remark">备注：{{ item.remark }}</div>
        <div class="time">{{ item.createdAt }}</div>

        <div class="actions">
          <ElButton
            v-if="item.status === 200"
            type="primary"
            size="small"
            @click="transit(item, 300, '接单')"
          >
            接单
          </ElButton>
          <ElButton
            v-if="item.status === 300"
            type="primary"
            size="small"
            @click="transit(item, 400, '出餐')"
          >
            出餐
          </ElButton>
          <ElButton
            v-if="item.status === 400"
            type="success"
            size="small"
            @click="transit(item, 500, '完成')"
          >
            完成
          </ElButton>
          <ElButton
            v-if="item.status === 100"
            type="danger"
            plain
            size="small"
            @click="cancel(item)"
          >
            取消
          </ElButton>
          <ElButton
            v-if="item.status === 100 && item.payStatus === 0 && auth.hasAuth('order:pos:settle')"
            type="warning"
            size="small"
            @click="openPay(item)"
          >
            收款
          </ElButton>
        </div>
      </div>
    </div>

    <ElPagination
      v-model:current-page="pageNo"
      class="pager"
      layout="prev, pager, next"
      :page-size="pageSize"
      :total="total"
      @current-change="loadData"
    />

    <ElDialog v-model="payVisible" title="收款" width="420px" :close-on-click-modal="false">
      <div v-if="payOrder" class="pay-head">
        <div class="order-no">{{ payOrder.no }}</div>
        <div class="money">¥{{ yuan(payOrder.payAmount) }}</div>
      </div>

      <ElRadioGroup v-model="payType" :disabled="paying" class="pay-types">
        <ElRadioButton :value="1">现金</ElRadioButton>
        <ElRadioButton :value="2">微信扫码</ElRadioButton>
        <ElRadioButton :value="3">支付宝条码</ElRadioButton>
      </ElRadioGroup>

      <ElInput
        v-if="payType === 3"
        v-model="authCode"
        placeholder="请扫描顾客的支付宝付款码"
        class="mt-3"
        :disabled="paying"
      />

      <ElAlert
        v-if="qrText"
        class="mt-3"
        type="info"
        :closable="false"
        show-icon
        title="请让顾客扫码支付，成功后自动刷新"
      >
        <div class="qr-text">{{ qrText }}</div>
      </ElAlert>

      <template #footer>
        <ElButton :disabled="paying" @click="closePay">关闭</ElButton>
        <ElButton type="primary" :loading="paying" @click="submitPay">
          {{ payType === 1 ? '确认收款' : '发起收款' }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

  import {
    fetchCancelOrder,
    fetchOrderPage,
    fetchPosOverview,
    fetchPosPay,
    fetchPosPayStatus,
    fetchSettleCash,
    fetchTransitOrder
  } from '@/api/order'
  import { useAuthStore } from '@/stores/auth'

  const auth = useAuthStore()

  const list = ref<Api.Order.Item[]>([])
  const loading = ref(false)
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = 20
  const status = ref(0)
  const autoRefresh = ref(true)
  // 今日经营概览：只取当前门店当天数据
  const overview = ref<Api.Order.PosOverview | null>(null)

  // 收款弹窗状态
  const payVisible = ref(false)
  const payOrder = ref<Api.Order.Item | null>(null)
  const payType = ref(1)
  const authCode = ref('')
  const paying = ref(false)
  const qrText = ref('')
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let refreshTimer: ReturnType<typeof setInterval> | null = null

  const yuan = (cent: number): string => ((cent || 0) / 100).toFixed(2)

  const statusTagType = (value: number): 'warning' | 'success' | 'info' | 'danger' => {
    if (value === 200) return 'warning'
    if (value === 300) return 'success'
    if (value === 400) return 'info'
    if (value === 600) return 'danger'

    return 'info'
  }

  // 拉取订单列表：门店由请求头决定
  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      const params: Api.Order.PageParams = { pageNo: pageNo.value, pageSize }
      if (status.value > 0) {
        params.status = status.value
      }
      const res = await fetchOrderPage(params)
      list.value = res.list ?? []
      total.value = res.total ?? 0
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      loading.value = false
    }
    // 概览与列表同批刷新，避免出现「列表有新单、概览还是旧数」
    void loadOverview()
  }

  // 拉取今日概览：失败不打断列表，卡片保留上一次数据
  const loadOverview = async (): Promise<void> => {
    try {
      overview.value = await fetchPosOverview()
    } catch {
      // 概览是辅助信息，失败静默
    }
  }

  // 点概览卡片按状态筛选，省去再点一次状态页签
  const filterByStatus = (value: number): void => {
    status.value = value
    pageNo.value = 1
    void loadData()
  }

  const reload = (): void => {
    pageNo.value = 1
    void loadData()
  }

  // 状态流转：接单 / 出餐 / 完成
  const transit = async (item: Api.Order.Item, to: number, label: string): Promise<void> => {
    try {
      await fetchTransitOrder(item.id, to)
      ElMessage.success(`${label}成功`)
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    }
  }

  const cancel = async (item: Api.Order.Item): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定取消订单 ${item.no} 吗？`, '取消订单', {
        confirmButtonText: '确认取消',
        cancelButtonText: '返回',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchCancelOrder(item.id, '门店取消')
      ElMessage.success('已取消')
      await loadData()
    } catch {
      // 错误提示由请求层统一弹出
    }
  }

  const openPay = (item: Api.Order.Item): void => {
    payOrder.value = item
    payType.value = 1
    authCode.value = ''
    qrText.value = ''
    payVisible.value = true
  }

  const stopPoll = (): void => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  const closePay = (): void => {
    stopPoll()
    payVisible.value = false
    paying.value = false
  }

  // 微信扫码：展示二维码后轮询支付结果
  const startPoll = (orderID: string): void => {
    stopPoll()
    pollTimer = setInterval(async () => {
      try {
        const res = await fetchPosPayStatus(orderID)
        if (res.paid) {
          stopPoll()
          ElMessage.success('收款成功')
          closePay()
          await loadData()
        }
      } catch {
        stopPoll()
      }
    }, 2000)
  }

  const submitPay = async (): Promise<void> => {
    const order = payOrder.value
    if (!order) return

    paying.value = true
    try {
      // 现金：直接置已支付，不生成支付单
      if (payType.value === 1) {
        await fetchSettleCash(order.id)
        ElMessage.success('现金收款成功')
        closePay()
        await loadData()

        return
      }

      // 微信 1 / 支付宝 2，与后端 PosPayType 对齐
      const type = payType.value === 2 ? 1 : 2
      const res = await fetchPosPay(order.id, type, authCode.value)

      if (res.paid) {
        ElMessage.success('收款成功')
        closePay()
        await loadData()

        return
      }

      if (type === 1) {
        const params = res.payParams || {}
        qrText.value = String(params.linkQrCode || params.codeUrl || params.qrCode || '')
        if (!qrText.value) {
          qrText.value = '渠道未返回二维码，请检查支付渠道配置'
        }
        startPoll(order.id)
      } else {
        ElMessage.warning('顾客尚未完成付款，请重试')
      }
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      paying.value = false
    }
  }

  // 自动刷新：只在看板可见时跑，切走由浏览器节流
  watch(autoRefresh, (on) => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
    if (on) {
      refreshTimer = setInterval(() => void loadData(), 15000)
    }
  })

  onMounted(() => {
    void loadData()
    if (autoRefresh.value) {
      refreshTimer = setInterval(() => void loadData(), 15000)
    }
  })

  onBeforeUnmount(() => {
    stopPoll()
    if (refreshTimer) {
      clearInterval(refreshTimer)
    }
  })
</script>

<style scoped>
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
    gap: 10px;
    margin-bottom: 12px;
  }

  .stat {
    padding: 12px 14px;
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
  }

  .stat.clickable {
    cursor: pointer;
  }

  .stat.clickable:hover {
    border-color: #409eff;
  }

  .stat-label {
    font-size: 12px;
    color: #6b7280;
  }

  .stat-value {
    margin-top: 6px;
    font-size: 20px;
    font-weight: 600;
  }

  .stat-value.warn {
    color: #e6a23c;
  }

  .total {
    margin-left: auto;
    font-size: 13px;
    color: #6b7280;
  }

  .flex-between {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .meta {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 8px;
    font-size: 13px;
    color: #6b7280;
  }

  .amount-row {
    margin-top: 10px;
  }

  .pay-status {
    font-size: 12px;
    color: #6b7280;
  }

  .remark {
    margin-top: 8px;
    font-size: 13px;
    color: #b45309;
  }

  .time {
    margin-top: 6px;
    font-size: 12px;
    color: #9ca3af;
  }

  .actions {
    display: flex;
    gap: 6px;
    margin-top: 12px;
  }

  .pager {
    justify-content: center;
    margin-top: 16px;
  }

  .pay-head {
    margin-bottom: 12px;
    text-align: center;
  }

  .pay-types {
    display: flex;
    justify-content: center;
    width: 100%;
  }

  .qr-text {
    margin-top: 6px;
    font-size: 12px;
    word-break: break-all;
  }

  .mt-3 {
    margin-top: 12px;
  }
</style>
