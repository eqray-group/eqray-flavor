<script setup lang="ts">
/**
 * 点单页：左侧分类、右侧菜单、底部票据式购物车。
 * 视觉方向：纸墨菜单——暖纸底、墨黑字、朱红只用于强调，标题与价格用宋体。
 */
import { fetchMenu, type MenuCategory, type MenuSpu, type MenuSku } from '@/api/dish'
import { useCartStore, type CartAddon } from '@/store/cart'
import { useSessionStore } from '@/store/session'

definePage({
  name: 'menu',
  layout: 'tabbar',
  style: {
    navigationStyle: 'custom',
  },
})

const session = useSessionStore()
const cart = useCartStore()

const loading = ref(true)
const failed = ref(false)
const categories = ref<MenuCategory[]>([])
const activeCategory = ref('')
const scrollInto = ref('')

/** 加购弹窗 */
const sheetVisible = ref(false)
const picking = ref<MenuSpu | null>(null)
const pickedSku = ref('')
const pickedAddons = reactive<Record<string, string[]>>({})
const pickedRemark = ref('')
/** 购物车条的数字跳动 */
const bump = ref(false)

const yuan = (cent: number): string => (cent / 100).toFixed(2)

/** 分类下的菜品：子分类的菜品一并展开，避免漏菜 */
function dishesOf(category: MenuCategory): MenuSpu[] {
  const own = category.items || []
  const children = (category.children || []).flatMap((child) => child.items || [])
  return [...own, ...children]
}

const activeDishes = computed(() => {
  const category = categories.value.find((item) => item.id === activeCategory.value)
  return category ? dishesOf(category) : []
})

const lowestPrice = (spu: MenuSpu): number =>
  Math.min(...spu.skus.filter((sku) => sku.soldOut !== 1).map((sku) => sku.price), Number.POSITIVE_INFINITY)

const soldOutAll = (spu: MenuSpu): boolean => spu.skus.every((sku) => sku.soldOut === 1)

async function load(): Promise<void> {
  loading.value = true
  failed.value = false
  try {
    const data = await fetchMenu()
    categories.value = data.categories || []
    if (!activeCategory.value && categories.value.length) {
      activeCategory.value = categories.value[0].id
    }
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

function switchCategory(id: string): void {
  activeCategory.value = id
  scrollInto.value = `sec-${id}`
}

function openSheet(spu: MenuSpu): void {
  if (soldOutAll(spu)) return
  picking.value = spu
  const firstAvailable = spu.skus.find((sku) => sku.soldOut !== 1)
  pickedSku.value = firstAvailable?.id || ''
  pickedRemark.value = ''
  Object.keys(pickedAddons).forEach((key) => delete pickedAddons[key])
  for (const group of spu.addonGroups || []) {
    pickedAddons[group.id] = []
  }
  sheetVisible.value = true
}

function toggleAddon(groupId: string, itemId: string, max: number): void {
  const current = pickedAddons[groupId] || []
  if (current.includes(itemId)) {
    pickedAddons[groupId] = current.filter((id) => id !== itemId)
    return
  }
  if (max > 0 && current.length >= max) {
    useGlobalToast().warning({ msg: `最多选 ${max} 项` })
    return
  }
  pickedAddons[groupId] = [...current, itemId]
}

/** 必选组没选够时不允许加购，这是最容易漏的校验 */
function addonsValid(): boolean {
  const spu = picking.value
  if (!spu) return false
  return (spu.addonGroups || []).every((group) => {
    if (group.required !== 1) return true
    const picked = pickedAddons[group.id] || []
    return picked.length >= Math.max(1, group.minSelect)
  })
}

function confirmPick(): void {
  const spu = picking.value
  if (!spu) return
  const sku: MenuSku | undefined = spu.skus.find((item) => item.id === pickedSku.value)
  if (!sku) {
    useGlobalToast().warning({ msg: '请选择规格' })
    return
  }
  if (!addonsValid()) {
    useGlobalToast().warning({ msg: '请选择必选项' })
    return
  }

  const addons: CartAddon[] = []
  for (const group of spu.addonGroups || []) {
    for (const id of pickedAddons[group.id] || []) {
      const item = group.items.find((addon) => addon.id === id)
      if (item) {
        addons.push({ addonItemId: item.id, name: item.name, price: item.price, quantity: 1 })
      }
    }
  }

  cart.add({
    key: [sku.id, ...addons.map((addon) => addon.addonItemId).sort(), pickedRemark.value].join('|'),
    skuId: sku.id,
    spuName: spu.name,
    skuName: spu.skus.length > 1 ? sku.name : '',
    unitPrice: sku.price,
    addons,
    addonText: addons.map((addon) => addon.name).join('、'),
    remark: pickedRemark.value.trim(),
  })

  sheetVisible.value = false
  bump.value = true
  setTimeout(() => (bump.value = false), 320)
}

function toCart(): void {
  if (!cart.count) return
  useRouter().push({ name: 'cart' })
}

function goLogin(): void {
  useRouter().push({ name: 'login' })
}

onShow(() => {
  if (!session.isLogin) {
    goLogin()
    return
  }
  void load()
})
</script>

<template>
  <view class="paper">
    <!-- 页头：店名 + 日期，像一页菜单的刊头 -->
    <view class="masthead">
      <view class="masthead-top">
        <text class="kicker">自取 · 堂食</text>
        <text class="date">{{ new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric' }) }}</text>
      </view>
      <text class="title">今日菜单</text>
      <view class="rule">
        <view class="rule-thin" />
        <view class="rule-thick" />
      </view>
    </view>

    <!-- 加载：用墨线骨架，而不是转圈 -->
    <view v-if="loading" class="skeleton">
      <view v-for="i in 6" :key="i" class="sk-row" :style="{ animationDelay: `${i * 60}ms` }">
        <view class="sk-line w-40" />
        <view class="sk-line w-70" />
      </view>
    </view>

    <!-- 失败态：给出可执行的下一步 -->
    <view v-else-if="failed" class="state">
      <text class="state-title">菜单没能展开</text>
      <text class="state-desc">可能是网络，也可能是本店今日还没上菜</text>
      <view class="state-btn" @click="load">重新展开</view>
    </view>

    <view v-else class="body">
      <!-- 左：分类 -->
      <scroll-view class="rail" scroll-y>
        <view
          v-for="category in categories"
          :key="category.id"
          class="rail-item"
          :class="{ active: category.id === activeCategory }"
          @click="switchCategory(category.id)"
        >
          <text>{{ category.name }}</text>
        </view>
      </scroll-view>

      <!-- 右：菜品 -->
      <scroll-view class="list" scroll-y :scroll-into-view="scrollInto" scroll-with-animation>
        <view v-for="(category, ci) in categories" :id="`sec-${category.id}`" :key="category.id" class="section">
          <view class="section-head">
            <text class="section-name">{{ category.name }}</text>
            <text class="section-count">{{ dishesOf(category).length }}</text>
          </view>

          <view
            v-for="(spu, index) in dishesOf(category)"
            :key="spu.id"
            class="dish"
            :class="{ out: soldOutAll(spu) }"
            :style="{ animationDelay: `${(ci * 3 + index) * 28}ms` }"
            @click="openSheet(spu)"
          >
            <view class="dish-main">
              <view class="dish-line">
                <text class="dish-name">{{ spu.name }}</text>
                <view v-if="soldOutAll(spu)" class="stamp">售罄</view>
              </view>
              <text v-if="spu.description" class="dish-desc">{{ spu.description }}</text>
              <view class="dish-meta">
                <text class="price">
                  <text class="price-sym">¥</text>{{ yuan(lowestPrice(spu) === Number.POSITIVE_INFINITY ? 0 : lowestPrice(spu)) }}
                </text>
                <text v-if="spu.skus.length > 1" class="price-hint">起</text>
              </view>
            </view>

            <view v-if="!soldOutAll(spu)" class="plus" @click.stop="openSheet(spu)">
              <text class="plus-glyph">＋</text>
            </view>
          </view>
        </view>

        <view class="list-foot">
          <text>— 本店菜品以当日供应为准 —</text>
        </view>
      </scroll-view>
    </view>

    <!-- 底部票据式购物车 -->
    <view v-if="cart.count" class="ticket" :class="{ bump }" @click="toCart">
      <view class="perforation" />
      <view class="ticket-body">
        <view class="ticket-left">
          <text class="ticket-count">{{ cart.count }} 件</text>
          <text class="ticket-total">
            <text class="price-sym">¥</text>{{ yuan(cart.subtotal) }}
          </text>
        </view>
        <view class="ticket-btn">去结算</view>
      </view>
    </view>

    <!-- 规格与加料：像一张点菜小票 -->
    <wd-popup v-model="sheetVisible" position="bottom" custom-style="background:#F7F2E7;border-radius:18rpx 18rpx 0 0;">
      <view v-if="picking" class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">{{ picking.name }}</text>
          <text class="sheet-close" @click="sheetVisible = false">收起</text>
        </view>

        <text v-if="picking.skus.length > 1" class="sheet-label">规格</text>
        <view v-if="picking.skus.length > 1" class="chips">
          <view
            v-for="sku in picking.skus"
            :key="sku.id"
            class="chip"
            :class="{ on: pickedSku === sku.id, disabled: sku.soldOut === 1 }"
            @click="sku.soldOut !== 1 && (pickedSku = sku.id)"
          >
            <text>{{ sku.name }}</text>
            <text class="chip-price">¥{{ yuan(sku.price) }}</text>
          </view>
        </view>

        <template v-for="group in picking.addonGroups || []" :key="group.id">
          <view class="sheet-label-row">
            <text class="sheet-label">{{ group.name }}</text>
            <text v-if="group.required === 1" class="require">必选</text>
          </view>
          <view class="chips">
            <view
              v-for="addon in group.items"
              :key="addon.id"
              class="chip"
              :class="{ on: (pickedAddons[group.id] || []).includes(addon.id) }"
              @click="toggleAddon(group.id, addon.id, group.maxSelect)"
            >
              <text>{{ addon.name }}</text>
              <text class="chip-price">+¥{{ yuan(addon.price) }}</text>
            </view>
          </view>
        </template>

        <text class="sheet-label">口味备注</text>
        <input v-model="pickedRemark" class="sheet-input" placeholder="少辣、不要葱…" placeholder-class="ph" />

        <view class="sheet-submit" @click="confirmPick">
          <text>加入菜单</text>
          <text class="sheet-submit-price">¥{{ yuan(picking.skus.find((s) => s.id === pickedSku)?.price || 0) }}</text>
        </view>
      </view>
    </wd-popup>
  </view>
</template>

<style lang="scss" scoped>
/* 纸墨色板：整套页面只认这几个变量 */
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
  background:
    radial-gradient(120% 60% at 100% 0%, #fbf7ee 0%, $paper 55%),
    $paper;
  color: $ink;
}

/* 刊头 */
.masthead {
  padding: 28rpx 32rpx 0;
}
.masthead-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}
.kicker {
  font-size: 22rpx;
  color: $cinnabar;
  letter-spacing: 4rpx;
}
.date {
  font-size: 22rpx;
  color: $ink-faint;
  letter-spacing: 1rpx;
}
.title {
  display: block;
  margin-top: 10rpx;
  font-family: $serif;
  font-size: 62rpx;
  font-weight: 600;
  line-height: 1.15;
  letter-spacing: 6rpx;
}
.rule {
  margin: 22rpx 0 6rpx;
}
.rule-thin {
  height: 1px;
  background: rgba(23, 19, 15, 0.28);
}
.rule-thick {
  height: 6rpx;
  margin-top: 5rpx;
  background: $ink;
}

/* 骨架 */
.skeleton {
  padding: 24rpx 32rpx;
}
.sk-row {
  margin-bottom: 34rpx;
  animation: ink-in 0.5s both;
}
.sk-line {
  height: 20rpx;
  margin-bottom: 12rpx;
  background: $paper-deep;
  border-radius: 4rpx;
}
.w-40 { width: 40%; }
.w-70 { width: 70%; }

/* 状态 */
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
  margin-top: 14rpx;
  font-size: 24rpx;
  color: $ink-faint;
  text-align: center;
}
.state-btn {
  margin-top: 36rpx;
  padding: 18rpx 46rpx;
  font-size: 26rpx;
  color: $paper;
  letter-spacing: 2rpx;
  background: $ink;
  border-radius: 999rpx;
}

/* 主体两栏 */
.body {
  display: flex;
  flex: 1;
  min-height: 0;
}
.rail {
  width: 150rpx;
  padding: 12rpx 0;
  border-right: 1px solid rgba(23, 19, 15, 0.1);
}
.rail-item {
  position: relative;
  padding: 26rpx 12rpx 26rpx 26rpx;
  font-size: 25rpx;
  color: $ink-soft;
  letter-spacing: 2rpx;
}
.rail-item.active {
  font-weight: 600;
  color: $ink;
}
.rail-item.active::before {
  position: absolute;
  top: 50%;
  left: 0;
  width: 5rpx;
  height: 30rpx;
  content: '';
  background: $cinnabar;
  transform: translateY(-50%);
}

.list {
  flex: 1;
  min-width: 0;
  padding: 0 32rpx 220rpx 26rpx;
}
.section-head {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  padding: 26rpx 0 8rpx;
}
.section-name {
  font-family: $serif;
  font-size: 30rpx;
  font-weight: 600;
  letter-spacing: 3rpx;
}
.section-count {
  font-size: 20rpx;
  color: $ink-faint;
}

/* 菜品行 */
.dish {
  display: flex;
  gap: 16rpx;
  align-items: flex-start;
  padding: 26rpx 0;
  border-bottom: 1px solid rgba(23, 19, 15, 0.08);
  animation: ink-in 0.45s both;
}
.dish.out {
  opacity: 0.45;
}
.dish-main {
  flex: 1;
  min-width: 0;
}
.dish-line {
  display: flex;
  align-items: center;
  gap: 14rpx;
}
.dish-name {
  font-family: $serif;
  font-size: 33rpx;
  font-weight: 600;
  letter-spacing: 1rpx;
}
.stamp {
  padding: 2rpx 12rpx;
  font-size: 20rpx;
  color: $cinnabar;
  letter-spacing: 2rpx;
  border: 2rpx solid rgba(178, 58, 47, 0.6);
  border-radius: 4rpx;
  transform: rotate(-4deg);
}
.dish-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 23rpx;
  line-height: 1.5;
  color: $ink-faint;
}
.dish-meta {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
  margin-top: 12rpx;
}
.price {
  font-family: $serif;
  font-size: 36rpx;
  font-weight: 600;
  letter-spacing: 0.5rpx;
}
.price-sym {
  margin-right: 2rpx;
  font-size: 22rpx;
  font-weight: 400;
}
.price-hint {
  font-size: 20rpx;
  color: $ink-faint;
}

/* 加号：只用线条，不加阴影 */
.plus {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56rpx;
  height: 56rpx;
  margin-top: 6rpx;
  border: 2rpx solid $ink;
  border-radius: 50%;
  transition: transform 0.12s;
}
.plus:active {
  background: $ink;
  transform: scale(0.92);
}
.plus:active .plus-glyph {
  color: $paper;
}
.plus-glyph {
  margin-top: -4rpx;
  font-size: 30rpx;
  line-height: 1;
}

.list-foot {
  padding: 40rpx 0 20rpx;
  font-size: 20rpx;
  color: $ink-faint;
  letter-spacing: 2rpx;
  text-align: center;
}

/* 票据式购物车 */
.ticket {
  position: fixed;
  right: 28rpx;
  bottom: 130rpx;
  left: 28rpx;
  z-index: 20;
  transition: transform 0.18s ease-out;
}
.ticket.bump {
  transform: scale(1.02);
}
/* 撕口：一排半圆模拟票据的齿孔 */
.perforation {
  height: 14rpx;
  background-image: radial-gradient(circle at 7rpx 14rpx, transparent 7rpx, $ink 7rpx);
  background-size: 14rpx 14rpx;
}
.ticket-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 24rpx;
  background: $ink;
  border-radius: 0 0 16rpx 16rpx;
}
.ticket-left {
  display: flex;
  align-items: baseline;
  gap: 14rpx;
  color: $paper;
}
.ticket-count {
  font-size: 22rpx;
  letter-spacing: 1rpx;
  opacity: 0.7;
}
.ticket-total {
  font-family: $serif;
  font-size: 40rpx;
  font-weight: 600;
}
.ticket-btn {
  padding: 14rpx 34rpx;
  font-size: 26rpx;
  color: #fff;
  letter-spacing: 2rpx;
  background: $cinnabar;
  border-radius: 999rpx;
}

/* 规格弹层 */
.sheet {
  max-height: 76vh;
  padding: 32rpx 32rpx calc(32rpx + env(safe-area-inset-bottom));
  overflow-y: auto;
}
.sheet-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-bottom: 16rpx;
  border-bottom: 1px solid rgba(23, 19, 15, 0.12);
}
.sheet-title {
  font-family: $serif;
  font-size: 38rpx;
  font-weight: 600;
}
.sheet-close {
  font-size: 24rpx;
  color: $ink-faint;
}
.sheet-label {
  display: block;
  margin: 26rpx 0 14rpx;
  font-size: 24rpx;
  color: $ink-soft;
  letter-spacing: 2rpx;
}
.sheet-label-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.require {
  padding: 1rpx 8rpx;
  font-size: 18rpx;
  color: $cinnabar;
  border: 1px solid rgba(178, 58, 47, 0.5);
  border-radius: 3rpx;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 14rpx;
}
.chip {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
  padding: 14rpx 22rpx;
  font-size: 25rpx;
  color: $ink;
  background: $paper;
  border: 1px solid rgba(23, 19, 15, 0.2);
  border-radius: 8rpx;
}
.chip.on {
  color: #fff;
  background: $cinnabar;
  border-color: $cinnabar;
}
.chip.disabled {
  color: $ink-faint;
  text-decoration: line-through;
  opacity: 0.6;
}
.chip-price {
  font-size: 21rpx;
  opacity: 0.75;
}
.sheet-input {
  width: 100%;
  padding: 18rpx 20rpx;
  font-size: 26rpx;
  background: $paper-deep;
  border-radius: 8rpx;
}
.ph {
  color: $ink-faint;
}
.sheet-submit {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 26rpx 32rpx;
  margin-top: 34rpx;
  font-size: 28rpx;
  color: #fff;
  letter-spacing: 2rpx;
  background: $ink;
  border-radius: 12rpx;
}
.sheet-submit-price {
  font-family: $serif;
  font-size: 34rpx;
}

@keyframes ink-in {
  from {
    opacity: 0;
    transform: translateY(12rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
