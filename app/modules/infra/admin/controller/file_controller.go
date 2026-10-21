// Package controller 存放 infra 模块的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/infra/service"
)

// FileController 提供后台的文件上传、列表与删除接口；C 端上传见 client/controller。
type FileController struct {
	service *service.FileService
}

func NewFileController() *FileController {
	return &FileController{service: service.NewFileService()}
}

// Upload 后台上传文件（头像等自助场景），只要求登录态、不校验权限点。
// 入参：multipart 字段 file；返回文件 id、名称、访问地址、类型与大小。
func (c *FileController) Upload(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "请选择要上传的文件"))
	}

	m, err := c.service.Upload(ctx.Context(), file)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"id":   m.ID,
		"name": m.Name,
		"url":  m.Url,
		"type": m.Type,
		"size": m.Size,
	})
}

// Index 分页查询文件列表，按上传时间倒序，支持按文件名关键字搜索。
// 入参：keyword、pageNo、pageSize（查询串）；返回文件列表与总数。
func (c *FileController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.FileListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Destroy 删除文件记录并清理物理文件，按记录上的配置选择存储驱动。
// 入参：路径 id；返回 null。
func (c *FileController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
