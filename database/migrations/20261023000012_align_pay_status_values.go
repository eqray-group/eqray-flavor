package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000012AlignPayStatusValues 把支付模块历史状态值翻译成 yudao 口径。
// 旧实现：pay_order 0 待支付 / 10 已支付 / 20 已退款 / 30 已关闭；
//
//	pay_refund 0 等待 / 10 成功 / 20 失败；
//	pay_notify_task、pay_notify_log 0 等待 / 10 成功 / 20 失败。
//
// yudao 口径：pay_order 0 未支付 / 10 支付成功 / 20 支付关闭（全额退款不改状态，只累加 refund_price）；
//
//	pay_refund 0 退款订单生成 / 1 退款成功 / 2 退款失败 / 99 退款关闭；
//	pay_notify 1 等待 / 2 成功 / 3 失败 / 4 请求成功但结果失败 / 5 请求失败。
//
// 修正顺序不可颠倒：pay_order 必须先把 20（旧已退款）改成 10，再把 30（旧已关闭）改成 20，
// 否则旧关闭单会被误判成已退款。
type M20261023000012AlignPayStatusValues struct{}

// Signature The unique signature for the migration.
func (r *M20261023000012AlignPayStatusValues) Signature() string {
	return "20261023000012_align_pay_status_values"
}

// Up 翻译支付单、退款单与通知任务 / 日志的历史状态值。
func (r *M20261023000012AlignPayStatusValues) Up() error {
	if facades.Schema().HasTable("pay_order") {
		// 旧 20（已退款）在新口径里仍是「支付成功」，退款进度由 refund_price 表达
		if err := facades.DB().Statement("UPDATE pay_order SET status = 10 WHERE status = 20"); err != nil {
			return err
		}
		// 旧 30（已关闭）→ 新 20（支付关闭）
		if err := facades.DB().Statement("UPDATE pay_order SET status = 20 WHERE status = 30"); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_refund") {
		if err := facades.DB().Statement("UPDATE pay_refund SET status = 1 WHERE status = 10"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE pay_refund SET status = 2 WHERE status = 20"); err != nil {
			return err
		}
	}

	for _, table := range []string{"pay_notify_task", "pay_notify_log"} {
		if !facades.Schema().HasTable(table) {
			continue
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 1 WHERE status = 0"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 2 WHERE status = 10"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 3 WHERE status = 20"); err != nil {
			return err
		}
	}

	return nil
}

// Down 把状态值还原成旧口径，仅用于回滚到旧实现。
func (r *M20261023000012AlignPayStatusValues) Down() error {
	if facades.Schema().HasTable("pay_order") {
		if err := facades.DB().Statement("UPDATE pay_order SET status = 30 WHERE status = 20"); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_refund") {
		if err := facades.DB().Statement("UPDATE pay_refund SET status = 20 WHERE status = 2"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE pay_refund SET status = 10 WHERE status = 1"); err != nil {
			return err
		}
	}

	for _, table := range []string{"pay_notify_task", "pay_notify_log"} {
		if !facades.Schema().HasTable(table) {
			continue
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 20 WHERE status = 3"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 10 WHERE status = 2"); err != nil {
			return err
		}
		if err := facades.DB().Statement("UPDATE " + table + " SET status = 0 WHERE status = 1"); err != nil {
			return err
		}
	}

	return nil
}
