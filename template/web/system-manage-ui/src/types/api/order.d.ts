declare namespace Api {
  namespace Order {
    /** 订单类型 */
    type OrderType = 1 | 2 | 3
    /** 订单状态，取值与字典 cater_order_status 一致 */
    type OrderStatus = 100 | 200 | 300 | 400 | 500 | 600 | 700 | 800
    /** 支付状态：0 未支付 / 10 已支付 / 20 已退款 */
    type PayStatus = 0 | 10 | 20
    /** 下单渠道：1 小程序 / 2 POS / 3 自助机 / 4 外卖平台 */
    type OrderChannel = 1 | 2 | 3 | 4

    /** 订单明细上的加料快照 */
    interface OrderAddonItem {
      id: string
      orderItemId: string
      addonItemId: string
      addonName: string
      /** 加价（分） */
      price: number
      quantity: number
    }

    /** 订单明细 */
    interface OrderItemRow {
      id: string
      spuId: string
      skuId: string
      spuName: string
      skuName: string
      /** 下单时原价（分） */
      originalPrice: number
      /** 成交单价（分） */
      unitPrice: number
      quantity: number
      /** 加料费合计（分） */
      addonAmount: number
      /** 分摊到本行的整单优惠（分） */
      itemDiscountAmount: number
      /** 小计（分） */
      subtotal: number
      /** 明细状态：1 正常 / 2 已退 */
      status: number
      orderer: string
      remark: string
      addons: OrderAddonItem[]
    }

    /** 订单状态日志 */
    interface OrderLogRow {
      id: string
      fromStatus: number
      toStatus: number
      /** 操作方：1 顾客 / 2 店员 / 3 系统 */
      operatorType: number
      operatorName: string
      reason: string
      createdAt: string
    }

    /** 订单列表行 */
    interface OrderItem {
      id: string
      no: string
      storeId: string
      memberId: string
      /** 订单类型 */
      type: OrderType
      typeText: string
      /** 下单渠道 */
      channel: OrderChannel
      channelText: string
      tableId: string
      tableName: string
      /** 订单状态 */
      status: OrderStatus
      statusText: string
      /** 支付状态 */
      payStatus: PayStatus
      payStatusText: string
      /** 商品小计（分） */
      itemAmount: number
      /** 加料费合计（分） */
      addonAmount: number
      /** 单品级优惠（分） */
      lineDiscountAmount: number
      /** 活动优惠（分） */
      activityAmount: number
      /** 优惠券抵扣（分） */
      couponAmount: number
      /** 积分抵扣（分） */
      pointsAmount: number
      /** 应付金额（分） */
      payAmount: number
      /** 已退金额（分） */
      refundAmount: number
      /** 优惠计算过程快照 JSON */
      discountSnapshot: string
      /** 取餐码 */
      pickupCode: string
      receiverName: string
      receiverPhone: string
      receiverAddress: string
      remark: string
      /** 商家备注，与顾客备注分列保存 */
      merchantRemark: string
      /** 改价差额（分）：正数加价、负数减免 */
      adjustAmount: number
      /** 配送骑手 ID，0 表示未派单 */
      riderId: string
      /** 派单时间 */
      dispatchTime: string | null
      payTime: string | null
      acceptTime: string | null
      completeTime: string | null
      cancelTime: string | null
      cancelReason: string
      /** 待支付超时时间 */
      expireTime: string
      createdAt: string
    }

    /** 订单详情 */
    interface OrderDetail extends OrderItem {
      items: OrderItemRow[]
      logs: OrderLogRow[]
    }

    /** 订单分页结果 */
    interface OrderList {
      list: OrderItem[]
      total: number
    }

    /** 订单列表查询参数 */
    interface OrderSearchParams {
      /** 订单号 / 取餐码 / 收货电话 */
      keyword?: string
      storeId?: string
      /** 订单状态 */
      status?: number
      /** 支付状态 */
      payStatus?: number
      /** 订单类型 */
      type?: number
      pageNo?: number
      pageSize?: number
    }

    /** 状态流转入参 */
    interface OrderTransitParams {
      /** 目标状态 */
      to: number
      /** 变更原因 */
      reason?: string
    }

    /** 修改订单备注入参 */
    interface OrderRemarkParams {
      /** 顾客备注 */
      remark?: string
      /** 商家备注 */
      merchantRemark?: string
    }

    /** 订单改价入参 */
    interface OrderAdjustPriceParams {
      /** 改价差额（分）：正数加价、负数减免 */
      delta: number
    }

    /** 订单派单入参 */
    interface OrderAssignRiderParams {
      /** 骑手 ID */
      riderId: string
    }

    /** 可派单骑手下拉项 */
    interface OrderRiderOption {
      riderId: string
      name: string
      /** 当前进行中配送单量 */
      currentOrders: number
      /** 接单并发上限 */
      maxOrders: number
      /** 是否还有接单额度 */
      available: boolean
    }

    /** 小票上的加料 */
    interface OrderTicketAddon {
      name: string
      quantity: number
    }

    /** 小票上的菜品 */
    interface OrderTicketItem {
      name: string
      skuName: string
      quantity: number
      /** 成交单价（分） */
      unitPrice: number
      /** 小计（分） */
      subtotal: number
      remark: string
      addons: OrderTicketAddon[]
    }

    /** 打印小票数据 */
    interface OrderTicket {
      storeName: string
      storeShortName: string
      orderNo: string
      pickupCode: string
      tableName: string
      typeText: string
      channelText: string
      receiverName: string
      receiverPhone: string
      receiverAddress: string
      remark: string
      merchantRemark: string
      items: OrderTicketItem[]
      /** 商品金额（分） */
      itemAmount: number
      /** 加料费（分） */
      addonAmount: number
      /** 优惠合计（分） */
      discountAmount: number
      /** 改价差额（分） */
      adjustAmount: number
      /** 应付金额（分） */
      payAmount: number
      payStatusText: string
      createdAt: string
      payTime: string
    }
  }
}
