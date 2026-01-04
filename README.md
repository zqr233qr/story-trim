# StoryTrim

StoryTrim 是一个基于 AI 的小说阅读辅助工具，旨在通过大语言模型 (LLM) 智能精简长篇网络小说中的冗余描述，提供高信噪比的阅读体验。

## ✨ 特性

- **智能分章**：自动识别小说 TXT 格式并分章。
- **AI 精简**：保留核心剧情与对话，去除注水内容。
- **流式响应**：打字机效果实时显示处理进度。
- **双栏/全本导出**：支持对比阅读及导出“脱水版”全本。
- **混合模式**：支持 Web 界面操作。

## 🛠️ 技术栈

- **后端**: Go (Gin, GORM, Cobra)
- **前端**: Vue 3, TypeScript, Tailwind CSS
- **AI**: 兼容 OpenAI 接口 (DeepSeek, ChatGPT 等)

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Node.js 18+

### 启动开发环境

1. 复制配置示例并填入 API Key:
   ```bash
   cp config.example.yaml config.yaml
   # 编辑 config.yaml 填入你的 llm.api_key
   ```

2. 一键启动前后端:
   ```bash
   chmod +x start_dev.sh
   ./start_dev.sh
   ```

访问: http://localhost:5173

## 📦 目录结构

- `cmd/`: 应用程序入口
- `internal/`: 核心业务逻辑
- `web/`: 前端 Vue 项目
- `pkg/`: 通用工具库

## 📄 License

MIT
