package frontend

import (
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/gin-gonic/gin"
)

func (f *Frontend) renderErr(c *gin.Context, code int, msg, back, backMsg string) {
	render.RenderHTML(c, code, "error.tmpl", gin.H{
		"Message":     msg,
		"Back":        back,
		"BackMessage": backMsg,
	})
}

func (f *Frontend) renderModeErr(c *gin.Context, id, msg string) {
	f.renderErr(c, http.StatusInternalServerError, msg, "/detail/"+id, "Go back to Summary")
}
