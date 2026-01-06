<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useBookStore, type Book } from '../stores/book'
import BookCard from '../components/BookCard.vue'

const router = useRouter()
const userStore = useUserStore()
const bookStore = useBookStore()
const fileInput = ref<HTMLInputElement | null>(null)

onMounted(() => {
  bookStore.fetchBooks()
})

const handleBookClick = (book: Book) => {
  bookStore.setActiveBook(book.id)
  router.push(`/reader/${book.id}`)
}

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleFileChange = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    await bookStore.uploadBook(file)
  } catch (e) {
    alert('上传失败')
  }
}
</script>

<template>
  <div class="h-full flex flex-col p-6 sm:p-8 overflow-y-auto pb-24 bg-[#fafaf9]">
    <!-- Header -->
    <header class="flex justify-between items-center mb-8">
      <div class="flex items-center gap-2">
        <div class="bg-stone-900 text-white p-1.5 rounded-lg">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path></svg>
        </div>
        <span class="text-xl font-bold tracking-tight text-stone-800">StoryTrim</span>
      </div>
      <div class="w-8 h-8 rounded-full bg-stone-200 flex items-center justify-center text-stone-500 font-bold text-xs cursor-pointer" @click="userStore.logout(); router.push('/login')">
        {{ userStore.username.charAt(0).toUpperCase() }}
      </div>
    </header>

    <!-- Upload Box -->
    <div @click="triggerUpload" class="mb-8 border-2 border-dashed border-stone-200 rounded-2xl p-8 flex flex-col items-center justify-center text-center hover:border-teal-400 hover:bg-teal-50 transition-colors cursor-pointer group select-none">
      <input type="file" ref="fileInput" class="hidden" accept=".txt" @change="handleFileChange" />
      <div class="w-12 h-12 bg-white rounded-full shadow-sm flex items-center justify-center mb-3 group-hover:scale-110 transition-transform text-teal-500">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
      </div>
      <h3 class="font-bold text-stone-700">导入新书</h3>
      <p class="text-xs text-stone-400 mt-1">支持 .txt 格式</p>
    </div>

    <!-- Book List -->
    <h3 class="text-sm font-bold text-stone-400 uppercase tracking-wider mb-4">我的书架</h3>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <BookCard v-for="book in bookStore.books" :key="book.id" :book="book" @click="handleBookClick(book)" />
    </div>
  </div>
</template>