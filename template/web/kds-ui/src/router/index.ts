import { createRouter, createWebHashHistory } from 'vue-router'

import { TOKEN_KEY } from '@/utils/http'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue')
    },
    {
      path: '/',
      name: 'Board',
      component: () => import('@/views/board/index.vue')
    },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ]
})

// 未登录一律回登录页：后厨屏幕长期挂着，不能停在空白页
router.beforeEach((to) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (!token && to.name !== 'Login') {
    return { name: 'Login' }
  }
  if (token && to.name === 'Login') {
    return { name: 'Board' }
  }

  return true
})

export default router
