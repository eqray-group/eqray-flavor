package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// SpuSkuRequest 是菜品保存时提交的一条规格；ID 为空表示新增，有值表示更新。
type SpuSkuRequest struct {
	// ID 规格 ID，新增时留空
	ID types.ID `form:"id" json:"id"`
	// Name 规格名称，如大份、中杯
	Name string `form:"name" json:"name"`
	// Code 规格编码，非空时全局唯一
	Code string `form:"code" json:"code"`
	// StandardPrice 标准价（分），必须大于 0
	StandardPrice int64 `form:"standardPrice" json:"standardPrice"`
	// MemberPrice 会员价（分），0 表示未设置
	MemberPrice int64 `form:"memberPrice" json:"memberPrice"`
	// CostPrice 成本价（分）
	CostPrice int64 `form:"costPrice" json:"costPrice"`
	// Status 上下架状态：0 下架 / 1 上架
	Status int8 `form:"status" json:"status"`
	// Sort 规格排序
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// SpuSaveRequest 是新增 / 编辑菜品的入参；规格为整表提交，未提交的存量规格会被删除。
type SpuSaveRequest struct {
	// CategoryID 所属分类 ID
	CategoryID types.ID `form:"categoryId" json:"categoryId"`
	// Name 菜品名称，同一分类下唯一
	Name string `form:"name" json:"name"`
	// Code 菜品编码，非空时全局唯一
	Code string `form:"code" json:"code"`
	// Unit 销售单位，默认「份」
	Unit string `form:"unit" json:"unit"`
	// Description 菜品介绍
	Description string `form:"description" json:"description"`
	// Image 主图地址
	Image string `form:"image" json:"image"`
	// Tags 标签，逗号分隔
	Tags string `form:"tags" json:"tags"`
	// Allergens 过敏原，逗号分隔
	Allergens string `form:"allergens" json:"allergens"`
	// Status 上下架状态：0 下架 / 1 上架
	Status int8 `form:"status" json:"status"`
	// Sort 分类内排序
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
	// Skus 规格列表，至少一条
	Skus []SpuSkuRequest `form:"skus" json:"skus"`
	// AddonGroupIDs 挂载的加料组 ID 列表，整表覆盖
	AddonGroupIDs []types.ID `form:"addonGroupIds" json:"addonGroupIds"`
}

// Rules 返回菜品入参的校验规则；规格内部字段由服务层逐条校验并给出中文错误。
func (r *SpuSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"categoryId":  "required|integer|min:1",
		"name":        "required|max:64",
		"code":        "max:64",
		"unit":        "max:16",
		"description": "max:500",
		"image":       "max:255",
		"tags":        "max:255",
		"allergens":   "max:255",
		"status":      "required|integer|in:0,1",
		"sort":        "integer|min:0",
		"remark":      "max:255",
		"skus":        "required",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *SpuSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"categoryId.required": "所属分类不能为空",
		"categoryId.integer":  "所属分类 ID 必须是整数",
		"categoryId.min":      "所属分类 ID 不能小于 1",
		"name.required":       "菜品名称不能为空",
		"name.max":            "菜品名称不能超过 64 个字符",
		"code.max":            "菜品编码不能超过 64 个字符",
		"unit.max":            "销售单位不能超过 16 个字符",
		"description.max":     "菜品介绍不能超过 500 个字符",
		"image.max":           "主图地址不能超过 255 个字符",
		"tags.max":            "标签不能超过 255 个字符",
		"allergens.max":       "过敏原不能超过 255 个字符",
		"status.required":     "上下架状态不能为空",
		"status.integer":      "上下架状态必须是整数",
		"status.in":           "上下架状态只能是 0 或 1",
		"sort.integer":        "排序必须是整数",
		"sort.min":            "排序不能小于 0",
		"remark.max":          "备注不能超过 255 个字符",
		"skus.required":       "菜品至少需要一个规格",
	}
}

// SpuStatusRequest 是菜品上下架切换入参。
type SpuStatusRequest struct {
	// Status 上下架状态：0 下架 / 1 上架
	Status int8 `form:"status" json:"status"`
}

// Rules 返回上下架切换入参的校验规则。
func (r *SpuStatusRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"status": "required|integer|in:0,1",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *SpuStatusRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"status.required": "上下架状态不能为空",
		"status.integer":  "上下架状态必须是整数",
		"status.in":       "上下架状态只能是 0 或 1",
	}
}
