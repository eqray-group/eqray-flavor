<!-- POS 点单：左分类、中菜品、右购物车；金额只信服务端试算 -->
<template>
  <div class="pos">
    <aside class="cats">
      <div
        v-for="cat in categories"
        :key="cat.id"
        class="cat"
        :class="{ active: cat.id === activeCategory }"
        @click="activeCategory = cat.id"
      >
        {{ cat.name }}
      </div>
    </aside>

    <main v-loading="menuLoading" class="dishes">
      <ElEmpty v-if="!spus.length" description="该分类暂无可售菜品" />
      <div class="dish-grid">
        <div
          v-for="spu in spus"
          :key="spu.id"
          class="dish"
          :class="{ off: isSoldOut(spu) }"
          @click="pick(spu)"
        >
          <div class="dish-name">{{ spu.name }}</div>
          <div class="dish-price">{{ priceText(spu) }}</div>
          <div v-if="isSoldOut(spu)" class="dish-off">已沽清</div>
        </div>
      </div>
    </main>

    <aside class="cart">
      <div class="cart-head">
        <div v-if="appendOrderId" class="append-bar">
          加菜模式 · {{ appendOrderNo || appendOrderId }}
          <ElButton link type="danger" size="small" @click="exitAppend">退出</ElButton>
        </div>

        <div class="member-bar">
          <template v-if="member">
            <span class="member-name">{{ member.nickname || member.mobile }}</span>
            <span class="member-point">积分 {{ member.point }}</span>
            <span class="member-balance">余额 ¥{{ yuan(walletBalance) }}</span>
            <ElButton link type="primary" size="small" @click="openRecharge">充值</ElButton>
            <ElButton link type="danger" size="small" @click="clearMember">移除</ElButton>
          </template>
          <ElButton v-else link type="primary" size="small" @click="openMember">识别会员（选填）</ElButton>
        </div>

        <ElRadioGroup v-model="orderType" size="small">
          <ElRadioButton :value="1">堂食</ElRadioButton>
          <ElRadioButton :value="3">自提</ElRadioButton>
        </ElRadioGroup>

        <div class="hold-row">
          <ElButton size="small" :disabled="!cart.length" @click="holdCurrent">挂单</ElButton>
          <ElButton size="small" :disabled="!held.length" @click="heldVisible = true">
            取单{{ held.length ? `（${held.length}）` : '' }}
          </ElButton>
        </div>
        <ElSelect
          v-if="orderType === 1"
          v-model="tableId"
          class="mt-2"
          placeholder="选择桌台"
          size="small"
          filterable
        >
          <ElOption
            v-for="table in tables"
            :key="table.id"
            :label="`${table.areaName || ''} ${table.name}（${table.seats}座）`"
            :value="table.id"
          />
        </ElSelect>
      </div>

      <div class="cart-lines">
        <ElEmpty v-if="!cart.length" description="请选择菜品" :image-size="56" />
        <div v-for="line in cart" :key="line.key" class="line">
          <div class="line-main">
            <div class="line-name">
              {{ line.spuName }}<span v-if="line.skuName">（{{ line.skuName }}）</span>
            </div>
            <div v-if="line.addonText" class="line-addon">{{ line.addonText }}</div>
            <div v-if="line.remark" class="line-remark">备注：{{ line.remark }}</div>
          </div>
          <div class="line-side">
            <ElInputNumber
              v-model="line.quantity"
              :min="1"
              :max="99"
              size="small"
              controls-position="right"
              @change="scheduleEstimate"
            />
            <div class="line-sub">{{ yuan(linePrice(line)) }}</div>
            <ElButton link type="danger" size="small" @click="removeLine(line)">删除</ElButton>
          </div>
        </div>
      </div>

      <div class="cart-foot">
        <ElInput v-model="remark" placeholder="整单备注（选填）" size="small" />
        <div v-if="member" class="points-row">
          <span class="hint">积分抵扣</span>
          <ElSwitch v-model="usePoints" size="small" @change="scheduleEstimate" />
        </div>
        <div class="sum">
          <span>应收</span>
          <span class="money">{{ yuan(estimate?.payAmount ?? 0) }}</span>
        </div>
        <div v-if="discount > 0" class="discount">已优惠 {{ yuan(discount) }}</div>
        <ElButton
          type="primary"
          size="large"
          class="submit"
          :loading="submitting"
          :disabled="!cart.length"
          @click="submit"
        >
          下单
        </ElButton>
      </div>
    </aside>

    <ElDialog v-model="pickVisible" :title="picking?.name || '选择规格'" width="480px">
      <div v-if="picking">
        <div class="block-label">规格</div>
        <ElRadioGroup v-model="pickSkuId">
          <ElRadioButton
            v-for="sku in picking.skus"
            :key="sku.id"
            :value="sku.id"
            :disabled="sku.soldOut === 1"
          >
            {{ sku.name || '默认' }} ¥{{ yuan(sku.price) }}
          </ElRadioButton>
        </ElRadioGroup>

        <template v-for="group in picking.addonGroups" :key="group.id">
          <div class="block-label">
            {{ group.name }}
            <span class="hint">
              {{ group.required === 1 ? '必选' : '可选' }}，最多 {{ group.maxSelect || 1 }} 项
            </span>
          </div>
          <ElCheckboxGroup :model-value="pickAddons[group.id] || []" @change="toggleAddon(group, $event)">
            <ElCheckbox v-for="item in group.items" :key="item.id" :value="item.id">
              {{ item.name }}
              <span v-if="item.price > 0">+¥{{ yuan(item.price) }}</span>
            </ElCheckbox>
          </ElCheckboxGroup>
        </template>

        <div class="block-label">备注</div>
        <ElInput v-model="pickRemark" placeholder="如：少辣、不要香菜" />
      </div>

      <template #footer>
        <ElButton @click="pickVisible = false">取消</ElButton>
        <ElButton type="primary" @click="confirmPick">加入</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="rechargeVisible" title="会员充值" width="380px">
      <div class="recharge-member">
        {{ member?.nickname || member?.mobile }} · 当前余额 ¥{{ yuan(walletBalance) }}
      </div>
      <ElForm label-width="70px">
        <ElFormItem label="实收">
          <ElInputNumber v-model="rechargeYuan" :min="0.01" :precision="2" :step="10" class="w-full" />
        </ElFormItem>
        <ElFormItem label="赠送">
          <ElInputNumber v-model="bonusYuan" :min="0" :precision="2" :step="0" class="w-full" />
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput v-model="rechargeRemark" placeholder="选填" />
        </ElFormItem>
      </ElForm>
      <div class="recharge-tip">现金充值不经支付渠道，提交后余额立即可用；赠送不能高于实收</div>

      <template #footer>
        <ElButton :disabled="recharging" @click="rechargeVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="recharging" @click="submitRecharge">确认充值</ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="heldVisible" title="挂起的单" width="440px">
      <ElEmpty v-if="!held.length" description="没有挂起的单" :image-size="56" />
      <div v-for="item in held" :key="item.id" class="held-item">
        <div>
          <div class="held-title">
            {{ item.tableName || (item.orderType === 1 ? '堂食' : '自提') }}
            <span class="held-count">{{ item.lines.length }} 项</span>
          </div>
          <div class="held-sub">
            {{ item.createdAt }} · ¥{{ yuan(item.amount) }}
            <span v-if="item.memberName"> · {{ item.memberName }}</span>
          </div>
        </div>
        <div class="held-actions">
          <ElButton link type="primary" size="small" @click="restoreHeld(item)">取单</ElButton>
          <ElButton link type="danger" size="small" @click="removeHeld(item)">删除</ElButton>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="memberVisible" title="选择会员" width="420px">
      <ElInput
        v-model="memberKeyword"
        placeholder="手机号或昵称"
        clearable
        @keyup.enter="searchMembers"
        @clear="searchMembers"
      >
        <template #append>
          <ElButton @click="searchMembers">搜索</ElButton>
        </template>
      </ElInput>

      <div v-loading="memberLoading" class="member-list">
        <ElEmpty
          v-if="!memberLoading && !memberList.length"
          description="没有匹配的会员"
          :image-size="56"
        />
        <div v-for="item in memberList" :key="item.id" class="member-item">
          <div>
            <div class="member-item-name">{{ item.nickname || item.mobile }}</div>
            <div class="member-item-sub">{{ item.mobile }} · 积分 {{ item.point }}</div>
          </div>
          <ElButton link type="primary" size="small" @click="selectMember(item)">选择</ElButton>
        </div>
      </div>
    </ElDialog>

    <ElDialog v-model="doneVisible" title="下单成功" width="360px" :close-on-click-modal="false">
      <div class="done">
        <div class="done-no">{{ doneOrder?.no }}</div>
        <div class="money">¥{{ yuan(doneOrder?.payAmount ?? 0) }}</div>
        <div v-if="doneOrder?.pickupCode" class="done-tip">取餐码 {{ doneOrder.pickupCode }}</div>
      </div>

      <template #footer>
        <ElButton :disabled="cashLoading" @click="finish(false)">稍后收款</ElButton>
        <ElButton type="primary" :loading="cashLoading" @click="finish(true)">现金收款</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRoute } from 'vue-router'

  import { fetchPosMenu } from '@/api/dish'
  import { fetchMemberPage } from '@/api/member'
  import {
    fetchAppendOrderItems,
    fetchPosCreateOrder,
    fetchPosEstimate,
    fetchSettleCash
  } from '@/api/order'
  import { fetchOpenTable, fetchTablePage } from '@/api/store'
  import { fetchCashRecharge, fetchMemberWallet } from '@/api/wallet'
  import { STORE_KEY } from '@/utils/http'

  const route = useRoute()

  /** 购物车一行：同一规格 + 同一加料组合 + 同一备注才合并 */
  interface CartLine {
    key: string
    skuId: string
    spuName: string
    skuName: string
    /** 菜单展示价（分），仅用于购物车小计展示，最终以试算为准 */
    unitPrice: number
    quantity: number
    addons: { addonItemId: string; name: string; price: number; quantity: number }[]
    addonText: string
    addonAmount: number
    remark: string
  }

  /** 挂单记录：只存在浏览器本地，保留取单与展示所需字段 */
  interface HeldOrder {
    id: string
    createdAt: string
    orderType: number
    tableId: string
    tableName: string
    memberId: string
    memberName: string
    memberPoint: number
    remark: string
    usePoints: boolean
    /** 挂起时的应收（分），仅用于列表展示 */
    amount: number
    lines: CartLine[]
  }

  const menuLoading = ref(false)
  const categories = ref<Api.Dish.MenuCategory[]>([])
  const activeCategory = ref('')
  const orderType = ref(1)
  const tableId = ref('')
  const tables = ref<Api.Store.TableItem[]>([])
  const remark = ref('')

  const cart = ref<CartLine[]>([])
  const estimate = ref<Api.Order.EstimateResult | null>(null)
  const submitting = ref(false)

  const pickVisible = ref(false)
  const picking = ref<Api.Dish.MenuSpu | null>(null)
  const pickSkuId = ref('')
  const pickAddons = reactive<Record<string, string[]>>({})
  const pickRemark = ref('')

  const doneVisible = ref(false)
  const doneOrder = ref<Api.Order.CreateResult['order'] | null>(null)
  const cashLoading = ref(false)

  // 加菜模式：从桌台页带 orderId 进来，提交时并入原单而不是新建订单
  const appendOrderId = ref('')
  const appendOrderNo = ref('')

  // 会员识别：选中后菜单按会员价重取，试算与下单都带 memberId
  const member = ref<Api.Member.Item | null>(null)
  const memberVisible = ref(false)
  const memberKeyword = ref('')
  const memberList = ref<Api.Member.Item[]>([])
  const memberLoading = ref(false)
  const usePoints = ref(false)

  // 会员钱包余额与充值：充值走现金，不生成支付单
  const walletBalance = ref(0)
  const rechargeVisible = ref(false)
  const recharging = ref(false)
  const rechargeYuan = ref(100)
  const bonusYuan = ref(0)
  const rechargeRemark = ref('')

  // 挂单：整单暂存到 localStorage，不落库。收银中途走开、临时换一单都不丢购物车
  const HELD_KEY = 'merchant_pos_held'
  const heldVisible = ref(false)
  const held = ref<HeldOrder[]>(readHeld())

  // 挂单必须按门店隔离：不同门店的菜品与价格完全不同，混在一起取单会算错钱
  const heldKey = (): string => `${HELD_KEY}_${localStorage.getItem(STORE_KEY) || '0'}`

  function readHeld(): HeldOrder[] {
    try {
      const raw = localStorage.getItem(heldKey())

      return raw ? (JSON.parse(raw) as HeldOrder[]) : []
    } catch {
      return []
    }
  }

  const persistHeld = (): void => {
    localStorage.setItem(heldKey(), JSON.stringify(held.value))
  }

  // 幂等请求号：同一购物车重复点「下单」不会生成两张单，成功后重置
  const requestNo = ref('')
  let estimateTimer: ReturnType<typeof setTimeout> | null = null

  const yuan = (cent: number): string => ((cent || 0) / 100).toFixed(2)

  // 当前分类下的菜品：分类带子级时把子级菜品一并展开
  const spus = computed<Api.Dish.MenuSpu[]>(() => {
    const current = categories.value.find((item) => item.id === activeCategory.value)
    if (!current) return []

    const collect = (cat: Api.Dish.MenuCategory): Api.Dish.MenuSpu[] => [
      ...cat.items,
      ...(cat.children ?? []).flatMap(collect)
    ]

    return collect(current)
  })

  const discount = computed(() => {
    const value = estimate.value
    if (!value) return 0

    return Math.max(0, value.itemAmount + value.addonAmount - value.payAmount)
  })

  const isSoldOut = (spu: Api.Dish.MenuSpu): boolean =>
    spu.skus.length > 0 && spu.skus.every((sku) => sku.soldOut === 1)

  /** 菜品卡片上的价格：取最低价并标注「起」 */
  const priceText = (spu: Api.Dish.MenuSpu): string => {
    const prices = spu.skus.filter((sku) => sku.soldOut !== 1).map((sku) => sku.price)
    if (!prices.length) return '—'
    const min = Math.min(...prices)

    return spu.skus.length > 1 ? `¥${yuan(min)} 起` : `¥${yuan(min)}`
  }

  const linePrice = (line: CartLine): number => line.unitPrice * line.quantity + line.addonAmount

  const loadMenu = async (): Promise<void> => {
    menuLoading.value = true
    try {
      const snapshot = await fetchPosMenu(member.value?.id)
      categories.value = (snapshot.categories ?? []).filter(
        (cat) => cat.items.length > 0 || (cat.children ?? []).length > 0
      )
      activeCategory.value = categories.value[0]?.id || ''
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      menuLoading.value = false
    }
  }

  // 堂食必须选台，这里只取空闲桌台，避免把单挂到别人正在用的台上
  const loadTables = async (): Promise<void> => {
    try {
      const res = await fetchTablePage({ status: 1, pageNo: 1, pageSize: 200 })
      tables.value = res.list ?? []
    } catch {
      tables.value = []
    }
  }

  // 打开会员检索：默认列出前若干会员，收银员可直接点选
  const openMember = (): void => {
    memberVisible.value = true
    memberKeyword.value = ''
    void searchMembers()
  }

  // 检索会员：纯数字按手机号查，否则按昵称查。
  // 两个条件同时传是 AND 关系，会互相抵消，所以只传一个。
  const searchMembers = async (): Promise<void> => {
    memberLoading.value = true
    try {
      const params: Record<string, unknown> = { pageNo: 1, pageSize: 20 }
      const keyword = memberKeyword.value.trim()
      if (keyword) {
        if (/^\d+$/.test(keyword)) {
          params.mobile = keyword
        } else {
          params.nickname = keyword
        }
      }
      const res = await fetchMemberPage(params)
      memberList.value = res.list ?? []
    } catch {
      memberList.value = []
    } finally {
      memberLoading.value = false
    }
  }

  // 切换会员前先确认：会员价与散客价不同口径，购物车必须清空重来
  const confirmResetCart = async (): Promise<boolean> => {
    if (!cart.value.length) return true

    try {
      await ElMessageBox.confirm('切换会员后价格口径会变化，当前购物车将被清空，是否继续？', '提示', {
        confirmButtonText: '继续',
        cancelButtonText: '取消',
        type: 'warning'
      })

      return true
    } catch {
      return false
    }
  }

  // 读取会员钱包余额；没有钱包时后端返回 0，不建钱包
  const loadWallet = async (): Promise<void> => {
    if (!member.value) {
      walletBalance.value = 0

      return
    }

    try {
      const res = await fetchMemberWallet(member.value.id)
      walletBalance.value = res.balance ?? 0
    } catch {
      walletBalance.value = 0
    }
  }

  // 选中会员：菜单按会员价重取，购物车清空避免旧价残留
  const selectMember = async (item: Api.Member.Item): Promise<void> => {
    if (!(await confirmResetCart())) return

    memberVisible.value = false
    member.value = item
    cart.value = []
    estimate.value = null
    await Promise.all([loadMenu(), loadWallet()])
  }

  const clearMember = async (): Promise<void> => {
    if (!(await confirmResetCart())) return

    member.value = null
    usePoints.value = false
    walletBalance.value = 0
    cart.value = []
    estimate.value = null
    await loadMenu()
  }

  const openRecharge = (): void => {
    rechargeYuan.value = 100
    bonusYuan.value = 0
    rechargeRemark.value = ''
    rechargeVisible.value = true
  }

  // 现金充值：金额在提交前转成分，前端只做基本校验，上限与赠送规则后端也会再拦一次
  const submitRecharge = async (): Promise<void> => {
    const current = member.value
    if (!current) return

    const price = Math.round((rechargeYuan.value || 0) * 100)
    const bonus = Math.round((bonusYuan.value || 0) * 100)
    if (price <= 0) {
      ElMessage.warning('请输入实收金额')

      return
    }
    if (bonus > price) {
      ElMessage.warning('赠送金额不能高于实收金额')

      return
    }

    recharging.value = true
    try {
      const res = await fetchCashRecharge(current.id, {
        price,
        bonusPrice: bonus,
        remark: rechargeRemark.value.trim()
      })
      walletBalance.value = res.balance
      rechargeVisible.value = false
      ElMessage.success(`充值成功，当前余额 ¥${yuan(res.balance)}`)
    } catch {
      // 错误提示由请求层统一弹出
    } finally {
      recharging.value = false
    }
  }

  const pick = (spu: Api.Dish.MenuSpu): void => {
    if (isSoldOut(spu)) {
      ElMessage.warning(`${spu.name} 已沽清`)

      return
    }

    picking.value = spu
    const available = spu.skus.find((sku) => sku.soldOut !== 1)
    pickSkuId.value = available?.id || spu.skus[0]?.id || ''
    pickRemark.value = ''
    Object.keys(pickAddons).forEach((key) => delete pickAddons[key])
    for (const group of spu.addonGroups) {
      pickAddons[group.id] = []
    }
    pickVisible.value = true
  }

  // 加料多选上限：超出时提示并丢弃多余选项
  const toggleAddon = (group: Api.Dish.MenuAddonGroup, values: (string | number | boolean)[]): void => {
    const picked = values.map(String)
    const limit = group.maxSelect > 0 ? group.maxSelect : group.items.length
    if (picked.length > limit) {
      ElMessage.warning(`${group.name} 最多选 ${limit} 项`)
      pickAddons[group.id] = picked.slice(0, limit)

      return
    }
    pickAddons[group.id] = picked
  }

  const confirmPick = (): void => {
    const spu = picking.value
    if (!spu) return

    const sku = spu.skus.find((item) => item.id === pickSkuId.value)
    if (!sku) {
      ElMessage.warning('请选择规格')

      return
    }

    // 必选组必须选够 minSelect
    const addons: CartLine['addons'] = []
    for (const group of spu.addonGroups) {
      const ids = pickAddons[group.id] || []
      if (group.required === 1 && ids.length < Math.max(1, group.minSelect)) {
        ElMessage.warning(`请选择${group.name}`)

        return
      }
      for (const id of ids) {
        const item = group.items.find((addon) => addon.id === id)
        if (item) {
          addons.push({ addonItemId: item.id, name: item.name, price: item.price, quantity: 1 })
        }
      }
    }

    const addonAmount = addons.reduce((sum, addon) => sum + addon.price * addon.quantity, 0)
    const key = [sku.id, ...addons.map((addon) => addon.addonItemId).sort(), pickRemark.value].join('|')
    const exist = cart.value.find((line) => line.key === key)
    if (exist) {
      exist.quantity += 1
    } else {
      cart.value.push({
        key,
        skuId: sku.id,
        spuName: spu.name,
        skuName: spu.skus.length > 1 ? sku.name : '',
        unitPrice: sku.price,
        quantity: 1,
        addons,
        addonText: addons.map((addon) => addon.name).join('、'),
        addonAmount,
        remark: pickRemark.value.trim()
      })
    }

    pickVisible.value = false
    scheduleEstimate()
  }

  const removeLine = (line: CartLine): void => {
    cart.value = cart.value.filter((item) => item.key !== line.key)
    scheduleEstimate()
  }

  // 挂起当前单：连会员与台位一起存，取单后可直接结算
  const holdCurrent = (): void => {
    if (!cart.value.length) return

    const table = tables.value.find((item) => item.id === tableId.value)
    held.value.unshift({
      id: newRequestNo(),
      createdAt: new Date().toLocaleString('zh-CN', { hour12: false }),
      orderType: orderType.value,
      tableId: tableId.value,
      tableName: orderType.value === 1 ? table?.name || '' : '',
      memberId: member.value?.id || '',
      memberName: member.value?.nickname || member.value?.mobile || '',
      memberPoint: member.value?.point || 0,
      remark: remark.value,
      usePoints: usePoints.value,
      amount: estimate.value?.payAmount ?? 0,
      lines: cart.value.map((line) => ({ ...line }))
    })
    persistHeld()

    // 挂起后清空工作区，避免与下一单混淆
    cart.value = []
    estimate.value = null
    remark.value = ''
    tableId.value = ''
    requestNo.value = ''
    member.value = null
    usePoints.value = false
    void loadMenu()
    ElMessage.success('已挂单')
  }

  // 取单：恢复购物车、台位与会员；价格口径可能已变，恢复后重新试算
  const restoreHeld = async (item: HeldOrder): Promise<void> => {
    if (cart.value.length) {
      try {
        await ElMessageBox.confirm('当前购物车还有菜品，取单会覆盖，是否继续？', '取单', {
          confirmButtonText: '继续',
          cancelButtonText: '取消',
          type: 'warning'
        })
      } catch {
        return
      }
    }

    cart.value = item.lines.map((line) => ({ ...line }))
    orderType.value = item.orderType
    tableId.value = item.tableId
    remark.value = item.remark
    usePoints.value = item.usePoints
    // 挂单里只存了展示需要的字段，其余按默认值补齐，够用即可
    member.value = item.memberId
      ? {
          id: item.memberId,
          mobile: '',
          nickname: item.memberName,
          avatar: '',
          name: '',
          levelId: '0',
          point: item.memberPoint,
          experience: 0,
          status: 1
        }
      : null

    held.value = held.value.filter((record) => record.id !== item.id)
    persistHeld()
    heldVisible.value = false
    requestNo.value = ''
    await loadMenu()
    scheduleEstimate()
  }

  const removeHeld = async (item: HeldOrder): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定删除挂单「${item.tableName || (item.orderType === 1 ? '堂食' : '自提')}」吗？删除后购物车内容不再保留。`,
        '删除挂单',
        { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    held.value = held.value.filter((record) => record.id !== item.id)
    persistHeld()
  }

  // 试算防抖：连续加点只发最后一次，避免刷接口
  const scheduleEstimate = (): void => {
    if (estimateTimer) clearTimeout(estimateTimer)
    estimateTimer = setTimeout(() => void recalc(), 300)
  }

  // 购物车转接口明细：试算、下单、加菜三处共用，避免字段漏写
  const buildItems = () =>
    cart.value.map((line) => ({
      skuId: line.skuId,
      quantity: line.quantity,
      addons: line.addons.map((addon) => ({
        addonItemId: addon.addonItemId,
        quantity: addon.quantity
      })),
      remark: line.remark
    }))

  const recalc = async (): Promise<void> => {
    if (!cart.value.length) {
      estimate.value = null

      return
    }

    try {
      estimate.value = await fetchPosEstimate({
        memberId: member.value?.id || '',
        usePoints: usePoints.value && !!member.value,
        items: buildItems()
      })
    } catch {
      estimate.value = null
    }
  }

  const newRequestNo = (): string =>
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `POS${Date.now()}`

  const submit = async (): Promise<void> => {
    if (!cart.value.length) return

    // 加菜模式：明细并入原单，不新建订单，也不碰桌台状态
    if (appendOrderId.value) {
      submitting.value = true
      try {
        const res = await fetchAppendOrderItems(appendOrderId.value, buildItems())
        doneOrder.value = {
          id: res.id,
          no: res.no,
          status: 0,
          statusText: '',
          payAmount: res.payAmount,
          pickupCode: '',
          expireTime: ''
        }
        doneVisible.value = true
        ElMessage.success(`已加菜，整单应收 ¥${yuan(res.payAmount)}`)
      } catch {
        // 错误提示由请求层统一弹出，购物车保留便于重试
      } finally {
        submitting.value = false
      }

      return
    }

    if (orderType.value === 1 && !tableId.value) {
      ElMessage.warning('堂食请先选择桌台')

      return
    }

    submitting.value = true
    try {
      const res = await fetchPosCreateOrder({
        type: orderType.value,
        tableId: orderType.value === 1 ? tableId.value : '',
        memberId: member.value?.id || '',
        usePoints: usePoints.value && !!member.value,
        requestNo: requestNo.value || (requestNo.value = newRequestNo()),
        remark: remark.value.trim(),
        items: buildItems()
      })
      // 堂食下单成功后开台并绑定台位单：桌台图据此显示「有台位单」。
      // 台位已被占用时开台会失败，属正常情况（重复开台），忽略即可，不影响下单结果。
      if (orderType.value === 1 && tableId.value) {
        try {
          await fetchOpenTable(tableId.value, res.order.id)
        } catch {
          // 静默：下单已成功，开台失败只影响桌台图展示
        }
      }

      doneOrder.value = res.order
      doneVisible.value = true
      if (res.idempotentHit) {
        ElMessage.warning('该请求已生成过订单，沿用原订单')
      }
    } catch {
      // 错误提示由请求层统一弹出，保留购物车与请求号便于重试
    } finally {
      submitting.value = false
    }
  }

  // 收尾：可选现金收款，随后清空购物车并换新的请求号
  const finish = async (settle: boolean): Promise<void> => {
    const order = doneOrder.value
    if (settle && order) {
      cashLoading.value = true
      try {
        await fetchSettleCash(order.id)
        ElMessage.success('现金收款成功')
      } catch {
        // 收款失败保留弹窗，收银员可重试或稍后在订单看板处理
        cashLoading.value = false

        return
      }
      cashLoading.value = false
    }

    doneVisible.value = false
    doneOrder.value = null
    cart.value = []
    estimate.value = null
    remark.value = ''
    tableId.value = ''
    requestNo.value = ''
    // 加菜完成后退回新单模式，避免下一次点菜又并到旧单上
    appendOrderId.value = ''
    appendOrderNo.value = ''
    // 换下一单：会员与积分抵扣一并清掉，避免把上一位顾客的身份带过去
    if (member.value) {
      member.value = null
      usePoints.value = false
      await loadMenu()
    }
    void loadTables()
  }

  // 退出加菜模式：清空当前购物车与订单上下文，回到新单
  const exitAppend = (): void => {
    appendOrderId.value = ''
    appendOrderNo.value = ''
    cart.value = []
    estimate.value = null
    void loadMenu()
  }

  onMounted(async () => {
    // 从桌台图进来时带上 tableId，直接按堂食预选台位
    const presetTable = route.query.tableId
    if (typeof presetTable === 'string' && presetTable) {
      orderType.value = 1
      tableId.value = presetTable
    }
    // 带 orderId 表示加菜：明细会并入这张单
    const presetOrder = route.query.orderId
    if (typeof presetOrder === 'string' && presetOrder) {
      appendOrderId.value = presetOrder
      appendOrderNo.value = typeof route.query.orderNo === 'string' ? route.query.orderNo : ''
      orderType.value = 1
    }

    await loadMenu()
    await loadTables()
  })
</script>

<style scoped>
  .pos {
    display: flex;
    height: calc(100vh - 61px);
  }

  .cats {
    width: 132px;
    overflow: auto;
    background: #fff;
    border-right: 1px solid #e5e7eb;
  }

  .cat {
    padding: 14px 12px;
    font-size: 14px;
    cursor: pointer;
    border-bottom: 1px solid #f2f3f5;
  }

  .cat.active {
    font-weight: 600;
    color: #409eff;
    background: #ecf5ff;
  }

  .dishes {
    flex: 1;
    padding: 12px;
    overflow: auto;
  }

  .dish-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 10px;
  }

  .dish {
    position: relative;
    padding: 14px 12px;
    cursor: pointer;
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
  }

  .dish:hover {
    border-color: #409eff;
  }

  .dish.off {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .dish-name {
    font-size: 14px;
    font-weight: 500;
  }

  .dish-price {
    margin-top: 8px;
    font-size: 15px;
    font-weight: 600;
    color: #f56c6c;
  }

  .dish-off {
    position: absolute;
    top: 8px;
    right: 8px;
    font-size: 12px;
    color: #909399;
  }

  .cart {
    display: flex;
    flex-direction: column;
    width: 340px;
    background: #fff;
    border-left: 1px solid #e5e7eb;
  }

  .cart-head {
    padding: 12px;
    border-bottom: 1px solid #f2f3f5;
  }

  .append-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 8px;
    margin-bottom: 8px;
    font-size: 13px;
    color: #b45309;
    background: #fdf6ec;
    border-radius: 6px;
  }

  .member-bar {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 8px;
    font-size: 13px;
  }

  .member-name {
    font-weight: 600;
  }

  .member-point {
    color: #909399;
  }

  .member-balance {
    color: #f56c6c;
  }

  .recharge-member {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 500;
    text-align: center;
  }

  .recharge-tip {
    font-size: 12px;
    color: #909399;
  }

  .w-full {
    width: 100%;
  }

  .points-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 8px;
  }

  .hold-row {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }

  .held-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0;
    border-bottom: 1px dashed #ebeef5;
  }

  .held-title {
    font-size: 14px;
    font-weight: 500;
  }

  .held-count {
    margin-left: 6px;
    font-size: 12px;
    font-weight: 400;
    color: #909399;
  }

  .held-sub {
    margin-top: 2px;
    font-size: 12px;
    color: #6b7280;
  }

  .held-actions {
    display: flex;
    gap: 4px;
  }

  .member-list {
    max-height: 320px;
    margin-top: 12px;
    overflow: auto;
  }

  .member-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0;
    border-bottom: 1px dashed #ebeef5;
  }

  .member-item-name {
    font-size: 14px;
    font-weight: 500;
  }

  .member-item-sub {
    margin-top: 2px;
    font-size: 12px;
    color: #6b7280;
  }

  .cart-lines {
    flex: 1;
    padding: 8px 12px;
    overflow: auto;
  }

  .line {
    display: flex;
    gap: 8px;
    justify-content: space-between;
    padding: 10px 0;
    border-bottom: 1px dashed #ebeef5;
  }

  .line-name {
    font-size: 14px;
    font-weight: 500;
  }

  .line-addon,
  .line-remark {
    margin-top: 2px;
    font-size: 12px;
    color: #6b7280;
  }

  .line-side {
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-end;
    width: 110px;
  }

  .line-sub {
    font-size: 13px;
    font-weight: 600;
  }

  .cart-foot {
    padding: 12px;
    border-top: 1px solid #f2f3f5;
  }

  .sum {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-top: 10px;
    font-size: 14px;
  }

  .money {
    font-size: 20px;
    font-weight: 700;
    color: #f56c6c;
  }

  .discount {
    margin-top: 2px;
    font-size: 12px;
    color: #67c23a;
    text-align: right;
  }

  .submit {
    width: 100%;
    margin-top: 10px;
  }

  .block-label {
    margin: 14px 0 8px;
    font-size: 13px;
    font-weight: 600;
    color: #303133;
  }

  .hint {
    margin-left: 6px;
    font-size: 12px;
    font-weight: 400;
    color: #909399;
  }

  .done {
    text-align: center;
  }

  .done-no {
    font-size: 16px;
    font-weight: 600;
  }

  .done-tip {
    margin-top: 6px;
    font-size: 12px;
    color: #6b7280;
  }

  .mt-2 {
    margin-top: 8px;
  }
</style>
