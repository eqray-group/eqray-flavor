import { http } from '@/utils/http'

// 会员钱包余额：没有钱包时后端返回余额 0，不会自动建钱包
export function fetchMemberWallet(memberId: string) {
  return http.get<{ walletId: string; userId?: string; balance: number }>(
    `/api/v1/pay/wallets/member/${memberId}`
  )
}

// 门店现金充值：金额单位为分，赠送金额可为 0
export function fetchCashRecharge(
  memberId: string,
  params: { price: number; bonusPrice: number; remark?: string }
) {
  return http.post<{ walletId: string; balance: number; price: number; bonusPrice: number }>(
    `/api/v1/pay/wallets/member/${memberId}/cash-recharge`,
    params as Record<string, unknown>
  )
}
