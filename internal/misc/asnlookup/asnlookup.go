package asnlookup

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/LaunchPad-Network/NetPeek/asnlookup2"
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/whois"
	"github.com/lfcypo/viperx"
	"github.com/patrickmn/go-cache"
)

var log = logger.New("ASN Lookup")

type ASNLookup struct {
	cache   *cache.Cache
	lookup2 asnlookup2.Lookup
}

var Lookup = NewASNLookup(24 * time.Hour)

// NewASNLookup 创建新的 ASNLookup
// defaultExpiration 默认缓存过期时间
func NewASNLookup(defaultExpiration time.Duration) *ASNLookup {
	l2config := asnlookup2.DefaultConfig()
	l2config.DataDir = viperx.GetString("asnlookup2.datadir", "./cache/asn_data")
	lookup2, err := asnlookup2.New(l2config)
	if err != nil {
		log.Fatal("Lookup2 initialization fail, err: ", err)
	}
	lookup2.Start(context.Background())

	return &ASNLookup{
		cache:   cache.New(defaultExpiration, time.Minute),
		lookup2: lookup2,
	}
}

// LookupByDNS 查询 ASN 名称
func (a *ASNLookup) LookupByDNS(asn string) (string, error) {
	if v, found := a.cache.Get(asn); found {
		return v.(string), nil
	}

	txt, err := queryASNText(asn)
	if err != nil {
		return "", err
	}

	name := parseASNName(txt)
	if name == "" {
		return "", fmt.Errorf("cannot parse ASN name from TXT: %s", txt)
	}

	a.cache.Set(asn, name, cache.DefaultExpiration)

	return name, nil
}

// LookupByWHOIS 查询 ASN 名称
func (a *ASNLookup) LookupByWHOIS(asn string) (string, error) {
	if v, found := a.cache.Get(asn); found {
		return v.(string), nil
	}

	txt := whois.Whois("AS" + asn)
	if txt == "" {
		return "", fmt.Errorf("cannot query WHOIS")
	}

	name := extractASNName(txt)
	if name == "" {
		return "", fmt.Errorf("cannot extract ASN name from WHOIS: %s", txt)
	}

	a.cache.Set(asn, name, cache.DefaultExpiration)

	return name, nil
}

// Lookup 查询 ASN 名称
func (a *ASNLookup) Lookup(asn string) string {
	parsed, err := asnlookup2.ParseASN("AS" + asn)
	if err != nil {
		return "AS NAME LOOKUP FAILURE"
	}
	q2, err := a.lookup2.Query(parsed)
	if err == nil {
		if q2.Name != "" {
			return q2.Name
		}
	}
	log.Debug("Lookup2 failed, err: ", err, ", fallback to Lookup1")

	name, err := a.LookupByDNS(asn)
	if err != nil {
		name, err = a.LookupByWHOIS(asn)
	}

	if name == "" {
		name = "AS" + asn
	}

	return name
}

// queryASNText 使用 dig 查询 ASN TXT 记录
func queryASNText(asn string) (string, error) {
	fqdn := fmt.Sprintf("as%s.asn.cymru.com", strings.TrimPrefix(asn, "AS"))
	ips, err := net.LookupTXT(fqdn)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no TXT record found for %s", fqdn)
	}
	return ips[0], nil
}

// parseASNName 解析 TXT 字符串，返回名称
func parseASNName(txt string) string {
	// 按 | 分割
	parts := strings.Split(txt, "|")
	if len(parts) < 5 {
		return strings.TrimSpace(txt)
	}

	// 第五个部分
	raw := strings.TrimSpace(parts[4])

	// 去掉末尾的 ", 国家/地区"
	raw = strings.SplitN(raw, ",", 2)[0]

	return strings.TrimSpace(raw)
}

func extractASNName(txt string) string {
	lines := strings.Split(txt, "\n")

	fields := []string{"org-name", "OrgName", "as-name", "ASName", "descr"}

	for _, field := range fields {
		re := regexp.MustCompile(fmt.Sprintf(`(?i)^%s\s*[:=]\s*(.+)$`, field))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if m := re.FindStringSubmatch(line); m != nil {
				return strings.TrimSpace(m[1])
			}
		}
	}

	return ""
}
