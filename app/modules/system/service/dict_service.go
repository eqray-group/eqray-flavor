package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

const (
	dictCachePrefix = "system:dict:"
	dictCacheTTL    = time.Hour
)

var dictTypeCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// 字典标签可用的颜色，与前端 ElTag 类型对齐。
var dictColorTypes = map[string]bool{
	"default": true,
	"primary": true,
	"success": true,
	"info":    true,
	"warning": true,
	"danger":  true,
}

// DictTypeListQuery 是字典类型列表的查询条件。
type DictTypeListQuery struct {
	Name     string
	Type     string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// DictTypeSaveInput 是新增 / 编辑字典类型的入参。
type DictTypeSaveInput struct {
	Name   string
	Type   string
	Status baseEnums.CommonStatusEnum
	Remark string
}

// DictDataListQuery 是字典数据列表的查询条件。
type DictDataListQuery struct {
	DictType string
	Label    string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// DictDataSaveInput 是新增 / 编辑字典数据的入参。
type DictDataSaveInput struct {
	Sort      int
	Label     string
	Value     string
	DictType  string
	Status    baseEnums.CommonStatusEnum
	ColorType string
	CSSClass  string
	Remark    string
}

// DictItem 是前端下拉 / 标签渲染需要的字典项。
type DictItem struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	ColorType string `json:"colorType"`
	CSSClass  string `json:"cssClass"`
	Sort      int    `json:"sort"`
}

// DictService 负责字典类型与字典数据的业务规则（docs 6.3）。
type DictService struct {
	types *repository.DictTypeRepository
	data  *repository.DictDataRepository
}

func NewDictService() *DictService {
	return &DictService{
		types: repository.NewDictTypeRepository(),
		data:  repository.NewDictDataRepository(),
	}
}

// PageTypes 分页查询字典类型。
func (s *DictService) PageTypes(ctx context.Context, q DictTypeListQuery) ([]model.SystemDictTypeModel, int64, error) {
	return s.types.Paginate(ctx, repository.DictTypeFilter{
		Name:   strings.TrimSpace(q.Name),
		Type:   strings.TrimSpace(q.Type),
		Status: q.Status,
	}, q.PageNo, q.PageSize)
}

// TypeOptions 返回全部启用字典类型，供字典数据弹窗选择。
func (s *DictService) TypeOptions(ctx context.Context) ([]model.SystemDictTypeModel, error) {
	return s.types.ListEnabled(ctx)
}

// TypeDetail 查询字典类型详情。
func (s *DictService) TypeDetail(ctx context.Context, id types.ID) (*model.SystemDictTypeModel, error) {
	m, err := s.types.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(systemErrs.SystemDictTypeNotExists, id)
	}

	return m, nil
}

// CreateType 新增字典类型。
func (s *DictService) CreateType(ctx context.Context, in DictTypeSaveInput) (types.ID, error) {
	name, code, remark, err := normalizeDictType(in)
	if err != nil {
		return 0, err
	}

	exists, err := s.types.ExistsByType(ctx, code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(systemErrs.SystemDictTypeExists, code)
	}

	m := &model.SystemDictTypeModel{
		Name:   name,
		Type:   code,
		Status: in.Status,
		Remark: remark,
	}
	if err := s.types.Create(ctx, m); err != nil {
		return 0, err
	}

	s.invalidate(code)

	return m.ID, nil
}

// UpdateType 编辑字典类型；类型编码创建后不可修改，避免字典数据失去归属。
func (s *DictService) UpdateType(ctx context.Context, id types.ID, in DictTypeSaveInput) error {
	existing, err := s.TypeDetail(ctx, id)
	if err != nil {
		return err
	}

	name, code, remark, err := normalizeDictType(in)
	if err != nil {
		return err
	}
	if code != existing.Type {
		return baseErrs.New(systemErrs.SystemDictTypeImmutable)
	}

	existing.Name = name
	existing.Status = in.Status
	existing.Remark = remark

	if err := s.types.Update(ctx, existing); err != nil {
		return err
	}

	s.invalidate(existing.Type)

	return nil
}

// DeleteType 物理删除字典类型及其数据；字典属于配置数据，不保留软删除痕迹。
func (s *DictService) DeleteType(ctx context.Context, id types.ID) error {
	existing, err := s.TypeDetail(ctx, id)
	if err != nil {
		return err
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.data.DeleteByType(ctx, tx, existing.Type); err != nil {
			return err
		}

		return s.types.Delete(ctx, tx, existing.ID)
	})
	if err != nil {
		return err
	}

	s.invalidate(existing.Type)

	return nil
}

// PageData 分页查询某个字典类型下的数据。
func (s *DictService) PageData(ctx context.Context, q DictDataListQuery) ([]model.SystemDictDataModel, int64, error) {
	return s.data.Paginate(ctx, repository.DictDataFilter{
		DictType: strings.TrimSpace(q.DictType),
		Label:    strings.TrimSpace(q.Label),
		Status:   q.Status,
	}, q.PageNo, q.PageSize)
}

// DataDetail 查询字典数据详情。
func (s *DictService) DataDetail(ctx context.Context, id types.ID) (*model.SystemDictDataModel, error) {
	m, err := s.data.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(systemErrs.SystemDictDataNotExists, id)
	}

	return m, nil
}

// CreateData 新增字典数据。
func (s *DictService) CreateData(ctx context.Context, in DictDataSaveInput) (types.ID, error) {
	m, err := s.prepareData(ctx, in, 0)
	if err != nil {
		return 0, err
	}

	if err := s.data.Create(ctx, m); err != nil {
		return 0, err
	}

	s.invalidate(m.DictType)

	return m.ID, nil
}

// UpdateData 编辑字典数据。
func (s *DictService) UpdateData(ctx context.Context, id types.ID, in DictDataSaveInput) error {
	existing, err := s.DataDetail(ctx, id)
	if err != nil {
		return err
	}

	m, err := s.prepareData(ctx, in, existing.ID)
	if err != nil {
		return err
	}
	m.ID = existing.ID

	if err := s.data.Update(ctx, m); err != nil {
		return err
	}

	// 允许把数据挪到另一个字典类型，两个类型的缓存都要失效
	s.invalidate(existing.DictType)
	if m.DictType != existing.DictType {
		s.invalidate(m.DictType)
	}

	return nil
}

// DeleteData 物理删除字典数据。
func (s *DictService) DeleteData(ctx context.Context, id types.ID) error {
	existing, err := s.DataDetail(ctx, id)
	if err != nil {
		return err
	}

	if err := s.data.Delete(ctx, existing.ID); err != nil {
		return err
	}

	s.invalidate(existing.DictType)

	return nil
}

// Items 返回某个字典类型下的启用项，命中缓存时直接返回（前端下拉 / 标签渲染用）。
func (s *DictService) Items(ctx context.Context, dictType string) ([]DictItem, error) {
	dictType = strings.TrimSpace(dictType)
	if dictType == "" {
		return []DictItem{}, nil
	}

	key := dictCachePrefix + dictType
	if cached := facades.Cache().GetString(key); cached != "" {
		items := make([]DictItem, 0)
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	list, err := s.data.ListEnabledByType(ctx, dictType)
	if err != nil {
		return nil, err
	}

	items := make([]DictItem, 0, len(list))
	for _, m := range list {
		items = append(items, DictItem{
			Label:     m.Label,
			Value:     m.Value,
			ColorType: m.ColorType,
			CSSClass:  m.CSSClass,
			Sort:      m.Sort,
		})
	}

	if payload, err := json.Marshal(items); err == nil {
		_ = facades.Cache().Put(key, string(payload), dictCacheTTL)
	}

	return items, nil
}

// prepareData 校验字典数据入参并组装 Model。
func (s *DictService) prepareData(ctx context.Context, in DictDataSaveInput, excludeID types.ID) (*model.SystemDictDataModel, error) {
	dictType := strings.TrimSpace(in.DictType)
	label := strings.TrimSpace(in.Label)
	value := strings.TrimSpace(in.Value)
	if dictType == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "字典类型不能为空")
	}
	if label == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "显示文本不能为空")
	}
	if value == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "字典键值不能为空")
	}
	if !in.Status.IsValid() {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	dictTypeModel, err := s.types.FindByType(ctx, dictType)
	if err != nil {
		return nil, err
	}
	if dictTypeModel == nil {
		return nil, baseErrs.New(systemErrs.SystemDictTypeNotExists, dictType)
	}

	colorType := strings.TrimSpace(in.ColorType)
	if colorType == "" {
		colorType = "default"
	}
	if !dictColorTypes[colorType] {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "标签样式取值不合法")
	}

	exists, err := s.data.ExistsByValue(ctx, dictType, value, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, baseErrs.New(systemErrs.SystemDictValueExists, value)
	}

	return &model.SystemDictDataModel{
		Sort:      in.Sort,
		Label:     label,
		Value:     value,
		DictType:  dictType,
		Status:    in.Status,
		ColorType: colorType,
		CSSClass:  strings.TrimSpace(in.CSSClass),
		Remark:    strings.TrimSpace(in.Remark),
	}, nil
}

// normalizeDictType 归一化并校验字典类型入参。
func normalizeDictType(in DictTypeSaveInput) (string, string, string, error) {
	name := strings.TrimSpace(in.Name)
	code := strings.TrimSpace(in.Type)
	if name == "" {
		return "", "", "", baseErrs.WithMsg(baseErrs.BadRequest, "字典名称不能为空")
	}
	if !dictTypeCodePattern.MatchString(code) {
		return "", "", "", baseErrs.WithMsg(baseErrs.BadRequest, "字典类型只能包含小写字母、数字与下划线，且以字母开头")
	}
	if !in.Status.IsValid() {
		return "", "", "", baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	return name, code, strings.TrimSpace(in.Remark), nil
}

// invalidate 清除某个字典类型的缓存。
func (s *DictService) invalidate(dictType string) {
	if dictType == "" {
		return
	}
	_ = facades.Cache().Forget(dictCachePrefix + dictType)
}
