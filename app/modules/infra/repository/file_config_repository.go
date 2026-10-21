package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

const fileConfigTable = model.InfraFileConfigTable

// FileConfigRepository 是 infra_file_config 表的仓储。
type FileConfigRepository struct{}

func NewFileConfigRepository() *FileConfigRepository {
	return &FileConfigRepository{}
}

// Create 插入文件配置。
func (r *FileConfigRepository) Create(ctx context.Context, m *model.InfraFileConfigModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(fileConfigTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"storage":    m.Storage,
		"master":     m.Master,
		"config":     m.Config,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新文件配置。
func (r *FileConfigRepository) Update(ctx context.Context, m *model.InfraFileConfigModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(fileConfigTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"storage":    m.Storage,
			"master":     m.Master,
			"config":     m.Config,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(infraErrs.InfraFileConfigNotExists, m.Name)
	}

	return nil
}

// Delete 软删除文件配置。
func (r *FileConfigRepository) Delete(ctx context.Context, m *model.InfraFileConfigModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(fileConfigTable).
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
		return baseErrs.New(infraErrs.InfraFileConfigNotExists, m.Name)
	}

	return nil
}

// FindByID 按主键查询未软删的文件配置；未命中返回 (nil, nil)。
func (r *FileConfigRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraFileConfigModel, error) {
	var m model.InfraFileConfigModel

	if err := facades.DB().WithContext(ctx).Table(fileConfigTable).
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

// Master 返回主配置；未配置时返回 (nil, nil)，调用方使用默认本地磁盘。
func (r *FileConfigRepository) Master(ctx context.Context) (*model.InfraFileConfigModel, error) {
	var m model.InfraFileConfigModel

	if err := facades.DB().WithContext(ctx).Table(fileConfigTable).
		Where("master", 1).
		WhereNull("deleted_at").
		OrderByDesc("id").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// List 返回全部文件配置，主配置排在最前。
func (r *FileConfigRepository) List(ctx context.Context) ([]model.InfraFileConfigModel, error) {
	list := make([]model.InfraFileConfigModel, 0)

	if err := facades.DB().WithContext(ctx).Table(fileConfigTable).
		WhereNull("deleted_at").
		OrderByDesc("master").
		OrderByDesc("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ClearMaster 清除除 excludeID 以外的主配置标记。
func (r *FileConfigRepository) ClearMaster(ctx context.Context, excludeID types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(fileConfigTable).
		Where("master", 1).
		Where("id <> ?", excludeID).
		WhereNull("deleted_at").
		Update(map[string]any{"master": 0})

	return err
}

// SetMaster 在事务内把指定配置设为主配置，并清除其它配置的主标记。
func (r *FileConfigRepository) SetMaster(ctx context.Context, id types.ID) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if _, err := tx.Table(fileConfigTable).
			Where("master", 1).
			Where("id <> ?", id).
			WhereNull("deleted_at").
			Update(map[string]any{"master": 0}); err != nil {
			return err
		}

		res, err := tx.Table(fileConfigTable).
			Where("id", id).
			WhereNull("deleted_at").
			Update(map[string]any{"master": 1})
		if err != nil {
			return err
		}
		if res.RowsAffected > 0 {
			return nil
		}

		// MySQL 在新旧值相同时 RowsAffected 为 0，需要用存在性区分「已是主配置」与「记录不存在」
		total, err := tx.Table(fileConfigTable).
			Where("id", id).
			WhereNull("deleted_at").
			Count()
		if err != nil {
			return err
		}
		if total == 0 {
			return baseErrs.New(infraErrs.InfraFileConfigNotExists, id)
		}

		return nil
	})
}
