/**
 * 菜单接口。后端按会员身份返回会员价，未登录会被拒绝，因此页面需先登录。
 */
import alovaInstance from './core/instance'

export interface MenuAddon {
  id: string
  name: string
  price: number
}

export interface MenuAddonGroup {
  id: string
  name: string
  required: number
  minSelect: number
  maxSelect: number
  items: MenuAddon[]
}

export interface MenuSku {
  id: string
  name: string
  status: number
  soldOut: number
  onlineOnly: number
  standardPrice: number
  storePrice: number
  /** 最终成交价（分），已按会员身份计算 */
  price: number
}

export interface MenuSpu {
  id: string
  name: string
  unit: string
  description: string
  tags: string
  skus: MenuSku[]
  addonGroups: MenuAddonGroup[]
}

export interface MenuCategory {
  id: string
  name: string
  children?: MenuCategory[]
  items: MenuSpu[]
}

export interface MenuSnapshot {
  storeId: string
  builtAt: string
  categories: MenuCategory[]
}

/** 拉取本店菜单；返回时已解包为 MenuSnapshot */
export function fetchMenu() {
  return alovaInstance.Get<MenuSnapshot>('/api/v1/client/dish/menus')
}
