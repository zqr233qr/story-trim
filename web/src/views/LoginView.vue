<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { api } from '../api'
import { Loader2, Sparkles, ArrowRight } from 'lucide-vue-next'

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
    if (isLogin.value) {
      const res = await api.login({ username: username.value, password: password.value })
      const data = res.data
      if (data.code === 0) {
        userStore.login(data.data.token, username.value)
        router.push('/dashboard')
      } else {
        errorMsg.value = data.msg
      }
    } else {
      const res = await api.register({ username: username.value, password: password.value })
      const data = res.data
      if (data.code === 0) {
        isLogin.value = true
        errorMsg.value = '注册成功，请登录'
        password.value = ''
      } else {
        errorMsg.value = data.msg
      }
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.msg || '请求失败'
  } finally {
    loading.value = false
  }
}

const skipLogin = () => {
  router.push('/dashboard')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center relative overflow-hidden bg-gradient-to-br from-slate-50 to-gray-100">
    <!-- 装饰背景球 -->
    <div class="absolute top-20 left-20 w-72 h-72 bg-teal-200 rounded-full mix-blend-multiply filter blur-2xl opacity-30 animate-blob"></div>
    <div class="absolute bottom-20 right-20 w-72 h-72 bg-purple-200 rounded-full mix-blend-multiply filter blur-2xl opacity-30 animate-blob animation-delay-2000"></div>

    <div class="bg-white/80 backdrop-blur-xl p-8 rounded-3xl shadow-2xl w-full max-w-md border border-white/50 relative z-10">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-gradient-to-tr from-teal-400 to-emerald-500 rounded-2xl mx-auto flex items-center justify-center shadow-lg mb-4">
          <Sparkles class="text-white w-8 h-8" />
        </div>
        <h1 class="text-2xl font-bold text-gray-800">StoryTrim</h1>
        <p class="text-gray-500 text-sm mt-2">AI 驱动的沉浸式阅读伴侣</p>
      </div>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">Username</label>
          <input 
            v-model="username" 
            type="text" 
            class="w-full px-4 py-3 rounded-xl bg-gray-50 border-transparent focus:bg-white focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all outline-none" 
            placeholder="输入您的用户名"
          >
        </div>
        <div>
          <label class="block text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">Password</label>
          <input 
            v-model="password" 
            type="password" 
            class="w-full px-4 py-3 rounded-xl bg-gray-50 border-transparent focus:bg-white focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all outline-none" 
            placeholder="••••••••"
            @keyup.enter="handleSubmit"
          >
        </div>
        
        <div v-if="errorMsg" class="text-red-500 text-sm text-center bg-red-50 py-2 rounded-lg">{{ errorMsg }}</div>

        <button 
          @click="handleSubmit" 
          :disabled="loading"
          class="w-full bg-gray-900 text-white py-3.5 rounded-xl font-medium shadow-lg shadow-gray-900/20 hover:bg-gray-800 hover:shadow-xl hover:-translate-y-0.5 transition-all duration-300 flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed"
        >
          <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
          <span v-else>{{ isLogin ? '登录' : '注册' }}</span>
          <ArrowRight v-if="!loading" class="w-4 h-4" />
        </button>

        <div class="flex justify-between items-center text-xs mt-6 px-1">
          <button @click="isLogin = !isLogin" class="text-teal-600 hover:text-teal-700 font-medium">
            {{ isLogin ? '创建新账号' : '返回登录' }}
          </button>
          <button @click="skipLogin" class="text-gray-400 hover:text-gray-600">
            游客试用 &rarr;
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.animation-delay-2000 {
  animation-delay: 2s;
}
</style>