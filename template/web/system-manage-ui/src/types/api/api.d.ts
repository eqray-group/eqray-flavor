/**
 * API 接口类型定义模块
 *
 * 提供所有后端接口的类型定义
 *
 * ## 主要功能
 *
 * - 通用类型（分页参数、响应结构等）
 * - 认证类型（登录、用户信息等）
 * - 系统管理类型（用户、角色等）
 * - 全局命名空间声明
 *
 * ## 使用场景
 *
 * - API 请求参数类型约束
 * - API 响应数据类型定义
 * - 接口文档类型同步
 *
 * ## 注意事项
 *
 * - 在 .vue 文件使用需要在 eslint.config.mjs 中配置 globals: { Api: 'readonly' }
 * - 使用全局命名空间，无需导入即可使用
 *
 * ## 使用方式
 *
 * ```typescript
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const response: Api.Auth.UserInfo = await fetchUserInfo()
 * ```
 *
 * @module types/api/api
 * @author Art Design Pro Team
 */

declare namespace Api {
  /** 通用类型 */
  namespace Common {
    /** 分页参数 */
    interface PaginationParams {
      /** 当前页码 */
      current: number
      /** 每页条数 */
      size: number
      /** 总条数 */
      total: number
    }

    /** 通用搜索参数 */
    type CommonSearchParams = Pick<PaginationParams, 'current' | 'size'>

    /** 分页响应基础结构 */
    interface PaginatedResponse<T = any> {
      records: T[]
      current: number
      size: number
      total: number
    }

    /** 启用状态 */
    type EnableStatus = '1' | '2'
  }

  /** 认证类型 */
  namespace Auth {
    /** 登录参数（与后端 auth/login 对齐） */
    interface LoginParams {
      username: string
      password: string
      captchaId: string
      captchaCode: string
    }

    /** 登录 / 续期响应：访问令牌短期有效，刷新令牌用于换新 */
    interface LoginResponse {
      accessToken: string
      refreshToken: string
      /** 访问令牌有效期（秒） */
      expiresIn: number
      /** 刷新令牌有效期（秒） */
      refreshExpiresIn: number
    }

    /** 图片验证码 */
    interface Captcha {
      captchaId: string
      image: string
    }

    /** 用户信息 */
    interface UserInfo {
      buttons: string[]
      roles: string[]
      /** 雪花 ID，后端序列化为字符串 */
      userId: string
      userName: string
      nickname?: string
      email: string
      mobile?: string | null
      /** 0 未知 / 1 男 / 2 女 */
      sex?: number
      avatar?: string
    }

    /** 个人资料更新参数 */
    interface ProfileUpdateParams {
      nickname: string
      email?: string
      mobile?: string
      sex: number
      avatar?: string
    }

    /** 修改密码参数 */
    interface ProfilePasswordParams {
      oldPassword: string
      newPassword: string
    }
  }

  /** 基础设施类型 */
  namespace Infra {
    /** 短信渠道配置（对齐 yudao system_sms_channel） */
    interface SmsChannelItem {
      id: string
      name: string
      /** DebugLog / AliyunSms / TencentSms */
      code: string
      apiKey: string
      /** 短信签名 */
      signature: string
      /** 扩展配置 JSON，腾讯云为 {"region","sdkAppId"} */
      extra: string
      status: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface SmsChannelList {
      list: SmsChannelItem[]
      total: number
    }

    interface SmsChannelSearchParams {
      keyword?: string
      code?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface SmsChannelSaveParams {
      name: string
      code: string
      apiKey?: string
      /** 编辑留空表示不修改 */
      apiSecret?: string
      signature?: string
      extra?: string
      status: number
      remark?: string
    }

    /** 短信模板（对齐 yudao system_sms_template） */
    interface SmsTemplateItem {
      id: string
      channelId: string
      code: string
      name: string
      /** 1 验证码 / 2 通知 */
      type: number
      apiTemplateId: string
      content: string
      /** 模板变量名，逗号分隔 */
      params: string
      status: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface SmsTemplateList {
      list: SmsTemplateItem[]
      total: number
    }

    interface SmsTemplateSearchParams {
      keyword?: string
      channelId?: string
      type?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface SmsTemplateSaveParams {
      channelId: string
      code: string
      name: string
      type: number
      apiTemplateId?: string
      content: string
      params?: string
      status: number
      remark?: string
    }

    /** 参数配置 */
    interface ConfigItem {
      id: string
      category: string
      name: string
      configKey: string
      value: string
      /** 1 系统内置 / 2 自定义 */
      type: number
      /** 0 否 / 1 是 */
      visible: number
      remark: string
      creator?: string
      createdAt?: string
      updatedAt?: string
    }

    /** 参数配置列表 */
    interface ConfigList {
      list: ConfigItem[]
      total: number
    }

    /** 参数配置搜索参数 */
    interface ConfigSearchParams {
      category?: string
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 参数配置新增参数 */
    interface ConfigCreateParams {
      category: string
      name: string
      configKey: string
      value: string
      type: number
      visible: number
      remark?: string
    }

    /** 参数配置编辑参数，参数键与类型不可修改 */
    type ConfigUpdateParams = Omit<ConfigCreateParams, 'configKey' | 'type'>

    /** 文件记录 */
    interface FileItem {
      id: string
      /** 原始文件名 */
      name: string
      /** 存储路径 */
      path: string
      /** 访问地址 */
      url: string
      /** MIME 类型 */
      type: string
      /** 文件大小（字节） */
      size: number
      createdAt: string
    }

    /** 文件列表 */
    interface FileList {
      list: FileItem[]
      total: number
    }

    /** 文件搜索参数 */
    interface FileSearchParams {
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 文件配置 */
    interface FileConfigItem {
      id: string
      name: string
      /** 存储类型：local 本地磁盘 / s3 S3 兼容存储（MinIO、OSS、COS 等） */
      storage: string
      /** 1 主配置 / 0 普通 */
      master: number
      /** 存储配置 JSON */
      config: string
      createdAt?: string
    }

    /** 文件配置新增 / 编辑参数 */
    interface FileConfigParams {
      name: string
      storage: string
      master: number
      config: string
    }

    /** 文件配置搜索参数 */
    interface FileConfigSearchParams {
      name?: string
      storage?: string
      pageNo?: number
      pageSize?: number
    }

    /** API 访问日志 */
    interface ApiAccessLogItem {
      id: string
      traceId: string
      userId: string
      userType: number
      applicationName: string
      requestMethod: string
      requestUrl: string
      requestParams: string
      responseBody: string
      userIp: string
      userAgent: string
      operateModule: string
      operateName: string
      beginAt: string
      endAt: string
      duration: number
      resultCode: number
      createdAt: string
    }

    /** API 访问日志列表 */
    interface ApiAccessLogList {
      list: ApiAccessLogItem[]
      total: number
    }

    /** API 错误日志 */
    interface ApiErrorLogItem {
      id: string
      traceId: string
      userId: string
      requestMethod: string
      requestUrl: string
      requestParams: string
      userIp: string
      userAgent: string
      name: string
      message: string
      exceptionStack: string
      processStatus: number
      processTime?: string
      createdAt: string
    }

    /** API 错误日志列表 */
    interface ApiErrorLogList {
      list: ApiErrorLogItem[]
      total: number
    }

    /** API 日志搜索参数 */
    interface ApiLogSearchParams {
      requestUrl?: string
      keyword?: string
      resultCode?: number
      pageNo?: number
      pageSize?: number
    }

    /** 上传成功后的文件信息 */
    interface UploadedFile {
      /** 雪花 ID，后端序列化为字符串 */
      id: string
      /** 原始文件名 */
      name: string
      /** 可直接访问的地址 */
      url: string
      /** MIME 类型 */
      type: string
      /** 文件大小（字节） */
      size: number
    }
  }

  /** 会员管理类型 */
  namespace Member {
    /** 会员档案（对齐 yudao member_user） */
    interface MemberItem {
      id: string
      mobile: string
      nickname: string
      avatar: string
      /** 真实姓名 */
      name: string
      /** 0 未知 / 1 男 / 2 女 */
      sex: number
      birthday: string | null
      areaId: number
      /** 会员等级 ID */
      levelId: string
      /** 积分 */
      point: number
      /** 经验值 */
      experience: number
      /** 标签 ID，逗号分隔 */
      tagIds: string
      /** 用户标记 */
      mark: string
      registerIp: string
      loginIp: string
      loginDate: string | null
      /** 会员分组 ID */
      groupId: string
      /** 0 禁用 / 1 启用 */
      status: number
      creator?: string
      createdAt?: string
      updatedAt?: string
    }

    /** 会员列表 */
    interface MemberList {
      list: MemberItem[]
      total: number
    }

    /** 会员搜索参数 */
    interface MemberSearchParams {
      nickname?: string
      mobile?: string
      levelId?: string
      groupId?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 会员新增 / 编辑参数 */
    interface MemberSaveParams {
      nickname: string
      mobile: string
      avatar?: string
      name?: string
      sex: number
      /** 格式：YYYY-MM-DD HH:mm:ss */
      birthday?: string
      areaId?: number
      levelId: string
      point?: number
      experience?: number
      tagIds?: string
      mark?: string
      groupId?: string
      status: number
    }

    /** 第三方登录渠道配置（对齐 yudao system_social_client） */
    interface SocialClientItem {
      id: string
      name: string
      /** wechat_mini / wechat_mp / wechat_open */
      socialType: string
      clientId: string
      /** 列表不返回密钥 */
      agentId: string
      /** 0 禁用 / 1 启用 */
      status: number
      createdAt?: string
      updatedAt?: string
    }

    interface SocialClientList {
      list: SocialClientItem[]
      total: number
    }

    interface SocialClientSearchParams {
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface SocialClientSaveParams {
      name: string
      socialType: string
      clientId: string
      /** 编辑留空表示不修改 */
      clientSecret?: string
      agentId?: string
      status: number
    }

    /** 会员第三方账号绑定 */
    interface SocialBindingItem {
      id: string
      /** wechat_mini / wechat_mp / wechat_open */
      socialType: string
      nickname: string
      avatar: string
      createdAt: string
    }

    /** 会员等级（对齐 yudao member_level） */
    interface LevelItem {
      id: string
      /** 等级序号：1-99 */
      level: number
      name: string
      icon: string
      /** 升级所需经验值 */
      experience: number
      /** 折扣百分比：100 无折扣 / 95 表示 95 折 */
      discountPercent: number
      description: string
      /** 0 禁用 / 1 启用 */
      status: number
      sort: number
      createdAt?: string
      updatedAt?: string
    }

    /** 会员等级列表 */
    interface LevelList {
      list: LevelItem[]
      total: number
    }

    /** 会员等级搜索参数 */
    interface LevelSearchParams {
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 会员等级新增 / 编辑参数 */
    interface LevelSaveParams {
      level: number
      name: string
      icon?: string
      experience: number
      discountPercent: number
      description?: string
      status: number
      sort: number
    }

    /** 签到奖励档位（对齐 yudao member_sign_in_config） */
    interface SignConfigItem {
      id: string
      /** 连续签到第几天，从 1 开始 */
      day: number
      /** 该档位奖励积分 */
      point: number
      /** 该档位奖励经验 */
      experience: number
      /** 0 禁用 / 1 启用 */
      status: number
      createdAt?: string
      updatedAt?: string
    }

    /** 签到档位列表 */
    interface SignConfigList {
      list: SignConfigItem[]
      total: number
    }

    /** 签到档位搜索参数 */
    interface SignConfigSearchParams {
      /** 连续签到第几天 */
      day?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 签到档位新增 / 编辑参数 */
    interface SignConfigSaveParams {
      day: number
      point: number
      experience: number
      status: number
    }

    /** 积分记录（对齐 yudao member_point_record） */
    interface PointRecordItem {
      id: string
      userId: string
      /** 业务类型：0 其它 / 10 后台调整 / 20 签到 / 30 消费 */
      bizType: number
      bizId: string
      title: string
      description: string
      /** 积分，正数增加 / 负数扣减 */
      point: number
      /** 变动后的积分 */
      totalPoint: number
      creator?: string
      createdAt: string
    }

    /** 积分记录分页 */
    interface PointRecordList {
      list: PointRecordItem[]
      total: number
    }

    /** 积分调整参数 */
    interface PointAdjustParams {
      /** 正数增加 / 负数扣减 */
      point: number
      title?: string
      description?: string
    }

    /** 等级调整参数 */
    interface LevelAdjustParams {
      levelId: string
      experience: number
      title?: string
      description?: string
    }

    /** 等级变更记录 */
    interface LevelRecordItem {
      id: string
      userId: string
      levelId: string
      experience: number
      title: string
      description: string
      createdAt: string
    }

    /** 等级变更记录分页 */
    interface LevelRecordList {
      list: LevelRecordItem[]
      total: number
    }

    /** 签到记录 */
    interface SignRecordItem {
      id: string
      userId: string
      signDate: string
      point: number
      createdAt: string
    }

    /** 签到记录分页 */
    interface SignRecordList {
      list: SignRecordItem[]
      total: number
    }

    /** 收货地址 */
    interface AddressItem {
      id: string
      userId: string
      name: string
      mobile: string
      areaId: number
      detailAddress: string
      /** 0 否 / 1 是 */
      defaultStatus: number
    }

    /** 会员标签 */
    interface TagItem {
      id: string
      name: string
      status: number
      remark: string
      createdAt?: string
    }

    /** 标签列表 */
    interface TagList {
      list: TagItem[]
      total: number
    }

    /** 标签搜索参数 */
    interface TagSearchParams {
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 标签新增 / 编辑参数 */
    interface TagSaveParams {
      name: string
      status: number
      remark?: string
    }

    /** 会员分组 */
    interface GroupItem {
      id: string
      name: string
      status: number
      remark: string
      createdAt?: string
    }

    /** 分组列表 */
    interface GroupList {
      list: GroupItem[]
      total: number
    }

    /** 分组搜索参数 */
    interface GroupSearchParams {
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 分组新增 / 编辑参数 */
    interface GroupSaveParams {
      name: string
      status: number
      remark?: string
    }
  }

  /** 控制台统计 */
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

  /** 系统管理类型 */
  namespace SystemManage {
    /** 用户列表（与后端 Base.Page 对齐） */
    interface UserList {
      list: UserListItem[]
      total: number
    }

    /** 用户列表项（与后端 model.SystemUser 对齐，ID 统一为字符串） */
    interface UserListItem {
      id: string
      username: string
      nickname: string
      remark: string
      email: string | null
      mobile: string | null
      /** 0 未知 / 1 男 / 2 女 */
      sex: number
      avatar: string
      /** 0 禁用 / 1 启用 */
      status: number
      loginIp: string
      lastLoginAt: string | null
      creator: string
      createdAt: string
      updater: string
      updatedAt: string
    }

    /** 用户搜索参数（后端只支持这些筛选条件） */
    interface UserSearchParams {
      username?: string
      nickname?: string
      mobile?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑用户参数 */
    interface UserSaveParams {
      username: string
      /** 仅新增时必填，编辑不改密码 */
      password?: string
      nickname?: string
      remark?: string
      email?: string
      mobile?: string
      sex?: number
      avatar?: string
      status: number
    }

    /** 重置密码参数 */
    interface UserResetPasswordParams {
      password: string
    }

    /** 菜单树节点（与后端 model.SystemMenuModel 对齐） */
    interface MenuItem {
      id: string
      name: string
      /** 权限标识，仅按钮类填写 */
      permission: string
      /** 1 目录 / 2 菜单 / 3 按钮 */
      type: number
      sort: number
      parentId: string
      path: string
      icon: string
      component: string
      componentName: string
      /** 0 禁用 / 1 启用 */
      status: number
      /** 0 隐藏 / 1 显示 */
      visible: number
      keepAlive: number
      alwaysShow: number
      createdAt: string
      updatedAt: string
      children: MenuItem[]
    }

    /** 角色列表（与后端 Base.Page 对齐） */
    interface RoleList {
      list: RoleItem[]
      total: number
    }

    /** 角色项 */
    interface RoleItem {
      id: string
      name: string
      code: string
      sort: number
      /** 0 禁用 / 1 启用 */
      status: number
      /** 1 内置角色 / 2 自定义角色 */
      type: number
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 角色搜索参数 */
    interface RoleSearchParams {
      name?: string
      code?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑角色参数 */
    interface RoleSaveParams {
      name: string
      code: string
      sort?: number
      status: number
      remark?: string
    }

    /** 新增 / 编辑菜单参数 */
    interface MenuSaveParams {
      name: string
      permission?: string
      type: number
      sort?: number
      parentId?: string
      path?: string
      icon?: string
      component?: string
      componentName?: string
      status: number
      visible?: number
      keepAlive?: number
      alwaysShow?: number
    }

    /** 字典类型 */
    interface DictTypeItem {
      id: string
      name: string
      /** 字典类型编码，创建后不可修改 */
      type: string
      /** 0 禁用 / 1 启用 */
      status: number
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 字典类型列表 */
    interface DictTypeList {
      list: DictTypeItem[]
      total: number
    }

    /** 字典类型搜索参数 */
    interface DictTypeSearchParams {
      name?: string
      type?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑字典类型参数 */
    interface DictTypeSaveParams {
      name: string
      type: string
      status: number
      remark?: string
    }

    /** 字典数据 */
    interface DictDataItem {
      id: string
      sort: number
      label: string
      value: string
      dictType: string
      /** 0 禁用 / 1 启用 */
      status: number
      /** default / primary / success / info / warning / danger */
      colorType: string
      cssClass: string
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 字典数据列表 */
    interface DictDataList {
      list: DictDataItem[]
      total: number
    }

    /** 字典数据搜索参数 */
    interface DictDataSearchParams {
      dictType?: string
      label?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑字典数据参数 */
    interface DictDataSaveParams {
      sort?: number
      label: string
      value: string
      dictType: string
      status: number
      colorType?: string
      cssClass?: string
      remark?: string
    }

    /** 前端下拉 / 标签渲染用的字典项 */
    interface DictItem {
      label: string
      value: string
      colorType: string
      cssClass: string
      sort: number
    }

    /** 登录日志 */
    interface LoginLogItem {
      id: string
      userId: string
      /** 1 后台用户 / 2 会员 */
      userType: number
      username: string
      /** 1 登录 / 2 登出 */
      type: number
      /** 0 失败 / 1 成功 */
      result: number
      userIp: string
      userAgent: string
      createdAt: string
    }

    /** 登录日志列表 */
    interface LoginLogList {
      list: LoginLogItem[]
      total: number
    }

    /** 登录日志搜索参数 */
    interface LoginLogSearchParams {
      username?: string
      type?: number
      result?: number
      pageNo?: number
      pageSize?: number
    }

    /** 操作日志 */
    interface OperateLogItem {
      id: string
      userId: string
      /** 1 后台用户 / 2 会员 */
      userType: number
      /** 操作类型（模块） */
      type: string
      /** 操作子类型（动作） */
      subType: string
      action: string
      /** 0 失败 / 1 成功 */
      success: number
      extra: string
      requestMethod: string
      requestUrl: string
      requestIp: string
      userAgent: string
      /** 耗时（毫秒） */
      duration: number
      createdAt: string
    }

    /** 操作日志列表 */
    interface OperateLogList {
      list: OperateLogItem[]
      total: number
    }

    /** 操作日志搜索参数 */
    interface OperateLogSearchParams {
      type?: string
      subType?: string
      keyword?: string
      success?: number
      pageNo?: number
      pageSize?: number
    }

    /** 角色列表 */
    type RoleList = Api.Common.PaginatedResponse<RoleListItem>

    /** 角色列表项 */
    interface RoleListItem {
      roleId: number
      roleName: string
      roleCode: string
      description: string
      enabled: boolean
      createTime: string
    }

    /** 角色搜索参数 */
    type RoleSearchParams = Partial<
      Pick<RoleListItem, 'roleId' | 'roleName' | 'roleCode' | 'description' | 'enabled'> &
        Api.Common.CommonSearchParams
    >
  }
}
