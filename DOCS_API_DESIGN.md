# StoryTrim API 接口文档 (v1.0)

## 文档说明

本文档记录 StoryTrim 项目的完整 API 接口设计，包括：
- 接口路径与请求方法
- 前端调用时机与业务逻辑
- 请求参数与返回格式
- App端和小程序端的差异化处理

**最后更新：** 2026-01-10
**状态：** 设计中

---

## 接口设计原则

1. **RESTful风格**：使用标准的 HTTP 方法和路径
2. **统一响应格式**：所有接口返回统一的JSON结构
3. **分页限制**：批量查询接口限制单次最多10条
4. **MD5寻址 vs ID寻址**：根据sync_state决定，sync_state=0/1优先使用ID，sync_state=2与小程序一致
5. **三级缓存策略**：Memory → Storage/SQLite → Network

---

## 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "success",
  "data": { ... }
}
```

### 失败响应
```json
{
  "code": 400,
  "msg": "参数错误",
  "data": null
}
```

### 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权（Token无效或过期） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 1001 | 书籍已存在 |
| 1002 | 用户名或密码错误 |
| 1003 | 章节内容不存在 |

---

## 一、用户认证相关

### 1.1 用户注册

**接口路径：** `POST /api/v1/auth/register`

**调用时机：**
- 小程序端：用户首次使用，点击注册按钮
- App端：用户首次使用，点击注册按钮

**请求参数：**
```json
{
  "username": "user123",
  "password": "password123"
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": 1,
    "username": "user123",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**前端处理逻辑：**
```typescript
// 1. 调用注册接口
const res = await request({ url: '/auth/register', method: 'POST', data })

// 2. 保存Token到本地存储
if (res.code === 0) {
  uni.setStorageSync('token', res.data.token)
  uni.setStorageSync('userId', res.data.user_id)
  uni.setStorageSync('username', res.data.username)

  // 3. 跳转到书架页
  uni.redirectTo({ url: '/pages/shelf/shelf' })
}
```

---

### 1.2 用户登录

**接口路径：** `POST /api/v1/auth/login`

**调用时机：**
- 小程序端：用户点击登录按钮
- App端：用户点击登录按钮
- 自动登录：App启动时检测本地有Token，静默登录验证

**请求参数：**
```json
{
  "username": "user123",
  "password": "password123"
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": 1,
    "username": "user123",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**前端处理逻辑：**
```typescript
// 1. 调用登录接口
const res = await request({ url: '/auth/login', method: 'POST', data })

// 2. 保存Token到本地存储
if (res.code === 0) {
  uni.setStorageSync('token', res.data.token)
  uni.setStorageSync('userId', res.data.user_id)

  // 3. 清空本地其他用户的数据
  // #ifdef APP-PLUS
  const currentUserId = uni.getStorageSync('userId')
  if (currentUserId && currentUserId !== res.data.user_id) {
    // 切换账号，清空本地数据
    await db.execute("DELETE FROM books WHERE user_id != ?", [res.data.user_id])
  }
  // #endif

  // 4. 跳转到书架页
  uni.redirectTo({ url: '/pages/shelf/shelf' })
}
```

---

## 二、公共接口（无需登录）

### 2.1 获取精简模式列表

**接口路径：** `GET /api/v1/common/prompts`

**调用时机：**
- App端：应用启动时，初始化精简模式列表
- 小程序端：应用启动时，初始化精简模式列表
- 设置页：用户进入设置页时刷新

**请求参数：** 无

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "id": 1,
      "name": "标准精简",
      "description": "去除冗余，保留核心剧情",
      "is_default": true,
      "target_ratio_min": 0.5,
      "target_ratio_max": 0.6
    },
    {
      "id": 2,
      "name": "极致浓缩",
      "description": "仅保留主线脉络",
      "is_default": false,
      "target_ratio_min": 0.3,
      "target_ratio_max": 0.4
    }
  ]
}
```

**前端处理逻辑：**
```typescript
// 1. 调用接口获取提示词列表
const res = await request({ url: '/common/prompts', method: 'GET' })

// 2. 存储到Pinia Store
if (res.code === 0) {
  bookStore.prompts = res.data

  // 3. 设置默认模式
  if (!bookStore.activeModeId) {
    const defaultPrompt = res.data.find(p => p.is_default)
    if (defaultPrompt) {
      bookStore.activeModeId = defaultPrompt.id.toString()
    }
  }
}
```

---

## 三、书籍管理相关

### 3.1 获取用户云端书架

**接口路径：** `GET /api/v1/books`

**调用时机：**
- App端：用户登录成功后，拉取云端书架
- App端：用户切换账号后，重新拉取书架
- 小程序端：应用启动时，拉取云端书架
- 书架页：用户下拉刷新时

**请求参数：** 无（从Token中获取UserID）

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "id": 101,
      "user_id": 1,
      "book_md5": "a1b2c3d4e5f6...",
      "fingerprint": "f1e2d3c4b5a6...",
      "title": "斗破苍穹",
      "total_chapters": 1000,
      "created_at": "2026-01-10T10:00:00Z",
      "ongoing_task": {
        "task_id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "full-trim",
        "status": "running",
        "progress": 45,
        "created_at": "2026-01-10T10:00:00Z"
      }
    },
    {
      "id": 102,
      "user_id": 1,
      "book_md5": "b2c3d4e5f6a7...",
      "fingerprint": "g6h5i4j3k2l1...",
      "title": "完美世界",
      "total_chapters": 800,
      "created_at": "2026-01-09T15:30:00Z"
    }
  ]
}
```

**字段说明：**
- `ongoing_task`: 可选字段，仅当存在运行中的任务时返回
  - `task_id`: 任务ID（UUID）
  - `type`: 任务类型（`full-trim`-全书精简）
  - `status`: 任务状态（`pending`-等待中/`running`-运行中/`completed`-已完成/`failed`-失败）
  - `progress`: 任务进度（0-100整数）
  - `created_at`: 任务创建时间

**App端处理逻辑：**
```typescript
const userId = uni.getStorageSync('userId')

// 1. 调用接口获取云端书架
const res = await request({ url: '/books', method: 'GET' })

if (res.code === 0 && res.data.length > 0) {
  // 2. 遍历云端书籍，与本地数据库同步
  for (const cloudBook of res.data) {
    // 2.1 查询本地是否存在此书
    const localBook = await db.select<any>(
      'SELECT id, sync_state FROM books WHERE book_md5 = ?',
      [cloudBook.book_md5]
    )

    if (localBook.length === 0) {
      // 情况A：本地book_md5不存在 → 插入本地book表（CLOUD_ONLY状态）
      await db.execute(
        `INSERT INTO books (user_id, cloud_id, book_md5, fingerprint, title, total_chapters, sync_state, created_at)
         VALUES (?, ?, ?, ?, ?, ?, 2, ?)`,
        [userId, cloudBook.id, cloudBook.book_md5, cloudBook.fingerprint, cloudBook.title, cloudBook.total_chapters, Date.now()]
      )

      console.log(`[Sync] Created cloud-only book: ${cloudBook.title}`)
    } else {
      // 情况B：本地book_md5存在 → 判断sync_state
      if (localBook[0].sync_state === 0) {
        // 只有sync_state=0（仅本地）才需要后续处理
        // 2.2.1 UPDATE本地book表（设置cloud_id和sync_state=1）
        await db.execute(
          'UPDATE books SET cloud_id = ?, sync_state = 1 WHERE id = ?',
          [cloudBook.id, localBook[0].id]
        )

        // 2.2.2 调用API获取章节目录（不获取原文内容）
        const chaptersRes = await request({ url: `/books/${cloudBook.id}`, method: 'GET' })

        if (chaptersRes.code === 0 && chaptersRes.data.chapters) {
          // 2.2.3 批量UPDATE本地chapters表的cloud_id字段
          for (const cloudChapter of chaptersRes.data.chapters) {
            await db.execute(
              'UPDATE chapters SET cloud_id = ? WHERE book_id = ? AND chapter_index = ?',
              [cloudChapter.id, localBook[0].id, cloudChapter.index]
            )
          }

          console.log(`[Sync] Updated local book: ${cloudBook.title}, synced ${chaptersRes.data.chapters.length} chapters`)
        }
      }
      // sync_state=1或2 → 跳过（已同步或仅云端记录）
    }
  }

  // 3. 重新加载本地书籍列表
  const localBooks = await db.select<any>('SELECT * FROM books WHERE user_id = ? ORDER BY created_at DESC', [userId])
  bookStore.books = localBooks.map(b => ({
    id: b.id,
    title: b.title,
    total_chapters: b.total_chapters,
    fingerprint: b.fingerprint,
    created_at: new Date(b.created_at).toISOString(),
    sync_state: b.sync_state,
    cloud_id: b.cloud_id
  }))
}
```

**小程序端处理逻辑：**
```typescript
// 1. 调用接口获取云端书架
const res = await request({ url: '/books', method: 'GET' })

if (res.code === 0) {
  // 2. 展示云端书籍，并处理任务状态
  bookStore.books = res.data.map(b => ({
    id: b.id,
    title: b.title,
    total_chapters: b.total_chapters,
    fingerprint: b.fingerprint,
    created_at: b.created_at,
    sync_state: 2,  // 小程序端都是CLOUD_ONLY状态
    ongoing_task: b.ongoing_task || null  // 保存任务状态
  }))

  // 3. 如果存在运行中的任务，定时刷新书架
  const hasRunningTask = res.data.some(b => b.ongoing_task?.status === 'running')
  if (hasRunningTask && !bookStore.refreshTimer) {
    bookStore.refreshTimer = setInterval(async () => {
      await bookStore.fetchBooks()
      const stillRunning = bookStore.books.some(b => b.ongoing_task?.status === 'running')
      if (!stillRunning) {
        clearInterval(bookStore.refreshTimer)
        bookStore.refreshTimer = null
      }
    }, 5000)  // 每5秒刷新一次
  }
}
```

---

### 3.2 获取书籍详情

**接口路径：** `GET /api/v1/books/:id`

**调用时机：**
- App端：用户点击书架上的某本书，进入阅读器前（仅当sync_state=2时）
- 小程序端：用户点击书架上的某本书，进入阅读器前

**请求参数：**
- 路径参数：`id`（云端书籍ID）
- Query参数：可选 `prompt_id`（精简模式ID）

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "book": {
      "id": 101,
      "user_id": 1,
      "book_md5": "a1b2c3d4e5f6...",
      "fingerprint": "f1e2d3c4b5a6...",
      "title": "斗破苍穹",
      "total_chapters": 1000
    },
    "chapters": [
      {
        "id": 5001,
        "book_id": 101,
        "index": 0,
        "title": "序章",
        "chapter_md5": "m1n2o3p4q5r6..."
      },
      {
        "id": 5002,
        "book_id": 101,
        "index": 1,
        "title": "第一章 天才少年",
        "chapter_md5": "a7b8c9d0e1f2..."
      }
    ],
    "trimmed_map": {
      "5001": [1, 2],
      "5002": [1]
    },
    "reading_history": {
      "last_chapter_id": 5002,
      "last_prompt_id": 1
    }
  }
}
```

**App端处理逻辑：**
```typescript
// 1. 获取本地书籍记录
const localBook = await db.select<any>('SELECT * FROM books WHERE id = ?', [bookId])

if (localBook.length > 0 && localBook[0].sync_state === 2) {
  // 2. 只有sync_state=2（CLOUD_ONLY）时才调用API获取云端书籍详情
  const res = await request({ url: `/books/${localBook[0].cloud_id}`, method: 'GET' })

  if (res.code === 0) {
    const { book, chapters, trimmed_map, reading_history } = res.data

    // 3. 设置为activeBook
    bookStore.activeBook = {
      id: localBook[0].id,
      title: book.title,
      total_chapters: book.total_chapters,
      fingerprint: book.fingerprint,
      chapters: chapters.map(ch => ({
        id: localBook[0].id,  // 使用本地章节ID
        book_id: book.id,           // 云端书籍ID
        index: ch.index,
        title: ch.title,
        chapter_md5: ch.chapter_md5,
        cloud_id: ch.id,          // 云端章节ID
        trimmed_prompt_ids: trimmed_map[ch.id] || [],
        isLoaded: false
      })),
      activeChapterIndex: reading_history ? chapters.findIndex(c => c.id === reading_history.last_chapter_id) : 0,
      activeModeId: reading_history?.last_prompt_id?.toString() || 'original'
    }

    // 4. 如果有阅读进度，自动跳转
    if (reading_history) {
      await bookStore.setChapter(bookStore.activeBook.activeChapterIndex)
    }
  }
}

**小程序端处理逻辑：**
```typescript
// 1. 调用API获取书籍详情
const res = await request({ url: `/books/${bookId}`, method: 'GET' })

if (res.code === 0) {
  const { book, chapters, trimmed_map, reading_history } = res.data

  // 2. 设置为activeBook
  bookStore.activeBook = {
    id: book.id,
    title: book.title,
    total_chapters: book.total_chapters,
    fingerprint: book.fingerprint,
    chapters: chapters.map(ch => ({
      id: ch.id,
      book_id: ch.book_id,
      index: ch.index,
      title: ch.title,
      chapter_md5: ch.chapter_md5,
      trimmed_prompt_ids: trimmed_map[ch.id] || [],
      isLoaded: false
    })),
    activeChapterIndex: reading_history ? chapters.findIndex(c => c.id === reading_history.last_chapter_id) : 0,
    activeModeId: reading_history?.last_prompt_id?.toString() || 'original'
  }

  // 3. 预加载第一张
  await bookStore.fetchChapter(bookId, bookStore.activeBook.chapters[0].id)
}
```

---

### 3.3 同步本地书籍到云端

**接口路径：** `POST /api/v1/books/sync-local`

**调用时机：**
- App端：用户在书架点击"同步"按钮
- App端：用户首次导入书籍后，自动提示同步

**请求参数：**
```json
{
  "book_name": "斗破苍穹",
  "book_md5": "a1b2c3d4e5f6...",
  "cloud_book_id": 0,  // 0表示新书，非0表示续传
  "total_chapters": 1000,
  "chapters": [
    {
      "local_id": 1,  // 本地章节ID
      "index": 0,
      "title": "序章",
      "md5": "m1n2o3p4q5r6...",
      "content": "序章内容...",
      "words_count": 500
    },
    {
      "local_id": 2,
      "index": 1,
      "title": "第一章 天才少年",
      "md5": "a7b8c9d0e1f2...",
      "content": "第一章内容...",
      "words_count": 3000
    }
  ]
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "book_id": 101,
    "chapters_map": {
      "1": 5001,  // local_id -> cloud_id
      "2": 5002
    }
  }
}
```

**App端处理逻辑：**
```typescript
async function syncBookToCloud(bookId: number) {
  // 1. 获取本地书籍信息
  const book = await db.select<any>('SELECT * FROM books WHERE id = ?', [bookId])

  // 2. 批量上传章节（每次200章）
  const BATCH_SIZE = 200
  let syncedCount = 0
  let cloudBookId = book.cloud_id || 0

  while (syncedCount < book.total_chapters) {
    // 2.1 获取本地章节（分批）
    const chapters = await db.select<any>(
      'SELECT id, chapter_index, title, md5, content, words_count FROM chapters WHERE book_id = ? ORDER BY chapter_index LIMIT ? OFFSET ?',
      [bookId, BATCH_SIZE, syncedCount]
    )

    if (chapters.length === 0) break

    // 2.2 构建请求参数
    const payload = {
      book_name: book.title,
      book_md5: book.book_md5,
      cloud_book_id: cloudBookId || 0,
      total_chapters: book.total_chapters,
      chapters: chapters.map(ch => ({
        local_id: ch.id,
        index: ch.chapter_index,
        title: ch.title,
        md5: ch.md5,
        content: ch.content,
        words_count: ch.words_count
      }))
    }

    // 2.3 调用API上传
    const res = await request({ url: '/books/sync-local', method: 'POST', data: payload })

    if (res.code === 0) {
      // 2.4 更新cloudBookId
      cloudBookId = res.data.book_id

      // 2.5 更新本地chapters表的cloud_id字段
      for (const ch of chapters) {
        const cloudChapterId = res.data.chapters_map[ch.id.toString()]
        if (cloudChapterId) {
          await db.execute(
            'UPDATE chapters SET cloud_id = ? WHERE id = ?',
            [cloudChapterId, ch.id]
          )
        }
      }

      // 2.6 更新进度
      syncedCount += chapters.length
      const progress = Math.floor((syncedCount / book.total_chapters) * 100)
      bookStore.syncProgress = progress
    } else {
      throw new Error(res.msg || '同步失败')
    }
  }

  // 3. 更新本地book表的cloud_id和sync_state
  await db.execute(
    'UPDATE books SET cloud_id = ?, sync_state = 1 WHERE id = ?',
    [cloudBookId, bookId]
  )

  uni.showToast({ title: '同步成功', icon: 'success' })

  // 4. 刷新书架
  await bookStore.fetchBooks()
}
```

---

### 3.4 更新阅读进度

**接口路径：** `POST /api/v1/books/:id/progress`

**调用时机：**
- App端：用户切换章节后，延迟5秒上报
- App端：用户退出阅读器时，立即上报
- 小程序端：用户切换章节后，立即上报

**请求参数：**
```json
{
  "chapter_id": 5002,
  "prompt_id": 1
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

**App端处理逻辑：**
```typescript
// 1. 本地SQLite实时记录
await db.execute(
  'INSERT OR REPLACE INTO reading_history (book_id, last_chapter_id, last_prompt_id, updated_at) VALUES (?, ?, ?, ?)',
  [bookId, chapterId, promptId || 0, Date.now()]
)

// 2. 异步上报云端（不阻塞）
request({ url: `/books/${bookId}/progress`, method: 'POST', data: { chapter_id, prompt_id } })
  .catch(err => {
    console.warn('Failed to upload progress:', err)
  })
```

**小程序端处理逻辑：**
```typescript
// 1. 直接上报云端
await request({ url: `/books/${bookId}/progress`, method: 'POST', data: { chapter_id, prompt_id } })

// 2. 更新本地Store
bookStore.activeBook.activeChapterIndex = newIndex
bookStore.activeBook.lastChapterId = chapterId
bookStore.activeBook.lastPromptId = promptId
```

---

### 3.5 删除书籍

**接口路径：** `DELETE /api/v1/books/:id`

**调用时机：**
- App端：用户在书架长按书籍，点击"删除"
- 小程序端：用户在书架点击书籍的"删除"按钮

**请求参数：** 无（路径参数：id为云端书籍ID）

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

**删除范围说明：**

**服务端处理：**
- Book表：软删除（设置`DeletedAt`字段）
- Chapter表：级联软删除（设置`DeletedAt`字段）
- ReadingHistory表：级联物理删除（用户隐私）
- UserProcessedChapter表：级联物理删除（用户足迹）
- Task表：保留任务记录（用于审计），标记BookID为空或添加DeletedAt

**保留数据：**
- ChapterContent表：不删除（全局共享内容）
- TrimResult表：不删除（全局共享精简结果）
- ChapterSummary表：不删除（全局共享摘要）
- SharedEncyclopedia表：不删除（全局共享百科）

**App端处理逻辑：**
```typescript
async function deleteBook(bookId: number) {
  // 1. 获取本地书籍信息
  const book = await db.select<any>('SELECT * FROM books WHERE id = ?', [bookId])

  if (book.length === 0) {
    uni.showToast({ title: '书籍不存在', icon: 'none' })
    return
  }

  const localBook = book[0]

  // 2. 根据sync_state决定是否调用云端API
  if (localBook.sync_state === 2) {
    // 降级模式：需要调用云端API删除
    if (localBook.cloud_id > 0) {
      const res = await request({ url: `/books/${localBook.cloud_id}`, method: 'DELETE' })
      if (res.code !== 0) {
        uni.showToast({ title: '删除失败', icon: 'none' })
        return
      }
    }
  }

  // 3. 删除本地SQLite记录
  await db.transaction(async () => {
    // 3.1 删除阅读进度
    await db.execute('DELETE FROM reading_history WHERE book_id = ?', [bookId])

    // 3.2 删除章节索引
    await db.execute('DELETE FROM chapters WHERE book_id = ?', [bookId])

    // 3.3 删除章节内容（contents表）
    await db.execute('DELETE FROM contents WHERE chapter_md5 IN (SELECT md5 FROM chapters WHERE book_id = ?)', [bookId])

    // 3.4 删除书籍记录
    await db.execute('DELETE FROM books WHERE id = ?', [bookId])
  })

  // 4. 清除Tier 1缓存（Pinia Store）
  bookStore.books = bookStore.books.filter(b => b.id !== bookId)
  if (bookStore.activeBook?.id === bookId) {
    bookStore.activeBook = null
  }

  // 5. 刷新书架
  await bookStore.fetchBooks()

  uni.showToast({ title: '删除成功', icon: 'success' })
}
```

**小程序端处理逻辑：**
```typescript
async function deleteBook(bookId: number) {
  // 1. 调用云端API删除
  const res = await request({ url: `/books/${bookId}`, method: 'DELETE' })

  if (res.code !== 0) {
    uni.showToast({ title: '删除失败', icon: 'none' })
    return
  }

  // 2. 清除Tier 2缓存（Storage）
  // 获取该书籍的所有章节MD5
  if (bookStore.activeBook && bookStore.activeBook.id === bookId) {
    for (const chapter of bookStore.activeBook.chapters) {
      uni.removeStorageSync(`st_content_${chapter.id}`)
      for (const promptId of chapter.trimmed_prompt_ids || []) {
        uni.removeStorageSync(`st_trim_${chapter.id}_${promptId}`)
      }
    }
  }

  // 3. 清除Tier 1缓存（Pinia Store）
  bookStore.books = bookStore.books.filter(b => b.id !== bookId)
  if (bookStore.activeBook?.id === bookId) {
    bookStore.activeBook = null
  }

  // 4. 刷新书架
  await bookStore.fetchBooks()

  uni.showToast({ title: '删除成功', icon: 'success' })
}
```

---

## 四、书籍解析与上传相关

### 4.1 获取章节解析规则

**接口路径：** `GET /api/v1/chapters/parse-rules`

**调用时机：**
- App端：导入新文件前，获取最新的解析规则
- App端：解析失败时，获取备用规则

**请求参数：** 无

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "title_regex": "^第[0-9零一二三四五六七八九十百千]+章[\\s\\S]*$",
    "default_rules": [
      {
        "name": "标准格式",
        "pattern": "^第[0-9零一二三四五六七八九十百千]+章[\\s\\S]*$",
        "description": "匹配第1章、第一章等格式"
      },
      {
        "name": "简化格式",
        "pattern": "^[0-9]+\\.[\\s\\S]*$",
        "description": "匹配1.、2.等格式"
      }
    ],
    "fallback_enabled": true
  }
}
```

**App端处理逻辑：**
```typescript
async function getParseRules() {
  // 1. 尝试从服务端获取规则
  try {
    const res = await request({ url: '/chapters/parse-rules', method: 'GET' })

    if (res.code === 0) {
      // 2. 保存规则到本地缓存
      uni.setStorageSync('parse_rules', res.data)

      // 3. 使用服务端规则解析
      return res.data
    }
  } catch (err) {
    console.warn('Failed to fetch parse rules from server, using local fallback:', err)
  }

  // 4. 兜底机制：使用本地默认规则
  const fallbackRules = {
    title_regex: "^第[0-9零一二三四五六七八九十百千]+章[\\s\\S]*$",
    default_rules: [
      {
        name: "标准格式",
        pattern: "^第[0-9零一二三四五六七八九十百千]+章[\\s\\S]*$",
        description: "匹配第1章、第一章等格式"
      }
    ],
    fallback_enabled: true
  }

  return fallbackRules
}

// 使用规则解析章节
async function parseChapters(content: string) {
  const rules = await getParseRules()

  // 尝试多个规则
  for (const rule of rules.default_rules) {
    const regex = new RegExp(rule.pattern, 'gm')
    const matches = content.match(regex)

    if (matches && matches.length > 10) {
      // 找到足够的章节标题，使用此规则
      console.log(`Using rule: ${rule.name}`)
      return extractChapters(content, regex)
    }
  }

  throw new Error('无法识别章节格式，请检查文件内容')
}
```

---

### 4.2 创建书籍（上传元数据）

**接口路径：** `POST /api/v1/books/create`

**调用时机：**
- App端：用户导入新文件后，先上传书籍元数据

**请求参数：**
```json
{
  "book_name": "斗破苍穹",
  "book_md5": "a1b2c3d4e5f6...",
  "fingerprint": "f1e2d3c4b5a6...",
  "total_chapters": 1000
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "book_id": 101
  }
}
```

**App端处理逻辑：**
```typescript
async function createBook(fileContent: string) {
  // 1. 计算BookMD5（全文计算）
  const bookMD5 = calculateMD5(fileContent)

  // 2. 解析章节
  const chapters = await parseChapters(fileContent)

  // 3. 计算Fingerprint（第一章MD5）
  const fingerprint = calculateMD5(chapters[0].content)

  // 4. 调用API创建书籍
  const res = await request({
    url: '/books/create',
    method: 'POST',
    data: {
      book_name: extractTitle(fileContent),
      book_md5: bookMD5,
      fingerprint: fingerprint,
      total_chapters: chapters.length
    }
  })

  if (res.code === 0) {
    const cloudBookId = res.data.book_id

    // 5. 保存到本地SQLite
    await db.execute(
      'INSERT INTO books (user_id, cloud_id, book_md5, fingerprint, title, total_chapters, sync_state, created_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?)',
      [userId, cloudBookId, bookMD5, fingerprint, extractTitle(fileContent), chapters.length, Date.now()]
    )

    // 6. 保存章节索引
    const localBookId = await getLastInsertId()
    for (const chapter of chapters) {
      const chapterMD5 = calculateMD5(chapter.content)
      await db.execute(
        'INSERT INTO chapters (book_id, chapter_index, title, md5, words_count) VALUES (?, ?, ?, ?, ?)',
        [localBookId, chapter.index, chapter.title, chapterMD5, chapter.content.length]
      )

      // 7. 保存章节内容到contents表
      await db.execute(
        'INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)',
        [chapterMD5, chapter.content]
      )
    }

    // 8. 开始分批上传
    await uploadChapters(localBookId, cloudBookId, chapters)
  }
}
```

---

### 4.3 分批上传章节内容

**接口路径：** `POST /api/v1/books/:id/chapters/upload`

**调用时机：**
- App端：创建书籍后，分批上传章节内容

**请求参数：**
```json
{
  "chapters": [
    {
      "local_index": 0,
      "title": "序章",
      "md5": "m1n2o3p4q5r6...",
      "content": "序章内容...",
      "words_count": 500
    },
    {
      "local_index": 1,
      "title": "第一章 天才少年",
      "md5": "a7b8c9d0e1f2...",
      "content": "第一章内容...",
      "words_count": 3000
    }
  ]
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "chapters_map": {
      "0": 5001,
      "1": 5002
    }
  }
}
```

**App端处理逻辑：**
```typescript
async function uploadChapters(localBookId: number, cloudBookId: number, chapters: any[]) {
  const BATCH_SIZE = 200  // 每批200章
  let uploadedCount = 0

  while (uploadedCount < chapters.length) {
    // 1. 获取当前批次
    const batch = chapters.slice(uploadedCount, uploadedCount + BATCH_SIZE)

    // 2. 构建请求参数
    const payload = {
      chapters: batch.map(ch => ({
        local_index: ch.index,
        title: ch.title,
        md5: ch.md5,
        content: ch.content,
        words_count: ch.content.length
      }))
    }

    // 3. 调用API上传
    const res = await request({
      url: `/books/${cloudBookId}/chapters/upload`,
      method: 'POST',
      data: payload
    })

    if (res.code === 0) {
      // 4. 更新本地chapters表的cloud_id
      for (const chapter of batch) {
        const cloudChapterId = res.data.chapters_map[chapter.index.toString()]
        if (cloudChapterId) {
          await db.execute(
            'UPDATE chapters SET cloud_id = ? WHERE book_id = ? AND chapter_index = ?',
            [cloudChapterId, localBookId, chapter.index]
          )
        }
      }

      // 5. 更新进度
      uploadedCount += batch.length
      const progress = Math.floor((uploadedCount / chapters.length) * 100)
      bookStore.uploadProgress = progress
    } else {
      throw new Error(res.msg || '上传失败')
    }
  }

  // 6. 更新sync_state为SYNCED
  await db.execute(
    'UPDATE books SET sync_state = 1 WHERE id = ?',
    [localBookId]
  )

  uni.showToast({ title: '上传成功', icon: 'success' })
  await bookStore.fetchBooks()
}
```

---

## 五、章节内容相关

### 4.1 批量获取章节原文

**接口路径：** `POST /api/v1/chapters/content`

**调用时机：**
- 小程序端：进入阅读器时，预加载前5章
- 小程序端：阅读过程中，检测到后续章节未缓存，批量加载
- App端：降级模式（sync_state=2）时，按需从云端拉取

**请求参数：**
```json
{
  "ids": [5001, 5002, 5003]  // 最多10个
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "chapter_id": 5001,
      "chapter_md5": "m1n2o3p4q5r6...",
      "content": "序章内容..."
    },
    {
      "chapter_id": 5002,
      "chapter_md5": "a7b8c9d0e1f2...",
      "content": "第一章内容..."
    }
  ]
}
```

**小程序端处理逻辑：**
```typescript
async function fetchBatchChapters(chapterIds: number[]) {
  // 1. 检查本地缓存（Tier 2）
  const uncachedIds: number[] = []
  const cachedContent: Map<string, string> = new Map()

  for (const id of chapterIds) {
    const cached = uni.getStorageSync(`st_content_${id}`)
    if (cached) {
      cachedContent.set(id.toString(), cached)
    } else {
      uncachedIds.push(id)
    }
  }

  // 2. 批量拉取未缓存的内容
  if (uncachedIds.length > 0) {
    const res = await request({ url: '/chapters/content', method: 'POST', data: { ids: uncachedIds } })

    if (res.code === 0) {
      // 3. 存储到缓存
      for (const item of res.data) {
        uni.setStorageSync(`st_content_${item.chapter_id}`, item.content)
        cachedContent.set(item.chapter_id.toString(), item.content)
      }
    }
  }

  // 4. 返回内容
  return chapterIds.map(id => cachedContent.get(id.toString()) || '')
}
```

---

### 4.2 批量获取精简内容（ID寻址）

**接口路径：** `POST /api/v1/chapters/trim`

**调用时机：**
- 小程序端：用户切换到精简模式时
- 小程序端：用户点击"预加载"按钮时

**请求参数：**
```json
{
  "ids": [5001, 5002],  // 最多10个
  "prompt_id": 1
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "chapter_id": 5001,
      "prompt_id": 1,
      "trimmed_content": "精简后的序章内容..."
    }
  ]
}
```

**小程序端处理逻辑：**
```typescript
async function fetchBatchTrimmed(chapterIds: number[], promptId: number) {
  // 1. 检查本地缓存（Tier 2）
  const uncachedIds: number[] = []
  const cachedContent: Map<string, string> = new Map()

  for (const id of chapterIds) {
    const cached = uni.getStorageSync(`st_trim_${id}_${promptId}`)
    if (cached) {
      cachedContent.set(id.toString(), cached)
    } else {
      uncachedIds.push(id)
    }
  }

  // 2. 批量拉取未缓存的内容
  if (uncachedIds.length > 0) {
    const res = await request({ url: '/chapters/trim', method: 'POST', data: { ids: uncachedIds, prompt_id: promptId } })

    if (res.code === 0) {
      // 3. 存储到缓存
      for (const item of res.data) {
        uni.setStorageSync(`st_trim_${item.chapter_id}_${item.prompt_id}`, item.trimmed_content)
        cachedContent.set(item.chapter_id.toString(), item.trimmed_content)
      }
    }
  }

  // 4. 返回内容
  return chapterIds.map(id => cachedContent.get(id.toString()) || '')
}
```

---

### 4.3 批量获取精简内容（MD5寻址）

**接口路径：** `POST /api/v1/contents/trim`

**调用时机：**
- App端：sync_state=1（已同步）时，用户切换章节，预加载精简内容

**请求参数：**
```json
{
  "md5s": ["a7b8c9d0e1f2...", "b3c4d5e6f7a8..."],  // 最多10个
  "prompt_id": 1
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "chapter_md5": "a7b8c9d0e1f2...",
      "prompt_id": 1,
      "trimmed_content": "精简后的第一章内容..."
    }
  ]
}
```

**App端处理逻辑：**
```typescript
async function fetchBatchTrimmedByMD5(md5s: string[], promptId: number) {
  // 1. 检查Storage缓存（Tier 2）
  const uncachedMD5s: string[] = []
  const cachedContent: Map<string, string> = new Map()

  for (const md5 of md5s) {
    const cached = uni.getStorageSync(`st_trim_${md5}_${promptId}`)
    if (cached) {
      cachedContent.set(md5, cached)
    } else {
      uncachedMD5s.push(md5)
    }
  }

  // 2. 批量拉取未缓存的内容
  if (uncachedMD5s.length > 0) {
    const res = await request({ url: '/contents/trim', method: 'POST', data: { md5s: uncachedMD5s, prompt_id: promptId } })

    if (res.code === 0) {
      // 3. 存储到Storage缓存
      for (const item of res.data) {
        uni.setStorage({
          key: `st_trim_${item.chapter_md5}_${item.prompt_id}`,
          data: item.trimmed_content
        })
        cachedContent.set(item.chapter_md5, item.trimmed_content)
      }
    }
  }

  // 4. 返回内容
  return md5s.map(md5 => cachedContent.get(md5) || '')
}
```

---

### 4.4 同步精简足迹（ID寻址）

**接口路径：** `POST /api/v1/chapters/sync-status`

**调用时机：**
- 小程序端：进入阅读器时，获取全书精简状态
- 小程序端：用户点击"刷新"按钮时

**请求参数：**
```json
{
  "book_id": 101
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "trimmed_map": {
      "5001": [1, 2],  // 章节ID -> 已精简的PromptID列表
      "5002": [1]
    }
  }
}
```

**小程序端处理逻辑：**
```typescript
async function syncTrimmedStatusByBookId(bookId: number) {
  // 1. 调用API获取精简状态
  const res = await request({ url: '/chapters/sync-status', method: 'POST', data: { book_id: bookId } })

  if (res.code === 0) {
    const { trimmed_map } = res.data

    // 2. 更新本地Store
    if (bookStore.activeBook) {
      bookStore.activeBook.chapters.forEach(chapter => {
        chapter.trimmed_prompt_ids = trimmed_map[chapter.id] || []
      })
    }
  }
}
```

---

### 4.5 同步精简足迹（MD5寻址）

**接口路径：** `POST /api/v1/contents/sync-status`

**调用时机：**
- App端：用户切换章节时，探测当前章+预加载章的精简状态
- App端：用户点击"刷新"按钮时

**请求参数：**
```json
{
  "md5s": ["a7b8c9d0e1f2...", "b3c4d5e6f7a8..."]
}
```

**返回示例：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "trimmed_map": {
      "a7b8c9d0e1f2...": [1, 2],  // MD5 -> 已精简的PromptID列表
      "b3c4d5e6f7a8...": [1]
    }
  }
}
```

**App端处理逻辑：**
```typescript
async function syncTrimmedStatusByMD5(md5s: string[]) {
  // 1. 调用API获取精简状态
  const res = await request({ url: '/contents/sync-status', method: 'POST', data: { md5s } })

  if (res.code === 0) {
    const { trimmed_map } = res.data

    // 2. 更新本地Store
    if (bookStore.activeBook) {
      bookStore.activeBook.chapters.forEach(chapter => {
        if (chapter.md5) {
          chapter.trimmed_prompt_ids = trimmed_map[chapter.md5] || []
        }
      })
    }
  }
}
```

---

## 五、AI精简相关

### 5.1 基于MD5的流式精简（App端）

**接口路径：** `GET /api/v1/trim/stream/by-md5` (WebSocket)

**调用时机：**
- App端：用户点击"精简"按钮，缓存未命中时

**WebSocket连接参数：**
- Query参数：`token={token}`
- 发送消息：
```json
{
  "content": "章节原文内容...",
  "md5": "a7b8c9d0e1f2...",
  "prompt_id": 1,
  "chapter_index": 1,
  "book_fingerprint": "f1e2d3c4b5a6..."
}
```

**流式响应：**
```json
{"c": "精简"}
{"c": "后的"}
{"c": "内容"}
// ... 最后一帧后发送CloseMessage
```

**App端处理逻辑：**
```typescript
const token = uni.getStorageSync('token')
const wsUrl = `${WS_BASE_URL}/trim/stream/by-md5?token=${token}`

const socketTask = uni.connectSocket({ url: wsUrl })

socketTask.onOpen(() => {
  // 发送精简请求
  socketTask.send({
    data: JSON.stringify({
      content: rawContent,
      md5: chapterMD5,
      prompt_id: promptId,
      chapter_index: chapterIndex,
      book_fingerprint: bookFingerprint
    })
  })
})

socketTask.onMessage((res) => {
  const data = JSON.parse(res.data as string)

  if (data.error) {
    // 错误处理
    console.error('Trim error:', data.error)
    uni.showToast({ title: '精简失败', icon: 'none' })
    socketTask.close({})
  } else if (data.c) {
    // 追加流式内容
    streamingContent.value += data.c
    currentTextLines.value = streamingContent.value.split('\n')
  }
})

socketTask.onClose(() => {
  // 精简完成
  console.log('Trim completed')

  // 1. 更新UI
  activeBook.value.activeModeId = promptId.toString()
  isMagicActive.value = true

  // 2. 保存到Storage缓存（异步）
  uni.setStorage({
    key: `st_trim_${chapterMD5}_${promptId}`,
    data: streamingContent.value
  })

  // 3. 更新章节的精简状态
  const chapter = activeBook.value.chapters[activeChapterIndex.value]
  if (!chapter.trimmed_prompt_ids.includes(promptId)) {
    chapter.trimmed_prompt_ids.push(promptId)
  }

  // 4. 关闭终端
  showTerminal.value = false
})

socketTask.onError((err) => {
  console.error('WebSocket error:', err)
  uni.showToast({ title: '连接失败', icon: 'none' })
})
```

---

### 5.2 基于ID的流式精简（小程序端）

**接口路径：** `GET /api/v1/trim/stream/by-id` (WebSocket)

**调用时机：**
- 小程序端：用户点击"精简"按钮，缓存未命中时

**WebSocket连接参数：**
- Query参数：`token={token}`
- 发送消息：
```json
{
  "book_id": 101,
  "chapter_id": 5002,
  "prompt_id": 1
}
```

**流式响应：**
```json
{"c": "精简"}
{"c": "后的"}
{"c": "内容"}
// ... 最后一帧后发送CloseMessage
```

**小程序端处理逻辑：**
```typescript
const token = uni.getStorageSync('token')
const wsUrl = `${WS_BASE_URL}/trim/stream/by-id?token=${token}`

const socketTask = uni.connectSocket({ url: wsUrl })

socketTask.onOpen(() => {
  // 发送精简请求
  socketTask.send({
    data: JSON.stringify({
      book_id: bookId,
      chapter_id: chapterId,
      prompt_id: promptId
    })
  })
})

socketTask.onMessage((res) => {
  const data = JSON.parse(res.data as string)

  if (data.error) {
    console.error('Trim error:', data.error)
    uni.showToast({ title: '精简失败', icon: 'none' })
    socketTask.close({})
  } else if (data.c) {
    // 追加流式内容
    streamingContent.value += data.c
    currentTextLines.value = streamingContent.value.split('\n')
  }
})

socketTask.onClose(() => {
  // 精简完成
  console.log('Trim completed')

  // 1. 更新UI
  activeBook.value.activeModeId = promptId.toString()
  isMagicActive.value = true

  // 2. 保存到缓存（Tier 2）
  uni.setStorage({
    key: `st_trim_${chapterId}_${promptId}`,
    data: streamingContent.value
  })

  // 3. 更新章节的精简状态
  const chapter = activeBook.value.chapters[activeChapterIndex.value]
  if (!chapter.trimmed_prompt_ids.includes(promptId)) {
    chapter.trimmed_prompt_ids.push(promptId)
  }

  // 4. 关闭终端
  showTerminal.value = false
})

socketTask.onError((err) => {
  console.error('WebSocket error:', err)
  uni.showToast({ title: '连接失败', icon: 'none' })
})
```

---

## 六、附录

### 6.1 前端决策矩阵

| 场景 | 原文获取 | 精简判定 | 精简拉取 | 兜底生成 |
| :--- | :--- | :--- | :--- | :--- |
| **App已同步** | SQLite | `contents/sync-status` | `contents/trim` | `stream/by-md5` |
| **小程序** | `chapters/content` | `books/:id` | `chapters/trim` | `stream/by-id` |
| **App降级** | `chapters/content` | `chapters/sync-status` | `chapters/trim` | `stream/by-id` |

### 6.2 App降级模式完整流程

**降级模式触发条件：**
- App重装或切换设备
- 用户登录后，发现本地book_md5与云端不一致
- 创建sync_state=2的本地书籍记录

**完整处理流程：**

**步骤1：书架同步（在书架页加载时执行）**
```typescript
// 获取云端书架并同步到本地（详见3.1接口）
// 创建sync_state=2的本地记录（如果本地不存在）
```

**步骤2：用户点击某本书进入阅读器**
```typescript
// 1. 检查本地书籍记录
const localBook = await db.select<any>(
  'SELECT * FROM books WHERE id = ?',
  [bookId]
)

if (localBook.length > 0 && localBook[0].sync_state === 2) {
  // 2. 降级模式：从云端获取书籍详情
  const res = await request({
    url: `/books/${localBook[0].cloud_id}`,  // 使用cloud_id获取
    method: 'GET'
  })

  if (res.code === 0) {
    const { book, chapters, trimmed_map, reading_history } = res.data

    // 3. 将chapters保存到本地（仅目录索引，不包含内容）
    await db.transaction(async () => {
      for (const cloudChapter of chapters) {
        await db.execute(
          'INSERT OR REPLACE INTO chapters (book_id, cloud_id, chapter_index, title, md5) VALUES (?, ?, ?, ?, ?)',
          [localBook[0].id, cloudChapter.id, cloudChapter.index, cloudChapter.title, cloudChapter.chapter_md5]
        )
      }
    })

    // 4. 设置为activeBook
    bookStore.activeBook = {
      id: localBook[0].id,  // 本地ID
      title: book.title,
      total_chapters: book.total_chapters,
      fingerprint: book.fingerprint,
      chapters: chapters.map(ch => ({
        id: localBook[0].id, // 保持本地ID
        book_id: localBook[0].id, // 本地book_id
        cloud_id: ch.id, // 云端章节ID
        index: ch.index,
        title: ch.title,
        chapter_md5: ch.chapter_md5,
        trimmed_prompt_ids: trimmed_map[ch.id] || [],
        isLoaded: false,
        modes: { original: [] } // 初始为空数组，需要按需加载
      })),
      activeChapterIndex: reading_history ? chapters.findIndex(c => c.id === reading_history.last_chapter_id) : 0,
      activeModeId: reading_history?.last_prompt_id?.toString() || 'original'
    }

    // 5. 预加载第一张（原文）
    await bookStore.fetchChapter(localBook[0].id, bookStore.activeBook.chapters[0].id)
  }
}
```

**步骤3：阅读过程中的内容获取**
```typescript
// 判断使用哪种数据源
const useLocalData = localBook[0].sync_state !== 2

if (useLocalData) {
  // sync_state=0或1：使用本地SQLite
  const content = await db.select<any>(
    'SELECT c.raw_content FROM contents c JOIN chapters ch ON c.chapter_md5 = ch.md5 WHERE ch.id = ?',
    [chapterId]
  )
  // 使用SQLite数据
} else {
  // sync_state=2：降级模式，调用云端API
  const res = await request({
    url: '/chapters/content',
    method: 'POST',
    data: { ids: [chapterId] }
  })

  if (res.code === 0) {
    const content = res.data.find(item => item.chapter_id === chapterId)?.content
    // 可选：存储到Tier 2缓存
    uni.setStorage({
      key: `st_content_${chapterId}`,
      data: content
    })
  }
}
```

**步骤4：同步阅读进度**
```typescript
// sync_state=2时才上报云端
if (localBook[0].sync_state === 2) {
  await request({
    url: `/books/${localBook[0].cloud_id}/progress`,
    method: 'POST',
    data: { chapter_id, prompt_id: promptId }
  })
}

// 本地始终记录
await db.execute(
  'INSERT OR REPLACE INTO reading_history (book_id, last_chapter_id, last_prompt_id, updated_at) VALUES (?, ?, ?, ?)',
  [localBook[0].id, chapterId, promptId || 0, Date.now()]
)
)
```

**步骤5：用户点击"下载全书"**
```typescript
// 将降级模式转为同步模式
async function downloadFullBook(bookId: number) {
  const book = await db.select<any>('SELECT * FROM books WHERE id = ?', [bookId])
  if (book[0].sync_state !== 2) return

  const chapters = await db.select<any>(
    'SELECT * FROM chapters WHERE book_id = ? ORDER BY chapter_index',
    [bookId]
  )

  const BATCH_SIZE = 10
  let downloadedCount = 0

  // 批量下载原文并保存到SQLite
  for (let i = 0; i < chapters.length; i += BATCH_SIZE) {
    const batch = chapters.slice(i, i + BATCH_SIZE)
    const cloudIds = batch
      .map(ch => ch.cloud_id)
      .filter((id): id is number => id > 0)

    if (cloudIds.length === 0) break

    const res = await request({
      url: '/chapters/content',
      method: 'POST',
      data: { ids: cloudIds }
    })

    if (res.code === 0) {
      for (const item of res.data) {
        // 存储到contents表
        await db.execute(
          'INSERT OR REPLACE INTO contents (chapter_md5, raw_content) VALUES (?, ?)',
          [item.chapter_md5, item.content]
        )
      }
      downloadedCount++
    }

    // 更新进度
    const progress = Math.floor((downloadedCount / chapters.length) * 100)
    // 显示进度UI
  }

  // 同步精简内容（可选，根据需求实现）

  // 更新sync_state为1
  await db.execute(
    'UPDATE books SET sync_state = 1, synced_count = ? WHERE id = ?',
    [downloadedCount, bookId]
  )

  uni.showToast({ title: '下载完成', icon: 'success' })
  await bookStore.fetchBooks()
}
```

**关键设计点：**
1. sync_state=2时，chapters表保存云端ID，但contents表为空
2. 阅读时动态从云端拉取，可选择缓存到Storage或下载全文
3. 用户可选择"下载全书"，将sync_state从2转为1
4. 进度上报只在sync_state=2时才调用云端API
5. 内容获取和精简获取都与小程序端保持一致

### 6.2 数据缓存策略

**App端两级缓存：**
1. **Tier 1 (Memory)**: Pinia Store - 当前阅读章节及前后各2章
2. **Tier 2 (SQLite)**: `contents` 表 - 所有原文内容
3. **Tier 3 (Storage)**: `uni.setStorage` - 临时缓存精简内容（LRU淘汰）

**小程序端两级缓存：**
1. **Tier 1 (Memory)**: Pinia Store - 当前阅读章节及前后各2章
2. **Tier 2 (Storage)**: `uni.setStorage` - 所有精简和原文内容

### 6.3 批量接口限制

| 接口 | 最大请求数 | 超限处理 |
|------|----------|---------|
| `POST /chapters/content` | 10 | 返回400错误 |
| `POST /chapters/trim` | 10 | 返回400错误 |
| `POST /contents/trim` | 10 | 返回400错误 |
| `POST /contents/sync-status` | 20 | 返回400错误 |

---

## 七、设计原则总结

### 7.1 App端数据访问策略（核心原则：本地优先）

**sync_state 状态定义：**
- `0` (LOCAL_ONLY)：仅本地，未上传
- `1` (SYNCED)：已同步，本地有完整数据
- `2` (CLOUD_ONLY)：仅云端（降级模式），本地无完整数据

**数据源选择规则：**

| sync_state | 原文来源 | 精简判定 | 精简获取 | 精简生成 | 进度上报 |
|----------|---------|----------|-----------|-----------|-----------|
| **0** (LOCAL_ONLY) | SQLite | 本地SQLite查询 | - | - | 本地SQLite |
| **1** (SYNCED) | SQLite | `contents/sync-status` | `contents/trim` | `stream/by-md5` | 本地SQLite |
| **2** (CLOUD_ONLY) | `chapters/content` | `chapters/sync-status` | `chapters/trim` | `stream/by-id` | 直接上报云端 |

### 7.2 App降级模式（sync_state=2）设计要点

**触发条件：**
- App重装或切换设备
- 用户切换账号
- 本地book_md5与云端不一致

**处理流程：**
1. 书架同步时，发现云端有而本地无 → 创建`sync_state=2`的本地记录
2. 用户点击进入阅读器 → 检测到`sync_state=2`：
   - 调用云端API获取书籍详情（含章节目录，不含内容）
   - 保存章节索引到本地chapters表（仅目录，不含内容）
3. 用户阅读时需要内容 → 动态从云端拉取（可缓存到Storage）
4. 用户可选择"下载全书" → 批量下载全文到SQLite，`sync_state`转为1`

**关键设计点：**
- `sync_state=2`时，chapters表有cloud_id和章节索引，但contents表为空
- `sync_state=2`时，所有云端数据调用（进度、原文、精简）与小程序端完全一致
- `sync_state=0或1`时，完全使用本地数据，不上报云端

---

**文档版本：** v1.1
**最后更新：** 2026-01-10
**主要变更：** 
- 修正App端获取书籍详情的调用时机（仅sync_state=2时调用云端API）
- 补充App端降级模式的完整处理流程
- 更新前端决策矩阵（区分sync_state的不同数据访问策略）
