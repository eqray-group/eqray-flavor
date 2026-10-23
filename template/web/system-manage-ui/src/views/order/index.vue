<!-- 订单管理页面：查询订单、查看明细与加料、执行状态流转与取消 -->
<template>
  <div class="order-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <div class="mb-3 flex items-center justify-end gap-2">
        <ElButton v-if="hasAuth('order:order:list')" :loading="exporting" @click="exportOrders">
          导出 CSV
        </ElButton>
      </div>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <OrderDetailDialog v-model:visible="detailVisible" :order-id="currentOrderId" />
      <OrderRemarkDialog
        v-model:visible="remarkVisible"
        :order="currentRow"
        @success="refreshData"
      />
      <OrderAdjustDialog
        v-model:visible="adjustVisible"
        :order="currentRow"
        @success="refreshData"
      />
      <OrderDispatchDialog
        v-model:visible="dispatchVisible"
        :order="currentRow"
        @success="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchCancelOrder,
    fetchCancelOrderRider,
    fetchExportOrders,
    fetchGetOrderList,
    fetchGetOrderTicket,
    fetchTransitOrder
  } from '@/api/order/order'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { exportFilename, saveBlob } from '@/utils/download'
  import OrderAdjustDialog from './modules/order-adjust-dialog.vue'
  import OrderDetailDialog from './modules/order-detail-dialog.vue'
  import OrderDispatchDialog from './modules/order-dispatch-dialog.vue'
  import OrderRemarkDialog from './modules/order-remark-dialog.vue'

  defineOptions({ name: 'OrderManage' })

  type OrderItem = Api.Order.OrderItem

  const { hasAuth } = useAuth()

  // 订单状态文案与后端 order/enums 一致
  const STATUS_TEXT: Record<number, string> = {
    100: '待支付',
    200: '待接单',
    300: '制作中',
    400: '待取餐',
    500: '已完成',
    600: '已取消',
    700: '退款中',
    800: '已退款'
  }

  const statusOptions = Object.entries(STATUS_TEXT).map(([value, label]) => ({
    label,
    value: Number(value)
  }))

  const typeOptions = [
    { label: '堂食', value: 1 },
    { label: '外卖', value: 2 },
    { label: '自提', value: 3 }
  ]

  const payStatusOptions = [
    { label: '未支付', value: 0 },
    { label: '已支付', value: 10 },
    { label: '已退款', value: 20 }
  ]

  const statusTagTypes: Record<number, 'warning' | 'success' | 'info' | 'danger' | 'primary'> = {
    100: 'warning',
    200: 'warning',
    300: 'primary',
    400: 'primary',
    500: 'success',
    600: 'info',
    700: 'danger',
    800: 'info'
  }

  const searchForm = ref<Api.Order.OrderSearchParams>({
    keyword: '',
    status: undefined,
    payStatus: undefined,
    type: undefined
  })

  const searchItems = computed(() => [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '订单号 / 取餐码 / 收货电话',
        clearable: true
      }
    },
    {
      label: '订单状态',
      key: 'status',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    },
    {
      label: '支付状态',
      key: 'payStatus',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: payStatusOptions
      }
    },
    {
      label: '订单类型',
      key: 'type',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: typeOptions
      }
    }
  ])

  const detailVisible = ref(false)
  const currentOrderId = ref('')

  // 经营操作弹窗共用同一个当前行
  const remarkVisible = ref(false)
  const adjustVisible = ref(false)
  const dispatchVisible = ref(false)
  const exporting = ref(false)
  const currentRow = ref<OrderItem | null>(null)

  // 是否已派单：骑手 ID 为 0 表示未派单
  const isDispatched = (row: OrderItem): boolean => !!row.riderId && String(row.riderId) !== '0'

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetOrderList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'no',
          label: '订单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'pickupCode',
          label: '取餐码',
          width: 90
        },
        {
          prop: 'typeText',
          label: '类型',
          width: 80
        },
        {
          prop: 'channelText',
          label: '渠道',
          width: 100
        },
        {
          prop: 'tableName',
          label: '桌台',
          width: 100
        },
        {
          prop: 'payAmount',
          label: '应付（元）',
          width: 110,
          formatter: (row: OrderItem) => formatYuan(row.payAmount)
        },
        {
          prop: 'status',
          label: '订单状态',
          width: 100,
          formatter: (row: OrderItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusText || '未知'
            )
        },
        {
          prop: 'payStatus',
          label: '支付状态',
          width: 100,
          formatter: (row: OrderItem) =>
            h(
              ElTag,
              {
                type: row.payStatus === 10 ? 'success' : row.payStatus === 20 ? 'info' : 'warning',
                effect: 'plain'
              },
              () => row.payStatusText || '未知'
            )
        },
        {
          prop: 'receiverName',
          label: '收货人',
          width: 110
        },
        { prop: 'createdAt', label: '下单时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 420,
          fixed: 'right',
          formatter: (row: OrderItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '详情',
                onClick: () => openDetail(row)
              }),
              // 接单：待接单 → 制作中
              ...(row.status === 200 && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '接单',
                      onClick: () => transitOrder(row, 300)
                    })
                  ]
                : []),
              // 出餐：制作中 → 待取餐
              ...(row.status === 300 && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '出餐',
                      onClick: () => transitOrder(row, 400)
                    })
                  ]
                : []),
              // 完成：待取餐 / 制作中 → 已完成
              ...([300, 400].includes(row.status) && hasAuth('order:order:status')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '完成',
                      onClick: () => transitOrder(row, 500)
                    })
                  ]
                : []),
              // 取消：仅待支付可取消，其余状态要走退款
              ...(row.status === 100 && hasAuth('order:order:cancel')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '取消',
                      onClick: () => cancelOrder(row)
                    })
                  ]
                : []),
              // 备注：未取消、未退款的订单都可改
              ...(![600, 800].includes(row.status) && hasAuth('order:order:remark')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '备注',
                      onClick: () => openRemark(row)
                    })
                  ]
                : []),
              // 改价：仅未支付且处于待支付 / 待接单
              ...(row.payStatus === 0 &&
              [100, 200].includes(row.status) &&
              hasAuth('order:order:adjust')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      text: '改价',
                      onClick: () => openAdjust(row)
                    })
                  ]
                : []),
              // 派单 / 取消派单：仅外卖订单，待接单与制作中可操作
              ...(row.type === 2 &&
              [200, 300].includes(row.status) &&
              hasAuth('order:order:dispatch')
                ? [
                    isDispatched(row)
                      ? h(ArtButtonTable, {
                          type: 'delete',
                          text: '取消派单',
                          onClick: () => cancelDispatch(row)
                        })
                      : h(ArtButtonTable, {
                          type: 'action',
                          text: '派单',
                          onClick: () => openDispatch(row)
                        })
                  ]
                : []),
              // 小票：只读动作，跟订单查询同权限
              ...(hasAuth('order:order:list')
                ? [
                    h(ArtButtonTable, {
                      type: 'view',
                      text: '小票',
                      onClick: () => printTicket(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后重新加载订单列表
  const handleSearch = (params: Api.Order.OrderSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开订单详情弹窗，按主键拉取明细与日志
  const openDetail = (row: OrderItem): void => {
    currentOrderId.value = row.id
    detailVisible.value = true
  }

  // 执行状态流转：带原因的流转（如拒单）弹窗收集原因，其余直接确认
  const transitOrder = async (row: OrderItem, to: number): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要将订单「${row.no}」流转到「${STATUS_TEXT[to]}」吗？`,
        '流转订单',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }

    try {
      await fetchTransitOrder(row.id, { to })
      ElMessage.success('已流转')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 打开备注弹窗
  const openRemark = (row: OrderItem): void => {
    currentRow.value = row
    remarkVisible.value = true
  }

  // 打开改价弹窗
  const openAdjust = (row: OrderItem): void => {
    currentRow.value = row
    adjustVisible.value = true
  }

  // 打开派单弹窗
  const openDispatch = (row: OrderItem): void => {
    currentRow.value = row
    dispatchVisible.value = true
  }

  // 取消派单：清空骑手并归还接单额度
  const cancelDispatch = async (row: OrderItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要取消订单「${row.no}」的派单吗？`, '取消派单', {
        confirmButtonText: '确定',
        cancelButtonText: '返回',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchCancelOrderRider(row.id)
      ElMessage.success('已取消派单')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 导出当前筛选条件下的订单，最多由后端截断到一万条
  const exportOrders = async (): Promise<void> => {
    exporting.value = true
    try {
      const blob = await fetchExportOrders({ ...searchParams })
      saveBlob(blob, exportFilename('orders'))
      ElMessage.success('导出成功')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      exporting.value = false
    }
  }

  // 转义小票里的文本，避免备注里的尖括号破坏结构
  const escapeHtml = (value: string): string =>
    value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')

  // 生成小票 HTML：宽 80mm，前端直接调浏览器打印
  const buildTicketHtml = (ticket: Api.Order.OrderTicket): string => {
    const lines = ticket.items
      .map((item) => {
        const addons = item.addons.length
          ? `<div class="addon">+ ${item.addons.map((a) => `${escapeHtml(a.name)}×${a.quantity}`).join('，')}</div>`
          : ''
        const remark = item.remark
          ? `<div class="addon">备注：${escapeHtml(item.remark)}</div>`
          : ''
        return `<div class="item">
          <div class="row"><span>${escapeHtml(item.name)}${item.skuName ? `（${escapeHtml(item.skuName)}）` : ''}</span><span>×${item.quantity}</span></div>
          ${addons}${remark}
        </div>`
      })
      .join('')

    const receiver = ticket.receiverAddress
      ? `<div>收货：${escapeHtml(ticket.receiverName)} ${escapeHtml(ticket.receiverPhone)}</div>
         <div>地址：${escapeHtml(ticket.receiverAddress)}</div>`
      : ''

    return `<!DOCTYPE html><html><head><meta charset="utf-8" />
      <title>${escapeHtml(ticket.orderNo)}</title>
      <style>
        * { box-sizing: border-box; }
        body { width: 80mm; margin: 0 auto; padding: 6mm 4mm; font: 12px/1.5 "Microsoft YaHei", sans-serif; color: #000; }
        h1 { margin: 0 0 4px; font-size: 16px; text-align: center; }
        .center { text-align: center; }
        .no { font-size: 14px; font-weight: 700; }
        .row { display: flex; justify-content: space-between; gap: 8px; }
        .item { margin: 4px 0; }
        .addon { padding-left: 8px; font-size: 11px; }
        .divider { border-top: 1px dashed #000; margin: 6px 0; }
        .total { font-size: 14px; font-weight: 700; }
        .tip { margin-top: 8px; text-align: center; font-size: 11px; }
      </style></head><body>
      <h1>${escapeHtml(ticket.storeShortName || ticket.storeName || '消费小票')}</h1>
      <div class="center no">${escapeHtml(ticket.orderNo)}</div>
      <div class="center">${escapeHtml(ticket.typeText)} · ${escapeHtml(ticket.channelText)}${ticket.tableName ? ` · ${escapeHtml(ticket.tableName)}` : ''}</div>
      ${ticket.pickupCode ? `<div class="center">取餐码：${escapeHtml(ticket.pickupCode)}</div>` : ''}
      <div class="divider"></div>
      ${receiver}
      <div class="divider"></div>
      ${lines}
      <div class="divider"></div>
      <div class="row"><span>商品金额</span><span>¥${((ticket.itemAmount || 0) / 100).toFixed(2)}</span></div>
      <div class="row"><span>加料费</span><span>¥${((ticket.addonAmount || 0) / 100).toFixed(2)}</span></div>
      <div class="row"><span>优惠合计</span><span>-¥${((ticket.discountAmount || 0) / 100).toFixed(2)}</span></div>
      ${ticket.adjustAmount ? `<div class="row"><span>改价差额</span><span>${ticket.adjustAmount > 0 ? '+' : '-'}¥${(Math.abs(ticket.adjustAmount) / 100).toFixed(2)}</span></div>` : ''}
      <div class="row total"><span>应付</span><span>¥${((ticket.payAmount || 0) / 100).toFixed(2)}</span></div>
      <div class="row"><span>支付状态</span><span>${escapeHtml(ticket.payStatusText)}</span></div>
      <div class="divider"></div>
      <div>下单时间：${escapeHtml(ticket.createdAt)}</div>
      ${ticket.payTime ? `<div>支付时间：${escapeHtml(ticket.payTime)}</div>` : ''}
      ${ticket.remark ? `<div>顾客备注：${escapeHtml(ticket.remark)}</div>` : ''}
      ${ticket.merchantRemark ? `<div>商家备注：${escapeHtml(ticket.merchantRemark)}</div>` : ''}
      <div class="tip">谢谢惠顾，欢迎再次光临</div>
      </body></html>`
  }

  // 打印小票：拉取数据后在新窗口渲染并唤起打印
  const printTicket = async (row: OrderItem): Promise<void> => {
    try {
      const ticket = await fetchGetOrderTicket(row.id)
      const win = window.open('', '_blank', 'width=460,height=680')
      if (!win) {
        ElMessage.warning('浏览器拦截了打印窗口，请允许后重试')
        return
      }
      win.document.write(buildTicketHtml(ticket))
      win.document.close()
      win.focus()
      win.print()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 取消订单：只允许待支付状态，原因可为空
  const cancelOrder = async (row: OrderItem): Promise<void> => {
    try {
      const { value } = await ElMessageBox.prompt(
        `确定要取消订单「${row.no}」吗？取消后不可恢复。`,
        '取消订单',
        {
          confirmButtonText: '确定取消',
          cancelButtonText: '返回',
          inputPlaceholder: '取消原因（选填）',
          inputValue: ''
        }
      )
      await fetchCancelOrder(row.id, value || undefined)
      ElMessage.success('已取消')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
