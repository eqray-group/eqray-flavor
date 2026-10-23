import { createRouter, createWebHashHistory } from 'vue-router'

import { TOKEN_KEY } from '@/utils/http'

// 门店端页面固定，直接用静态路由，不走后端菜单驱动
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: () => import('@/layouts/BasicLayout.vue'),
      redirect: '/orders',
      children: [
        {
          path: 'orders',
          name: 'Orders',
          component: () => import('@/views/order/index.vue'),
          meta: { title: '订单看板' }
        },
        {
          path: 'tables',
          name: 'Tables',
          component: () => import('@/views/table/index.vue'),
          meta: { title: '桌台' }
        },
        {
          path: 'reports',
          name: 'Reports',
          component: () => import('@/views/report/index.vue'),
          meta: { title: '报表' }
        },
        {
          path: 'pos',
          name: 'Pos',
          component: () => import('@/views/pos/index.vue'),
          meta: { title: '点单' }
        },
        {
          path: 'sold-out',
          name: 'SoldOut',
          component: () => import('@/views/dish/sold-out.vue'),
          meta: { title: '沽清' }
        },
        {
          path: 'print-tasks',
          name: 'PrintTasks',
          component: () => import('@/views/print-task/index.vue'),
          meta: { title: '打印任务' }
        }
      ]
    }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (!to.meta.public && !token) {
    return { path: '/login' }
  }
  if (to.path === '/login' && token) {
    return { path: '/' }
  }

  return true
})

export default router
