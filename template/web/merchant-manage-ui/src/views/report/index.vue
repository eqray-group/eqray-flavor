<!-- 经营报表：区间营收趋势、现金与线上构成、菜品销量排行 -->
<template>
  <div class="page">
    <div class="toolbar">
      <ElRadioGroup v-model="days" @change="loadData">
        <ElRadioButton :value="7">近 7 天</ElRadioButton>
        <ElRadioButton :value="14">近 14 天</ElRadioButton>
        <ElRadioButton :value="30">近 30 天</ElRadioButton>
      </ElRadioGroup>
      <ElButton :loading="loading" @click="loadData">刷新</ElButton>
      <span v-if="report" class="range">{{ report.from }} ~ {{ report.to }}</span>
    </div>

    <div class="stats">
      <div class="stat">
        <div class="stat-label">营业额</div>
        <div class="stat-value">¥{{ yuan(report?.revenue ?? 0) }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">已支付单</div>
        <div class="stat-value">{{ report?.orders ?? 0 }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">客单价</div>
        <div class="stat-value">¥{{ yuan(report?.avgPrice ?? 0) }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">退款</div>
        <div class="stat-value">¥{{ yuan(report?.refundAmount ?? 0) }}</div>
      </div>
    </div>

    <div class="panel-row">
      <ElCard class="panel" shadow="never">
        <template #header>
          <span class="panel-title">收款构成</span>
        </template>
        <div v-for="item in paymentRows" :key="item.label" class="pay-row">
          <div class="pay-head">
            <span>{{ item.label }}</span>
            <span class="pay-amount">¥{{ yuan(item.amount) }} · {{ item.orders }} 单</span>
          </div>
          <div class="bar-track">
            <div class="bar-fill" :class="item.className" :style="{ width: item.percent + '%' }" />
          </div>
          <div class="pay-percent">{{ item.percent }}%</div>
        </div>
        <ElEmpty v-if="!report?.orders" description="区间内没有收款" :image-size="56" />
      </ElCard>

      <ElCard class="panel" shadow="never">
        <template #header>
          <span class="panel-title">菜品销量 Top {{ top }}</span>
        </template>
        <ElEmpty v-if="!report?.topDishes.length" description="区间内没有销量数据" :image-size="56" />
        <ElTable v-else :data="report.topDishes" size="small" border>
          <ElTableColumn type="index" label="#" width="46" />
          <ElTableColumn label="菜品" min-width="140">
            <template #default="{ row }">
              {{ row.spuName }}<span v-if="row.skuName" class="sku">（{{ row.skuName }}）</span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="quantity" label="销量" width="80" />
          <ElTableColumn label="金额" width="100">
            <template #default="{ row }">¥{{ yuan(row.amount) }}</template>
          </ElTableColumn>
        </ElTable>
      </ElCard>
    </div>

    <ElCard shadow="never">
      <template #header>
        <span class="panel-title">每日营收</span>
      </template>
      <ElEmpty v-if="!report?.trend.length" description="暂无数据" :image-size="56" />
      <div v-else class="trend">
        <div v-for="item in report.trend" :key="item.date" class="trend-row">
          <span class="trend-date">{{ item.date.slice(5) }}</span>
          <div class="bar-track">
            <div class="bar-fill primary" :style="{ width: barWidth(item.revenue) + '%' }" />
          </div>
          <span class="trend-value">¥{{ yuan(item.revenue) }}</span>
          <span class="trend-orders">{{ item.orders }} 单</span>
        </div>
      </div>
    </ElCard>

    <ElCard class="mt-12" shadow="never">
      <template #header>
        <span class="panel-title">时段分布</span>
        <span class="panel-hint">悬停查看该时段金额与单量</span>
      </template>
      <div class="hour-chart">
        <div
          v-for="item in report?.hourly ?? []"
          :key="item.hour"
          class="hour-col"
          :title="`${item.hour}:00  ¥${yuan(item.revenue)}  ${item.orders} 单`"
        >
          <div class="hour-bar-wrap">
            <div class="hour-bar" :style="{ height: hourHeight(item.revenue) + '%' }" />
          </div>
          <span class="hour-label">{{ item.hour }}</span>
        </div>
      </div>
    </ElCard>

    <ElCard class="mt-12" shadow="never">
      <template #header>
        <span class="panel-title">退款明细</span>
        <span class="panel-hint">按订单维度统计，现金退款同样在内</span>
      </template>
      <ElEmpty v-if="!report?.refunds.length" description="区间内没有退款" :image-size="56" />
      <ElTable v-else :data="report.refunds" size="small" border>
        <ElTableColumn prop="orderNo" label="订单号" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="应收" width="100">
          <template #default="{ row }">¥{{ yuan(row.payAmount) }}</template>
        </ElTableColumn>
        <ElTableColumn label="已退" width="100">
          <template #default="{ row }">¥{{ yuan(row.refundAmount) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="statusText" label="状态" width="100" />
        <ElTableColumn prop="updatedAt" label="最后更新" width="180" />
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'

  import { fetchPosReport } from '@/api/order'

  const loading = ref(false)
  const days = ref(7)
  const top = 10
  const report = ref<Api.Order.PosReport | null>(null)

  const yuan = (cent: number): string => ((cent || 0) / 100).toFixed(2)

  // 收款构成：现金与线上各自的金额占比，两边加起来不一定等于营业额（可能有退款），只做相对比较
  const paymentRows = computed(() => {
    const value = report.value
    const cash = value?.cashAmount ?? 0
    const online = value?.onlineAmount ?? 0
    const sum = cash + online
    const percent = (amount: number): number => (sum > 0 ? Math.round((amount / sum) * 100) : 0)

    return [
      {
        label: '现金',
        amount: cash,
        orders: value?.cashOrders ?? 0,
        percent: percent(cash),
        className: 'cash'
      },
      {
        label: '微信 / 支付宝',
        amount: online,
        orders: value?.onlineOrders ?? 0,
        percent: percent(online),
        className: 'online'
      }
    ]
  })

  // 柱宽按区间内最高营收归一化，最低留 2% 让「有单但额小」的日期也能看见
  const barWidth = (revenue: number): number => {
    const max = Math.max(...(report.value?.trend ?? []).map((item) => item.revenue), 0)
    if (max <= 0 || revenue <= 0) return 0

    return Math.max(2, Math.round((revenue / max) * 100))
  }

  // 时段柱高同样按区间内最高时段归一化，有单但额小的时段至少留 4% 可见
  const hourHeight = (revenue: number): number => {
    const max = Math.max(...(report.value?.hourly ?? []).map((item) => item.revenue), 0)
    if (max <= 0 || revenue <= 0) return 0

    return Math.max(4, Math.round((revenue / max) * 100))
  }

  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      report.value = await fetchPosReport(days.value, top)
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .range {
    margin-left: auto;
    font-size: 13px;
    color: #6b7280;
  }

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

  .stat-label {
    font-size: 12px;
    color: #6b7280;
  }

  .stat-value {
    margin-top: 6px;
    font-size: 20px;
    font-weight: 600;
  }

  .panel-row {
    display: grid;
    grid-template-columns: minmax(280px, 1fr) minmax(320px, 1.2fr);
    gap: 12px;
    margin-bottom: 12px;
  }

  .panel-title {
    font-size: 14px;
    font-weight: 600;
  }

  .pay-row {
    margin-bottom: 14px;
  }

  .pay-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 6px;
    font-size: 13px;
  }

  .pay-amount {
    font-weight: 600;
  }

  .bar-track {
    flex: 1;
    height: 10px;
    overflow: hidden;
    background: #f0f2f5;
    border-radius: 6px;
  }

  .bar-fill {
    height: 100%;
    border-radius: 6px;
    transition: width 0.3s;
  }

  .bar-fill.cash {
    background: #67c23a;
  }

  .bar-fill.online {
    background: #409eff;
  }

  .bar-fill.primary {
    background: #409eff;
  }

  .pay-percent {
    margin-top: 4px;
    font-size: 12px;
    color: #909399;
    text-align: right;
  }

  .trend-row {
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 4px 0;
    font-size: 13px;
  }

  .trend-date {
    width: 52px;
    color: #6b7280;
  }

  .trend-value {
    width: 92px;
    text-align: right;
  }

  .trend-orders {
    width: 56px;
    font-size: 12px;
    color: #909399;
    text-align: right;
  }

  .sku {
    color: #909399;
  }

  .mt-12 {
    margin-top: 12px;
  }

  .panel-hint {
    margin-left: 8px;
    font-size: 12px;
    font-weight: 400;
    color: #909399;
  }

  .hour-chart {
    display: flex;
    gap: 4px;
    align-items: flex-end;
    height: 160px;
  }

  .hour-col {
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: flex-end;
    height: 100%;
  }

  .hour-bar-wrap {
    display: flex;
    align-items: flex-end;
    height: 100%;
  }

  .hour-bar {
    width: 100%;
    background: #409eff;
    border-radius: 3px 3px 0 0;
    transition: height 0.3s;
  }

  .hour-label {
    margin-top: 4px;
    font-size: 11px;
    color: #909399;
    text-align: center;
  }
</style>
