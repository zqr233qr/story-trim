import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useBookStore } from './book'

export const useUserStore = defineStore('user', () => {
  const token = ref(uni.getStorageSync('token') || '')
  const username = ref(uni.getStorageSync('username') || '')

  const isLoggedIn = () => {
    return !!token.value
  }

  const setLogin = (t: string, u: string) => {
    console.log('[UserStore] setLogin:', u)
    token.value = t
    username.value = u
    uni.setStorageSync('token', t)
    uni.setStorageSync('username', u)

    const bookStore = useBookStore()
    bookStore.fetchBooks()
  }

  const logout = () => {
    console.log('[UserStore] logout')
    token.value = ''
    username.value = ''
    uni.removeStorageSync('token')
    uni.removeStorageSync('username')
  }

  return { token, username, isLoggedIn, setLogin, logout }
})
