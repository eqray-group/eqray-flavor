// Package service 实现 order 模块的业务层：试算、下单、状态机与支付对接。
// 订单与支付解耦：订单侧发起支付走 pay 模块；支付结果由 pay 的签名回调通知订单。
package service

import (
	dishRepository "eqray-flavor/app/modules/dish/repository"
	orderRepository "eqray-flavor/app/modules/order/repository"
	promotionService "eqray-flavor/app/modules/promotion/service"
	storeRepository "eqray-flavor/app/modules/store/repository"
	storeService "eqray-flavor/app/modules/store/service"
)

// OrderService 是订单模块的业务入口，聚合门店 / 菜品 / 订单 / 营销四个仓储。
// 试算与下单共用 resolveLines 装配口径，保证「试算价 = 下单价」。
type OrderService struct {
	stores     *storeRepository.StoreRepository
	tables     *storeRepository.StoreTableRepository
	riders     *storeRepository.StoreRiderRepository
	employees  *storeRepository.StoreEmployeeRepository
	skus       *dishRepository.DishSkuRepository
	spus       *dishRepository.DishSpuRepository
	dishStores *dishRepository.DishStoreRepository
	addons     *dishRepository.DishAddonRepository
	orders     *orderRepository.CaterOrderRepository
	promotions *promotionService.PromotionService
	// printTasks 支付成功后生成门店打印任务（结账单 + 厨打单）
	printTasks *storeService.StorePrintTaskService
}

// NewOrderService 构造订单服务。
func NewOrderService() *OrderService {
	return &OrderService{
		stores:     storeRepository.NewStoreRepository(),
		tables:     storeRepository.NewStoreTableRepository(),
		riders:     storeRepository.NewStoreRiderRepository(),
		employees:  storeRepository.NewStoreEmployeeRepository(),
		skus:       dishRepository.NewDishSkuRepository(),
		spus:       dishRepository.NewDishSpuRepository(),
		dishStores: dishRepository.NewDishStoreRepository(),
		addons:     dishRepository.NewDishAddonRepository(),
		orders:     orderRepository.NewCaterOrderRepository(),
		promotions: promotionService.NewPromotionService(),
		printTasks: storeService.NewStorePrintTaskService(),
	}
}
