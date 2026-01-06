import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')

  const isLoggedIn = () => !!token.value

  const setLogin = (t: string, u: string) => {
    token.value = t
    username.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('username', u)
  }

  const logout = () => {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  return { token, username, isLoggedIn, setLogin, logout }
})
