import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import ReaderView from '../views/ReaderView.vue'
import { useUserStore } from '../stores/user'

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
      component: DashboardView
    },
    {
      path: '/reader',
      name: 'reader',
      component: ReaderView
    },
    {
      path: '/',
      redirect: '/dashboard'
    }
  ]
})

// 简单的路由守卫 (可选)
router.beforeEach((to, _, next) => {
  const userStore = useUserStore()
  // 如果去 dashboard 且没 token，这里允许通行，因为我们支持游客
  // 如果你有必须登录的页面，可以在这里拦截
  if (to.path === '/reader' && !userStore.isLoggedIn && !localStorage.getItem('token')) {
     // 允许游客进入 reader，只要 book store 有数据
  }
  next()
})

export default router
