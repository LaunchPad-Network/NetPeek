package validator

import (
	"net"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

// IsIP 判断给定字符串是否为合法 IP 地址，分别返回 isIPv4, isIPv6
func IsIP(s string) (bool, bool) {
	ip := net.ParseIP(s)
	if ip == nil {
		return false, false
	}
	// To4 returns a 4-byte representation for IPv4, otherwise nil
	if ip.To4() != nil {
		return true, false
	}
	return false, true
}

// IsCIDR 判断给定字符串是否为合法 CIDR，分别返回 isIPv4CIDR, isIPv6CIDR
func IsCIDR(s string) (bool, bool) {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil || ipnet == nil {
		return false, false
	}
	// Determine address family by IP in IPNet
	ip := ipnet.IP
	if ip.To4() != nil {
		return true, false
	}
	return false, true
}

func IsDomain(s string) bool {
	if s == "" {
		return false
	}
	if net.ParseIP(s) != nil {
		return false
	}
	raw := s
	if strings.HasSuffix(raw, ".") {
		raw = strings.TrimSuffix(raw, ".")
		if raw == "" {
			return false
		}
	}
	ascii, err := idna.Lookup.ToASCII(raw)
	if err != nil {
		return false
	}
	if len(ascii) > 255 {
		return false
	}
	labels := strings.Split(ascii, ".")
	labelRe := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	for _, lab := range labels {
		if lab == "" {
			return false
		}
		if len(lab) < 1 || len(lab) > 63 {
			return false
		}
		if !labelRe.MatchString(lab) {
			return false
		}
	}
	return true
}

func IsValidProtocol(s string) bool {
	reg := regexp.MustCompile(`^[0-9A-Za-z_-]+$`)
	return reg.MatchString(s)
}
