<script setup lang="ts">
import { ref, computed, getCurrentInstance } from "vue";
import { onShow, onHide } from "@dcloudio/uni-app";
import { useUserStore } from "@/stores/user";
import { useBookStore } from "@/stores/book";
import { taskApi } from "@/api/task";
import BookCard from "@/components/BookCard.vue";
import DeleteConfirmModal from "@/components/DeleteConfirmModal.vue";
import BookActionSheet from "@/components/BookActionSheet.vue";
import TaskIndicator from "@/components/TaskIndicator.vue";
import TaskProgressSheet from "@/components/TaskProgressSheet.vue";
import LoginConfirmModal from "@/components/LoginConfirmModal.vue";
import SimpleAlertModal from "@/components/SimpleAlertModal.vue";

const userStore = useUserStore();
const bookStore = useBookStore();
const statusBarHeight = ref(uni.getSystemInfoSync().statusBarHeight || 0);
const renderTrigger = ref(0);

// 通用提示弹窗
const alertVisible = ref(false);
const alertMsg = ref("");
const alertTitle = ref("提示");

// 登录引导相关
const showLoginModal = ref(false);
const loginTipContent = ref("");

const openLoginModal = (msg: string) => {
  loginTipContent.value = msg;
  showLoginModal.value = true;
};

const handleLoginConfirm = () => {
  uni.navigateTo({ url: "/pages/login/login" });
};

// 获取组件实例，用于暴露方法给 renderjs
const instance = getCurrentInstance();

// --- RenderJS 交互逻辑 ---
const tempBookId = ref(0);
let bookIdPromise: Promise<number> | null = null;
let bookIdResolver: ((id: number) => void) | null = null;
let parseStartTime = 0;

const resetBookLock = () => {
  tempBookId.value = 0;
  bookIdPromise = new Promise((resolve) => {
    bookIdResolver = resolve;
  });
};

// 初始化锁
resetBookLock();

// 1. 接收书籍基本信息
const onBookInfo = async (info: { title: string; total: number; bookMD5: string; coverBase64?: string }) => {
  console.log("[Logic] onBookInfo called:", info.title, "MD5:", info.bookMD5, "Chapters:", info.total);
  if (info.coverBase64) {
      console.log("[Logic] Received cover image, length:", info.coverBase64.length);
      // 保存封面到本地
      try {
          // #ifdef APP-PLUS
          const fs = uni.getFileSystemManager();
          const fileName = `${info.bookMD5}.jpg`; // 统一存为 jpg
          const dir = '_doc/covers';
          
          // 确保目录存在 (异步稍微麻烦，这里假设已初始化或忽略错误)
          try { fs.accessSync(dir); } catch { try { fs.mkdirSync(dir, true); } catch(e){} }
          
          const filePath = `${dir}/${fileName}`;
          // 去掉 Base64 头部
          const base64Data = info.coverBase64.replace(/^data:image\/\w+;base64,/, "");
          
          fs.writeFile({
              filePath: filePath,
              data: base64Data,
              encoding: 'base64',
              success: () => console.log('[Logic] Cover saved to:', filePath),
              fail: (e) => console.error('[Logic] Save cover failed:', e)
          });
          // #endif
      } catch (e) {
          console.error('[Logic] Save cover process error:', e);
      }
  }

  parseStartTime = Date.now();
  resetBookLock();

  try {
    const id = await bookStore.createBookRecord(
      info.title,
      info.total,
      info.bookMD5,
    );
    console.log("[Logic] Book created with ID:", id);
    tempBookId.value = id;
    if (bookIdResolver) bookIdResolver(id);
  } catch (e: any) {
    console.error("[Logic] Create book error:", e);
    onUploadError(e.message);
    resetBookLock();
  }
};

// 2. 接收批量章节数据
const onBatchChapters = async (batch: { chapters: any[]; progress: number }) => {
  console.log("[Logic] onBatchChapters called, progress:", batch.progress);
  
  if (!bookIdPromise) {
    console.log("[Logic] No bookIdPromise, returning early");
    return;
  }
  
  try {
    const bookId = await bookIdPromise;
    console.log("[Logic] Inserting chapters for book:", bookId, "count:", batch.chapters.length);
    bookStore.uploadProgress = batch.progress;
    await bookStore.insertChapters(bookId, batch.chapters);
    console.log("[Logic] Chapters inserted successfully");
  } catch (e: any) {
    console.error("Batch insert failed", e);
  }
};

// 3. 完成
const onParseSuccess = async () => {
  const time = Date.now() - parseStartTime;
  console.log(`[Logic] Parse finished in ${time}ms`);

  bookStore.uploadProgress = 100;
  uni.hideLoading();
  uni.showToast({ title: "导入成功", icon: "success" });

  setTimeout(() => {
    bookStore.fetchBooks();
    bookStore.uploadProgress = 0;
  }, 1000);
};

const onUploadError = (msg: string) => {
  uni.hideLoading();
  bookStore.uploadProgress = 0;
  alertTitle.value = "导入失败";
  alertMsg.value = msg;
  alertVisible.value = true;
};

const showParsingLoading = () => {
  console.log("[Logic] showParsingLoading called");
};

// 暴露方法给 renderjs
if (instance && instance.proxy) {
  (instance.proxy as any).onBookInfo = onBookInfo;
  (instance.proxy as any).onBatchChapters = onBatchChapters;
  (instance.proxy as any).onParseSuccess = onParseSuccess;
  (instance.proxy as any).onUploadError = onUploadError;
  (instance.proxy as any).showParsingLoading = showParsingLoading;
}

// --- End RenderJS Logic ---

// 任务中心相关
const showTaskSheet = ref(false);
const hasActiveTasks = ref(false);

const refreshTasks = async () => {
  if (!userStore.isLoggedIn()) return;
  try {
    const res = await taskApi.getActiveTasksCount();
    if (res.code === 0) {
      hasActiveTasks.value = res.data?.has_active || false;
    }
  } catch (e) {
    console.warn('[Shelf] Check active tasks failed', e);
  }
};

const showTaskCenter = () => {
  showTaskSheet.value = true;
};

// 底部操作菜单相关
const showActionSheet = ref(false);
const actionBook = ref<any>(null);

onShow(async () => {
  // #ifdef APP-PLUS
  await bookStore.init();
  // #endif
  await bookStore.fetchBooks();
  
  // 检查是否有正在运行的任务
  await refreshTasks();
});

const handleBookClick = (book: any) => {
  bookStore.setActiveBook(book);
  uni.navigateTo({ url: `/pages/reader/reader?id=${book.id}` });
};

const triggerUpload = () => {
  console.log("[Logic] Triggering RenderJS");
  renderTrigger.value++;
};

const handleSyncBook = async (book: any) => {
  if (!userStore.isLoggedIn()) {
    openLoginModal('同步功能需要登录账号，登录后即可多端同步阅读进度。');
    return;
  }

  try {
    await bookStore.syncBookToCloud(book.id);
    uni.showToast({ title: "同步成功", icon: "success" });
    await bookStore.fetchBooks();
  } catch (e: any) {
    if (e.message && e.message.includes("已存在云端")) {
      uni.showToast({ title: "云端已存在本书", icon: "none" });
    } else {
      uni.showToast({ title: "同步失败", icon: "none" });
    }
    console.error("[Shelf] Sync error:", e);
  }
};

// 删除相关
const showDeleteModal = ref(false);
const bookToDelete = ref<any>(null);

const handleDeleteBook = (book: any) => {
  if (!userStore.isLoggedIn() && book.sync_state === 1) {
    uni.showToast({
      title: '该书籍为云端书籍，未登录状态下无法删除',
      icon: 'none',
      duration: 2000
    });
    return;
  }
  bookToDelete.value = book;
  showDeleteModal.value = true;
};

const handleBookOptions = (book: any) => {
  actionBook.value = book;
  showActionSheet.value = true;
  uni.vibrateShort({});
};

const handleSheetAction = (action: 'sync' | 'delete') => {
  if (!actionBook.value) return;
  
  if (action === 'sync') {
    handleSyncBook(actionBook.value);
  } else if (action === 'delete') {
    handleDeleteBook(actionBook.value);
  }
};

const confirmDelete = () => {
  if (bookToDelete.value) {
    bookStore.deleteBook(
      Number(bookToDelete.value.id), 
      bookToDelete.value.sync_state || 0, 
      bookToDelete.value.cloud_id
    );
    bookToDelete.value = null;
  }
};

const handleLogout = () => {
  userStore.logout();
  uni.reLaunch({ url: "/pages/login/login" });
};

// 过滤展示的书籍：未登录时隐藏仅云端书籍 (sync_state === 2)
const displayBooks = computed(() => {
  if (userStore.isLoggedIn()) {
    return bookStore.books;
  }
  return bookStore.books.filter(book => book.sync_state !== 2);
});
</script>

<template>
  <view class="flex flex-col h-screen bg-stone-50">
    <scroll-view scroll-y class="flex-1">
      <view
        class="p-6 pb-32"
        :style="{ paddingTop: statusBarHeight + 10 + 'px' }"
      >
        <!-- Header -->
        <view class="flex justify-between items-center mb-8 pt-4">
          <view class="flex items-center gap-2">
            <view class="bg-stone-900 text-white p-1.5 rounded-lg">
              <text class="text-xs font-bold">ST</text>
            </view>
            <text class="text-xl font-bold tracking-tight text-stone-800"
              >StoryTrim</text
            >
          </view>
          
          <view class="flex items-center gap-3">
            <view
              @click="handleLogout"
              class="w-10 h-10 rounded-full bg-stone-200 flex items-center justify-center text-stone-500 font-bold text-xs active:opacity-50"
            >
              {{ (userStore.username || "G").charAt(0).toUpperCase() }}
            </view>
          </view>
        </view>

        <!-- Upload Box -->
        <view
          @click="triggerUpload"
          class="mb-10 border border-stone-100 rounded-[2rem] p-10 flex flex-col items-center justify-center text-center bg-white shadow-[0_8px_30px_rgb(0,0,0,0.04)] active:scale-[0.98] transition-all duration-300"
        >
          <view
            class="w-12 h-12 bg-stone-50 text-stone-600 rounded-2xl shadow-inner border border-stone-100 flex items-center justify-center mb-4"
          >
            <image src="/static/icons/upload.svg" class="w-6 h-6 opacity-60" />
          </view>
          <text class="font-bold text-stone-900 tracking-tight">导入本地书籍</text>
          <text class="text-[10px] text-stone-400 mt-1 uppercase tracking-widest font-medium">支持 TXT / EPUB 格式</text>
        </view>

        <!-- Renderjs Bridge -->
        <view
          :change:prop="filePicker.trigger"
          :prop="renderTrigger"
          class="hidden"
        ></view>

        <!-- Book List Header -->
        <view class="flex items-end justify-between mb-5 px-1">
          <view>
            <text class="text-xs font-black text-stone-900 uppercase tracking-[0.2em]"
              >我的书架</text
            >
            <view class="w-4 h-0.5 bg-stone-900 mt-1"></view>
          </view>
          <text class="text-[10px] text-stone-400 font-medium tracking-wide"
            >{{ displayBooks.length }} 本</text
          >
        </view>

        <view class="flex flex-col">
          <BookCard
            v-for="book in displayBooks"
            :key="book.id"
            :book="book"
            @click="handleBookClick(book)"
            @sync="handleSyncBook(book)"
            @delete="handleDeleteBook"
            @longpress="handleBookOptions"
          />

          <view v-if="displayBooks.length === 0" class="py-20 text-center">
            <text class="text-stone-300 text-sm italic">书架空空如也</text>
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- FIXED COMPONENTS OUTSIDE SCROLL-VIEW -->

    <!-- Delete Confirm Modal -->
    <DeleteConfirmModal
      v-model:visible="showDeleteModal"
      title="删除书籍"
      content="确定删除本书吗？此操作不可恢复。"
      confirm-text="删除"
      @confirm="confirmDelete"
    />

    <!-- Custom Book Action Sheet -->
    <BookActionSheet
      v-model="showActionSheet"
      :title="actionBook?.title || '书籍操作'"
      :show-sync="actionBook?.sync_state === 0"
      @action="handleSheetAction"
    />

    <!-- Task Indicator (Floating Pill) - Only show when has tasks -->
    <TaskIndicator 
      :has-active-tasks="hasActiveTasks"
      @click="showTaskCenter"
    />

    <!-- Task Dashboard Sheet -->
    <TaskProgressSheet
      v-model="showTaskSheet"
      @update:modelValue="(val: boolean) => !val && refreshTasks()"
    />

    <!-- Upload Progress Modal -->
    <view
      v-if="bookStore.uploadProgress > 0 && bookStore.uploadProgress < 100"
      class="fixed inset-0 z-[200] bg-black/60 flex items-center justify-center"
    >
      <view class="bg-white p-6 rounded-2xl w-64 flex flex-col items-center">
        <view
          class="w-12 h-12 border-4 border-teal-100 border-t-teal-500 rounded-full animate-spin mb-4"
        ></view>
        <text class="font-bold text-lg mb-1"
          >{{ bookStore.uploadProgress }}%</text
        >
        <text class="text-xs text-stone-400">正在本地解析...</text>
      </view>
    </view>

    <!-- Sync Progress Modal -->
    <view
      v-if="bookStore.syncProgress > 0"
      class="fixed inset-0 z-[200] bg-black/60 flex items-center justify-center"
    >
      <view class="bg-white p-6 rounded-2xl w-64 flex flex-col items-center">
        <view
          class="w-12 h-12 border-4 border-blue-100 border-t-blue-500 rounded-full animate-spin mb-4"
        ></view>
        <text class="font-bold text-lg mb-1"
          >{{ bookStore.syncProgress }}%</text
        >
        <text class="text-xs text-stone-400">正在同步至云端...</text>
      </view>
    </view>

    <!-- Login Confirm Modal -->
    <LoginConfirmModal
      v-model:visible="showLoginModal"
      :content="loginTipContent"
      @confirm="handleLoginConfirm"
    />

    <!-- Simple Alert Modal -->
    <SimpleAlertModal
      v-model:visible="alertVisible"
      :title="alertTitle"
      :content="alertMsg"
    />
  </view>
</template>

<script module="filePicker" lang="renderjs">
import SparkMD5 from 'spark-md5'
import JSZip from 'jszip'

export default {
  methods: {
    // 归一化内容：去除所有非字母数字字符，转小写
    normalizeContent(content) {
      return content.replace(/[^a-zA-Z0-9\u4e00-\u9fa5]/g, '').toLowerCase();
    },

    // 计算章节内容的归一化 MD5
    calculateChapterMD5(content) {
      const normalized = this.normalizeContent(content);
      return SparkMD5.hash(normalized);
    },

    // 智能提取纯文本 (保留段落结构，移除样式和脚本)
    cleanHtmlContent(html) {
      if (!html) return '';
      
      // 1. 提取 body 内容 (如果存在)
      const bodyMatch = html.match(/<body[^>]*>([\s\S]*)<\/body>/i);
      let content = bodyMatch ? bodyMatch[1] : html;

      // 2. 移除 head, style, script 及其内容
      content = content.replace(/<head[^>]*>[\s\S]*?<\/head>/gi, '')
                       .replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
                       .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '');

      // 3. 将块级元素和换行标签替换为换行符
      content = content.replace(/<\/p>/gi, '\n')
                       .replace(/<\/div>/gi, '\n')
                       .replace(/<br\s*\/?>/gi, '\n')
                       .replace(/<\/h[1-6]>/gi, '\n\n'); // 标题后多空一行

      // 4. 移除剩余所有标签
      content = content.replace(/<[^>]+>/g, '');

      // 5. 处理常见实体
      content = content.replace(/&nbsp;/g, ' ')
                       .replace(/&lt;/g, '<')
                       .replace(/&gt;/g, '>')
                       .replace(/&amp;/g, '&')
                       .replace(/&quot;/g, '"')
                       .replace(/&apos;/g, "'")
                       .replace(/&[a-z]+;/g, ' '); // 兜底其他实体

      // 6. 规范化空白字符：将连续空行合并，去除每行首尾空格
      return content.split('\n')
                    .map(line => line.trim())
                    .filter(line => line.length > 0)
                    .join('\n'); // 暂时用单换行，阅读器端可能需要处理
    },

    trigger(newValue, oldValue, ownerInstance, instance) {
      if (newValue === 0) return

      const input = document.createElement('input')
      input.type = 'file'
      input.accept = '.txt,text/plain,.epub,application/epub+zip'
      input.style.display = 'none'
      document.body.appendChild(input)

      input.onchange = (res) => {
        const file = res.target.files[0]
        if (!file) {
          document.body.removeChild(input)
          return
        }

        const fileName = file.name.toLowerCase();
        const isEpub = fileName.endsWith('.epub');
        const isTxt = fileName.endsWith('.txt');

        if (!isEpub && !isTxt) {
           ownerInstance.callMethod('onUploadError', '不支持的文件格式，仅支持 TXT 或 EPUB');
           document.body.removeChild(input);
           return;
        }

        if (file.size > 200 * 1024 * 1024) {
           ownerInstance.callMethod('onUploadError', '文件过大(>200MB)');
           document.body.removeChild(input)
           return
        }

        ownerInstance.callMethod('showParsingLoading')

        if (isEpub) {
          const reader = new FileReader()
          reader.onload = (e) => {
            this.parseEpub(file.name, e.target.result, ownerInstance)
            document.body.removeChild(input)
          }
          reader.readAsArrayBuffer(file)
        } else {
          const reader = new FileReader()
          reader.onload = (e) => {
            this.parseAndUpload(file.name, e.target.result, ownerInstance)
            document.body.removeChild(input)
          }
          reader.readAsText(file)
        }
      }

      input.click()
    },

    // 辅助函数：无视命名空间查找元素 (解决部分 EPUB 解析不到 spine 的问题)
    getElements(xmlDoc, tagName) {
      // 1. 尝试直接获取
      let nodes = xmlDoc.getElementsByTagName(tagName);
      if (nodes.length > 0) return Array.from(nodes);
      
      // 2. 尝试带命名空间的获取 (常见 OPF 命名空间)
      nodes = xmlDoc.getElementsByTagNameNS("http://www.idpf.org/2007/opf", tagName);
      if (nodes.length > 0) return Array.from(nodes);
      
      // 3. 暴力遍历：匹配 localName
      const allNodes = xmlDoc.getElementsByTagName("*");
      const result = [];
      for (let i = 0; i < allNodes.length; i++) {
        // 兼容带前缀的情况 (如 opf:itemref)
        if (allNodes[i].localName === tagName || allNodes[i].nodeName === tagName || allNodes[i].nodeName.endsWith(':' + tagName)) {
          result.push(allNodes[i]);
        }
      }
      return result;
    },

    parseEpub: async function(fileName, data, ownerInstance) {
      console.log('[RenderJS] Parsing EPUB:', fileName);
      try {
        const zip = await JSZip.loadAsync(data);
        
        // 1. 寻找 container.xml 获取 OPF 路径
        const containerXml = await zip.file("META-INF/container.xml").async("string");
        // 兼容单引号和双引号
        const opfPathMatch = containerXml.match(/full-path=["']([^"']+)["']/);
        if (!opfPathMatch) throw new Error("无效的 EPUB 格式 (未找到 OPF)");
        
        const opfPath = opfPathMatch[1];
        console.log('[RenderJS] OPF Path:', opfPath);
        
        const lastSlashIndex = opfPath.lastIndexOf('/');
        const opfDir = lastSlashIndex !== -1 ? opfPath.substring(0, lastSlashIndex) : '';
        
        const opfContent = await zip.file(opfPath).async("string");
        const parser = new DOMParser();
        const xmlDoc = parser.parseFromString(opfContent, "text/xml");

        // 2. 获取书名 (尝试多种 tag 格式)
        let title = fileName.replace('.epub', '');
        const titleNodes = [
            ...this.getElements(xmlDoc, "title"),
            ...xmlDoc.getElementsByTagName("dc:title")
        ];
        if (titleNodes.length > 0 && titleNodes[0].textContent) {
            title = titleNodes[0].textContent;
        }

        // --- 尝试提取封面 ---
        let coverBase64 = null;
        try {
            // 1. 找 meta name="cover"
            const metaNodes = this.getElements(xmlDoc, "meta");
            let coverId = null;
            for (const meta of metaNodes) {
                if (meta.getAttribute("name") === "cover") {
                    coverId = meta.getAttribute("content");
                    break;
                }
            }
            
            // 2. 如果没找到，尝试找 manifest item properties="cover-image"
            if (!coverId) {
                 const items = this.getElements(xmlDoc, "item");
                 for (const item of items) {
                     if (item.getAttribute("properties") === "cover-image") {
                         coverId = item.getAttribute("id");
                         break;
                     }
                 }
            }

            // 3. 读取图片
            if (coverId) {
                const items = this.getElements(xmlDoc, "item");
                let coverHref = null;
                for (const item of items) {
                    if (item.getAttribute("id") === coverId) {
                        coverHref = item.getAttribute("href");
                        break;
                    }
                }

                if (coverHref) {
                    const decodedCoverHref = decodeURIComponent(coverHref);
                    const coverFullPath = opfDir ? `${opfDir}/${decodedCoverHref}` : decodedCoverHref;
                    const coverFile = zip.file(coverFullPath);
                    if (coverFile) {
                        console.log('[RenderJS] Found cover image:', coverFullPath);
                        const blob = await coverFile.async("blob");
                        // 简单转 Base64，实际生产环境建议用 Canvas 压缩
                        const reader = new FileReader();
                        coverBase64 = await new Promise((resolve) => {
                            reader.onloadend = () => resolve(reader.result);
                            reader.readAsDataURL(blob);
                        });
                        // 截断一下日志
                        console.log('[RenderJS] Cover size:', coverBase64.length);
                    }
                }
            }
        } catch (e) {
            console.warn('[RenderJS] Extract cover failed:', e);
        }
        // ------------------

        // 3. 解析 Manifest 和 Spine
        const manifest = {};
        const items = this.getElements(xmlDoc, "item");
        console.log(`[RenderJS] Found items in manifest: ${items.length}`);
        
        for (let i = 0; i < items.length; i++) {
          manifest[items[i].getAttribute("id")] = items[i].getAttribute("href");
        }

        const spine = [];
        const itemrefs = this.getElements(xmlDoc, "itemref");
        console.log(`[RenderJS] Found itemrefs in spine: ${itemrefs.length}`);
        
        for (let i = 0; i < itemrefs.length; i++) {
          spine.push(itemrefs[i].getAttribute("idref"));
        }
        
        console.log(`[RenderJS] Parsed Spine: ${spine.length} items`);

        // 4. 读取内容
        const chapters = [];
        for (let i = 0; i < spine.length; i++) {
          const id = spine[i];
          const href = manifest[id];
          if (!href) continue; 

          // 关键修复：解码 URL (如 "Chapter%201.html" -> "Chapter 1.html")
          const decodedHref = decodeURIComponent(href);
          
          // 简单的路径拼接 (暂不支持 ../ 等复杂相对路径，EPUB 规范通常不建议)
          const fullPath = opfDir ? `${opfDir}/${decodedHref}` : decodedHref;
          
          const file = zip.file(fullPath);
          if (!file) {
              // 尝试不带 opfDir 的情况 (部分不规范 EPUB)
              const fallbackFile = zip.file(decodedHref);
              if (fallbackFile) {
                  // console.log(`[RenderJS] Found file via fallback path: ${decodedHref}`);
                  // 虽然逻辑上不太可能，但为了鲁棒性
                  var targetFile = fallbackFile;
              } else {
                 console.warn(`[RenderJS] File not found in zip: ${fullPath} (original href: ${href})`);
                 continue;
              }
          }
          
          const targetZipFile = file || targetFile; 

          const html = await targetZipFile.async("string");
          // 使用新的清洗函数
          const chapterText = this.cleanHtmlContent(html);
          
          // 如果内容太短（可能是只有图片的封面页），跳过
          if (chapterText.length < 5) continue;

          // 提取可能的章节标题
          let chapterTitle = `第 ${chapters.length + 1} 章节`;
          
          // 提取标题时不使用 cleanHtmlContent，而是简单去除标签即可，防止过度清洗
          const simpleStrip = (s) => s.replace(/<[^>]+>/g, '').trim();
          
          // 优先尝试读取 title 标签
          const titleMatch = html.match(/<title[^>]*>(.*?)<\/title>/i);
          if (titleMatch && titleMatch[1]) {
             chapterTitle = simpleStrip(titleMatch[1]);
          } else {
             // 否则尝试找 h1-h2
             const hMatch = html.match(/<h[1-2][^>]*>(.*?)<\/h[1-2]>/i);
             if (hMatch) {
                chapterTitle = simpleStrip(hMatch[1]);
             }
          }

          chapters.push({
            index: chapters.length,
            title: chapterTitle,
            content: chapterText,
            md5: this.calculateChapterMD5(chapterText),
            length: [...chapterText].length
          });
        }
        
        console.log(`[RenderJS] Extracted ${chapters.length} valid chapters`);

        if (chapters.length === 0) {
            throw new Error("未能提取到有效章节内容，请确认文件是否加密或格式特殊");
        }

        // 5. 发送元数据
        console.log('[RenderJS] Calculating Book MD5...');
        const bookMD5 = SparkMD5.ArrayBuffer.hash(data);
        console.log('[RenderJS] Book MD5:', bookMD5);
        
        ownerInstance.callMethod('onBookInfo', {
          title: title,
          total: chapters.length,
          bookMD5: bookMD5,
          coverBase64: coverBase64 // 传递封面
        });

        // 6. 分批上传
        this.batchUpload(chapters, ownerInstance);

      } catch (e) {
        console.error('[RenderJS] EPUB Parse Error:', e);
        ownerInstance.callMethod('onUploadError', 'EPUB 解析失败: ' + e.message);
      }
    },

    batchUpload(chapters, ownerInstance) {
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
        setTimeout(sendNextBatch, 50);
      };

      sendNextBatch();
    },

    // 原有的 TXT 解析逻辑
    parseAndUpload(fileName, text, ownerInstance) {
      console.log('[RenderJS] parseAndUpload called, text length:', text.length);
      const CHAPTER_REGEX = /(?:^|\n)\s*(第[0-9一二三四五六七八九十百千万]+[章回节][^\r\n]*)/g;

      const matches = [...text.matchAll(CHAPTER_REGEX)];
      const chapters = [];

      if (matches.length > 0 && matches[0].index > 0) {
         const content = text.substring(0, matches[0].index);
         if (content.trim().length > 10) {
            const md5 = this.calculateChapterMD5(content.trim());
            chapters.push({ index: 0, title: '序章', content: content, md5: md5, length: content.length });
         }
      }

      for (let i = 0; i < matches.length; i++) {
        const m = matches[i];
        const title = m[1].trim();
        const start = m.index + m[0].length;
        const end = (i < matches.length - 1) ? matches[i+1].index : text.length;
        const content = text.substring(start, end);

        if (content.trim().length < 5) continue;

        const md5 = this.calculateChapterMD5(content.trim());
        chapters.push({
          index: chapters.length,
          title: title,
          content: content,
          md5: md5,
          length: [...content].length
        });
      }

      if (chapters.length === 0) {
         const md5 = this.calculateChapterMD5(text.trim());
         chapters.push({ index: 0, title: fileName.replace('.txt',''), content: text, md5: md5, length: [...text].length });
      }

      const bookMD5 = SparkMD5.hash(text);
      ownerInstance.callMethod('onBookInfo', {
        title: fileName.replace('.txt', ''),
        total: chapters.length,
        bookMD5: bookMD5
      });

      this.batchUpload(chapters, ownerInstance);
    }
  }
}
</script>

<style scoped>
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>