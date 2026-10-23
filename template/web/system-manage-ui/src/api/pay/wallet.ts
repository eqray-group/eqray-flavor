import request from '@/utils/http'

// 分页查询会员钱包：按用户 ID 与用户类型筛选
export function fetchGetPayWalletList(params: Api.Pay.WalletSearchParams) {
  return request.get<Api.Pay.WalletList>({
    url: '/api/v1/pay/wallets',
    params
  })
}

// 查询钱包详情
export function fetchGetPayWalletDetail(id: string) {
  return request.get<Api.Pay.WalletItem>({
    url: `/api/v1/pay/wallets/${id}`
  })
}

// 分页查询钱包流水：支持按业务类型与收支方向筛选
export function fetchGetPayWalletTransactions(
  id: string,
  params: Api.Pay.WalletTransactionSearchParams
) {
  return request.get<Api.Pay.WalletTransactionList>({
    url: `/api/v1/pay/wallets/${id}/transactions`,
    params
  })
}

// 调整钱包余额：price 正数加款、负数扣款，后端写一条流水
export function fetchChangePayWalletBalance(id: string, data: Api.Pay.WalletChangeParams) {
  return request.post<null>({
    url: `/api/v1/pay/wallets/${id}/balance`,
    params: data
  })
}

// 分页查询充值套餐（含已关闭）
export function fetchGetPayWalletPackageList(params: { pageNo?: number; pageSize?: number }) {
  return request.get<Api.Pay.WalletPackageList>({
    url: '/api/v1/pay/wallet-packages',
    params
  })
}

// 查询开启状态的充值套餐，供档位下拉与充值页面使用
export function fetchGetPayWalletPackageOptions() {
  return request.get<Api.Pay.WalletPackageItem[]>({
    url: '/api/v1/pay/wallet-packages/options'
  })
}

// 新增充值套餐
export function fetchCreatePayWalletPackage(data: Api.Pay.WalletPackageSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/pay/wallet-packages',
    params: data
  })
}

// 编辑充值套餐：已产生的充值记录不受影响
export function fetchUpdatePayWalletPackage(id: string, data: Api.Pay.WalletPackageSaveParams) {
  return request.put<null>({
    url: `/api/v1/pay/wallet-packages/${id}`,
    params: data
  })
}

// 删除充值套餐（软删除），历史充值记录保留
export function fetchDeletePayWalletPackage(id: string) {
  return request.del<null>({
    url: `/api/v1/pay/wallet-packages/${id}`
  })
}

// 分页查询充值记录：按钱包与充值状态筛选
export function fetchGetPayWalletRechargeList(params: Api.Pay.WalletRechargeSearchParams) {
  return request.get<Api.Pay.WalletRechargeList>({
    url: '/api/v1/pay/wallet-recharges',
    params
  })
}

// 查询充值记录详情
export function fetchGetPayWalletRechargeDetail(id: string) {
  return request.get<Api.Pay.WalletRechargeItem>({
    url: `/api/v1/pay/wallet-recharges/${id}`
  })
}
