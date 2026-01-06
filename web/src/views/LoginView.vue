<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { api } from '../api'

const router = useRouter()
const userStore = useUserStore()

const isLogin = ref(true)
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const handleSubmit = async () => {
  if (!username.value || !password.value) return
  loading.value = true
  errorMsg.value = ''

  try {
    // Mock Login for UI Preview
    // if (isLogin.value) {
    //   res = await api.login({ username: username.value, password: password.value })
    // } else { ... }

    // 模拟登录成功 (即使后端未启动)
    setTimeout(() => {
      const mockToken = 'mock-token-' + Date.now()
      userStore.setLogin(mockToken, username.value)
      router.push('/shelf')
    }, 500)
    
    // Original Logic (Commented out)
    /*
    if (res && res.data.code === 0) {
      userStore.setLogin(res.data.data.token, username.value)
      router.push('/shelf')
    } else {
      errorMsg.value = res?.data.msg || '操作失败'
    }
    */
  } catch (e: any) {
    errorMsg.value = e.message || '网络错误'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#fafaf9] flex flex-col justify-center items-center p-6">
    <div class="w-full max-w-sm">
      <!-- Logo Area -->
      <div class="text-center mb-10">
        <div class="w-16 h-16 bg-stone-900 text-white rounded-2xl mx-auto flex items-center justify-center mb-4 shadow-xl shadow-stone-200">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path></svg>
        </div>
        <h1 class="text-2xl font-bold text-stone-800 tracking-tight font-serif">StoryTrim</h1>
        <p class="text-stone-400 text-sm mt-2">AI 驱动的极简阅读体验</p>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div class="space-y-4">
          <input v-model="username" type="text" placeholder="用户名" 
            class="w-full px-5 py-4 bg-white border border-stone-200 rounded-xl text-stone-800 focus:border-teal-500 focus:ring-2 focus:ring-teal-100 outline-none transition-all placeholder-stone-300 font-medium" />
          
          <input v-model="password" type="password" placeholder="密码" 
            class="w-full px-5 py-4 bg-white border border-stone-200 rounded-xl text-stone-800 focus:border-teal-500 focus:ring-2 focus:ring-teal-100 outline-none transition-all placeholder-stone-300 font-medium" />
        </div>

        <div v-if="errorMsg" class="text-red-500 text-xs text-center font-medium">{{ errorMsg }}</div>

        <button type="submit" :disabled="loading" 
          class="w-full bg-stone-900 text-white py-4 rounded-xl font-bold shadow-lg shadow-stone-200 hover:bg-teal-600 transition-all active:scale-[0.98] disabled:opacity-70 disabled:cursor-not-allowed flex justify-center items-center">
          <svg v-if="loading" class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
          {{ isLogin ? '进入阅读' : '注册账号' }}
        </button>
      </form>

      <!-- Toggle -->
      <div class="mt-8 text-center">
        <button @click="isLogin = !isLogin" class="text-sm text-stone-400 hover:text-teal-600 font-medium transition-colors">
          {{ isLogin ? '还没有账号？点击注册' : '已有账号？直接登录' }}
        </button>
      </div>
    </div>
  </div>
</template>