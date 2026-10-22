# 项目框架设计文档

> 技术栈：Goravel 1.18 + Gin + MySQL 8.0
> 表结构参考 **yudao（ruoyi-vue-pro）**，字段命名遵循 Goravel 约定。
> 文档性质：架构与数据设计说明；除第 15 节（可直接复制的代码模板）外，其余章节不含实现代码。

---

## 1. 设计目标

1. **模块清晰**：分为基础模块、系统关联模块、基础设施模块、会员模块，边界明确。
2. **依赖单向**：三个业务模块都只依赖 base，彼此互不依赖。
3. **分层固定**：业务模块内部统一为 `Controller / Service / Repository / Model / Request / Route` 六层；base 作为底座不对外提供路由与控制器，只含 Service / Repository / Model / Request 及各类基础封装。
4. **表结构有参照**：数据表划分与字段构成对齐 yudao，降低团队认知成本，便于对照与迁移。
5. **约定优先**：响应体、错误码、表名、权限标识、迁移命名全部有统一约定。

---

## 2. 技术栈与依赖变更

数据库由 Postgres 调整为 **MySQL 8.0**，需要同步调整以下配置：

| 位置 | 变更内容 |
|---|---|
| `go.mod` | 移除 `github.com/goravel/postgres`，引入 `github.com/goravel/mysql` |
| `config/database.go` | 连接驱动改为 `mysqlfacades.Mysql("mysql")`，连接名改为 `mysql` |
| `config/auth.go` | 守卫由默认 `user` 改为 `admin` 与 `member` 两个独立 guard，各配 provider |
| `.env` | `DB_CONNECTION=mysql`，补齐 `DB_HOST` / `DB_PORT=3306` / `DB_DATABASE` / `DB_USERNAME` / `DB_PASSWORD` |

数据库层约束：

- 引擎 `InnoDB`，字符集 `utf8mb4`，排序规则 `utf8mb4_unicode_ci`。
- 时间字段统一 `datetime`，由应用层写入，不用 `timestamp`。
- 金额字段统一 `bigint`，以「分」为单位存储（如 12.34 元存 1234），禁止用 `float`/`double`/`decimal`；接口出入参同样以「分」传递，由前端负责展现换算。折扣率、费率等比例字段不属于金额，仍用 `decimal` 保留精度。
- 状态类字段统一 `tinyint`（1 启用 / 0 禁用）。
- 所有 `varchar` 必须显式指定长度。
- 主键由应用侧生成雪花 ID，建表时不使用 `AUTO_INCREMENT`，类型为有符号 `bigint`。

---

## 3. 模块划分与依赖

| 模块 | 定位 | 职责范围 | 依赖 |
|---|---|---|---|
| **base 基础模块** | 技术底座 | 统一响应封装、系统错误码与业务异常基础设施、通用中间件、认证组件（双令牌签发与校验、登录日志）、分页/查询请求基类、通用工具（树构建、雪花 ID、加解密、导出） | 无 |
| **system 系统关联模块** | 后台管理与系统关联关系 | 系统用户（员工账号）、认证登录、角色、菜单权限点、字典、通知公告、操作日志，以及各关联表 | base |
| **infra 基础设施模块** | 与业务无关的基础设施 | 参数配置、文件与文件配置、定时任务与任务日志、API 访问与错误日志 | base |
| **member 会员模块** | 面向 C 端的会员体系 | 会员账号、注册与登录、会员资料、会员等级与等级记录、积分与积分记录、签到、标签、分组 | base |

### 依赖关系

```
             base
        ╱     │     ╲
   system   infra   member
```

- `base` 不含业务表，只持有一张认证日志表（`auth_login_log`），被三个模块共同依赖。
- `system`、`infra`、`member` **除下列场景外互不依赖**，各自维护自己的业务表：
  - `member` 的短信验证码登录依赖 `infra` 的 `InfraSmsService`（渠道 / 模板配置与 SDK 发送都在 infra），
    与 yudao 中 member 依赖 system 的短信能力一致；`system` 不依赖任何业务模块。
- 后台「系统用户」与 C 端「会员」是两套独立账号体系，不共用业务表；共用 base 的认证组件，登录日志以 `user_type` 区分来源。

### 表归属

| 归属 | 表 |
|---|---|
| base 模块 | `auth_login_log` |
| system 模块 | `system_users`、`system_role`、`system_menu`、`system_user_role`、`system_role_menu`、`system_dict_type`、`system_dict_data`、`system_notice`、`system_operate_log` |
| infra 模块 | `infra_config`、`infra_file`、`infra_file_config`、`infra_sms_channel`、`infra_sms_template`、`infra_api_access_log`、`infra_api_error_log` |
| member 模块 | `member_user`、`member_level`、`member_level_record`、`member_point_record`、`member_sign_record`、`member_tag`、`member_group`、`member_address`、`member_social_client`、`member_social_user`、`member_social_binding` |

> 原则：**谁的表谁来读写**。system 不查 `infra_*` 与 `member_*`，member 不查 `system_*`。`auth_login_log` 由 base 的认证组件独占读写，业务模块只能调用其 Service。

> 表前缀由早期约定的 `sys_` 调整为 `system_` / `infra_` / `member_`，与 yudao 保持一致，便于直接对照其建表语句与文档。认证日志表因归属 base，改用 `auth_` 前缀以表明其框架级身份。

---

## 4. 目录结构

```
app/
├── facades/                       # 框架门面（保持现有不动）
├── modules/
│   ├── base/                      # 基础模块
│   │   ├── types/                 # 通用类型：ID（雪花 int64，JSON 序列化为字符串）
│   │   ├── model/                 # Model 基类：审计字段与 FillCreate / FillUpdate 填充
│   │   ├── auth/                  # 认证组件（base 唯一持有表的组件）
│   │   │   ├── model/             # auth_login_log
│   │   │   ├── repository/
│   │   │   └── service/           # 双令牌签发与校验、登录日志写入
│   │   ├── enums/                 # 跨模块共用枚举（如 user_type）
│   │   ├── errs/                  # 系统错误码定义 + 业务异常类型 + 请求上下文错误槽（Capture / Captured）
│   │   ├── response/              # code/msg/data 响应封装（Success / Page / Fail）
│   │   ├── param/                 # 通用请求参数：分页 Page、Bind / Validate / PathID
│   │   ├── result/                # 响应输出助手（Success / Page / Fail）
│   │   ├── middleware/            # recovery / exception_handler（含错误槽）/ cors / request_log / token_auth / member_auth / permission
│   │   └── support/               # 树构建、雪花ID（节点号取 SNOWFLAKE_NODE_ID）、1062 判重、加解密、导出
│   ├── system/                    # 系统关联模块（仅管理端）
│   │   ├── admin/{controller,request}/   # 前缀 /api/v1/system
│   │   ├── enums/ errs/ model/ repository/ service/
│   │   ├── middleware/            # operate_log（按路由元数据采集操作日志）
│   │   └── route.go               # 聚合：admin.Register()
│   ├── infra/                     # 基础设施模块
│   │   ├── admin/{controller,request}/   # 前缀 /api/v1/infra
│   │   ├── client/{controller,request}/  # 前缀 /api/v1/client/infra
│   │   ├── enums/ errs/ model/ repository/ service/
│   │   ├── middleware/            # api_access_log（落库）、api_error_log
│   │   ├── job/                   # 定时任务处理器
│   │   └── route.go               # 聚合：admin.Register() + client.Register()
│   └── member/                    # 会员模块
│       ├── admin/{controller,request}/   # 前缀 /api/v1/member
│       ├── client/{controller,request}/  # 前缀 /api/v1/client/member
│       ├── enums/ errs/ model/ repository/ service/
│       └── route.go               # 聚合：admin.Register() + client.Register()
routes/
├── web.go                         # 保留欢迎页
├── grpc.go                        # 保留
└── modules.go                     # 聚合：system / infra / member 的 Register()
database/
└── migrations/
config/                            # database.go 改 MySQL、auth.go 加双 guard
```

`bootstrap/app.go` 的 `WithRouting` 中追加 `routes.Modules()` 即可挂载全部模块路由。

---

## 5. 分层规范

### 5.1 调用链

```
HTTP 请求
  → Request     参数绑定与校验
  → Controller  编排（取参 → 调 Service → 出响应），不含业务判断
  → Service     业务逻辑、事务边界、跨模块调用
  → Repository  数据访问，只操作本模块表，返回 Model 或分页结果
  → Model       表映射（字段、表名、db tag）
```

### 5.2 各层约束

| 层 | 允许做 | 禁止做 |
|---|---|---|
| Controller | 调用 Request 校验、调用 Service、返回统一响应 | 直接访问 Repository、写业务分支、开事务 |
| Service | 业务规则、事务、调用其他模块 Service、写日志 | 直接操作 `http.Context`、拼接 SQL |
| Repository | 单表/多表查询、条件构建、返回结构体 | 业务判断、返回 HTTP 结构、开事务 |
| Model | 字段与 `db` tag、表名（`TableName()`）、`FillCreate` / `FillUpdate` | 写查询逻辑、声明 ORM 关联 |

事务只允许在 Service 层开启，Repository 内不出现事务代码。

### 5.3 Repository 层约定

四层结构中 Repository 为强制层，为避免其退化成纯透传的样板代码，约定如下：

| 项 | 约定 |
|---|---|
| 粒度 | 每个实体一个 Repository，放在模块的 `repository/` 下 |
| 方法命名 | 按业务查询语义命名，如 `FindByUsername`、`PageWithDataScope`、`ListByRoleIDs` |
| 职责 | 承载多表关联、动态条件组装、分页查询；`First by id` 这类裸查询不必单独成方法 |
| 表名 | 引用 Model 导出的 `XxxTable` 常量，不在 Repository 里重复写表名字面量 |
| 唯一键 | 写入捕获 1062（`support.IsDuplicate`），转成与该字段「已存在」相同的业务错误码（6.8(5)） |
| 分页 | 只处理 1 ~ 200 的入参（校验在 Controller）；软删除用 `RowsAffected` 判定未命中 |
| 横切条件 | 软删除等条件在 Repository 入口统一注入，不散落到 Service |
| 事务 | 自身不开事务，由 Service 控制并向下传递 |
| 返回 | 返回 Model 或分页结果，不返回 HTTP 结构、不返回 Request / Response 对象 |
| 接口 | 不为每个 Repository 预抽 interface；仅当 Service 测试需要 mock 时，由使用方就近定义小接口 |
| 跨模块 | 禁止被其他模块 import（见第 10 节） |

反模式：为一个查询形态固定不变的实体，写一整套 `FindByXxx` 的逐字段 getter。方法应对应**真实的业务查询**，而不是把查询构造器已有的能力再枚举一遍。

> 判断标准很简单：如果这个方法的方法体只有一行、且只是把入参原样传给查询构造器，那它不该存在——除非它是为了承载数据权限注入等横切逻辑。

---

## 6. 数据库设计

### 6.1 通用字段约定

对齐 yudao 的 `BaseDO`，但命名遵循 Goravel 约定：

| yudao 字段 | 本项目字段 | 类型 | 说明 |
|---|---|---|---|
| `creator` | `creator` | varchar(64) | 创建者（存用户 ID） |
| `create_time` | `created_at` | datetime | 创建时间 |
| `updater` | `updater` | varchar(64) | 更新者（存用户 ID） |
| `update_time` | `updated_at` | datetime | 更新时间 |
| `deleted` | `deleted_at` | datetime NULL | 软删除时间，NULL 表示未删除 |

其他公共约定：

- 主键统一 `id bigint`，**不使用数据库自增**，由 base 的 `support` 以雪花算法在应用侧生成。雪花 ID 为 int64（19 位），所有 ID 字段对外传递时必须序列化为字符串，见 8.4。
- 除特别说明外，**所有表**（含关联表与日志表）都带上述 5 个审计字段，与 yudao 一致。
- `status` 取值 `1` 启用 / `0` 禁用（与 yudao 的 0/1 相反，见 6.6 差异说明）。
- 跨表引用只存 ID，**不建物理外键**，由 Service 层保证一致性。
- `creator` / `updater` 由 base 的 Model 基类统一填充，业务代码不手工赋值：Repository 写入前调用 `FillCreate`（写 `id`、`creator`、`created_at`，并把 `updated_at` 置为同一时刻）与 `FillUpdate`（写 `updater`、`updated_at`）。查询构造器不触发框架钩子，因此这两个方法是**显式调用**，不是自动回调（见 15.3）。当前登录者从 base 的上下文读取；无登录态的场景（定时任务、数据初始化）留空。**审计字段只是「业务不显式写入」的一个例子，完整的自动写入清单见 6.9。**

### 6.2 base 模块表

base 只持有一张认证日志表。令牌采用**双令牌 + Redis** 方案，状态存 Redis、可主动失效（见 9.2）。

#### auth_login_log 登录日志
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| user_id | bigint | 用户 ID，可空 |
| user_type | tinyint | 用户类型（1 后台用户 / 2 会员） |
| username | varchar(64) | 登录账号 |
| type | tinyint | 1 登录 / 2 登出 |
| result | tinyint | 1 成功 / 0 失败 |
| user_ip | varchar(50) | 来源 IP |
| user_agent | varchar(512) | 客户端信息 |

索引：`idx_user_id(user_id)`、`idx_username(username)`、`idx_created_at(created_at)`。

> 本表由 base 的认证组件独占读写，业务模块只能调用其 Service，不直接操作表。

### 6.3 system 模块表

#### system_users 用户
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| username | varchar(64) | 登录名，唯一 |
| password | varchar(255) | bcrypt 哈希 |
| nickname | varchar(64) | 昵称 |
| remark | varchar(500) | 备注 |
| email | varchar(128) | 邮箱 |
| mobile | varchar(20) | 手机号 |
| sex | tinyint | 0 未知 / 1 男 / 2 女 |
| avatar | varchar(512) | 头像地址 |
| status | tinyint | 状态 |
| login_ip | varchar(50) | 最后登录 IP |
| last_login_at | datetime | 最后登录时间 |
| creator / created_at / updater / updated_at / deleted_at | — | 审计字段 |

索引：`uk_username(username)`、`uk_email(email)`、`uk_mobile(mobile)`。

> 简单系统不引入部门与岗位：`system_users` 只保留账号本身的信息，人员组织关系交给外部系统或后续按需扩展。

#### system_role 角色
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| name | varchar(64) | 角色名称 |
| code | varchar(64) | 角色标识，唯一，如 `SUPER_ADMIN` |
| sort | int | 排序 |
| status | tinyint | 状态 |
| type | tinyint | 1 内置角色 / 2 自定义角色 |
| remark | varchar(500) | 备注 |

索引：`uk_code(code)`。

#### system_menu 菜单与权限点
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| name | varchar(64) | 菜单名称 |
| permission | varchar(128) | 权限标识，如 `system:user:create` |
| type | tinyint | 1 目录 / 2 菜单 / 3 按钮 |
| sort | int | 排序 |
| parent_id | bigint | 父菜单 ID，0 为顶级 |
| path | varchar(255) | 前端路由地址 |
| icon | varchar(64) | 图标 |
| component | varchar(255) | 前端组件路径 |
| component_name | varchar(64) | 组件名称 |
| status | tinyint | 状态 |
| visible | tinyint | 是否显示 |
| keep_alive | tinyint | 是否缓存 |
| always_show | tinyint | 是否始终显示 |

索引：`idx_permission(permission)`、`idx_parent_id(parent_id)`。

> `permission` 不做唯一约束，与 yudao 保持一致：只有按钮类节点才填写该值，目录与菜单类节点留空；唯一性由业务层与菜单管理界面保证，不依赖数据库约束。

#### system_user_role 用户角色关联
`id`、`user_id`、`role_id` + 审计字段。索引：`uk_user_role(user_id, role_id)`。

#### system_role_menu 角色菜单关联
`id`、`role_id`、`menu_id` + 审计字段。索引：`uk_role_menu(role_id, menu_id)`。

#### system_dict_type 字典类型
`id`、`name`、`type`（字典类型编码）、`status`、`remark` + 审计字段。索引：`uk_type(type)`。

#### system_dict_data 字典数据
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| sort | int | 排序 |
| label | varchar(64) | 显示文本 |
| value | varchar(64) | 实际值 |
| dict_type | varchar(64) | 所属字典类型编码 |
| status | tinyint | 状态 |
| color_type | varchar(32) | 标签样式（default/primary/success/info/warning/danger） |
| css_class | varchar(64) | 自定义样式类 |
| remark | varchar(500) | 备注 |

索引：`uk_type_value(dict_type, value)`、`idx_dict_type(dict_type)`。

#### system_notice 通知公告
`id`、`title`(varchar 128)、`type`(tinyint，1 通知 / 2 公告)、`content`(text)、`status` + 审计字段。

#### system_operate_log 操作日志
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| trace_id | varchar(64) | 链路追踪 ID |
| user_id | bigint | 操作人 ID |
| user_type | tinyint | 用户类型 |
| type | varchar(64) | 操作类型 |
| sub_type | varchar(64) | 操作子类型 |
| biz_id | varchar(64) | 业务主键 |
| action | varchar(255) | 操作描述 |
| success | tinyint | 是否成功 |
| extra | text | 扩展信息（脱敏） |
| request_method | varchar(16) | HTTP 方法 |
| request_url | varchar(512) | 请求地址 |
| user_ip | varchar(50) | 来源 IP |
| user_agent | varchar(512) | 客户端信息 |

索引：`idx_user_id(user_id)`、`idx_trace_id(trace_id)`、`idx_created_at(created_at)`。

> 与 `infra_api_access_log` 的分工：本表记录**有业务语义的操作**（如「新增用户」「分配角色」），由 system 模块的中间件按**路由元数据**自动采集，业务代码不显式调用，且只在后台管理端产生；`infra_api_access_log` 记录**每个请求**的出入参，覆盖面更广。两者并存，前者偏审计、后者偏排查。本表与 `auth_login_log`、`infra_api_access_log`、`infra_api_error_log` 同属「业务不显式写入」的数据，统一约定见 6.9。

### 6.4 infra 模块表

#### infra_config 参数配置
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| category | varchar(64) | 参数分组 |
| name | varchar(128) | 参数名称 |
| config_key | varchar(128) | 参数键，唯一 |
| value | varchar(512) | 参数值 |
| type | tinyint | 1 系统内置 / 2 自定义 |
| visible | tinyint | 是否对外公开 |
| remark | varchar(500) | 备注 |

索引：`uk_config_key(config_key)`。

> 参数值读取走 Cache，通过 base 的缓存工具统一封装。

#### infra_file 文件
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| config_id | bigint | 文件配置 ID |
| name | varchar(255) | 原文件名 |
| path | varchar(512) | 存储路径 |
| url | varchar(1024) | 访问地址 |
| type | varchar(64) | 文件 MIME 类型 |
| size | bigint | 文件大小（字节） |

索引：`idx_config_id(config_id)`。

#### infra_file_config 文件配置
`id`、`name`(varchar 64)、`storage`(varchar 32，存储类型)、`master`(tinyint，是否主配置)、`config`(text，存储配置 JSON) + 审计字段。

#### infra_job 定时任务
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| name | varchar(64) | 任务名称 |
| status | tinyint | 状态 |
| handler_name | varchar(255) | 处理器名称（对应 Go 注册的 key） |
| handler_param | varchar(255) | 处理器参数 |
| cron_expression | varchar(255) | cron 表达式 |
| retry_count | int | 最大重试次数 |
| retry_interval | int | 重试间隔（毫秒） |
| monitor_timeout | int | 监控超时（毫秒） |

#### infra_job_log 定时任务日志
`id`、`job_id`、`handler_name`、`handler_param`、`execute_index`(第几次执行)、`begin_at`、`end_at`、`duration`(耗时毫秒)、`status`(执行状态)、`result`(text，结果) + 审计字段。索引：`idx_job_id(job_id)`。

#### infra_api_access_log API 访问日志
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| trace_id | varchar(64) | 链路追踪 ID |
| user_id | bigint | 用户 ID |
| user_type | tinyint | 用户类型 |
| application_name | varchar(64) | 应用名 |
| request_method | varchar(16) | HTTP 方法 |
| request_url | varchar(512) | 请求地址 |
| request_params | text | 请求参数（脱敏） |
| response_body | text | 响应体（脱敏、可截断） |
| user_ip | varchar(50) | 来源 IP |
| user_agent | varchar(512) | 客户端信息 |
| operate_module | varchar(128) | 操作模块 |
| operate_name | varchar(128) | 操作名 |
| begin_at | datetime | 开始时间 |
| end_at | datetime | 结束时间 |
| duration | int | 耗时（毫秒） |
| result_code | int | 业务响应码 |

索引：`idx_trace_id(trace_id)`、`idx_user_id(user_id)`、`idx_created_at(created_at)`。

#### infra_api_error_log API 错误日志
`id`、`trace_id`、`user_id`、`user_type`、`application_name`、`request_method`、`request_url`、`request_params`、`user_ip`、`user_agent`、`name`(异常类名)、`message`(text)、`exception_stack`(text)、`process_status`(处理状态)、`process_time` + 审计字段。

### 6.5 member 模块表

#### member_user 会员
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| mobile | varchar(20) | 手机号，唯一 |
| password | varchar(255) | bcrypt 哈希 |
| status | tinyint | 状态 |
| nickname | varchar(64) | 昵称 |
| avatar | varchar(512) | 头像地址 |
| name | varchar(64) | 真实姓名 |
| sex | tinyint | 0 未知 / 1 男 / 2 女 |
| birthday | date | 生日 |
| area_id | int | 地区编码 |
| remark | varchar(500) | 备注 |
| point | int | 当前积分 |
| tag_ids | varchar(255) | 标签 ID 集合，逗号分隔 |
| level_id | bigint | 会员等级 ID |
| experience | int | 经验值 |
| login_ip | varchar(50) | 最后登录 IP |
| last_login_at | datetime | 最后登录时间 |

索引：`uk_mobile(mobile)`、`idx_level_id(level_id)`、`idx_group_id(group_id)`。

#### member_level 会员等级
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| name | varchar(64) | 等级名称，如「黄金会员」 |
| level | int | 等级值，越大越高 |
| experience | int | 升级所需经验值 |
| icon | varchar(512) | 等级图标 |
| background_url | varchar(512) | 等级背景图 |
| discount_percent | int | 折扣百分比，如 80 表示 8 折（比例字段，非金额） |
| status | tinyint | 状态 |

索引：`uk_level(level)`。

#### member_level_record 会员等级记录
`id`、`user_id`、`level_id`、`level`(冗余等级值)、`experience`、`remark` + 审计字段。

#### member_point_record 会员积分记录
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| user_id | bigint | 会员 ID |
| biz_id | varchar(64) | 业务编号 |
| biz_type | varchar(64) | 业务类型 |
| title | varchar(128) | 积分标题 |
| description | varchar(500) | 积分描述 |
| point | int | 变动积分（正加负减） |
| total_point | int | 变动后总积分 |

索引：`idx_user_id(user_id)`。

#### member_sign_record 会员签到记录
`id`、`user_id`、`day`(int，签到日期 yyyyMMdd)、`point`(赠送积分)、`experience`(赠送经验) + 审计字段。索引：`uk_user_day(user_id, day)`。

#### member_tag 会员标签
`id`、`name`(varchar 64) + 审计字段。

#### member_group 会员分组
`id`、`name`(varchar 64)、`remark`、`status` + 审计字段。

#### member_address 会员收货地址
`id`、`user_id`、`name`(收件人)、`mobile`、`area_id`、`detail_address`、`default_status` + 审计字段。索引：`idx_user(user_id)`。

#### member_social_client 会员第三方登录渠道配置
对齐 yudao 的 `system_social_client`，由后台页面管理。
`id`、`name`、`social_type`、`client_id`(微信 appid)、`client_secret`(微信 appsecret)、`agent_id`、`status` + 审计字段。
索引：`uk_social_type(social_type)`。`client_secret` 列表不返回，编辑留空表示不修改。

#### member_social_user 会员第三方账号
`id`、`type`(渠道：wechat_mini / wechat_mp / wechat_open)、`openid`、`unionid`、`nickname`、`avatar`、`raw_data` + 审计字段。
索引：`uk_type_openid(type, openid)`、`idx_unionid(unionid)`。

#### infra_sms_channel 短信渠道
对齐 yudao 的 `system_sms_channel`：渠道凭据存库、后台页面维护，属基础设施能力。
`id`、`name`、`code`(DebugLog / AliyunSms / TencentSms)、`api_key`、`api_secret`、`signature`、`extra`(JSON，腾讯云 region / sdkAppId)、`status`、`remark` + 审计字段。索引：`uk_code(code)`。

#### infra_sms_template 短信模板
对齐 yudao 的 `system_sms_template`。`id`、`channel_id`、`code`、`name`、`type`(1 验证码 / 2 通知)、`api_template_id`、`content`(`{name}` 占位)、`params`、`status`、`remark` + 审计字段。索引：`uk_code(code)`、`idx_channel(channel_id)`。

#### member_social_binding 会员第三方绑定
`id`、`user_id`、`social_type`、`social_user_id` + 审计字段。
索引：`uk_social_user(social_user_id)`、`uk_user_type(user_id, social_type)`；解绑物理删除，允许重新绑定。

### 6.6 与 yudao 的差异说明

| 项 | yudao | 本项目 | 原因 |
|---|---|---|---|
| 审计时间字段 | `create_time` / `update_time` | `created_at` / `updated_at` | 沿用 Goravel 命名习惯；由 `base/model.BaseModel` 统一声明，不用 `orm.Model` |
| 逻辑删除 | `deleted` bit(1) | `deleted_at` datetime | 沿用 Goravel 约定字段，由 Repository 显式过滤与写入（15.3）；唯一键冲突处理见 6.8(2) |
| 创建/更新人 | `creator` / `updater` | 同 yudao | 直接沿用 |
| `status` 取值 | 0 开启 / 1 关闭 | 1 启用 / 0 禁用 | 更符合直觉，业务代码可读性更好 |
| 主键 | `bigint AUTO_INCREMENT` | 雪花 ID（`bigint`，应用侧生成） | 不暴露业务量、天然支持分库分表；代价是需处理字符串序列化与时钟回拨 |
| 表前缀 | `system_` / `infra_` / `member_` | 同 yudao | 便于直接对照 |
| 金额单位 | 分（Integer/Long） | bigint 分 | 一致 |
| 认证模式 | OAuth2，access + refresh 令牌落库，支持强制下线与多端管理 | 双令牌 + Redis，access 短效 / refresh 长效 | 不建会话表，令牌状态放 Redis；每次请求一次 Redis 查询，换取登出即失效与多端独立 |
| 认证表 | `system_oauth2_*`、`system_login_log` 归 system 模块 | 只保留 `auth_login_log`（归 base），令牌表取消 | 令牌状态落 Redis，登录日志仍统一留存 |

### 6.7 迁移文件规范

- 命名：`YYYYMMDDHHMMSS_snake_case.go`，**文件名必须与 `Signature()` 返回的字符串一致**。
- **一个表对应一个 `create_*_table.go`**：只建这一张表，且直接建出最终结构（含索引与注释），
  不放种子数据，也不把多张表塞进同一个文件；表结构变更通过新增迁移完成。
- **数据填充单独放 `seed_*.go`**：菜单、初始账号、默认等级、默认短信渠道等各自独立成文件，
  内容必须幂等（先查再插）。有依赖的种子用编号保证顺序（账号 → 角色 → 授权 → 菜单）。
- 建表顺序（编号即顺序）：
  1. **base**：`auth_login_log`
  2. **system**：`system_users` → `system_menu` → `system_role` → `system_user_role` → `system_role_menu` → `system_dict_type` → `system_dict_data` → `system_notice` → `system_operate_log`
  3. **infra**：`infra_config` → `infra_file_config` → `infra_file` → `infra_sms_channel` → `infra_sms_template` → `infra_api_access_log` → `infra_api_error_log`
  4. **member**：`member_level` → `member_user` → `member_level_record` → `member_point_record` → `member_sign_record` → `member_tag` → `member_group` → `member_address` → `member_social_client` → `member_social_user` → `member_social_binding`
- 每个迁移的 `Down()` 必须实现 `DropIfExists`，保证可回滚。
- 表名在迁移中显式指定，同时把 Model 的表名方法写死，避免复数化推导歧义。
- 建表必须写**表注释与列注释**：表注释与本文档的表标题一致，列注释取自 6.x 的字段说明（状态类字段注明取值含义）。Schema Builder 用 `table.Comment("后台用户表")` / `table.String("nickname", 64).Comment("昵称")`，原生 SQL 用 `COMMENT='...'`。`SHOW CREATE TABLE` 与 DBA 排查都靠它。
- 日志类表（`*_log`）数据量大，建议建表时即按 `created_at` 建索引，并规划定期归档。
- 唯一键直接建在业务键上，不使用 `deleted_key` 生成列；关联表解除关系时用物理删除，避免占用唯一键（6.8(4)、SQL 写法见 15.9）。
- 迁移文件删除或改名后，同步删除 `migrations` 表中对应执行记录；`bootstrap/migrations.go` 必须手工注册每一个迁移。

### 6.8 唯一约束约定

唯一性必须由数据库兜底：Service 层的「先查后插」存在并发窗口，只能用作友好提示，不能作为唯一性保证。本节集中说明各表的唯一键策略，6.2 ~ 6.5 各表小节里列出的 `uk_*` 即最终形态，不再追加 `deleted_key`。

#### （1）命名、列顺序与归一化

| 项 | 约定 |
|---|---|
| 命名 | `uk_` 前缀 + 语义缩写：`uk_username`、`uk_user_role`、`uk_type_value` |
| 列顺序 | 多列唯一键按「区分度高 → 低」排列，便于复用索引前缀做等值查询 |
| 大小写 | 唯一键涉及的 varchar 列沿用表的 `utf8mb4_unicode_ci`，唯一性**天然大小写不敏感**；Service 查重直接用 `=`，不允许改写成 `_bin` 或手工 `LOWER()` |
| 归一化 | 写入前归一化：手机号去空格与 `+86`、账号与邮箱转小写、编码去首尾空格。唯一性是「归一化之后的值」的唯一性 |
| 可空字段 | `system_users.email` / `mobile` 这类可选唯一字段用 **NULL** 表示未填写，**禁止用空字符串**——唯一索引不比较 NULL，但会把两个 `''` 判为重复，导致第二个用户写不进去；Model 中对应字段必须声明为 `*string`，否则查询时 `converting NULL to string is unsupported` |

#### （2）软删除与唯一索引的冲突

问题：软删除的行仍在表里，唯一索引被旧值永久占位，删掉 `zhangsan` 后再建同名账号会撞 1062；而 MySQL **没有 `WHERE deleted_at IS NULL` 的部分索引**，无法只对活跃行建唯一约束。

本项目统一选择**不让位**：唯一索引只含业务键，软删除后该值永久占用，不使用 `deleted_key` 生成列，也不在删除时改写业务键。

- `system_users`：后台账号删除等于作废，账号名不复用反而能避免审计混淆。
- `system_role.code`：角色删除后标识不复用；`ExistsByCode` 查重不过滤 `deleted_at`，与唯一键行为保持一致（否则 Service 检查通过、数据库却撞 1062）。
- 需要「删掉再建同名」的场景，说明这条记录本就不该软删除，应物理删除或换业务键。
- 关联表不参与软删除，见 (4)。

唯一键直接建在业务键上：

```sql
CREATE TABLE system_users (
  id           bigint      NOT NULL,
  username     varchar(64) NOT NULL,
  -- …其他业务字段
  deleted_at   datetime    NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台用户表';
```

#### （3）事实记录与日志表

一次性事实记录删除后仍必须占位，唯一索引只含业务键：

| 表 | 唯一键 | 为什么不允许复用 |
|---|---|---|
| `member_sign_record` | `uk_user_day(user_id, day)` | 签到是「一天一次」的事实，删掉记录不应等于「可以再签一次」 |

日志类表（`auth_login_log`、`system_operate_log`、`infra_job_log`、`infra_api_access_log`、`infra_api_error_log`）属于另一类：它们**不建任何唯一键**，只追加不删除（也不做软删除），治理靠归档任务而非删除。

#### （4）关联表的复合唯一

关联表的唯一性表达的是「同一组合只允许一条关系」，删除语义是「解除关系」。唯一键不含删除时间，因此关联表**不做软删除**，解除关系时直接物理删除，才能反复分配：

| 表 | 唯一键 |
|---|---|
| `system_user_role` | `uk_user_role(user_id, role_id)` |
| `system_role_menu` | `uk_role_menu(role_id, menu_id)` |
| `member_tag_ref` | `uk_user_tag(user_id, tag_id)` |

反例：关联表若沿用软删除，`uk_user_role(user_id, role_id)` 会和历史行冲突，「给用户重新分配之前解除的角色」直接撞 1062。因此 `Sync` 的做法是「多余的直接 `Delete`、缺少的 `Insert`」，删除角色时 `DeleteByRole` 也走物理删除。

#### （5）应用侧必须处理 1062

预查重不能替代唯一索引，唯一索引也不能替代友好文案，两者都要有：

- 前置校验：Service 先查重，命中即返回对应业务错误码（`errs.SystemUsernameExists` 之类）。
- 落库兜底：Repository 写入/更新时捕获 MySQL `1062 Duplicate entry`，转换成**同一个**业务错误码抛出；禁止把驱动原始错误透传给前端。判重 helper 的建议实现见 15.3。
- 二者返回同一个错误码，前端无需区分「已存在」与「并发撞键」。

#### （6）全量清单

| 表 | 唯一键 | 备注 |
|---|---|---|
| `system_users` | `uk_username`、`uk_email`、`uk_mobile` | 账号删除后不复用，唯一键直接建在业务键上；`email`、`mobile` 用 NULL 表示未填 |
| `system_role` | `uk_code` | 删除后标识不复用；内置角色 `SUPER_ADMIN` 由初始化写入，不允许删除 |
| `system_menu` | 无 | `permission` 仅按钮类节点填写，唯一性由业务层保证（见 6.3） |
| `system_user_role` / `system_role_menu` | 见 (4) | 复合唯一，解除关系时物理删除 |
| `system_dict_type` | `uk_type` | 删除后标识不复用 |
| `system_dict_data` | `uk_type_value(dict_type, value)` | 删除后键值不复用 |
| `infra_config` | `uk_config_key` | `config_key` 由代码常量引用，改名等于换键 |
| `infra_file_config` / `infra_file` / `infra_job` / `infra_job_log` | 无 | 「至多一个主文件配置」由 Service 保证；任务名允许重复，用 `id` 区分 |
| `member_user` | `uk_mobile` | 手机号即账号，删除后不复用 |
| `member_level` | `uk_level` | 删除等级前需确认无会员占用，属于 Service 职责 |
| `member_sign_record` | `uk_user_day(user_id, day)` | 见 (3) |
| `member_social_client` | `uk_social_type(social_type)` | 每渠道一条配置 |
| `member_social_user` | `uk_type_openid(type, openid)` | 同一渠道下 openid 唯一 |
| `member_social_binding` | `uk_social_user(social_user_id)` / `uk_user_type(user_id, social_type)` | 见 (4)，解绑物理删除 |
| `member_tag` / `member_group` / `member_address` / `system_notice` / `auth_login_log` / `system_operate_log` / `infra_api_*_log` | 无 | — | 名称类字段允许重复或由业务层保证，日志表只追加 |

> 唯一键只保护「值唯一」，不保护「语义正确」：`member_user.level_id` 指向已删除等级、`system_user_role.role_id` 指向不存在的角色，数据库都不作约束（本项目不建物理外键，见 6.1），一致性由 Service 层负责。

### 6.9 自动写入约定

统一成一句：**凡是框架、Model 基类或中间件能够推断出来的数据，业务代码一律不显式写入；漏写或手写都算缺陷。** 业务层只负责「没有任何自动来源」的业务字段。

「不显式写入」的执行位置是 **Repository**：查询构造器不触发框架钩子，填充动作必须在写入前显式调用（`Base.FillCreate` / `Base.FillUpdate`），Service 与 Controller 都不参与。

| 数据 | 写入方 | 业务是否显式写 | 说明 |
|---|---|---|---|
| `id` | Repository 写入前调 `Base.FillCreate` | 否 | 雪花 ID，含时钟回拨处理 |
| `creator` / `created_at` | 同上 `FillCreate` | 否 | 当前登录者取自 base 上下文 |
| `updater` / `updated_at` | Repository 写入前调 `Base.FillUpdate` | 否 | 软删除同样先调它 |
| `deleted_at` | Repository 显式写列 | 否 | 构造器不会自动加软删除条件，读写两侧都需显式处理；关联表解除关系用物理删除，见 6.8(4) |
| `system_operate_log` | system 的 `OperateLog` 中间件 | 否 | 语义在路由注册时用元数据声明（9.5） |
| `infra_api_access_log` | infra 的 `ApiAccessLog` 中间件 | 否 | 每个请求一条 |
| `infra_api_error_log` | infra 的 `ApiErrorLog` 中间件 | 否 | 只记系统异常（见 9.5） |
| `auth_login_log` | base 认证组件 | 否 | 业务只调其 Service，不直接写表 |
| `infra_job_log` | infra 任务调度器 | 否 | 处理器只返回结果/错误，由调度器落库 |
| 业务字段（`status`、`point`、`level_id`、`sort`…） | Service / Repository | **是** | 自动约定的反面：这些字段没有任何自动来源 |

落到代码上的三条约束：

1. Controller / Service 中不出现审计字段赋值、不出现 `log.Create(...)`、不自行生成 `trace_id`；`trace_id` 由 `RequestLog` 生成放入上下文，日志中间件与业务都只读取。
2. 无登录态场景（定时任务、数据初始化）的 `creator` / `updater` **留空**，不要用 `"system"` 之类占位值伪造操作人；这类场景的留痕由 `infra_job_log` 承担，不写 `system_operate_log`（跨模块也不允许，见第 10 节）。若确需人工补记，只能由归属 system 的 Service 暴露接口，其他模块不得直连其表。
3. 自动采集只覆盖 HTTP 请求链路。任务、第三方回调等非 HTTP 场景产生的动作**不补写**操作日志与 API 日志。

本节与 6.1（审计字段）、6.3（日志分工）、9.5（中间件位置）说的是同一件事的三个视角（字段 / 日志 / 中间件），实现冲突时以本节为准。

---

## 7. 接口与路由设计

### 7.1 路由总览

统一前缀 `/api/v1`。**后台管理与 C 端使用不同前缀**，同一业务域的代码仍在同一个模块里：
后台走 `/api/v1/{模块}`，C 端走 `/api/v1/client/{模块}`。模块内部按端拆目录，
`admin/route.go` 与 `client/route.go` 各自注册，模块的 `route.go` 只做聚合。

```
/api/v1
├── client/                          # C 端（小程序 / H5 / App），只校验会员令牌
│   ├── infra/
│   │   └── files                   文件上传（头像、评价图等自助场景）
│   └── member/
│       ├── auth/{sms-code,sms-login,social-bind-login}
│       ├── auth/{social/login,refresh,logout,me}
│       ├── auth/social/{authorize-url,bind,bindings}
│       ├── profile                 我的资料（GET / PUT）
│       ├── addresses               我的收货地址（增删改查 + 设默认）
│       ├── sign                    每日签到（POST /sign、GET /sign/status、GET /sign/records）
│       ├── points/records          我的积分流水
│       ├── level/records           我的等级变更记录
│       └── levels                  等级列表 + 我的进度
│
├── system/                          # 后台管理端
│   ├── auth/{login,logout,refresh,captcha}
│   ├── profile                     我的资料
│   ├── password                    修改密码
│   ├── profile-menus               我的菜单与权限
│   ├── users                       用户管理
│   │   └── PUT  /{id}/reset-password
│   ├── roles
│   │   ├── GET  /{id}/menus        角色已授权菜单
│   │   └── PUT  /{id}/menus        分配菜单权限
│   ├── menus                       菜单管理
│   ├── dict-types                  字典类型
│   ├── dict-data                   字典数据
│   ├── notices                     通知公告
│   ├── members                     会员管理（实现在 member 模块）
│   ├── member-levels               会员等级管理（实现在 member 模块）
│   └── logs/{operate,login}        操作日志 / 登录日志
│
├── infra/                           # 后台：基础设施
│   ├── configs                     参数配置
│   ├── files                       文件上传（自助）/ 列表 / 删除
│   ├── file-configs                文件配置
│   ├── sms-channels / sms-templates 短信渠道与模板
│   └── api-logs/{access,error}     API 访问 / 错误日志
│
└── member/                          # 后台：会员管理
    ├── users                       会员档案
    ├── levels / tags / groups      等级 / 标签 / 分组
    └── social-clients              第三方登录渠道配置
```

后台短信配置页挂在 `/api/v1/infra/sms-channels` 与 `/api/v1/infra/sms-templates`，
权限点 `infra:sms-channel:*` / `infra:sms-template:*`；会员端 `/api/v1/client/member/auth/sms-*` 只负责
验证码的生成、校验与登录态，发送动作委托 infra 的 `InfraSmsService`。

会员端登录态与后台登录态使用**同一套双令牌实现、不同的 Redis 作用域**：
会员令牌落在 `auth:member:access:<hash>`，由 `base/middleware.MemberAuth` 校验，后台 `TokenAuth` 无法识别，
反之亦然。第三方渠道（微信小程序 / 公众号 / 开放平台）的凭据放在 `member_social_client` 表，
由后台 `会员管理 → 社交客户端` 页面配置（对齐 yudao 的 `system_social_client`，数据库优先，不走环境变量）。

会员手机号登录（`sms-login`）与第三方绑定登录（`social-bind-login`）共用 `createUserIfAbsent` 语义：
手机号不存在即创建会员（昵称默认「用户 + 后 4 位」），手机号已存在则直接登录。
短信验证码存 Redis：6 位数字、5 分钟有效、一次性消费、单手机号 60 秒发送间隔、每日上限 10 条；
**短信是基础设施能力，渠道与模板配置在 infra 模块**（`infra_sms_channel` / `infra_sms_template`，
对齐 yudao 把 `system_sms_*` 放 system 的做法），后台 `基础设施 → 短信渠道 / 短信模板` 维护，
member 通过 `infra/service.InfraSmsService.Send` 复用发送能力（与 yudao 中 member 依赖 system 一致），
渠道实现按 `code` 选择官方 SDK：`DebugLog`（本地日志，默认）、`AliyunSms`（阿里云 `dysmsapi` SDK）、
`TencentSms`（腾讯云 `sms/v20210111` SDK）。

后台管理会员数据的路由统一挂在 `/api/v1/member/*`，由 `member/admin/route.go` 注册，
受 base 的 `token_auth` + `permission` 中间件保护；会员端接口挂在 `/api/v1/client/member/*`，
由 `member/client/route.go` 注册，只经过 `member_auth`，不进入后台权限体系。
`infra` 同理：`/api/v1/infra/*` 与 `/api/v1/client/infra/*` 分别由 `infra/admin`、`infra/client` 注册，
目前 C 端只开放文件上传。

`system/logs/login` 的数据来自 base 的认证组件，system 通过其 Service 读取，不直接查 `auth_login_log`。

### 7.2 权限标识规范

仅后台路由需要权限校验，格式 `模块:资源:动作`：

- `system:user:list` / `system:user:create` / `system:user:update` / `system:user:delete` / `system:user:reset-password`
- `system:role:list` / `system:role:assign-menu`
- `system:member:list` / `system:member:update`
- `infra:config:list` / `infra:job:update` / `infra:file:upload`

权限标识与 `system_menu.permission` 一一对应，路由注册时声明所需标识，由权限中间件校验。

---

## 8. 接口返回与错误码（对齐 yudao）

### 8.1 统一响应结构

```json
{ "code": 0, "msg": "", "data": {} }
```

| 字段 | 类型 | 说明 |
|---|---|---|
| code | int | 状态码，`0` 表示成功，非 0 为错误码 |
| msg | string | 提示信息，成功时为空字符串 |
| data | T | 业务数据，失败时省略 |

对齐 yudao 的三条硬性约定：

- 字段固定为 `code` / `msg` / `data` 三个，**不额外增加 `success` 字段**，判断成功一律用 `code == 0`。
- 成功时 `msg` 返回**空字符串**，而不是 `"ok"` 之类的文案。
- **HTTP 响应状态码统一为 200**，业务结果只由 body 中的 `code` 表达，不借助 HTTP 状态码传递业务语义。
- 成功响应**固定带 `data` 字段**（无数据时为 `null`），避免前端按字段存在性判断；失败响应省略 `data`。落地上是 `response.Body` 与 `response.FailBody` 两个结构体（15.6）。

> 例外：静态资源、第三方回调等非标准接口不套用该包装，单独定义返回内容，与 yudao 的处理方式一致。

### 8.2 分页约定

请求参数（对应 yudao 的 `PageParam`）：

| 参数 | 类型 | 默认值 | 约束 | 说明 |
|---|---|---|---|---|
| pageNo | int | 1 | ≥ 1 | 页码，从 1 开始；小于 1 时按 1 处理 |
| pageSize | int | 10 | 1 ~ 200 | 每页条数；非正数或超过 200 直接返回 `400` |

响应结构（对应 yudao 的 `PageResult`）：

```json
{ "code": 0, "msg": "", "data": { "list": [], "total": 0 } }
```

分页响应**只包含 `list` 与 `total`**，不返回 `page` / `size` / `pages`，与 yudao 一致。

落地要点：`param.Page.Normalize` 负责校验并返回错误，`pageSize` 不在 1 ~ 200 时返回 `400`，Controller 必须处理该错误。分页接口**不接受 `pageSize = -1`**：不分页的导出场景不套用 `param.Page`，由专用接口自行取全量，Repository 也不为负数准备分支。

### 8.3 错误码体系

对齐 yudao 的两级结构：系统错误码 + 业务错误码。

#### （1）系统错误码（0 ~ 999）

固定段，直接沿用 yudao 的 `GlobalErrorCodeConstants`：

| 错误码 | 常量名 | 提示信息 |
|---|---|---|
| 0 | SUCCESS | 成功 |
| 400 | BAD_REQUEST | 请求参数不正确 |
| 401 | UNAUTHORIZED | 账号未登录 |
| 403 | FORBIDDEN | 没有该操作权限 |
| 404 | NOT_FOUND | 请求未找到 |
| 405 | METHOD_NOT_ALLOWED | 请求方法不正确 |
| 423 | LOCKED | 请求失败，请稍后重试（并发请求） |
| 429 | TOO_MANY_REQUESTS | 请求过于频繁，请稍后重试 |
| 500 | INTERNAL_SERVER_ERROR | 系统异常 |
| 501 | NOT_IMPLEMENTED | 功能未实现/未开启 |
| 502 | ERROR_CONFIGURATION | 错误的配置项 |
| 900 | REPEATED_REQUESTS | 重复请求，请稍后重试 |
| 901 | DEMO_DENY | 演示模式，禁止写操作 |
| 999 | UNKNOWN | 未知错误 |

系统错误码定义在 base 模块，全项目共用，base 自身不再单独占用业务错误码区间。

#### （2）业务错误码（10 位 4 段）

对齐 yudao 的 `ServiceErrorCodeRange`：

| 段 | 位数 | 含义 | 说明 |
|---|---|---|---|
| 第一段 | 1 | 类型 | `1` = 业务级异常，其余预留 |
| 第二段 | 3 | 模块 | 与 yudao 的模块区间一致 |
| 第三段 | 3 | 子模块 | 模块内自行分配 |
| 第四段 | 3 | 错误码 | 子模块内自增 |

模块区间（与 `design/12 §3` 一致）：

| 模块 | 区间 |
|---|---|
| infra | `1-001-000-000 ~ 1-001-999-999` |
| system | `1-002-000-000 ~ 1-002-999-999` |
| member | `1-003-000-000 ~ 1-003-999-999` |
| store | `1-004-000-000 ~ 1-004-999-999` |
| dish | `1-005-000-000 ~ 1-005-999-999` |
| order | `1-006-000-000 ~ 1-006-999-999` |
| promotion | `1-007-000-000 ~ 1-007-999-999` |
| pay | `1-008-000-000 ~ 1-008-999-999` |

子模块分配（本项目）：

| 模块 | 第三段 | 子模块 |
|---|---|---|
| system | 001 / 002 / 003 / 004 / 005 | 认证登录 / 用户 / 角色 / 菜单权限 / 字典 |
| infra | 001 / 002 / 003 / 004 | 参数配置 / 文件 / 文件配置 / 短信 |
| member | 001 / 002 / 003 / 004 / 005 / 006 / 007 / 008 | 会员 / 等级 / 标签 / 分组 / 社交 / 认证 / 地址 / 签到 |
| store | 001 / 002 / 003 / 004 / 005 | 门店 / 桌台 / 员工 / 打印 / 骑手 |

示例：

| 错误码 | 含义 |
|---|---|
| `1-002-001-001` | 验证码错误或已过期 |
| `1-002-001-002` | 账号或密码错误 |
| `1-002-002-001` | 用户不存在 |
| `1-002-003-001` | 角色已被用户引用，不可删除 |
| `1-001-002-001` | 文件上传失败 |
| `1-001-003-001` | 文件配置不存在 |
| `1-003-001-004` | 积分不足 |
| `1-003-008-001` | 今日已签到 |
| `1-004-001-001` | 门店不存在 |

#### （3）错误码的组织与使用

- 系统错误码定义在 base 的 `errs` 包；业务错误码按模块定义在各自模块的 `errs` 文件中，一模块一文件。常量命名（不加 `Err` 前缀、保留模块前缀）见 11.2。
- 每个错误码是一个包含 `code` 与 `msg` 的结构体常量，配套构造函数，支持占位符格式化：
  `errs.New(errs.SystemUserNotExists, "用户【%s】不存在", username)`。
- 业务异常类型对应 yudao 的 `ServiceException`：Service 层直接抛出错，由全局异常中间件（即 9.5 中的 `ExceptionHandler`）统一转换为 `{"code","msg","data":null}`，Controller 不写错误分支。
- 错误码的提示文案全部由代码常量维护，不落库、不支持运行时动态覆盖，避免出现两个真实来源。

### 8.4 ID 字段的序列化

采用雪花 ID 后，主键与外键都是 int64（19 位），而 JavaScript 的 `Number` 只能安全表示 2^53-1（16 位），直接以数字返回会被**静默截断**。因此约定：

- **所有 ID 类型字段在 JSON 中一律以字符串返回**，统一声明为 base 的 `types.ID` 类型（见下），由类型自身保证序列化行为，不需要逐字段加标签。
- 前端不对 ID 做数值运算，只作为标识透传与展示。
- 请求侧不受影响：路径参数 `/{id}` 本身是字符串，解析时转 `int64`；Body 中的 ID 字段同样接收字符串。
- 非 ID 的整型字段（如 `status`、`sort`、`type`）保持数字，不受此约定影响。

涉及的字段包括所有表的主键，以及 `dept_id`、`user_id`、`role_id`、`menu_id`、`level_id`、`config_id`、`job_id`、`biz_id` 等外键与业务 ID。

> 反向说明：`area_id`（地区编码）、`size`（文件字节数）、`point`（积分）等虽然也是数值，但不属于**实体标识**，保持数字类型，不套用 `types.ID`。

> 数据库层不需要为此做任何处理，字符串化只发生在 JSON 编解码边界。

#### 统一 ID 类型

逐个字段手写 `json:",string"` 迟早会漏，因此 base 提供统一的 `ID` 类型，从类型层面根除：

```go
// app/modules/base/types/id.go
package types

type ID int64

func (i ID) String() string { return strconv.FormatInt(int64(i), 10) }

// 序列化一律输出字符串
func (i ID) MarshalJSON() ([]byte, error) { return json.Marshal(i.String()) }

// 反序列化同时兼容字符串与数字，容错前端传参格式不统一
func (i *ID) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*i = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*i = ID(v)
	return nil
}

// 存取数据库：查询构造器通过这两个方法读写
func (i ID) Value() (driver.Value, error) { return int64(i), nil }

func (i *ID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*i = 0
	case int64:
		*i = ID(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*i = ID(n)
	default:
		return fmt.Errorf("types.ID: 无法从 %T 扫描", src)
	}
	return nil
}
```

使用约定：

| 位置 | 写法 |
|---|---|
| Model 主键与外键 | `ID types.ID` |
| VO / DTO 中的 ID 字段 | `ID types.ID` |
| 路径参数 | `strconv.ParseInt(ctx.Request().Input("id"), 10, 64)` 后转 `types.ID` |
| Redis 值等裸 int64 场景 | 显式 `strconv.FormatInt(int64(id), 10)` |

这样 Model、VO、Request 全部声明为 `types.ID`，序列化行为自动统一，无需再逐字段加 tag，也杜绝了新字段忘记加 `,string` 的问题。

> 注意两点：Model 中主键需显式声明 `db:"id"`，不要依赖字段名推导；`types.ID` 只用于**实体标识**，`status`、`sort`、`type` 这类枚举与计数字段不要套用。

---

## 9. 认证与权限流程

### 9.1 两套独立认证体系

| 维度 | 后台（system） | 会员（member） |
|---|---|---|
| 账号表 | `system_users` | `member_user` |
| 登录入口 | `/api/v1/system/auth/login` | `/api/v1/client/member/auth/{sms-login,social/login}` |
| 认证 guard | `admin` | `member` |
| 中间件 | `token_auth` | `member_auth` |
| 令牌载荷 | 无，令牌即 Redis key | 无，令牌即 Redis key |
| 权限控制 | RBAC 逐项校验 | 仅登录态校验 |
| 令牌存储 | Redis，双令牌（见 9.2） | Redis，双令牌（见 9.2） |
| 登录日志 | `auth_login_log`，`user_type=1` | 同表，`user_type=2` |

两套令牌互不通用：`token_auth` 只认后台用户的令牌，`member_auth` 只认会员的令牌，交叉访问返回系统错误码 `401`（账号未登录）。

### 9.2 令牌方案（双令牌 + Redis）

采用**双令牌**方案，令牌状态全部落在 Redis，不使用 JWT：

| 令牌 | 用途 | 有效期 | 存储 |
|---|---|---|---|
| access_token | 随每次业务请求放在 `Authorization: Bearer` 中 | `ACCESS_TOKEN_TTL`，默认 1800 秒 | `auth:access:{sha256(token)}` → 用户 ID |
| refresh_token | 只用于换取新的一对令牌，不参与业务请求 | `REFRESH_TOKEN_TTL`，默认 604800 秒 | `auth:refresh:{sha256(token)}` → 用户 ID |

- 登录：校验账号密码与状态 → 生成两个 32 字节随机串 → 写入 Redis → 返回 `accessToken` / `refreshToken` / `expiresIn` / `refreshExpiresIn`。
- 鉴权：`token_auth` 中间件用访问令牌查 Redis，命中才放行并把用户 ID 注入上下文；未命中返回 `401`。
- 续期：`POST /auth/refresh` 用刷新令牌换新的一对令牌，**旧刷新令牌立即删除**（轮换），旧访问令牌按自身短 TTL 自然过期。
- 登出：`POST /auth/logout` 删除两个 key，立即失去访问权与续期能力；该接口不挂鉴权，访问令牌已过期时也能调用。
- Redis 只存令牌的 SHA-256，不存明文：Redis 数据被读走也拿不到可用令牌。
- 多端登录：每次登录生成独立的 key，天然互不影响；需要踢下线时删除对应用户的 `auth:*` 键即可。
- 代价：每个受保护请求多一次 Redis 查询；Redis 不可用时登录态整体失效（fail closed），这是有状态方案的必然取舍。

Redis 连接在 `config/database.go` 的 `redis.default`，缓存驱动在 `config/cache.go`（`default=redis`），有效期在 `config/token.go`。

### 9.3 后台权限（RBAC）

1. 登录：校验账号密码 → 校验 `status` → 调用 base 认证组件签发令牌 → 写 `auth_login_log`。
2. `token_auth` 用访问令牌查 Redis 并注入当前用户；白名单路由跳过。
3. `permission` 中间件查询权限集合：`system_user_role` → `system_role_menu` → `system_menu.permission`，与路由声明的标识比对，不通过返回系统错误码 `403`（没有该操作权限）。
4. 权限集合缓存，key 形如 `rbac:perms:{userId}`；角色、菜单、角色授权或用户角色变更时主动清除缓存。
5. 超级管理员（角色 `code = SUPER_ADMIN`）直接放行。
6. 简单系统不做数据权限：接口级 RBAC 已覆盖「能不能访问」，不按组织架构再收窄数据范围。

### 9.4 会员认证

1. 注册：校验手机号唯一 → 密码哈希 → 赋默认等级 → 可选赠送注册积分（写 `member_point_record`）。
2. 登录：校验密码与 `status` → 签发令牌 → 更新 `login_ip` / `last_login_at` → 写登录日志。
3. 请求：`member_auth` 解析令牌并校验 `type=member`。
4. 不做接口级权限控制，业务限制（积分不足、等级门槛、重复签到）在 Service 层判断。

### 9.5 中间件链

```
后台：ApiErrorLog → ApiAccessLog → Recovery → ExceptionHandler → CORS → RequestLog → TokenAuth(白名单) → Permission → OperateLog
会员：ApiErrorLog → ApiAccessLog → Recovery → ExceptionHandler → CORS → RequestLog → MemberAuth(白名单)
```

三处顺序不是随手排的：

- `ApiAccessLog` 特意注册在靠外层，是为了统计**包含后续所有中间件在内的完整耗时**。用户身份在 `c.Next()` 返回后从上下文读取——此时 `TokenAuth` 已把用户写入上下文，因此最外层也能拿到身份，无需把它挪到鉴权之后。同时这样还能记录到被 `Permission` 拦截（403）的请求。
- `ApiErrorLog` 比 `ApiAccessLog` 更靠外，是为了让错误日志的落库耗时不计入访问日志的 `duration`：中间件的前置阶段由外向内进入、**后置阶段由内向外返回**，最外层的 `ApiErrorLog` 最后落库，`ApiAccessLog` 统计耗时（后置阶段）时它还没开始写库。
- `Recovery` 不再是注册链上的第一个：两个日志中间件都在它之外，原因见 9.5「错误日志的边界」。

`Recovery`、`ExceptionHandler`、`CORS`、`RequestLog`、`TokenAuth`、`MemberAuth`、`Permission` 实现于 base；`ApiAccessLog`、`ApiErrorLog` 实现于 infra；`OperateLog` 实现于 system（写 `system_operate_log`）。

`ExceptionHandler` 是 8.3 中「全局异常中间件」的落地，把 Service 抛出的业务异常转换为 `{"code","msg","data":null}`；`Recovery` 只兜住它没接住的异常。

`OperateLog` 注册在 `Permission` 之内，只采集**已通过鉴权并实际执行**的操作；被鉴权拦下的请求由 `ApiAccessLog` 留痕。操作语义在路由注册时声明：

```go
r.Post("/users", userController.Create).Middleware(operate("用户管理", "新增用户"))
```

中间件在后置阶段读取该元数据与最终响应码，落库时一并写入 `success`。

#### 错误日志的边界：`Recovery`（base）与 `ApiErrorLog`（infra）

`ApiErrorLog` 注册在 `Recovery` **之外**，这是由 Go 的 defer 展开顺序决定的：panic 发生时 defer 由内向外展开，**只有 `Recovery` 之外的中间件，才能在其后置阶段看到已经归一化的异常**。若把 `ApiErrorLog` 放进 `Recovery` 之内，它的后置逻辑会先于 `Recovery` 执行，读到的错误槽恒为空，落库永远写不出东西。（这也与 Gin 自带的 `Logger` 放在 `Recovery` 之外是同一个道理。）

两者职责严格切分，互不越界：

| | `Recovery`（base） | `ApiErrorLog`（infra） |
|---|---|---|
| 触发 | 任意 panic（nil 解引用、越界、未预期错误） | 后置阶段发现错误槽非空 |
| 职责 | 归一化：`recover()` → 写出 `500 INTERNAL_SERVER_ERROR` 响应 → 把异常写入请求上下文的「错误槽」 | 读取错误槽 → 组装记录 → 落库 `infra_api_error_log` |
| 依赖 | 只用 base 自身能力（响应封装、上下文、框架日志），**不触碰任何数据表** | 依赖 base 的上下文与错误槽，只写本模块表 |
| 失败处理 | 写错误槽不会失败；响应写出失败仅记框架日志 | 落库失败仅记框架日志，**不再抛出**，避免 panic 套 panic |

调用方式（三步，始终单向依赖）：

1. base 只提供读写接口：`errs.Capture(ctx, recovered)` 写入错误槽（异常类型名、message、堆栈），`errs.Captured(ctx)` 读出。错误槽随请求上下文创建、随请求结束释放，不跨请求、不落库、不做任何持久化。
2. infra 侧不注册回调、不 import base 的任何错误处理分支，只在 `c.Next()` 之后读一次错误槽：**非空才落库**，为空时整条链路只多一次内存判断。
3. 由于 base 不依赖 infra（第 10 节），**落库动作必定发生在 infra 侧**；`Recovery` 中不会、也不允许出现写 `infra_api_error_log` 的代码。被有意排除的替代方案是「base 暴露 `OnRecover` 钩子、infra 在 bootstrap 注册回调」：它会在 panic 路径上引入跨模块同步调用，并把注册顺序变成隐式依赖，因此不采用。

**只收系统异常**：这张表只记 `500`/未预期异常，等价于「服务端故障清单」，也才能让条目少到可被值班直接消费；「账号密码错误」「积分不足」属于正常业务失败，不是故障，写进来只会把真正的故障淹没。具体划分：业务异常由 `ExceptionHandler` 转为正常响应（HTTP 仍 200、`code` 非 0），鉴权拦截返回 401/403，这两类都由 `ApiAccessLog.result_code` 留痕，**不写错误日志**。该划分与 yudao 一致——业务异常在全局异常处理器内消化，错误日志只兜未捕获的异常。

> 边界提醒：把数据库、网络等 I/O 层的裸 `error` 直接 panic，会被 `ExceptionHandler` 判为非业务异常落到 `500` 并写入错误日志。Service 应把可预期的失败包装成业务错误码抛出，只让真正意料之外的错误进入这张表。

字段来源：

| 字段 | 来源 |
|---|---|
| `trace_id`、`user_id`、`user_type`、`user_ip`、`user_agent` | base 注入的请求上下文 |
| `application_name` | 应用配置（`config/app.go` 的 `name`） |
| `request_method`、`request_url`、`request_params` | 请求本身；`request_params` 复用 `ApiAccessLog` 的脱敏函数，禁止维护两套脱敏逻辑 |
| `name`、`message`、`exception_stack` | 错误槽（`Recovery` 写入） |
| `process_status`、`process_time` | 建表默认「未处理」/ NULL，由后台「标记已处理」接口更新，中间件不写 |

> 若某条链路需要进一步降低落库对响应的影响，可在 `ApiErrorLog` 中异步入库，但必须先复制出所需字段再交给 goroutine，**不得在 goroutine 中复用 `*gin.Context`**（请求结束后上下文会被回收）。异步化只允许出现在这一处：日志丢失的优先级低于请求正确性。

---

## 10. 跨模块调用约定

1. 只允许调用对方 Service，禁止跨模块 import 对方的 Repository / Model。
2. 三个业务模块当前互不依赖。若后续需要跨模块取数（如后台展示会员摘要），由数据归属方暴露 Service 接口，调用方单向依赖，并同步更新本文档的依赖图。
3. 跨模块调用不共享事务；若需强一致，采用「先校验、后提交、失败补偿」策略，不做分布式事务。
4. `infra` 的 API 日志中间件、`system` 的操作日志中间件，需要当前用户信息时通过 base 的上下文工具读取，不反向查询 `system_users`。

---

## 11. 命名规范

| 项 | 规范 | 示例 |
|---|---|---|
| base 表 | `auth_` 前缀 | `auth_login_log` |
| system 表 | `system_` 前缀 + 小写下划线 | `system_users`、`system_role_menu` |
| infra 表 | `infra_` 前缀 | `infra_config`、`infra_job_log` |
| member 表 | `member_` 前缀 | `member_user`、`member_level` |
| 模块目录包名 | 小写单词 | `system`、`infra`、`member` |
| 文件名 | 小写下划线 | `user_service.go`、`job_service.go` |
| 路由路径 | 小写中划线 | `/reset-password`、`/api-logs/access` |
| 权限标识 | `模块:资源:动作` | `system:user:create`、`infra:job:update` |
| 错误码常量 | 模块 + 语义，**不叠加 `Err` 前缀**（结论与理由见 11.2） | `errs.SystemPasswordWrong`、`errs.MemberPointsNotEnough` |
| Model 结构体 | PascalCase 单数形式 + `Model` 后缀 | `SystemUserModel`（对应 `system_users`）、`MemberLevelModel` |
| 请求结构体 | PascalCase + `Request` 后缀，动作在前 | `UserCreateRequest`、`UserUpdateRequest`、`LoginRequest` |
| ID 类型 | 主键与外键统一用 `types.ID` | `ID types.ID`（见 8.4） |
| 枚举类型 | PascalCase 语义名 + `Enum` 后缀 | `CommonStatusEnum`、`LoginTypeEnum` |
| 枚举常量 | 语义名 + 取值，不带 `Enum` 后缀 | `CommonStatusEnabled`、`LoginTypeLogout` |
| 表与字段注释 | 建表必须写表注释与全部字段注释，内容与 6.x 字段说明一致，状态类字段注明取值 | `COMMENT '状态：0 禁用 / 1 启用'` |

> 表名与结构体名的对应关系在 Model 的 `TableName()` 中显式声明，避免框架复数化推导错误（如 `system_users` 对应 `SystemUserModel`）。

### 11.1 枚举定义

Go 没有 `enum` 关键字，统一用「自定义类型 + `const` 块」表达枚举，禁止把裸 `int` 或字符串字面量散落在业务代码里。

```go
// app/modules/base/enums/common_status.go
package enums

type CommonStatusEnum int8

const (
	CommonStatusDisabled CommonStatusEnum = 0 // 禁用
	CommonStatusEnabled  CommonStatusEnum = 1 // 启用
)
```

约定：

| 项 | 要求 |
|---|---|
| 底层类型 | `int8`，与表里的 `tinyint` 对齐 |
| 定义位置 | 只被单个模块使用的放该模块 `enums/` 包；跨模块共用的必须放 base 的 `enums/`，不得各模块重复定义 |
| 启停状态 | 用户、角色、菜单等表的 `status` 统一使用 `base/enums.CommonStatusEnum`，不按业务再定义 `XxxStatusEnum` |
| 连续值 | 用 `iota`；需要从 `1` 开始时写 `iota + 1` |
| 必须实现 | `String()` 用于日志与描述，`IsValid()` 用于参数校验 |
| 参数校验 | Request 中先过 `IsValid()`，不把取值硬编码进校验 tag |

需要注意的三点：

- **零值陷阱**：`var s CommonStatusEnum` 默认是 `0`，而本项目 `status` 的 `0` 恰是「禁用」这个有效值。`IsValid()` **拦不住零值**（0 是合法值），所以「字段是否传了」必须由请求校验的 `required` 保证；Service 里的 `IsValid()` 只负责越界判断。
- **穷尽性**：Go 的 `switch` 不检查是否覆盖全部枚举值，新增值后漏改不会编译报错。枚举数量上来后，在 CI 接入 `exhaustive` linter。
- **序列化**：不实现额外方法时 JSON 输出为数字；若前端期望字符串，实现 `MarshalJSON` 统一转换。数据库侧统一映射为 `tinyint`，Model 中声明为枚举类型并配 `db` tag 即可，不要加 `gorm` tag。

枚举增多后可用 `stringer`（只生成 `String()`）或 `go-enum`（`String()` / `MarshalJSON` / `UnmarshalJSON` / `IsValid()` 全套）生成样板代码，注释中写 `//go:generate` 后用 `go generate ./...` 执行。

### 11.2 错误码命名

错误码常量是**带 `code` 与 `msg` 的结构体值**（见 8.3），不是 Go 的 `error` 值，因此命名只由「模块 + 语义」两段构成：

```go
// app/modules/system/errs/system_error.go
var (
	SystemPasswordWrong  = ErrorCode{Code: 1_002_001_000, Msg: "登录失败，账号或密码不正确"}
	SystemUsernameExists = ErrorCode{Code: 1_002_002_002, Msg: "用户【%s】已存在"}
)

// 使用：Service 中直接抛出，Controller 不写错误分支
return errs.New(errs.SystemUsernameExists, username)
```

#### （1）不叠加 `Err` 前缀（确认维持）

结论：**维持第 11 节表格中的约定，不写成 `errs.ErrSystemPasswordWrong`**。理由三条：

1. **冗余**：包名 `errs` 已经表明这是错误相关标识，再叠 `Err` 构成 stutter（`errs.ErrSystemPasswordWrong` 读作「errs·Err·System…」），Go 的命名风格明确避免这种重复。
2. **类型误导**：`Err` 前缀在 Go 社区是留给 `error` 类型哨兵值的（`sql.ErrNoRows`、`os.ErrNotExist`、`fs.ErrNotExist`）。给一个结构体常量加 `Err` 前缀，会诱导调用方写出 `return errs.ErrSystemPasswordWrong` 这类类型不匹配的代码，也让 `errors.Is` 的使用者误判这是可比较的 error。
3. **无信息增量**：调用点已经足够清晰——`errs.New(errs.SystemPasswordWrong, ...)`；前缀不多表达任何信息。

同时排除反面做法：本项目**不定义 `error` 类型的错误哨兵**。Service 一律用 `errs.New(...)` 构造并抛出，判断错误用「取错误码比对」，不用 `errors.Is`。

#### （2）模块前缀必须保留

`System` / `Infra` / `Member` 前缀不是可选项：

- 项目里存在**五个同名的 `errs` 包**（base 的系统错误码 + 四个模块的业务错误码）。跨模块阅读时 `xxx.PasswordWrong` 看不出归属，带模块前缀才能「看名字就知道去翻哪张错误码表」。
- 与 8.3 的错误码区间一一对应：`System*` → `1-002-*-*`，`Infra*` → `1-001-*-*`，`Member*` → `1-003-*-*`，`Store*` → `1-004-*-*`。
- base 的系统错误码**不带**模块前缀（`errs.BadRequest`、`errs.InternalServerError`），因为它们属于固定段（0 ~ 999）、全项目共用，没有模块归属。

若同一文件同时需要 base 与模块的 `errs`，把 base 侧以 `baseErrs` 别名引入，不要为此改命名。

#### （3）系统错误码常量名同步为 PascalCase

8.3 表格里的「常量名」列沿用 yudao 的 `GlobalErrorCodeConstants` 原样（`BAD_REQUEST`、`INTERNAL_SERVER_ERROR`），仅用于对照；**Go 侧定义时同样使用 PascalCase**：

`errs.Success`、`errs.BadRequest`、`errs.Unauthorized`、`errs.Forbidden`、`errs.NotFound`、`errs.MethodNotAllowed`、`errs.Locked`、`errs.TooManyRequests`、`errs.InternalServerError`、`errs.NotImplemented`、`errs.ErrorConfiguration`、`errs.RepeatedRequests`、`errs.DemoDeny`、`errs.Unknown`。

取值与 yudao 完全一致（`400`/`401`/`500`…），只是名字符合 Go 风格与本节 (1) 的结论，避免同一份文档里出现两套命名风格。

#### （4）语义粒度

- 写法统一为「主语 + 状态/动作」：`SystemUserNotExists`、`SystemRoleUsedByUser`、`MemberPointsNotEnough`；不写成 `SystemUserNotExistError`。
- 一个错误码只表达一件事；「参数不合法」这类粗粒度判断归系统错误码 `400`，不另造业务码。
- 错误码的 `code` 与常量名一一对应且**不可回收**：删除某个错误码后其数值不再复用，因为前端与历史日志可能长期留存该值。

### 11.3 注释规范

Go 侧按标准 doc comment 形式书写（`// Name ...`），但**内容必须是业务语义，不能是路由复述**：

```go
// 反例：等于把函数名与路由抄一遍，零信息量
// Index GET /api/v1/system/users

// 正例：做什么 + 关键约束 + 入参 + 返回
// Index 分页查询后台用户，支持账号、昵称、手机号与状态筛选。
// 入参：username、nickname、mobile、status、pageNo、pageSize（查询串）；返回用户列表与总数。
func (c *UserController) Index(ctx http.Context) http.Response {
```

分层要求：

| 层           | 注释要点                                                    |
| ------------ | ----------------------------------------------------------- |
| Controller   | 业务动作与可选筛选；`入参：` 来源（查询串 / 请求体 / 路径）；`返回：` 结构 |
| Service      | 业务规则与边界：唯一键、事务范围、软删、缓存版本前移        |
| Repository   | 查询条件、排序方式、「不存在返回 nil」这类约定              |
| Request      | `Rules` 校验了什么、`Messages` 覆盖哪些提示                 |
| 私有函数     | 一句话说明用途（键构造、入参清洗、类型转换）                |
| 构造函数     | 装配了哪些依赖                                              |
| 迁移         | `Up` 的建表 / 补列 / 种子内容与幂等前提；`Down` 的回滚动作  |

前端沿用同一原则：函数上方用 `// 中文说明`，写触发时机与副作用，例如
「搜索：合并搜索栏条件后重新查询」「弹窗提交成功：按操作类型刷新列表」。

---

## 12. 现有代码的迁移

| 现有内容 | 处理方式 |
|---|---|
| `app/http/controllers/user_controller.go` | 已删除（示例代码，业务改由 `app/modules/*/controller/` 承接） |
| `routes/web.go` 中的 `/users` 示例路由 | 已随示例控制器移除 |
| `routes/web.go` 的欢迎页与 `public` 静态目录 | 保留 |
| `config/database.go` | 由 Postgres 改为 MySQL 连接配置 |
| `config/auth.go` | 守卫由 `user` 改为 `admin` / `member` 双 guard |
| `go.mod` | 替换数据库驱动依赖 |
| `database/migrations/` 现有 jobs 表迁移 | 与 `infra_job` 职责重叠，落地时移除，统一用 `infra_job` |
| `app/facades/`、`config/` 其他文件、`bootstrap/` | 全部保留，作为模块的依赖门面 |

---

## 13. 开发落地顺序

1. **数据库切换**：`go.mod` 换驱动、`config/database.go` 改连接、`.env` 配 MySQL 并验证连通。
2. **base 模块**：响应封装 → 错误码 → **雪花 ID 生成器（含时钟回拨处理）** → Model 基类（审计字段与 `FillCreate` / `FillUpdate`）→ 通用请求基类 → 中间件（`Recovery`、`ExceptionHandler` 与错误槽、`CORS`、`RequestLog`、`TokenAuth` / `MemberAuth` / `Permission`，链序见 9.5）→ 认证组件（双令牌签发与 Redis 校验 + 登录日志表）。雪花 ID 必须最先落地，各模块建表与写入都依赖它。
3. **system 模块**：迁移 → Model → Repository → Service → Controller → 路由。建议顺序：菜单 → 角色（含角色授权）→ 用户（含分配角色、重置密码）→ 认证登录 → 字典 → 通知公告 → 日志。
4. **infra 模块**：参数配置 → 文件配置与文件上传 → 定时任务与任务日志 → API 访问/错误日志中间件。
5. **member 模块**：迁移 → Model → 等级 → 会员 → 注册/登录/找回密码 → 个人资料 → 积分与签到 → 标签与分组 → 后台会员管理接口。
6. **聚合路由**：新增 `routes/modules.go`，在 `bootstrap/app.go` 的 `WithRouting` 中调用，并配置 `config/auth.go` 的双 guard。
7. **自测**：用 Goravel 的 HTTP 测试能力覆盖后台登录、权限拦截、用户 CRUD、会员注册登录、文件上传五条主链路。

---

## 14. 后续可扩展方向

- **消息能力**：参考 yudao 补充短信（`system_sms_channel` / `template` / `log` / `code`）、邮件（`system_mail_account` / `template` / `log`）、站内信（`system_notify_template` / `message`）。
- **地区数据**：补充 `system_area` 或引入第三方地区库，替换 member 中的 `area_id` 裸编码。
- **会员成长体系**：签到连续奖励、经验值自动升级（定时任务扫描 `member_point_record`）。
- **组织架构与数据权限**：当前只做接口级 RBAC；后续若接入部门/岗位，再按组织树扩展数据范围。
- **日志治理**：`infra_api_access_log` / `system_operate_log` / `auth_login_log` 数据量大，补充归档与清理定时任务。
- **分库分表**：主键已采用雪花 ID，天然支持水平拆分；后续按业务域或时间分库时，无需改造主键方案。

---

## 15. 代码模板

本章给出按**当前已落地实现**整理的模板（查询构造器 + `base/model.BaseModel`），新增一个实体时逐节复制改名即可。模板与 5.x / 6.x 的文字约定若有不一致，优先以本章为准，差异清单见 15.10。

占位约定：表 `system_xxx`、结构体 `SystemXxxModel`、文件前缀 `xxx_`、错误码常量 `SystemXxxNotExists`。

### 15.1 一个实体要新增的文件

| 文件 | 说明 |
|---|---|
| `app/modules/{module}/model/xxx.go` | 表映射与表名常量，嵌入 `model.BaseModel` |
| `app/modules/{module}/enums/xxx_type.go` | 本模块专属枚举；启停状态复用 base 公共枚举，不建 `xxx_status.go`（11.1） |
| `app/modules/{module}/repository/xxx_repository.go` | 单表读写，唯一出现 SQL/构造器的地方 |
| `app/modules/{module}/service/xxx_service.go` | 业务规则、事务边界，直接持有具体仓储类型 |
| `app/modules/{module}/admin/controller/xxx_controller.go` | 取参 → 调 Service → 出响应（C 端放 `client/controller`） |
| `app/modules/{module}/admin/request/xxx_request.go` | 入参、校验规则与中文文案（C 端放 `client/request`） |
| `app/modules/{module}/errs/{module}_error.go` | 本模块业务错误码（一模块一文件，只追加） |
| `app/modules/{module}/admin/route.go` | 追加管理端路由（C 端写 `client/route.go`） |

**controller / request / route 按端分目录**（见 7.1）：`app/modules/{module}/admin/{controller,request,route.go}`
与 `app/modules/{module}/client/{controller,request,route.go}`，`service` / `repository` / `model` / `errs`
两端共用；`{module}/route.go` 只负责调用 `admin.Register()` 与 `client.Register()`。
`system` 目前只有 `admin/`（C 端不需要系统管理能力），`infra` / `member` 两端都有。
| `database/migrations/YYYYMMDDHHMMSS_create_xxx_table.go` | 建表 |

跨层复用的助手放在 base，不放业务模块：`base/result`（`Success` / `Page` / `Fail`）、`base/param`（`Bind` / `Validate` / `PathID` / `Page`）、`base/response`（`Body` / `FailBody` 与对应构造函数）、`base/support`（`IsDuplicate` / `NewID` / `UserID`）。

命名对照：表 `system_notice` → Model `SystemNotice` → 文件 `notice_controller.go`（文件名取表名去掉模块前缀，见 11 节）。

### 15.2 Model

```go
// Package model 存放 system 模块的表映射。
package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/model"
)

// SystemXxxTable 是表名常量，Model 与 Repository 共用同一来源。
const SystemXxxTable = "system_xxx"

// SystemXxxModel 对应 system_xxx 表。
type SystemXxxModel struct {
	model.BaseModel
	Name   string                     `db:"name" json:"name"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 显式声明表名，不依赖复数化推导（11 节）。
func (m *SystemXxxModel) TableName() string {
	return SystemXxxTable
}
```

五条硬约束：

1. 标签用 `db:"..."`（查询构造器按它映射列），**不写 `gorm:"..."`**；`gorm` 标签在本项目的查询构造器链路里不生效。
2. 主键与外键用 `types.ID`，时间字段用 `types.DateTime`，不自建 ID 类型、不用裸 `time.Time`（8.4、6.1）。
3. 审计字段靠嵌入 `model.BaseModel` 获得，**不要重复声明** `id` / `creator` / `created_at` / `updater` / `updated_at` / `deleted_at`。
4. 表名以常量导出，**Repository 引用该常量**而不是再写一遍字面量，避免改表名漏改一处。

`status` 字段不单独定义枚举，直接复用 base 的公共启停枚举（11.1）：

```go
Status baseEnums.CommonStatusEnum `db:"status" json:"status"` // 0 禁用 / 1 启用
```

只有 `xxx` 表确实存在专属业务枚举（如 `xxx_type`）时，才在 `app/modules/system/enums/` 下新增文件，按 11.1 实现 `String()` 与 `IsValid()`，命名形如 `XxxTypeEnum` / `XxxTypeBuiltin`。

### 15.3 Repository

```go
// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
)

const xxxTable = model.SystemXxxModelTable

// XxxRepository 是 system_xxx 表的仓储。
type XxxRepository struct{}

// NewXxxRepository 构造仓储实例；仓储无状态，不需要注入依赖。
func NewXxxRepository() *XxxRepository {
	return &XxxRepository{}
}

// Paginate 按关键字分页，只查未软删的数据。
// pageSize 由 Controller 的 param.Page.Normalize 保证落在 1 ~ 200（docs 8.2）。
func (r *XxxRepository) Paginate(ctx context.Context, pageNo, pageSize int, keyword string) ([]model.SystemXxxModel, int64, error) {
	list := make([]model.SystemXxxModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(xxxTable).WhereNull("deleted_at")
	if keyword != "" {
		query = query.WhereLike("name", "%"+keyword+"%")
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Find 按主键查询未软删的记录；未命中返回 (nil, nil)，由 Service 决定是否抛业务异常。
func (r *XxxRepository) Find(ctx context.Context, id types.ID) (*model.SystemXxxModel, error) {
	var m model.SystemXxxModel

	if err := facades.DB().WithContext(ctx).Table(xxxTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}

	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Create 插入一条记录，主键与审计字段由 Base.FillCreate 填充。
func (r *XxxRepository) Create(ctx context.Context, m *model.SystemXxxModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(xxxTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemXxxNameExists, m.Name)
	}

	return err
}

// Update 全量更新业务字段与审计字段。
func (r *XxxRepository) Update(ctx context.Context, m *model.SystemXxxModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(xxxTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemXxxNameExists, m.Name)
	}

	return err
}

// Delete 软删除：只写 deleted_at，不物理删除。
func (r *XxxRepository) Delete(ctx context.Context, m *model.SystemXxxModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(xxxTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(systemErrs.SystemXxxNotExists, m.ID)
	}

	return nil
}
```

七条铁律：

1. **写操作用 `map[string]any`，不用结构体**。构造器把结构体转列时会跳过零值字段，`status=0`、空 `remark` 这类合法值会被静默吞掉。
2. **软删除靠显式 `WhereNull("deleted_at")` + 写 `deleted_at`**，不依赖 GORM 软删除；构造器不会自动加这个条件。
3. **不触发 GORM hook**，所以 `FillCreate` / `FillUpdate` 必须在写之前显式调用，否则审计字段为空。
4. **不开事务**（5.3）。需要事务时由 Service 用 `facades.DB().Transaction(func(tx contractsdb.Tx) error { ... })` 开启，并把 `tx` 作为参数传进 Repository 的重载方法。
5. **唯一键冲突用 `support.IsDuplicate` 判定**：helper 已在 `app/modules/base/support/duplicate.go`，判断 MySQL `1062` 后转成与该字段「已存在」相同的业务错误码，禁止把驱动原始错误透传给前端（6.8(5)）。
6. **分页只接受 1 ~ 200**：`pageSize` 的校验放在 Controller 的 `param.Page.Normalize`，Repository 直接 `Paginate`，不为 `-1` 准备分支（8.2）。
7. **`RowsAffected == 0` 只在软删除时判定「不存在」**。`Update` 不能据此判断：MySQL 在字段值未变化时同样返回 0 行，会把「重复提交同一份数据」误报成记录不存在。

### 15.4 Service

```go
// Package service 存放 system 模块的业务逻辑，事务只在这一层开启。
package service

import (
	"context"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// XxxService 是 system_xxx 的业务入口。
// 直接持有具体仓储类型；等真要写单测时，再在测试文件就近定义小接口（5.3）。
type XxxService struct {
	repo *repository.XxxRepository
}

// NewXxxService 构造 Service。
func NewXxxService() *XxxService {
	return &XxxService{repo: repository.NewXxxRepository()}
}

// Page 分页查询。
func (s *XxxService) Page(ctx context.Context, pageNo, pageSize int, keyword string) ([]model.SystemXxxModel, int64, error) {
	return s.repo.Paginate(ctx, pageNo, pageSize, keyword)
}

// Detail 查询单条，不存在时抛业务异常。
func (s *XxxService) Detail(ctx context.Context, id types.ID) (*model.SystemXxxModel, error) {
	m, err := s.repo.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(systemErrs.SystemXxxNotExists, id)
	}
	return m, nil
}

// Create 新建一条记录。
func (s *XxxService) Create(ctx context.Context, name string, status baseEnums.CommonStatusEnum, remark string) (*model.SystemXxxModel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, baseErrs.New(systemErrs.SystemXxxNameEmpty)
	}
	if !status.IsValid() {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	m := &model.SystemXxxModel{Name: name, Status: status, Remark: remark}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}
```

约定：

- 错误一律 `return nil, baseErrs.New(...)` / `baseErrs.WithMsg(...)` 抛出，**Controller 不写业务判断**；业务错误码常量来自本模块的 `systemErrs`，两者在文件里以 `baseErrs` / `systemErrs` 区分（11.2）。
- 不出现 `log.Create(...)`、不生成 `trace_id`、不赋值审计字段（6.9）。
- 直接持有具体仓储类型，**不预抽 interface**。等确实需要 mock 时，再由测试文件就近定义小接口并把字段改成该接口（5.3），避免为「将来可能的测试」提前付出抽象成本。

### 15.5 Request

```go
// Package request 存放 system 模块的请求参数与校验规则。
package request

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
)

// XxxCreateRequest 创建入参。
type XxxCreateRequest struct {
	Name   string                     `form:"name" json:"name"`
	Status baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Remark string                     `form:"remark" json:"remark"`
}

// Rules 由 Controller 交给 param.Bind 执行。
func (r *XxxCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"status": "required|integer|in:0,1",
		"remark": "max:500",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *XxxCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "名称不能为空",
		"name.max":        "名称不能超过 64 个字符",
		"status.required": "状态不能为空",
		"status.integer":  "状态必须为整数",
		"status.in":       "状态取值不合法",
		"remark.max":      "备注不能超过 500 个字符",
	}
}
```

- 更新入参单独一个结构体（`XxxUpdateRequest`），**字段同样必填**，保持 PUT 的幂等语义，不做部分更新。
- 字段类型用枚举（状态字段为 `baseEnums.CommonStatusEnum`），不用裸 `int8`；取值范围仍由 `Rules` 与 Service 的 `IsValid()` 双重把关（11.1）。
- `Messages` 是可选的：实现了 `MessagesProvider` 就会被 `param.Bind` 自动带上，没实现则回退框架默认文案。

### 15.6 Controller

```go
// Package controller 存放 system 模块的控制器（admin/controller），只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// XxxController 提供 system_xxx 的 CRUD 接口。
type XxxController struct {
	service *service.XxxService
}

// NewXxxController 构造 xxx 控制器。
func NewXxxController() *XxxController {
	return &XxxController{service: service.NewXxxService()}
}

// Index 分页查询 xxx，支持按关键字筛选。
// 入参：keyword、pageNo、pageSize（查询串）；返回列表与总数。
func (c *XxxController) Index(ctx http.Context) http.Response {
	var page param.Page
	page.PageNo = ctx.Request().QueryInt("pageNo", 1)
	page.PageSize = ctx.Request().QueryInt("pageSize", param.DefaultPageSize)
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), page.PageNo, page.PageSize, ctx.Request().Query("keyword"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增 xxx；名称非空且唯一。
// 入参：name、status、remark；返回新记录。
func (c *XxxController) Store(ctx http.Context) http.Response {
	var req request.XxxCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Create(ctx.Context(), req.Name, req.Status, req.Remark)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Show 查询单条 xxx 详情。
// 入参：路径 id；返回记录详情，不存在返回业务错误码。
func (c *XxxController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}
```

- 所有出口统一走 `result.Success` / `result.Page` / `result.Fail`，Controller 里**不出现 `ctx.Response()` 与 `response.*`**；HTTP 状态码始终 200（8.1），成功响应固定带 `data`，失败响应不带 `data`。
- **Controller 内不写 `fail`、`pathID`、响应拼装**：响应输出由 `base/result` 提供、参数与路径解析由 `base/param` 提供，避免每个模块复制一份、也避免包内重名冲突。
- 校验与绑定合并为 `param.Bind`，顺序固定为「先校验、再绑定」，两步失败都返回 `400`。

### 15.7 Route

管理端接口写在 `{module}/admin/route.go`（C 端写 `{module}/client/route.go`），模块根目录的 `route.go` 只做聚合：

```go
// admin/route.go
func Register() {
	xxxController := controller.NewXxxController()

	facades.Route().Prefix("/api/v1/system").Group(func(r route.Router) {
		r.Get("/xxxs", xxxController.Index)
		r.Post("/xxxs", xxxController.Store)
		r.Get("/xxxs/{id}", xxxController.Show)
		r.Put("/xxxs/{id}", xxxController.Update)
		r.Delete("/xxxs/{id}", xxxController.Destroy)
	})
}
```

```go
// route.go —— 只聚合
func Register() {
	admin.Register()
	client.Register() // 该模块还没有 C 端接口时省略这一行
}
```

后台路由追加 `.Middleware(operate("模块名", "操作名"))` 声明操作语义（9.5），并挂到 `permission` 组内；C 端路由挂 `member_auth`，不做接口级权限控制。

### 15.8 错误码

```go
// app/modules/system/errs/system_error.go
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

var (
	SystemXxxNotExists  = baseErrs.ErrorCode{Code: 1_002_0XX_001, Msg: "xx【%v】不存在"}
	SystemXxxNameEmpty  = baseErrs.ErrorCode{Code: 1_002_0XX_002, Msg: "名称不能为空"}
	SystemXxxNameExists = baseErrs.ErrorCode{Code: 1_002_0XX_003, Msg: "名称【%s】已存在"}
)
```

- 业务错误码一律落在**本模块**的 `errs` 包，base 只保留 0 ~ 999 的系统错误码（8.3(3)）；跨模块阅读时靠 `systemErrs` / `baseErrs` 别名区分。
- `1_002_0XX_001` 中的 `0XX` 换成该实体的子模块段（8.3(2) 已分配 001 ~ 008，新实体顺延）。
- 同一错误码的「预查重」与「1062 兜底」必须共用**同一个**常量（6.8(5)）。

### 15.9 迁移

无唯一键表的常规写法（Schema Builder）：

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"eqray-flavor/app/facades"
)

type M20261017000002CreateSystemXxxTable struct{}

func (r *M20261017000002CreateSystemXxxTable) Signature() string {
	return "20261017000002_create_system_xxx_table"
}

// Up 建 system_xxx 表；表已存在直接返回，保证重复执行幂等。
func (r *M20261017000002CreateSystemXxxTable) Up() error {
	if facades.Schema().HasTable("system_xxx") {
		return nil
	}

	return facades.Schema().Create("system_xxx", func(table schema.Blueprint) {
		table.BigInteger("id").Comment("主键") // 不用自增，应用侧雪花 ID
		table.String("name", 64).Comment("名称")
		table.TinyInteger("status").Default(1).Comment("状态：0 禁用 / 1 启用")
		table.String("remark", 500).Default("").Comment("备注")
		table.String("creator", 64).Default("").Comment("创建人")
		table.DateTime("created_at").Nullable().Comment("创建时间")
		table.String("updater", 64).Default("").Comment("更新人")
		table.DateTime("updated_at").Nullable().Comment("更新时间")
		table.DateTime("deleted_at").Nullable().Comment("删除时间")

		table.Primary("id")
		table.Index("created_at")

		table.Comment("xxx 表")
	})
}

// Down 回滚时删表，用 DropIfExists 避免表不存在时报错。
func (r *M20261017000002CreateSystemXxxTable) Down() error {
	return facades.Schema().DropIfExists("system_xxx")
}
```

唯一键直接建在业务键上。需要写原生 SQL（如生成列、复杂默认值）时用下面的模板，注意不要引入 `deleted_key`：

```go
// Up 用原生 SQL 建表，便于写唯一键与复杂默认值；表已存在直接返回。
func (r *M20261017000002CreateSystemXxxTable) Up() error {
	if facades.Schema().HasTable("system_xxx") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_xxx (
  id           bigint       NOT NULL COMMENT '主键',
  name         varchar(64)  NOT NULL COMMENT '名称',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  remark       varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  deleted_at   datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='xxx 表'`)
}

// Down 回滚时删表，与 Up 的建表动作对应。
func (r *M20261017000002CreateSystemXxxTable) Down() error {
	return facades.Schema().DropIfExists("system_xxx")
}
```

- `HasTable` 判断保留，保证重复执行幂等；`Down()` 必须 `DropIfExists`（6.7）。
- 建表顺序按 6.7 的依赖顺序插入，不要追加到文件末尾就算完成。
- 唯一键直接建在业务键上：Schema Builder 的 `table.Unique("uk_xxx", []string{"col"})` 或原生 SQL 均可（`system_users` 用的是后者）；关联表唯一键同样只含业务键，解除关系时物理删除（6.8(4)）。

### 15.10 与正文的关系

本章是 6.1、6.6、6.8(2)、6.9、8.1、8.2、8.4、11.1 的可执行版本：正文写规则，本章给落地写法。以下口径以本章代码为准：

| 主题 | 现行口径 | 落点 |
|---|---|---|
| 审计字段填充 | 构造器不触发钩子，Repository 写入前显式调 `FillCreate` / `FillUpdate` | 6.1、6.9、15.2 |
| 软删除 | 读写两侧都显式处理：查询带 `WhereNull("deleted_at")`，删除写 `deleted_at` | 6.6、6.9、15.3 |
| 写库形态 | 写用 `map[string]any` 显式列名，规避零值跳过；读用结构体按 `db` tag 映射 | 6.9、15.3 |
| Model 标签 | 只用 `db` tag，不用 `gorm` tag | 8.4、11.1、15.2 |
| 表名 | Model 导出 `XxxTable` 常量，Repository 引用它，不再重复字面量 | 6.7、15.2、15.3 |
| 业务错误码 | 落在**模块** `errs` 包，base 只放系统错误码 | 8.3(3)、11.2、15.8 |
| 响应结构 | 成功固定带 `data`（可为 `null`），失败省略 `data` | 8.1、15.6 |
| 分页 | `pageSize` 只接受 1 ~ 200，越界返回 `400`；分页接口不给 `-1` 开口子 | 8.2、15.3、15.6 |

后续若正文再出现与本章代码冲突的表述，以本章为准并同步回改正文。

### 15.11 新增实体自检清单

- [ ] Model 嵌入 `Base`，字段全带 `db` tag，主键/外键用 `types.ID`，表名以常量导出
- [ ] 状态类字段用 `enums` 里的自定义类型，不写裸 `int8`
- [ ] Repository 读用结构体、写用 `map`，所有查询带 `WhereNull("deleted_at")`
- [ ] 写前调 `FillCreate` / `FillUpdate`，Service 不碰审计字段
- [ ] 分页接口校验 `pageSize` 为 1 ~ 200，越界返回 `400`，不透传 `-1`
- [ ] 唯一键表的写入用 `support.IsDuplicate` 捕获 1062，与预查重共用同一错误码
- [ ] 软删除用 `RowsAffected` 判定「不存在」；普通更新不用它判断
- [ ] Request 有独立 `Rules` 与中文 `Messages`，更新入参字段同样必填
- [ ] Service 直接持有具体仓储类型，不预抽 interface（需要 mock 时再由测试就近定义）
- [ ] Controller 只做取参/调 Service/出响应，错误走 `result.Fail`
- [ ] 路由挂到正确前缀与中间件组，后台路由声明操作语义与权限标识
- [ ] 错误码落在**本模块** `errs` 文件，编码段与该实体子模块一致
- [ ] 迁移有 `HasTable` 幂等判断与 `DropIfExists` 的 `Down()`；唯一键直接建在业务键上，不使用 `deleted_key`；关联表解除关系用物理删除
- [ ] 表注释与每个字段的 `COMMENT` 都写了，状态类字段注明取值含义
