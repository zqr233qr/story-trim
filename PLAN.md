# StoryTrim 开发计划 (Development Plan)

## ✅ 已完成阶段 (Completed Stages)

### 阶段 1-4: 核心功能与基础体验
- [x] **后端基础**: Gin 框架，流式接口，文件分章。
- [x] **前端基础**: Vue 3 + Tailwind CSS，双栏阅读器。
- [x] **流式交互**: SSE 打字机效果，自动处理队列。
- [x] **本地导出**: 纯前端 TXT 导出。

### 阶段 5: 数据持久化 (Backend Persistence)
- [x] **ORM 集成**: GORM + SQLite。
- [x] **数据模型**: `User`, `Book`, `Chapter`。
- [x] **存库逻辑**: 上传解析入库，精简结果回写数据库。

### 阶段 6: 用户系统与鉴权 (Auth System)
- [x] **认证模块**: JWT 注册/登录接口 (`/api/auth/*`)。
- [x] **权限控制**: 中间件实现 Token 验证，支持 UserID 关联。

### 阶段 7: 前端架构升级 (Frontend Refactor)
- [x] **多页面路由**: `Vue Router` 实现 Login / Dashboard / Reader 切换。
- [x] **状态管理**: `Pinia` 管理 User 和 Book 状态。
- [x] **登录交互**: 完整的注册登录流程。

### 阶段 8: 业务完善与体验突破 (Experience Polish)
- [x] **书架同步**: 实现 `GET /api/books`，Dashboard 渲染真实列表。
- [x] **UI 重构**: 移植现代化设计（弥散光背景、纸张质感、Grid 布局）。
- [x] **视图切换**: 支持 对照/精简/原版 三种阅读模式。

---

## 🚀 未来规划 (Future Roadmap)

### 📅 阶段 9: 移动端适配与 PWA (Mobile)
- [ ] **响应式优化**: 适配手机屏幕，隐藏侧边栏，优化触摸交互。
- [ ] **PWA**: 支持添加到桌面，离线缓存静态资源。

### 📅 阶段 10: 智能上下文 (AI Context / RAG)
- [ ] **剧情记忆**: 每章处理完提取关键摘要 (Summary)。
- [ ] **连贯性优化**: 将前文摘要作为 Context 传入 Prompt，防止伏笔丢失。

### 📅 阶段 11: 商业化与部署 (Deployment)
- [ ] **MySQL 迁移**: 生产环境切换数据库。
- [ ] **Docker**: 编写 Dockerfile 和 docker-compose.yml。
- [ ] **支付对接**: 实现会员订阅逻辑。