package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000004CreateStoreEmployeeTable 建 store_employee 门店员工关联表。
// 门店与后台账号多对多；同一门店内同一账号只允许一条记录，关系删除采用物理删除。
type M20261022000004CreateStoreEmployeeTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000004CreateStoreEmployeeTable) Signature() string {
	return "20261022000004_create_store_employee_table"
}

// Up 建 store_employee 表。
// 唯一键落在 (store_id, user_id) 业务键上，门店员工是数据权限的来源（design/03 §4）。
func (r *M20261022000004CreateStoreEmployeeTable) Up() error {
	if facades.Schema().HasTable("store_employee") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_employee (
  id         bigint       NOT NULL COMMENT '主键',
  store_id   bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  user_id    bigint       NOT NULL COMMENT '后台账号 ID，关联 system_users',
  role       tinyint      NOT NULL DEFAULT 4 COMMENT '门店角色：1 店长 / 2 收银 / 3 后厨 / 4 服务员 / 5 骑手',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '启停状态：0 禁用 / 1 启用；只有启用记录计入可见门店',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间；关联表实际采用物理删除，保留字段与基类一致',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_user (store_id, user_id),
  KEY idx_store_status (store_id, status),
  KEY idx_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门店员工关联表（员工与门店多对多）'`)
}

// Down 删除 store_employee 表。
func (r *M20261022000004CreateStoreEmployeeTable) Down() error {
	return facades.Schema().DropIfExists("store_employee")
}
