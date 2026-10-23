// Package admin 注册菜品模块的后台管理接口。
// 菜品是品牌级数据，接口只挂后台令牌与权限点，不参与门店上下文校验。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/dish/admin/controller"
	storeMiddleware "eqray-flavor/app/modules/store/middleware"
)

// Prefix 是菜品后台接口前缀。
const Prefix = "/api/v1/dish"

// Register 注册菜品后台接口，由 dish.Register() 统一调用。
func Register() {
	categoryController := controller.NewCategoryController()
	spuController := controller.NewSpuController()
	addonController := controller.NewAddonController()
	storeSkuController := controller.NewStoreSkuController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件在 base，dish 只声明自己的权限点
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 菜品分类 ----------
		// 分类是树结构，一次返回整棵树，不再单独提供详情接口
		perm("dish:category:list").Get("/categories/tree", categoryController.Tree)
		perm("dish:category:create").Middleware(operate("菜品管理", "新增菜品分类")).
			Post("/categories", categoryController.Store)
		perm("dish:category:update").Middleware(operate("菜品管理", "编辑菜品分类")).
			Put("/categories/{id}", categoryController.Update)
		perm("dish:category:delete").Middleware(operate("菜品管理", "删除菜品分类")).
			Delete("/categories/{id}", categoryController.Destroy)

		// ---------- 菜品 ----------
		// 规格与加料组随菜品一起提交，因此不单独开放 SKU 接口
		perm("dish:spu:list").Get("/spus", spuController.Index)
		perm("dish:spu:list").Get("/spus/{id}", spuController.Show)
		perm("dish:spu:create").Middleware(operate("菜品管理", "新增菜品")).
			Post("/spus", spuController.Store)
		perm("dish:spu:update").Middleware(operate("菜品管理", "编辑菜品")).
			Put("/spus/{id}", spuController.Update)
		perm("dish:spu:update").Middleware(operate("菜品管理", "菜品上下架")).
			Put("/spus/{id}/status", spuController.UpdateStatus)
		perm("dish:spu:delete").Middleware(operate("菜品管理", "删除菜品")).
			Delete("/spus/{id}", spuController.Destroy)

		// ---------- 加料组与加料项 ----------
		perm("dish:addon:list").Get("/addon-groups", addonController.Index)
		perm("dish:addon:list").Get("/addon-groups/{id}", addonController.Show)
		perm("dish:addon:create").Middleware(operate("菜品管理", "新增加料组")).
			Post("/addon-groups", addonController.Store)
		perm("dish:addon:update").Middleware(operate("菜品管理", "编辑加料组")).
			Put("/addon-groups/{id}", addonController.Update)
		perm("dish:addon:delete").Middleware(operate("菜品管理", "删除加料组")).
			Delete("/addon-groups/{id}", addonController.Destroy)
		perm("dish:addon:create").Middleware(operate("菜品管理", "新增加料项")).
			Post("/addon-groups/{id}/items", addonController.StoreItem)
		perm("dish:addon:update").Middleware(operate("菜品管理", "编辑加料项")).
			Put("/addon-items/{id}", addonController.UpdateItem)
		perm("dish:addon:delete").Middleware(operate("菜品管理", "删除加料项")).
			Delete("/addon-items/{id}", addonController.DestroyItem)

		// ---------- 门店可售与门店价 ----------
		// 门店由请求头 X-Store-Id 决定，统一经过门店上下文中间件校验可见范围
		perm("dish:store-sku:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/store-skus", storeSkuController.Index)
		perm("dish:store-sku:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("菜品管理", "保存门店菜品配置")).Put("/store-skus/{id}", storeSkuController.Save)
		perm("dish:store-sku:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("菜品管理", "批量调整门店菜品")).Post("/store-skus/batch", storeSkuController.Batch)
	})
}
