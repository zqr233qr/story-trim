<script lang="ts">
import { useBookStore } from '@/stores/book'
import { db } from '@/utils/sqlite'

// 逻辑层：负责接收数据并入库
export default {
  data() {
    return {
      tempBookId: 0,
      bookIdPromise: null as Promise<number> | null,
      bookIdResolver: null as ((id: number) => void) | null,
      parseStartTime: 0
    }
  },
  created() {
    this.resetBookLock()
  },
  methods: {
    resetBookLock() {
      this.tempBookId = 0
      this.bookIdPromise = new Promise(resolve => {
        this.bookIdResolver = resolve
      })
    },

    // 1. 接收书籍基本信息
    async onBookInfo(info: { title: string, total: number, fingerprint: string }) {
      console.log('[Logic] Creating book:', info.title, 'Chapters:', info.total)
      this.parseStartTime = Date.now()
      this.resetBookLock() // 重置锁
      
      const bookStore = useBookStore()
      bookStore.uploadProgress = 0
      
      try {
        const id = await bookStore.createBookRecord(info.title, info.fingerprint, info.total)
        this.tempBookId = id
        if (this.bookIdResolver) this.bookIdResolver(id) // 解锁
      } catch (e: any) {
        this.onUploadError('创建书籍失败: ' + e.message)
      }
    },

    // 2. 接收批量章节数据
    async onBatchChapters(batch: { chapters: any[], progress: number }) {
      // 等待书籍 ID 生成
      if (!this.bookIdPromise) return
      const bookId = await this.bookIdPromise
      
      const bookStore = useBookStore()
      bookStore.uploadProgress = batch.progress
      
      try {
        await bookStore.insertChapters(bookId, batch.chapters)
      } catch (e: any) {
        console.error('Batch insert failed', e)
      }
    },
// ...,

    // 3. 完成
    async onParseSuccess() {
      const time = Date.now() - this.parseStartTime
      console.log(`[Logic] Parse finished in ${time}ms`)
      
      const bookStore = useBookStore()
      bookStore.uploadProgress = 100
      uni.hideLoading()
      uni.showToast({ title: '导入成功', icon: 'success' })
      
      // 稍等刷新，让 Toast 显示一会
      setTimeout(() => {
        bookStore.fetchBooks()
        bookStore.uploadProgress = 0 // 重置进度条
      }, 1000)
    },

    onUploadError(msg: string) {
      uni.hideLoading()
      const bookStore = useBookStore()
      bookStore.uploadProgress = 0
      uni.showModal({ title: '导入失败', content: msg, showCancel: false })
    },
    
    // UI Loading 控制
    showParsingLoading() {
      // 实际上使用了自定义进度条，所以不需要 uni.showLoading
    }
  }
}
</script>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { useBookStore } from '@/stores/book'
import BookCard from '@/components/BookCard.vue'

const userStore = useUserStore()
const bookStore = useBookStore()
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0)
const renderTrigger = ref(0)

onShow(async () => {
  // #ifdef APP-PLUS
  await bookStore.init() 
  // #endif
  bookStore.fetchBooks()
})

const handleBookClick = (book: any) => {
  bookStore.setActiveBook(book)
  uni.navigateTo({ url: `/pages/reader/reader?id=${book.id}` })
}

const triggerUpload = () => {
  console.log('[Logic] Triggering RenderJS')
  renderTrigger.value++
}

const handleSyncBook = async (book: any) => {
  uni.showLoading({ title: '正在同步至云端...' })
  try {
    await bookStore.syncBookToCloud(book.id)
    uni.showToast({ title: '同步成功', icon: 'success' })
    // 重新拉取列表以更新 UI 状态
    await bookStore.fetchBooks()
  } catch (e) {
    uni.showToast({ title: '同步失败', icon: 'none' })
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
            <text class="text-xs font-bold">ST</text>
          </view>
          <text class="text-xl font-bold tracking-tight text-stone-800">StoryTrim</text>
        </view>
        <view @click="handleLogout" class="w-10 h-10 rounded-full bg-stone-200 flex items-center justify-center text-stone-500 font-bold text-xs active:opacity-50">
          {{ (userStore.username || 'G').charAt(0).toUpperCase() }}
        </view>
      </view>

      <!-- Upload Box -->
      <view @click="triggerUpload" class="mb-8 border-2 border-dashed border-stone-200 rounded-3xl p-10 flex flex-col items-center justify-center text-center bg-white/50 active:bg-stone-100 transition-colors">
        <view class="w-14 h-14 bg-stone-900 text-white rounded-full shadow-lg flex items-center justify-center mb-4">
          <text class="text-2xl">+</text>
        </view>
        <text class="font-bold text-stone-800">导入本地书籍</text>
        <text class="text-xs text-stone-400 mt-1">支持 TXT 格式</text>
      </view>

      <!-- Renderjs Bridge -->
      <view :change:prop="filePicker.trigger" :prop="renderTrigger" class="hidden"></view>

      <!-- Book List -->
      <view class="flex items-center justify-between mb-4 px-1">
        <text class="text-sm font-bold text-stone-400 uppercase tracking-widest">我的书架</text>
        <text class="text-[10px] text-stone-300">{{ bookStore.books.length }} 本书</text>
      </view>
      
      <view class="flex flex-col">
        <BookCard v-for="book in bookStore.books" :key="book.id" :book="book" 
          @click="handleBookClick(book)" 
          @sync="handleSyncBook(book)" />
        
        <view v-if="bookStore.books.length === 0" class="py-20 text-center">
          <text class="text-stone-300 text-sm italic">书架空空如也</text>
        </view>
      </view>
    </view>

    <!-- Upload Progress Modal -->
    <view v-if="bookStore.uploadProgress > 0 && bookStore.uploadProgress < 100" class="fixed inset-0 z-[200] bg-black/60 flex items-center justify-center">
      <view class="bg-white p-6 rounded-2xl w-64 flex flex-col items-center">
        <view class="w-12 h-12 border-4 border-teal-100 border-t-teal-500 rounded-full animate-spin mb-4"></view>
        <text class="font-bold text-lg mb-1">{{ bookStore.uploadProgress }}%</text>
        <text class="text-xs text-stone-400">正在本地解析...</text>
      </view>
    </view>

    <!-- Sync Progress Modal -->
    <view v-if="bookStore.syncProgress > 0" class="fixed inset-0 z-[200] bg-black/60 flex items-center justify-center">
      <view class="bg-white p-6 rounded-2xl w-64 flex flex-col items-center">
        <view class="w-12 h-12 border-4 border-blue-100 border-t-blue-500 rounded-full animate-spin mb-4"></view>
        <text class="font-bold text-lg mb-1">{{ bookStore.syncProgress }}%</text>
        <text class="text-xs text-stone-400">正在同步至云端...</text>
      </view>
    </view>
  </scroll-view>
</template>

<script module="filePicker" lang="renderjs">
export default {
  methods: {
    // 简易哈希算法 (RenderJS 内部使用)
    simpleHash(str) {
      // 极速采样: 头中尾各取100字符
      const len = str.length;
      if (len < 500) {
        let hash = 0;
        for (let i = 0; i < len; i++) {
          hash = ((hash << 5) - hash) + str.charCodeAt(i);
          hash |= 0;
        }
        return Math.abs(hash).toString(16);
      }
      let sample = str.substring(0, 100) + 
                   str.substring(Math.floor(len/2), Math.floor(len/2) + 100) + 
                   str.substring(len - 100);
      let hash = 0;
      for (let i = 0; i < sample.length; i++) {
        hash = ((hash << 5) - hash) + sample.charCodeAt(i);
        hash |= 0;
      }
      return Math.abs(hash).toString(16);
    },

    trigger(newValue, oldValue, ownerInstance, instance) {
      if (newValue === 0) return
      
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = '.txt,text/plain'
      input.style.display = 'none'
      document.body.appendChild(input)

      input.onchange = (res) => {
        const file = res.target.files[0]
        if (!file) {
          document.body.removeChild(input)
          return
        }
        
        // RenderJS 读取速度极快，200MB内通常没问题
        if (file.size > 200 * 1024 * 1024) {
           alert('文件过大(>200MB)')
           document.body.removeChild(input)
           return
        }

        ownerInstance.callMethod('showParsingLoading')
        
        const reader = new FileReader()
        reader.onload = (e) => {
          const text = e.target.result
          this.parseAndUpload(file.name, text, ownerInstance)
          document.body.removeChild(input)
        }
        reader.readAsText(file) 
      }
      
      input.click()
    },

    // 核心解析逻辑 (运行在 RenderJS 线程)
    parseAndUpload(fileName, text, ownerInstance) {
      const CHAPTER_REGEX = /(?:^|\n)\s*(第[0-9一二三四五六七八九十百千万]+[章回节][^\r\n]*)/g;
      
      // 1. 快速正则分章
      const matches = [...text.matchAll(CHAPTER_REGEX)];
      const chapters = [];
      let fingerprint = '';
      
      // 序章处理
      if (matches.length > 0 && matches[0].index > 0) {
         const content = text.substring(0, matches[0].index);
         if (content.trim().length > 10) {
            const md5 = this.simpleHash(content.trim());
            fingerprint = md5;
            chapters.push({ index: 0, title: '序章', content: content, md5: md5 });
         }
      }

      // 正文处理
      for (let i = 0; i < matches.length; i++) {
        const m = matches[i];
        const title = m[1].trim();
        const start = m.index + m[0].length;
        const end = (i < matches.length - 1) ? matches[i+1].index : text.length;
        const content = text.substring(start, end);
        
        // 忽略空章
        if (content.trim().length < 5) continue;

        const md5 = this.simpleHash(content.trim());
        if (!fingerprint) fingerprint = md5;

        chapters.push({
          index: chapters.length, // 重新编号
          title: title,
          content: content,
          md5: md5
        });
      }
      
      // 兜底：如果没匹配到章节，当做全文一章
      if (chapters.length === 0) {
         const md5 = this.simpleHash(text.trim());
         chapters.push({ index: 0, title: fileName.replace('.txt',''), content: text, md5: md5 });
         fingerprint = md5;
      }

      // 2. 发送元数据
      ownerInstance.callMethod('onBookInfo', {
        title: fileName.replace('.txt', ''),
        total: chapters.length,
        fingerprint: fingerprint
      });

      // 3. 分批发送章节数据 (避免 bridge 阻塞)
      const BATCH_SIZE = 50;
      let sentCount = 0;
      
      const sendNextBatch = () => {
        if (sentCount >= chapters.length) {
          ownerInstance.callMethod('onParseSuccess');
          return;
        }
        
        const end = Math.min(sentCount + BATCH_SIZE, chapters.length);
        const batch = chapters.slice(sentCount, end);
        
        const progress = Math.floor((end / chapters.length) * 100);
        ownerInstance.callMethod('onBatchChapters', {
          chapters: batch,
          progress: progress
        });
        
        sentCount = end;
        // 使用 setTimeout 释放 UI 线程
        setTimeout(sendNextBatch, 50);
      };
      
      sendNextBatch();
    }
  }
}
</script>

<style scoped>
.pb-safe { padding-bottom: env(safe-area-inset-bottom); }
</style>