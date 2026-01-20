package server

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func PanicRecoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				stackStr := string(stack)
				log.Printf("panic error: %v", err)
				log.Printf("panic url: %v", c.Request.URL)
				log.Printf("panic stack: %v", stackStr)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"code": 500,
						"msg":  "服务器内部错误，请稍后再试",
					},
				)
			}
		}()
		c.Next()
	}
}
