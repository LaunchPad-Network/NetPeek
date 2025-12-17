package traceroute

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net"
	"strings"
	"time"

	"github.com/syepes/network_exporter/pkg/common"
	"github.com/syepes/network_exporter/pkg/mtr"
)

// PTR lookup
func lookupPTR(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

// padRight pads string to fixed width (rune-safe)
func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}

// splitLines splits string into lines of max width (rune-safe)
func splitLines(s string, width int) []string {
	r := []rune(s)
	var lines []string
	for len(r) > width {
		lines = append(lines, string(r[:width]))
		r = r[width:]
	}
	if len(r) > 0 {
		lines = append(lines, string(r))
	}
	return lines
}

func formatString(out *mtr.MtrResult) string {
	const (
		ipWidth  = 39
		ptrWidth = 48
	)

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf(
		"Start: %v, DestAddr: %v\n",
		time.Now().Format("2006-01-02 15:04:05"),
		out.DestAddr,
	))

	if len(out.Hops) == 0 {
		buffer.WriteString("Expected at least one hop\n")
		return buffer.String()
	}

	// Header
	buffer.WriteString(fmt.Sprintf(
		"%-3s %-39s %10s%c %10s %10s %10s %10s %10s %-48s\n",
		"", "IP", "Loss", '%', "Snt", "Last", "Avg", "Best", "Worst", "PTR",
	))

	var lastHop int

	for _, hop := range out.Hops {
		if hop.Success {
			ptr := lookupPTR(hop.AddressTo)
			if ptr == "" {
				ptr = "-"
			}

			ptrLines := splitLines(ptr, ptrWidth)

			for i, line := range ptrLines {
				if i == 0 {
					buffer.WriteString(fmt.Sprintf(
						"%-3d %-39s %10.1f%c %10d %10.2f %10.2f %10.2f %10.2f %-48s\n",
						hop.TTL,
						padRight(hop.AddressTo, ipWidth),
						hop.Loss,
						'%',
						hop.Snt,
						common.Time2Float(hop.LastTime),
						common.Time2Float(hop.AvgTime),
						common.Time2Float(hop.BestTime),
						common.Time2Float(hop.WorstTime),
						padRight(line, ptrWidth),
					))
				} else {
					buffer.WriteString(fmt.Sprintf(
						"%-3s %-39s %10s%c %10s %10s %10s %10s %10s %-48s\n",
						"",
						padRight(hop.AddressTo, ipWidth),
						"", '%', "", "", "", "", "",
						padRight(line, ptrWidth),
					))
				}
			}

			lastHop = hop.TTL
		} else {
			lastHop++
			buffer.WriteString(fmt.Sprintf(
				"%-3d %-39s %10.1f%c %10d %10.2f %10.2f %10.2f %10.2f %-48s\n",
				lastHop,
				padRight("???", ipWidth),
				100.0,
				'%',
				0, 0.0, 0.0, 0.0, 0.0,
				padRight("???", ptrWidth),
			))
		}
	}

	return buffer.String()
}

type mtrHopView struct {
	Hop     int
	Host    string
	Loss    float32
	Snt     int
	Last    float32
	Avg     float32
	Best    float32
	Worst   float32
	PTR     string
	Success bool
	IsFinal bool
}

type mtrView struct {
	Start    string
	DestAddr string
	Hops     []mtrHopView
}

func buildMtrView(out *mtr.MtrResult) *mtrView {
	view := &mtrView{
		Start:    time.Now().Format("2006-01-02 15:04:05"),
		DestAddr: out.DestAddr,
		Hops:     make([]mtrHopView, 0, len(out.Hops)),
	}

	for index, hop := range out.Hops {
		hopView := mtrHopView{
			Hop:     hop.TTL,
			Host:    hop.AddressTo,
			Loss:    float32(hop.Loss),
			Snt:     hop.Snt,
			Last:    common.Time2Float(hop.LastTime),
			Avg:     common.Time2Float(hop.AvgTime),
			Best:    common.Time2Float(hop.BestTime),
			Worst:   common.Time2Float(hop.WorstTime),
			Success: hop.Success,
			IsFinal: false,
		}

		if !hop.Success {
			hopView.Host = "???"
			hopView.Loss = 100
			hopView.Snt = 0
			hopView.Last = 0
			hopView.Avg = 0
			hopView.Best = 0
			hopView.Worst = 0
			hopView.PTR = ""

			if index == len(out.Hops)-1 {
				hopView.IsFinal = true
			}
		} else {
			ptr := lookupPTR(hop.AddressTo)
			hopView.PTR = ptr
		}

		view.Hops = append(view.Hops, hopView)
	}

	return view
}

//go:embed template.html
var htmlTmplFS embed.FS

func formatHTML(out *mtr.MtrResult) (string, error) {
	view := buildMtrView(out)

	tmpl, err := template.ParseFS(htmlTmplFS, "template.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		return "", err
	}

	return buf.String(), nil
}
