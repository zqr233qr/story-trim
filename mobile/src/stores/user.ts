import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(uni.getStorageSync('token') || '')
  const username = ref(uni.getStorageSync('username') || '')

  const isLoggedIn = () => !!token.value

  const setLogin = (t: string, u: string) => {
    token.value = t
    username.value = u
    uni.setStorageSync('token', t)
    uni.setStorageSync('username', u)
  }

  const logout = () => {
    token.value = ''
    username.value = ''
    uni.removeStorageSync('token')
    uni.removeStorageSync('username')
  }

  return { token, username, isLoggedIn, setLogin, logout }
})
