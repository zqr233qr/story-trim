package domain

// Chapter 代表小说的一个章节
type Chapter struct {
	Index   int    `json:"index"`   // 章节序号
	Title   string `json:"title"`   // 章节标题
	Content string `json:"content"` // 章节原始内容
}

// ProcessedChapter 代表处理后的章节
type ProcessedChapter struct {
	Chapter
	Summary        string `json:"summary"`          // 章节摘要
	TrimmedContent string `json:"trimmed_content"`  // 缩减后的内容
}
