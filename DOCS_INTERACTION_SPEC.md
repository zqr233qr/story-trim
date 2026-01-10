# StoryTrim 前后端交互逻辑与接口调用规范 (v1.4)

## 1. 基础元数据初始化 (Metadata)

### 1.1 获取精简模式列表
- **接口路径**: `GET /api/v1/common/prompts`
- **调用时机**: App/小程序启动初始化，或进入设置页时。
- **业务目的**: 获悉云端支持的 AI 处理预设。这是所有 Magic 模式交互的基石。
- **调用优先级**: **P0**
- **输出参数**:
  ```json
  {
    "code": 0,
    "data": [
      { "id": 1, "name": "极致精简", "description": "保留主线，剔除细节", "is_default": true }
    ]
  }
  ```

---

## 2. 书籍资源发现 (Resource Discovery)

### 2.1 获取用户云端书架
- **接口路径**: `GET /api/v1/books`
- **调用时机**: 书架首页加载/下拉刷新。
- **业务目的**: 获取用户已同步的资产。App 端以此判断哪些书需要“资产恢复”，小程序端则以此作为阅读入口。
- **调用优先级**: **P0**
- **输出参数**:
  ```json
  {
    "code": 0,
    "data": [
      { "id": 101, "book_md5": "...", "title": "斗破苍穹", "total_chapters": 1000 }
    ]
  }
  ```

### 2.2 获取书籍目录详情 (核心决策点)
- **接口路径**: `GET /api/v1/books/:id`
- **调用时机**: 用户在书架点击某本已上云的书籍。
- **业务目的**: 建立阅读器上下文。下发全书目录、用户已精简的足迹图谱及阅读断点。
- **输出参数**:
  ```json
  {
    "code": 0,
    "data": {
      "book": { "id": 101, "fingerprint": "..." },
      "chapters": [{ "id": 5001, "index": 0, "title": "序章" }],
      "trimmed_map": { "5001": [1, 2] }, // 关键：告知前端哪些章已 Ready，优先走静态拉取
      "reading_history": { "last_chapter_id": 5001, "last_prompt_id": 1 }
    }
  }
  ```

---

## 3. 内容获取与预加载 (Content & Preload)

### 3.1 批量获取章节原文 (预加载)
- **接口路径**: `POST /api/v1/chapters/content`
- **调用时机**: 1. 小程序/降级 App 进入阅读页加载首章；2. 阅读过程中检测到后续 3-5 章未在 Tier 2 缓存时。
- **业务目的**: 提升阅读连贯性，消除翻页白屏。
- **参数传递**: `{ "ids": [5001, 5002, 5003] }` (上限 10)。
- **输出参数**:
  ```json
  {
    "code": 0,
    "data": [
      { "chapter_id": 5001, "chapter_md5": "...", "content": "..." }
    ]
  }
  ```

### 3.2 批量获取已精简内容 (ID 寻址)
- **接口路径**: `POST /api/v1/chapters/trim`
- **调用时机**: 小程序/降级 App 切换到 Magic 模式时。根据 `GetBookDetail` 返回的 `trimmed_map` 判定，若云端已有且缓存无，则调用。
- **业务目的**: 静态下载已生成的精简文，避免不必要的 WebSocket 流式开销。
- **参数传递**: `{ "ids": [5001], "prompt_id": 1 }`。
- **输出参数**:
  ```json
  {
    "code": 0,
    "data": [
      { "chapter_id": 5001, "prompt_id": 1, "trimmed_content": "..." }
    ]
  }
  ```

### 3.3 批量获取已精简内容 (MD5 寻址)
- **接口路径**: `POST /api/v1/contents/trim`
- **调用时机**: App 端（SYNCED 状态）切章时。若探测到云端已有该 MD5 的精简文，则异步拉取并存入本地 SQLite。
- **业务目的**: 资产跨书找回后的物理下载。
- **参数传递**: `{ "md5s": ["md5_abc"], "prompt_id": 1 }`。

---

## 4. 资产对齐与足迹同步 (Sync)

### 4.1 按章探测精简足迹 (MD5 寻址)
- **接口路径**: `POST /api/v1/contents/sync-status`
- **时机**: App 阅读器切章时（探测当前章 + 预加载章）。
- **目的**: 在不查询全书 MD5 的前提下，“即时发现”该内容是否曾被自己或他人精简过。
- **输出参数**: `{ "trimmed_map": { "md5_xxx": [1, 2] } }`。

### 4.2 全书足迹刷新 (ID 寻址)
- **接口路径**: `POST /api/v1/chapters/sync-status`
- **时机**: 用户停留在目录页点击刷新，或全书处理任务进行中。
- **目的**: 获取基于云端 ID 的最新精简进度。
- **参数传递**: `{ "book_id": 101 }`。

---

## 5. AI 内容生产 (Stream & Task)

### 5.1 实时流式精简
- **WebSocket 路径**: 
    - `/api/v1/trim/stream/by-md5` (App 离线优先)
    - `/api/v1/trim/stream/by-id` (云端模式)
- **时机**: 缓存彻底未命中，必须消耗 Token 生成时。
- **业务说明**: 结果会自动触发后端 Job 队列生成摘要。

### 5.2 全书处理任务
- **接口路径**: `POST /api/v1/tasks/full-trim`
- **时机**: 用户希望一次性处理全书。
- **目的**: 异步流水线作业（精简 -> 摘要 -> 百科）。

---

## 6. 进度管理 (Persistence)

### 6.1 阅读进度上报
- **接口路径**: `POST /api/v1/books/:id/progress`
- **时机**: 阅读器退出或切章。
- **逻辑**: 本地 SQLite 实时记录，云端异步镜像。

---

## 7. 前端决策矩阵

| 场景 | 原文获取 | 精简判定 | 精简拉取 | 兜底生成 |
| :--- | :--- | :--- | :--- | :--- |
| **App已同步 (sync_state=1)** | SQLite | `chapters/sync-status` | `chapters/trim` | `stream/by-id` |
| **小程序** | `chapters/content` | `books/:id` (一次性) | `chapters/trim` | `stream/by-id` |
| **App降级** | `chapters/content` | `chapters/sync-status` | `chapters/trim` | `stream/by-id` |

**说明：**
- App已同步（sync_state=1）：原文使用SQLite，精简使用Storage缓存，云端为兜底
- App仅本地（sync_state=0）：原文使用SQLite，精简暂不支持（离线）
- App降级（sync_state=2）：与小程序端一致，强制云端
- 小程序：强制云端

---

## 8. 离线处理逻辑

### 8.1 网络异常检测
- App端：在切换精简模式前，检测网络状态
- 小程序端：在发起API请求前，检测网络状态

### 8.2 离线场景处理

**App端（sync_state=0或1）：**
- **原文阅读**：直接使用SQLite数据，无需网络
- **精简模式**：
  - 检查Storage缓存
  - 缓存命中：直接使用缓存内容
  - 缓存未命中：提示"当前精简模式需联网，已切换至原文"，自动切换到原文

**App端（sync_state=2，降级模式）：**
- **原文阅读**：
  - 检查Storage缓存
  - 缓存命中：直接使用缓存内容
  - 缓存未命中：提示"网络不可用，请检查网络连接"
- **精简模式**：
  - 检查Storage缓存
  - 缓存命中：直接使用缓存内容
  - 缓存未命中：提示"当前精简模式需联网，已切换至原文"，自动切换到原文

**小程序端：**
- **原文阅读**：
  - 检查Storage缓存
  - 缓存命中：直接使用缓存内容
  - 缓存未命中：提示"网络不可用，请检查网络连接"
- **精简模式**：
  - 检查Storage缓存
  - 缓存命中：直接使用缓存内容
  - 缓存未命中：提示"当前精简模式需联网，已切换至原文"，自动切换到原文

### 8.3 Storage缓存策略
- **Key命名**：`st_content_{cloud_id}` 或 `st_trim_{md5}_{prompt_id}`
- **LRU淘汰**：容量不足时，淘汰最早的内容
- **缓存时效**：无固定过期时间，仅按容量淘汰
