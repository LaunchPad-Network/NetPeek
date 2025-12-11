package frontend

import (
	"io/fs"
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/service/frontend/assets"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/spf13/viper"
)

func (f *Frontend) setup() {
	f.setupStatic()
	f.setupTemplates()
	f.setupRoutes()
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
	f.engine.GET("/ct", f.ctStage1)
	f.engine.GET("/ct2", f.ctStage2)

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
