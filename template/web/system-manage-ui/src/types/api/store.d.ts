/**
 * API 类型：门店组织树、营业时间与配送规则
 * 全局命名空间 Api 的合并声明，使用处无需 import。
 */

declare namespace Api {
  namespace Store {
    /** 门店节点，总部 / 大区 / 门店共用同一结构 */
    interface StoreItem {
      /** 雪花 ID，后端序列化为字符串 */
      id: string
      /** 上级节点 ID，'0' 表示总部根节点 */
      parentId: string
      /** 节点类型：1 总部 / 2 大区 / 3 门店 */
      type: number
      /** 门店编码，全局唯一，创建后不可修改 */
      code: string
      name: string
      shortName: string
      /** 经营状态：1 营业 / 2 休息 / 3 装修 / 4 停业 */
      status: number
      contactName: string
      contactMobile: string
      /** 区县 ID，'0' 表示未选择 */
      districtId: string
      address: string
      /** 经度，未配置为 0 */
      longitude: number
      /** 纬度，未配置为 0 */
      latitude: number
      /** 营业开始时间 HH:mm */
      openTime: string
      /** 营业结束时间 HH:mm；小于开始时间表示跨夜 */
      closeTime: string
      /** 配送规则 JSON 字符串，结构见 DeliveryRule */
      deliveryRule: string
      /** 是否自动接单：1 自动 / 0 人工 */
      autoAccept: number
      /** 是否开放线上：1 是 / 0 否 */
      onlineEnabled: number
      /** 是否开启外卖：1 是 / 0 否，默认关闭 */
      deliveryEnabled: number
      sort: number
      remark: string
      createdAt?: string
    }

    /** 组织树节点，children 为空数组表示叶子 */
    interface StoreTreeNode extends StoreItem {
      children: StoreTreeNode[]
    }

    /** 门店分页结果 */
    interface StoreList {
      list: StoreItem[]
      total: number
    }

    /** 门店列表查询参数 */
    interface StoreSearchParams {
      /** 名称 / 编码 / 简称关键字 */
      keyword?: string
      type?: number
      status?: number
      /** 组织节点 ID：只看该节点及下级中的门店，雪花 ID 用字符串 */
      entityId?: string
      pageNo?: number
      pageSize?: number
    }

    /** 配送规则：距离用米、金额用分、时长用分钟 */
    interface DeliveryRule {
      /** 配送半径（米），0 表示不限范围 */
      radius: number
      /** 起送价（分） */
      minAmount: number
      /** 配送费（分） */
      deliveryFee: number
      /** 每份打包费（分） */
      packingFee: number
      /** 预计送达时长（分钟） */
      etaMinutes: number
    }

    /** 门店新增 / 编辑参数，deliveryRule 为 DeliveryRule 的 JSON 字符串 */
    interface StoreSaveParams {
      /** 雪花 ID，必须按字符串提交，避免 JS Number 丢精度；总部传 '0' */
      parentId: string
      type: number
      /** 编辑时后端忽略该字段 */
      code: string
      name: string
      shortName?: string
      status: number
      contactName?: string
      contactMobile?: string
      /** 区县雪花 ID，按字符串提交；未选择传 '0' */
      districtId?: string
      address?: string
      longitude?: number
      latitude?: number
      /** HH:mm */
      openTime: string
      /** HH:mm，小于 openTime 表示跨夜 */
      closeTime: string
      deliveryRule?: string
      autoAccept?: number
      onlineEnabled?: number
      /** 是否开启外卖：默认 0 */
      deliveryEnabled?: number
      sort?: number
      remark?: string
    }

    /** 门店切换器使用的可见门店，id 为雪花字符串 */
    interface VisibleStore {
      id: string
      code: string
      name: string
      /** 经营状态：1 营业 / 2 休息 / 3 装修 / 4 停业 */
      status: number
    }

    /** 门店员工关联记录，含门店与后台账号展示字段 */
    interface EmployeeItem {
      id: string
      storeId: string
      userId: string
      /** 门店角色：1 店长 / 2 收银 / 3 后厨 / 4 服务员 / 5 骑手 */
      role: number
      roleName: string
      /** 启停状态：0 禁用 / 1 启用 */
      status: number
      remark: string
      storeName: string
      username: string
      nickname: string
      mobile: string
      createdAt?: string
      updatedAt?: string
    }

    /** 门店员工分页结果 */
    interface EmployeeList {
      list: EmployeeItem[]
      total: number
    }

    /** 门店员工列表查询参数，门店范围由请求头 X-Store-Id 或可见集合决定 */
    interface EmployeeSearchParams {
      /** 账号 / 昵称 / 手机号关键字 */
      keyword?: string
      role?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 门店员工新增 / 编辑参数，雪花 ID 一律用字符串 */
    interface EmployeeSaveParams {
      storeId: string
      userId: string
      role: number
      /** 0 禁用 / 1 启用 */
      status: number
      remark?: string
    }

    /** 员工分配弹窗的后台账号下拉项 */
    interface EmployeeUserOption {
      id: string
      username: string
      nickname: string
      mobile: string
      /** 账号状态：0 禁用 / 1 启用 */
      status: number
    }

    /** 桌台区域，按门店归属 */
    interface TableAreaItem {
      id: string
      storeId: string
      name: string
      sort: number
      remark: string
      /** 所属门店名称，列表展示用 */
      storeName: string
      /** 区域下桌台数量 */
      tableCount: number
      createdAt?: string
      updatedAt?: string
    }

    /** 桌台区域分页结果 */
    interface TableAreaList {
      list: TableAreaItem[]
      total: number
    }

    /** 桌台区域列表查询参数 */
    interface TableAreaSearchParams {
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 桌台区域新增 / 编辑参数 */
    interface TableAreaSaveParams {
      storeId: string
      name: string
      sort?: number
      remark?: string
    }

    /** 桌台区域下拉项，桌台筛选与弹窗共用 */
    interface TableAreaOption {
      id: string
      storeId: string
      storeName: string
      name: string
      sort: number
    }

    /** 桌台记录，含门店、区域与状态展示字段 */
    interface TableItem {
      id: string
      storeId: string
      areaId: string
      /** 桌号，如 A01 */
      name: string
      /** 桌码，全局唯一 */
      code: string
      seats: number
      sort: number
      /** 桌台状态：1 空闲 / 2 使用中 / 3 待清台 */
      status: number
      /** 当前台位单 ID，'0' 表示没有进行中的台位单 */
      currentOrderId: string
      /** 开台时间，清台后为 null */
      openedAt: string | null
      remark: string
      storeName: string
      areaName: string
      statusName: string
      createdAt?: string
      updatedAt?: string
    }

    /** 桌台分页结果 */
    interface TableList {
      list: TableItem[]
      total: number
    }

    /** 桌台列表查询参数，门店范围由请求头 X-Store-Id 或可见集合决定 */
    interface TableSearchParams {
      keyword?: string
      areaId?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 桌台新增 / 编辑参数；编辑时后端忽略 code */
    interface TableSaveParams {
      storeId: string
      areaId: string
      name: string
      code?: string
      seats: number
      sort?: number
      remark?: string
    }

    /** 批量生成桌台参数 */
    interface TableBatchParams {
      storeId: string
      areaId: string
      prefix: string
      start: number
      count: number
      seats: number
    }

    /** 批量生成桌台结果 */
    interface TableBatchResult {
      count: number
    }

    /** 骑手档案，关联门店员工与后台账号 */
    interface RiderItem {
      id: string
      storeId: string
      employeeId: string
      /** 骑手状态：1 空闲 / 2 接单中 / 3 配送中 / 4 休息 / 5 停用 */
      status: number
      statusName: string
      /** 接单并发上限 */
      maxOrders: number
      /** 当前进行中的配送单数量 */
      currentOrders: number
      /** 今日累计接单量，跨日自动归零展示 */
      todayOrders: number
      /** 今日单量统计日期 */
      statDate: string | null
      remark: string
      storeName: string
      userId: string
      username: string
      nickname: string
      mobile: string
      /** 门店员工启停状态：0 禁用 / 1 启用 */
      employeeStatus: number
      createdAt?: string
      updatedAt?: string
    }

    /** 骑手档案分页结果 */
    interface RiderList {
      list: RiderItem[]
      total: number
    }

    /** 骑手档案列表查询参数，门店范围由请求头 X-Store-Id 或可见集合决定 */
    interface RiderSearchParams {
      /** 后台账号 / 昵称 / 手机号关键字 */
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 骑手新增参数；userId 为 '0' 时按 mobile 自动开通后台账号 */
    interface RiderCreateParams {
      storeId: string
      userId?: string
      /** 自动开通账号时作为昵称 */
      name?: string
      /** 自动开通账号时作为登录账号 */
      mobile?: string
      maxOrders: number
      status: number
      remark?: string
    }

    /** 骑手编辑参数；门店、员工与账号创建后不可变更 */
    interface RiderUpdateParams {
      maxOrders: number
      status: number
      remark?: string
    }

    /** 骑手新增结果，initialPassword 仅在自动开通账号时返回一次 */
    interface RiderCreateResult {
      id: string
      employeeId: string
      userId: string
      username: string
      initialPassword?: string
    }

    /** 骑手弹窗选择已有后台账号时的下拉项 */
    interface RiderUserOption {
      id: string
      username: string
      nickname: string
      mobile: string
      /** 账号状态：0 禁用 / 1 启用 */
      status: number
    }

    /** 打印设备，按门店归属；密钥只返回脱敏值 */
    interface PrinterItem {
      id: string
      storeId: string
      /** 设备名称，如前台小票机 */
      name: string
      /** 设备号，同一门店内唯一 */
      deviceNo: string
      /** 设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号 */
      type: number
      typeName: string
      /** 接入方式：1 云打印机 / 2 局域网 / 3 本地客户端 */
      accessMode: number
      accessModeName: string
      /** 非敏感配置 JSON 字符串，禁止包含密钥 */
      config: string
      /** 密钥脱敏值，如 ****abcd；未配置时为空 */
      secretMask: string
      /** 是否已配置密钥 */
      hasSecret: boolean
      /** 默认打印份数 */
      copies: number
      /** 启停状态：0 禁用 / 1 启用 */
      status: number
      statusName: string
      remark: string
      storeName: string
      createdAt?: string
      updatedAt?: string
    }

    /** 打印设备分页结果 */
    interface PrinterList {
      list: PrinterItem[]
      total: number
    }

    /** 打印设备列表查询参数，门店范围由请求头 X-Store-Id 或可见集合决定 */
    interface PrinterSearchParams {
      /** 设备名称 / 设备号关键字 */
      keyword?: string
      type?: number
      accessMode?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 打印设备新增 / 编辑参数；secret 留空表示编辑时沿用原密钥 */
    interface PrinterSaveParams {
      storeId: string
      name: string
      deviceNo: string
      type: number
      accessMode: number
      config?: string
      secret?: string
      copies: number
      status: number
      remark?: string
    }

    /** 打印模板，按门店与「设备类型 + 打印类型」唯一 */
    interface PrintTemplateItem {
      id: string
      storeId: string
      name: string
      /** 适用设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号 */
      printerType: number
      printerTypeName: string
      /** 业务打印类型：1 结账单 / 2 预结单 / 3 厨打 / 4 退菜 / 5 标签 / 6 退款单 / 7 交班 / 8 日结 */
      printType: number
      printTypeName: string
      /** 纸宽，单位毫米 */
      paperWidth: number
      /** 模板正文 */
      content: string
      /** 当前版本号，正文变化时后端自动递增 */
      version: number
      /** 启停状态：0 禁用 / 1 启用 */
      status: number
      statusName: string
      remark: string
      storeName: string
      createdAt?: string
      updatedAt?: string
    }

    /** 打印模板分页结果 */
    interface PrintTemplateList {
      list: PrintTemplateItem[]
      total: number
    }

    /** 打印模板列表查询参数，门店范围由请求头 X-Store-Id 或可见集合决定 */
    interface PrintTemplateSearchParams {
      /** 模板名称关键字 */
      keyword?: string
      printerType?: number
      printType?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 打印模板新增 / 编辑参数；版本号由后端维护 */
    interface PrintTemplateSaveParams {
      storeId: string
      name: string
      printerType: number
      printType: number
      paperWidth: number
      content: string
      status: number
      remark?: string
    }
  }
}
