# StoryTrim 数据库架构设计 (V3.0 - Cloud Shared Cache)

## 1. 设计目标
- **内容去重 (Deduplication)**: 全局共享原始文本，节省存储空间。
- **跨用户共享缓存 (Global Cache)**: 相同章节在相同算法下的处理结果全局复用，降低 Token 成本。
- **逻辑解耦**: 将“原文”、“摘要记忆”、“精简结果”、“业务逻辑”四者分离。
- **版本控制**: 支持 Prompt 模板迭代导致的缓存逻辑失效。

---

## 2. 核心表结构 (Core Schema)

### A. 提示词模板 (prompts)
存储精简逻辑的指令集。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 模板唯一标识 |
| `name` | string(50) | 模板名称 (e.g., "极致精简", "剧情保留") |
| `version` | string(20) | 指令版本 (e.g., "v1.0", "20240101") |
| `content` | text | 实际发送给 AI 的 System Prompt |
| `is_system` | bool | 是否系统预设 (系统级模板参与全局共享) |

### B. 原始文本池 (raw_contents) - 去重核心
存储归一化后的原始章节内容。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `content_md5` | string(32) (PK) | 归一化后的原文指纹 (去除空白、标点后的 MD5) |
| `content` | longtext | 原始章节文本 |
| `token_count` | int | 预计算的 Token 消耗量 |

### C. 剧情摘要池 (raw_summaries) - 全局记忆
存储章节的剧情记忆点，不随精简模式改变。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `content_md5` | string(32) (PK) | 关联原文指纹 |
| `summary` | text | AI 生成的 200 字剧情摘要 |
| `summary_version` | string(20) | 摘要算法版本 (系统全局唯一) |

### D. 精简结果缓存池 (trim_results) - 共享缓存
存储特定算法下的精简产物。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 唯一 ID |
| `content_md5` | string(32) | 关联原文指纹 (Index) |
| `prompt_id` | uint | 关联模板 ID (Index) |
| `prompt_version`| string(20) | 关联模板版本 (Index) |
| `trimmed_content`| longtext | AI 处理后的精简文本 |

### E. 书籍业务表 (books)
记录用户个人书架。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 书籍 ID |
| `user_id` | uint | 所属用户 ID (Index) |
| `title` | string(255) | 书名 |
| `fingerprint` | string(32) | 书籍指纹 (第一章归一化 MD5) |
| `total_chapters` | int | 总章节数 |

### F. 公共百科池 (shared_encyclopedias)
存储基于书籍指纹共享的剧情设定集。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 唯一 ID |
| `book_fingerprint` | string(32) | 关联书籍指纹 (Index) |
| `range_end` | int | 覆盖进度截止章节 (e.g., 50, 100) |
| `content` | text | 百科内容 (Markdown) |
| `version` | string(20) | 算法版本 |

### G. 章节关联表 (chapters)
建立书与内容的关联。
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 唯一 ID |
| `book_id` | uint | 关联书籍 ID (Index) |
| `index` | int | 章节序号 |
| `title` | string(255) | 章节标题 |
| `content_md5` | string(32) | 指向原文池的逻辑外键 (Index) |

---

## 3. 核心查询逻辑

### 3.1 获取精简缓存 (Trim Cache Lookup)
```sql
SELECT trimmed_content FROM trim_results 
WHERE content_md5 = ? AND prompt_id = ? AND prompt_version = ?
```

### 3.2 获取记忆摘要 (Summary Lookup)
用于在处理第 N 章时获取 N-1 章的背景。
```sql
SELECT summary FROM raw_summaries 
WHERE content_md5 = ? AND summary_version = ?
```

---

## 4. 注意事项
1. **归一化算法**: Go 与前端（如需）计算 MD5 前，必须统一执行：`去除 \n \r \t \s 和所有标点符号`。
2. **异步生成**: `raw_summaries` 应在 `trim_results` 生成后异步触发，不阻塞用户阅读流程。
3. **缓存失效**: 升级 `prompts.version` 后，新请求将不再命中 `trim_results` 的旧版本，从而触发重新生成。
