<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useBookStore, type Book } from '@/stores/book'
import { api } from '@/api'

const userStore = useUserStore()
const bookStore = useBookStore()
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)

onMounted(() => {
  bookStore.fetchBooks()
})

const handleBookClick = (book: Book) => {
  bookStore.setActiveBook(book.id)
  uni.navigateTo({ url: `/pages/reader/reader?id=${book.id}` })
}

// 获取后端配置，供 renderjs 使用
const uploadUrl = (import.meta.env.PROD ? '/api/v1' : 'http://192.168.3.178:8080/api/v1') + '/upload'
const userToken = uni.getStorageSync('token')

// 供 renderjs 调用的通知方法
const onUploadSuccess = () => {
  uni.hideLoading()
  uni.showToast({ title: '导入成功', icon: 'success' })
  bookStore.fetchBooks()
}

const onUploadError = (msg: string) => {
  uni.hideLoading()
  uni.showModal({ title: '上传失败', content: msg || '网络错误', showCancel: false })
}

const onUploadStart = () => {
  uni.showLoading({ title: '上传中...' })
}

const triggerUpload = () => {
  // #ifdef MP-WEIXIN
  uni.chooseMessageFile({
    count: 1,
    type: 'file',
    extension: ['.txt'],
    success: (res) => handleUpload(res.tempFiles[0].path, res.tempFiles[0].name)
  })
  // #endif

  // #ifndef MP-WEIXIN
  renderTrigger.value = Date.now()
  // #endif
}

const renderTrigger = ref(0)

const handleUpload = async (path: string, name: string) => {
  uni.showLoading({ title: '处理中...' })
  try {
    await api.upload(path, name)
    await bookStore.fetchBooks()
    uni.showToast({ title: '导入成功', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: '导入失败', icon: 'none' })
  } finally {
    uni.hideLoading()
  }
}

const handleLogout = () => {
  userStore.logout()
  uni.reLaunch({ url: '/pages/login/login' })
}
</script>

<template>
  <scroll-view scroll-y class="h-screen bg-stone-50">
    <view class="p-6 pb-24" :style="{ paddingTop: (statusBarHeight + 10) + 'px' }">
      <!-- Header -->
      <view class="flex justify-between items-center mb-8 pt-4">
        <view class="flex items-center gap-2">
          <view class="bg-stone-900 text-white p-1.5 rounded-lg">
            <text class="i-heroicons-book-open w-5 h-5 text-white"></text>
          </view>
          <text class="text-xl font-bold tracking-tight text-stone-800">StoryTrim</text>
        </view>
        <view @click="handleLogout" class="w-10 h-10 rounded-full bg-stone-200 flex items-center justify-center text-stone-500 font-bold text-xs">
          {{ userStore.username.charAt(0).toUpperCase() }}
        </view>
      </view>

      <!-- Upload Box -->
      <view @click="triggerUpload" class="mb-8 border-2 border-dashed border-stone-200 rounded-2xl p-8 flex flex-col items-center justify-center text-center hover:bg-stone-100 transition-colors active:scale-98">
        <view class="w-12 h-12 bg-white rounded-full shadow-sm flex items-center justify-center mb-3 text-teal-500">
          <text class="text-2xl">+</text>
        </view>
        <text class="font-bold text-stone-700">导入新书</text>
        <text class="text-xs text-stone-400 mt-1">支持 .txt 格式</text>
      </view>

      <!-- Renderjs Bridge -->
      <view :change:prop="filePicker.trigger" :prop="renderTrigger" :data-url="uploadUrl" :data-token="userToken" class="hidden file-picker-bridge"></view>

      <!-- Book List -->
      <text class="text-sm font-bold text-stone-400 uppercase tracking-wider mb-4 block">我的书架</text>
      <view class="grid grid-cols-1 gap-4">
        <view v-for="book in bookStore.books" :key="book.id" @click="handleBookClick(book)"
          class="bg-white p-4 rounded-2xl shadow-sm border border-stone-100 flex items-center gap-4 active:bg-stone-50">
          <view class="w-12 h-16 bg-stone-100 rounded-lg flex items-center justify-center text-stone-300">
            <text class="i-heroicons-document-text text-xl"></text>
          </view>
          <view class="flex-1">
            <view class="font-bold text-stone-800">{{ book.title }}</view>
            <view class="text-xs text-stone-400 mt-1">{{ book.total_chapters }} 章节</view>
          </view>
          <view class="text-teal-500 text-xs font-bold">{{ book.progress || 0 }}%</view>
        </view>
      </view>
    </view>
  </scroll-view>
</template>

<script module="filePicker" lang="renderjs">
export default {
  data() {
    return {
      inputEl: null
    }
  },
  mounted() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '*/*' 
    input.style.display = 'none'
    input.onchange = (e) => {
      const file = e.target.files[0]
      if (!file) return
      
      if (!file.name.toLowerCase().endsWith('.txt')) {
        alert('目前仅支持 .txt 格式的文件')
        return
      }

      this.uploadFile(file)
    }
    document.body.appendChild(input)
    this.inputEl = input
  },
  methods: {
    trigger(newValue, oldValue) {
      if (newValue > 0 && this.inputEl) {
        this.inputEl.click()
      }
    },
    uploadFile(file) {
      // 从 DOM 元素中读取配置
      const bridge = document.querySelector('.file-picker-bridge')
      const url = bridge.getAttribute('data-url')
      const token = bridge.getAttribute('data-token')

      this.$ownerInstance.callMethod('onUploadStart')

      const formData = new FormData()
      formData.append('file', file)

      const xhr = new XMLHttpRequest()
      xhr.open('POST', url, true)
      xhr.setRequestHeader('Authorization', 'Bearer ' + token)
      
      xhr.onload = () => {
        if (xhr.status === 200) {
          this.$ownerInstance.callMethod('onUploadSuccess')
        } else {
          this.$ownerInstance.callMethod('onUploadError', '服务器返回错误: ' + xhr.status)
        }
      }
      
      xhr.onerror = () => {
        this.$ownerInstance.callMethod('onUploadError', '网络连接失败，请检查后端是否开启及局域网IP是否正确')
      }

      xhr.send(formData)
    }
  }
}
</script>