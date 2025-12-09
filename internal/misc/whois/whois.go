package whois

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AkaereProbeResult struct {
	Supported bool     // 是否支持彩色输出
	Schemes   []string // 支持的 scheme 列表
	Raw       string   // 原始服务器响应
}

// 基础 Whois 查询
func Whois(query string) string {
	whoisServer := viper.GetString("servers.whois")
	if whoisServer == "" {
		return ""
	}
	if !strings.Contains(whoisServer, ":") {
		whoisServer += ":43"
	}

	conn, err := net.DialTimeout("tcp", whoisServer, 5*time.Second)
	if err != nil {
		return err.Error()
	}
	defer conn.Close()

	_, err = conn.Write([]byte(query + "\r\n"))
	if err != nil {
		return err.Error()
	}

	buf := make([]byte, 65536)
	n, err := conn.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return err.Error()
	}
	return string(buf[:n])
}

// AkaereProtocol 探测
func AkaereProtocolProbe() (*AkaereProbeResult, error) {
	whoisServer := viper.GetString("servers.whois")
	if whoisServer == "" {
		return nil, fmt.Errorf("whois server not configured")
	}
	if !strings.Contains(whoisServer, ":") {
		whoisServer += ":43"
	}

	conn, err := net.DialTimeout("tcp", whoisServer, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("X-WHOIS-COLOR-PROBE: v1.0\r\n\r\n"))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return &AkaereProbeResult{Raw: line}, err
	}

	result := &AkaereProbeResult{
		Raw: line,
	}

	if strings.HasPrefix(line, "X-WHOIS-COLOR-SUPPORT:") {
		result.Supported = true
		if idx := strings.Index(line, "schemes="); idx != -1 {
			schemesPart := strings.TrimSpace(line[idx+len("schemes="):])
			schemesPart = strings.TrimRight(schemesPart, "\r\n")
			result.Schemes = strings.Split(schemesPart, ",")
		}
	}

	return result, nil
}

func AkaereProtocolWhois(scheme, query string) (string, error) {
	whoisServer := viper.GetString("servers.whois")
	if whoisServer == "" {
		return "", fmt.Errorf("whois server not configured")
	}
	if !strings.Contains(whoisServer, ":") {
		whoisServer += ":43"
	}

	conn, err := net.DialTimeout("tcp", whoisServer, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	req := fmt.Sprintf("X-WHOIS-COLOR: scheme=%s\r\n%s\r\n\r\n", scheme, query)
	_, err = conn.Write([]byte(req))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	var builder strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		builder.WriteString(line)
	}

	return builder.String(), nil
}
