package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000049CreateStorePrintTaskTable 创建门店打印任务表。
// 背景：门店已经能配打印设备与打印模板，但没有任务表，下单后没有任何打印记录可查、可重打。
// 设计：一条任务 = 一台设备 × 一次打印；内容在生成时就渲染好，
// 避免事后模板被改导致重打内容与当初不一致。
type M20261023000049CreateStorePrintTaskTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000049CreateStorePrintTaskTable) Signature() string {
	return "20261023000049_create_store_print_task_table"
}

// Up 建表并补齐查询索引。
func (r *M20261023000049CreateStorePrintTaskTable) Up() error {
	if facades.Schema().HasTable("store_print_task") {
		return nil
	}

	statement := `CREATE TABLE store_print_task (
		id bigint unsigned NOT NULL COMMENT '主键',
		store_id bigint unsigned NOT NULL DEFAULT 0 COMMENT '门店 ID',
		printer_id bigint unsigned NOT NULL DEFAULT 0 COMMENT '打印设备 ID',
		printer_name varchar(64) NOT NULL DEFAULT '' COMMENT '设备名称快照',
		template_id bigint unsigned NOT NULL DEFAULT 0 COMMENT '打印模板 ID',
		template_version int NOT NULL DEFAULT 0 COMMENT '生成时的模板版本号',
		print_type tinyint NOT NULL DEFAULT 0 COMMENT '打印类型：1 结账单 / 2 预结单 / 3 厨打单 / 4 退菜单 / 5 标签 / 6 退款单 / 7 交班单 / 8 日结单',
		order_id bigint unsigned NOT NULL DEFAULT 0 COMMENT '订单 ID，非订单类打印为 0',
		order_no varchar(64) NOT NULL DEFAULT '' COMMENT '订单号快照',
		content mediumtext COMMENT '渲染后的打印正文，重打时直接复用',
		status tinyint NOT NULL DEFAULT 0 COMMENT '状态：0 待打印 / 1 已打印 / 2 打印失败 / 3 已取消',
		copies tinyint NOT NULL DEFAULT 1 COMMENT '打印份数',
		retry_count int NOT NULL DEFAULT 0 COMMENT '重打次数',
		error_msg varchar(255) NOT NULL DEFAULT '' COMMENT '失败原因',
		printed_at datetime NULL COMMENT '打印成功时间',
		creator varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
		created_at datetime NULL COMMENT '创建时间',
		updater varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
		updated_at datetime NULL COMMENT '更新时间',
		deleted_at datetime NULL COMMENT '删除时间',
		PRIMARY KEY (id),
		KEY idx_store_status (store_id, status),
		KEY idx_printer_status (printer_id, status),
		KEY idx_order (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店打印任务'`

	return facades.DB().Statement(statement)
}

// Down 删除打印任务表。
func (r *M20261023000049CreateStorePrintTaskTable) Down() error {
	if !facades.Schema().HasTable("store_print_task") {
		return nil
	}

	return facades.DB().Statement("DROP TABLE IF EXISTS store_print_task")
}
