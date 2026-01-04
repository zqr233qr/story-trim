# StoryTrim 开发计划 (Development Plan)

## ✅ 阶段 4: 导出与优化 (Export & Polish) [Completed]
- [x] **导出功能**: 纯前端实现全本合并与下载
- [x] **批量处理**: 实现自动处理队列与错误中断机制
- [x] **UI 优化**: 集成夜间模式、进度条与侧边栏状态指示

## 🚀 StoryTrim 2.0: 云端阅读伴侣 (Cloud Reader)

### 📅 阶段 5: 后端重构 - 数据持久化 (Backend Persistence) [Next Step]
- [ ] **ORM 集成**: 引入 GORM + SQLite (为 MySQL 迁移做准备)
- [ ] **数据模型设计**:
    - `User`: 用户体系
    - `Book`: 书籍元数据
    - `Chapter`: 章节内容 (原文 & 精简文)
- [ ] **API 改造**:
    - 重构 `/upload`: 解析 TXT -> 存入数据库 -> 返回 `book_id`
    - 新增 `/books`: 书架管理
    - 新增 `/chapters`: 章节按需加载

### 📅 阶段 6: 用户系统与鉴权 (Auth System)
- [ ] **认证模块**: 实现 JWT 登录/注册接口
- [ ] **混合模式策略**:
    - 游客: 仅本地存储/内存处理 (限前3章)
    - 用户: 云端存储，全本同步
- [ ] **权限中间件**: 拦截 API 请求并注入 User Context

### 📅 阶段 7: 前端架构升级 (Frontend Refactor)
- [ ] **路由与状态**: 引入 Vue Router + Pinia
- [ ] **页面拆分**: 登录页 / 书架页 / 阅读器页
- [ ] **体验优化**: 实现“未登录试用”到“登录同步”的平滑过渡

### 📅 阶段 8: 核心体验突破 (Core Experience)
- [ ] **对比阅读 (Diff View)**: 只有 Pro 用户可用的高级视图
- [ ] **移动端适配**: 响应式布局优化

---
## 💡 技术栈变更
- **DB**: SQLite (Dev) / MySQL (Prod)
- **ORM**: GORM
- **Auth**: JWT
- **Frontend**: Vue Router, Pinia
