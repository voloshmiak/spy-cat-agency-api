package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
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
		var requestBodyBytes []byte
		if c.Request.Body != nil {
			requestBodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))

		log.Printf(
			"Request: Method=%s, Path=%s, Body=%s",
			c.Request.Method,
			c.Request.URL.Path,
			string(requestBodyBytes),
		)

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		t := time.Now()

		c.Next()

		log.Printf(
			"Response: StatusCode=%d, Latency=%v, Body=%s",
			c.Writer.Status(),
			time.Since(t),
			blw.body.String(),
		)
	}
}
