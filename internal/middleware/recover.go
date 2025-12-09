package middleware

import (
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/logger"

	"github.com/gin-gonic/gin"
)

var recoverLog = logger.New("Recover Middleware")

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				recoverLog.Errorf("panic: %+v", err)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
