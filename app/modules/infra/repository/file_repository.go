// Package repository 存放 infra 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

const fileTable = model.InfraFileTable

// FileRepository 是 infra_file 表的仓储。
type FileRepository struct{}

func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

// Create 插入文件记录。
func (r *FileRepository) Create(ctx context.Context, m *model.InfraFileModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(fileTable).Insert(map[string]any{
		"id":         m.ID,
		"config_id":  m.ConfigID,
		"name":       m.Name,
		"path":       m.Path,
		"url":        m.Url,
		"type":       m.Type,
		"size":       m.Size,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// FindByID 按主键查询未软删的文件；未命中返回 (nil, nil)。
func (r *FileRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraFileModel, error) {
	var m model.InfraFileModel

	if err := facades.DB().WithContext(ctx).Table(fileTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}

	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FileFilter 是文件列表的筛选条件，零值表示不限制。
type FileFilter struct {
	Keyword string
}

// Paginate 按条件分页查询文件，按上传时间倒序。
func (r *FileRepository) Paginate(ctx context.Context, filter FileFilter, pageNo, pageSize int) ([]model.InfraFileModel, int64, error) {
	list := make([]model.InfraFileModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(fileTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 软删除文件记录。
func (r *FileRepository) Delete(ctx context.Context, m *model.InfraFileModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(fileTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(infraErrs.InfraFileNotExists, m.ID)
	}

	return nil
}
