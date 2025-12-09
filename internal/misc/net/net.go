package net

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/LaunchPad-Network/NetPeek/internal/logger"
)

var log = logger.New("Net")

func createConnectionTimeoutRoundTripper(timeout int) http.RoundTripper {
	context := net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return &http.Transport{
		DialContext: context.DialContext,

		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func FetchURLWithTimeout(url string, timeout int) (*http.Response, error) {
	log.Debugf("Fetching URL: %s", url)

	client := &http.Client{
		Transport: createConnectionTimeoutRoundTripper(timeout),
	}

	return client.Get(url)
}

func FetchURLWithTimeoutAsPlaintext(url string, timeout int) (string, error) {
	resp, err := FetchURLWithTimeout(url, timeout)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
