<script setup lang="ts">
/**
 * 我的订单：按时间倒序，状态用印章式标签区分。
 * 只读列表，详情与退款后续接。
 */
import { fetchOrders, type OrderRow } from '@/api/order'
import { useSessionStore } from '@/store/session'

definePage({
  name: 'orders',
  layout: 'tabbar',
  style: {
    navigationStyle: 'custom',
    enablePullDownRefresh: true,
  },
})

const session = useSessionStore()
const router = useRouter()

const loading = ref(true)
const failed = ref(false)
const list = ref<OrderRow[]>([])
const pageNo = ref(1)
const finished = ref(false)

const yuan = (cent: number): string => (cent / 100).toFixed(2)

/** 状态 → 印章颜色：待支付朱红、制作中墨黑、已完成淡化 */
function stampClass(status: number): string {
  if (status === 100) return 'stamp-wait'
  if (status === 300 || status === 200) return 'stamp-live'
  if (status === 600 || status === 800) return 'stamp-dead'
  return 'stamp-done'
}

function formatTime(raw: string): string {
  if (!raw) return ''
  const text = raw.includes('T') ? raw : raw.replace(' ', 'T')
  const date = new Date(text)
  if (Number.isNaN(date.getTime())) return raw
  const pad = (n: number): string => `${n}`.padStart(2, '0')
  return `${date.getMonth() + 1}月${date.getDate()}日 ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function load(reset = false): Promise<void> {
  if (reset) {
    pageNo.value = 1
    finished.value = false
  }
  loading.value = true
  failed.value = false
  try {
    const page = await fetchOrders({ pageNo: pageNo.value, pageSize: 10 })
    list.value = reset ? page.list || [] : [...list.value, ...(page.list || [])]
    finished.value = list.value.length >= (page.total || 0)
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

function loadMore(): void {
  if (finished.value || loading.value) return
  pageNo.value += 1
  void load()
}

onShow(() => {
  if (!session.isLogin) {
    router.push({ name: 'login' })
    return
  }
  void load(true)
})

onPullDownRefresh(() => {
  void load(true).finally(() => uni.stopPullDownRefresh())
})

onReachBottom(loadMore)
</script>

<template>
  <view class="paper">
    <view class="masthead">
      <text class="kicker">ORDER HISTORY</text>
      <text class="title">我的订单</text>
      <view class="rule" />
    </view>

    <view v-if="loading && !list.length" class="skeleton">
      <view v-for="i in 4" :key="i" class="sk-card" :style="{ animationDelay: `${i * 70}ms` }">
        <view class="sk-line w-30" />
        <view class="sk-line w-60" />
      </view>
    </view>

    <view v-else-if="failed" class="state">
      <text class="state-title">订单没取到</text>
      <text class="state-desc">检查网络后再试一次</text>
      <view class="state-btn" @click="load(true)">重新获取</view>
    </view>

    <view v-else-if="!list.length" class="state">
      <text class="state-title">还没有下过单</text>
      <text class="state-desc">点几样菜，第一单就写在这里</text>
      <view class="state-btn" @click="router.replaceAll({ name: 'menu' })">去点菜</view>
    </view>

    <scroll-view v-else class="list" scroll-y>
      <view v-for="order in list" :key="order.id" class="order">
        <view class="order-top">
          <text class="order-no">{{ order.no }}</text>
          <text class="stamp" :class="stampClass(order.status)">{{ order.statusText }}</text>
        </view>

        <view class="order-mid">
          <text class="order-type">{{ order.typeText }}</text>
          <text v-if="order.tableName" class="order-table">{{ order.tableName }}</text>
          <text class="order-time">{{ formatTime(order.createdAt) }}</text>
        </view>

        <view class="order-bottom">
          <text v-if="order.pickupCode" class="pickup">取餐码 {{ order.pickupCode }}</text>
          <text class="order-amount">
            <text class="sym">¥</text>{{ yuan(order.payAmount) }}
          </text>
        </view>
      </view>

      <view class="list-foot">
        <text>{{ finished ? '— 没有更多了 —' : '上拉加载更多' }}</text>
      </view>
    </scroll-view>
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

.masthead {
  padding: 26rpx 32rpx 0;
}
.kicker {
  font-size: 20rpx;
  color: $cinnabar;
  letter-spacing: 5rpx;
}
.title {
  display: block;
  margin-top: 8rpx;
  font-family: $serif;
  font-size: 56rpx;
  font-weight: 600;
  letter-spacing: 5rpx;
}
.rule {
  height: 6rpx;
  margin-top: 18rpx;
  background: $ink;
}

.list {
  flex: 1;
  min-height: 0;
  padding: 20rpx 32rpx 120rpx;
}

.order {
  padding: 26rpx 0;
  border-bottom: 1px dashed rgba(23, 19, 15, 0.18);
  animation: ink-in 0.4s both;
}
.order-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.order-no {
  font-size: 23rpx;
  color: $ink-soft;
  letter-spacing: 1rpx;
}
.stamp {
  padding: 3rpx 14rpx;
  font-size: 20rpx;
  letter-spacing: 2rpx;
  border: 2rpx solid currentcolor;
  border-radius: 4rpx;
  transform: rotate(-3deg);
}
.stamp-wait { color: $cinnabar; }
.stamp-live { color: $ink; }
.stamp-done { color: $ink-faint; }
.stamp-dead { color: $ink-faint; text-decoration: line-through; }

.order-mid {
  display: flex;
  gap: 14rpx;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: $ink-faint;
}
.order-type {
  color: $ink-soft;
}
.order-bottom {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-top: 14rpx;
}
.pickup {
  font-size: 22rpx;
  color: $cinnabar;
  letter-spacing: 1rpx;
}
.order-amount {
  font-family: $serif;
  font-size: 38rpx;
  font-weight: 600;
  margin-left: auto;
}
.sym {
  font-size: 22rpx;
  font-weight: 400;
}

.list-foot {
  padding: 36rpx 0;
  font-size: 20rpx;
  color: $ink-faint;
  letter-spacing: 2rpx;
  text-align: center;
}

.skeleton {
  padding: 26rpx 32rpx;
}
.sk-card {
  padding: 26rpx 0;
  border-bottom: 1px dashed rgba(23, 19, 15, 0.12);
  animation: ink-in 0.5s both;
}
.sk-line {
  height: 20rpx;
  margin-bottom: 14rpx;
  background: $paper-deep;
  border-radius: 4rpx;
}
.w-30 { width: 30%; }
.w-60 { width: 60%; }

.state {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 60rpx;
}
.state-title {
  font-family: $serif;
  font-size: 38rpx;
}
.state-desc {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: $ink-faint;
}
.state-btn {
  margin-top: 32rpx;
  padding: 16rpx 44rpx;
  font-size: 26rpx;
  color: $paper;
  background: $ink;
  border-radius: 999rpx;
}

@keyframes ink-in {
  from { opacity: 0; transform: translateY(10rpx); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
