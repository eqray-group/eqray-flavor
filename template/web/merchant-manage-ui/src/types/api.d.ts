/**
 * 门店端用到的接口类型，字段与后端 JSON 一一对应。
 * 只声明本端实际使用的字段，避免把整个管理后台的类型搬过来。
 */

/** 登录与当前用户 */
declare namespace Api {
  namespace Auth {
    interface LoginParams {
      username: string
      password: string
      captchaId: string
      captchaCode: string
    }

    interface LoginResponse {
      accessToken: string
      refreshToken: string
      expiresIn: number
      refreshExpiresIn: number
    }

    interface Captcha {
      captchaId: string
      /** base64 图片 */
      image: string
    }

    interface UserInfo {
      userId: string
      userName: string
      nickname?: string
      buttons: string[]
      roles: string[]
    }
  }

  /** 门店 */
  namespace Store {
    interface VisibleItem {
      id: string
      name: string
      shortName?: string
    }

    /** 桌台：1 空闲 / 2 使用中 / 3 待清台 */
    interface TableItem {
      id: string
      name: string
      areaId: string
      areaName: string
      seats: number
      status: number
      currentOrderId: string
      /** 台位单号，没有台位单时为空 */
      currentOrderNo?: string
      /** 台位单应收（分） */
      currentOrderAmount?: number
      /** 已开台分钟数 */
      openedMinutes?: number
    }

    interface TablePage {
      list: TableItem[]
      total: number
    }

    interface PrintTaskItem {
      id: string
      storeName: string
      orderNo: string
      printerName: string
      printTypeName: string
      /** 0 待打印 / 1 已打印 / 2 失败 / 3 已取消 */
      status: number
      statusName: string
      retryCount: number
      errorMsg: string
      content: string
      createdAt: string
    }

    interface PrintTaskPage {
      list: PrintTaskItem[]
      total: number
    }
  }

  /** 会员 */
  namespace Member {
    interface Item {
      id: string
      mobile: string
      nickname: string
      avatar: string
      name: string
      levelId: string
      /** 当前积分 */
      point: number
      /** 成长值 */
      experience: number
      status: number
    }

    interface Page {
      list: Item[]
      total: number
    }
  }

  /** 菜品与门店菜品配置 */
  namespace Dish {
    interface CategoryNode {
      id: string
      name: string
      children?: CategoryNode[]
    }

    /** 门店维度规格配置：品牌菜品信息 + 门店可售与沽清 */
    interface StoreSkuItem {
      skuId: string
      skuName: string
      skuCode: string
      spuId: string
      spuName: string
      categoryId: string
      categoryName: string
      /** 品牌标准价（分） */
      standardPrice: number
      /** 是否已配置过该门店；未配置按不可售处理 */
      configured: boolean
      available: number
      /** 0 正常 / 1 已沽清 */
      soldOut: number
      onlineOnly: number
      sort: number
      remark: string
      storePrice: number
      storeMemberPrice: number
      startAt?: string | null
      endAt?: string | null
    }

    interface StoreSkuPage {
      list: StoreSkuItem[]
      total: number
    }

    /** 菜单里的加料项 */
    interface MenuAddonItem {
      id: string
      name: string
      /** 加料价（分） */
      price: number
    }

    /** 菜单里的加料组：minSelect / maxSelect 决定必选与多选上限 */
    interface MenuAddonGroup {
      id: string
      name: string
      /** 1 表示必选 */
      required: number
      minSelect: number
      maxSelect: number
      items: MenuAddonItem[]
    }

    /** 菜单里的规格：price 是按会员身份算好的最终价 */
    interface MenuSku {
      id: string
      name: string
      status: number
      /** 1 表示沽清 */
      soldOut: number
      onlineOnly: number
      standardPrice: number
      storePrice: number
      /** 最终展示价（分） */
      price: number
    }

    /** 菜单里的菜品 */
    interface MenuSpu {
      id: string
      name: string
      unit: string
      description: string
      tags: string
      skus: MenuSku[]
      addonGroups: MenuAddonGroup[]
    }

    /** 菜单分类：children 为子分类，items 为直属菜品 */
    interface MenuCategory {
      id: string
      name: string
      children?: MenuCategory[]
      items: MenuSpu[]
    }

    interface MenuSnapshot {
      storeId: string
      builtAt: string
      categories: MenuCategory[]
    }
  }

  /** 订单 */
  namespace Order {
    interface Item {
      id: string
      no: string
      type: number
      typeText: string
      channelText: string
      tableName: string
      pickupCode: string
      /** 100 待支付 / 200 待接单 / 300 制作中 / 400 待取餐 / 500 已完成 / 600 已取消 / 700 退款中 / 800 已退款 */
      status: number
      statusText: string
      /** 0 未支付 / 10 已支付 / 20 已退款 */
      payStatus: number
      payStatusText: string
      payAmount: number
      refundAmount: number
      remark: string
      createdAt: string
      payTime?: string | null
    }

    interface Page {
      list: Item[]
      total: number
    }

    interface PageParams {
      status?: number
      payStatus?: number
      type?: number
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** POS 收款结果 */
    interface PayResult {
      orderNo: string
      payAmount: number
      payOrderNo: string
      channelCode: string
      paid: boolean
      payParams: Record<string, unknown>
    }

    interface PayStatusResult {
      orderNo: string
      paid: boolean
      payStatus: number
      waitScan: boolean
    }

    /** 试算明细行 */
    interface EstimateLine {
      skuId: string
      spuName: string
      skuName: string
      addons: { addonItemId: string; name: string; price: number; quantity: number }[]
      unitPrice: number
      quantity: number
      addonAmount: number
      lineDiscount: number
      subtotal: number
      orderDiscount: number
      payable: number
    }

    /** 试算结果：金额只信这个接口，前端不做任何计算 */
    interface EstimateResult {
      lines: EstimateLine[]
      itemAmount: number
      addonAmount: number
      lineDiscountAmount: number
      activityAmount: number
      couponAmount: number
      pointsAmount: number
      payAmount: number
      activityNames: string[]
      couponName: string
      notes: string[]
    }

    /** 门店今日经营概览，金额单位为分 */
    interface PosOverview {
      date: string
      revenue: number
      paidOrders: number
      avgPrice: number
      refundAmount: number
      pendingPay: number
      pendingAccept: number
      making: number
      waitingPickup: number
    }

    /** 报表的一天 */
    interface ReportTrendRow {
      date: string
      orders: number
      revenue: number
    }

    /** 菜品销量排行的一行 */
    interface ReportDishRow {
      spuName: string
      skuName: string
      quantity: number
      amount: number
    }

    /** 时段分布的一行，固定 24 条 */
    interface ReportHourRow {
      hour: string
      orders: number
      revenue: number
    }

    /** 退款明细的一行 */
    interface ReportRefundRow {
      orderNo: string
      payAmount: number
      refundAmount: number
      status: number
      statusText: string
      /** 订单表没有退款时间，这里是最后更新时间 */
      updatedAt: string
    }

    /** 门店经营报表，金额单位为分 */
    interface PosReport {
      days: number
      from: string
      to: string
      revenue: number
      orders: number
      avgPrice: number
      refundAmount: number
      cashAmount: number
      cashOrders: number
      onlineAmount: number
      onlineOrders: number
      trend: ReportTrendRow[]
      hourly: ReportHourRow[]
      topDishes: ReportDishRow[]
      refunds: ReportRefundRow[]
    }

    /** 下单结果 */
    interface CreateResult {
      order: {
        id: string
        no: string
        status: number
        statusText: string
        payAmount: number
        pickupCode: string
        expireTime: string
      }
      idempotentHit: boolean
    }
  }
}
