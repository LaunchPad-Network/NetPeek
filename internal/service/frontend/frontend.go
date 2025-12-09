package frontend

import (
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/config"
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/proxyreq"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/summaryparser"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/validator"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/whois"
	"github.com/LaunchPad-Network/NetPeek/internal/router"
	"github.com/LaunchPad-Network/NetPeek/internal/service/frontend/assets"
	"github.com/robert-nix/ansihtml"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	f.engine.GET("/favicon.ico", f.redirectFavicon)

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
		f.engine.GET("/whois", func(c *gin.Context) {
			f.renderErr(c, http.StatusNotAcceptable, "Not supported", "/", "Go back to home")
		})
	}
}

func (f *Frontend) serveRobots(c *gin.Context) {
	c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func (f *Frontend) redirectFavicon(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/static/favicon.ico")
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

func (f *Frontend) handleWhois(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		render.RenderHTML(c, http.StatusOK, "whois.tmpl", gin.H{
			"Title": "WHOIS Query",
		})
		return
	}

	var res string
	probeRes, err := whois.AkaereProtocolProbe()
	if err != nil {
		res = whois.Whois(q)
	} else if !probeRes.Supported {
		res = whois.Whois(q)
	} else {
		akRes, err := whois.AkaereProtocolWhois(probeRes.Schemes[0], q)
		if err != nil {
			res = whois.Whois(q)
		} else {
			res = string(ansihtml.ConvertToHTML([]byte(akRes)))
		}
	}

	render.RenderHTML(c, http.StatusOK, "whois_res.tmpl", gin.H{
		"Title":  "WHOIS Query - " + q,
		"Query":  q,
		"Result": template.HTML(strings.TrimSpace(res)),
	})
}

func (f *Frontend) handleDetail(c *gin.Context) {
	id := c.Param("id")
	srv := serverslist.GetServerByID(id)
	if srv == nil {
		serverslist.NotifyFetchServersList()
		f.renderErr(c, http.StatusNotFound,
			"PoP Not found. Please try again later.", "/", "Go back to home")
		return
	}

	mode := c.Query("mode")
	q := c.Query("q")
	if mode != "" && q != "" {
		f.handleDetailMode(c, id, mode, q)
		return
	}

	summaryResp, err := proxyreq.BirdRequest(id, "show protocols")
	if err != nil {
		log.Errorf("Failed to fetch BGP summary for %s: %v", id, err)
		f.renderErr(c, http.StatusInternalServerError,
			"Failed to fetch BGP summary.", "/", "Go back to home")
		return
	}

	table, err := summaryparser.SummaryParse(summaryResp)
	if err != nil {
		log.Errorf("Failed to parse BGP summary for %s: %v", id, err)
		f.renderErr(c, http.StatusInternalServerError,
			"Failed to parse BGP summary.", "/", "Go back to home")
		return
	}

	render.RenderHTML(c, http.StatusOK, "summary.tmpl", gin.H{
		"Title":        id,
		"Server":       srv,
		"SummaryTable": table,
	})
}

func (f *Frontend) handleDetailMode(c *gin.Context, id, mode, q string) {
	switch mode {
	case "route":
		f.handleRoute(c, id, q)
	case "filter":
		f.handleFilter(c, id, q)
	case "traceroute":
		f.handleTraceroute(c, id, q)
	default:
		f.renderModeErr(c, id, "Invalid request.")
	}
}

func (f *Frontend) handleRoute(c *gin.Context, id, q string) {
	isV4, isV6 := validator.IsIP(q)
	isV4CIDR, isV6CIDR := validator.IsCIDR(q)
	if !(isV4 || isV6 || isV4CIDR || isV6CIDR) {
		f.renderModeErr(c, id, "Invalid IP address or CIDR notation.")
		return
	}

	cmd := "show route for " + q + " all"
	resp, err := proxyreq.BirdRequest(id, cmd)
	if err != nil {
		log.Errorf("Failed to fetch route for %s (%s): %v", id, q, err)
		f.renderModeErr(c, id, "Failed to fetch information.")
		return
	}
	if strings.Contains(resp, "syntax error, unexpected CF_SYM_UNDEFINED") {
		f.renderModeErr(c, id, "Invalid parameter. Please try again later.")
		return
	}

	f.renderBird(c, id, "show route for "+q, cmd, resp)
}

func (f *Frontend) handleFilter(c *gin.Context, id, q string) {
	if !validator.IsValidProtocol(q) {
		f.renderModeErr(c, id, "Invalid protocol name.")
		return
	}

	cmd := "show route filtered all protocol '" + q + "'"
	resp, err := proxyreq.BirdRequest(id, cmd)
	if err != nil || strings.Contains(resp, "syntax error, unexpected CF_SYM_UNDEFINED") {
		if err != nil {
			log.Errorf("Failed to fetch filtered routes for %s (%s): %v", id, q, err)
		}
		f.renderModeErr(c, id, "Failed to fetch information. Please try again later.")
		return
	}

	f.renderBird(c, id, "filtered routes "+q, cmd, resp)
}

func (f *Frontend) handleTraceroute(c *gin.Context, id, q string) {
	isV4, isV6 := validator.IsIP(q)
	isDomain := validator.IsDomain(q)
	if !(isV4 || isV6 || isDomain) {
		f.renderModeErr(c, id, "Invalid IP address or domain name.")
		return
	}

	resp, err := proxyreq.TracerouteRequest(id, q)
	if err != nil {
		log.Errorf("Failed to perform traceroute for %s (%s): %v", id, q, err)
		f.renderModeErr(c, id, "Failed to perform traceroute.")
		return
	}

	f.renderBird(c, id, "traceroute "+q, "traceroute "+q, resp)
}

func (f *Frontend) handleProtocol(c *gin.Context) {
	id := c.Param("id")
	p := c.Param("protocol")

	if !validator.IsValidProtocol(p) {
		f.renderErr(c, http.StatusBadRequest,
			"Invalid protocol name.", "/detail/"+id, "Go back to Summary")
		return
	}

	srv := serverslist.GetServerByID(id)
	if srv == nil {
		serverslist.NotifyFetchServersList()
		f.renderErr(c, http.StatusNotFound,
			"PoP Not found. Please try again later.", "/", "Go back to home")
		return
	}

	cmd := "show protocols all '" + p + "'"
	resp, err := proxyreq.BirdRequest(id, cmd)
	if err != nil || strings.Contains(resp, "syntax error, unexpected CF_SYM_UNDEFINED") {
		if err != nil {
			log.Errorf("Failed to fetch protocol details for %s (%s): %v", id, p, err)
		}
		f.renderErr(c, http.StatusInternalServerError,
			"Invalid protocol name or failed to fetch protocol details. Please try again later.",
			"/detail/"+id, "Go back to Summary")
		return
	}

	f.renderBird(c, id, p, cmd, resp)
}

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

func (f *Frontend) renderBird(c *gin.Context, id, q, cmd, raw string) {
	srv := serverslist.GetServerByID(id)
	render.RenderHTML(c, http.StatusOK, "bird.tmpl", gin.H{
		"Title":   id + " - " + q,
		"Server":  srv,
		"Command": cmd,
		"Raw":     summaryparser.SmartFormatter(strings.TrimSpace(raw)),
	})
}

func SetupRouter() *gin.Engine {
	f := New()
	return f.Engine()
}
