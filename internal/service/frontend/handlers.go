package frontend

import (
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/config"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
	"github.com/LaunchPad-Network/NetPeek/internal/service/frontend/assets"
	"github.com/gin-gonic/gin"
)

func (f *Frontend) serveRobots(c *gin.Context) {
	c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func (f *Frontend) serveFavicon(c *gin.Context) {
	c.FileFromFS("static/favicon.ico", http.FS(assets.Static))
}

func (f *Frontend) handleHome(c *gin.Context) {
	mode := c.Query("mode")
	q := c.Query("q")
	if mode != "" && q != "" {
		f.handleHomeMode(c, mode, q)
		return
	}

	viewMode, err := c.Cookie("lg_view_mode")
	if err == nil {
		if viewMode == "list" {
			render.RenderHTML(c, http.StatusOK, "list.tmpl", gin.H{
				"ServersList": serverslist.GetServersList(),
			})
			return
		}
		if viewMode == "map" {
			render.RenderHTML(c, http.StatusOK, "map.tmpl", nil)
			return
		}
	}

	c.Redirect(http.StatusFound, "/ct")
}

func (f *Frontend) handleHomeMode(c *gin.Context, mode, q string) {
	switch mode {
	case "whois":
		c.Redirect(http.StatusFound, "/whois?q="+q)
	default:
		f.renderErr(c, http.StatusBadRequest, "Invalid request.", "/", "Go back to home")
	}
}

func (f *Frontend) setViewMode(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GetBrandingInfo().MapUrl == "" {
			mode = "list"
		}
		c.SetCookie("lg_view_mode", mode, 3600*24*365, "/", "", false, false)
		c.Redirect(http.StatusFound, "/")
	}
}

func (f *Frontend) ctStage1(c *gin.Context) {
	c.SetCookie("ct", "1", 3600*24*365, "/", "", false, false)
	c.Redirect(http.StatusFound, "/ct2")
}

func (f *Frontend) ctStage2(c *gin.Context) {
	ct, err := c.Cookie("ct")
	if err != nil || ct != "1" {
		f.renderErr(c, http.StatusTeapot,
			"To use this website, please enable your browser's cookies and then click the refresh link below.",
			"/", "Refresh")
		return
	}
	c.Redirect(http.StatusFound, "/list")
}
