package router

import (
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/config"
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var ginDebugLog = logger.New("GinDebug")

func init() {
	if config.IsDevelopment() {
		gin.SetMode(gin.DebugMode)

		gin.DebugPrintFunc = func(format string, values ...interface{}) {
			ginDebugLog.Debugf(strings.TrimRight(format, "\n"), values...)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.RedirectTrailingSlash = false
	r.HandleMethodNotAllowed = true
	r.RemoteIPHeaders = viper.GetStringSlice("net.remote_ip_headers")

	r.Use(
		middleware.Recover(),
		logger.GinLoggerMiddleware("Http"),
	)

	setupDebugRoute(&r.RouterGroup)

	return r
}
