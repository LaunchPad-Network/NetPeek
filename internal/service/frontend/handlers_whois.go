package frontend

import (
	"net/http"
	"strings"

	"html/template"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/whois"
	"github.com/gin-gonic/gin"
	"github.com/robert-nix/ansihtml"
)

func (f *Frontend) handleWhoisNotSupported(c *gin.Context) {
	f.renderErr(c, http.StatusNotAcceptable, "Not supported", "/", "Go back to home")
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
