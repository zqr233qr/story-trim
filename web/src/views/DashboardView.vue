<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { useBookStore } from '../stores/book'
import { useUserStore } from '../stores/user'
import { UploadCloud, Loader2, BookOpen, Clock } from 'lucide-vue-next'

const router = useRouter()
const bookStore = useBookStore()
const userStore = useUserStore()

const isLoading = ref(false)
const books = ref<any[]>([])
const loadingBooks = ref(false)

const fetchBooks = async () => {
  if (!userStore.isLoggedIn) return
  loadingBooks.value = true
  try {
    const res = await api.getBooks()
    if (res.data.code === 0) {
      books.value = res.data.data
    }
  } catch (err) {
    console.error('Failed to fetch books', err)
  } finally {
    loadingBooks.value = false
  }
}

onMounted(() => {
  fetchBooks()
})

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  isLoading.value = true
  try {
    const res = await api.upload(file)
    const data = res.data as any
    if (data.code === 0) {
      const d = data.data
      bookStore.setBook(d.book_id, d.filename, d.chapters)
      router.push('/reader')
    } else {
      alert('上传失败: ' + data.msg)
    }
  } catch (err) {
    alert('上传请求失败')
  } finally {
    isLoading.value = false
  }
}

const openBook = async (bookID: number) => {
  try {
    const res = await api.getBookDetail(bookID)
    const data = res.data as any
    if (data.code === 0) {
      const { book, trimmed_ids, reading_history } = data.data
      bookStore.setBook(book.id, book.title, book.chapters, trimmed_ids, reading_history)
      router.push('/reader')
    }
  } catch (err) {
    alert('打开书籍失败')
  }
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-[#FDFBF7]">
    <!-- Header -->
    <header class="bg-white/70 backdrop-blur-md sticky top-0 z-40 px-8 py-4 flex justify-between items-center border-b border-white/30">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 bg-gradient-to-tr from-teal-500 to-emerald-400 rounded-xl flex items-center justify-center shadow-md">
          <BookOpen class="text-white w-5 h-5" />
        </div>
        <span class="font-bold text-lg tracking-tight text-gray-800">StoryTrim</span>
      </div>
      <div class="flex items-center gap-6">
        <nav class="hidden md:flex gap-6 text-sm font-medium text-gray-500">
          <a href="#" class="text-gray-900">我的书架</a>
          <button v-if="userStore.isLoggedIn" @click="handleLogout" class="hover:text-red-500 transition-colors">退出</button>
          <router-link v-else to="/login" class="text-teal-600 font-bold">登录</router-link>
        </nav>
        <div class="w-10 h-10 rounded-full bg-teal-100 flex items-center justify-center text-teal-700 font-bold border-2 border-white shadow-sm" title="User">
          {{ userStore.username ? userStore.username.charAt(0).toUpperCase() : 'G' }}
        </div>      </div>
    </header>

    <main class="max-w-7xl mx-auto px-8 py-12">
      <!-- Upload Area -->
      <div class="mb-12 relative group cursor-pointer">
        <div class="absolute -inset-0.5 bg-gradient-to-r from-teal-400 to-emerald-500 rounded-3xl opacity-20 group-hover:opacity-40 transition duration-500 blur"></div>
        <div class="relative bg-white rounded-3xl p-10 text-center border border-gray-100 shadow-sm group-hover:shadow-md transition-all">
          <div class="w-16 h-16 bg-teal-50 text-teal-600 rounded-2xl flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
            <Loader2 v-if="isLoading" class="w-8 h-8 animate-spin" />
            <UploadCloud v-else class="w-8 h-8" />
          </div>
          <h2 class="text-2xl font-bold text-gray-800 mb-2">拖拽上传小说</h2>
          <p class="text-gray-400 max-w-md mx-auto">支持 .txt 格式。AI 将自动为您去水、精简，保留最纯粹的故事。</p>
          <input type="file" accept=".txt" @change="handleFileUpload" class="absolute inset-0 w-full h-full opacity-0 cursor-pointer" />
        </div>
      </div>

      <!-- Books Grid -->
      <div v-if="books.length > 0 || bookStore.currentBookId > 0">
        <h3 class="text-xl font-bold text-gray-800 mb-6 flex items-center gap-2">
          我的书架
          <span class="text-xs font-normal text-gray-400 bg-gray-100 px-2 py-1 rounded-md">{{ books.length }} 本</span>
        </h3>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <!-- 动态列表 -->
          <div 
            v-for="book in books" 
            :key="book.id"
            @click="openBook(book.id)"
            class="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer group"
          >
            <div class="flex gap-5">
              <div class="w-24 h-32 bg-gray-800 rounded-lg shadow-md shrink-0 flex items-center justify-center text-white font-serif relative overflow-hidden group-hover:shadow-lg transition-all">
                <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"></div>
                <span class="relative z-10 text-center px-2 line-clamp-2 text-sm">{{ book.title }}</span>
              </div>
              <div class="flex flex-col justify-between flex-1 py-1">
                <div>
                  <h4 class="font-bold text-lg text-gray-900 group-hover:text-teal-600 transition-colors line-clamp-1">{{ book.title }}</h4>
                  <p class="text-xs text-gray-400 mt-1 flex items-center gap-1">
                    <Clock class="w-3 h-3" /> {{ new Date(book.updated_at).toLocaleDateString() }}
                  </p>
                </div>
                <div>
                  <div class="flex items-center gap-2 text-xs text-gray-500 mb-2">
                    <span class="bg-teal-50 text-teal-700 px-1.5 py-0.5 rounded flex items-center gap-1">
                      {{ book.total_chapters }} 章节
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="loadingBooks" class="text-center py-12 text-gray-400">
        <Loader2 class="w-8 h-8 animate-spin mx-auto mb-2" />
        <p>加载书架中...</p>
      </div>

      <div v-else class="text-center py-12 text-gray-400 bg-white/50 rounded-3xl border border-dashed border-gray-200">
        <p>书架空空如也，快上传你的第一本书吧</p>
      </div>
    </main>
  </div>
</template>
