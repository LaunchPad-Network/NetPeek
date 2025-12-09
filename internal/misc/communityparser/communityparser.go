package communityparser

import (
	"regexp"
	"strconv"
	"strings"
)

type CommunityEntry struct {
	Pattern     *regexp.Regexp
	Description string
	IsLarge     bool
	GroupCount  int
}

type BGPCommunityProcessor struct {
	outPrefix        string
	communityEntries []CommunityEntry
	largeEntries     []CommunityEntry
}

var communityValueRegex = regexp.MustCompile(`\((\d+),\s*(\d+)(?:,\s*(\d+))?\)`)

func NewBGPCommunityProcessor(communityDefinition, outPrefix string) *BGPCommunityProcessor {
	processor := &BGPCommunityProcessor{
		outPrefix: outPrefix,
	}

	processor.parseCommunityDefinitions(communityDefinition)

	return processor
}

func convertWildcardToRegex(patternPart string) (string, int) {
	result := ""
	groupCount := 0

	for i := 0; i < len(patternPart); i++ {
		ch := patternPart[i]

		if ch == 'x' {
			result += `(\d)`
			groupCount++
		} else if i+2 < len(patternPart) && patternPart[i:i+3] == "nnn" {
			result += `(\d+)`
			groupCount++
			i += 2
		} else {
			if strings.ContainsAny(string(ch), `\.+*?()|[]{}^$`) {
				result += `\` + string(ch)
			} else {
				result += string(ch)
			}
		}
	}

	return result, groupCount
}

func (p *BGPCommunityProcessor) parseCommunityDefinitions(definitions string) {
	lines := strings.Split(definitions, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue
		}

		patternStr := strings.TrimSpace(parts[0])
		description := strings.TrimSpace(parts[1])

		patternParts := strings.Split(patternStr, ":")
		isLarge := len(patternParts) == 3

		regexPattern := "^"
		groupCount := 0

		for i, part := range patternParts {
			part = strings.TrimSpace(part)

			if i > 0 {
				regexPattern += ":"
			}

			convertedPart, groups := convertWildcardToRegex(part)
			regexPattern += convertedPart
			groupCount += groups
		}
		regexPattern += "$"

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			continue
		}

		entry := CommunityEntry{
			Pattern:     re,
			Description: description,
			IsLarge:     isLarge,
			GroupCount:  groupCount,
		}

		if isLarge {
			p.largeEntries = append(p.largeEntries, entry)
		} else {
			p.communityEntries = append(p.communityEntries, entry)
		}
	}
}

func normalizeCommunityString(community string) string {
	normalized := strings.ReplaceAll(community, " ", "")
	normalized = strings.Trim(normalized, "()")
	normalized = strings.ReplaceAll(normalized, ",", ":")
	return normalized
}

func (p *BGPCommunityProcessor) findMatchingEntry(community string, isLarge bool) (CommunityEntry, []string, bool) {
	var entries []CommunityEntry
	if isLarge {
		entries = p.largeEntries
	} else {
		entries = p.communityEntries
	}

	normalized := normalizeCommunityString(community)

	for _, entry := range entries {
		if matches := entry.Pattern.FindStringSubmatch(normalized); matches != nil {
			if len(matches) > 1 {
				return entry, matches[1:], true
			}
			return entry, nil, true
		}
	}

	return CommunityEntry{}, nil, false
}

func (p *BGPCommunityProcessor) formatDescription(desc string, groups []string) string {
	if len(groups) == 0 {
		return desc
	}

	result := desc

	for i, group := range groups {
		placeholder := "$" + strconv.Itoa(i)
		result = strings.ReplaceAll(result, placeholder, group)
	}

	return result
}

func (p *BGPCommunityProcessor) FormatBGPText(s string) string {
	var result strings.Builder
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}

		if strings.Contains(line, "BGP.community:") || strings.Contains(line, "BGP.large_community:") {
			formattedLine := communityValueRegex.ReplaceAllStringFunc(line, func(match string) string {
				cleanMatch := strings.ReplaceAll(match, " ", "")
				parts := strings.Split(strings.Trim(cleanMatch, "()"), ",")
				isLarge := len(parts) == 3

				entry, groups, found := p.findMatchingEntry(match, isLarge)
				if !found {
					return match
				}

				formattedDesc := p.formatDescription(entry.Description, groups)

				var displayText string
				if p.outPrefix != "" {
					displayText = "[" + p.outPrefix + ": " + formattedDesc + "]"
				} else {
					displayText = "[" + formattedDesc + "]"
				}

				return `<abbr class="smart-community" title="` + strings.ReplaceAll(match, " ", "") + `">` + displayText + `</abbr>`
			})
			result.WriteString(formattedLine)
		} else {
			result.WriteString(line)
		}
	}

	return result.String()
}
