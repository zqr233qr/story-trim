import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { api } from "@/api";
import {
  getBatchChapterContents,
  getBatchTrimmedById,
  syncTrimmedStatusById,
  updateReadingProgress,
  getChapterTrimStatusByMd5,
  getChapterTrimStatusById,
} from "@/api/content";
import { taskApi } from "@/api/task";
import type { Book, Chapter, Prompt } from "@/api";
import { useUserStore } from "./user";
// #ifdef APP-PLUS
import { AppRepository } from "@/adapter/app-repository";
import { db } from "@/utils/sqlite";
const repo = new AppRepository();
// #endif

export type { Book, Chapter, Prompt };

export interface BookUI extends Book {
  progress: number;
  chapters: any[];
  activeChapterIndex: number;
  activeModeId?: string;
  book_trimmed_ids?: number[];
  sync_state?: number; // 0: Local, 1: Synced, 2: CloudOnly
  cloud_id?: number;
  user_id?: number;
  full_trim_status?: string; // 'running' | 'completed' | 'failed'
  full_trim_progress?: number; // 0-100
  full_trim_task_id?: string;
  full_trim_prompt_id?: number;
  full_trim_available?: boolean;
}

export const useBookStore = defineStore("book", () => {
  const books = ref<BookUI[]>([]);
  const activeBook = ref<BookUI | null>(null);
  const prompts = ref<Prompt[]>([]);
  const isLoading = ref(false);
  const uploadProgress = ref(0);
  const syncProgress = ref(0);
  const downloadProgress = ref(0);

  // 初始化数据库
  const init = async () => {
    // #ifdef APP-PLUS
    await repo.init();
    // #endif
  };

  // 1. Fetch Books (Local First)
  const fetchBooks = async () => {
    isLoading.value = true;
    try {
      // #ifdef APP-PLUS
      const userStore = useUserStore();
      let cloudBooks: Book[] = [];

       // 仅在已登录时尝试从云端同步
       const currentUserId = userStore.isLoggedIn() ? Number(userStore.userId || 0) : 0;
       if (userStore.isLoggedIn()) {
         try {
           const res = await api.getBooks();
           // console.log('[FetchBooks] Cloud API result:', res.code, res.data?.length);
           if (res.code === 0) {
             cloudBooks = res.data || [];
             console.log(
               "[FetchBooks] Syncing",
               cloudBooks.length,
               "books from cloud",
             );
 
             for (const cloudBook of cloudBooks) {
               await repo.syncBookFromCloud(cloudBook, currentUserId);
             }
           }
         } catch (e) {
           console.warn("[FetchBooks] Failed to sync books from cloud", e);
         }
       }
 
       // 无论是否登录，都从本地加载书籍
       const localBooks = await repo.getBooks();
       const filteredBooks = localBooks.filter(
         (b) => !b.userId || b.userId === 0 || b.userId === currentUserId,
       );
       console.log("[FetchBooks] Local DB books count:", filteredBooks.length);


      const bookMap = new Map<string, any>();
      const noMd5Books: any[] = [];

       for (const b of filteredBooks) {

        if (b.bookMD5 && b.bookMD5.length > 5) {
          const key = b.bookMD5;
          if (!bookMap.has(key)) {
            bookMap.set(key, b);
          } else {
            const existing = bookMap.get(key);
            if ((b.syncState || 2) < (existing.syncState || 2)) {
              bookMap.set(key, b);
            }
          }
        } else {
          noMd5Books.push(b);
        }
      }

      const finalBooks = [...Array.from(bookMap.values()), ...noMd5Books];
      const cloudMap = new Map<number, Book>();
      cloudBooks.forEach((cb) => cloudMap.set(cb.id, cb));

      books.value = finalBooks.map((b) => {
        const cloudInfo = b.cloudId ? cloudMap.get(Number(b.cloudId)) : null;
           return {
             id: Number(b.id),
             title: b.title,
             book_md5: b.bookMD5,
             total_chapters: b.totalChapters,
             created_at: new Date(b.createdAt).toISOString(),
             progress: 0,
             activeChapterIndex: 0,
             chapters: [],
             sync_state: b.syncState || 0,
             cloud_id: b.cloudId,
             user_id: b.userId || 0,
             full_trim_status: cloudInfo?.full_trim_status,
             full_trim_progress: cloudInfo?.full_trim_progress,
           };


      });

      books.value.sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      );
      // #endif

      // #ifndef APP-PLUS
      const res = await api.getBooks();
      if (res.code === 0) {
        const data = res.data || [];
        books.value = data.map((b) => ({
          ...b,
          progress: 0,
          activeChapterIndex: 0,
          chapters: [],
          full_trim_status: b.full_trim_status,
          full_trim_progress: b.full_trim_progress,
        }));
      }
      // #endif
    } catch (e) {
      console.error("Failed to fetch books", e);
      if (!books.value) books.value = [];
    } finally {
      isLoading.value = false;
    }
  };

  // 2. Add Book (此方法仅用于 H5/MP，App 端走 RenderJS -> createBookRecord)
  const createBookRecord = async (
    title: string,
    total: number,
    bookMD5: string,
  ) => {
    // #ifdef APP-PLUS
    try {
      return await repo.createBook(title, total, bookMD5);
    } catch (e: any) {
      if (e.message && e.message.includes("已存在")) {
        throw new Error(e.message);
      }
      throw new Error("创建书籍失败：" + e.message);
    }
    // #endif
    return 0;
  };

  // 3. RenderJS 专用：批量插入章节
  const insertChapters = async (bookId: number, chapters: any[]) => {
    // #ifdef APP-PLUS
    await repo.insertChapters(bookId, chapters);
    // #endif
  };

  // 5. Fetch Book Detail (不批量同步精简状态，按需查询)
  const fetchBookDetail = async (bookId: number) => {
    // #ifdef APP-PLUS
    try {
      const book = await repo.getBook(bookId);
      console.log("[Store] Fetched book:", book);
      if (book) {
        const syncState = book.syncState || 0;

        if (syncState === 2) {
          const cloudBookId = book.cloudId || Number(book.id);
          const res = await api.getBookDetail(cloudBookId);
          if (res.code === 0) {
            activeBook.value = {
              id: Number(book.id),
              cloud_id: cloudBookId,
              title: res.data.book.title,
              book_md5: res.data.book.book_md5 || book.bookMD5,
              total_chapters: res.data.book.total_chapters,
              created_at: res.data.book.created_at,
              progress: 0,
              activeChapterIndex: 0,
              chapters: res.data.chapters.map((c) => ({
                id: Number(c.id),
                book_id: Number(c.book_id),
                index: c.index,
                title: c.title,
                md5: c.md5,
                cloud_id: Number(c.id),
                trimmed_prompt_ids: [],
                isLoaded: false,
                modes: {},
              })),
              sync_state: 2,
            };
          }
        } else {
          const chapters = await repo.getChapters(bookId);
          console.log("[Store] Fetched chapters count:", chapters.length);

          activeBook.value = {
            id: Number(book.id),
            cloud_id: book.cloudId,
            title: book.title,
            book_md5: book.bookMD5,
            total_chapters: book.totalChapters,
            created_at: new Date(book.createdAt).toISOString(),
            progress: 0,
            activeChapterIndex: 0,
            chapters: chapters.map((c) => ({
              id: Number(c.id),
              book_id: Number(c.bookId),
              index: c.index,
              title: c.title,
              md5: c.md5,
              cloud_id: c.cloudId,
              trimmed_prompt_ids: [],
              isLoaded: false,
              modes: {},
            })),
            sync_state: syncState,
          };
        }
      }
    } catch (e) {
      console.error("[Store] Fetch book detail failed", e);
    }
    return;
    // #endif
  };

  // 6. Fetch Chapter Content
  const fetchChapter = async (bookId: number, chapterId: number) => {
    if (!activeBook.value) return;
    const chapter = activeBook.value.chapters.find((c) => c.id === chapterId);
    if (!chapter) return;

    const syncState = activeBook.value.sync_state || 0;

    // #ifdef APP-PLUS
    if (syncState === 2) {
      const cloudChapterId = chapter.cloud_id || chapterId;
      const res = await getBatchChapterContents([cloudChapterId]);
      if (res.code === 0 && res.data[0] && res.data[0].content) {
        const content = res.data[0].content;
        const lines = content.split("\n");
        chapter.modes["original"] = lines;
        chapter.isLoaded = true;
        const cacheKey = `chapter_${cloudChapterId}`;
        uni.setStorageSync(cacheKey, lines);
      }
    } else {
      const content = await repo.getChapterContent(bookId, chapterId);
      const lines = content.split("\n");
      chapter.modes["original"] = lines;
      chapter.isLoaded = true;
      const cacheKey = `chapter_${chapterId}`;
      uni.setStorageSync(cacheKey, lines);
    }
    // #endif
  };

  // 辅助 Actions
  const setActiveBook = (book: BookUI) => {
    activeBook.value = book;
  };

  const setChapter = async (index: number) => {
    if (activeBook.value && activeBook.value.chapters[index]) {
      activeBook.value.activeChapterIndex = index;
      const chapter = activeBook.value.chapters[index];
      if (!chapter.isLoaded)
        await fetchChapter(activeBook.value.id, chapter.id);
    }
  };

  const fetchPrompts = async () => {
    try {
      const res = await api.getPrompts();
      if (res.code === 0) prompts.value = res.data;
    } catch (e) {
      console.warn("Failed to fetch prompts, using default");
      // 离线兜底 Prompt
      if (prompts.value.length === 0) {
        prompts.value = [
          {
            id: 1,
            name: "标准精简",
            description: "去除冗余，保留核心剧情",
            content: "",
            is_system: true,
            version: "1.0",
          },
          {
            id: 2,
            name: "极致浓缩",
            description: "仅保留主线脉络",
            content: "",
            is_system: true,
            version: "1.0",
          },
        ];
      }
    }
  };

  // 判断精简内容缓存是否有效（非空且含可见文字）
  const hasValidTrimLines = (lines: unknown): lines is string[] => {
    if (!Array.isArray(lines)) return false;
    return lines.some(
      (line) => typeof line === "string" && line.trim().length > 0,
    );
  };

  // 7. Fetch Trimmed Content (Local Cache Only)
  const fetchChapterTrim = async (
    bookId: number,
    chapterId: number,
    promptId: number,
  ): Promise<string[] | null> => {
    if (!activeBook.value) return null;
    const chapter = activeBook.value.chapters.find((c) => c.id === chapterId);
    if (!chapter) return null;

    const modeKey = `mode_${promptId}`;
    if (hasValidTrimLines(chapter.modes?.[modeKey]))
      return chapter.modes[modeKey];
    if (!chapter.modes) chapter.modes = {};

    const cacheKey = `trim_${chapterId}_${promptId}`;
    try {
      const cached = uni.getStorageSync(cacheKey);
      if (hasValidTrimLines(cached)) {
        chapter.modes[modeKey] = cached;
        return cached;
      }
    } catch (e) {
      console.warn("[Store] Read storage cache failed", e);
    }

    // 从云端获取
    const cloudChapterId = chapter.cloud_id || chapterId;
    try {
      const res = await getBatchTrimmedById([cloudChapterId], promptId);
      if (
        res.code === 0 &&
        res.data[0] &&
        res.data[0].trimmed_content !== undefined
      ) {
        const lines = res.data[0].trimmed_content.split("\n");
        if (hasValidTrimLines(lines)) {
          chapter.modes[modeKey] = lines;
          uni.setStorageSync(cacheKey, lines);
          return lines;
        }
      }
    } catch (e) {
      console.warn("[Store] Fetch trim from cloud failed", e);
    }

    return null;
  };

  // 手动保存精简内容 (供 UI 层调用)
  const saveChapterTrim = async (
    bookId: number,
    chapterId: number,
    promptId: number,
    content: string,
  ) => {
    if (!activeBook.value) return;
    const chapter = activeBook.value.chapters.find((c) => c.id === chapterId);
    if (!chapter) return;

    const lines = content.split("\n");
    chapter.modes[`mode_${promptId}`] = lines;

    const cacheKey = `trim_${chapterId}_${promptId}`;
    uni.setStorageSync(cacheKey, lines);
  };

  // 确保章节的精简状态是最新的（查询云端）
  const ensureTrimmedStatus = async (chapterId: number): Promise<boolean> => {
    if (!activeBook.value) return false;

    const chapter = activeBook.value.chapters.find((c) => c.id === chapterId);
    if (!chapter) return false;

    const syncState = activeBook.value.sync_state || 0;

    try {
      if (syncState === 0) {
        // 本地书籍：根据 MD5 查询
        if (chapter.md5) {
          const res = await getChapterTrimStatusByMd5(chapter.md5);
          if (res.code === 0 && res.data.prompt_ids) {
            chapter.trimmed_prompt_ids = res.data.prompt_ids;
            return res.data.prompt_ids.length > 0;
          }
        }
      } else {
        // 云端书籍：根据章节 ID 查询
        const cloudChapterId = chapter.cloud_id || chapterId;
        const res = await getChapterTrimStatusById(cloudChapterId, activeBook.value.book_md5, chapter.md5);
        if (res.code === 0 && res.data.prompt_ids) {
          chapter.trimmed_prompt_ids = res.data.prompt_ids;
          return res.data.prompt_ids.length > 0;
        }
      }
    } catch (e) {
      console.warn("[Store] Ensure trim status failed", e);
    }

    return false;
  };

  // 批量获取章节内容
  const fetchBatchChapters = async (ids: number[], promptId: number) => {
    if (!activeBook.value) return;
    const syncState = activeBook.value.sync_state || 0;

    // #ifdef APP-PLUS
    if (syncState === 2) {
      const cloudIds = ids.map((id) => {
        const chapter = activeBook.value?.chapters.find((c) => c.id === id);
        return chapter?.cloud_id || id;
      });
      try {
        const res = await getBatchChapterContents(cloudIds);
        if (res.code === 0 && res.data && activeBook.value) {
          for (let idx = 0; idx < res.data.length; idx++) {
            const item = res.data[idx];
            if (item && item.content) {
              const chapter = activeBook.value.chapters.find(
                (c) => (c.cloud_id || c.id) === cloudIds[idx],
              );
              if (chapter) {
                const lines = item.content.split("\n");
                chapter.modes["original"] = lines;
                chapter.isLoaded = true;
                const cacheKey = `chapter_${cloudIds[idx]}`;
                uni.setStorageSync(cacheKey, lines);
              }
            }
          }
        }
      } catch (e) {
        console.warn("[Store] Fetch chapters from cloud failed", ids, e);
      }
    } else {
      for (const id of ids) {
        const chapter = activeBook.value.chapters.find((c) => c.id === id);
        if (!chapter) continue;
        if (!chapter.isLoaded) {
          try {
            const content = await repo.getChapterContent(
              activeBook.value!.id,
              id,
            );
            const lines = content.split("\n");
            chapter.modes["original"] = lines;
            chapter.isLoaded = true;
            const cacheKey = `chapter_${id}`;
            uni.setStorageSync(cacheKey, lines);
          } catch (e) {
            console.warn("[Store] Fetch chapter from local failed", id, e);
          }
        }
      }
    }
    // #endif
  };
  const updateProgress = async (
    bookId: number,
    chapterId: number,
    promptId: number,
  ) => {
    if (!activeBook.value) return;

    // 1. 立即保存到本地 SQLite
    try {
      // #ifdef APP-PLUS
      await repo.updateProgress(bookId, chapterId, promptId);
      // #endif
    } catch (e) {
      console.warn("[Progress] Save to local failed", e);
    }

    // 2. 异步上报云端 (sync_state=1/2)
    const cloudBookId = activeBook.value.cloud_id;
    if (cloudBookId) {
      setTimeout(async () => {
        try {
          await updateReadingProgress(cloudBookId, chapterId, promptId);
        } catch (e) {
          console.warn("[Progress] Sync to cloud failed", e);
        }
      }, 1000);
    }
  };

  // 上传书籍元数据建立同步 (App端)
  const syncBookToCloud = async (bookId: number) => {
    // #ifdef APP-PLUS
    const book = await repo.getBook(bookId);
    if (!book || book.platform !== "app") return;

    const total = book.totalChapters;
    if (total === 0) return;

    syncProgress.value = 1;
     try {
       const res = await repo.uploadBookZip(bookId, (progress) => {
         syncProgress.value = Math.max(
           syncProgress.value,
           Math.min(100, Math.floor(progress)),
         );
       });
 
       const cloudBookId = res.book_id;
       const mappings = res.chapter_mappings || [];
       const userId = Number(useUserStore().userId || 0);


      const mappingStart = Date.now();
      const batchSize = 200;
      await db.transaction(async () => {
        for (let idx = 0; idx < mappings.length; idx += batchSize) {
          const batch = mappings.slice(idx, idx + batchSize);
          if (batch.length === 0) continue;

          const whenParts: string[] = [];
          const params: Array<number> = [];
          const ids: number[] = [];

          for (const mapping of batch) {
            whenParts.push("WHEN ? THEN ?");
            params.push(mapping.local_id, mapping.cloud_id);
            ids.push(mapping.local_id);
          }

          const idPlaceholders = ids.map(() => "?").join(", ");
          const sql = `UPDATE chapters SET cloud_id = CASE id ${whenParts.join(" ")} END WHERE id IN (${idPlaceholders})`;
          await db.execute(sql, [...params, ...ids]);
        }

         await db.execute(
           "UPDATE books SET cloud_id = ?, sync_state = 1, user_id = ? WHERE id = ?",
           [cloudBookId, userId, bookId],
         );

      });
      console.log("[Sync] Mapping update cost", {
        costMs: Date.now() - mappingStart,
        mappings: mappings.length,
        batches: Math.ceil(mappings.length / batchSize),
      });
    } catch (e) {
      console.error("[Sync] Failed:", e);
      throw e;
    } finally {
      setTimeout(() => {
        syncProgress.value = 0;
      }, 1000);
    }
    // #endif
  };

  // 下载云端书籍到本地 (App端)
  const downloadBookFromCloud = async (book: BookUI) => {
    // #ifdef APP-PLUS
    if (!book.cloud_id) {
      throw new Error("云端书籍ID为空");
    }

    downloadProgress.value = 1;
     try {
       const userId = Number(useUserStore().userId || 0);
       await repo.downloadBookContent(book.id, book.cloud_id, userId, (progress) => {
         downloadProgress.value = progress;
       });

      await fetchBooks();
    } catch (e) {
      console.error("[Download] Failed:", e);
      throw e;
    } finally {
      setTimeout(() => {
        downloadProgress.value = 0;
      }, 1000);
    }
    // #endif
  };

  const activeChapter = computed(() => {
    if (!activeBook.value) return null;
    return activeBook.value.chapters[activeBook.value.activeChapterIndex];
  });

  // 启动全文精简任务（不再管理轮询，任务中心组件自己管理）
  const startFullTrimTask = async (
    bookId: number,
    promptId: number,
  ): Promise<boolean> => {
    const res = await taskApi.startFullTrim(bookId, promptId);
    if (res.code === 0) {
      const book = books.value.find(
        (b) =>
          String(b.cloud_id) === String(bookId) ||
          String(b.id) === String(bookId),
      );
      if (book) {
        book.full_trim_status = "pending";
        book.full_trim_progress = 0;
        book.full_trim_task_id = res.data.task_id;
        book.full_trim_prompt_id = promptId;
      }
    }
    return res.code === 0;
  };

  const deleteBook = async (
    bookId: number,
    syncState: number,
    cloudId?: number,
  ) => {
    console.log("[DeleteBook] params:", { bookId, syncState, cloudId });
    try {
      // #ifdef APP-PLUS
      if (cloudId) {
        console.log(
          "[DeleteBook] calling api.deleteBook with cloudId:",
          cloudId,
        );
        const res = await api.deleteBook(cloudId);
        if (res.code !== 0) {
          throw new Error(res.msg || "云端书籍不存在");
        }
      }
      console.log("[DeleteBook] calling repo.deleteBook with bookId:", bookId);
      await repo.deleteBook(bookId);
      // #endif

      // #ifndef APP-PLUS
      console.log("[DeleteBook] calling api.deleteBook with bookId:", bookId);
      const res = await api.deleteBook(bookId);
      if (res.code !== 0) {
        throw new Error(res.msg || "云端书籍不存在");
      }
      // #endif

      books.value = books.value.filter((b) => Number(b.id) !== bookId);
      uni.showToast({ title: "已删除", icon: "none" });
    } catch (e: any) {
      console.error("[DeleteBook] 删除失败:", e);
      uni.showToast({ title: e?.message || "云端书籍不存在", icon: "none" });
    }
  };

  return {
    books,
    activeBook,
    prompts,
    isLoading,
    uploadProgress,
    syncProgress,
    downloadProgress,
    activeChapter,
    init,
    fetchBooks,
    fetchBookDetail,
    fetchChapter,
    createBookRecord,
    insertChapters,
    saveChapterTrim,
    setActiveBook,
    setChapter,
    fetchPrompts,
    fetchChapterTrim,
    fetchBatchChapters,
    updateProgress,
    syncBookToCloud,
    downloadBookFromCloud,
    deleteBook,
    startFullTrimTask,
    ensureTrimmedStatus,
  };
});
