package migrations

import (
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000073RemovePayMenusBeforeYudao 清空旧版支付模块的页面与按钮。
// 支付模块按 yudao 重构期间接口与页面全部下线，先撤掉菜单避免白屏；
// 新页面（应用、渠道、支付单、退款单、通知、转账、钱包、充值套餐）随实现完成后再补种子。
type M20261022000073RemovePayMenusBeforeYudao struct{}

// Signature The unique signature for the migration.
func (r *M20261022000073RemovePayMenusBeforeYudao) Signature() string {
	return "20261022000073_remove_pay_menus_before_yudao"
}

// Up 删除 /pay 下全部页面、页面下的按钮以及 pay: 前缀的权限点，并前移权限缓存版本。
func (r *M20261022000073RemovePayMenusBeforeYudao) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	ids := make([]types.ID, 0, 32)

	// 支付目录下的全部页面
	pages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("type", 2).
		WhereIn("path", []any{"/pay/app", "/pay/channel", "/pay/order", "/pay/refund", "/pay/notify-task", "/pay/transfer", "/pay/wallet", "/pay/recharge-package"}).
		WhereNull("deleted_at").
		Get(&pages); err != nil {
		return err
	}
	for _, page := range pages {
		ids = append(ids, page.ID)
	}

	// 旧版页面组件路径兜底（历史路径与新版 path 不完全一致）
	legacyPages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("type", 2).
		WhereLike("component", "/pay/%").
		WhereNull("deleted_at").
		Get(&legacyPages); err != nil {
		return err
	}
	for _, page := range legacyPages {
		ids = append(ids, page.ID)
	}

	// pay: 前缀的按钮权限点
	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("type", 3).
		Where(func(q db.Query) db.Query {
			return q.WhereLike("permission", "pay:%")
		}).
		WhereNull("deleted_at").
		Get(&buttons); err != nil {
		return err
	}
	for _, button := range buttons {
		ids = append(ids, button.ID)
	}

	// 页面下的子按钮
	if len(ids) > 0 {
		values := make([]any, 0, len(ids))
		for _, id := range ids {
			values = append(values, id)
		}

		children := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").
			Select("id").
			WhereIn("parent_id", values).
			WhereNull("deleted_at").
			Get(&children); err != nil {
			return err
		}
		for _, child := range children {
			ids = append(ids, child.ID)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// 去重后统一删除授权与菜单
	seen := make(map[types.ID]bool, len(ids))
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		values = append(values, id)
	}

	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").WhereIn("menu_id", values).Delete(); err != nil {
			return err
		}
	}

	if _, err := facades.DB().Table("system_menu").WhereIn("id", values).Delete(); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// Down 不恢复旧版支付菜单。
func (r *M20261022000073RemovePayMenusBeforeYudao) Down() error {
	return nil
}
