package parser

import (
	"math"
	"regexp"
	"strings"
)

// Rule 定义解析规则
type Rule struct {
	Name    string // 规则名称
	Pattern string // 正则表达式 (Go RE2)
	Weight  int    // 基础权重
}

// ChapterIndex 章节索引信息
type ChapterIndex struct {
	Index int    // 章节序号
	Title string // 章节标题
	Start int    // 起始位置 (byte offset)
	End   int    // 结束位置
	Len   int    // 内容长度
}

// Result 竞速中间结果
type matchResult struct {
	rule    Rule
	indices [][]int // RegExp FindAllIndex 的结果: [[start, end], [start, end]...]
	score   float64
}

// DefaultRules 预置规则集 (按推荐优先级排序)
// 注意：Go RE2 不支持前瞻断言 (?=)，所以我们需要用 (?m) 开启多行模式，并匹配行首 ^
var DefaultRules = []Rule{
	{
		Name:    "Strict_Chinese",
		Pattern: `(?m)^第[0-9零一二三四五六七八九十百千万]+[章回节][ \t\f].*`, // 必须有空格
		Weight:  100,
	},
	{
		Name:    "Normal_Chinese",
		Pattern: `(?m)^第[0-9零一二三四五六七八九十百千万]+[章回节].*`, // 允许无空格
		Weight:  90,
	},
	{
		Name:    "Strict_English",
		Pattern: `(?m)^Chapter\s+\d+.*`,
		Weight:  80,
	},
	{
		Name:    "Loose_Number",
		Pattern: `(?m)^\d+\.\s+.*`, // 1. Title
		Weight:  60,
	},
	{
		Name:    "Loose_Direct",
		Pattern: `(?m)^[0-9零一二三四五六七八九十百千万]+\s+.*`, // 1 Title 或 一 Title
		Weight:  40,
	},
}

// SmartParseTXT 智能解析 TXT 内容
// content: 全文内容
// customRules: 可选的自定义规则，如果为 nil 则使用 DefaultRules
// 返回: 章节列表, 选用的规则名称
func SmartParseTXT(content string, customRules []Rule) ([]ChapterIndex, string) {
	rules := customRules
	if len(rules) == 0 {
		rules = DefaultRules
	}

	var bestResult *matchResult

	// 1. 竞速阶段
	for _, rule := range rules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		// FindAllIndex 返回 [[start, end], [start, end]...]
		// 限制 max 匹配数为 -1 (无限制)
		matches := re.FindAllStringIndex(content, -1)
		if len(matches) == 0 {
			continue
		}

		// 计算评分
		score := calculateScore(len(content), matches, rule.Weight)

		if bestResult == nil || score > bestResult.score {
			bestResult = &matchResult{
				rule:    rule,
				indices: matches,
				score:   score,
			}
		}
	}

	// 2. 提取阶段
	if bestResult == nil {
		// 兜底：全文作为一章
		return []ChapterIndex{{
			Index: 0,
			Title: "全文",
			Start: 0,
			End:   len(content),
			Len:   len(content),
		}}, "Fallback"
	}

	return extractChapters(content, bestResult.indices), bestResult.rule.Name
}

// calculateScore 计算健康分
// 核心逻辑：章节长度越均匀（标准差越小），得分越高
func calculateScore(totalLen int, matches [][]int, weight int) float64 {
	count := len(matches)
	if count == 0 {
		return -1
	}

	// 1. 提取每章长度
	// 长度 = 下一章Start - 当前章Start
	// 最后一章长度 = totalLen - 最后一章Start
	lengths := make([]float64, 0, count)

	for i := 0; i < count; i++ {
		currentStart := matches[i][0]
		var nextStart int
		if i == count-1 {
			nextStart = totalLen
		} else {
			nextStart = matches[i+1][0]
		}

		l := float64(nextStart - currentStart)
		// 极短章节惩罚 (如 < 50 字节)
		if l < 50 {
			// 这可能是错误的匹配（如把文中的数字匹配成了标题）
			// 我们不直接剔除，但严重拉低均值和增大方差
		}
		lengths = append(lengths, l)
	}

	// 2. 计算平均长度
	var sum float64
	for _, l := range lengths {
		sum += l
	}
	avg := sum / float64(count)

	// 阈值过滤：如果平均每章不到 200 字，极大概率是匹配错了（比如匹配到了行号）
	if avg < 200 {
		return -10000
	}

	// 3. 计算标准差 (Standard Deviation)
	var varianceSum float64
	for _, l := range lengths {
		varianceSum += math.Pow(l-avg, 2)
	}
	stdDev := math.Sqrt(varianceSum / float64(count))

	// 4. 计算变异系数 (CV = stdDev / avg)
	// CV 越小，说明越均匀。
	// CV 典型值参考：
	// - 极好: < 0.5 (章节长度非常统一)
	// - 正常: 0.5 - 1.5
	// - 差: > 2.0
	cv := stdDev / avg

	// 5. 最终得分公式
	// Score = 基础分(权重) + 数量加成 - 离散度惩罚

	// 数量加成：每多一章 +0.1 分 (微弱鼓励匹配更多章节，防止漏配)
	// 但要有上限，防止匹配到几千个行号
	countBonus := math.Min(float64(count)*0.1, 50.0)

	// 离散度惩罚：CV * 系数。系数越大，对均匀度要求越高。
	// 假设 CV=1.0 (标准差=均值)，扣除 50 分。
	uniformityPenalty := cv * 50.0

	finalScore := float64(weight) + countBonus - uniformityPenalty

	return finalScore
}

// extractChapters 根据索引提取章节
func extractChapters(content string, matches [][]int) []ChapterIndex {
	chapters := make([]ChapterIndex, 0, len(matches))
	totalLen := len(content)

	for i := 0; i < len(matches); i++ {
		start := matches[i][0] // 标题开始
		// titleEnd := matches[i][1] // 标题结束

		var end int
		if i == len(matches)-1 {
			end = totalLen
		} else {
			end = matches[i+1][0]
		}

		// 提取标题 (去除首尾空白)
		// 注意：正则匹配的是整个标题行，matches[i] 是 [start, end]
		title := strings.TrimSpace(content[matches[i][0]:matches[i][1]])

		chapters = append(chapters, ChapterIndex{
			Index: i,
			Title: title,
			Start: start,
			End:   end,
			Len:   end - start,
		})
	}
	return chapters
}
