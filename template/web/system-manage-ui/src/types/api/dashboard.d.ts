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
