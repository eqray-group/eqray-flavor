import request from '@/utils/http'

// 分页查询优惠券模板：支持券名称关键字、券类型与启用状态筛选
export function fetchGetCouponTemplateList(params: Api.Promotion.CouponTemplateSearchParams) {
  return request.get<Api.Promotion.CouponTemplateList>({
    url: '/api/v1/promotion/coupon-templates',
    params
  })
}

// 查询优惠券模板详情
export function fetchGetCouponTemplateDetail(id: string) {
  return request.get<Api.Promotion.CouponTemplateItem>({
    url: `/api/v1/promotion/coupon-templates/${id}`
  })
}

// 新增优惠券模板
export function fetchCreateCouponTemplate(data: Api.Promotion.CouponTemplateSaveParams) {
  return request.post<string>({
    url: '/api/v1/promotion/coupon-templates',
    params: data
  })
}

// 修改优惠券模板
export function fetchUpdateCouponTemplate(
  id: string,
  data: Api.Promotion.CouponTemplateSaveParams
) {
  return request.put<string>({
    url: `/api/v1/promotion/coupon-templates/${id}`,
    params: data
  })
}

// 删除优惠券模板
export function fetchDeleteCouponTemplate(id: string) {
  return request.del<null>({
    url: `/api/v1/promotion/coupon-templates/${id}`
  })
}

// 向指定会员发放优惠券
export function fetchIssueCoupon(id: string, data: Api.Promotion.CouponIssueParams) {
  return request.post<number>({
    url: `/api/v1/promotion/coupon-templates/${id}/issue`,
    params: data
  })
}

// 分页查询会员券：支持按模板、会员与券状态筛选
export function fetchGetCouponList(params: Api.Promotion.CouponSearchParams) {
  return request.get<Api.Promotion.CouponList>({
    url: '/api/v1/promotion/coupons',
    params
  })
}

// 作废会员券：未使用 → 已过期
export function fetchRevokeCoupon(id: string) {
  return request.post<null>({
    url: `/api/v1/promotion/coupons/${id}/revoke`
  })
}

// 分页查询营销活动：支持活动名称关键字、玩法与启用状态筛选
export function fetchGetActivityList(params: Api.Promotion.ActivitySearchParams) {
  return request.get<Api.Promotion.ActivityList>({
    url: '/api/v1/promotion/activities',
    params
  })
}

// 查询营销活动详情，含参与商品
export function fetchGetActivityDetail(id: string) {
  return request.get<Api.Promotion.ActivityItem>({
    url: `/api/v1/promotion/activities/${id}`
  })
}

// 新增营销活动
export function fetchCreateActivity(data: Api.Promotion.ActivitySaveParams) {
  return request.post<string>({
    url: '/api/v1/promotion/activities',
    params: data
  })
}

// 修改营销活动
export function fetchUpdateActivity(id: string, data: Api.Promotion.ActivitySaveParams) {
  return request.put<string>({
    url: `/api/v1/promotion/activities/${id}`,
    params: data
  })
}

// 删除营销活动
export function fetchDeleteActivity(id: string) {
  return request.del<null>({
    url: `/api/v1/promotion/activities/${id}`
  })
}

// 查询积分抵扣规则
export function fetchGetPointRule() {
  return request.get<Api.Promotion.PointRule>({
    url: '/api/v1/promotion/point-rule'
  })
}

// 保存积分抵扣规则
export function fetchUpdatePointRule(data: Api.Promotion.PointRuleSaveParams) {
  return request.put<null>({
    url: '/api/v1/promotion/point-rule',
    params: data
  })
}
