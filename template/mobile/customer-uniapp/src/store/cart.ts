/**
 * 购物车：本地持久化，避免小程序被切后台后丢单。
 * 只存「点了什么」，不存优惠结果——金额一律由服务端试算。
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

const CART_KEY = 'customer_cart'

export interface CartAddon {
  addonItemId: string
  name: string
  price: number
  quantity: number
}

export interface CartLine {
  /** 同菜品 + 同加料 + 同备注视为一行 */
  key: string
  skuId: string
  spuName: string
  skuName: string
  unitPrice: number
  quantity: number
  addons: CartAddon[]
  addonText: string
  remark: string
}

function readCart(): CartLine[] {
  try {
    const raw = uni.getStorageSync(CART_KEY) as string
    return raw ? (JSON.parse(raw) as CartLine[]) : []
  } catch {
    return []
  }
}

export const useCartStore = defineStore('cart', () => {
  const lines = ref<CartLine[]>(readCart())

  const persist = (): void => {
    uni.setStorageSync(CART_KEY, JSON.stringify(lines.value))
  }

  /** 件数：按总份数算，加 3 份就是 3 */
  const count = computed(() => lines.value.reduce((sum, line) => sum + line.quantity, 0))

  /** 展示用小计（分）：仅用于购物车条，不含活动与券 */
  const subtotal = computed(() =>
    lines.value.reduce(
      (sum, line) =>
        sum + line.unitPrice * line.quantity + line.addons.reduce((a, addon) => a + addon.price * addon.quantity, 0) * line.quantity,
      0,
    ),
  )

  const add = (line: Omit<CartLine, 'quantity'>): void => {
    const exist = lines.value.find((item) => item.key === line.key)
    if (exist) {
      exist.quantity += 1
    } else {
      lines.value.push({ ...line, quantity: 1 })
    }
    persist()
  }

  const inc = (key: string): void => {
    const line = lines.value.find((item) => item.key === key)
    if (!line) return
    line.quantity += 1
    persist()
  }

  const dec = (key: string): void => {
    const index = lines.value.findIndex((item) => item.key === key)
    if (index < 0) return
    const line = lines.value[index]
    if (line.quantity > 1) {
      line.quantity -= 1
    } else {
      lines.value.splice(index, 1)
    }
    persist()
  }

  const remove = (key: string): void => {
    lines.value = lines.value.filter((item) => item.key !== key)
    persist()
  }

  const clear = (): void => {
    lines.value = []
    persist()
  }

  return { lines, count, subtotal, add, inc, dec, remove, clear }
})
