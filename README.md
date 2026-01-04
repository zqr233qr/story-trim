# StoryTrim

StoryTrim 是一个基于 AI 的小说阅读辅助工具，旨在通过大语言模型 (LLM) 智能精简长篇网络小说中的冗余描述，提供高信噪比的阅读体验。

![UI Preview](https://via.placeholder.com/800x400?text=StoryTrim+Preview) 
*(UI 预览图)*

## ✨ 特性

- **智能分章**：自动识别小说 TXT 格式并分章入库。
- **云端同步**：支持用户注册/登录，多端同步书架进度。
- **AI 精简**：保留核心剧情与对话，去除注水内容。
- **流式响应**：打字机效果实时显示处理进度。
- **沉浸阅读**：提供“对照模式”、“精简版”、“原版”三种视图，纸张质感护眼配色。
- **全本导出**：一键导出处理后的全本 TXT。

## 🛠️ 技术栈

- **后端**: Go 1.21+
    - Web 框架: [Gin](https://github.com/gin-gonic/gin)
    - ORM: [GORM](https://gorm.io/) (SQLite/MySQL)
    - Auth: JWT + Bcrypt
- **前端**: Vue 3 + TypeScript
    - 构建工具: Vite
    - UI 框架: Tailwind CSS (v3.4) + Headless UI
    - 状态管理: Pinia
    - 路由: Vue Router
- **AI**: 兼容 OpenAI 接口 (DeepSeek, ChatGPT 等)

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Node.js 18+

### 1. 配置环境
复制配置示例并填入 API Key:
```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml 填入你的 llm.api_key 和 auth.jwt_secret
```

### 2. 启动服务 (一键脚本)
```bash
chmod +x start_dev.sh
./start_dev.sh
```
该脚本会自动编译启动后端 (Port 8080) 和前端 (Port 5173)。

### 3. 访问
打开浏览器访问: http://localhost:5173

## 📦 目录结构

- `cmd/server`: 后端入口
- `internal/`: 核心业务逻辑 (API, Service, Domain)
- `pkg/`: 通用工具库 (Config, Database)
- `web/`: 前端 Vue 项目源码

## 📄 License

MIT