// order_pos_kds_service.go 提供后厨 KDS 看板数据。
// 一次返回「待接单 / 制作中 / 待取餐」的全部订单与明细，避免看板逐单查详情的 N+1。
package service

import (
	"context"
	"fmt"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
)

// kdsOrderLimit 是看板单次返回的订单上限，防止异常状态下把整表拉出来。
const kdsOrderLimit = 200

// KdsItemRow 是 KDS 卡片上的一行菜。
type KdsItemRow struct {
	// SpuName 菜品名
	SpuName string `json:"spuName"`
	// SkuName 规格名，无规格时为空
	SkuName string `json:"skuName"`
	// Quantity 数量
	Quantity int `json:"quantity"`
	// Remark 明细备注，如「少辣」
	Remark string `json:"remark"`
	// Orderer 点菜人，堂食多人点菜时使用
	Orderer string `json:"orderer"`
	// Addons 加料文案，已拼成「加芝士x2」形式
	Addons []string `json:"addons"`
}

// KdsOrderRow 是 KDS 看板上的一张订单卡。
type KdsOrderRow struct {
	ID         types.ID       `json:"id"`
	No         string         `json:"no"`
	Status     int16          `json:"status"`
	StatusText string         `json:"statusText"`
	Type       int8           `json:"type"`
	TypeText   string         `json:"typeText"`
	TableName  string         `json:"tableName"`
	PickupCode string         `json:"pickupCode"`
	Remark     string         `json:"remark"`
	CreatedAt  types.DateTime `json:"createdAt"`
	// PayTime 支付时间：后厨计时以它为准，未支付时为 null
	PayTime *types.DateTime `json:"payTime"`
	Items   []KdsItemRow    `json:"items"`
}

// KdsBoard 返回门店当前在做的订单（待接单 / 制作中 / 待取餐），含明细与加料。
// 入参：门店与操作人；返回按订单 ID 升序排列的看板数据。
// 说明：只读接口，不做状态过滤之外的业务判断；超时判定交给前端按支付时间计算，
// 避免服务端和看板的时钟口径不一致。
func (s *OrderService) KdsBoard(ctx context.Context, storeID, userID types.ID) ([]KdsOrderRow, error) {
	if storeID == 0 {
		return nil, errs.New(orderErrs.OrderPosStoreRequired)
	}
	if userID > 0 {
		if err := s.ensureStoreVisible(ctx, userID, storeID); err != nil {
			return nil, err
		}
	}

	statuses := []any{
		int16(orderEnums.OrderStatusPendingAccept),
		int16(orderEnums.OrderStatusMaking),
		int16(orderEnums.OrderStatusWaitingPickup),
	}
	orders, err := s.orders.ListActiveByStore(ctx, storeID, statuses, kdsOrderLimit)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return make([]KdsOrderRow, 0), nil
	}

	orderIDs := make([]types.ID, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	itemsByOrder, err := s.orders.ListItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}
	addonsByOrder, err := s.orders.ListAddonsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	rows := make([]KdsOrderRow, 0, len(orders))
	for _, order := range orders {
		row := KdsOrderRow{
			ID:         order.ID,
			No:         order.No,
			Status:     int16(order.Status),
			StatusText: order.Status.String(),
			Type:       int8(order.Type),
			TypeText:   order.Type.String(),
			TableName:  order.TableTitle,
			PickupCode: order.PickupCode,
			Remark:     order.Remark,
			CreatedAt:  order.CreatedAt,
			PayTime:    order.PayTime,
			Items:      make([]KdsItemRow, 0, len(itemsByOrder[order.ID])),
		}

		// 加料按明细归组，卡片上每条菜自己带加料，不串行
		addonsByItem := make(map[types.ID][]string, len(itemsByOrder[order.ID]))
		for _, addon := range addonsByOrder[order.ID] {
			text := addon.AddonName
			if addon.Quantity > 1 {
				text = fmt.Sprintf("%sx%d", addon.AddonName, addon.Quantity)
			}
			addonsByItem[addon.OrderItemID] = append(addonsByItem[addon.OrderItemID], text)
		}

		for _, item := range itemsByOrder[order.ID] {
			row.Items = append(row.Items, KdsItemRow{
				SpuName:  item.SpuName,
				SkuName:  item.SkuName,
				Quantity: item.Quantity,
				Remark:   item.Remark,
				Orderer:  item.Orderer,
				Addons:   addonsByItem[item.ID],
			})
		}

		rows = append(rows, row)
	}

	return rows, nil
}
