package port

type SplitChapter struct {
	Index   int
	Title   string
	Content string
}

type SplitterPort interface {
	Split(content string) []SplitChapter
}
