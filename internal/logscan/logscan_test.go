package logscan

import "testing"

func TestMatchFindsFailurePatternCaseInsensitive(t *testing.T) {
	pattern, ok := Match("server started\nFATAL database unavailable", []string{"panic", "fatal"})
	if !ok || pattern != "fatal" {
		t.Fatalf("expected fatal match, got %q %v", pattern, ok)
	}
}

func TestMatchIgnoresEmptyPatterns(t *testing.T) {
	if _, ok := Match("all good", []string{"", "panic"}); ok {
		t.Fatal("unexpected match")
	}
}
