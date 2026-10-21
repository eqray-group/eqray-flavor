// Package controller 存放 infra 模块 C 端的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/infra/service"
)

// FileController 提供 C 端的文件上传接口，供头像、评价图等自助场景使用。
type FileController struct {
	service *service.FileService
}

func NewFileController() *FileController {
	return &FileController{service: service.NewFileService()}
}

// Upload C 端上传文件（头像、凭证图片等），只要求会员令牌。
// 入参：multipart 字段 file；返回文件 id、名称、可直接访问的 url、类型与大小，记录统一存入 infra_file。
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
