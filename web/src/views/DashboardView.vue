<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Book } from '../api'
import { useBookStore } from '../stores/book'
import { Plus, BookOpen, Loader2, LogOut } from 'lucide-vue-next'
import { useUserStore } from '../stores/user'

const router = useRouter()
const bookStore = useBookStore()
const userStore = useUserStore()

const books = ref<Book[]>([])
const loading = ref(true)
const isUploading = ref(false)

const loadBooks = async () => {
  try {
    const res = await api.getBooks()
    if (res.data.code === 0) {
      books.value = res.data.data || [] // 防御性赋值
    }
  } finally {
    loading.value = false
  }
}

const handleFileUpload = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  isUploading.value = true
  try {
    const res = await api.upload(file)
    if (res.data.code === 0) {
      if (!books.value) books.value = []
      books.value.unshift(res.data.data) // 上传成功后立即添加到列表首部
    }
  } finally {
    isUploading.value = false
  }
}

const openBook = async (bookID: number) => {
  try {
    const res = await api.getBookDetail(bookID)
    if (res.data.code === 0) {
      const { book, chapters, trimmed_ids, reading_history } = res.data.data
      bookStore.setBook(book.id, book.title, chapters, trimmed_ids, reading_history)
      router.push('/reader')
    }
  } catch (err) {
    alert('打开书籍失败')
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

onMounted(() => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }
  loadBooks()
})
</script>

<template>
  <div class="min-h-screen bg-[#F9F7F1] p-8">
    <header class="flex justify-between items-center mb-12 max-w-5xl mx-auto">
      <h1 class="text-3xl font-serif font-bold text-slate-800">My Library</h1>
      <button @click="userStore.logout(); router.push('/login')" class="p-2 text-slate-400 hover:text-slate-800 transition-colors">
        <LogOut class="w-5 h-5" />
      </button>
    </header>

    <div class="max-w-5xl mx-auto grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
      <!-- Upload Card -->
      <label class="aspect-[3/4] border-2 border-dashed border-gray-300 rounded-xl flex flex-col items-center justify-center cursor-pointer hover:border-teal-500 hover:bg-white transition-all group relative overflow-hidden">
        <input type="file" accept=".txt" class="hidden" @change="handleFileUpload" :disabled="isUploading" />
        <div v-if="isUploading" class="absolute inset-0 bg-white/80 flex items-center justify-center z-10">
          <Loader2 class="w-8 h-8 animate-spin text-teal-600" />
        </div>
        <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mb-4 group-hover:bg-teal-50 transition-colors">
          <Plus class="w-6 h-6 text-gray-400 group-hover:text-teal-600" />
        </div>
        <span class="text-sm font-bold text-gray-400 group-hover:text-teal-600">Import Book</span>
      </label>

      <!-- Book Cards -->
      <div v-for="book in books" :key="book.id" @click="openBook(book.id)" class="aspect-[3/4] bg-white rounded-xl shadow-sm hover:shadow-xl transition-all cursor-pointer p-6 flex flex-col justify-between border border-gray-100 hover:-translate-y-1">
        <div class="w-10 h-10 bg-teal-50 rounded-lg flex items-center justify-center">
          <BookOpen class="w-5 h-5 text-teal-600" />
        </div>
        <div>
          <h3 class="font-bold text-slate-800 line-clamp-2 mb-1">{{ book.title }}</h3>
          <div class="flex justify-between items-center">
            <p class="text-[10px] text-gray-400">{{ book.total_chapters }} Chapters</p>
            <p class="text-[10px] text-gray-300">{{ formatDate(book.created_at) }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>