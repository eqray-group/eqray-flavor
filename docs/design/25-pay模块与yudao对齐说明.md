# 25 · 支付模块与 yudao 对齐说明

> 关联：[12 枚举与字典](./12-枚举与数据字典总表.md)、[03 数据模型 §8](./03-数据模型设计.md)、[22 进度](./22-开发进度与待办.md)
>
> 基线：`ruoyi-vue-pro`（yudao）commit `5656a0b`，核对文件：
> `enums/order/PayOrderStatusEnum.java`、`enums/refund/PayRefundStatusEnum.java`、
> `enums/notify/PayNotifyStatusEnum.java`、`enums/notify/PayNotifyTypeEnum.java`、
> `dal/dataobject/order/PayOrderDO.java`、`dal/dataobject/refund/PayRefundDO.java`、
> `dal/dataobject/notify/PayNotifyTaskDO.java`、`enums/DictTypeConstants.java`。
>
> 结论先行：**表结构与核心状态机已对齐；差异集中在版本选型（单商户 vs 多商户）与钱包扩展。**

---

## 1. 为什么会有「对齐」这件事

支付模块最初是精简版（两张表 + 环境变量配置），2026-10-22 整体重写为对齐 yudao 的支付中心，
表与字段大体照抄，但**枚举取值被改过**：订单状态多了一个「已退款」、退款状态用了 0/10/20、
通知状态用了 0/10/20。这些值一旦落库再改，历史数据就要写迁移翻译，因此趁订单模块未开工一次性校正。

---

## 2. 已对齐项

### 2.1 枚举取值（本次修正）

| 枚举 | yudao | 修正前 | 修正后 |
| --- | --- | --- | --- |
| `PayOrderStatusEnum` | 0 未支付 / 10 支付成功 / 20 支付关闭 | 0 / 10 / **20 已退款** / **30 已关闭** | 0 待支付 / 10 支付成功 / 20 支付关闭 |
| `PayRefundStatusEnum` | 0 生成 / 1 成功 / 2 失败 / 99 关闭 | 0 / **10** / **20**，无 99 | 0 / 1 / 2 / 99 |
| `PayNotifyStatusEnum` | 1 等待 / 2 成功 / 3 失败 / 4 请求成功但结果失败 / 5 请求失败 | **0 / 10 / 20**，只有 3 个 | 1 / 2 / 3 / 4 / 5 |
| `PayNotifyTypeEnum` | 1 支付单 / 2 退款单 | 1 / 2 / **3 转账** | 1 / 2 |
| `PayEnableStatusEnum` | 0 开启 / 1 关闭 | 一致 | 不变 |

**支付单不再有「已退款」状态**：yudao 的全额退款只累加 `refund_price`，状态停在「支付成功」，
是否退完用 `refund_price >= price` 判断。`pay_order.status = 20` 只表示未付款关闭或主动关单。
原来 `MarkRefunded` 会把退完的单置为「已退款」，已删除该逻辑。

### 2.2 通知任务机制（本次修正）

| 项 | yudao `PayNotifyTaskDO` | 本项目 |
| --- | --- | --- |
| 重试频率 | `NOTIFY_FREQUENCY = [15, 15, 30, 180, 1800, 1800, 1800, 3600]` 秒 | 一致 |
| 最大通知次数 | 8 档 + 首次 = 9 次 | 一致 |
| 状态语义 | 区分「请求失败」与「请求成功但结果失败」 | 一致（4 / 5） |
| 终态 | 成功、最终失败 | 一致，且终态不再被调度捞出 |

`dispatch` 判定规则：业务方响应 `success` → 2；HTTP 2xx 但响应不是 success → 4；
网络异常或非 2xx → 5；次数用尽 → 3。

### 2.3 表结构

12 张表与 yudao 同名同义：`pay_app`、`pay_channel`、`pay_order`、`pay_order_extension`、
`pay_refund`、`pay_notify_task`、`pay_notify_log`、`pay_transfer`、`pay_wallet`、
`pay_wallet_transaction`、`pay_wallet_recharge_package`、`pay_wallet_recharge`。

核心字段命名保持一致：`merchant_order_id`、`channel_order_no`、`channel_user_id`、
`channel_fee_rate`、`channel_fee_price`、`refund_price`、`extension_id`、`no`、`user_ip`、
`expire_time`、`success_time`、`channel_extras`、`channel_error_code/msg`、`channel_notify_data`。

### 2.4 退款并发防护（本次补强）

yudao 在 `updateOrderRefundPrice` 里校验「已退 + 本次 ≤ 支付金额」，但那是先查再比。
本项目改为条件更新：`WHERE refund_price + ? <= price`，返回 0 行即拒绝，
并发两笔退款同时通过前置校验时不会退穿。对应错误码 `1-008-004-007`。

### 2.5 配置口径

渠道的商户号、密钥、证书全部存 `pay_channel.config`（JSON），按渠道主键取用；
`config/pay.go` 只保留 `pay.notify_base_url`。原文件里的 `wechat.*` / `alipay.*` / `expire_minutes`
已经没有任何代码读取，本次删除，避免有人照着配 `.env` 却发现不生效。

---

## 3. 有意保留的差异

| 差异 | yudao | 本项目 | 理由 |
| --- | --- | --- | --- |
| 商户层级 | 最新版有 `pay_merchant`，订单/退款/通知都带 `merchant_id` | 无商户表，`pay_app` 即商户 | 单品牌单商户，多一层商户只会增加联表成本；对齐的是经典单商户版 |
| 订单通知状态 | `pay_order.notify_status` + 字典 `pay_order_notify_status` | 由 `pay_notify_task` 表独立表达 | 用任务表能看次数、下次时间与响应原文，比订单上一个状态字段更有用 |
| 订单退款状态 | `pay_order.refund_status` + `refund_times` | 只有 `refund_price` | 退款进度由退款单列表 + 累计金额表达，避免两处状态不一致 |
| 退款单类型 | `type`（部分 / 全部）、`trade_no`、`channel_extras`、`expire_time`、`notify_time` | 未建这些列 | 当前退款金额已能表达部分/全部；等接会员退款审批时再补 `type` |
| 转账 | `PayTransferService` + 后台页 + 转账接口 | 只有表、模型与仓储，未开放接口 | 一期无提现/打款场景，先留结构与 yudao 对齐，等业务确认再做 |
| 通知类型 | 只有 1 / 2 | 一致（已删 3 转账） | 转账不发通知，与 yudao 相同 |
| 钱包 | 当前 master 的 pay 模块已无 wallet 数据对象 | 自建 4 张表 + 余额支付渠道 `wallet` | 会员储值是本项目必需能力；`biz_type` 采用数值枚举（1 充值 / 2 充值退款 / 3 支付 / 4 支付退款 / 5 手工调整），比原字符串 `refund` 混用更利于对账 |

> 钱包的 `wallet` 渠道是本项目扩展：不写 `pay_channel`，下单时由支付单服务直接转钱包服务扣余额，
> 并落一张已支付的支付单。yudao 无此渠道编码，前端渠道列表里会多一项「钱包余额」。

---

## 4. 数据库迁移（本次新增）

| 迁移 | 作用 |
| --- | --- |
| `20261023000011_alter_pay_wallet_transaction_biz_type` | 钱包流水 `biz_type` 由字符串翻译为数值并改列类型 |
| `20261023000012_align_pay_status_values` | 翻译历史状态值：订单 20→10、30→20；退款 10→1、20→2；通知 0→1、10→2、20→3 |
| `20261023000013_add_pay_composite_indexes` | 补 `(app_id, merchant_order_id)`、`(app_id, merchant_refund_id)` 复合索引 |
| `20261023000014_seed_pay_dicts` | 补 5 个支付字典（此前只在文档里，库里没有） |
| `20261023000015_widen_pay_order_text_columns` | `subject` 32→128、`body` 128→255、`channel_extras` 256→1024 |

**执行顺序不能改**：状态值翻译必须在代码上线前跑完，否则新旧语义会混在一张表里。

`20261023000012` 的顺序也有讲究：先把旧 `20`（已退款）改成 `10`，再把旧 `30`（已关闭）改成 `20`；
反过来会把旧的已关闭单误判成已支付。

---

## 5. 仍然存在的风险

1. **退款单没有唯一约束**。`(app_id, merchant_refund_id)` 目前是普通索引，幂等靠代码先查。
   并发提交同一商户退款单号可能落两条（yudao 同样没有唯一键）。若要彻底杜绝，
   需要去掉软删除或改用「删除时间置 0」的唯一键技巧，不要在没想清楚前加唯一索引。
2. **支付单同理**。`merchant_order_id` 在应用内应当唯一，但表上是普通索引。
3. **转账是半成品**：表、模型、仓储齐全，但没有 service、路由与页面。接业务前不要对外承诺。
4. **真实商户号未联调**：微信平台证书序列号、支付宝验签开关、回调地址可达性都还是纸面验证。
5. **钱包余额支付没有独立幂等键**：靠 `merchant_order_id` 查重 + 支付单事务，重复提交同一单号会被拒，
   但钱包扣款与支付单落库在同一事务内，若渠道侧（无渠道）异常需人工核对。

---

## 6. 改动清单（便于 review）

后端：

- `app/modules/pay/enums/pay.go`：四个枚举重写，删除 `PayOrderStatusRefund` 与 `PayNotifyTypeTransfer`
- `app/modules/pay/model/`：`pay_order` / `pay_refund` / `pay_notify_task` / `pay_wallet_transaction` 注释与类型
- `app/modules/pay/repository/`：`MarkRefunded` 加累计校验并去掉改状态；`FindByMerchantOrderID` 不再过滤状态；
  `ListDueTasks` 按新状态集合捞取；`biz_type` 数值化
- `app/modules/pay/service/`：退款三处累计校验、通知状态判定与重试频率、钱包业务类型
- `app/modules/pay/errs/pay_error.go`：新增 `1-008-004-007`
- `config/pay.go`：只留 `notify_base_url`
- 新增 `app/modules/pay/service/pay_core_test.go`：9 组用例锁定枚举值、重试频率、金额校验、脱敏、签名

前端：

- `views/pay/order`、`views/pay/refund`、`views/pay/notify`、`views/pay/notify/modules/notify-log-drawer`、
  `views/pay/order/modules/refund-create-dialog`：状态选项与判断改为新取值
- `views/pay/wallet/modules/wallet-transaction-drawer`：业务类型下拉与展示改为数值映射
- `types/api/pay.d.ts`：`bizType` 改为 number

文档：

- `AGENTS.md`：补 dish / pay 模块、新增 §4.10 支付约定
- `docs/design/12`：支付字典改为新取值并写明三个易踩点
- `docs/design/22`：支付模块状态描述改为支付中心
