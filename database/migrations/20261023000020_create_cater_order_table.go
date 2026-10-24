package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000020CreateCaterOrderTable 创建订单主表 cater_order。
// 状态机：100 待支付 / 200 待接单 / 300 制作中 / 400 待取餐 / 500 已完成 / 600 已取消 / 700 退款中 / 800 已退款；
// pay_status 与 pay 模块对齐：0 未支付 / 10 已支付 / 20 已退款。
// 幂等：(store_id, request_no) 唯一，request_no 缺省时服务端生成，保证全渠道可兜底。
type M20261023000020CreateCaterOrderTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000020CreateCaterOrderTable) Signature() string {
	return "20261023000020_create_cater_order_table"
}

// Up 创建 cater_order 表。
func (r *M20261023000020CreateCaterOrderTable) Up() error {
	if facades.Schema().HasTable("cater_order") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE cater_order (
  id                bigint        NOT NULL COMMENT '订单编号',
  no                varchar(32)   NOT NULL COMMENT '订单号，对外展示',
  store_id          bigint        NOT NULL COMMENT '门店 ID',
  member_id         bigint        NOT NULL DEFAULT 0 COMMENT '会员 ID，0 表示散客',
  type              tinyint       NOT NULL COMMENT '订单类型：1 堂食 / 2 外卖 / 3 自提',
  channel           tinyint       NOT NULL COMMENT '下单渠道：1 小程序 / 2 POS / 3 自助机 / 4 外卖平台',
  table_id          bigint        NOT NULL DEFAULT 0 COMMENT '堂食桌台 ID，非堂食为 0',
  table_name        varchar(64)   NOT NULL DEFAULT '' COMMENT '桌台名称快照',
  request_no        varchar(64)   NOT NULL COMMENT '幂等请求号，客户端缺省时服务端生成',
  status            smallint      NOT NULL DEFAULT 100 COMMENT '订单状态：100 待支付 / 200 待接单 / 300 制作中 / 400 待取餐 / 500 已完成 / 600 已取消 / 700 退款中 / 800 已退款',
  pay_status        tinyint       NOT NULL DEFAULT 0 COMMENT '支付状态：0 未支付 / 10 已支付 / 20 已退款',
  item_amount       bigint        NOT NULL DEFAULT 0 COMMENT '商品小计（分），不含加料，按成交单价计',
  addon_amount      bigint        NOT NULL DEFAULT 0 COMMENT '加料费合计（分）',
  line_discount_amount bigint     NOT NULL DEFAULT 0 COMMENT '单品级优惠合计（分），来自单品券 / 加价购',
  activity_amount   bigint        NOT NULL DEFAULT 0 COMMENT '活动优惠合计（分）',
  coupon_amount     bigint        NOT NULL DEFAULT 0 COMMENT '优惠券抵扣（分）',
  points_amount     bigint        NOT NULL DEFAULT 0 COMMENT '积分抵扣（分）',
  points_used       int           NOT NULL DEFAULT 0 COMMENT '实际使用的积分数：下单时按当时兑换比例折算并落库',
  adjust_amount     bigint        NOT NULL DEFAULT 0 COMMENT '改价差额（分）：正数加价、负数减免',
  pay_amount        bigint        NOT NULL DEFAULT 0 COMMENT '应付金额（分）',
  refund_amount     bigint        NOT NULL DEFAULT 0 COMMENT '已退款金额（分）',
  discount_snapshot text          NULL COMMENT '优惠计算过程快照 JSON，可完整复现计价结果',
  pickup_code       varchar(8)    NOT NULL DEFAULT '' COMMENT '取餐码',
  receiver_name     varchar(64)   NOT NULL DEFAULT '' COMMENT '收货人姓名，外卖订单使用',
  receiver_phone    varchar(20)   NOT NULL DEFAULT '' COMMENT '收货人电话',
  receiver_address  varchar(255)  NOT NULL DEFAULT '' COMMENT '收货地址',
  rider_id          bigint        NOT NULL DEFAULT 0 COMMENT '配送骑手 ID，0 表示未派单',
  dispatch_time     datetime      NULL COMMENT '派单时间',
  remark            varchar(255)  NOT NULL DEFAULT '' COMMENT '订单备注（顾客填写）',
  merchant_remark   varchar(255)  NOT NULL DEFAULT '' COMMENT '商家备注，与顾客备注分列存放',
  pay_order_id      bigint        NOT NULL DEFAULT 0 COMMENT '关联支付单 ID',
  pay_time          datetime      NULL COMMENT '支付成功时间',
  accept_time       datetime      NULL COMMENT '接单时间',
  complete_time     datetime      NULL COMMENT '完成时间',
  cancel_time       datetime      NULL COMMENT '取消时间',
  cancel_reason     varchar(255)  NOT NULL DEFAULT '' COMMENT '取消 / 拒单原因',
  expire_time       datetime      NOT NULL COMMENT '待支付超时时间，超时由定时任务关单',
  creator           varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at        datetime      NULL COMMENT '创建时间',
  updater           varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at        datetime      NULL COMMENT '更新时间',
  deleted_at        datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_no (no),
  UNIQUE KEY uk_store_request (store_id, request_no),
  KEY idx_store_status (store_id, status, created_at),
  KEY idx_member (member_id, created_at),
  KEY idx_store_pickup (store_id, pickup_code),
  KEY idx_pay_order (pay_order_id),
  KEY idx_status_expire (status, expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单主表'`)
}

// Down 删除 cater_order 表。
func (r *M20261023000020CreateCaterOrderTable) Down() error {
	return facades.Schema().DropIfExists("cater_order")
}
