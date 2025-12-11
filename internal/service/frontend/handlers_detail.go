package frontend

import (
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/proxyreq"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/render"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/summaryparser"
	"github.com/gin-gonic/gin"
)

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
