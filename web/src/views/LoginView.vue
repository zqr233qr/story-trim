<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useUserStore } from '../stores/user'
import { Loader2 } from 'lucide-vue-next'

const router = useRouter()
const userStore = useUserStore()

const isLogin = ref(true)
const loading = ref(false)
const form = ref({ username: '', password: '' })
const errorMsg = ref('')

const handleSubmit = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    if (isLogin.value) {
      const res = await api.login(form.value)
      if (res.data.code === 0) {
        userStore.setToken(res.data.data.token)
        router.push('/dashboard')
      } else {
        errorMsg.value = res.data.msg
      }
    } else {
      const res = await api.register(form.value)
      if (res.data.code === 0) {
        isLogin.value = true
        errorMsg.value = '注册成功，请登录'
      } else {
        errorMsg.value = res.data.msg
      }
    }
  } catch (e) {
    errorMsg.value = '网络错误，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[#F9F7F1]">
    <div class="bg-white p-8 rounded-2xl shadow-xl w-full max-w-md">
      <h1 class="text-2xl font-serif font-bold text-center text-slate-800 mb-8">StoryTrim</h1>
      
      <form @submit.prevent="handleSubmit" class="space-y-6">
        <div>
          <label class="block text-sm font-medium text-slate-600 mb-1">用户名</label>
          <input v-model="form.username" type="text" required class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-teal-500 focus:border-transparent outline-none transition-all" />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-slate-600 mb-1">密码</label>
          <input v-model="form.password" type="password" required class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-teal-500 focus:border-transparent outline-none transition-all" />
        </div>

        <div v-if="errorMsg" class="text-red-500 text-xs text-center">{{ errorMsg }}</div>

        <button type="submit" :disabled="loading" class="w-full bg-slate-900 text-white py-3 rounded-lg font-bold hover:bg-slate-800 transition-colors flex justify-center items-center">
          <Loader2 v-if="loading" class="w-5 h-5 animate-spin mr-2" />
          {{ isLogin ? '登 录' : '注 册' }}
        </button>
      </form>

      <div class="mt-6 text-center text-sm text-slate-500">
        {{ isLogin ? '还没有账号？' : '已有账号？' }}
        <button @click="isLogin = !isLogin" class="text-teal-600 font-bold hover:underline ml-1">
          {{ isLogin ? '立即注册' : '去登录' }}
        </button>
      </div>
    </div>
  </div>
</template>
