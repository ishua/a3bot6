package taskparser

import (
	"reflect"
	"testing"
	"time"
)

func TestIsTaskLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"dash task", "- foo", true},
		{"star task", "* foo", true},
		{"plain text", "foo", false},
		{"dash no space", "-foo", false},
		{"leading space", " - foo", false},
		{"empty", "", false},
		{"heading", "### Today", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTaskLine(tt.line); got != tt.want {
				t.Errorf("isTaskLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsSubtaskLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"tab dash", "\t- foo", true},
		{"spaces dash", "  - foo", true},
		{"spaces star", "   * foo", true},
		{"task line", "- foo", false},
		{"indented text", "  foo", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSubtaskLine(tt.line); got != tt.want {
				t.Errorf("isSubtaskLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsSectionHeader(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"header", "### Today", true},
		{"double space", "###  winter 2031", true},
		{"h2", "## In Progress", false},
		{"h1", "# Long", false},
		{"no space", "###", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSectionHeader(tt.line); got != tt.want {
				t.Errorf("isSectionHeader(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestIsStopLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		stopMarker string
		want       bool
	}{
		{"matches marker", "### Long", "### Long", true},
		{"separator", "---", "### Long", true},
		{"separator no marker", "---", "", true},
		{"not stop", "something", "### Long", false},
		{"empty marker no match", "something", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStopLine(tt.line, tt.stopMarker); got != tt.want {
				t.Errorf("isStopLine(%q, %q) = %v, want %v", tt.line, tt.stopMarker, got, tt.want)
			}
		})
	}
}

func TestExtractPriority(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantPri  string
		wantRest string
	}{
		{"A", "(A) do thing", "A", "do thing"},
		{"B", "(B) do thing", "B", "do thing"},
		{"none", "do thing", "", "do thing"},
		{"not marker", "(AB) do thing", "", "(AB) do thing"},
		{"lowercase", "(a) do thing", "", "(a) do thing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPri, gotRest := extractPriority(tt.line)
			if gotPri != tt.wantPri || gotRest != tt.wantRest {
				t.Errorf("extractPriority(%q) = (%q, %q), want (%q, %q)",
					tt.line, gotPri, gotRest, tt.wantPri, tt.wantRest)
			}
		})
	}
}

func TestExtractDueDate(t *testing.T) {
	valid, _ := time.Parse("2006-01-02", "2030-07-13")
	tests := []struct {
		name     string
		text     string
		wantDue  time.Time
		wantRest string
		wantWarn bool
	}{
		{"valid only", "due:2030-07-13", valid, "", false},
		{"valid with text", "foo due:2030-07-13 bar", valid, "foo bar", false},
		{"no due", "plain text", time.Time{}, "plain text", false},
		{"invalid month", "due:2030-13-45", time.Time{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDue, gotRest, gotWarn := extractDueDate(tt.text)
			if !gotDue.Equal(tt.wantDue) {
				t.Errorf("extractDueDate(%q) due = %v, want %v", tt.text, gotDue, tt.wantDue)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractDueDate(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
			if (gotWarn != nil) != tt.wantWarn {
				t.Errorf("extractDueDate(%q) warn = %v, want warn=%v", tt.text, gotWarn, tt.wantWarn)
			}
		})
	}
}

func TestExtractEmojis(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantEmojis []string
		wantRest   string
	}{
		{"single", "clean kitchen 🏡", []string{"🏡"}, "clean kitchen"},
		{"emoji before due", "💰 due:2030-07-13", []string{"💰"}, "due:2030-07-13"},
		{"multiple", "🏡 📞", []string{"🏡", "📞"}, ""},
		{"none", "plain text", []string{}, "plain text"},
		{"duplicate", "🏡 x 🏡", []string{"🏡", "🏡"}, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEmojis, gotRest := extractEmojis(tt.text)
			if !reflect.DeepEqual(gotEmojis, tt.wantEmojis) {
				t.Errorf("extractEmojis(%q) emojis = %v, want %v", tt.text, gotEmojis, tt.wantEmojis)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractEmojis(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
		})
	}
}

func TestExtractWikilinks(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantLinks []string
		wantRest  string
		wantWarn  bool
	}{
		{"single", "[[finance]] x", []string{"finance"}, "x", false},
		{"multiple", "a [[b]] c [[d]] e", []string{"b", "d"}, "a c e", false},
		{"embedded", "--[[x]]--", []string{"x"}, "----", false},
		{"unclosed", "[[unclosed", []string{}, "", true},
		{"none", "plain text", []string{}, "plain text", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, gotRest, gotWarns := extractWikilinks(tt.text)
			if !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("extractWikilinks(%q) links = %v, want %v", tt.text, gotLinks, tt.wantLinks)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractWikilinks(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
			if (len(gotWarns) > 0) != tt.wantWarn {
				t.Errorf("extractWikilinks(%q) warns = %v, want warn=%v", tt.text, gotWarns, tt.wantWarn)
			}
		})
	}
}

func TestExtractInlineLinks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		want     []InlineLink
		wantRest string
	}{
		{"single", "[link](url)", []InlineLink{{Text: "link", URL: "url"}}, ""},
		{"with text", "see [here](http://x) ok", []InlineLink{{Text: "here", URL: "http://x"}}, "see ok"},
		{"multiple", "[a](b) [c](d)", []InlineLink{{Text: "a", URL: "b"}, {Text: "c", URL: "d"}}, ""},
		{"none", "plain text", []InlineLink{}, "plain text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, gotRest := extractInlineLinks(tt.text)
			if !reflect.DeepEqual(gotLinks, tt.want) {
				t.Errorf("extractInlineLinks(%q) links = %v, want %v", tt.text, gotLinks, tt.want)
			}
			if gotRest != tt.wantRest {
				t.Errorf("extractInlineLinks(%q) rest = %q, want %q", tt.text, gotRest, tt.wantRest)
			}
		})
	}
}
