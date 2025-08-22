package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		log.Println(fmt.Sprintf("%s | %s | %d | %s", c.Request.Method, time.Since(t),
			c.Writer.Status(), c.Request.URL.Path))
	}
}
