import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import DashboardView from '../views/DashboardView.vue'
import LoginView from '../views/LoginView.vue'
import ReaderView from '../views/ReaderView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      meta: { requiresAuth: true }
    },
    {
      path: '/reader',
      name: 'reader',
      component: ReaderView,
      meta: { requiresAuth: true }
    },
    {
      path: '/',
      redirect: '/dashboard'
    }
  ]
})

// 全局前置守卫：校验登录状态
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    // 需要登录但未登录，重定向到登录页
    next('/login')
  } else if (to.name === 'login' && userStore.isLoggedIn) {
    // 已登录状态访问登录页，重定向到首页
    next('/dashboard')
  } else {
    next()
  }
})

export default router