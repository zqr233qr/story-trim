package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Process bodies
		reqLog := processBody(requestBody)
		respLog := processBody(w.body.Bytes())

		// Log with structured data
		event := log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP())

		// Use Interface to allow zerolog to serialize the map/slice directly
		if reqLog != nil {
			event.Interface("req_body", reqLog)
		}
		if respLog != nil {
			event.Interface("resp_body", respLog)
		}

		event.Msg("API Request")
	}
}

// processBody tries to parse data as JSON and truncate long string fields.
// If it's not JSON, it returns the string (truncated if too long).
func processBody(data []byte) interface{} {
	if len(data) == 0 {
		return nil
	}

	var jsonVal interface{}
	if err := json.Unmarshal(data, &jsonVal); err == nil {
		return truncateJSONFields(jsonVal)
	}

	// Not JSON, return as string (with simple truncation if extremely long)
	s := string(data)
	if len([]rune(s)) > 1000 { // fallback limit for non-json raw text
		return string([]rune(s)[:1000]) + "...(raw body truncated)"
	}
	return s
}

// truncateJSONFields recursively traverses the JSON structure and truncates long strings.
func truncateJSONFields(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		runes := []rune(val)
		if len(runes) > 200 {
			return string(runes[:5]) + "..."
		}
		return val
	case map[string]interface{}:
		newMap := make(map[string]interface{}, len(val))
		for k, v := range val {
			newMap[k] = truncateJSONFields(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(val))
		for i, v := range val {
			newSlice[i] = truncateJSONFields(v)
		}
		return newSlice
	default:
		return val
	}
}