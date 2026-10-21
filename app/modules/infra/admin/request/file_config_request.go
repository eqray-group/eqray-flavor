package request

import (
	"github.com/goravel/framework/contracts/http"
)

// FileConfigRequest 是新增 / 编辑文件配置的入参。
type FileConfigRequest struct {
	Name    string `form:"name" json:"name"`
	Storage string `form:"storage" json:"storage"`
	Master  int8   `form:"master" json:"master"`
	Config  string `form:"config" json:"config"`
}

// Rules 返回校验规则；存储配置的 JSON 结构由 Service 校验。
func (r *FileConfigRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":    "required|max:64",
		"storage": "in:local,s3",
		"master":  "integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *FileConfigRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":  "配置名称不能为空",
		"name.max":       "配置名称不能超过 64 个字符",
		"storage.in":     "存储类型不合法",
		"master.integer": "是否主配置取值不合法",
		"master.in":      "是否主配置取值不合法",
	}
}
