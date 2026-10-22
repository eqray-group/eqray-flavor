/**
 * API 类型：会员档案、等级、标签、分组、积分、签到、地址与社交登录
 * 全局命名空间 Api 的合并声明，使用处无需 import。
 */

declare namespace Api {
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
}
