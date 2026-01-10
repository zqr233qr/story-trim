package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github/zqr233qr/story-trim/internal/adapter/handler/apix"
	"github/zqr233qr/story-trim/internal/core/port"
	"github/zqr233qr/story-trim/pkg/errno"
)

func AuthMiddleware(userSvc port.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// 如果 Header 没带，尝试从 Query 中获取 (用于 WebSocket)
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			apix.Error(c, 401, errno.AuthErrCode, "Authentication required")
			c.Abort()
			return
		}

		userID, err := userSvc.ValidateToken(tokenStr)
		if err != nil {
			apix.Error(c, 401, errno.AuthErrCode, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func SoftAuthMiddleware(userSvc port.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr != "" {
			userID, err := userSvc.ValidateToken(tokenStr)
			if err == nil {
				c.Set("userID", userID)
			}
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// SkipPaths 跳过记录的路径
	SkipPaths []string
}

// LoggerMiddleware 日志中间件，记录请求参数和响应
func LoggerMiddleware(config ...LoggerConfig) gin.HandlerFunc {
	cfg := LoggerConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// 跳过指定路径
		if skipPaths[c.FullPath()] {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（可重复读取）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 使用自定义 ResponseWriter 来捕获响应
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(startTime)

		// 记录请求信息
		logRequest(c, requestBody, w.body.Bytes(), latency, c.Writer.Status())
	}
}

// responseWriter 自定义 ResponseWriter 用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 拦截 Write 方法以捕获响应体
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// logRequest 记录请求和响应日志
func logRequest(c *gin.Context, requestBody, responseBody []byte, latency time.Duration, statusCode int) {
	// 构建日志结构
	logData := map[string]interface{}{
		"time":       time.Now().Format("2006-01-02 15:04:05"),
		"method":     c.Request.Method,
		"path":       c.FullPath(),
		"query":      c.Request.URL.RawQuery,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"latency":    latency.String(),
		"status":     statusCode,
	}

	// 记录请求参数
	var reqParams map[string]interface{}
	if len(requestBody) > 0 {
		// 尝试解析 JSON 请求体
		if err := json.Unmarshal(requestBody, &reqParams); err == nil {
			// 过滤大段文本内容
			filterLargeContent(reqParams)
			logData["request_body"] = reqParams
		} else {
			// 如果不是 JSON，记录原始内容
			if len(requestBody) < 500 {
				logData["request_body"] = string(requestBody)
			} else {
				logData["request_body"] = string(requestBody[:500]) + "... (truncated)"
			}
		}
	} else if c.Request.URL.RawQuery != "" {
		// GET 请求记录查询参数
		reqParams = make(map[string]interface{})
		for k, v := range c.Request.URL.Query() {
			if len(v) == 1 {
				reqParams[k] = v[0]
			} else {
				reqParams[k] = v
			}
		}
		logData["request_params"] = reqParams
	}

	// 记录路径参数
	if len(c.Params) > 0 {
		pathParams := make(map[string]string)
		for _, param := range c.Params {
			pathParams[param.Key] = param.Value
		}
		logData["path_params"] = pathParams
	}

	// 记录响应体
	if len(responseBody) > 0 {
		var respData map[string]interface{}
		if err := json.Unmarshal(responseBody, &respData); err == nil {
			// 过滤响应中的大段文本内容
			filterLargeContent(respData)
			logData["response_body"] = respData
		} else {
			if len(responseBody) < 500 {
				logData["response_body"] = string(responseBody)
			} else {
				logData["response_body"] = string(responseBody[:500]) + "... (truncated)"
			}
		}
	}

	// 记录错误信息
	if len(c.Errors) > 0 {
		errors := make([]string, 0, len(c.Errors))
		for _, err := range c.Errors {
			errors = append(errors, err.Error())
		}
		logData["errors"] = errors
	}

	// 输出日志
	logJSON, _ := json.Marshal(logData)
	log.Printf("[HTTP] %s", string(logJSON))
}

// filterLargeContent 过滤请求和响应中的大段文本内容
func filterLargeContent(data map[string]interface{}) {
	// 需要过滤的字段列表
	filterFields := []string{
		"content",         // 章节内容
		"raw_content",     // 原文内容
		"trimmed_content", // 精简内容
		"text",            // 文本内容
	}

	for key, value := range data {
		// 检查是否需要过滤
		shouldFilter := false
		for _, field := range filterFields {
			if strings.Contains(strings.ToLower(key), field) {
				shouldFilter = true
				break
			}
		}

		if shouldFilter {
			if strValue, ok := value.(string); ok {
				runes := []rune(strValue)
				if len(runes) > 10 {
					// 显示前10个字符 + 后缀
					data[key] = string(runes[:10]) + "... (filtered: " + strconv.Itoa(len(runes)) + " chars)"
				}
			}
			continue
		}

		// 递归处理嵌套结构
		switch v := value.(type) {
		case map[string]interface{}:
			filterLargeContent(v)
		case []interface{}:
			for _, item := range v {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					filterLargeContent(nestedMap)
				}
			}
		}
	}
}
