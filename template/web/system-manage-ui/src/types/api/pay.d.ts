/**
 * API 类型：支付应用、支付渠道、支付单与渠道调用记录
 * 全局命名空间 Api 的合并声明，使用处无需 import。
 */

declare namespace Api {
  namespace Pay {
    /** 支付应用，一个应用可配置多个渠道，回调地址按应用维度配置 */
    interface AppItem {
      /** 雪花 ID，后端序列化为字符串 */
      id: string
      /** 应用名 */
      name: string
      /** 开启状态：0 开启 / 1 关闭（yudao 口径，与全站 1 启用相反） */
      status: number
      /** 状态文案 */
      statusName: string
      remark: string
      /** 支付结果回调地址，业务侧接收通知用 */
      orderNotifyUrl: string
      /** 退款结果回调地址 */
      refundNotifyUrl: string
      /** 转账结果回调地址，可为空 */
      transferNotifyUrl: string
      /** 应用标识，创建后不可修改 */
      appKey: string
      /** 已配置渠道数量 */
      channelCount: number
      /** 应用密钥明文，仅新增接口返回一次 */
      appSecret?: string
      createdAt?: string
      updatedAt?: string
    }

    /** 支付应用分页结果 */
    interface AppList {
      list: AppItem[]
      total: number
    }

    /** 支付应用列表查询参数 */
    interface AppSearchParams {
      /** 应用名 / appKey 关键字 */
      keyword?: string
      /** 0 开启 / 1 关闭 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 支付应用新增 / 编辑参数；appSecret 留空表示由系统生成或保持原值 */
    interface AppSaveParams {
      name: string
      status: number
      remark?: string
      orderNotifyUrl: string
      refundNotifyUrl: string
      transferNotifyUrl?: string
      appSecret?: string
    }

    /** 支付应用下拉选项 */
    interface AppOption {
      id: string
      name: string
    }

    /** 支付渠道，同一应用下渠道编码唯一 */
    interface ChannelItem {
      id: string
      /** 所属应用 ID */
      appId: string
      /** 渠道编码：wx_lite / wx_native / alipay_qr 等 */
      code: string
      /** 渠道中文名 */
      channelName: string
      /** 开启状态：0 开启 / 1 关闭 */
      status: number
      statusName: string
      remark: string
      /** 渠道费率，单位百分比 */
      feeRate: number
      /** 渠道配置 JSON，密钥字段已脱敏 */
      config: string
      createdAt?: string
      updatedAt?: string
    }

    /** 支付渠道分页结果 */
    interface ChannelList {
      list: ChannelItem[]
      total: number
    }

    /** 支付渠道列表查询参数 */
    interface ChannelSearchParams {
      /** 应用 ID，不传表示全部应用 */
      appId?: string
      /** 渠道编码 */
      code?: string
      /** 0 开启 / 1 关闭 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 支付渠道新增 / 编辑参数；config 中回传掩码表示沿用原密钥 */
    interface ChannelSaveParams {
      appId: string
      code: string
      status: number
      remark?: string
      feeRate: number
      config: string
    }

    /** 渠道编码下拉选项，由后端枚举下发 */
    interface ChannelOption {
      code: string
      name: string
    }

    /** 渠道调用记录，每次调起渠道下单都会新增一条 */
    interface OrderExtensionItem {
      id: string
      /** 支付单号 */
      no: string
      orderId: string
      channelId: string
      channelCode: string
      userIp: string
      /** 该次调用后的支付状态 */
      status: number
      /** 渠道额外参数 JSON，如二维码链接 */
      channelExtras: string
      channelErrorCode: string
      channelErrorMsg: string
      /** 渠道回调原文 */
      channelNotifyData: string
      createdAt?: string
    }

    /** 支付单 */
    interface OrderItem {
      id: string
      appId: string
      channelId: string
      channelCode: string
      channelName: string
      /** 付款用户 ID，'0' 表示未登录用户 */
      userId: string
      userType: number
      /** 业务侧订单号，用于下单幂等 */
      merchantOrderId: string
      subject: string
      body: string
      /** 业务回调地址 */
      notifyUrl: string
      /** 支付金额，单位分 */
      price: number
      /** 渠道费率快照，单位百分比 */
      channelFeeRate: number
      /** 渠道手续费快照，单位分 */
      channelFeePrice: number
      /** 0 待支付 / 10 已支付 / 20 已退款 / 30 已关闭 */
      status: number
      statusName: string
      /** 已退金额，单位分 */
      refundPrice: number
      userIp: string
      /** 支付单失效时间 */
      expireTime: string
      /** 支付成功时间，未支付为空 */
      successTime?: string
      extensionId: string
      /** 支付单号 */
      no: string
      /** 渠道订单号 */
      channelOrderNo: string
      /** 渠道用户标识，如微信 openid */
      channelUserId: string
      createdAt?: string
      updatedAt?: string
    }

    /** 支付单详情，附带渠道调用历史 */
    interface OrderDetail extends OrderItem {
      extensions: OrderExtensionItem[]
    }

    /** 支付单分页结果 */
    interface OrderList {
      list: OrderItem[]
      total: number
    }

    /** 支付单列表查询参数 */
    interface OrderSearchParams {
      /** 支付单号 / 商户订单号 / 渠道订单号关键字 */
      keyword?: string
      appId?: string
      channelCode?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 创建支付单（下单并调起渠道）参数，金额单位分 */
    interface OrderCreateParams {
      appId: string
      channelCode: string
      userId?: string
      userType?: number
      /** 业务订单号，同一应用下未关闭前不可重复 */
      merchantOrderId: string
      subject: string
      body?: string
      /** 支付金额，单位分 */
      price: number
      notifyUrl?: string
      /** 微信 JSAPI / 小程序支付必填 */
      openId?: string
      /** 支付宝条码支付必填 */
      authCode?: string
      userIp?: string
      /** 支付单有效期，单位分钟，0 表示默认 30 分钟 */
      expireMinutes?: number
    }

    /** 待支付单重新调起渠道参数 */
    interface OrderSubmitParams {
      openId?: string
      authCode?: string
    }

    /** 下单 / 重新调起的结果，前端按渠道取 payParams 调起支付 */
    interface OrderSubmitResult {
      orderId: string
      no: string
      channelCode: string
      /** 渠道额外参数 JSON */
      channelExtras: string
      /** 调起参数：JSAPI 签名参数 / 二维码链接 / 支付链接 / orderStr */
      payParams: Record<string, unknown>
    }

    /** 退款单，一个支付单可对应多笔部分退款 */
    interface RefundItem {
      id: string
      /** 退款单号 */
      no: string
      appId: string
      channelId: string
      channelCode: string
      channelName: string
      /** 支付单主键 */
      orderId: string
      /** 支付单号 */
      orderNo: string
      /** 业务支付订单号 */
      merchantOrderId: string
      /** 业务退款单号，用于退款幂等 */
      merchantRefundId: string
      /** 0 等待退款 / 10 退款成功 / 20 退款失败 */
      status: number
      statusName: string
      /** 原支付金额，单位分 */
      payPrice: number
      /** 本次退款金额，单位分 */
      refundPrice: number
      /** 退款原因 */
      reason: string
      /** 渠道订单号 */
      channelOrderNo: string
      /** 渠道退款单号 */
      channelRefundNo: string
      /** 渠道错误码 */
      channelErrorCode: string
      /** 渠道错误文案 */
      channelErrorMsg: string
      /** 渠道退款回调原文 */
      channelNotifyData: string
      userIp: string
      /** 退款成功时间，未成功为空 */
      successTime?: string
      /** 审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回 */
      auditStatus: number
      auditStatusName: string
      /** 审核时间 */
      auditTime?: string | null
      /** 审核人后台用户 ID */
      auditBy: string
      /** 审核意见 / 驳回原因 */
      auditRemark: string
      createdAt?: string
      updatedAt?: string
    }

    /** 退款单分页结果 */
    interface RefundList {
      list: RefundItem[]
      total: number
    }

    /** 退款单列表查询参数 */
    interface RefundSearchParams {
      /** 退款单号 / 支付单号 / 渠道退款单号关键字 */
      keyword?: string
      appId?: string
      orderId?: string
      channelCode?: string
      status?: number
      /** 审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回 */
      auditStatus?: number
      pageNo?: number
      pageSize?: number
    }

    /** 发起退款参数，金额单位分 */
    interface RefundCreateParams {
      /** 支付单主键 */
      orderId: string
      /** 业务退款单号，同一应用下不可重复 */
      merchantRefundId: string
      /** 本次退款金额，单位分 */
      refundPrice: number
      reason?: string
      userIp?: string
      /** 是否需要审核：true 时只生成退款单，审核通过后才真正退钱 */
      needAudit?: boolean
    }

    /** 退款审核参数 */
    interface RefundAuditParams {
      /** 审核意见；驳回时必填 */
      remark?: string
    }

    /** 业务回调通知任务，支付 / 退款成功时自动登记 */
    interface NotifyTaskItem {
      id: string
      appId: string
      /** 1 支付 / 2 退款 / 3 转账 */
      type: number
      typeName: string
      /** 业务数据编号（支付单 / 退款单 ID） */
      dataId: string
      merchantOrderId: string
      /** 业务方回调地址 */
      notifyUrl: string
      /** 0 等待通知 / 10 通知成功 / 20 通知失败 */
      status: number
      statusName: string
      /** 下一次通知时间 */
      nextNotifyTime: string
      /** 最后一次执行时间 */
      lastExecuteTime: string
      /** 已通知次数 */
      notifyTimes: number
      /** 最大通知次数 */
      maxNotifyTimes: number
      createdAt?: string
    }

    /** 通知任务分页结果 */
    interface NotifyTaskList {
      list: NotifyTaskItem[]
      total: number
    }

    /** 通知任务列表查询参数 */
    interface NotifyTaskSearchParams {
      /** 商户订单号关键字 */
      keyword?: string
      appId?: string
      /** 1 支付 / 2 退款 / 3 转账 */
      type?: number
      /** 0 等待 / 10 成功 / 20 失败 */
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 通知日志，每次派发写一条 */
    interface NotifyLogItem {
      id: string
      taskId: string
      /** 第几次通知 */
      notifyTimes: number
      /** 业务方响应内容（已截断） */
      response: string
      /** 0 等待 / 10 成功 / 20 失败 */
      status: number
      createdAt?: string
    }

    /** 通知日志分页结果 */
    interface NotifyLogList {
      list: NotifyLogItem[]
      total: number
    }

    /** 会员钱包，金额单位分 */
    interface WalletItem {
      id: string
      userId: string
      userType: number
      /** 可用余额，单位分 */
      balance: number
      /** 冻结金额，单位分 */
      freezePrice: number
      /** 累计支出，单位分 */
      totalExpense: number
      /** 累计充值，单位分 */
      totalRecharge: number
      createdAt?: string
      updatedAt?: string
    }

    /** 钱包分页结果 */
    interface WalletList {
      list: WalletItem[]
      total: number
    }

    /** 钱包列表查询参数 */
    interface WalletSearchParams {
      userId?: string
      userType?: number
      pageNo?: number
      pageSize?: number
    }

    /** 钱包流水，price 恒为正数，方向由 type 决定 */
    interface WalletTransactionItem {
      id: string
      walletId: string
      /** 业务类型：recharge / payment / refund / manual */
      bizType: number
      /** 业务编号 */
      bizId: string
      /** 1 收入 / 2 支出 */
      type: number
      /** 变动金额，单位分 */
      price: number
      /** 变动后余额，单位分 */
      balance: number
      remark: string
      createdAt?: string
    }

    /** 钱包流水分页结果 */
    interface WalletTransactionList {
      list: WalletTransactionItem[]
      total: number
    }

    /** 钱包流水查询参数 */
    interface WalletTransactionSearchParams {
      /** recharge / payment / refund / manual */
      bizType?: number
      /** 1 收入 / 2 支出 */
      type?: number
      pageNo?: number
      pageSize?: number
    }

    /** 调整钱包余额参数，price 正数加款、负数扣款 */
    interface WalletChangeParams {
      price: number
      remark?: string
    }

    /** 充值套餐，到账金额 = payPrice + bonusPrice */
    interface WalletPackageItem {
      id: string
      name: string
      /** 实付金额，单位分 */
      payPrice: number
      /** 赠送金额，单位分 */
      bonusPrice: number
      /** 0 开启 / 1 关闭 */
      status: number
      description: string
      createdAt?: string
      updatedAt?: string
    }

    /** 充值套餐分页结果 */
    interface WalletPackageList {
      list: WalletPackageItem[]
      total: number
    }

    /** 充值套餐新增 / 编辑参数，金额单位分 */
    interface WalletPackageSaveParams {
      name: string
      payPrice: number
      bonusPrice: number
      status: number
      description?: string
    }

    /** 充值记录 */
    interface WalletRechargeItem {
      id: string
      walletId: string
      /** 到账总金额，单位分 */
      totalPrice: number
      /** 实付金额，单位分 */
      payPrice: number
      /** 赠送金额，单位分 */
      bonusPrice: number
      /** 套餐编号，'0' 表示自定义金额 */
      packageId: string
      /** 0 未支付 / 10 已支付 / 20 已退款 */
      status: number
      statusName: string
      payOrderId: string
      payChannelCode: string
      payTime?: string
      refundOrderId: string
      refundPrice: number
      refundTime?: string
      createdAt?: string
    }

    /** 充值记录分页结果 */
    interface WalletRechargeList {
      list: WalletRechargeItem[]
      total: number
    }

    /** 充值记录查询参数 */
    interface WalletRechargeSearchParams {
      walletId?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }
  }
}
