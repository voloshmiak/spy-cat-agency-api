package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		t := time.Now()

		c.Next()

		statusCode := c.Writer.Status()

		log.Println(fmt.Sprintf("%s | %s | %d | %s | %s", c.Request.Method, time.Since(t),
			statusCode, c.Request.URL.Path, blw.body.String()))
	}
}
