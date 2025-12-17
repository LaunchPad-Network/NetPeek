package proxyreq

import (
	"net/url"
	"strconv"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/net"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/proxyreqsign"

	"github.com/lfcypo/viperx"
	"github.com/spf13/viper"
)

func buildProxyUrl(node, kind, q string) (string, error) {
	reqS, err := proxyreqsign.Sign(q)
	if err != nil {
		return "", err
	}
	proxyUrl := "http://" + node + viper.GetString("servers.proxy_suffix") + ":" + viperx.GetString("servers.proxy_port", "10179") + "/" + kind + "?q=" + url.QueryEscape(reqS.Query) + "&ts=" + strconv.FormatInt(reqS.Ts, 10) + "&sig=" + reqS.Signature
	return proxyUrl, nil
}

func BirdRequest(node, q string) (string, error) {
	url, err := buildProxyUrl(node, "bird", q)
	if err != nil {
		return "", err
	}
	return net.FetchURLWithTimeoutAsPlaintext(url, viperx.GetInt("servers.timeout", 5))
}

func TracerouteRequest(node, q string) (string, error) {
	url, err := buildProxyUrl(node, "traceroute", q)
	if err != nil {
		return "", err
	}
	return net.FetchURLWithTimeoutAsPlaintext(url, viperx.GetInt("servers.timeout", 5))
}

func TracerouteHTMLRequest(node, q string) (string, error) {
	url, err := buildProxyUrl(node, "tracerouteh", q)
	if err != nil {
		return "", err
	}
	return net.FetchURLWithTimeoutAsPlaintext(url, viperx.GetInt("servers.timeout", 5))
}
