package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000047AddPayRefundAuditFields 给退款单补审核相关字段。
// 背景：退款此前是「提交即执行」，店员点一下钱就出去了。
// 这里加一层审核，但不改 yudao 的退款状态机：
// status 仍是 0 生成 / 1 成功 / 2 失败 / 99 关闭，
// 审核进度单独用 audit_status 表达，审核通过才真正提交渠道。
type M20261023000047AddPayRefundAuditFields struct{}

// Signature The unique signature for the migration.
func (r *M20261023000047AddPayRefundAuditFields) Signature() string {
	return "20261023000047_add_pay_refund_audit_fields"
}

// Up 给 pay_refund 增加审核状态、审核时间、审核人与审核意见。
func (r *M20261023000047AddPayRefundAuditFields) Up() error {
	if !facades.Schema().HasTable("pay_refund") {
		return nil
	}

	columns := []struct {
		name      string
		statement string
	}{
		{
			name:      "audit_status",
			statement: "ALTER TABLE pay_refund ADD COLUMN audit_status tinyint NOT NULL DEFAULT 0 COMMENT '审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回'",
		},
		{
			name:      "audit_time",
			statement: "ALTER TABLE pay_refund ADD COLUMN audit_time datetime NULL COMMENT '审核时间'",
		},
		{
			name:      "audit_by",
			statement: "ALTER TABLE pay_refund ADD COLUMN audit_by bigint NOT NULL DEFAULT 0 COMMENT '审核人后台用户 ID'",
		},
		{
			name:      "audit_remark",
			statement: "ALTER TABLE pay_refund ADD COLUMN audit_remark varchar(255) NOT NULL DEFAULT '' COMMENT '审核意见 / 驳回原因'",
		},
	}

	for _, column := range columns {
		if facades.Schema().HasColumn("pay_refund", column.name) {
			continue
		}
		if err := facades.DB().Statement(column.statement); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除本次新增的四个列。
func (r *M20261023000047AddPayRefundAuditFields) Down() error {
	if !facades.Schema().HasTable("pay_refund") {
		return nil
	}

	for _, name := range []string{"audit_status", "audit_time", "audit_by", "audit_remark"} {
		if !facades.Schema().HasColumn("pay_refund", name) {
			continue
		}
		if err := facades.DB().Statement("ALTER TABLE pay_refund DROP COLUMN " + name); err != nil {
			return err
		}
	}

	return nil
}
