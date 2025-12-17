package traceroute

import (
	"net"
	"strings"
	"time"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/validator"
	"github.com/spf13/viper"
	"github.com/syepes/network_exporter/pkg/common"
	"github.com/syepes/network_exporter/pkg/mtr"
)

var icmpID = common.IcmpID{}

func callMTR(q string) (*mtr.MtrResult, error) {
	if viper.GetBool("traceroute.disable") {
		return nil, ErrNotSupported
	}

	q = strings.TrimSpace(q)
	if q == "" {
		return nil, ErrEmptyTarget
	}

	isV4, isV6 := validator.IsIP(q)
	isDomain := validator.IsDomain(q)
	if !isV4 && !isV6 && !isDomain {
		return nil, ErrInvalidTarget
	}
	target := q

	if isDomain {
		names, err := net.LookupHost(target)
		if err != nil {
			return nil, err
		}
		target = names[0]
		_, isV6 = validator.IsIP(target)
	}

	return mtr.Mtr(
		target,
		"",
		30,
		2,
		1*time.Second,
		int(icmpID.Get()),
		56,
		"icmp",
		"80",
		isV6,
	)
}

func CallTraceroute(q string) (string, error) {
	out, err := callMTR(q)
	if err != nil {
		return "", err
	}

	return formatString(out), nil
}

func CallTracerouteHTML(q string) (string, error) {
	out, err := callMTR(q)
	if err != nil {
		return "", err
	}

	return formatHTML(out)
}
