// Package model 提供所有业务 Model 的基类。
package model

import (
	"context"
	"strconv"

	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
)

// BaseModel 对应 docs 6.1 的 5 个审计字段，业务 Model 匿名嵌入即可。
// 字段用 db tag 映射：查询构造器（facades.DB()）按 db tag 读写列，
// 见 docs/framework/queries.md；gorm tag 在这里不起作用。
type BaseModel struct {
	ID        types.ID        `db:"id" json:"id"`
	Creator   string          `db:"creator" json:"creator"`
	CreatedAt types.DateTime  `db:"created_at" json:"createdAt"`
	Updater   string          `db:"updater" json:"updater"`
	UpdatedAt types.DateTime  `db:"updated_at" json:"updatedAt"`
	DeletedAt *types.DateTime `db:"deleted_at" json:"-"`
}

// FillCreate 在 Insert 之前填充主键与审计字段。
//
// 查询构造器不触发 GORM hook，因此由 Repository 在写入前显式调用一次；
// 业务层（Service / Controller）仍然不碰这些字段，见 docs 6.9。
func (b *BaseModel) FillCreate(ctx context.Context) {
	if b.ID == 0 {
		b.ID = support.NewID()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = types.Now()
	}
	if uid := support.UserID(ctx); uid != 0 {
		b.Creator = strconv.FormatInt(uid, 10)
	}
	b.Updater = b.Creator
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = b.CreatedAt
	}
}

// FillUpdate 在 Update / 软删除之前填充更新者与更新时间。
func (b *BaseModel) FillUpdate(ctx context.Context) {
	b.UpdatedAt = types.Now()
	if uid := support.UserID(ctx); uid != 0 {
		b.Updater = strconv.FormatInt(uid, 10)
	}
}
