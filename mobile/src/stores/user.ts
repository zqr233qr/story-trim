import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useBookStore } from './book'

export const useUserStore = defineStore('user', () => {
  const parseUserId = (t: string): number => {
    try {
      if (!t) return 0
      if (typeof atob === 'undefined') return 0
      const payload = t.split('.')[1]
      if (!payload) return 0
      const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
      const decoded = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => `%${('00' + c.charCodeAt(0).toString(16)).slice(-2)}`)
          .join(''),
      )
      const data = JSON.parse(decoded)
      return Number(data.userID || 0)
    } catch (e) {
      return 0
    }
  }

  const storedToken = uni.getStorageSync('token') || ''
  const token = ref(storedToken)
  const username = ref(uni.getStorageSync('username') || '')
  const userId = ref(
    Number(uni.getStorageSync('user_id') || parseUserId(storedToken) || 0),
  )

  const isLoggedIn = () => {
    return !!token.value
  }

  const setLogin = (t: string, u: string) => {
    console.log('[UserStore] setLogin:', u)
    token.value = t
    username.value = u
    userId.value = parseUserId(t)
    uni.setStorageSync('token', t)
    uni.setStorageSync('username', u)
    uni.setStorageSync('user_id', userId.value)

    const bookStore = useBookStore()
    bookStore.fetchBooks()
  }

  const logout = () => {
    console.log('[UserStore] logout')
    token.value = ''
    username.value = ''
    userId.value = 0
    uni.removeStorageSync('token')
    uni.removeStorageSync('username')
    uni.removeStorageSync('user_id')
  }

  return { token, username, userId, isLoggedIn, setLogin, logout }
})
