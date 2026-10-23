declare namespace Api {
  namespace Promotion {
    /** 优惠券类型：1 满减 / 2 折扣 / 3 单品 / 4 兑换 / 5 代金 */
    type CouponType = 1 | 2 | 3 | 4 | 5
    /** 券状态：1 未使用 / 2 锁定 / 3 已使用 / 4 已过期 */
    type CouponStatus = 1 | 2 | 3 | 4
    /** 适用范围：1 全品牌 / 2 指定门店 / 3 指定商品 */
    type CouponScope = 1 | 2 | 3
    /** 渠道限制：0 不限 / 1 仅线上 / 2 仅线下 */
    type CouponChannelLimit = 0 | 1 | 2
    /** 发放方式：1 主动领取 / 2 消费返券 / 3 新人礼 / 4 定向发券 */
    type CouponIssueType = 1 | 2 | 3 | 4
    /** 活动玩法：1 满减 / 2 第二份半价 / 3 限时折扣 / 4 套餐特惠 / 5 加价购 */
    type ActivityType = 1 | 2 | 3 | 4 | 5
    /** 活动层级：1 商品级 / 2 订单级 */
    type ActivityLevel = 1 | 2
    /** 活动门店范围：1 全品牌 / 2 指定门店 */
    type ActivityScope = 1 | 2
    /** 名额周期：1 整个活动期 / 2 每日重置 */
    type ActivityQuotaCycle = 1 | 2

    /** 优惠券模板列表行 */
    interface CouponTemplateItem {
      id: string
      name: string
      /** 券类型 */
      type: CouponType
      typeText: string
      description: string
      /** 使用门槛（分） */
      thresholdAmount: number
      /** 抵扣金额（分） */
      discountAmount: number
      /** 折扣率（万分比），7500 表示 75 折 */
      discountRate: number
      /** 折扣封顶（分） */
      maxDiscountAmount: number
      /** 适用范围 */
      scope: CouponScope
      scopeText: string
      /** 适用对象 ID 列表，逗号分隔 */
      scopeIds: string
      /** 渠道限制 */
      channelLimit: CouponChannelLimit
      channelLimitText: string
      /** 发放方式 */
      issueType: CouponIssueType
      issueTypeText: string
      /** 发行总量，0 表示不限 */
      totalCount: number
      /** 已发放数量 */
      issuedCount: number
      /** 每人限领张数 */
      perMemberLimit: number
      startTime: string
      endTime: string
      /** 领取后有效天数 */
      validDays: number
      /** 启用状态：0 关闭 / 1 开启 */
      status: number
      createdAt: string
    }

    /** 券模板分页结果 */
    interface CouponTemplateList {
      list: CouponTemplateItem[]
      total: number
    }

    /** 券模板列表查询参数 */
    interface CouponTemplateSearchParams {
      /** 券名称 */
      keyword?: string
      /** 券类型 */
      type?: number
      /** 启用状态 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 券模板保存参数 */
    interface CouponTemplateSaveParams {
      /** 模板 ID，0 表示新增 */
      id?: string
      name: string
      /** 券类型 */
      type: CouponType
      description?: string
      /** 使用门槛（分） */
      thresholdAmount?: number
      /** 抵扣金额（分） */
      discountAmount?: number
      /** 折扣率（万分比） */
      discountRate?: number
      /** 折扣封顶（分） */
      maxDiscountAmount?: number
      /** 适用范围 */
      scope: CouponScope
      /** 适用对象 ID 列表 */
      scopeIds?: string[]
      /** 渠道限制 */
      channelLimit?: CouponChannelLimit
      /** 发放方式 */
      issueType: CouponIssueType
      /** 发行总量 */
      totalCount?: number
      /** 每人限领张数 */
      perMemberLimit?: number
      startTime: string
      endTime: string
      /** 领取后有效天数 */
      validDays?: number
      /** 启用状态 */
      status?: number
    }

    /** 会员券列表行 */
    interface CouponItem {
      id: string
      templateId: string
      templateName: string
      memberId: string
      /** 券状态 */
      status: CouponStatus
      statusText: string
      /** 占用该券的订单 ID */
      orderId: string
      useTime: string | null
      expireTime: string
      getTime: string
      createdAt: string
    }

    /** 会员券分页结果 */
    interface CouponList {
      list: CouponItem[]
      total: number
    }

    /** 会员券列表查询参数 */
    interface CouponSearchParams {
      templateId?: string
      memberId?: string
      /** 券状态 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 阶梯满减的一档 */
    interface ActivityTier {
      /** 该档门槛（分） */
      thresholdAmount: number
      /** 该档减免金额（分） */
      discountAmount: number
    }

    /** 套餐中的一个商品 */
    interface ActivityComboItem {
      spuId: string
      quantity: number
    }

    /** 营销活动列表行 */
    interface ActivityItem {
      id: string
      name: string
      /** 活动玩法 */
      type: ActivityType
      typeText: string
      /** 作用层级 */
      level: ActivityLevel
      levelText: string
      description: string
      /** 玩法规则 JSON */
      rule: string
      /** 适用门店范围 */
      scope: ActivityScope
      scopeText: string
      /** 适用门店 ID 列表，逗号分隔 */
      scopeIds: string
      /** 优先级，数值大的先判定 */
      priority: number
      /** 互斥组 */
      mutexGroup: string
      /** 是否独占 */
      exclusive: number
      /** 是否可叠加 */
      stackable: number
      /** 名额上限 */
      quotaLimit: number
      /** 名额周期 */
      quotaCycle: ActivityQuotaCycle
      startTime: string
      endTime: string
      /** 启用状态：0 未启用 / 1 已启用 */
      status: number
      statusText: string
      /** 参与商品的 SPU ID 列表 */
      productSpuIds: string[]
      createdAt: string
    }

    /** 营销活动分页结果 */
    interface ActivityList {
      list: ActivityItem[]
      total: number
    }

    /** 营销活动列表查询参数 */
    interface ActivitySearchParams {
      /** 活动名称 */
      keyword?: string
      /** 活动玩法 */
      type?: number
      /** 启用状态 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 营销活动保存参数 */
    interface ActivitySaveParams {
      /** 活动 ID，0 表示新增 */
      id?: string
      name: string
      /** 活动玩法 */
      type: ActivityType
      /** 作用层级 */
      level: ActivityLevel
      description?: string
      /** 满减门槛（分） */
      thresholdAmount?: number
      /** 减免金额（分） */
      discountAmount?: number
      /** 阶梯满减档位 */
      tiers?: ActivityTier[]
      /** 折扣率（万分比） */
      discountRate?: number
      /** 折扣封顶（分） */
      maxDiscountAmount?: number
      /** 第二份折扣率（万分比） */
      secondHalfRate?: number
      /** 套餐商品组合 */
      comboItems?: ActivityComboItem[]
      /** 套餐价（分） */
      comboPrice?: number
      /** 加价购商品 SPU ID */
      addonSpuId?: string
      /** 加价购价格（分） */
      addonPrice?: number
      /** 适用门店范围 */
      scope: ActivityScope
      /** 适用门店 ID 列表 */
      scopeIds?: string[]
      priority?: number
      /** 互斥组 */
      mutexGroup?: string
      /** 是否独占 */
      exclusive?: boolean
      /** 是否可叠加 */
      stackable?: boolean
      /** 名额上限 */
      quotaLimit?: number
      /** 名额周期 */
      quotaCycle?: ActivityQuotaCycle
      startTime: string
      endTime: string
      /** 启用状态 */
      status?: number
      /** 参与商品的 SPU ID 列表 */
      productSpuIds?: string[]
    }

    /** 积分抵扣规则 */
    interface PointRule {
      /** 兑换比例：多少积分抵 1 元（分） */
      exchangeRate: number
      /** 单笔抵扣上限（分） */
      maxDeductAmount: number
      /** 使用积分抵扣的最低订单金额（分） */
      minOrderAmount: number
      /** 启用状态 */
      status: number
    }

    /** 积分抵扣规则保存参数 */
    interface PointRuleSaveParams {
      exchangeRate: number
      maxDeductAmount?: number
      minOrderAmount?: number
      status?: number
    }

    /** 发券参数 */
    interface CouponIssueParams {
      /** 目标会员 ID */
      memberId: string
      /** 发放张数 */
      count?: number
    }
  }
}
