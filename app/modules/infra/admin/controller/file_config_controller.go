package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/infra/admin/request"
	"eqray-flavor/app/modules/infra/service"
)

// FileConfigController 提供文件配置的增删改查接口。
type FileConfigController struct {
	service *service.FileConfigService
}

func NewFileConfigController() *FileConfigController {
	return &FileConfigController{service: service.NewFileConfigService()}
}

// Index 查询全部存储配置（低频数据不分页），列表会隐藏密钥字段。
// 无入参；返回 { list }。
func (c *FileConfigController) Index(ctx http.Context) http.Response {
	list, err := c.service.List(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"list": list})
}

// Store 新增存储配置，storage 支持 local / s3，config 为对应驱动的 JSON 参数。
// 入参：name、storage、master、config；返回 null。
func (c *FileConfigController) Store(ctx http.Context) http.Response {
	var req request.FileConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Create(ctx.Context(), service.FileConfigInput{
		Name:    req.Name,
		Storage: req.Storage,
		Master:  req.Master,
		Config:  req.Config,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update 编辑存储配置。
// 入参：路径 id + name、storage、master、config；返回 null。
func (c *FileConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.FileConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.FileConfigInput{
		Name:    req.Name,
		Storage: req.Storage,
		Master:  req.Master,
		Config:  req.Config,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SetMaster 把指定配置设为主配置，事务内保证主配置唯一，之后上传的文件都走该配置。
// 入参：路径 id；返回 null。
func (c *FileConfigController) SetMaster(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SetMaster(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除存储配置；主配置或被文件引用的配置由服务层拒绝删除。
// 入参：路径 id；返回 null。
func (c *FileConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
