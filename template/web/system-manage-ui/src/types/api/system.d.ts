/**
 * API 类型：用户、角色、菜单、字典与日志（命名空间仍为 SystemManage）
 * 全局命名空间 Api 的合并声明，使用处无需 import。
 */

declare namespace Api {
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
