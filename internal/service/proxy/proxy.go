package proxy

import (
	"strconv"

	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/bird"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/proxyreqsign"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/traceroute"
	"github.com/LaunchPad-Network/NetPeek/internal/router"

	"github.com/gin-gonic/gin"
)

var log = logger.New("Proxy")

func SetupRouter() *gin.Engine {
	r := router.SetupRouter()

	r.GET("/bird", birdHandler)
	r.GET("/traceroute", tracerouteHandler)

	return r
}

func securityCheck(c *gin.Context) (string, bool) {
	q := c.Query("q")
	ts := c.Query("ts")
	sig := c.Query("sig")
	if q == "" || ts == "" || sig == "" {
		c.String(400, "Invalid parameters")
		return "", false
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		c.String(400, "Invalid parameters")
		return "", false
	}
	spr := proxyreqsign.SignedProxyRequest{
		Query:     q,
		Ts:        tsInt,
		Signature: sig,
	}
	if !spr.Verify() {
		c.String(403, "Invalid authentication")
		return "", false
	}
	return q, true
}

func birdHandler(c *gin.Context) {
	q, ok := securityCheck(c)
	if !ok {
		return
	}
	c.Writer.WriteHeader(200)
	bird.CallBirdRestricted(q, c.Writer)
}

func tracerouteHandler(c *gin.Context) {
	q, ok := securityCheck(c)
	if !ok {
		return
	}
	r, err := traceroute.CallTraceroute(q)
	if err != nil {
		log.Errorf("traceroute error: %v", err)
		c.String(500, "Internal server error")
		return
	}
	c.String(200, r)
}
