/// <reference types="@uni-helper/vite-plugin-uni-pages/client" />
import { pages, subPackages } from 'virtual:uni-pages'

/**
 * 路由由页面文件自动生成：新增 src/pages 下的页面即可跳转，不需要在这里登记。
 * 登录态不在这里拦——各页面的 onShow 已经按需跳登录页，重复拦截会导致两次跳转。
 */
function generateRoutes() {
  const routes = pages.map((page) => {
    return { ...page, path: `/${page.path}` }
  })

  if (subPackages && subPackages.length > 0) {
    subPackages.forEach((subPackage) => {
      subPackage.pages.forEach((page: { path: string }) => {
        routes.push({ ...page, path: `/${subPackage.root}/${page.path}` } as (typeof routes)[number])
      })
    })
  }

  return routes
}

const router = createRouter({
  routes: generateRoutes(),
})

export default router
