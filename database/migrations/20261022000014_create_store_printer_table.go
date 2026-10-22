package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000014CreateStorePrinterTable 建 store_printer 打印设备表。
// 设备按门店归属，同一门店内设备号唯一；密钥只保存 AES 密文与脱敏值。
type M20261022000014CreateStorePrinterTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000014CreateStorePrinterTable) Signature() string {
	return "20261022000014_create_store_printer_table"
}

// Up 建 store_printer 表。
func (r *M20261022000014CreateStorePrinterTable) Up() error {
	if facades.Schema().HasTable("store_printer") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_printer (
  id          bigint        NOT NULL COMMENT '主键',
  store_id    bigint        NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  name        varchar(64)   NOT NULL COMMENT '设备名称，如前台小票机、后厨打印机',
  device_no   varchar(64)   NOT NULL COMMENT '设备号，同一门店内唯一',
  type        tinyint       NOT NULL DEFAULT 1 COMMENT '设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号',
  access_mode tinyint       NOT NULL DEFAULT 1 COMMENT '接入方式：1 云打印机 / 2 局域网 / 3 本地客户端',
  config      varchar(2000) NOT NULL DEFAULT '' COMMENT '非敏感配置 JSON，如厂商、地址、端口；禁止写入密钥',
  secret      varchar(2048) NOT NULL DEFAULT '' COMMENT '设备密钥的 AES 密文，禁止明文存储',
  secret_mask varchar(32)   NOT NULL DEFAULT '' COMMENT '密钥脱敏展示值，如 ****abcd',
  copies      int           NOT NULL DEFAULT 1 COMMENT '默认打印份数，1 ~ 9',
  status      tinyint       NOT NULL DEFAULT 1 COMMENT '启停状态：0 禁用 / 1 启用',
  remark      varchar(255)  NOT NULL DEFAULT '' COMMENT '备注',
  creator     varchar(64)   NOT NULL DEFAULT '' COMMENT '创建人',
  created_at  datetime      NULL COMMENT '创建时间',
  updater     varchar(64)   NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at  datetime      NULL COMMENT '更新时间',
  deleted_at  datetime      NULL COMMENT '删除时间；实际采用物理删除，保留字段与基类一致',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_device (store_id, device_no),
  KEY idx_store_type_status (store_id, type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印设备表（门店打印设备与密钥配置）'`)
}

// Down 删除 store_printer 表。
func (r *M20261022000014CreateStorePrinterTable) Down() error {
	return facades.Schema().DropIfExists("store_printer")
}
