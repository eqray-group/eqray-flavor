// promotion_client_service.go 实现 promotion 模块的 C 端能力：会员查看与领取优惠券。
// 会员身份一律从令牌取，请求体不接受会员 ID；发券与查看都按会员维度过滤。
package service

import (
	"context"

	"eqray-flavor/app/modules/base/errs"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
	promotionErrs "eqray-flavor/app/modules/promotion/errs"
	promotionModel "eqray-flavor/app/modules/promotion/model"
	promotionRepository "eqray-flavor/app/modules/promotion/repository"
)

// MyCouponRow 是会员持有券的输出行。
type MyCouponRow struct {
	// ID 券 ID
	ID types.ID `json:"id"`
	// TemplateID 券模板 ID
	TemplateID types.ID `json:"templateId"`
	// Name 券名称
	Name string `json:"name"`
	// Type 券类型
	Type int8 `json:"type"`
	// TypeText 券类型描述
	TypeText string `json:"typeText"`
	// Description 券说明
	Description string `json:"description"`
	// ThresholdAmount 使用门槛（分）
	ThresholdAmount int64 `json:"thresholdAmount"`
	// DiscountAmount 抵扣金额（分）
	DiscountAmount int64 `json:"discountAmount"`
	// DiscountRate 折扣率（万分比）
	DiscountRate int32 `json:"discountRate"`
	// Status 券状态
	Status int8 `json:"status"`
	// StatusText 券状态描述
	StatusText string `json:"statusText"`
	// ExpireTime 过期时间
	ExpireTime types.DateTime `json:"expireTime"`
}

// MyCoupons 查询会员持有的优惠券。
// 入参：会员 ID、状态筛选（0 表示不限）；返回券列表。
// 说明：只返回会员自己的券，过期券一并列出由前端提示。
func (s *PromotionService) MyCoupons(ctx context.Context, memberID types.ID, status enums.CouponStatus) ([]MyCouponRow, error) {
	if memberID == 0 {
		return nil, errs.New(promotionErrs.CouponNotOwned)
	}

	list, _, err := s.coupons.Page(ctx, promotionRepository.CouponListFilter{
		MemberID: memberID,
		Status:   couponStatusPointer(status),
	}, 1, 100)
	if err != nil {
		return nil, err
	}

	templateIDs := make([]types.ID, 0, len(list))
	for _, item := range list {
		templateIDs = append(templateIDs, item.TemplateID)
	}
	templates, err := s.templates.FindByIDs(ctx, templateIDs)
	if err != nil {
		return nil, err
	}
	templateMap := make(map[types.ID]promotionModel.CouponTemplateModel, len(templates))
	for _, item := range templates {
		templateMap[item.ID] = item
	}

	rows := make([]MyCouponRow, 0, len(list))
	for _, item := range list {
		row := MyCouponRow{
			ID: item.ID, TemplateID: item.TemplateID,
			Status: int8(item.Status), StatusText: item.Status.String(),
			ExpireTime: item.ExpireTime,
		}
		if template, ok := templateMap[item.TemplateID]; ok {
			row.Name = template.Name
			row.Type = int8(template.Type)
			row.TypeText = template.Type.String()
			row.Description = template.Description
			row.ThresholdAmount = template.ThresholdAmount
			row.DiscountAmount = template.DiscountAmount
			row.DiscountRate = template.DiscountRate
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// AvailableTemplateRow 是可领取券模板的输出行。
type AvailableTemplateRow struct {
	// ID 模板 ID
	ID types.ID `json:"id"`
	// Name 券名称
	Name string `json:"name"`
	// Type 券类型
	Type int8 `json:"type"`
	// TypeText 券类型描述
	TypeText string `json:"typeText"`
	// Description 券说明
	Description string `json:"description"`
	// ThresholdAmount 使用门槛（分）
	ThresholdAmount int64 `json:"thresholdAmount"`
	// DiscountAmount 抵扣金额（分）
	DiscountAmount int64 `json:"discountAmount"`
	// DiscountRate 折扣率（万分比）
	DiscountRate int32 `json:"discountRate"`
	// MaxDiscountAmount 折扣封顶（分）
	MaxDiscountAmount int64 `json:"maxDiscountAmount"`
	// Remaining 剩余可领张数，-1 表示不限量
	Remaining int `json:"remaining"`
	// Taken 已领取张数
	Taken int `json:"taken"`
	// EndTime 结束时间
	EndTime types.DateTime `json:"endTime"`
}

// AvailableTemplates 查询当前可主动领取的券模板。
// 入参：会员 ID；返回可领模板列表，含本人已领张数与剩余量。
// 说明：只展示发放方式为主动领取、已启用、在有效期内的模板。
func (s *PromotionService) AvailableTemplates(ctx context.Context, memberID types.ID) ([]AvailableTemplateRow, error) {
	now := types.Now()
	templates, _, err := s.templates.Page(ctx, promotionRepository.TemplateListFilter{
		Status: enabledStatus(),
	}, 1, 200)
	if err != nil {
		return nil, err
	}

	rows := make([]AvailableTemplateRow, 0, len(templates))
	for i := range templates {
		template := &templates[i]
		if template.IssueType != enums.CouponIssueActive {
			continue
		}
		if now.Time().Before(template.StartTime.Time()) || now.Time().After(template.EndTime.Time()) {
			continue
		}
		// 限量模板发完即不再展示
		if template.TotalCount > 0 && template.IssuedCount >= template.TotalCount {
			continue
		}

		row := AvailableTemplateRow{
			ID: template.ID, Name: template.Name,
			Type: int8(template.Type), TypeText: template.Type.String(),
			Description:      template.Description,
			ThresholdAmount:  template.ThresholdAmount,
			DiscountAmount:   template.DiscountAmount,
			DiscountRate:     template.DiscountRate,
			MaxDiscountAmount: template.MaxDiscountAmount,
			EndTime:          template.EndTime,
		}
		if template.TotalCount > 0 {
			row.Remaining = template.TotalCount - template.IssuedCount
		} else {
			row.Remaining = -1
		}

		// 本人已领张数：用于前端展示「已领取 N 张」
		if memberID > 0 {
			taken, err := s.coupons.CountByTemplateAndMember(ctx, template.ID, memberID)
			if err != nil {
				return nil, err
			}
			row.Taken = taken
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// Take 会员主动领取优惠券。
// 入参：模板 ID 与会员 ID；返回领取的张数。
// 并发保护：发行量与每人限领都在事务内校验，超发直接拒绝。
func (s *PromotionService) Take(ctx context.Context, templateID, memberID types.ID) (int, error) {
	if memberID == 0 {
		return 0, errs.New(promotionErrs.CouponNotOwned)
	}

	template, err := s.templates.FindByID(ctx, templateID)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, errs.New(promotionErrs.CouponTemplateNotExists)
	}
	if template.Status != baseEnums.CommonStatusEnabled {
		return 0, errs.New(promotionErrs.CouponTemplateDisabled)
	}
	// 只允许主动领取的模板被自己领取
	if template.IssueType != enums.CouponIssueActive {
		return 0, errs.New(promotionErrs.CouponTemplateIssueTypeInvalid)
	}

	return s.Issue(ctx, IssueInput{TemplateID: templateID, MemberID: memberID, Count: 1})
}

// couponStatusPointer 把状态筛选转成可选指针；0 表示不限。
func couponStatusPointer(status enums.CouponStatus) *enums.CouponStatus {
	if status == 0 {
		return nil
	}

	return &status
}

// enabledStatus 返回启用状态指针，供仓储筛选复用。
func enabledStatus() *int8 {
	status := int8(baseEnums.CommonStatusEnabled)

	return &status
}
