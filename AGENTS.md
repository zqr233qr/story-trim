# StoryTrim AI Agent 指导文档

## 项目概述

StoryTrim 是一款 AI 辅助阅读器，支持 App 端（本地优先）和小程序端（云端优先）双端，提供基于 AI 的章节精简、摘要生成和阅读进度同步功能。

**核心功能：**

- 智能精简：多级精简模式，保留不同细节程度
- 摘要生成：自动生成章节摘要，用于跨书内容关联
- 阅读进度同步：多端实时同步阅读进度和精简足迹
- 离线支持：App端支持完全离线阅读

**MVP 范围：**

- 基于智能精简的辅助阅读
- 双端数据同步与资产恢复
- 多级精简模式（极致精简、标准精简等）

**技术栈：**

- 服务端：Go (Gin, GORM)
- 前端：Vue 3 + TypeScript + UniApp
- 数据库：SQLite（开发阶段，方便调试），生产环境计划使用MySQL
- App端：SQLite
- AI：OpenAI API
- 通信：RESTful API + WebSocket（流式精简）

---

## 核心设计原则

### 1. 内容中心化

- **ChapterContent** 通过 `ChapterMD5` 实现全局去重
- **TrimResult** 通过 `ChapterMD5 + PromptID + Level` 实现跨书、跨用户共享
- 相同内容的精简结果只生成一次，节省 AI 调用成本

### 2. 用户隔离性

- **Book** 采用 `UserID + BookMD5` 用户级别唯一
- 不同用户导入同一本书，会创建独立的 Book 记录（ID 不同）
- ReadingHistory、UserProcessedChapter 按用户隔离

### 3. 双端支持

- **App 端**：本地优先，支持离线，SQLite 存储
- **小程序端**：云端优先，依赖 API，Storage 缓存
- **降级模式**：App 重装后，sync_state=2，行为与小程端一致

### 4. 同步一致性

- ReadingHistory 支持多端实时同步
- UserProcessedChapter 记录用户精简足迹
- sync_state 控制 App 端数据访问策略

### 5. 三级缓存策略（App端）

- **Tier 1 (Memory)**: Pinia Store - 当前章节及前后各2章
- **Tier 2 (Storage)**: uni.setStorage - 临时缓存精简内容（LRU淘汰）
- **Tier 3 (SQLite)**: contents 表 - 完整本地存储

---

## 数据库操作规范

### 服务端表结构

详见 `DOCS_DATABASE_DESIGN.md`

**关键表关系：**

```
User → Book (1:N)
Book → Chapter (1:N)
Chapter → ChapterContent (N:1) - 通过 ChapterMD5
ChapterContent → TrimResult (1:N) - 通过 ChapterMD5 + PromptID + Level
```

**唯一索引规则：**

- User: `Username` UNIQUE
- Book: `UserID + BookMD5` 用户级别唯一
- Chapter: `BookID + Index` 联合唯一
- ChapterContent: `ChapterMD5` PRIMARY KEY
- TrimResult: `ChapterMD5 + PromptID + Level` 联合唯一
- UserProcessedChapter: `UserID + BookID + ChapterID + PromptID` 联合唯一

### App端表结构

详见 `DOCS_DATABASE_DESIGN.md` 第二部分

**sync_state 状态定义：**

- `0` (LOCAL_ONLY)：仅本地，未上传
- `1` (SYNCED)：已同步，本地有完整数据
- `2` (CLOUD_ONLY)：仅云端，本地无完整数据

---

## API 开发规范

详见 `DOCS_API_DESIGN.md`

### 响应格式

```json
// 成功
{"code": 0, "msg": "success", "data": {...}}

// 失败
{"code": 400, "msg": "参数错误", "data": null}
```

### 错误码

- 0: 成功
- 400: 参数错误
- 401: 未授权
- 404: 资源不存在
- 500: 服务器内部错误
- 1001: 书籍已存在
- 1002: 用户名或密码错误
- 1003: 章节内容不存在

### 批量接口限制

| 接口                       | 最大请求数 |
| -------------------------- | ---------- |
| POST /chapters/content     | 10         |
| POST /chapters/trim        | 10         |
| POST /contents/trim        | 10         |
| POST /contents/sync-status | 20         |

### 认证机制

- 公共接口：`/api/v1/common/*` 无需 Token
- 其他接口：需要从 Header 或 Query 中获取 Token
- Token 验证使用 JWT，从 Token 中解析 UserID

---

## 前端决策矩阵

| 场景                            | 原文获取           | 精简判定               | 精简拉取        | 兜底生成        |
| :------------------------------ | :----------------- | :--------------------- | :-------------- | :-------------- |
| **App 已同步 (sync_state=0/1)** | SQLite             | 本地SQLite查询         | `contents/trim` | `stream/by-md5` |
| **App 降级 (sync_state=2)**     | `chapters/content` | `chapters/sync-status` | `chapters/trim` | `stream/by-id`  |
| **小程序**                      | `chapters/content` | `books/:id`            | `chapters/trim` | `stream/by-id`  |

**关键规则：**

- App端 sync_state=0/1 时，完全使用本地数据，不上报云端
- App端 sync_state=2 时，行为与小程端完全一致
- 小程序端强制云端

---

## 代码规范

### Go 代码规范

- 使用 Gin 框架，路由分组：`/api/v1/`
- 使用 GORM 操作数据库，遵循已定义的 Model 结构
- 使用统一错误处理，参考 `pkg/errno/code.go`
- 日志使用 zerolog，参考 `pkg/logger/logger.go`
- WebSocket 流式传输格式：`{"c": "内容片段"}`，结束后发送 CloseMessage

### 注释规范

- 所有函数、结构体、复杂逻辑必须添加明确的中文注释
- 公开函数（首字母大写）必须包含功能说明、参数说明、返回值说明
- 复杂业务逻辑需要分段注释说明流程

### TypeScript 代码规范

- 使用 Pinia Store 管理状态
- API 调用统一封装，处理 Token、错误提示
- 条件编译区分 App 和小程序：`// #ifdef APP-PLUS`
- 使用 SQLite 封装层：`mobile/src/utils/sqlite.ts`

### 注释规范

- 所有函数、接口、类型定义需要添加明确的中文注释
- 复杂业务逻辑需要分段注释说明流程
- 公开组件和方法必须包含功能说明、参数说明、返回值说明

---

## 常见任务指南

### 1. 添加新的 API 端点

1. 检查 `internal/adapter/handler/http/` 下的现有路由
2. 在对应的 handler 文件中添加处理函数
3. 遵循统一响应格式，使用 `pkg/errno` 返回错误
4. 更新 `DOCS_API_DESIGN.md` 文档
5. 运行 `go mod tidy` 确保依赖完整

### 2. 修改数据库表结构

1. 先检查 `DOCS_DATABASE_DESIGN.md` 确认设计
2. 修改 `internal/adapter/repository/gorm/model.go` 中的 GORM Model
3. 更新文档
4. **不要**修改已确认的唯一索引策略
5. 当前处于开发阶段，无用户数据，无需考虑数据库迁移逻辑

### 3. 实现 App 端功能

1. 先判断 sync_state，决定数据源
2. sync_state=0/1：优先使用 SQLite
3. sync_state=2：行为与小程端一致，调用云端 API
4. 使用三级缓存：Memory → Storage → SQLite → Network
5. 更新 `DOCS_API_DESIGN.md` 中的前端逻辑示例

### 4. 实现流式精简

1. App 端使用 `/trim/stream/by-md5`（WebSocket）
2. 小程序端使用 `/trim/stream/by-id`（WebSocket）
3. 发送格式：`{"c": "内容片段"}`，结束后 Close
4. 前端收到完整内容后，保存到本地缓存/SQLite
5. 超时处理：设置合理的超时时间

### 5. 处理多账号切换

1. App 端 books 表有 `user_id` 字段
2. 登录成功后，使用 `user_id` 过滤查询：`WHERE (sync_state=0) OR (user_id=?)`
3. ReadingHistory 按 user_id 隔离

---

## 测试要求

### 单元测试

- Service 层核心逻辑需要单元测试
- Repository 层数据库操作需要测试
- 关键算法（如指纹生成、MD5计算）需要测试

### 集成测试

- API 端点需要集成测试
- WebSocket 流式传输需要测试
- 跨端同步逻辑需要测试

### 边界测试

- 批量接口超限处理
- 网络异常重试机制
- 缓存失效场景
- 并发写入冲突

---

## 性能优化

### 数据库优化

- 利用已定义的索引，避免全表扫描
- 批量操作使用事务
- ChapterContent 和 TrimResult 通过 MD5 去重，减少存储

### API 优化

- 批量接口分批处理，限制单次请求量
- 使用缓存减少重复查询
- 流式传输减少内存占用

### 前端优化

- 预加载相邻章节
- 虚拟滚动处理长列表
- 按需加载内容

---

## 安全规范

- 密码使用 bcrypt 加密
- Token 使用 JWT，设置合理过期时间
- 防止 SQL 注入（使用 GORM 参数化查询）
- 敏感信息不记录日志
- API 限流（防止滥用）

---

## 交互要求

- Thinking 思考过程用中文表述
- Reply 回答也要用中文回复

---

## 注意事项

- **当前处于设计重构阶段**：所有文档和设计方案仅供参考
- **闭环确认**：如果遇到无法确认的业务逻辑闭环，请与用户进行商讨
- **数据安全**：开发阶段无真实用户数据，可以直接修改表结构
