package frontend

import (
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/service/frontend/assets"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func (f *Frontend) setup() {
	f.setupCookieTest()
	f.setupStatic()
	f.setupTemplates()
	f.setupRoutes()
}

func (f *Frontend) setupCookieTest() {
	f.engine.Use(func(c *gin.Context) {
		ct, err := c.Cookie("ct")
		if (err != nil || ct != "1") && c.Request.URL.Path != "/ct" && !strings.HasPrefix(c.Request.URL.Path, "/api") {
			redirect := c.Request.RequestURI
			if redirect == "/" {
				redirect = "/list"
			}

			c.Redirect(http.StatusFound, "/ct?redirect="+url.QueryEscape(redirect))
			c.Abort()
		}

		c.Next()
	})
}

func (f *Frontend) setupStatic() {
	staticFiles, err := fs.Sub(assets.Static, "static")
	if err != nil {
		log.Fatal("Failed to load static files:", err)
	}
	f.engine.StaticFS("/static", http.FS(staticFiles))
}

func (f *Frontend) setupTemplates() {
	templatesFiles, err := fs.Sub(assets.Templates, "templates")
	if err != nil {
		log.Fatal("Failed to load template files:", err)
	}

	engine := goview.New(goview.Config{
		Root:      "",
		Extension: "html",
		Master:    "base.tmpl",
	})

	engine.SetFileHandler(func(cfg goview.Config, tplFile string) (string, error) {
		data, err := fs.ReadFile(templatesFiles, tplFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})

	f.engine.HTMLRender = ginview.Wrap(engine)
}

func (f *Frontend) setupRoutes() {
	f.engine.GET("/robots.txt", f.serveRobots)
	f.engine.GET("/favicon.ico", f.serveFavicon)

	f.engine.GET("/", f.handleHome)
	f.engine.GET("/ct", f.cookieTestStage1)
	f.engine.GET("/ct2", f.cookieTestStage2)

	f.engine.GET("/list", f.setViewMode("list"))
	f.engine.GET("/map", f.setViewMode("map"))

	f.engine.GET("/detail/:id", f.handleDetail)
	f.engine.GET("/detail/:id/:protocol", f.handleProtocol)

	if viper.GetString("servers.whois") != "" {
		f.engine.GET("/whois", f.handleWhois)
	} else {
		f.engine.GET("/whois", f.handleWhoisNotSupported)
	}
}
