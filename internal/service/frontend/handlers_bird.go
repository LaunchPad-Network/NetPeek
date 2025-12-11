package frontend

import (
	"net/http"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/birdformatter"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/proxyreq"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/validator"
	"github.com/gin-gonic/gin"
)

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

func (f *Frontend) renderBird(c *gin.Context, id, q, cmd, raw string) {
	srv := serverslist.GetServerByID(id)
	render.RenderHTML(c, http.StatusOK, "bird.tmpl", gin.H{
		"Title":   id + " - " + q,
		"Server":  srv,
		"Command": cmd,
		"Raw":     birdformatter.SmartFormatter(strings.TrimSpace(raw)),
	})
}
