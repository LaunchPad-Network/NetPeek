package render

import (
	"net/http"

	"github.com/LaunchPad-Network/NetPeek/internal/config"
	"github.com/LaunchPad-Network/NetPeek/internal/version"
	"github.com/gin-gonic/gin"
)

type VersionInfo struct {
	Commit    string
	CommitSHA string
}

func RenderHTML(c *gin.Context, code int, tpl string, obj any) bool {
	branding := config.GetBrandingInfo()
	version := VersionInfo{
		Commit:    version.CommitHash(),
		CommitSHA: string([]rune(version.CommitHash())[:min(10, len([]rune(version.CommitHash())))]),
	}

	switch v := obj.(type) {
	case nil:
		data := gin.H{
			"branding": branding,
			"version":  version,
		}
		c.HTML(code, tpl, data)
		return true

	case map[string]any:
		v["branding"] = branding
		v["version"] = version
		c.HTML(code, tpl, v)
		return true

	case gin.H:
		v["branding"] = branding
		v["version"] = version
		c.HTML(code, tpl, v)
		return true

	default:
		c.AbortWithStatus(http.StatusInternalServerError)

		return true
	}
}
