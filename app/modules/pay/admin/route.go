// Package admin 注册支付模块的后台管理接口。
// 支付应用、渠道与支付单一律挂后台令牌与权限点，写操作留痕。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/pay/admin/controller"
)

// Prefix 是支付后台接口前缀。
const Prefix = "/api/v1/pay"

// Register 注册支付后台接口，由 pay.Register() 统一调用。
func Register() {
	appController := controller.NewPayAppController()
	channelController := controller.NewPayChannelController()
	orderController := controller.NewPayOrderController()
	refundController := controller.NewPayRefundController()
	notifyTaskController := controller.NewPayNotifyTaskController()
	walletController := controller.NewPayWalletController()
	walletPackageController := controller.NewPayWalletPackageController()
	walletRechargeController := controller.NewPayWalletRechargeController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件在 base，pay 只声明自己的权限点
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 支付应用 ----------
		// 静态路径必须排在动态路径之前，否则 /apps/options 会被 /apps/{id} 吃掉
		perm("pay:app:list").Get("/apps/options", appController.Options)
		perm("pay:app:list").Get("/apps", appController.Index)
		perm("pay:app:list").Get("/apps/{id}", appController.Show)
		perm("pay:app:create").Middleware(operate("支付管理", "新增支付应用")).
			Post("/apps", appController.Store)
		perm("pay:app:update").Middleware(operate("支付管理", "编辑支付应用")).
			Put("/apps/{id}", appController.Update)
		perm("pay:app:delete").Middleware(operate("支付管理", "删除支付应用")).
			Delete("/apps/{id}", appController.Destroy)

		// ---------- 支付渠道 ----------
		// 渠道配置里的密钥只在写入时接收明文，列表与详情均只返回脱敏值。
		perm("pay:channel:list").Get("/channels/options", channelController.Options)
		perm("pay:channel:list").Get("/channels", channelController.Index)
		perm("pay:channel:list").Get("/channels/{id}", channelController.Show)
		perm("pay:channel:create").Middleware(operate("支付管理", "新增支付渠道")).
			Post("/channels", channelController.Store)
		perm("pay:channel:update").Middleware(operate("支付管理", "编辑支付渠道")).
			Put("/channels/{id}", channelController.Update)
		perm("pay:channel:delete").Middleware(operate("支付管理", "删除支付渠道")).
			Delete("/channels/{id}", channelController.Destroy)

		// ---------- 支付单 ----------
		// 创建即调起渠道；金额只认服务端试算结果，前端无法改写应付金额。
		// 导出是静态路径，必须排在 /orders/{id} 之前
		perm("pay:order:list").Get("/orders/export", orderController.Export)
		perm("pay:order:list").Get("/orders", orderController.Index)
		perm("pay:order:list").Get("/orders/{id}", orderController.Show)
		perm("pay:order:create").Middleware(operate("支付管理", "创建支付单")).
			Post("/orders", orderController.Store)
		perm("pay:order:submit").Middleware(operate("支付管理", "重新调起支付")).
			Post("/orders/{id}/submit", orderController.Submit)
		// 补单：主动查渠道，把回调丢失的支付单补成已支付
		perm("pay:order:sync").Middleware(operate("支付管理", "同步支付单")).
			Post("/orders/{id}/sync", orderController.Sync)
		perm("pay:order:close").Middleware(operate("支付管理", "关闭支付单")).
			Post("/orders/{id}/close", orderController.Close)

		// ---------- 退款 ----------
		// 退款金额与可退额度由服务层校验，前端不能绕过；退款结果以渠道回调为准
		perm("pay:refund:list").Get("/refunds/export", refundController.Export)
		perm("pay:refund:list").Get("/refunds", refundController.Index)
		perm("pay:refund:list").Get("/refunds/{id}", refundController.Show)
		perm("pay:refund:create").Middleware(operate("支付管理", "发起退款")).
			Post("/refunds", refundController.Store)
		// 审核与发起拆开授权：能提申请的人不一定能放款
		perm("pay:refund:audit").Middleware(operate("退款管理", "审核通过退款")).
			Put("/refunds/{id}/approve", refundController.Approve)
		perm("pay:refund:audit").Middleware(operate("退款管理", "驳回退款")).
			Put("/refunds/{id}/reject", refundController.Reject)

		// ---------- 回调通知任务 ----------
		// 只读：任务由支付 / 退款成功时自动登记，派发靠定时任务，后台不提供手工重发
		perm("pay:notify:list").Get("/notify-tasks", notifyTaskController.Index)
		perm("pay:notify:list").Get("/notify-tasks/{id}/logs", notifyTaskController.Logs)

		// ---------- 会员钱包 ----------
		// 余额调整走条件更新 + 流水，后台只能调整金额，不能直接改余额字段
		// 门店充值：静态路径必须排在 /wallets/{id} 之前，且用独立权限点与「调账」区分
		perm("pay:wallet:recharge").Get("/wallets/member/{memberId}", walletController.MemberWallet)
		perm("pay:wallet:recharge").Middleware(operate("支付管理", "门店现金充值")).
			Post("/wallets/member/{memberId}/cash-recharge", walletController.CashRecharge)
		perm("pay:wallet:list").Get("/wallets", walletController.Index)
		perm("pay:wallet:list").Get("/wallets/{id}", walletController.Show)
		perm("pay:wallet:list").Get("/wallets/{id}/transactions", walletController.Transactions)
		perm("pay:wallet:list").Get("/wallets/{id}/transactions/export", walletController.ExportTransactions)
		perm("pay:wallet:update").Middleware(operate("支付管理", "调整钱包余额")).
			Post("/wallets/{id}/balance", walletController.ChangeBalance)

		// ---------- 充值套餐 ----------
		perm("pay:wallet-package:list").Get("/wallet-packages/options", walletPackageController.Options)
		perm("pay:wallet-package:list").Get("/wallet-packages", walletPackageController.Index)
		perm("pay:wallet-package:create").Middleware(operate("支付管理", "新增充值套餐")).
			Post("/wallet-packages", walletPackageController.Store)
		perm("pay:wallet-package:update").Middleware(operate("支付管理", "编辑充值套餐")).
			Put("/wallet-packages/{id}", walletPackageController.Update)
		perm("pay:wallet-package:delete").Middleware(operate("支付管理", "删除充值套餐")).
			Delete("/wallet-packages/{id}", walletPackageController.Destroy)

		// ---------- 充值记录 ----------
		perm("pay:wallet-recharge:list").Get("/wallet-recharges", walletRechargeController.Index)
		perm("pay:wallet-recharge:list").Get("/wallet-recharges/{id}", walletRechargeController.Show)
	})
}
