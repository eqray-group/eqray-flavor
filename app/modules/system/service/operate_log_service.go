package service

import (
	"context"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// OperateLogInput 是一次写操作留痕所需的数据，由 OperateLog 中间件组装。
type OperateLogInput struct {
	UserID        int64
	UserType      baseEnums.UserTypeEnum
	Type          string
	SubType       string
	Action        string
	Success       bool
	Extra         string
	RequestMethod string
	RequestURL    string
	RequestIP     string
	UserAgent     string
	Duration      int64
}

// OperateLogListQuery 是操作日志列表的查询条件。
type OperateLogListQuery struct {
	Type     string
	SubType  string
	Success  *int8
	Keyword  string
	PageNo   int
	PageSize int
}

// OperateLogService 负责操作日志的写入与查询（docs 6.3 / 9.5）。
type OperateLogService struct {
	logs *repository.OperateLogRepository
}

func NewOperateLogService() *OperateLogService {
	return &OperateLogService{logs: repository.NewOperateLogRepository()}
}

// Record 追加一条操作日志。日志写入失败不影响主流程，由中间件忽略返回的错误。
func (s *OperateLogService) Record(ctx context.Context, in OperateLogInput) error {
	success := int8(0)
	if in.Success {
		success = 1
	}

	return s.logs.Create(ctx, &model.SystemOperateLogModel{
		UserID:        in.UserID,
		UserType:      in.UserType,
		Type:          in.Type,
		SubType:       in.SubType,
		Action:        in.Action,
		Success:       success,
		Extra:         in.Extra,
		RequestMethod: in.RequestMethod,
		RequestURL:    in.RequestURL,
		RequestIP:     in.RequestIP,
		UserAgent:     in.UserAgent,
		Duration:      in.Duration,
	})
}

// Page 分页查询操作日志。
func (s *OperateLogService) Page(ctx context.Context, q OperateLogListQuery) ([]model.SystemOperateLogModel, int64, error) {
	return s.logs.Paginate(ctx, repository.OperateLogFilter{
		Type:    q.Type,
		SubType: q.SubType,
		Success: q.Success,
		Keyword: q.Keyword,
	}, q.PageNo, q.PageSize)
}
