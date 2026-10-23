<!-- 后厨看板：待接单 / 制作中 / 待取餐三列，超时高亮，新单响铃 -->
<template>
  <div class="board">
    <header class="bar">
      <div class="bar-left">
        <span class="brand">后厨看板</span>
        <span class="store">{{ storeName }}</span>
        <span class="dot" :class="online ? 'ok' : 'bad'">
          {{ online ? '已连接' : '连接中断，重试中' }}
        </span>
      </div>

      <div class="bar-right">
        <span class="summary">
          待接 {{ countOf(200) }} · 制作 {{ countOf(300) }} · 待取 {{ countOf(400) }}
        </span>
        <button class="tool" type="button" @click="toggleSound">
          {{ soundOn ? '声音：开' : '声音：关' }}
        </button>
        <button class="tool" type="button" @click="toggleFullscreen">全屏</button>
        <button class="tool" type="button" :disabled="loading" @click="manualRefresh">刷新</button>
        <button class="tool danger" type="button" @click="logout">退出</button>
      </div>
    </header>

    <main class="columns">
      <section v-for="col in columns" :key="col.status" class="column">
        <div class="col-head">
          <span>{{ col.title }}</span>
          <span class="badge">{{ col.list.length }}</span>
        </div>

        <div class="col-body">
          <article
            v-for="order in col.list"
            :key="order.id"
            class="card"
            :class="cardClass(order)"
          >
            <div class="card-head">
              <span class="table">{{ order.tableName || order.pickupCode || order.typeText }}</span>
              <span class="elapsed">{{ elapsedText(order) }}</span>
            </div>

            <div class="card-sub">
              <span>{{ order.no }}</span>
              <span v-if="order.typeText">{{ order.typeText }}</span>
            </div>

            <ul class="items">
              <li v-for="(item, index) in order.items" :key="index">
                <span class="qty">{{ item.quantity }}×</span>
                <span class="name">
                  {{ item.spuName }}<em v-if="item.skuName">（{{ item.skuName }}）</em>
                </span>
                <span v-if="item.addons.length" class="addons">{{ item.addons.join('、') }}</span>
                <span v-if="item.remark" class="item-remark">{{ item.remark }}</span>
              </li>
            </ul>

            <p v-if="order.remark" class="order-remark">备注：{{ order.remark }}</p>

            <button
              class="act"
              type="button"
              :disabled="!!acting[order.id]"
              @click="move(order, col.next)"
            >
              {{ acting[order.id] ? '处理中…' : col.action }}
            </button>
          </article>

          <p v-if="!col.list.length" class="empty">暂无订单</p>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
  import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'

  import { fetchVisibleStores } from '@/api/auth'
  import { fetchKdsBoard, transitOrder, type KdsOrder } from '@/api/kds'
  import { REFRESH_KEY, STORE_KEY, TOKEN_KEY, notify } from '@/utils/http'

  /** 轮询间隔：后厨要新单尽快出现，8 秒是响应速度与请求量的折中 */
  const POLL_MS = 8000
  /** 超时阈值（分钟）：先黄后红，让厨师一眼看到压单 */
  const WARN_MIN = 10
  const DANGER_MIN = 20

  const router = useRouter()

  const orders = ref<KdsOrder[]>([])
  const loading = ref(false)
  const online = ref(true)
  const soundOn = ref(true)
  const storeName = ref('')
  const nowTs = ref(Date.now())
  const acting = ref<Record<string, boolean>>({})

  // 已出现过的订单：用于判断「新到单」，首屏不响铃
  const known = new Set<string>()
  let bootstrap = true
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let tickTimer: ReturnType<typeof setInterval> | null = null

  const columns = computed(() => [
    { status: 200, title: '待接单', action: '接单', next: 300, list: byStatus(200) },
    { status: 300, title: '制作中', action: '出餐', next: 400, list: byStatus(300) },
    { status: 400, title: '待取餐', action: '完成', next: 500, list: byStatus(400) }
  ])

  function byStatus(status: number): KdsOrder[] {
    return orders.value.filter((item) => item.status === status)
  }

  const countOf = (status: number): number => byStatus(status).length

  /** 支付时间优先：后厨计时从收款那一刻算，未支付（外卖挂单）退回下单时间 */
  const startedAt = (order: KdsOrder): number => {
    const raw = order.payTime || order.createdAt
    if (!raw) return Date.now()
    // Go 的 time.Time 默认序列化成「2026-10-23 14:05:33」，补 T 才能被各浏览器一致解析
    const text = raw.includes('T') ? raw : raw.replace(' ', 'T')
    const time = new Date(text).getTime()

    return Number.isNaN(time) ? Date.now() : time
  }

  const elapsedMinutes = (order: KdsOrder): number =>
    Math.max(0, Math.floor((nowTs.value - startedAt(order)) / 60000))

  const elapsedText = (order: KdsOrder): string => {
    const minutes = elapsedMinutes(order)
    if (minutes < 60) return `${minutes} 分钟`

    return `${Math.floor(minutes / 60)} 时 ${minutes % 60} 分`
  }

  const cardClass = (order: KdsOrder): string => {
    const minutes = elapsedMinutes(order)
    if (minutes >= DANGER_MIN) return 'danger'
    if (minutes >= WARN_MIN) return 'warn'

    return ''
  }

  /** 新单提示音：用 WebAudio 合成，避免额外音频文件与跨域问题 */
  const beep = (): void => {
    if (!soundOn.value) return
    try {
      const Ctor =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
      if (!Ctor) return

      const ctx = new Ctor()
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      osc.connect(gain)
      gain.connect(ctx.destination)
      osc.frequency.value = 880
      gain.gain.value = 0.08
      osc.start()
      setTimeout(() => {
        osc.stop()
        void ctx.close()
      }, 220)
    } catch {
      // 浏览器自动播放策略可能拦截，静默失败即可
    }
  }

  const load = async (silent = false): Promise<void> => {
    if (!silent) loading.value = true
    try {
      const res = await fetchKdsBoard()
      online.value = true

      let added = 0
      for (const order of res) {
        if (!known.has(order.id)) added++
      }
      known.clear()
      for (const order of res) known.add(order.id)

      orders.value = res
      // 首次加载是「看到存量」，不该响铃，否则每次进页面都被吓一跳
      if (!bootstrap && added > 0) {
        beep()
        notify(`新到 ${added} 单`, 'success')
      }
      bootstrap = false
    } catch {
      online.value = false
    } finally {
      loading.value = false
    }
  }

  const manualRefresh = (): void => {
    void load()
  }

  const move = async (order: KdsOrder, next: number): Promise<void> => {
    if (acting.value[order.id]) return

    acting.value = { ...acting.value, [order.id]: true }
    try {
      await transitOrder(order.id, next)
      await load(true)
    } catch {
      // 提示已由请求层弹出；失败不本地改状态，等下一轮轮询纠正
    } finally {
      const rest = { ...acting.value }
      delete rest[order.id]
      acting.value = rest
    }
  }

  const toggleSound = (): void => {
    soundOn.value = !soundOn.value
  }

  const toggleFullscreen = (): void => {
    if (document.fullscreenElement) {
      void document.exitFullscreen()

      return
    }
    void document.documentElement.requestFullscreen()
  }

  const logout = (): void => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
    localStorage.removeItem(STORE_KEY)
    void router.replace('/login')
  }

  onMounted(async () => {
    // 顶栏显示门店名，取不到就留空，不影响看板
    try {
      const stores = await fetchVisibleStores()
      const current = stores.find((item) => item.id === localStorage.getItem(STORE_KEY))
      storeName.value = current?.name || stores[0]?.name || ''
      if (!localStorage.getItem(STORE_KEY) && stores[0]) {
        localStorage.setItem(STORE_KEY, stores[0].id)
      }
    } catch {
      storeName.value = ''
    }

    await load()
    pollTimer = setInterval(() => void load(true), POLL_MS)
    tickTimer = setInterval(() => {
      nowTs.value = Date.now()
    }, 1000)
  })

  onBeforeUnmount(() => {
    if (pollTimer) clearInterval(pollTimer)
    if (tickTimer) clearInterval(tickTimer)
  })
</script>

<style scoped>
  .board {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    background: #1c1f26;
    border-bottom: 1px solid #2a2e37;
  }

  .bar-left,
  .bar-right {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .brand {
    font-size: 18px;
    font-weight: 600;
  }

  .store {
    font-size: 14px;
    color: #a9b1bf;
  }

  .dot {
    padding: 2px 8px;
    font-size: 12px;
    border-radius: 10px;
  }

  .dot.ok {
    color: #18a058;
    background: rgb(24 160 88 / 12%);
  }

  .dot.bad {
    color: #ff7875;
    background: rgb(255 120 117 / 12%);
  }

  .summary {
    font-size: 14px;
    color: #a9b1bf;
  }

  .tool {
    padding: 6px 12px;
    font-size: 13px;
    color: #d5dae3;
    background: #23262f;
    border: 1px solid #2f3541;
    border-radius: 6px;
  }

  .tool.danger {
    color: #ff7875;
  }

  .columns {
    display: grid;
    flex: 1;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
    min-height: 0;
    padding: 10px;
  }

  .column {
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: #171a20;
    border: 1px solid #23262f;
    border-radius: 10px;
  }

  .col-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    font-size: 16px;
    font-weight: 600;
    border-bottom: 1px solid #23262f;
  }

  .badge {
    padding: 1px 10px;
    font-size: 13px;
    color: #14161a;
    background: #d5dae3;
    border-radius: 10px;
  }

  .col-body {
    flex: 1;
    padding: 10px;
    overflow-y: auto;
  }

  .card {
    padding: 12px;
    margin-bottom: 10px;
    background: #1f232b;
    border: 1px solid #2a2e37;
    border-left: 4px solid #3f8cff;
    border-radius: 8px;
  }

  .card.warn {
    border-left-color: #f0a020;
  }

  .card.danger {
    background: #2a1d1f;
    border-left-color: #ff5c5c;
  }

  .card-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .table {
    font-size: 20px;
    font-weight: 700;
  }

  .elapsed {
    font-size: 14px;
    color: #a9b1bf;
  }

  .card.danger .elapsed {
    font-weight: 700;
    color: #ff7875;
  }

  .card-sub {
    display: flex;
    justify-content: space-between;
    margin-top: 2px;
    font-size: 12px;
    color: #7b8494;
  }

  .items {
    padding: 0;
    margin: 10px 0 0;
    list-style: none;
  }

  .items li {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: baseline;
    padding: 5px 0;
    font-size: 17px;
    border-bottom: 1px dashed #2a2e37;
  }

  .qty {
    font-weight: 700;
    color: #3f8cff;
  }

  .name em {
    font-size: 13px;
    color: #8b93a1;
  }

  .addons {
    font-size: 13px;
    color: #f0a020;
  }

  .item-remark {
    font-size: 13px;
    color: #ff7875;
  }

  .order-remark {
    margin: 8px 0 0;
    font-size: 13px;
    color: #ff7875;
  }

  .act {
    width: 100%;
    padding: 10px;
    margin-top: 12px;
    font-size: 15px;
    font-weight: 600;
    color: #fff;
    background: #18a058;
    border: none;
    border-radius: 6px;
  }

  .act:disabled {
    opacity: 0.6;
  }

  .empty {
    padding: 24px 0;
    font-size: 14px;
    color: #5c6472;
    text-align: center;
  }
</style>
