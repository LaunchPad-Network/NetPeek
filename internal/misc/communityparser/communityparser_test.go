package communityparser

import (
	"testing"
)

const definitions = `
65535:666,exact match
65535:1xxx,triple X wildcard match $0
65535:xxx1,triple X wildcard match B $0
65535:x0,match single digit wildcard $0
65535:nnn,match any number $0
65535:0:nnn,large community test $0
65535:x:nnn,large wildcard $0 $1
`

func TestExactMatch(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "TEST")

	input := "BGP.community: (65535, 666)"
	output := p.FormatBGPText(input)

	want := `<abbr class="smart-community" title="(65535, 666)">[TEST: exact match]</abbr>`
	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestSingleDigitWildcard(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 10)"
	output := p.FormatBGPText(input)

	// 65535:x0 → groups = ["1"]
	want := `<abbr class="smart-community" title="(65535, 10)">[match single digit wildcard 1]</abbr>`
	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestNNNWildcard(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 123456)"
	output := p.FormatBGPText(input)

	// 65535:nnn → group = ["123456"]
	want := `<abbr class="smart-community" title="(65535, 123456)">[match any number 123456]</abbr>`
	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestLargeCommunity(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "LARGE")

	input := "BGP.large_community: (65535, 0, 400)"
	output := p.FormatBGPText(input)

	// 65535:0:nnn → group = ["400"]
	want := `<abbr class="smart-community" title="(65535, 0, 400)">[LARGE: large community test 400]</abbr>`
	if output != "BGP.large_community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestLargeWildcard(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.large_community: (65535, 3, 999)"
	output := p.FormatBGPText(input)

	// 65535:x:nnn → groups = ["3", "999"]
	want := `<abbr class="smart-community" title="(65535, 3, 999)">[large wildcard 3 999]</abbr>`
	if output != "BGP.large_community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestMultipleCommunitiesInOneLine(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 666) (65535, 10)"
	output := p.FormatBGPText(input)

	want := "BGP.community: " +
		`<abbr class="smart-community" title="(65535, 666)">[exact match]</abbr> ` +
		`<abbr class="smart-community" title="(65535, 10)">[match single digit wildcard 1]</abbr>`

	if output != want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestTripleXWildcardSuffix(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 1123)"
	output := p.FormatBGPText(input)

	// 1xxx → group = ["123"]
	want := `<abbr class="smart-community" title="(65535, 1123)">[triple X wildcard match 123]</abbr>`

	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestTripleXWildcardSuffixWithLeadingZero(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 1023)"
	output := p.FormatBGPText(input)

	want := `<abbr class="smart-community" title="(65535, 1023)">[triple X wildcard match 23]</abbr>`

	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestTripleXWildcardPrefix(t *testing.T) {
	p := NewBGPCommunityProcessor(definitions, "")

	input := "BGP.community: (65535, 9871)"
	output := p.FormatBGPText(input)

	// xxx1 → group = ["987"]
	want := `<abbr class="smart-community" title="(65535, 9871)">[triple X wildcard match B 987]</abbr>`

	if output != "BGP.community: "+want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}
