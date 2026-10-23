package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000015WidenPayOrderTextColumns 放宽支付单与拓展单的文本列长度。
// 建表时 subject / body 用了 32 / 128，但请求校验与 yudao 的商品标题上限是 128，
// 业务侧传长标题会直接写库失败；拓展单的 channel_extras 要存二维码链接与调起参数，
// 256 字符在支付宝 WAP 场景下不够。这里统一放宽，与代码校验口径对齐。
type M20261023000015WidenPayOrderTextColumns struct{}

// Signature The unique signature for the migration.
func (r *M20261023000015WidenPayOrderTextColumns) Signature() string {
	return "20261023000015_widen_pay_order_text_columns"
}

// Up 调整三处列长度；MODIFY 幂等，重复执行同样成立。
func (r *M20261023000015WidenPayOrderTextColumns) Up() error {
	if facades.Schema().HasTable("pay_order") {
		if err := facades.DB().Statement(`
ALTER TABLE pay_order
  MODIFY COLUMN subject varchar(128) NOT NULL COMMENT '商品标题'`); err != nil {
			return err
		}
		if err := facades.DB().Statement(`
ALTER TABLE pay_order
  MODIFY COLUMN body varchar(255) NOT NULL DEFAULT '' COMMENT '商品描述'`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_order_extension") {
		if err := facades.DB().Statement(`
ALTER TABLE pay_order_extension
  MODIFY COLUMN channel_extras varchar(1024) NOT NULL DEFAULT ''
  COMMENT '渠道额外参数 JSON，如二维码链接、调起参数'`); err != nil {
			return err
		}
	}

	return nil
}

// Down 还原为建表时的长度，仅用于回滚；已有超长数据会被截断或报错，需先自行清理。
func (r *M20261023000015WidenPayOrderTextColumns) Down() error {
	if facades.Schema().HasTable("pay_order") {
		if err := facades.DB().Statement(`
ALTER TABLE pay_order
  MODIFY COLUMN subject varchar(32) NOT NULL COMMENT '商品标题'`); err != nil {
			return err
		}
		if err := facades.DB().Statement(`
ALTER TABLE pay_order
  MODIFY COLUMN body varchar(128) NOT NULL DEFAULT '' COMMENT '商品描述'`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_order_extension") {
		return facades.DB().Statement(`
ALTER TABLE pay_order_extension
  MODIFY COLUMN channel_extras varchar(256) NOT NULL DEFAULT '' COMMENT '渠道额外参数 JSON'`)
	}

	return nil
}
