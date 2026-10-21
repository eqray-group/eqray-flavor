// Package param 提供通用分页参数与请求参数校验辅助。
package param

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"
	contractsValidation "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/validation"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
)

const (
	defaultPageNo = 1
	// DefaultPageSize 是未传 pageSize 时的默认条数。
	DefaultPageSize = 10
	maxPageSize     = 200
)

// Page 是分页请求基类，对应 docs 8.2。
type Page struct {
	PageNo   int `form:"pageNo" json:"pageNo"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

// Normalize 校验并兜底分页参数：pageNo 小于 1 时取 1，pageSize 必须落在 1 ~ 200。
// 分页接口不接受 pageSize = -1：不分页的导出场景不套用 Page，由专用接口自行处理。
func (p *Page) Normalize() error {
	if p.PageNo < defaultPageNo {
		p.PageNo = defaultPageNo
	}
	if p.PageSize < 1 || p.PageSize > maxPageSize {
		return errs.WithMsg(errs.BadRequest, "pageSize 取值范围 1 ~ 200")
	}

	return nil
}

// FromQuery 从查询串解析分页参数：pageNo 小于 1 时取 1，pageSize 缺省取 DefaultPageSize。
// 各模块的分页接口统一用它，避免每处重复拼 Page 再 Normalize。
func FromQuery(ctx http.Context) (Page, error) {
	page := Page{
		PageNo:   ctx.Request().QueryInt("pageNo", defaultPageNo),
		PageSize: ctx.Request().QueryInt("pageSize", DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return page, err
	}

	return page, nil
}

// Validatable 是请求参数需要实现的校验规则接口。
type Validatable interface {
	Rules(ctx http.Context) map[string]any
}

// MessagesProvider 是可选接口，实现后用于覆盖框架默认的英文提示。
type MessagesProvider interface {
	Messages(ctx http.Context) map[string]string
}

// Bind 先校验再绑定请求参数，两步都失败时统一返回带 400 错误码的业务异常。
func Bind(ctx http.Context, req Validatable) error {
	if err := Validate(ctx, req); err != nil {
		return err
	}
	if err := ctx.Request().Bind(req); err != nil {
		return errs.WithMsg(errs.BadRequest, "参数解析失败")
	}
	return nil
}

// Validate 按规则校验请求数据，失败时返回带 400 错误码的业务异常。
// 返回的 msg 取自校验器的第一条错误，便于前端直接提示。
// 没有任何规则的请求（如分配角色、角色授权）直接跳过校验，框架不接受空规则集合。
func Validate(ctx http.Context, req Validatable) error {
	rules := req.Rules(ctx)
	if len(rules) == 0 {
		return nil
	}

	var options []contractsValidation.Option
	if provider, ok := req.(MessagesProvider); ok {
		options = append(options, validation.Messages(provider.Messages(ctx)))
	}

	validator, err := ctx.Request().Validate(rules, options...)
	if err != nil {
		return err
	}
	if validator.Fails() {
		return errs.WithMsg(errs.BadRequest, validator.Errors().One())
	}
	return nil
}

// PathID 解析路径参数 id，非法时返回 400 业务异常。
func PathID(ctx http.Context) (types.ID, error) {
	v, err := strconv.ParseInt(ctx.Request().Route("id"), 10, 64)
	if err != nil || v <= 0 {
		return 0, errs.WithMsg(errs.BadRequest, "路径参数 id 不合法")
	}
	return types.ID(v), nil
}
