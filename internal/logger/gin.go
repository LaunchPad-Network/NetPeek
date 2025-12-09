package logger

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func InitGinLogger() {
	log := New("Gin")

	gin.DefaultWriter = io.Writer(log.WriterLevel(logrus.DebugLevel))
	gin.DefaultErrorWriter = io.Writer(log.WriterLevel(logrus.ErrorLevel))
}

func GinLoggerMiddleware(service string) gin.HandlerFunc {
	log := New(service + " Router")
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()
		clientIP, exists := c.Get("logger.clientIP")
		if !exists {
			clientIP = c.ClientIP()
		}
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		entry := log.WithFields(logrus.Fields{
			"status":    status,
			"method":    method,
			"path":      path,
			"query":     query,
			"ip":        clientIP,
			"latency":   latency,
			"userAgent": userAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request handled")
		}
	}
}
