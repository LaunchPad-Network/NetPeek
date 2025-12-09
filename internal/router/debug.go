package router

import (
	"net/http"
	"runtime"

	"github.com/LaunchPad-Network/NetPeek/internal/config"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func gc(c *gin.Context) {
	runtime.GC()
	c.AbortWithStatus(http.StatusNoContent)
}

func setupDebugRoute(r *gin.RouterGroup) {
	if !config.IsDevelopment() { // 务必不要在生产环境暴露
		return
	}

	rg := r.Group("/debug")
	rg.GET("/gc", gc)

	pprof.Register(r) // debug/pprof
}
