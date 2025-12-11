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

type WhoisQueryStrategy interface {
	Query(q string) (string, bool)
}

type AkaereWhoisStrategy struct{}
type DefaultWhoisStrategy struct{}

func (s *AkaereWhoisStrategy) Query(q string) (string, bool) {
	probeRes, err := whois.AkaereProtocolProbe()
	if err != nil {
		return "", false
	}

	if !probeRes.Supported {
		return "", false
	}
	if len(probeRes.Schemes) == 0 {
		return "", false
	}

	availableSchemes := make(map[string]bool)
	for _, scheme := range probeRes.Schemes {
		availableSchemes[scheme] = true
	}

	prioritySchemes := []string{"ripe-dark", "bgptools-dark", "ripe", "bgptools"}
	var selectedScheme string
	for _, scheme := range prioritySchemes {
		if availableSchemes[scheme] {
			selectedScheme = scheme
			break
		}
	}
	if selectedScheme == "" {
		selectedScheme = probeRes.Schemes[0]
	}
	log.Debugf("selected whois color scheme %s", selectedScheme)

	res, err := whois.AkaereProtocolWhois(selectedScheme, q)
	if err != nil {
		return "", false
	}

	htmlResult := ansihtml.ConvertToHTML([]byte(res))
	return string(htmlResult), true
}

func (s *DefaultWhoisStrategy) Query(q string) (string, bool) {
	res := whois.Whois(q)
	return strings.TrimSpace(res), true
}

type WhoisExecutor struct {
	strategies []WhoisQueryStrategy
}

func NewWhoisExecutor() *WhoisExecutor {
	return &WhoisExecutor{
		strategies: []WhoisQueryStrategy{
			&AkaereWhoisStrategy{},
			&DefaultWhoisStrategy{},
		},
	}
}

func (e *WhoisExecutor) Execute(q string) template.HTML {
	for _, strategy := range e.strategies {
		if res, ok := strategy.Query(q); ok {
			return template.HTML(res)
		}
	}
	return template.HTML("")
}

func (f *Frontend) handleWhois(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		render.RenderHTML(c, http.StatusOK, "whois.tmpl", gin.H{
			"Title": "WHOIS Query",
		})
		return
	}

	executor := NewWhoisExecutor()
	result := executor.Execute(q)

	render.RenderHTML(c, http.StatusOK, "whois_res.tmpl", gin.H{
		"Title":  "WHOIS Query - " + q,
		"Query":  q,
		"Result": result,
	})
}

func (f *Frontend) handleWhoisNotSupported(c *gin.Context) {
	f.renderErr(c, http.StatusNotAcceptable, "Not supported", "/", "Go back to home")
}
