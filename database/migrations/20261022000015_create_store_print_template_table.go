package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000015CreateStorePrintTemplateTable 建 store_print_template 打印模板表。
// 模板按门店归属，同一门店下设备类型与打印类型唯一；正文变化时版本号递增。
type M20261022000015CreateStorePrintTemplateTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000015CreateStorePrintTemplateTable) Signature() string {
	return "20261022000015_create_store_print_template_table"
}

// Up 建 store_print_template 表。
func (r *M20261022000015CreateStorePrintTemplateTable) Up() error {
	if facades.Schema().HasTable("store_print_template") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_print_template (
  id           bigint       NOT NULL COMMENT '主键',
  store_id     bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  name         varchar(64)  NOT NULL COMMENT '模板名称，便于运营区分',
  printer_type tinyint      NOT NULL COMMENT '适用设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号',
  print_type   tinyint      NOT NULL COMMENT '业务打印类型：1 结账单 / 2 预结单 / 3 厨打 / 4 退菜 / 5 标签 / 6 退款单 / 7 交班 / 8 日结',
  paper_width  int          NOT NULL DEFAULT 80 COMMENT '纸宽，单位毫米，常用 58 / 76 / 80',
  content      mediumtext   NOT NULL COMMENT '模板正文，打印渲染器按版本解释',
  version      int          NOT NULL DEFAULT 1 COMMENT '当前模板版本号，新建为 1，正文变化后递增',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '启停状态：0 禁用 / 1 启用',
  remark       varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  deleted_at   datetime     NULL COMMENT '删除时间；实际采用物理删除，保留字段与基类一致',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_kind (store_id, printer_type, print_type),
  KEY idx_store_print_status (store_id, print_type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打印模板表（门店各机型小票模板）'`)
}

// Down 删除 store_print_template 表。
func (r *M20261022000015CreateStorePrintTemplateTable) Down() error {
	return facades.Schema().DropIfExists("store_print_template")
}
