# StoryTrim 技术架构全景手册 (V3.0 - Intelligence & Precision)

## 1. 核心设计哲学
- **物理极致共享**：指纹去重存库。
- **业务严格隔离**：足迹决定可见性。
- **两层指令框架**：系统协议 (逻辑约束) + 业务模板 (风格控制)。
- **三态上下文感知**：区分“碎片化阅读”与“全本深度阅读”的缓存版本。

## 2. 核心数据结构升级

### 2.1 三态缓存索引 (Trim Cache Strategy)
缓存表 `trim_results` 的查询键扩展为：`content_md5` + `prompt_id` + `prompt_version` + `context_level`。
- `level 0`: 无上下文 (Stateless)。
- `level 1`: 仅携带前 N 章剧情摘要。
- `level 2`: 携带剧情摘要 + 全局百科背景。

### 2.2 步进式百科 (Stepping Encyclopedia)
- 存储于 `books.global_context`。
- 采用 **异步合并更新**：每 $K$ 章采集该区间的 Summary，调用 LLM 与旧百科合并，生成新版 Markdown 格式百科。

## 3. 业务流水线逻辑

### 3.1 增强型精简流水线 (Enhanced Trim Pipeline)
1. **上下文采集**：
   - 检查 `config.memory.context_mode`。
   - 若 mode >= 1：拉取 $N-1$ 章 `raw_summaries`。
   - 若 mode == 2：拉取 `books.global_context`。
2. **Prompt 组装 (The Assembler)**：
   - `SystemPrompt = [Config.Protocol] + [Global_Context] + [Previous_Summaries] + [Template.Content]`。
3. **缓存命中算法**：
   - 优先查找当前 `context_mode` 匹配的物理缓存。
   - 若命中：记录用户足迹，启动 **MockStreamer** 进行拟真流式输出。
   - 若未命中：调用真实 LLM，处理结束后同步更新“指纹摘要”并触发百科异步更新。

### 3.2 归一化 MD5 算法
- 处理：`去除所有 \s, \n, \r, \t 及所有中英文标点符号`。
- 目的：确保“内容去重”不受排版波动影响。

## 4. 可配置参数说明 (Configurable Params)
- `memory.encyclopedia_interval`: 控制百科更新频率，平衡成本与逻辑准确度。
- `memory.summary_limit`: 控制精简时的“回望”跨度。
- `protocol.base_instruction`: 系统底层的逻辑契约，保障 AI 的“常识”边界。

---

## 5. 存储架构快照
- `raw_contents`: 原始章节文本（PK: MD5）。
- `raw_summaries`: 章节剧情记忆（PK: MD5, Version）。
- `trim_results`: 精简缓存（Index: MD5+PromptID+Version+ContextLevel）。
- `user_processed_chapters`: 用户业务可见性（UniqueIndex: User+Book+Chapter+Prompt）。
- `reading_histories`: 用户进度足迹（UniqueIndex: User+Book）。