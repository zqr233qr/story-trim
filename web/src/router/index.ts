import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue')
    },
    {
      path: '/shelf',
      name: 'shelf',
      component: () => import('../views/BookshelfView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/reader/:id',
      name: 'reader',
      component: () => import('../views/ReaderView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/',
      redirect: '/shelf'
    }
  ]
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.isLoggedIn()) {
    next('/login')
  } else {
    next()
  }
})

export default router
