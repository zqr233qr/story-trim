package splitter

import (
	"regexp"
	"strings"

	"github/zqr233qr/story-trim/internal/core/port"
)

type regexSplitter struct {
	re *regexp.Regexp
}

func NewRegexSplitter() port.SplitterPort {
	return &regexSplitter{
		re: regexp.MustCompile(`(?m)^第[一二三四五六七八九十百千万零\d]+章.*$`),
	}
}

func (s *regexSplitter) Split(content string) []port.SplitChapter {
	indices := s.re.FindAllStringIndex(content, -1)
	if len(indices) == 0 {
		return []port.SplitChapter{{Index: 0, Title: "正文", Content: content}}
	}

	var chapters []port.SplitChapter
	for i := 0; i < len(indices); i++ {
		end := len(content)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}

		title := strings.TrimSpace(content[indices[i][0]:indices[i][1]])
		body := strings.TrimSpace(content[indices[i][1]:end])

		chapters = append(chapters, port.SplitChapter{
			Index:   i,
			Title:   title,
			Content: body,
		})
	}
	return chapters
}
