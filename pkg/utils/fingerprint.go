package utils

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strings"
)

var puncRegexp = regexp.MustCompile(`[^a-zA-Z0-9\p{L}\p{N}]`)

// GetContentFingerprint 计算归一化后的 MD5
func GetContentFingerprint(content string) string {
	// 1. 仅保留字母和数字 (自动移除空白、换行、中英文标点、特殊符号)
	normalized := puncRegexp.ReplaceAllString(content, "")
	normalized = strings.ToLower(normalized)

	// 2. 计算 MD5
	hash := md5.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:])
}
