# StoryTrim 多端数据同步与存储架构技术文档 (v1.1)

## 1. 架构核心理念
StoryTrim 采用 **“内容哈希（MD5）为物理灵魂，云端标识（CloudID）为逻辑链接”** 的设计哲学。
- **内容中心化**：通过 `ChapterMD5` 实现全网内容的物理去重与精简结果复用。
- **体验平滑化**：通过三级缓存机制，确保从 App（厚客户端）到小程序（轻客户端）以及“降级模式”下的 App 都能获得秒级的阅读体验。

---

## 2. 核心标识符定义
| 标识符 | 作用域 | 说明 |
| :--- | :--- | :--- |
| **ChapterMD5** | 全局唯一 | 章节原文归一化后的哈希。用于内容去重、精简结果（TrimResult）匹配。 |
| **BookMD5** | 本地/云端 | 书籍原始文件全文哈希。用于 App 重装后识别本地文件与云端记录的关联。 |
| **BookFingerprint** | 业务逻辑 | 通常取第一章 MD5。用于聚合书籍百科（Encyclopedia）和章节摘要（Summary）。 |
| **CloudID** | 云端唯一 | 服务端生成的自增 ID（BookID/ChapterID）。用于小程序及降级 App 的高效索引。 |

---

## 3. 三级缓存架构 (Tiered Storage)

### Tier 1: 内存缓存 (Memory Cache)
- **载体**：Uni-app Pinia Store。
- **范围**：当前阅读章节及前后相邻各 2 章。
- **目标**：响应 UI 翻页，消除渲染延迟。

### Tier 2: 持久化缓存 (Storage Cache)
- **载体**：`uni.setStorage` / `uni.getStorage`。
- **适用**：小程序、**CLOUD_ONLY 状态的 App**。
- **策略**：存储从 API 拉取的 `ChapterContent` 和 `TrimResult`。
- **Key 命名**：`st_content_{md5}` 或 `st_trim_{md5}_{prompt_id}`。

### Tier 3: 本地数据库 (Local Database)
- **载体**：App 端专属 SQLite。
- **适用**：**LOCAL_ONLY 或 SYNCED 状态的 App**。
- **目标**：全量存储书籍，支持完全离线阅读，作为上云的“源数据”。

---

## 4. App 本地数据库 (SQLite) 结构设计

### 4.1 `books` (书籍表)
用于管理书籍元数据及同步状态。
```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,                 -- 所属用户ID（用于切换账号时过滤书籍）
    cloud_id INTEGER DEFAULT 0,      -- 云端 BookID，0 表示未上云
    book_md5 TEXT UNIQUE,            -- 文件全文哈希
    fingerprint TEXT,                -- 第一章MD5
    title TEXT NOT NULL,             -- 书名
    total_chapters INTEGER DEFAULT 0,
    sync_state INTEGER DEFAULT 0,    -- 0:仅本地, 1:已同步, 2:仅云端(降级模式)
    synced_count INTEGER DEFAULT 0,   -- 已同步章节数
    created_at INTEGER               -- 创建时间（时间戳）
);
CREATE INDEX idx_book_cloud ON books(cloud_id);
CREATE INDEX idx_book_user ON books(user_id);
```

**账号切换逻辑：**
- 使用 `user_id` 过滤查询：`WHERE (sync_state=0) OR (user_id=?)`
- sync_state=0（仅本地）的书籍对所有用户可见
- sync_state=1/2的书籍仅对应用户可见

### 4.2 `chapters` (章节索引表)
用于维护书籍结构。
```sql
CREATE TABLE chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,                 -- 本地书籍ID
    cloud_id INTEGER DEFAULT 0,      -- 云端章节ID
    chapter_index INTEGER,           -- 章节下标
    title TEXT,                      -- 章节标题
    chapter_md5 TEXT,                -- 内容哈希
    words_count INTEGER,
    FOREIGN KEY(book_id) REFERENCES books(id)
);
CREATE INDEX idx_chap_md5 ON chapters(chapter_md5);
```

### 4.3 `contents` (物理内容表)
实现本地层面的内容寻址与去重。
```sql
CREATE TABLE contents (
    chapter_md5 TEXT PRIMARY KEY,    -- 主键
    raw_content TEXT                 -- 原文内容
);
```

### 4.4 `reading_history` (进度表)
```sql
CREATE TABLE reading_history (
    book_id INTEGER PRIMARY KEY,
    last_chapter_id INTEGER,
    last_prompt_id INTEGER,
    scroll_offset REAL,
    updated_at INTEGER              -- 更新时间（时间戳）
);
```

---

## 5. 关键业务流程设计

### 5.1 App 上传与“孪生”建立
1. **本地解析**：App 解析 TXT，计算 `BookMD5` 和各章 `ChapterMD5`。
2. **接口调用**：`/app/content/upload` 提交数据。
3. **ID 映射**：服务端返回 `CloudID` 列表。
4. **入库记录**：App 将 `CloudID` 更新至 `books` 和 `chapters` 表，并将 `sync_state` 改为 `SYNCED`。

### 5.2 精简状态同步 (`trimmed_status`)
- **App 请求**：发送 `[]chapter_md5`。
- **Mini/降级 App 请求**：发送 `[]cloud_chapter_id`。
- **后端策略**：直接查询 `UserProcessedChapter` (用户足迹表)。
    - *优化原因*：足迹表索引记录了“用户-内容-模式”的归属权，比扫描大体量的精简结果表快得多。
- **响应**：返回已拥有的模式 ID 列表。

### 5.3 降级模式 (App 重装/换机) 资产恢复
1. **书架找回**：App 调用 `/common/books` 拉取云端列表。
2. **初始化**：App 发现本地 SQLite 无此 `BookMD5`，创建 `sync_state = CLOUD_ONLY` 的记录。
3. **阅读行为**：
    - 用户点击章节 -> 发现本地无内容 -> 自动调用云端接口：`/chapters/content`。
    - 内容返回后存入 **Tier 2 (Storage Cache)**。
4. **行为模式**：sync_state=2时，App端行为与小程序端完全一致，强制云端。

---

## 6. 章节解析架构

### 6.1 服务端统一管理解析规则
- **接口**：`GET /api/v1/chapters/parse-rules`
- **返回**：最新的章节解析正则表达式列表
- **设计目的**：支持动态更新解析规则，无需发版

### 6.2 App端兜底机制
- **默认规则**：App内置常见的章节解析正则
- **降级策略**：服务端接口调用失败时，使用本地默认规则


### 6.3 解析流程
1. App尝试从服务端获取最新解析规则
2. 使用规则列表依次尝试解析
3. 找到第一个能识别足够章节（>10章）的规则
4. 对每个章节计算关键字段：
   - BookMD5：书籍全文哈希
   - ChapterMD5：章节内容归一化MD5
   - Fingerprint：第一章MD5（书籍指纹）
   - WordsCount：章节字数统计
5. 解析失败时，提示用户检查文件格式

---

## 7. 书籍删除架构

### 7.1 服务端软删除策略
- **Book表**：添加`DeletedAt`字段，软删除
- **Chapter表**：添加`DeletedAt`字段，级联软删除
- **保留数据**：
  - ChapterContent（全局共享内容）
  - TrimResult（全局共享精简结果）
  - ChapterSummary（全局共享摘要）
  - SharedEncyclopedia（全局共享百科）
- **删除数据**：
  - ReadingHistory（用户隐私）
  - UserProcessedChapter（用户足迹）

### 7.2 App端物理删除
- **sync_state=0/1**：直接删除本地SQLite记录
- **sync_state=2**：先调用云端API删除，再删除本地记录
- **删除范围**：
  - books、chapters、contents、reading_history

### 7.3 小程序端缓存清理
- 调用云端API删除
- 清除Tier 2缓存（Storage）
- 清除Tier 1缓存（Pinia Store）

---

## 8. Uni-app 代码实现策略：适配器模式

```typescript
// 数据源适配器接口
interface IDataProvider {
    getChapterContent(book: Book, chapter: Chapter): Promise<string>;
    getTrimmedStatus(book: Book, chapters: Chapter[]): Promise<any>;
    generateTrimmedContent(book: Book, chapter: Chapter, promptId: number): Promise<string>;
}

// App 模式适配器
class AppDataProvider implements IDataProvider {
    async getChapterContent(book, chapter) {
        // sync_state=0或1：使用本地SQLite
        if (book.sync_state !== 2) {
            return await db.query('SELECT raw_content FROM contents WHERE chapter_md5 = ?', [chapter.chapter_md5]);
        }
        // sync_state=2：调用云端API（与小程序一致）
        return await miniApi.fetchContent(chapter.cloud_id);
    }

    async getTrimmedStatus(book, chapters) {
        // sync_state=0或1：使用MD5寻址
        if (book.sync_state !== 2) {
            const md5s = chapters.map(ch => ch.chapter_md5);
            return await api.request('/contents/sync-status', { md5s });
        }
        // sync_state=2：使用ID寻址（与小程序一致）
        return await api.request('/chapters/sync-status', { book_id: book.cloud_id });
    }

    async generateTrimmedContent(book, chapter, promptId) {
        // sync_state=0或1：使用MD5流式精简
        if (book.sync_state !== 2) {
            return await api.stream('/trim/stream/by-md5', {
                content: chapter.content,
                md5: chapter.chapter_md5,
                prompt_id: promptId
            });
        }
        // sync_state=2：使用ID流式精简（与小程序一致）
        return await api.stream('/trim/stream/by-id', {
            book_id: book.cloud_id,
            chapter_id: chapter.cloud_id,
            prompt_id: promptId
        });
    }
}

// 小程序模式适配器
class MiniDataProvider implements IDataProvider {
    async getChapterContent(book, chapter) {
        // 优先 Tier 2 缓存
        let cached = uni.getStorageSync(`st_content_${chapter.id}`);
        if (cached) return cached;

        // 查云端
        let remote = await api.request('/chapters/content', { ids: [chapter.id] });
        const content = remote.data.find(item => item.chapter_id === chapter.id)?.content;

        // 存储到缓存
        uni.setStorage({key: `st_content_${chapter.id}`, data: content});
        return content;
    }

    async getTrimmedStatus(book, chapters) {
        // 使用ID寻址
        return await api.request('/chapters/sync-status', { book_id: book.id });
    }

    async generateTrimmedContent(book, chapter, promptId) {
        // 使用ID流式精简
        return await api.stream('/trim/stream/by-id', {
            book_id: book.id,
            chapter_id: chapter.id,
            prompt_id: promptId
        });
    }
}
```

---

## 9. 前端决策矩阵

| 场景 | 原文获取 | 精简判定 | 精简拉取 | 兜底生成 |
| :--- | :--- | :--- | :--- | :--- |
| **App已同步 (sync_state=1)** | SQLite | `chapters/sync-status` | `chapters/trim` | `stream/by-id` |
| **App仅本地 (sync_state=0)** | SQLite | 本地SQLite查询 | - | - |
| **小程序** | `chapters/content` | `books/:id` | `chapters/trim` | `stream/by-id` |
| **App降级 (sync_state=2)** | `chapters/content` | `chapters/sync-status` | `chapters/trim` | `stream/by-id` |

**设计原则：**
- App端本地优先（sync_state=0或1时）
- App端降级模式（sync_state=2时）行为与小程序端完全一致
- 小程序端强制云端

---

## 10. 总结
本设计通过 **"足迹(权利)全端同步"** 与 **"内容(物理)按需分级加载"**，彻底解决了多端切换、重装设备后的数据孤岛问题。
- **对用户**：阅读进度和精简资产永不丢失，哪怕切换到小程序也能立刻继续。
- **对开发者**：一套代码逻辑，通过简单的状态判断即可在 App、降级 App、小程序之间无缝切换。
- **对运维**：极度去重的内容池最大程度降低了存储和带宽成本。

---

## 版本变更记录

### v1.1 (2026-01-10)
- **新增**：章节解析架构（服务端下发+兜底机制）
- **新增**：书籍删除架构（软删除策略）
- **新增**：前端决策矩阵（统一数据访问标准）
- **优化**：适配器模式代码示例（反映最新设计）
- **优化**：App端本地数据库结构（添加user_id字段）
- **优化**：精简内容表结构（分离为独立表）
- **确认**：App降级模式精简生成路径（stream/by-id）
- **确认**：书籍唯一性策略（用户级别唯一）
- **确认**：精简结果共享策略（全局共享）

