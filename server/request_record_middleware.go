package server

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// 写入响应体
func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestRecordMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 保存请求体
		requestBodyByteSlice, _ := io.ReadAll(c.Request.Body)
		requestBodyByteBuffer := bytes.NewBuffer(requestBodyByteSlice)
		c.Request.Body = io.NopCloser(requestBodyByteBuffer)

		// 保存响应体
		bodyBuffer := &bytes.Buffer{}
		w := &CustomResponseWriter{
			ResponseWriter: c.Writer,
			body:           bodyBuffer,
		}
		c.Writer = w

		c.Next()

		// 保存请求记录
		cost := time.Since(start)
		u, _ := uuid.NewV6()
		request_id := u.String()

		requestQueryStringSlice := strings.Split(c.Request.URL.String(), "?")
		var requestQueryString string
		if len(requestQueryStringSlice) > 1 {
			requestQueryString = requestQueryStringSlice[len(requestQueryStringSlice)-1]
		}

		requestModel := RequestRecordModel{
			RequestID:          request_id,
			UserName:           c.Request.Header.Get("username"),
			RequestURL:         c.Request.URL.String(),
			ClientIP:           c.ClientIP(),
			RequestMethod:      c.Request.Method,
			RequestBody:        string(requestBodyByteSlice),
			RequestQueryString: requestQueryString,
			ResponseStatus:     strconv.Itoa(w.Status()),
			ResponseBody:       w.body.String(),
			Latency:            cost.String(),
		}
		requestModel.Create()
	}
}
