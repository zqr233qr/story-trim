<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)

const isLogin = ref(true)
const username = ref('zqr')
const password = ref('123456')
const loading = ref(false)
const errorMsg = ref('')

const handleSubmit = async () => {
  if (!username.value || !password.value) return
  loading.value = true
  errorMsg.value = ''

  try {
    let res;
    if (isLogin.value) {
      res = await api.login({ username: username.value, password: password.value })
    } else {
      await api.register({ username: username.value, password: password.value })
      res = await api.login({ username: username.value, password: password.value })
    }

    if (res && res.code === 0) {
      userStore.setLogin(res.data.token, username.value)
      uni.reLaunch({ url: '/pages/shelf/shelf' })
    } else {
      errorMsg.value = res?.msg || '操作失败'
    }
  } catch (e: any) {
    errorMsg.value = e.message || '网络错误'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="min-h-screen bg-stone-50 flex flex-col justify-center items-center p-6" :style="{ paddingTop: statusBarHeight + 'px' }">
    <view class="w-full max-w-sm">
      <!-- Logo Area -->
      <view class="text-center mb-10">
        <view class="w-16 h-16 bg-stone-900 text-white rounded-2xl mx-auto flex items-center justify-center mb-4 shadow-xl shadow-stone-200">
          <text class="i-heroicons-book-open w-8 h-8 text-white"></text>
        </view>
        <view class="text-2xl font-bold text-stone-800 tracking-tight">StoryTrim</view>
        <view class="text-stone-400 text-sm mt-2">AI 驱动的极简阅读体验</view>
      </view>

      <!-- Form -->
      <view class="space-y-4">
        <input v-model="username" type="text" placeholder="用户名" 
          class="w-full h-14 px-5 bg-white border border-stone-200 rounded-xl text-stone-800 focus:border-teal-500 transition-all placeholder-stone-300 font-medium" />
        
        <input v-model="password" password placeholder="密码" 
          class="w-full h-14 px-5 bg-white border border-stone-200 rounded-xl text-stone-800 focus:border-teal-500 transition-all placeholder-stone-300 font-medium" />

        <view v-if="errorMsg" class="text-red-500 text-xs text-center font-medium">{{ errorMsg }}</view>

        <button @click="handleSubmit" :disabled="loading" 
          class="w-full h-14 bg-stone-900 text-white rounded-xl font-bold shadow-lg flex justify-center items-center active:scale-95 transition-transform">
          <text v-if="loading" class="animate-spin mr-2">...</text>
          <text>{{ isLogin ? '进入阅读' : '注册账号' }}</text>
        </button>
      </view>

      <!-- Toggle -->
      <view class="mt-8 text-center">
        <text @click="isLogin = !isLogin" class="text-sm text-stone-400 hover:text-teal-600 font-medium">
          {{ isLogin ? '还没有账号？点击注册' : '已有账号？直接登录' }}
        </text>
      </view>
    </view>
  </view>
</template>

<style>
/* 可以在这里添加一些针对小程序的样式补丁 */
</style>
