package utils

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var (
	// 匹配所有标点符号的正则
	punctuationRegex = regexp.MustCompile(`[[:punct:]]`)
)

// GetContentFingerprint 计算内容的归一化 MD5 指纹
func GetContentFingerprint(content string) string {
	// 1. 归一化处理
	normalized := NormalizeText(content)
	
	// 2. 计算 MD5
	hash := md5.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// NormalizeText 去除换行、空格、制表符及所有标点符号
func NormalizeText(text string) string {
	// 去除空白字符
	f := func(r rune) bool {
		return unicode.IsSpace(r)
	}
	text = strings.Join(strings.FieldsFunc(text, f), "")

	// 去除标点符号 (包含中文标点)
	// 更加彻底的写法是只保留汉字、数字和英文字母
	var builder strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	
	return builder.String()
}
