package birdformatter

import (
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/asnlookup"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/communityparser"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
)

type SmartFormatterOptions struct {
	Server          *serverslist.Server
	CurrentProtocol string
	IsRouteOutput   bool
}

func SmartFormatter(s string, options SmartFormatterOptions) template.HTML {
	var result string
	s = template.HTMLEscapeString(s)
	for _, line := range strings.Split(s, "\n") {
		var lineFormatted string
		if strings.HasPrefix(strings.TrimSpace(line), "Neighbor AS:") || strings.HasPrefix(strings.TrimSpace(line), "Local AS:") {
			lineFormatted = regexp.MustCompile(`(\d+)`).ReplaceAllString(line, `<a href="/whois?q=AS${1}" class="smart-whois" target="_blank">${1}</a>`)
		} else if strings.HasPrefix(strings.TrimSpace(line), "BGP.as_path:") || strings.HasPrefix(strings.TrimSpace(line), "bgp_path:") {
			lineFormatted = regexp.MustCompile(`(\d+)`).ReplaceAllStringFunc(line, func(s string) string {
				return `<a href="/whois?q=AS` + s + `" class="smart-whois" target="_blank"><abbr class="smart-asn" title="` + asnlookup.Lookup.Lookup(s) + `">` + s + `</abbr></a>`
			})
		} else if strings.HasPrefix(strings.TrimSpace(line), "Routes:") && options.Server != nil && options.CurrentProtocol != "" {
			lineFormatted = regexp.MustCompile(`\b([1-9]\d*)\s+filtered\b`).ReplaceAllString(line, `<a href="/detail/`+options.Server.Id+`?mode=filter&q=`+url.QueryEscape(options.CurrentProtocol)+`" class="smart-whois" target="_blank">${1} filtered</a>`)
		} else {
			lineFormatted = regexp.MustCompile(`([a-zA-Z0-9\-]*\.([a-zA-Z]{2,3}){1,2})(\s|$)`).ReplaceAllString(line, `<a href="/whois?q=${1}" class="smart-whois" target="_blank">${1}</a>${3}`)
			lineFormatted = regexp.MustCompile(`\[AS(\d+)`).ReplaceAllString(lineFormatted, `[<a href="/whois?q=AS${1}" class="smart-whois" target="_blank">AS${1}</a>`)
			lineFormatted = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`).ReplaceAllString(lineFormatted, `<a href="/whois?q=${1}" class="smart-whois" target="_blank">${1}</a>`)
			lineFormatted = regexp.MustCompile(`(?i)(([a-f\d]{0,4}:){3,10}[a-f\d]{0,4})`).ReplaceAllString(lineFormatted, `<a href="/whois?q=${1}" class="smart-whois" target="_blank">${1}</a>`)
		}
		result += lineFormatted + "\n"
	}
	if options.IsRouteOutput {
		result = communityparser.ProcessBirdOutput(result)
	}
	return template.HTML(result)
}
