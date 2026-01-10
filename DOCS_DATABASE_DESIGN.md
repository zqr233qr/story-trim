# StoryTrim 数据库设计文档 (v1.0)

## 文档说明

本文档记录 StoryTrim 项目的完整数据库设计，包括：
- 服务端表结构（GORM模型）
- App端SQLite表结构
- 设计决策与待讨论问题

**最后更新：** 2026-01-10
**状态：** 设计讨论中

**注意：** 当前服务端开发阶段使用SQLite数据库，便于调试。生产环境计划使用MySQL。

---

## 一、服务端数据库设计

### 1.1 用户表 (User)

**表名：** `users`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| Username | string(255) | UNIQUE | 用户名，唯一 |
| OpenID | string(64) | INDEX | 微信小程序OpenID |
| PasswordHash | string(255) | - | 密码哈希值（bcrypt） |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- 支持两种登录方式：用户名+密码、微信小程序（OpenID）
- Username 唯一约束，防止重复注册
- PasswordHash 仅用于账号密码登录，OpenID 登录时可为空

---

### 1.2 书籍表 (Book)

**表名：** `books`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键（云端BookID） |
| UserID | uint | INDEX | 所属用户ID |
| BookMD5 | string(32) | INDEX | 书籍全文MD5（用于本地文件绑定） |
| Fingerprint | string(32) | INDEX | 书籍指纹（第一章归一化MD5） |
| Title | string(255) | - | 书名 |
| TotalChapters | int | - | 总章节数 |
| CreatedAt | time.Time | - | 创建时间 |
| DeletedAt | time.Time | INDEX | 删除时间（软删除标记，NULL表示未删除） |

**设计说明：**
- **BookMD5**：整本书的全文哈希，用于App端本地文件与云端记录的绑定
- **Fingerprint**：第一章内容的归一化MD5，用于书籍聚合（摘要、百科）
- 唯一性策略：`UserID + BookMD5` 组合唯一（用户级别唯一）
- 不同用户导入同一本书，会创建独立的Book记录（ID不同）

**使用场景：**
1. App上传时：根据 BookMD5 判断是否已存在（同用户）
2. 多端同步：App重装后，通过 BookMD5 识别同一本书
3. 内容聚合：通过 Fingerprint 聚合章节摘要和百科

---

### 1.3 章节索引表 (Chapter)

**表名：** `chapters`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键（云端ChapterID） |
| BookID | uint | UNIQUE INDEX(idx_book_index) | 所属书籍ID |
| Index | int | UNIQUE INDEX(idx_book_index) | 章节序号（从0开始） |
| Title | string(255) | - | 章节标题 |
| ChapterMD5 | string(32) | INDEX | 章节内容归一化MD5 |
| CreatedAt | time.Time | - | 创建时间 |
| DeletedAt | time.Time | INDEX | 删除时间（软删除标记，NULL表示未删除） |

**设计说明：**
- **联合唯一索引**：`BookID + Index` 确保同一本书的章节序号不重复
- **ChapterMD5**：用于内容去重和精简结果查找
- Chapter表不存储实际内容，只存储索引信息

**使用场景：**
1. 获取书籍目录：按 BookID 查询所有 Chapter，按 Index 排序
2. 内容查询：通过 ChapterMD5 去 ChapterContent 表查找实际内容
3. 精简查询：通过 ChapterMD5 去 TrimResult 表查找精简结果

---

### 1.4 章节内容表 (ChapterContent)

**表名：** `chapter_contents`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ChapterMD5 | string(32) | PRIMARY KEY | 章节内容归一化MD5（主键） |
| Content | longtext | - | 原文内容 |
| WordsCount | int | - | 字数统计 |
| TokenCount | int | - | 预估Token数 |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- **物理去重表**：全局唯一，同一章节内容只存储一次
- **主键设计**：ChapterMD5 作为主键，确保内容唯一
- 跨书共享：不同书籍的相同章节内容，指向同一个 ChapterContent 记录

**使用场景：**
1. 内容存储：App上传书籍时，所有章节内容插入此表（已存在则跳过）
2. 内容读取：通过 ChapterMD5 直接获取原文
3. 节省存储：相同内容在多本书中出现时，只存储一份

**示例：**
```
书A 第1章内容 → ChapterMD5 = "abc123"
书B 第1章内容 → ChapterMD5 = "abc123" (相同内容)
→ 两者都指向 chapter_contents 中 "abc123" 的记录
```

---

### 1.5 精简结果表 (TrimResult)

**表名：** `trim_results`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| ChapterMD5 | string(32) | UNIQUE INDEX(idx_trim_lookup) | 章节内容归一化MD5 |
| PromptID | uint | UNIQUE INDEX(idx_trim_lookup) | 精简提示词ID |
| Level | int | UNIQUE INDEX(idx_trim_lookup) | 精简级别（0-单章，1-全文） |
| TrimmedContent | longtext | - | 精简后的内容 |
| TrimWords | int | - | 精简后的字数 |
| TrimRate | decimal(5,2) | - | 精简率（精简后/原文字数*100） |
| ConsumeToken | int | - | 消耗的Token数 |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- **三元组唯一索引**：`ChapterMD5 + PromptID + Level` 确保同一内容同一模式只精简一次
- **全局共享策略**：同一章节内容，同一精简模式，所有用户共享结果
- 精简级别：
  - Level = 0：单章精简（用户点击章节精简时）
  - Level = 1：全文精简（批量任务处理时）

**使用场景：**
1. 缓存查询：用户请求精简时，先查此表，存在则直接返回
2. 成本节约：同一内容的精简结果被所有用户共享，避免重复调用LLM
3. 精简率统计：TrimRate 用于分析不同提示词的精简效果

**示例：**
```
用户A 精简 ChapterMD5="abc123" PromptID=1 → 生成 TrimResult
用户B 精简 ChapterMD5="abc123" PromptID=1 → 直接复用刚才的结果
用户C 精简 ChapterMD5="abc123" PromptID=2 → 需要重新生成（模式不同）
```

---

### 1.6 章节摘要表 (ChapterSummary)

**表名：** `chapter_summaries`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| ChapterMD5 | string(32) | UNIQUE INDEX(idx_chapter_summary) | 章节内容归一化MD5 |
| BookID | uint | INDEX | 所属书籍ID |
| BookFingerprint | string(32) | INDEX | 书籍指纹 |
| ChapterIndex | int | - | 章节索引 |
| Content | text | - | 章节摘要 |
| ConsumeToken | int | - | 消耗的Token数 |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- 摘要按章节生成，用于跨书内容关联
- BookFingerprint + ChapterIndex 组合，用于按书籍聚合摘要

**使用场景：**
1. 全书处理：批量精简任务时，为每章生成摘要
2. 百科生成：根据摘要聚合生成全书百科
3. 内容关联：用户在不同书籍中看到相似内容时，提供摘要参考

---

### 1.7 书籍百科表 (SharedEncyclopedia)

**表名：** `shared_encyclopedias`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| BookFingerprint | string(32) | UNIQUE INDEX(idx_book_enc) | 书籍指纹（第一章MD5） |
| RangeEnd | int | UNIQUE INDEX(idx_book_enc) | 百科涵盖的章节范围终点 |
| Content | text | - | 百科内容（Markdown格式） |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- 百科按书籍指纹+章节范围聚合
- RangeEnd 表示该百科涵盖到第几章
- 支持增量生成：随着章节精简完成，百科范围逐步扩大

**使用场景：**
1. 全书精简完成后，生成全书百科
2. 用户阅读时，提供全书的人物、背景等百科信息
3. 多本书籍（同一指纹）共享同一百科

---

### 1.8 用户精简足迹表 (UserProcessedChapter)

**表名：** `user_processed_chapters`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| UserID | uint | UNIQUE INDEX(idx_user_trim) | 用户ID |
| BookID | uint | UNIQUE INDEX(idx_user_trim) | 书籍ID |
| ChapterID | uint | UNIQUE INDEX(idx_user_trim) | 章节ID |
| PromptID | uint | UNIQUE INDEX(idx_user_trim) | 精简提示词ID |
| ChapterMD5 | string(32) | UNIQUE INDEX(idx_user_md5_trim) | 章节内容MD5（跨书关联用） |
| CreatedAt | time.Time | - | 创建时间 |

**设计说明：**
- **四元组唯一索引**：`UserID + BookID + ChapterID + PromptID`
- 记录用户的"精简足迹"：用户精简过哪些章节的哪些模式
- ChapterMD5 用于跨书查询：用户精简过某个MD5后，在其他书看到相同内容时，可以提示"你已精简过类似内容"

**使用场景：**
1. 足迹同步：App端和小程序端同步用户的精简记录
2. 权限管理：判断用户是否有权限访问某个精简结果
3. 跨书提示：用户在不同书籍中看到相同内容时，提示"你已精简过"

---

### 1.9 阅读进度表 (ReadingHistory)

**表名：** `reading_histories`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| UserID | uint | UNIQUE INDEX(idx_user_book) | 用户ID |
| BookID | uint | UNIQUE INDEX(idx_user_book) | 书籍ID |
| LastChapterID | uint | - | 最后阅读的章节ID |
| LastPromptID | uint | - | 最后使用的精简模式ID |
| UpdatedAt | time.Time | - | 更新时间 |

**设计说明：**
- **联合唯一索引**：`UserID + BookID`，每个用户每本书只有一条进度记录
- UpdatedAt 自动更新，用于同步判断
- 支持多端同步：App和小程序实时同步阅读进度

**使用场景：**
1. 阅读记录：记录用户读到哪一章，用的什么模式
2. 多端恢复：用户切换设备后，自动跳转到最后阅读的位置
3. 断点续读：打开书时，自动定位到上次阅读的章节

---

### 1.10 异步任务表 (Task)

**表名：** `tasks`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | string(36) | PRIMARY KEY | 任务ID（UUID） |
| UserID | uint | INDEX | 所属用户ID |
| BookID | uint | INDEX | 关联书籍ID |
| Type | string(20) | - | 任务类型（full-trim, export, 等） |
| Status | string(20) | - | 任务状态（pending, running, completed, failed） |
| Progress | int | - | 任务进度（0-100） |
| Error | text | - | 错误信息 |
| CreatedAt | time.Time | - | 创建时间 |
| UpdatedAt | time.Time | - | 更新时间 |

**设计说明：**
- 任务ID使用UUID，避免自增ID暴露业务信息
- 异步处理：耗时操作（如全书精简）使用任务队列处理
- 进度跟踪：Progress 字段用于前端展示进度条

**使用场景：**
1. 全书精简任务：用户点击"全书精简"后，创建异步任务
2. 任务监控：前端轮询或WebSocket监听任务进度
3. 任务重试：失败的任务可以重新提交

---

### 1.11 提示词配置表 (Prompt)

**表名：** `prompts`

| 字段 | 类型 | 索引 | 说明 |
|------|------|------|------|
| ID | uint | PRIMARY KEY | 自增主键 |
| Name | string(50) | - | 提示词名称（前端展示） |
| Description | string(255) | - | 提示词描述（前端展示） |
| PromptContent | text | - | 精简提示词内容 |
| SummaryPromptContent | text | - | 摘要提示词内容 |
| Type | int | - | 提示词类型（0-精简，1-摘要） |
| TargetRatioMin | float64 | - | 目标精简剩余率最小值（如 0.5） |
| TargetRatioMax | float64 | - | 目标精简剩余率最大值（如 0.6） |
| BoundaryRatioMin | float64 | - | 边界字数剩余率最小值 |
| BoundaryRatioMax | float64 | - | 边界字数剩余率最大值 |
| IsSystem | bool | - | 是否为系统提示词（不可删除） |
| IsDefault | bool | - | 是否为默认提示词 |

**设计说明：**
- 支持两种提示词：精简提示词（Type=0）和摘要提示词（Type=1）
- TargetRatio：控制精简后的字数范围（原文字数的50%-60%）
- BoundaryRatio：边界检查，防止精简过度或过少
- 系统提示词不可删除，用户可以创建自定义提示词

**使用场景：**
1. 前端展示：用户可以选择不同的精简模式（极致精简、保留剧情等）
2. LLM调用：根据选择的PromptID，使用对应的PromptContent
3. 结果质量控制：通过TargetRatio确保精简结果符合预期

---

## 二、App端SQLite数据库设计

### 2.1 书籍表 (books)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | 本地书籍ID |
| user_id | INTEGER | - | 所属用户ID（sync_state=0时为NULL，表示本地通用书籍） |
| cloud_id | INTEGER | DEFAULT 0 | 云端书籍ID（0表示未同步） |
| book_md5 | TEXT | UNIQUE | 书籍全文MD5 |
| fingerprint | TEXT | - | 书籍指纹（第一章MD5） |
| title | TEXT | NOT NULL | 书名 |
| total_chapters | INTEGER | - | 总章节数 |
| sync_state | INTEGER | DEFAULT 0 | 同步状态（0-仅本地，1-已同步，2-仅云端） |
| synced_count | INTEGER | DEFAULT 0 | 已同步章节数 |
| process_status | TEXT | - | 处理状态（new, processing, ready） |
| created_at | INTEGER | - | 创建时间（时间戳） |

**同步状态说明：**
- `0` (LOCAL_ONLY)：仅本地，未上传，user_id为NULL（本地通用书籍）
- `1` (SYNCED)：已同步，本地有完整数据，user_id绑定具体用户
- `2` (CLOUD_ONLY)：仅云端（降级模式），本地无完整数据，user_id绑定具体用户

**设计说明：**
- `user_id`：
  - sync_state=0时为NULL，表示本地通用书籍，不绑定用户
  - sync_state=1/2时绑定具体用户，仅对应用户可用
  - 支持多账号切换，查询时过滤：`WHERE (sync_state=0) OR (user_id=?)`
- `cloud_id` 用于与云端建立关联
- `sync_state` 判断数据来源，决定使用哪种数据访问策略
- `synced_count` 用于上传进度跟踪和断点续传

**同步逻辑说明：**
1. 用户登录后，从云端获取书架信息（返回：id, book_md5, title等）
2. 遍历云端书籍：
    - **本地book_md5不存在** → 插入本地book表（cloud_id=云端id, book_md5=云端md5, user_id=当前用户, sync_state=2）
      - 不获取章节信息（只要保证book表有记录即可）
    - **本地book_md5存在且sync_state=0** → UPDATE本地book表（cloud_id=云端id, sync_state=1, user_id=当前用户）
      - 调用API根据cloud_id获取章节目录及索引（不获取原文内容）
      - 批量UPDATE本地chapters表的cloud_id字段（根据book_id和chapter_index匹配）
    - **本地book_md5存在且sync_state=1或2且user_id=当前用户** → 跳过（已同步或仅云端记录）
    - **本地book_md5存在但user_id!=当前用户** → 更新user_id为当前用户

---

### 2.2 章节表 (chapters)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | 本地章节ID |
| book_id | INTEGER | FOREIGN KEY | 所属书籍ID |
| cloud_id | INTEGER | DEFAULT 0 | 云端章节ID（0表示未同步） |
| chapter_index | INTEGER | - | 章节序号 |
| title | TEXT | - | 章节标题 |
| md5 | TEXT | - | 章节内容MD5 |
| words_count | INTEGER | - | 字数统计 |

**索引：**
- `idx_chap_md5`：md5 字段索引

**设计说明：**
- `cloud_id` 用于与云端章节建立映射
- `md5` 用于去重和精简结果查找

---

### 2.3 内容去重表 (contents)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| chapter_md5 | TEXT | PRIMARY KEY | 章节内容MD5 |
| raw_content | TEXT | - | 原文内容 |

**设计说明：**
- 与服务端 ChapterContent 表对应
- 实现本地层面的内容去重

---

### 2.4 阅读进度表 (reading_history)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| book_id | INTEGER | PRIMARY KEY | 书籍ID（本地ID） |
| last_chapter_id | INTEGER | - | 最后阅读的章节ID |
| last_prompt_id | INTEGER | - | 最后使用的精简模式ID |
| scroll_offset | REAL | - | 滚动位置偏移 |
| updated_at | INTEGER | - | 更新时间（时间戳） |

**设计说明：**
- 本地阅读进度记录
- 与云端 ReadingHistory 同步

---

## 三、设计决策与待讨论问题

### 问题1：App端降级模式的CloudID处理 ✅ 已确认

**场景描述：**
- 用户在App1导入书籍，上传到云端，获得CloudBookID=101
- 用户重装App2，登录后拉取到CloudBookID=101
- App2本地SQLite需要创建一条记录，并建立与CloudID的关联

**当前情况：**
- 云端Book表的ID本身就是CloudID
- App端SQLite有`cloud_id`字段用于存储关联

**已确认方案：** **选项B**
- **API返回**：只返回`id`字段（服务端Book.ID，即CloudID）
  ```json
  {"id": 101, "title": "..."}
  ```
- **App端理解**：将`id`存储到本地`books.cloud_id`字段
- 优点：不冗余，简洁清晰
- 补充：App端books表添加了`user_id`字段，支持多账号切换

**对齐点1：App端获取书籍详情的调用时机**
根据"本地优先"的设计原则：
- **sync_state=0 (LOCAL_ONLY)**：本地有完整数据，无需调用云端API
- **sync_state=1 (SYNCED)**：本地有完整数据，无需调用云端API
- **sync_state=2 (CLOUD_ONLY)**：本地无完整数据，需要调用云端API获取

**结论：** 只有sync_state=2时，App端才需要调用API获取书籍详情。

---

### 问题2：书籍唯一性策略 ✅ 已确认

**设计问题：** 书籍的唯一性边界是什么？

**当前设计（用户级别唯一）：**
```sql
-- Book表
UserID + BookMD5 = 唯一组合
```

**含义：**
- 同一用户，同一本文件，只会有一个Book记录
- 不同用户导入同一本书，会创建独立的Book记录

**选项对比：**

| 选项 | 策略 | 优点 | 缺点 |
|------|------|------|------|
| **选项A（当前）** | 用户级别唯一 | 隔离性好，用户数据独立 | 相同BookMD5的内容可能重复存储 |
| **选项B** | 全局级别唯一 | 节省存储，内容完全复用 | 需要BookUser关联表，复杂度增加 |

**选项B实现示例：**
```sql
-- 需要新增关联表
CREATE TABLE book_users (
  book_id INTEGER,
  user_id INTEGER,
  created_at INTEGER,
  PRIMARY KEY (book_id, user_id)
);

-- Book表改为全局唯一
-- BookMD5 = 全局唯一（不依赖UserID）
```

**您倾向于哪种策略？** ✅ **已确认：选项A（当前）**

---

### 问题3：精简结果共享策略 ✅ 已确认

**设计问题：** 同一章节内容的精简结果，是否跨用户跨书共享？

**当前设计（全局共享）：**
```go
// TrimResult
ChapterMD5 + PromptID + Level = 唯一
```

**示例场景：**
- 用户A的《书A》第1章："第一章 天才少年" → ChapterMD5="abc"
- 用户B的《书B》第1章："第一章 天才少年" → ChapterMD5="abc"
- 用户A精简 → 保存到TrimResult（ChapterMD5="abc", PromptID=1）
- 用户B精简 → 复用刚才的结果

**优点：**
- 节省AI调用成本（相同内容只精简一次）
- 同一内容的精简质量一致
- 支持跨书内容复用

**潜在问题：**
- 用户隐私？不同用户看到完全相同的精简结果
- 个性化？用户可能期望不同的精简风格

**选项对比：**

| 选项 | 策略 | 优点 | 缺点 |
|------|------|------|------|
| **选项A（当前）** | 全局共享 | 节省成本，质量一致 | 缺乏个性化 |
| **选项B** | 用户级别独立 | 支持个性化 | 成本高，重复计算 |

**选项B实现示例：**
```go
// TrimResult需要添加UserID字段
UserID + ChapterMD5 + PromptID + Level = 唯一
```

**您认可"内容中心化"的设计哲学吗？还是希望为每个用户独立存储？** ✅ **已确认：选项A（当前）**

---

### 对齐点2：App端云端数据调用策略

**设计原则：**
- App端本地优先（sync_state=0或1时）
- App端降级模式（sync_state=2时）与小程序端行为一致
- 小程序端强制云端

**具体策略：**
1. **获取书籍详情**：只有sync_state=2（CLOUD_ONLY）时才调用API
2. **获取章节原文**：sync_state=2时，使用`chapters/content`（与小程序一致）
3. **获取精简内容**：sync_state=2时，使用`chapters/trim`（与小程序一致）
4. **同步精简状态**：sync_state=2时，使用`chapters/sync-status`（与小程序一致）
5. **上报阅读进度**：sync_state=2时，直接上报云端（与小程序一致）
6. **AI精简生成**：sync_state=2时，使用`stream/by-id`（与小程序一致）

**说明：**
- App端在sync_state=0或1时，完全使用本地数据，不上报云端，不调用相关API
- App端在sync_state=2时，行为与小程序端完全一致 ✅ **已确认：选项A（当前）**

---

### 问题4：表结构补充需求 ✅ 已部分确认

基于当前设计，是否有需要补充的字段或表？

**已确认添加：**
| 建议项 | 状态 | 说明 |
|--------|------|------|
| App端books表添加`user_id` | ✅ 已确认 | 支持多账号切换，用于切换账号时过滤书籍 |

**可选补充项（未确认）：**

| 建议项 | 说明 | 优先级 |
|--------|------|--------|
| Book表添加`Author`字段 | 部分同名书籍需要区分作者 | P2 |
| Book表添加`CoverURL`字段 | 支持封面展示 | P3 |
| Task表添加`Result`字段 | 存储任务结果（如导出文件路径） | P2 |
| 添加`BookTag`表 | 支持书籍标签分类 | P3 |

**您是否需要添加以上字段或表？或者有其他补充需求？**

---

## 四、数据关系图

### 4.1 服务端核心关系

```
User (用户)
  └─ 1:N ──► Book (书籍)
               │
               ├─ 1:N ──► Chapter (章节)
               │              │
               │              ├─ N:1 ──► ChapterContent (原文内容)
               │              │              └─ 1:N ──► TrimResult (精简结果)
               │              │
               │              ├─ N:1 ──► ChapterSummary (章节摘要)
               │              │
               │              └─ N:1 ──► SharedEncyclopedia (书籍百科)
               │
               ├─ 1:N ──► ReadingHistory (阅读进度)
               │
               └─ 1:N ──► Task (异步任务)

UserProcessedChapter (用户精简足迹)
  └─ 关联 ──► Chapter (章节)
  └─ 关联 ──► Book (书籍)
  └─ 关联 ──► Prompt (提示词)

Prompt (提示词)
  └─ 1:N ──► TrimResult (精简结果)
```

### 4.2 App端核心关系

```
books (书籍)
  ├─ 1:N ──► chapters (章节)
  │            │
  │            ├─ N:1 ──► contents (原文内容)
  │            │
  │
  └─ 1:1 ──► reading_history (阅读进度)
```

---

## 五、设计原则总结

### 5.1 核心设计理念

1. **内容中心化**
   - ChapterContent 通过 ChapterMD5 实现全局去重
   - TrimResult 通过 ChapterMD5+PromptID 实现跨书共享

2. **用户隔离性**
   - Book 采用 UserID + BookMD5 用户级别唯一
   - 不同用户的书籍数据完全独立

3. **双端支持**
   - App端支持本地优先，离线可用
   - 小程序端强制云端，数据依赖API

4. **同步一致性**
   - ReadingHistory 支持多端实时同步
   - UserProcessedChapter 记录用户精简足迹

### 5.2 数据访问策略

**App端（本地优先）：**
- 原文：优先从SQLite contents表读取
- 精简：优先从Storage缓存读取，不存在则调用API
- 进度：本地实时记录，异步同步云端

**小程序端（云端优先）：**
- 原文：调用API /chapters/content 获取
- 精简：调用API /chapters/trim 获取
- 进度：直接上报云端

---

## 六、已确认的设计决策

### 已确认问题1：App端降级模式的CloudID处理 ✅
- **选择方案：** 选项B（API返回id，App端理解为CloudID）
- **补充确认：** App端books表添加user_id字段，支持多账号切换

### 已确认问题3补充：章节同步逻辑 ✅
- **判断条件：** 本地book_md5存在且sync_state=0时才需要调用获取章节索引
- **更新策略：** 同步后sync_state设为1（已同步状态）
- **cloud_id更新：** 根据book_id和chapter_index批量UPDATE本地chapters表的cloud_id

## 七、已确认的设计决策

- [x] **问题1**：App端降级模式的CloudID处理方式 ✅ 已确认（选项B + 添加user_id）
- [x] **问题2**：书籍唯一性策略 ✅ 已确认（选项A：用户级别唯一）
  - 采用 `UserID + BookMD5` 组合唯一
  - 不同用户导入同一本书，会创建独立的Book记录（ID不同）
- [x] **问题3**：精简结果共享策略 ✅ 已确认（选项A：全局共享）
  - 采用 `ChapterMD5 + PromptID + Level` 三元组唯一
  - 同一章节内容的精简结果，所有用户共享
- [x] **问题4**：是否需要补充其他字段或表 ✅ 已确认
  - Book表和Chapter表添加 `DeletedAt` 字段，支持软删除
  - 其他字段根据后续功能需要动态补充
