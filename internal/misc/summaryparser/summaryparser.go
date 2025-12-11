package summaryparser

import (
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// summary
type SummaryRowData struct {
	Name        string `json:"name"`
	Proto       string `json:"proto"`
	Table       string `json:"table"`
	State       string `json:"state"`
	MappedState string `json:"-"`
	Since       string `json:"since"`
	Info        string `json:"info"`
}

// utility functions to allow filtering of results in the template

func (r SummaryRowData) NameHasPrefix(prefix string) bool {
	return strings.HasPrefix(r.Name, prefix)
}

func (r SummaryRowData) NameContains(prefix string) bool {
	return strings.Contains(r.Name, prefix)
}

func (r SummaryRowData) ProtocolMatches(protocols []string) bool {
	for _, protocol := range protocols {
		if strings.EqualFold(r.Proto, protocol) {
			return true
		}
	}
	return false
}

// pre-compiled regexp and constant statemap for summary rendering
var splitSummaryLine = regexp.MustCompile(`^([\w-]+)\s+(\w+)\s+([\w-]+)\s+(\w+)\s+([0-9\-\. :]+)(.*)$`)
var summaryStateMap = map[string]string{
	"up":      "green",
	"down":    "zinc",
	"start":   "red",
	"passive": "blue",
}

func SummaryRowDataFromLine(line string) *SummaryRowData {
	lineSplitted := splitSummaryLine.FindStringSubmatch(line)
	if lineSplitted == nil {
		return nil
	}

	var row SummaryRowData
	row.Name = strings.TrimSpace(lineSplitted[1])
	row.Proto = strings.TrimSpace(lineSplitted[2])
	row.Table = strings.TrimSpace(lineSplitted[3])
	row.State = strings.TrimSpace(lineSplitted[4])
	row.Since = strings.TrimSpace(lineSplitted[5])
	row.Info = strings.TrimSpace(lineSplitted[6])

	if strings.Contains(row.Info, "Passive") {
		row.MappedState = summaryStateMap["passive"]
	} else {
		row.MappedState = summaryStateMap[row.State]
	}

	return &row
}

type TemplateSummary struct {
	Raw    string
	Header []string
	Rows   []SummaryRowData
}

func SummaryParse(data string) (TemplateSummary, error) {
	args := TemplateSummary{
		Raw: data,
	}

	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) <= 1 {
		// Likely backend returned an error message
		return args, errors.New(strings.TrimSpace(data))
	}

	// extract the table header
	for _, col := range strings.Split(lines[0], " ") {
		colTrimmed := strings.TrimSpace(col)
		if len(colTrimmed) == 0 {
			continue
		}
		if col == "Table" {
			continue
		}
		args.Header = append(args.Header, col)
	}

	nameFilter := viper.GetString("frontend.name_filter")

	// Build regexp for nameFilter
	nameFilterRegexp := regexp.MustCompile(nameFilter)

	// sort the remaining rows
	rows := lines[1:]
	sort.Strings(rows)

	// parse each line
	for _, line := range rows {
		row := SummaryRowDataFromLine(line)
		if row == nil {
			continue
		}

		// Filter row name
		if nameFilter != "" && nameFilterRegexp.MatchString(row.Name) {
			continue
		}

		// add to the result
		args.Rows = append(args.Rows, *row)
	}

	return args, nil
}
