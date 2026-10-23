export interface TabbarItem {
  name: string
  value?: number
  active: boolean
  title: string
  icon: string
}

// 顾客端只有两个页签：点单与订单。图标沿用 wot 内置图标，避免额外引入图标包
const tabbarItems = ref<TabbarItem[]>([
  { name: 'menu', active: true, title: '点单', icon: 'home' },
  { name: 'orders', active: false, title: '订单', icon: 'list' },
])

export function useTabbar() {
  const tabbarList = computed(() => tabbarItems.value)

  const activeTabbar = computed(() => {
    const item = tabbarItems.value.find(item => item.active)
    return item || tabbarItems.value[0]
  })

  const getTabbarItemValue = (name: string) => {
    const item = tabbarItems.value.find(item => item.name === name)
    return item?.value
  }

  const setTabbarItem = (name: string, value: number) => {
    const tabbarItem = tabbarItems.value.find(item => item.name === name)
    if (tabbarItem) {
      tabbarItem.value = value
    }
  }

  const setTabbarItemActive = (name: string) => {
    tabbarItems.value.forEach((item) => {
      if (item.name === name) {
        item.active = true
      }
      else {
        item.active = false
      }
    })
  }

  return {
    tabbarList,
    activeTabbar,
    getTabbarItemValue,
    setTabbarItem,
    setTabbarItemActive,
  }
}
