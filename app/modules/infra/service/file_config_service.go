package service

import (
	"context"
	"encoding/json"
	"strings"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
)

const (
	// fileConfigLocalStorage 本地磁盘；fileConfigS3Storage 兼容 S3 协议的对象存储（MinIO / OSS / COS 等）。
	fileConfigLocalStorage = "local"
	fileConfigS3Storage    = "s3"
	// defaultS3Region 是 S3 协议要求的区域默认值，MinIO 通常可任意填。
	defaultS3Region = "us-east-1"
	// defaultS3Dir 是对象存储中的默认目录前缀。
	defaultS3Dir = "uploads"
)

// LocalStorageConfig 是 local 存储的配置项。
type LocalStorageConfig struct {
	// Dir 是相对 public 磁盘的存储目录，默认 uploads
	Dir string `json:"dir"`
}

// S3StorageConfig 是 S3 兼容存储的配置项，MinIO 只需填写 endpoint / bucket / accessKey / secretKey。
type S3StorageConfig struct {
	Bucket    string `json:"bucket"`
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	// URL 是访问前缀，留空时按 endpoint + bucket 推导
	URL string `json:"url"`
	// UsePathStyle 为 true 时使用 path-style 访问，MinIO 必须为 true
	UsePathStyle bool `json:"usePathStyle"`
	// Dir 是对象存储中的目录前缀
	Dir string `json:"dir"`
}

// FileStorageTarget 描述上传 / 删除文件时使用的存储目标。
type FileStorageTarget struct {
	ConfigID types.ID
	Storage  string
	Dir      string
	// S3 仅在 Storage 为 s3 时有值
	S3 *S3StorageConfig
}

// FileConfigInput 是新增 / 编辑文件配置的入参。
type FileConfigInput struct {
	Name    string
	Storage string
	Master  int8
	Config  string
}

// FileConfigService 负责文件配置的维护与主配置解析（docs 6.4）。
type FileConfigService struct {
	configs *repository.FileConfigRepository
}

func NewFileConfigService() *FileConfigService {
	return &FileConfigService{configs: repository.NewFileConfigRepository()}
}

// List 返回全部文件配置。
func (s *FileConfigService) List(ctx context.Context) ([]model.InfraFileConfigModel, error) {
	return s.configs.List(ctx)
}

// Create 新增文件配置；标记为主配置时会清掉其它配置的主标记。
func (s *FileConfigService) Create(ctx context.Context, in FileConfigInput) error {
	storage, config, err := normalizeStorageConfig(in.Storage, in.Config)
	if err != nil {
		return err
	}

	m := &model.InfraFileConfigModel{
		Name:    strings.TrimSpace(in.Name),
		Storage: storage,
		Master:  in.Master,
		Config:  config,
	}

	if err := s.configs.Create(ctx, m); err != nil {
		return err
	}
	if m.Master == 1 {
		return s.configs.ClearMaster(ctx, m.ID)
	}

	return nil
}

// Update 更新文件配置。
func (s *FileConfigService) Update(ctx context.Context, id types.ID, in FileConfigInput) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(infraErrs.InfraFileConfigNotExists, id)
	}

	storage, config, err := normalizeStorageConfig(in.Storage, in.Config)
	if err != nil {
		return err
	}

	m.Name = strings.TrimSpace(in.Name)
	m.Storage = storage
	m.Master = in.Master
	m.Config = config

	if err := s.configs.Update(ctx, m); err != nil {
		return err
	}
	if m.Master == 1 {
		return s.configs.ClearMaster(ctx, m.ID)
	}

	return nil
}

// SetMaster 把指定配置设为主配置，其它配置的主标记会被清除。
func (s *FileConfigService) SetMaster(ctx context.Context, id types.ID) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(infraErrs.InfraFileConfigNotExists, id)
	}
	if m.Master == 1 {
		return nil
	}

	return s.configs.SetMaster(ctx, id)
}

// Delete 删除文件配置；主配置不允许删除，避免上传失去落点。
func (s *FileConfigService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(infraErrs.InfraFileConfigNotExists, id)
	}
	if m.Master == 1 {
		return baseErrs.New(infraErrs.InfraFileConfigMasterForbidden, m.Name)
	}

	return s.configs.Delete(ctx, m)
}

// Active 返回上传时使用的存储配置：主配置存在则用它，否则回落到本地默认目录。
func (s *FileConfigService) Active(ctx context.Context) (*FileStorageTarget, error) {
	m, err := s.configs.Master(ctx)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return &FileStorageTarget{Storage: fileConfigLocalStorage, Dir: uploadDir}, nil
	}

	return buildStorageTarget(m)
}

// Target 按配置 ID 返回存储目标，供删除文件时定位磁盘；配置不存在时返回 (nil, nil)。
func (s *FileConfigService) Target(ctx context.Context, id types.ID) (*FileStorageTarget, error) {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	return buildStorageTarget(m)
}

// buildStorageTarget 把数据库中的配置解析成存储目标，并补齐缺省字段。
func buildStorageTarget(m *model.InfraFileConfigModel) (*FileStorageTarget, error) {
	storage, config, err := normalizeStorageConfig(m.Storage, m.Config)
	if err != nil {
		return nil, err
	}

	target := &FileStorageTarget{ConfigID: m.ID, Storage: storage}

	if storage == fileConfigS3Storage {
		var conf S3StorageConfig
		if err := json.Unmarshal([]byte(config), &conf); err != nil {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "S3 存储配置必须是合法的 JSON")
		}
		target.Dir = conf.Dir
		target.S3 = &conf
		return target, nil
	}

	var local LocalStorageConfig
	_ = json.Unmarshal([]byte(config), &local)
	target.Dir = local.Dir

	return target, nil
}

// normalizeStorageConfig 校验存储类型与配置结构，并返回补齐默认值后的 JSON 字符串。
func normalizeStorageConfig(storage, config string) (string, string, error) {
	storage = strings.TrimSpace(storage)
	if storage == "" {
		storage = fileConfigLocalStorage
	}

	switch storage {
	case fileConfigLocalStorage:
		local := LocalStorageConfig{Dir: uploadDir}
		if strings.TrimSpace(config) != "" {
			if err := json.Unmarshal([]byte(config), &local); err != nil {
				return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "存储配置必须是合法的 JSON")
			}
		}
		local.Dir = strings.Trim(strings.TrimSpace(local.Dir), "/")
		if local.Dir == "" {
			local.Dir = uploadDir
		}

		return storage, mustJSON(local), nil

	case fileConfigS3Storage:
		if strings.TrimSpace(config) == "" {
			return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "请填写 S3 存储配置")
		}

		var conf S3StorageConfig
		if err := json.Unmarshal([]byte(config), &conf); err != nil {
			return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "存储配置必须是合法的 JSON")
		}

		conf.Bucket = strings.TrimSpace(conf.Bucket)
		conf.Endpoint = strings.TrimSuffix(strings.TrimSpace(conf.Endpoint), "/")
		conf.AccessKey = strings.TrimSpace(conf.AccessKey)
		conf.SecretKey = strings.TrimSpace(conf.SecretKey)
		conf.Region = strings.TrimSpace(conf.Region)
		conf.URL = strings.TrimSuffix(strings.TrimSpace(conf.URL), "/")
		conf.Dir = strings.Trim(strings.TrimSpace(conf.Dir), "/")

		if conf.Bucket == "" || conf.Endpoint == "" || conf.AccessKey == "" || conf.SecretKey == "" {
			return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "S3 存储配置缺少 bucket / endpoint / accessKey / secretKey")
		}
		if conf.Region == "" {
			conf.Region = defaultS3Region
		}
		if conf.Dir == "" {
			conf.Dir = defaultS3Dir
		}
		if conf.URL == "" {
			conf.URL = conf.Endpoint + "/" + conf.Bucket
		}

		return storage, mustJSON(conf), nil

	default:
		return "", "", baseErrs.New(infraErrs.InfraStorageNotSupported, storage)
	}
}

// mustJSON 序列化已确认结构合法的配置，失败时回退为空对象。
func mustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}

	return string(data)
}
