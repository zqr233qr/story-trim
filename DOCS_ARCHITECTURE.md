# StoryTrim 多端数据同步与存储架构技术文档 (v1.0)

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
    cloud_id INTEGER DEFAULT 0,      -- 云端 BookID，0 表示未上云
    book_md5 TEXT UNIQUE,            -- 文件全文哈希
    fingerprint TEXT,                -- 第一章MD5
    title TEXT NOT NULL,             -- 书名
    total_chapters INTEGER DEFAULT 0,
    sync_state INTEGER DEFAULT 0,    -- 0:仅本地, 1:已同步, 2:仅云端(降级模式)
    created_at DATETIME
);
CREATE INDEX idx_book_cloud ON books(cloud_id);
```

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
    raw_content TEXT,                -- 原文内容
    trimmed_data TEXT                -- JSON格式存储: {"mode_1": "...", "mode_2": "..."}
);
```

### 4.4 `reading_history` (进度表)
```sql
CREATE TABLE reading_history (
    book_id INTEGER PRIMARY KEY,
    last_chapter_id INTEGER,
    last_prompt_id INTEGER,
    scroll_offset REAL,
    updated_at DATETIME
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
    - 用户点击章节 -> 发现本地无内容 -> 自动调用 Mini 接口：`/mini/raw/content?chapter_id=xxx`。
    - 内容返回后存入 **Tier 2 (Storage Cache)**。
4. **全量恢复**：用户点击“下载全书” -> App 后台静默调用 Mini 接口批量拉取原文 -> 写入 Tier 3 (SQLite) -> 状态转为 `SYNCED`。

---

## 6. Uni-app 代码实现策略：适配器模式

```typescript
// 数据源适配器接口
interface IDataProvider {
    getChapterContent(book: Book, chapter: Chapter): Promise<string>;
    getTrimmedStatus(book: Book, chapters: Chapter[]): Promise<any>;
}

// App 模式适配器
class AppDataProvider implements IDataProvider {
    async getChapterContent(book, chapter) {
        if (book.sync_state === SYNCED) {
            return await db.query('SELECT raw_content FROM contents WHERE chapter_md5 = ?', [chapter.chapter_md5]);
        }
        // 如果是 CLOUD_ONLY，则动态降级到 Mini 接口
        return await miniApi.fetchContent(chapter.cloud_id);
    }
}

// 小程序模式适配器
class MiniDataProvider implements IDataProvider {
    async getChapterContent(book, chapter) {
        // 优先 Tier 2 缓存
        let cached = uni.getStorageSync(`cache_${chapter.chapter_md5}`);
        if (cached) return cached;
        
        // 查云端
        let remote = await miniApi.fetchContent(chapter.cloud_id);
        uni.setStorage({key: `cache_${chapter.chapter_md5}`, data: remote});
        return remote;
    }
}
```

---

## 7. 总结
本设计通过 **“足迹(权利)全端同步”** 与 **“内容(物理)按需分级加载”**，彻底解决了多端切换、重装设备后的数据孤岛问题。
- **对用户**：阅读进度和精简资产永不丢失，哪怕切换到小程序也能立刻继续。
- **对开发者**：一套代码逻辑，通过简单的状态判断即可在 App、降级 App、小程序之间无缝切换。
- **对运维**：极度去重的内容池最大程度降低了存储和带宽成本。
