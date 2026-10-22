# AGENTS.md

面向 AI 助手的规则集。只写 agent 从代码里推不出来的约定；模板类内容看文中指向的真实文件。
完整设计见 `docs/framework-design.md`（冲突以它为准），业务设计在 `docs/design/`，
Goravel 文档本地副本在 `docs/framework/`（按需查，不要批量读）。用户当轮指令优先于本文件。

## 1. 速览

- 后端：Go 1.25 + Goravel 1.18（Gin）+ MySQL 8 + Redis，端口 `11111`。
- 前端：Vue 3.5 + Vite 7 + TS + Element Plus + Tailwind 4 + Pinia，在 `template/web/system-manage-ui/`，端口 `22222`（代理 `/api` → 11111）。
- 前端按形态分 `template/web/`（PC）与 `template/mobile/`（uni-app）：
  `web/system-manage-ui` 可用，`web/merchant-manage-ui`、`mobile/customer-uniapp`、`mobile/rider-uniapp` 是占位目录，对应模块开工前不要写代码。详见 `template/README.md`。
- 认证：账号密码 + 图形验证码，访问 / 刷新双令牌；RBAC 权限点 `模块:资源:动作`，`SUPER_ADMIN` 直接放行；菜单由后端 `system_menu` 下发（`VITE_ACCESS_MODE=backend`）。
- 初始账号 `admin` / `admin123`。
- 模块：`system`（认证、用户、角色、菜单、字典、登录/操作日志）、`infra`（参数、文件、文件配置、API 日志、短信）、`member`（会员、等级、积分、签到、地址、社交登录）、`store`（门店组织树、门店员工与可见门店数据权限、桌台区域与桌台、骑手档案、打印设备与模板）。
- 定时任务只在 `bootstrap/schedule.go`：每天 03:10 清理 API 访问日志（`bootstrap/app.go` 的 `WithSchedule` 装载）。任务后台页面已删除，不要再引入。

## 2. 命令

```bash
# 后端
go run .                              # http://127.0.0.1:11111
go run . artisan migrate              # 执行迁移（新文件必须先在 bootstrap/migrations.go 注册）
go run . artisan migrate:status
go build ./... && go vet ./...        # 提交前必跑

# 前端（先 cd template/web/system-manage-ui）
pnpm install
pnpm dev                              # http://127.0.0.1:22222
npx vue-tsc --noEmit                  # 类型检查
npx eslint "src/views/member" --fix   # 单目录校验 + 修复
```

代码风格以 `eslint.config.mjs` / `.prettierrc` / `.stylelintrc.cjs` 为准，不要把这些规则抄进本文。

## 3. 结构

```
app/modules/<模块>/
  admin/ | client/        # 接入层：controller + request，按端分目录
  route.go                # 只做聚合：admin.Register() / client.Register()
  service/ repository/ model/ errs/
bootstrap/                # app.go 装配、migrations.go 注册表、schedule.go 调度
database/migrations/      # 一个表一个 create_*，种子单独 seed_*
template/
  web/system-manage-ui/    # 管理后台：src/{api,views,types} 按 system / infra / member / store 分
  web/merchant-manage-ui/  # 商户端（占位）
  mobile/customer-uniapp/  # 顾客端（占位）
  mobile/rider-uniapp/     # 骑手端（占位，走后台账号）
```

| 模块 | 管理端前缀 | 管理端中间件 | C 端前缀 | C 端中间件 |
| --- | --- | --- | --- | --- |
| `system` | `/api/v1/system` | TokenAuth + Permission + OperateLog | — | — |
| `infra` | `/api/v1/infra` | 同上 | `/api/v1/client/infra`（仅上传） | MemberAuth |
| `member` | `/api/v1/member` | 同上 | `/api/v1/client/member` | MemberAuth |
| `store` | `/api/v1/store` | 同上 | — | — |

- 取当前人：管理端 `support.UserID`，C 端 `support.MemberID`。
- C 端 `request` **禁止出现 memberId**，按会员过滤一律由 service 用令牌里的 ID 强制带条件。
- C 端不参与 RBAC，也不写 `system_operate_log`。

**决策表**

| 问题 | 选择 |
| --- | --- |
| 接口给后台用户，还是会员本人？ | 后台 → `admin/` + TokenAuth + `perm`；会员本人 → `client/` + MemberAuth |
| 短信、文件这类通用能力放哪？ | 放 `infra`，其他模块调它的 service，不复制实现 |
| 会员令牌怎么签？ | `baseAuth.IssueScoped(ScopeMember, ...)`，不在 member 里自建 auth |
| 列表页表格用什么？ | `ArtTable` + `useTable`，不用原生 `ElTable` |

## 4. 后端规则

### 4.1 分层

| 层 | 做 | 不做 |
| --- | --- | --- |
| `controller` | 取参、`param.Bind`/`param.PathID`、调 service、`result.*` | 业务规则、直接查库 |
| `service` | 校验、组合仓储、事务、缓存 | 依赖 `http.Context` |
| `repository` | 查询，返回 model | 业务分支 |
| `request` | 结构体 + `Rules()` + `Messages()`（中文提示必填） | 访问数据库 |
| `errs` | `baseErrs.ErrorCode{Code, Msg}` | 拼接文案 |

### 4.2 新增管理端接口

1. `admin/request/xxx_request.go`：入参 + `Rules` + 中文 `Messages`。
2. `service/xxx_service.go`：业务方法，错误用 `baseErrs.New(code, args...)`。
3. `repository`：需要新查询时才加。
4. `admin/controller/xxx_controller.go`：绑定 → 调 service → `result.Success/Fail/Page`。
5. `admin/route.go` 挂 `perm("模块:资源:动作")`；写操作加 `operate("类型", "子类型")`。
6. 错误码写进对应 `errs/*_error.go`。
7. 新表/新字段：加迁移并注册（见 4.6）。
8. 新页面：补菜单与权限点种子（见 4.7）。
9. `go build ./... && go vet ./...`。

端的分工：管理端 controller/request 放 `admin/`，C 端放 `client/`；service/repository/model/errs 两端共用。

### 4.3 路由与权限

```go
facades.Route().Prefix("/api/v1/member").Group(func(r route.Router) {
    auth := r.Middleware(&baseMiddleware.TokenAuth{})
    perm := func(code string) route.Router {
        return auth.Middleware(&baseMiddleware.Permission{Code: code})
    }
    perm("member:tag:list").Get("/tags", tagController.Index)
})
```

- 路径参数用 `{id}` + `param.PathID(ctx)`。
- 静态路径必须排在动态路径前：`/levels/options` 要在 `/levels/{id}` 之前。
- 权限点动作：`list/create/update/delete`，特殊动作另起（`infra:api-log:update`、`member:user:points`）。
- 上传属自助能力：后台 `POST /api/v1/infra/files` 只挂 TokenAuth，C 端 `POST /api/v1/client/infra/files` 挂 MemberAuth。

### 4.4 响应、参数与错误码

- `result.Success(ctx, data)` / `result.Fail(ctx, err)` / `result.Page(ctx, list, total)`；业务错误 HTTP 200，`code != 0`。
- 分页一律 `param.FromQuery(ctx)`（默认 10 条），**不要手拼 `param.Page{}` 再 Normalize**。
- 错误码 `1_{模块}_{子模块}_{序号}`：模块 `001` infra、`002` system、`003` member、`004` store。
- 子模块号登记在各 `errs/*_error.go` 文件头，新增前先看占用情况（member：user 001 / level 002 / tag 003 / group 004 / social 005 / auth 006 / address 007 / sign 008；store：store 001 / table 002 / employee 003 / printer 004 / rider 005）。
- 中间件里 `Abort` 的中断必须调 `result.MarkFail(ctx, code)` 或 `result.AbortFail(ctx, err)`，否则错误日志 `message` 为空。

### 4.5 迁移

- 文件名 `YYYYMMDDHHMMSS_snake_case.go`，`Signature()` 必须与文件名一致（不含 `.go`）。
- **必须在 `bootstrap/migrations.go` 注册**，否则 `migrate` 提示 `Nothing to migrate`。
- 一个表一个 `create_*_table.go`，只写最终结构；改结构新增 `alter_*`，不在旧文件上改。
- 种子（菜单、账号、默认等级、短信渠道、签到档位）独立 `seed_*.go`，幂等（先查再插）。
- 建表用 `facades.DB().Statement("CREATE TABLE ...")`，`Up` 先 `HasTable` 判空。
- 已执行的迁移不要改；删文件不会撤销 DB 变更，需另写清理迁移并删 `migrations` 记录。
- 改菜单/权限后前移缓存：`facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)`。

### 4.6 菜单与权限点

`system_menu`：`type` 1 目录 / 2 菜单 / 3 按钮。新增页面需插页面菜单 + 按钮权限点，并授权内置角色
（`system_role_menu`），写法照 `database/migrations/20261020000037_seed_member_sign_config_menu.go`。

- `component` 填 `/member/sign-config`，前端解析到 `src/views/member/sign-config/index.vue`；填错即白屏。
- 页面 `name` 按层级加前缀：`/member/sign-config` → `MemberSignConfig`。
- 查询类下拉接口（如 `/levels/options`）可复用列表权限，不必新建权限点。

### 4.7 文件存储

- `infra_file_config.storage` 支持 `local` / `s3`，主配置唯一由业务层保证（`master` 无唯一约束）。
- `config` JSON：local `{"dir":"uploads"}`；s3 `{"bucket","endpoint","region","accessKey","secretKey","url","usePathStyle","dir"}`。
- 落盘 `{dir}/YYYY/MM/DD/{雪花ID}.{ext}`，原文件名存 `infra_file.name`；local 地址 `/storage/{path}`，s3 用 `driver.Url(path)`。
- 删物理文件按 `infra_file.config_id` 找原配置的磁盘，配置缺失才回落 local。

### 4.8 会员

- 令牌：`baseAuth.IssueScoped(baseAuth.ScopeMember, ...)`，不要在 member 里再包一层；`ScopeAdmin`/`ScopeMember` 隔离，Redis 键 `auth:member:access:<hash>`。
- 第三方登录：渠道配置只存 `member_social_client`（后台 `member:social-client:*`），不读环境变量；`/auth/social/login` 未绑定直接报错，**不自动注册**；解绑至少保留一个绑定且物理删除。
- 手机号登录：6 位码、Redis 5 分钟、一次性、60 秒重发限制、每号每日 10 条；`sms-login` 手机号不存在时自动注册（昵称「用户 + 后 4 位」）。
- 短信能力在 infra：`infraSmsService.Send(ctx, templateCode, mobile, values)`；member 依赖 infra，其余模块互不依赖。
- 签到奖励读 `member_sign_in_config`：`day` 1~31、同日唯一、至少保留一条启用档位；按连续天数取第一个 ≥ 的档位，超出封顶取最高档。
- 签到同时加积分与经验，经验达门槛自动升级（只升不降）并写 `member_level_record`；幂等由 `uk_user_date` 兜底。
- 后台档位接口 `/api/v1/member/sign-configs`，权限点 `member:sign-config:*`。
- C 端自助接口都在 `memberAuth` 分组：`/profile`、`/addresses`（5 个）、`/sign` + `/sign/status` + `/sign/records`、`/points/records`、`/level/records`、`/levels`。
- 只读接口先过 `MemberService.EnsureActive`，禁用会员报 `MemberDisabled`。

### 4.9 门店与数据权限

- 门店级接口统一读请求头 `X-Store-Id`，兼容查询串 `storeId`；中间件在
  `app/modules/store/middleware/store_context.go`，通过后写 `support.StoreID(ctx)`。
  上下文只能指向未删除的 `type=3` 门店，总部 / 大区节点会被拒绝。
- 未带门店标识不拦截：服务层按当前用户可见门店集合过滤，空集合返回空列表。
- 可见范围：超管全量；**没有任何员工关联**视为总部账号、全量；存在任何关联（含禁用）
  即受限；只有启用员工记录 + 未删除 `type=3` 门店有效。
- 缓存键 `store:visible:<userID>`，TTL 30 分钟；编码 `a` 全量、`r:1,2` 受限、
  `r:` 受限但为空；员工增删改后必须调 `StoreService.InvalidateScope(userID)`。
- 门店上下文接口的请求体门店必须与 `support.StoreID` 一致，否则报
  `StoreContextMismatch`；员工新增未传门店时按上下文回填，编辑未传时沿用原门店。
- `store_employee` 物理删除，唯一键 `(store_id,user_id)`；角色枚举存 `role`，
  字典 `store_role_type`；员工权限点 `store:employee:list/create/update/delete`。
- 骑手档案 `store_rider` 与员工一对一，`employee_id` 唯一，软删除；状态 `1` 空闲 /
  `2` 接单中 / `3` 配送中 / `4` 休息 / `5` 停用，字典 `store_rider_status`；权限点
  `store:rider:list/create/update/delete`。抢单 / 释放名额只走仓储条件更新
  （`TryAccept` / `Release`），编辑停用状态与删除也要带 `current_orders` 条件，不允许先查再改。
- 骑手自动开通账号通过 `support.AccountProvisioner`（system 实现），手机号为账号，
  初始密码只在创建响应返回一次；自动账号不分配 system 角色，骑手端接口不依赖 RBAC。
- 停用骑手同步禁用员工；有进行中订单时禁止休息 / 停用 / 删除；并发上限不能低于当前负载。
  员工页不能修改已关联骑手档案的门店 / 账号 / 角色 / 状态，也不能直接删除员工。
- 打印设备 `store_printer` 按门店归属，唯一键 `(store_id,device_no)`，物理删除；`config`
  只存非敏感 JSON，`secret` 存 AES 密文、`secret_mask` 存末四位脱敏值。列表 / 编辑
  只返回 `secretMask` 与 `hasSecret`，禁止新增回显明文的接口；编辑 `secret` 留空表示
  沿用原密钥。权限点 `store:printer:list/create/update/delete`。
- 打印模板 `store_print_template` 唯一键 `(store_id,printer_type,print_type)`，物理删除；
  `version` 新建为 1，只有正文变化才 +1，客户端不能提交版本号。权限点
  `store:print-template:list/create/update/delete`。打印任务、厂商适配与模板快照属
  T-PRINT-01/02，不要在 store 模块提前建任务表。
- 桌台区域 `store_table_area` 按门店归属，唯一键 `(store_id,name)`，物理删除；
  区域下仍有桌台时禁止删除。权限点 `store:table-area:list/create/update/delete`。
- 桌台 `store_table` 的 `code` 全局唯一、`(store_id,name)` 唯一，物理删除；桌码默认
  按「门店编码-桌号」生成且创建后不可改。批量生成按「前缀 + 补零序号」，单次最多
  100 张，补零宽度取整批最大序号位数且至少两位；权限点
  `store:table:list/create/update/delete/operate`。
- 桌台状态 `1` 空闲 / `2` 使用中 / `3` 待清台，字典 `table_status`；开台仅允许空闲，
  清台允许使用中或待清台；使用中或仍有 `current_order_id` 的桌台禁止删除。
- `/api/v1/store/stores/visible` 只挂 TokenAuth，供切换器使用，不要求门店列表权限。
- 前端门店上下文在 `useStoreContextStore`（持久化 `currentStoreId` / `currentStoreName`），
  HTTP 拦截器注入请求头；退出登录与切换账号时清空；切换后刷新当前页。

### 4.10 注释

每个函数/方法都要文档注释：**做什么 + 关键约束 + 入参 + 返回**，中文描述业务语义。

```go
// Index 分页查询后台用户，支持账号、昵称、手机号与状态筛选。
// 入参：username、nickname、mobile、status、pageNo、pageSize（查询串）；返回用户列表与总数。
func (c *UserController) Index(ctx http.Context) http.Response {
```

- 禁止 `// Index GET /api/v1/system/users` 这类复述路由的注释。
- 私有函数、构造函数、迁移 `Up`/`Down` 也要一句说明。
- 前端函数上方写 `// 中文说明`，包含触发时机与副作用。

## 5. 前端规则

### 5.1 列表页

必须 `ArtSearchBar` + `ElCard.art-table-card` + `ArtTableHeader` + `ArtTable` + `useTable`，不要写原生 `ElTable`。

- 不加「序号」列（全站约定）。
- 行内操作 `ArtButtonTable`（`type: edit/delete/action/view`），权限用 `hasAuth('模块:资源:动作')`；页面按钮 `v-auth`。
- 新增/编辑/删除后分别调 `refreshCreate` / `refreshUpdate` / `refreshRemove`；工具条刷新用 `refreshData`。
- 参考：`src/views/member/sign-config/index.vue`、`src/views/infra/file-config/index.vue`。

新增页面流程：`src/api/<模块>/xxx.ts` 封装接口 → `src/types/api/<模块>.d.ts` 加 `Api.<模块>` 类型 →
`src/views/<模块>/xxx/index.vue` 写列表页（`modules/` 放弹窗）→ 后端补菜单 + 权限点 seed →
重新登录验证菜单与按钮。

### 5.2 弹窗

- `v-model:visible` + `type: 'add' | 'edit'` + `xxx-data` 传当前行；成功 `emit('success')`，父页面按类型刷新。
- `watch(visible)` 回填并 `clearValidate()`；关闭确认用 `try { await ElMessageBox.confirm(...) } catch { return }`。
- 请求错误提示由 http 层统一弹出，`catch` 留空并注释，不要重复 `ElMessage.error`。
- 参考：`src/views/member/sign-config/modules/sign-config-dialog.vue`。

### 5.3 API 与类型

- 接口封装在 `src/api/{system,infra,member,store}/*.ts`，`request.get<T>` 返回后端 `data`，POST/PUT 用 `params`（http 层转 JSON）。
- 雪花 ID 一律 `string`。
- 类型按模块拆在 `src/types/api/{common,auth,infra,store,member,dashboard,system}.d.ts`，各自 `declare namespace Api { namespace Xxx { ... } }`，声明合并到全局 `Api`，使用处不写 import。system 模块的命名空间是 **`SystemManage`**，不是 `System`。

### 5.4 路由名与样式

- 路由由后端菜单驱动，`MenuProcessor` 按层级加前缀（`System` + `User` → `SystemUser`），页面 `defineOptions({ name })` 必须与路由名一致，否则 KeepAlive 缓存残留。
- Tailwind 4 原子类为主；次要文字 `text-xs text-g-600`；页面根容器加 `art-full-height`；CSS 变量用 Element 变量。

## 6. 已知坑

1. **Goravel 三参数 `Where` 会编译成 `id IN ('!=', 123)`**：用 `Where("id <> ?", id)`、`WhereLike("name", "%"+kw+"%")`、`WhereIn("id", []any{...})`。
2. **`SUM/AVG/MAX` 空集返回 NULL**：用 `Select("COALESCE(SUM(size), 0) AS size")` + `First(&row)`（`Get` 只接受 slice）。
3. **`RowsAffected == 0` 不等于记录不存在**（新旧值相同）：再查一次存在性，参考 `FileConfigRepository.SetMaster`。
4. **`v-auth` 同时读 `user.info.buttons` 与 `meta.authList`**：后端动态菜单没有 `authList`，只读它会隐藏全部按钮。
5. **权限缓存 30 分钟**：改菜单/权限后要前移 `rbac:version`，前端还需重新登录清本地路由。
6. **中间件 `Abort` 不调 `MarkFail` / `AbortFail`** → 错误日志 `message` 为空。
7. **S3/MinIO**：`usePathStyle=true`，`url` 前缀决定访问地址，Bucket 未公开读会 403。

## 7. 排查表

| 现象 | 先查 |
| --- | --- |
| 菜单白屏 | `system_menu.component` 是否为 `/模块/页面`；`src/views/模块/页面/index.vue` 是否存在 |
| 菜单不显示 | 迁移是否注册执行；`system_role_menu` 是否授权；`rbac:version`；是否重新登录 |
| 按钮不显示 | 权限点是否在 `system_menu.permission`；角色是否授权；`info.buttons` 是否刷新 |
| 接口 403 | 路由 `perm(...)` 与前端权限点是否一致 |
| 接口 500 且日志无信息 | 是否 `Abort` 未 `MarkFail` / `AbortFail`；看 `storage/logs` |
| 文件配置出现多个主配置 | `ClearMaster` 是否用了 `Where("id <> ?", id)` |
| 上传 404 / 图片 403 | local 查 `/storage` 静态路由；s3 查 `url` 前缀与 Bucket 策略 |
| 签到报「档位未配置」 | `member_sign_in_config` 是否有 `status=1` 档位；`20261020000036` 是否执行 |
| 打印设备号重复绑定 | 同一 `store_id` 是否已有该 `device_no`（`uk_store_device`）；请求头门店是否正确 |
| 打印接口出现明文密钥 | 列表 DTO 是否误带 `Secret`；模型必须保持 `json:"-"`，且不要新增回显明文的接口 |
| 模板版本没有递增 | 只有 `content` 变化才 +1；仅改名称 / 纸宽 / 状态不递增 |

## 8. 提交清单

**后端**

- [ ] 请求参数有 `Rules` + 中文 `Messages`
- [ ] 业务错误用错误码，没有硬编码文案
- [ ] 路由挂了正确的 `perm`，权限点已写入菜单种子
- [ ] 迁移已注册到 `bootstrap/migrations.go` 且幂等
- [ ] 菜单/权限变更已前移 `rbac:version`
- [ ] 新增/修改的函数都有中文文档注释（说什么、入参、返回）
- [ ] `go build ./...`、`go vet ./...` 通过

**前端**

- [ ] 列表页用 `ArtSearchBar + ArtTableHeader + ArtTable + useTable`，无「序号」列
- [ ] 权限：页面按钮 `v-auth`，行内操作 `hasAuth`
- [ ] 增删改后调用对应 `refreshCreate / refreshUpdate / refreshRemove`
- [ ] 弹窗 `v-model:visible` + `emit('success')`，类型写在 `Api.{模块}`
- [ ] 在 `template/web/system-manage-ui/` 下 `npx vue-tsc --noEmit`、`eslint` 通过
