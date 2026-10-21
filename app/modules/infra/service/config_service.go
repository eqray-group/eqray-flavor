package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
)

const (
	// configCachePrefix 是参数值缓存键前缀；值变更后按 key 精确失效（docs 6.4）。
	configCachePrefix = "infra:config:"
	configCacheTTL    = 10 * time.Minute
)

// ConfigCreateInput 是新增参数配置的入参。
type ConfigCreateInput struct {
	Category  string
	Name      string
	ConfigKey string
	Value     string
	Type      int8
	Visible   int8
	Remark    string
}

// ConfigUpdateInput 是编辑参数配置的入参；config_key 与 type 创建后不可修改。
type ConfigUpdateInput struct {
	Category string
	Name     string
	Value    string
	Visible  int8
	Remark   string
}

// ConfigListQuery 是参数配置列表的查询条件。
type ConfigListQuery struct {
	Category string
	Keyword  string
	PageNo   int
	PageSize int
}

// ConfigService 负责参数配置的读写与缓存（docs 6.4）。
// 业务代码通过 GetString / GetInt 读取配置，不直接查表。
type ConfigService struct {
	configs *repository.ConfigRepository
}

func NewConfigService() *ConfigService {
	return &ConfigService{configs: repository.NewConfigRepository()}
}

// Page 分页查询参数配置。
func (s *ConfigService) Page(ctx context.Context, q ConfigListQuery) ([]model.InfraConfigModel, int64, error) {
	return s.configs.Paginate(ctx, repository.ConfigFilter{
		Category: q.Category,
		Keyword:  q.Keyword,
	}, q.PageNo, q.PageSize)
}

// Create 新增参数配置；参数键重复由前置校验与唯一键双重兜底。
func (s *ConfigService) Create(ctx context.Context, in ConfigCreateInput) error {
	key := strings.TrimSpace(in.ConfigKey)

	exists, err := s.configs.ExistsByKey(ctx, key, 0)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(infraErrs.InfraConfigKeyExists, key)
	}

	m := &model.InfraConfigModel{
		Category:  strings.TrimSpace(in.Category),
		Name:      strings.TrimSpace(in.Name),
		ConfigKey: key,
		Value:     in.Value,
		Type:      in.Type,
		Visible:   in.Visible,
		Remark:    strings.TrimSpace(in.Remark),
	}
	if m.Type == 0 {
		m.Type = 2
	}

	return s.configs.Create(ctx, m)
}

// Update 更新参数配置并让缓存失效。
func (s *ConfigService) Update(ctx context.Context, id types.ID, in ConfigUpdateInput) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(infraErrs.InfraConfigNotExists, id)
	}

	m.Category = strings.TrimSpace(in.Category)
	m.Name = strings.TrimSpace(in.Name)
	m.Value = in.Value
	m.Visible = in.Visible
	m.Remark = strings.TrimSpace(in.Remark)

	if err := s.configs.Update(ctx, m); err != nil {
		return err
	}

	s.invalidate(m.ConfigKey)

	return nil
}

// Delete 删除参数配置；系统内置参数只允许改值，不允许删除。
func (s *ConfigService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(infraErrs.InfraConfigNotExists, id)
	}
	if m.Type == 1 {
		return baseErrs.New(infraErrs.InfraConfigBuiltinForbidden, m.ConfigKey)
	}

	if err := s.configs.Delete(ctx, m); err != nil {
		return err
	}

	s.invalidate(m.ConfigKey)

	return nil
}

// GetString 读取参数值；未配置时返回默认值。值走缓存，命中后不再查库。
func (s *ConfigService) GetString(ctx context.Context, key, defaultValue string) string {
	cacheKey := configCachePrefix + key
	if cached := facades.Cache().GetString(cacheKey); cached != "" {
		return cached
	}

	m, err := s.configs.FindByKey(ctx, key)
	if err != nil || m == nil {
		return defaultValue
	}

	_ = facades.Cache().Put(cacheKey, m.Value, configCacheTTL)

	return m.Value
}

// GetInt 读取整型参数值；未配置或无法解析时返回默认值。
func (s *ConfigService) GetInt(ctx context.Context, key string, defaultValue int) int {
	value := s.GetString(ctx, key, "")
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

// invalidate 清掉单个参数值的缓存。
func (s *ConfigService) invalidate(key string) {
	_ = facades.Cache().Forget(configCachePrefix + key)
}
