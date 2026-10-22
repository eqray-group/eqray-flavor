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

// PrintTemplateController 提供打印模板的增删改查接口。
type PrintTemplateController struct {
	service *service.StorePrintTemplateService
}

// NewPrintTemplateController 构造打印模板控制器。
func NewPrintTemplateController() *PrintTemplateController {
	return &PrintTemplateController{service: service.NewStorePrintTemplateService()}
}

// Index 分页查询打印模板，支持名称、设备类型、打印类型与状态筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *PrintTemplateController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PrintTemplateListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if value := ctx.Request().QueryInt("printerType", 0); value > 0 {
		item := enums.PrinterTypeEnum(value)
		q.PrinterType = &item
	}
	if value := ctx.Request().QueryInt("printType", 0); value > 0 {
		item := enums.PrintTypeEnum(value)
		q.PrintType = &item
	}
	if value := ctx.Request().QueryInt("status", -1); value >= 0 {
		if value > 1 {
			return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "模板状态取值不合法"))
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

// Store 新增打印模板；同一门店下「设备类型 + 打印类型」只能有一个模板。
func (c *PrintTemplateController) Store(ctx http.Context) http.Response {
	var req request.PrintTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toPrintTemplateSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑打印模板；模板不能跨门店移动，正文变化时版本号自动递增。
func (c *PrintTemplateController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PrintTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toPrintTemplateSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除打印模板；历史打印任务应依赖自身快照，不回查本表。
func (c *PrintTemplateController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toPrintTemplateSaveInput 把请求体转换为服务层入参。
func toPrintTemplateSaveInput(req request.PrintTemplateSaveRequest) service.PrintTemplateSaveInput {
	return service.PrintTemplateSaveInput{
		StoreID:     req.StoreID,
		Name:        req.Name,
		PrinterType: enums.PrinterTypeEnum(req.PrinterType),
		PrintType:   enums.PrintTypeEnum(req.PrintType),
		PaperWidth:  req.PaperWidth,
		Content:     req.Content,
		Status:      baseEnums.CommonStatusEnum(req.Status),
		Remark:      req.Remark,
	}
}
