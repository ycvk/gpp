import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: {
      title: '代理控制',
      keepAlive: true
    }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: {
      title: '设置',
      keepAlive: false
    }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior() {
    // 始终返回顶部位置（桌面应用不需要记住滚动位置）
    return { top: 0 }
  }
})

// 全局前置守卫
router.beforeEach(async (to, from) => {
  // 1. 页面title更新
  if (to.meta?.title) {
    document.title = `GPP加速器 - ${to.meta.title}`
  }
  
  // 2. 路由验证（如果需要）
  if (to.name === 'Settings') {
    // 可以在这里添加设置页面的访问权限检查
    // 例如检查是否有未保存的更改等
  }
  
  return true
})

// 全局解析守卫
router.beforeResolve(async (to) => {
  // 确保页面数据预加载（如果需要）
  if (to.name === 'Settings') {
    // 预加载设置数据
    try {
      // await loadSettings() // 未来实现
    } catch (error) {
      console.warn('预加载设置数据失败:', error)
    }
  }
})

// 全局后置钩子
router.afterEach((to, from) => {
  // 路由切换统计（可选）
  console.log(`路由切换: ${String(from.name || 'unknown')} -> ${String(to.name || 'unknown')}`)
})

// 错误处理
router.onError((error) => {
  console.error('路由错误:', error)
  // 可以在这里添加错误上报逻辑
})

export { routes }
export default router