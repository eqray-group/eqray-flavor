// Package promotion 聚合营销模块的路由。
// 管理端维护券模板、会员券与营销活动，C 端供会员查看与领取优惠券。
package promotion

import (
	"eqray-flavor/app/modules/promotion/admin"
	"eqray-flavor/app/modules/promotion/client"
)

// Register 注册营销模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
	client.Register()
}
