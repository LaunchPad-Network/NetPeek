package frontend

import (
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/router"
	"github.com/gin-gonic/gin"
)

var log = logger.New("Frontend")

type Frontend struct {
	engine *gin.Engine
}

func New() *Frontend {
	f := &Frontend{
		engine: router.SetupRouter(),
	}
	f.setup()
	return f
}

func (f *Frontend) Engine() *gin.Engine {
	return f.engine
}

func SetupRouter() *gin.Engine {
	f := New()
	return f.Engine()
}
