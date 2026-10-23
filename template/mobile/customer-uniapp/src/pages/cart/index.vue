<script setup lang="ts">
/**
 * 确认订单：像一张待结账的小票。
 * 金额只信服务端试算，本地 subtotal 仅用于与接口结果对照，不参与下单。
 */
import { createOrder, fetchEstimate, payOrder, type EstimateResult } from '@/api/order'
import { useCartStore } from '@/store/cart'
import { useSessionStore } from '@/store/session'

definePage({
  name: 'cart',
  style: {
    navigationBarTitleText: '确认订单',
    navigationBarBackgroundColor: '#F7F2E7',
    navigationBarTextStyle: 'black',
  },
})

const session = useSessionStore()
const cart = useCartStore()
const router = useRouter()

const remark = ref('')
const submitting = ref(false)
const estimate = ref<EstimateResult | null>(null)
const estimateLoading = ref(false)
/** 扫码进店时会带桌号，带了就是堂食 */
const tableId = ref('')
const requestNo = ref('')

const yuan = (cent: number): string => (cent / 100).toFixed(2)

const isDineIn = computed(() => !!tableId.value)

function newRequestNo(): string {
  return `C${Date.now()}${Math.floor(Math.random() * 1000)}`
}

async function recalc(): Promise<void> {
  if (!cart.lines.length) {
    estimate.value = null
    return
  }
  estimateLoading.value = true
  try {
    estimate.value = await fetchEstimate({
      items: cart.lines.map((line) => ({
        skuId: line.skuId,
        quantity: line.quantity,
        addons: line.addons.map((addon) => ({ addonItemId: addon.addonItemId, quantity: addon.quantity })),
        remark: line.remark,
      })),
    })
  } catch {
    estimate.value = null
  } finally {
    estimateLoading.value = false
  }
}

/** 微信支付参数：后端透传的字段名与 wx.requestPayment 对齐 */
interface WxPayParams {
  timeStamp: string
  nonceStr: string
  package: string
  signType: 'MD5' | 'HMAC-SHA256' | 'RSA'
  paySign: string
}

function requestPay(params: unknown): Promise<void> {
  return new Promise((resolve, reject) => {
    const pay = params as WxPayParams
    if (!pay?.paySign) {
      reject(new Error('no pay params'))
      return
    }
    // 类型声明里 App 支付必填 orderInfo，小程序端用不到，这里按实际字段透传
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: pay.timeStamp,
      nonceStr: pay.nonceStr,
      package: pay.package,
      signType: pay.signType,
      paySign: pay.paySign,
      success: () => resolve(),
      fail: (err: unknown) => reject(err instanceof Error ? err : new Error('payment canceled')),
    } as unknown as UniApp.RequestPaymentOptions)
  })
}

async function submit(): Promise<void> {
  if (!cart.lines.length || submitting.value) return

  submitting.value = true
  try {
    const created = await createOrder({
      type: isDineIn.value ? 1 : 3,
      tableId: tableId.value || undefined,
      remark: remark.value.trim(),
      requestNo: requestNo.value || (requestNo.value = newRequestNo()),
      items: cart.lines.map((line) => ({
        skuId: line.skuId,
        quantity: line.quantity,
        addons: line.addons.map((addon) => ({ addonItemId: addon.addonItemId, quantity: addon.quantity })),
        remark: line.remark,
      })),
    })

    cart.clear()
    // 下单成功先清购物车与幂等号，避免用户返回后重复提交
    requestNo.value = ''

    try {
      const params = await payOrder(created.order.id, { channelCode: 'wx_lite' })
      await requestPay(params)
      useGlobalToast().success({ msg: '支付成功' })
    } catch {
      // 拉起支付失败或用户取消：订单已生成，去订单列表继续付
      useGlobalToast().warning({ msg: '订单已提交，可在订单中继续支付' })
    }

    router.replaceAll({ name: 'orders' })
  } catch {
    // 业务错误已由请求层提示；保留购物车与请求号便于重试
  } finally {
    submitting.value = false
  }
}

onLoad((query) => {
  tableId.value = (query?.tableId as string) || ''
  if (!session.isLogin) {
    router.push({ name: 'login' })
    return
  }
  void recalc()
})
</script>

<template>
  <view class="paper">
    <view class="sheet-head">
      <text class="kicker">{{ isDineIn ? '堂食' : '自取' }}</text>
      <text class="title">确认订单</text>
      <view class="hairline" />
    </view>

    <view v-if="!cart.lines.length" class="empty">
      <text class="empty-title">还没有菜</text>
      <text class="empty-desc">回到菜单挑几样</text>
      <view class="empty-btn" @click="router.replaceAll({ name: 'menu' })">去点菜</view>
    </view>

    <scroll-view v-else class="receipt" scroll-y>
      <!-- 明细：名称与金额两端对齐，中间用点线牵引 -->
      <view v-for="line in cart.lines" :key="line.key" class="row">
        <view class="row-main">
          <view class="row-title">
            <text class="name">{{ line.spuName }}</text>
            <text v-if="line.skuName" class="sku">{{ line.skuName }}</text>
          </view>
          <text v-if="line.addonText" class="addon">+ {{ line.addonText }}</text>
          <text v-if="line.remark" class="remark">备注：{{ line.remark }}</text>
        </view>

        <view class="qty">
          <text class="qty-btn" @click="cart.dec(line.key); recalc()">−</text>
          <text class="qty-num">{{ line.quantity }}</text>
          <text class="qty-btn" @click="cart.inc(line.key); recalc()">＋</text>
        </view>

        <text class="amount">¥{{ yuan(line.unitPrice * line.quantity) }}</text>
      </view>

      <view class="note-block">
        <text class="note-label">整单备注</text>
        <input v-model="remark" class="note-input" placeholder="口味、餐具、到店时间…" placeholder-class="ph" />
      </view>

      <view class="divider">
        <text class="divider-text">合计</text>
      </view>

      <view class="sum">
        <view class="sum-row">
          <text class="sum-label">菜品小计</text>
          <text class="sum-value">¥{{ yuan(estimate?.itemAmount ?? cart.subtotal) }}</text>
        </view>
        <view v-if="(estimate?.activityAmount ?? 0) > 0" class="sum-row discount">
          <text class="sum-label">{{ estimate?.activityNames?.[0] || '活动优惠' }}</text>
          <text class="sum-value">−¥{{ yuan(estimate?.activityAmount ?? 0) }}</text>
        </view>
        <view v-if="(estimate?.couponAmount ?? 0) > 0" class="sum-row discount">
          <text class="sum-label">{{ estimate?.couponName || '优惠券' }}</text>
          <text class="sum-value">−¥{{ yuan(estimate?.couponAmount ?? 0) }}</text>
        </view>
        <view class="sum-final">
          <text class="sum-final-label">应付</text>
          <text class="sum-final-value">
            <text class="sym">¥</text>{{ yuan(estimate?.payAmount ?? 0) }}
          </text>
        </view>
      </view>
    </scroll-view>

    <view v-if="cart.lines.length" class="bar">
      <view class="bar-left">
        <text class="bar-label">{{ estimateLoading ? '正在核价…' : '应付' }}</text>
        <text class="bar-amount">
          <text class="sym">¥</text>{{ yuan(estimate?.payAmount ?? 0) }}
        </text>
      </view>
      <view class="bar-btn" :class="{ disabled: submitting }" @click="submit">
        {{ submitting ? '提交中…' : '提交并支付' }}
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
$paper: #f7f2e7;
$paper-deep: #efe7d6;
$ink: #17130f;
$ink-soft: #5c5448;
$ink-faint: #9a9183;
$cinnabar: #b23a2f;
$serif: 'Songti SC', 'Noto Serif CJK SC', 'Source Han Serif SC', 'SimSun', serif;

.paper {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: $paper;
  color: $ink;
}

.sheet-head {
  padding: 24rpx 32rpx 0;
}
.kicker {
  font-size: 22rpx;
  color: $cinnabar;
  letter-spacing: 4rpx;
}
.title {
  display: block;
  margin-top: 8rpx;
  font-family: $serif;
  font-size: 54rpx;
  font-weight: 600;
  letter-spacing: 5rpx;
}
.hairline {
  height: 1px;
  margin-top: 20rpx;
  background: rgba(23, 19, 15, 0.25);
}

.receipt {
  flex: 1;
  min-height: 0;
  padding: 8rpx 32rpx 40rpx;
}

.row {
  display: flex;
  gap: 16rpx;
  align-items: flex-start;
  padding: 26rpx 0;
  border-bottom: 1px dashed rgba(23, 19, 15, 0.16);
}
.row-main {
  flex: 1;
  min-width: 0;
}
.row-title {
  display: flex;
  align-items: baseline;
  gap: 10rpx;
}
.name {
  font-family: $serif;
  font-size: 30rpx;
  font-weight: 600;
}
.sku {
  font-size: 21rpx;
  color: $ink-faint;
}
.addon,
.remark {
  display: block;
  margin-top: 6rpx;
  font-size: 21rpx;
  color: $ink-faint;
}
.qty {
  display: flex;
  align-items: center;
  gap: 14rpx;
}
.qty-btn {
  width: 44rpx;
  height: 44rpx;
  font-size: 26rpx;
  line-height: 42rpx;
  text-align: center;
  border: 1px solid rgba(23, 19, 15, 0.35);
  border-radius: 50%;
}
.qty-num {
  min-width: 32rpx;
  font-size: 26rpx;
  text-align: center;
}
.amount {
  width: 120rpx;
  font-family: $serif;
  font-size: 30rpx;
  font-weight: 600;
  text-align: right;
}

.note-block {
  padding: 26rpx 0;
}
.note-label {
  display: block;
  margin-bottom: 12rpx;
  font-size: 22rpx;
  color: $ink-soft;
  letter-spacing: 2rpx;
}
.note-input {
  padding: 18rpx 20rpx;
  font-size: 25rpx;
  background: $paper-deep;
  border-radius: 8rpx;
}
.ph {
  color: $ink-faint;
}

.divider {
  position: relative;
  margin: 10rpx 0 6rpx;
  text-align: center;
}
.divider::before {
  position: absolute;
  top: 50%;
  right: 0;
  left: 0;
  height: 1px;
  content: '';
  background: rgba(23, 19, 15, 0.14);
}
.divider-text {
  position: relative;
  padding: 0 18rpx;
  font-size: 20rpx;
  color: $ink-faint;
  letter-spacing: 6rpx;
  background: $paper;
}

.sum {
  padding-top: 14rpx;
}
.sum-row {
  display: flex;
  justify-content: space-between;
  padding: 8rpx 0;
  font-size: 24rpx;
  color: $ink-soft;
}
.sum-row.discount {
  color: $cinnabar;
}
.sum-final {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-top: 18rpx;
  margin-top: 10rpx;
  border-top: 2rpx solid $ink;
}
.sum-final-label {
  font-size: 24rpx;
  letter-spacing: 2rpx;
}
.sum-final-value {
  font-family: $serif;
  font-size: 52rpx;
  font-weight: 600;
}
.sym {
  margin-right: 2rpx;
  font-size: 24rpx;
  font-weight: 400;
}

.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
  background: $ink;
}
.bar-left {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  color: $paper;
}
.bar-label {
  font-size: 22rpx;
  opacity: 0.65;
}
.bar-amount {
  font-family: $serif;
  font-size: 42rpx;
  font-weight: 600;
}
.bar-btn {
  padding: 18rpx 40rpx;
  font-size: 27rpx;
  letter-spacing: 2rpx;
  color: #fff;
  background: $cinnabar;
  border-radius: 999rpx;
}
.bar-btn.disabled {
  opacity: 0.6;
}

.empty {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.empty-title {
  font-family: $serif;
  font-size: 40rpx;
}
.empty-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: $ink-faint;
}
.empty-btn {
  margin-top: 32rpx;
  padding: 16rpx 44rpx;
  font-size: 26rpx;
  color: $paper;
  background: $ink;
  border-radius: 999rpx;
}
</style>
