/**
 * API 类型：控制台统计
 * 全局命名空间 Api 的合并声明，使用处无需 import。
 */

declare namespace Api {
  namespace Dashboard {
    /** 趋势数据点 */
    interface TrendItem {
      /** 月-日，如 10-19 */
      date: string
      count: number
    }

    /** 最近登录记录 */
    interface LoginItem {
      username: string
      userIp: string
      /** 0 失败 / 1 成功 */
      result: number
      createdAt: string
    }

    /** 经营趋势数据点 */
    interface BusinessTrendItem {
      /** 月-日，如 10-19 */
      date: string
      /** 当日下单量，含未支付 */
      orderCount: number
      /** 当日营业额（分），只统计已支付订单 */
      saleAmount: number
    }

    /** 热销菜品 */
    interface TopProductItem {
      name: string
      quantity: number
      /** 销售额（分） */
      amount: number
    }

    /** 经营看板 */
    interface Business {
      /** 今日下单量，含未支付 */
      todayOrderCount: number
      /** 今日已支付订单数 */
      todayPaidCount: number
      /** 今日营业额（分） */
      todaySaleAmount: number
      /** 今日退款额（分） */
      todayRefundAmount: number
      /** 今日客单价（分） */
      todayAvgAmount: number
      orderTrend: BusinessTrendItem[]
      topProducts: TopProductItem[]
    }

    /** 控制台聚合数据 */
    interface Overview {
      counts: {
        user: number
        role: number
        menu: number
        file: number
        fileSize: number
        todayLogin: number
        todayRequest: number
        todayError: number
      }
      business: Business
      loginTrend: TrendItem[]
      requestTrend: TrendItem[]
      recentLogins: LoginItem[]
      system: {
        goVersion: string
        serverTime: string
      }
    }
  }
}
