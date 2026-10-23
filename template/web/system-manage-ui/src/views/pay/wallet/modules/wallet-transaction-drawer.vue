<!-- 钱包流水抽屉：按收支方向与业务类型筛选，展示变动金额与变动后余额 -->
<template>
  <ElDrawer v-model="drawerVisible" title="钱包流水" size="720px">
    <div class="mb-4 flex items-center gap-3">
      <span class="text-sm text-g-600">
        用户 {{ props.wallet?.userId ?? '' }} · 可用余额
        {{ formatYuan(props.wallet?.balance ?? 0) }} 元
      </span>
      <ElSelect
        v-model="filter.type"
        class="w-32"
        clearable
        placeholder="收支方向"
        @change="loadData"
      >
        <ElOption label="收入" :value="1" />
        <ElOption label="支出" :value="2" />
      </ElSelect>
      <ElSelect
        v-model="filter.bizType"
        class="w-36"
        clearable
        placeholder="业务类型"
        @change="loadData"
      >
        <ElOption label="充值" :value="1" />
        <ElOption label="充值退款" :value="2" />
        <ElOption label="支付" :value="3" />
        <ElOption label="支付退款" :value="4" />
        <ElOption label="手工调整" :value="5" />
      </ElSelect>
      <ElButton
        v-auth="'pay:wallet:list'"
        class="ml-auto"
        :loading="exporting"
        @click="exportTransactions"
      >
        导出 CSV
      </ElButton>
    </div>

    <ElTable v-loading="loading" :data="list" size="small" border max-height="560">
      <ElTableColumn prop="createdAt" label="时间" width="170" />
      <ElTableColumn label="方向" width="90">
        <template #default="{ row }">
          <ElTag :type="row.type === 1 ? 'success' : 'warning'" effect="plain">
            {{ row.type === 1 ? '收入' : '支出' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="金额（元）" width="120">
        <template #default="{ row }">
          <span :class="row.type === 1 ? 'text-success' : 'text-danger'">
            {{ row.type === 1 ? '+' : '-' }}{{ formatYuan(row.price) }}
          </span>
        </template>
      </ElTableColumn>
      <ElTableColumn label="变动后余额（元）" width="140">
        <template #default="{ row }">{{ formatYuan(row.balance) }}</template>
      </ElTableColumn>
      <ElTableColumn label="业务类型" width="110">
        <template #default="{ row }">{{ BIZ_TYPE_LABELS[row.bizType] ?? '未知' }}</template>
      </ElTableColumn>
      <ElTableColumn prop="remark" label="备注" min-width="160" show-overflow-tooltip />
    </ElTable>

    <div class="mt-3 flex justify-end">
      <ElPagination
        v-model:current-page="filter.pageNo"
        v-model:page-size="filter.pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, prev, pager, next"
        small
        @current-change="loadData"
      />
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { fetchExportPayWalletTransactions, fetchGetPayWalletTransactions } from '@/api/pay/wallet'
  import { exportFilename, saveBlob } from '@/utils/download'

  defineOptions({ name: 'PayWalletTransactionDrawer' })

  interface Props {
    visible: boolean
    wallet?: Api.Pay.WalletItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    wallet: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const exporting = ref(false)
  const list = ref<Api.Pay.WalletTransactionItem[]>([])
  const total = ref(0)

  // 业务类型文案，取值对齐 yudao PayWalletBizTypeEnum（5 为后台手工调整）
  const BIZ_TYPE_LABELS: Record<number, string> = {
    1: '充值',
    2: '充值退款',
    3: '支付',
    4: '支付退款',
    5: '手工调整'
  }

  const filter = reactive({
    type: undefined as number | undefined,
    bizType: undefined as number | undefined,
    pageNo: 1,
    pageSize: 20
  })

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 按当前筛选条件加载流水分页
  const loadData = async (): Promise<void> => {
    if (!props.wallet) return

    loading.value = true
    try {
      const res = await fetchGetPayWalletTransactions(props.wallet.id, {
        type: filter.type,
        bizType: filter.bizType,
        pageNo: filter.pageNo,
        pageSize: filter.pageSize
      })
      list.value = res?.list ?? []
      total.value = res?.total ?? 0
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  // 按当前筛选条件导出该会员的钱包流水，用于余额对账
  const exportTransactions = async (): Promise<void> => {
    if (!props.wallet) return

    exporting.value = true
    try {
      const blob = await fetchExportPayWalletTransactions(props.wallet.id, {
        type: filter.type,
        bizType: filter.bizType
      })
      saveBlob(blob, exportFilename(`wallet-${props.wallet.id}-transactions`))
      ElMessage.success('导出成功')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      exporting.value = false
    }
  }

  // 打开抽屉时重置筛选并加载首页
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      filter.type = undefined
      filter.bizType = undefined
      filter.pageNo = 1
      await loadData()
    },
    { immediate: true }
  )
</script>
