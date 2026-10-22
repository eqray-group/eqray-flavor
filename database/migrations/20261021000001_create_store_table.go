package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261021000001CreateStoreTable 建 store 门店表。
// 一张表承载总部 / 大区 / 门店三级组织树：parent_id 自引用，type 区分层级。
type M20261021000001CreateStoreTable struct{}

// Signature The unique signature for the migration.
func (r *M20261021000001CreateStoreTable) Signature() string {
	return "20261021000001_create_store_table"
}

// Up 建 store 表。
// code 是外部对接与门店切换的稳定标识，建唯一键保证不会重复。
func (r *M20261021000001CreateStoreTable) Up() error {
	if facades.Schema().HasTable("store") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store (
  id             bigint        NOT NULL COMMENT '主键',
  parent_id      bigint        NOT NULL DEFAULT 0 COMMENT '上级节点 ID，0 表示根节点（总部）',
  type           tinyint       NOT NULL DEFAULT 3 COMMENT '节点类型：1 总部 / 2 大区 / 3 门店',
  code           varchar(32)   NOT NULL COMMENT '门店编码，全局唯一，创建后不可修改',
  name           varchar(64)   NOT NULL COMMENT '门店名称',
  short_name     varchar(32)   NOT NULL DEFAULT '' COMMENT '门店简称，小票与看板展示用',
  status         tinyint       NOT NULL DEFAULT 1 COMMENT '经营状态：1 营业 / 2 休息 / 3 装修 / 4 停业',
  contact_name   varchar(32)   NOT NULL DEFAULT '' COMMENT '联系人',
  contact_mobile varchar(20)   NOT NULL DEFAULT '' COMMENT '联系电话',
  district_id    bigint        NOT NULL DEFAULT 0 COMMENT '区县 ID，省市区通过该 ID 反查；0 表示未选择',
  address        varchar(255)  NOT NULL DEFAULT '' COMMENT '详细地址',
  longitude      decimal(10,7) NOT NULL DEFAULT 0 COMMENT '经度，配送范围校验用',
  latitude       decimal(10,7) NOT NULL DEFAULT 0 COMMENT '纬度，配送范围校验用',
  open_time      char(5)       NOT NULL DEFAULT '09:00' COMMENT '营业开始时间 HH:mm，如 09:00',
  close_time     char(5)       NOT NULL DEFAULT '23:00' COMMENT '营业结束时间 HH:mm；小于开始时间表示跨夜，相等表示全天',
  delivery_rule  varchar(1000) NOT NULL DEFAULT '' COMMENT '配送规则 JSON：{"radius":3000,"minAmount":2000,"deliveryFee":500,"packingFee":100,"etaMinutes":40}，半径米、金额分、时长分钟',
  auto_accept    tinyint       NOT NULL DEFAULT 0 COMMENT '自动接单：1 外卖/自提单支付成功即自动接单并转制作中，0 需人工接单；堂食不受影响',
  online_enabled tinyint       NOT NULL DEFAULT 1 COMMENT '是否开放线上：0 否 / 1 是',
  sort           int           NOT NULL DEFAULT 0 COMMENT '同级排序，越小越靠前',
  remark         varchar(255)  NOT NULL DEFAULT '' COMMENT '备注',
  creator        varchar(64)   NOT NULL DEFAULT '' COMMENT '创建人',
  created_at     datetime      NULL COMMENT '创建时间',
  updater        varchar(64)   NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at     datetime      NULL COMMENT '更新时间',
  deleted_at     datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code),
  KEY idx_parent (parent_id),
  KEY idx_type_status (type, status),
  KEY idx_district (district_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门店表（总部 / 大区 / 门店组织树）'`)
}

// Down 删除 store 表。
func (r *M20261021000001CreateStoreTable) Down() error {
	return facades.Schema().DropIfExists("store")
}
