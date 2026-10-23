// Package service 实现 order 模块的业务层：试算、下单、状态机与支付对接。
// 订单与支付解耦：订单侧发起支付走 pay 模块；支付结果由 pay 的签名回调通知订单。
package service

import (
	"context"
	"time"

	dishRepository "eqray-flavor/app/modules/dish/repository"
	orderRepository "eqray-flavor/app/modules/order/repository"
	storeRepository "eqray-flavor/app/modules/store/repository"
)

// OrderService 是订单模块的业务入口，聚合门店 / 菜品 / 订单三个仓储。
// 试算与下单共用 resolveLines 装配口径，保证「试算价 = 下单价」。
type OrderService struct {
	stores     *storeRepository.StoreRepository
	tables     *storeRepository.StoreTableRepository
	skus       *dishRepository.DishSkuRepository
	spus       *dishRepository.DishSpuRepository
	dishStores *dishRepository.DishStoreRepository
	addons     *dishRepository.DishAddonRepository
	orders     *orderRepository.CaterOrderRepository
}

// NewOrderService 构造订单服务。
func NewOrderService() *OrderService {
	return &OrderService{
		stores:     storeRepository.NewStoreRepository(),
		tables:     storeRepository.NewStoreTableRepository(),
		skus:       dishRepository.NewDishSkuRepository(),
		spus:       dishRepository.NewDishSpuRepository(),
		dishStores: dishRepository.NewDishStoreRepository(),
		addons:     dishRepository.NewDishAddonRepository(),
		orders:     orderRepository.NewCaterOrderRepository(),
	}
}

// 定时任务与回调使用的上下文超时上限，避免单次调度拖死连接。
const orderServiceTimeout = 5 * time.Minute

// newTimeoutContext 供调度与回调入口构造带超时的上下文。
func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), orderServiceTimeout)
}
