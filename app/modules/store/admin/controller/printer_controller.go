package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// PrinterController 提供打印设备的增删改查接口。
type PrinterController struct {
	service *service.StorePrinterService
}

// NewPrinterController 构造打印设备控制器。
func NewPrinterController() *PrinterController {
	return &PrinterController{service: service.NewStorePrinterService()}
}

// Index 分页查询打印设备，支持名称 / 设备号、类型、接入方式与状态筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定，列表不返回明文密钥。
func (c *PrinterController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PrinterListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if value := ctx.Request().QueryInt("type", 0); value > 0 {
		item := enums.PrinterTypeEnum(value)
		q.Type = &item
	}
	if value := ctx.Request().QueryInt("accessMode", 0); value > 0 {
		item := enums.PrinterAccessModeEnum(value)
		q.AccessMode = &item
	}
	if value := ctx.Request().QueryInt("status", -1); value >= 0 {
		if value > 1 {
			return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "设备状态取值不合法"))
		}
		item := baseEnums.CommonStatusEnum(value)
		q.Status = &item
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增打印设备：设备号在同一门店内唯一，密钥加密保存。
// 入参：设备字段；返回新设备 ID，不返回明文密钥。
func (c *PrinterController) Store(ctx http.Context) http.Response {
	var req request.PrinterSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toPrinterSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑打印设备；设备不能跨门店移动，密钥留空表示沿用原值。
func (c *PrinterController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PrinterSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toPrinterSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除打印设备，释放该门店的设备号占用。
func (c *PrinterController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toPrinterSaveInput 把请求体转换为服务层入参。
func toPrinterSaveInput(req request.PrinterSaveRequest) service.PrinterSaveInput {
	return service.PrinterSaveInput{
		StoreID:    req.StoreID,
		Name:       req.Name,
		DeviceNo:   req.DeviceNo,
		Type:       enums.PrinterTypeEnum(req.Type),
		AccessMode: enums.PrinterAccessModeEnum(req.AccessMode),
		Config:     req.Config,
		Secret:     req.Secret,
		Copies:     req.Copies,
		Status:     baseEnums.CommonStatusEnum(req.Status),
		Remark:     req.Remark,
	}
}
