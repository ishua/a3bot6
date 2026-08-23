package main

import (
	"strings"
	"testing"
)

func TestSplitMessageShort(t *testing.T) {
	text := "short message"
	parts := splitMessage(text)
	if len(parts) != 1 || parts[0] != text {
		t.Fatalf("short message should be returned as one part, got %q", parts)
	}
}

func TestSplitMessageNoPieces(t *testing.T) {
	parts := splitMessage("")
	if len(parts) != 1 || parts[0] != "" {
		t.Fatalf("empty message should produce one empty part, got %q", parts)
	}
}

func TestSplitMessageLong(t *testing.T) {
	text := strings.Repeat("a", maxMsgLen*2+123)
	parts := splitMessage(text)
	joined := strings.Join(parts, "")
	if joined != text {
		t.Fatalf("joined parts do not reconstruct original text")
	}
	for i, p := range parts {
		if len(p) > maxMsgLen {
			t.Fatalf("part %d exceeds %d chars: %d", i, maxMsgLen, len(p))
		}
		if len(p) == 0 {
			t.Fatalf("part %d is empty", i)
		}
	}
}

func TestSplitMessageUnicode(t *testing.T) {
	// Cyrillic runes are multi-byte in UTF-8; splitting must preserve runes.
	text := strings.Repeat("яж", maxMsgLen+10)
	parts := splitMessage(text)
	if strings.Join(parts, "") != text {
		t.Fatalf("unicode text not reconstructed")
	}
	for i, p := range parts {
		if len([]rune(p)) > maxMsgLen {
			t.Fatalf("part %d exceeds %d runes", i, maxMsgLen)
		}
	}
}
