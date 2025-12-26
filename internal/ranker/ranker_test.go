package ranker

import (
	"testing"

	"github.com/sakolb/bcd/internal/entry"
)

func TestScoreSubsequenceMatch(t *testing.T) {
	tests := []struct {
		query   string
		target  string
		matched bool
	}{
		{"cfg", "config", true},
		{"cfg", "my-config", true},
		{"cfg", ".config.yaml", true},
		{"abc", "aXbXc", true},
		{"abc", "abc", true},
		{"abc", "cab", false},
		{"abc", "ab", false},
		{"", "anything", false}, // empty query matches nothing at score level
		{"a", "", false},
	}

	for _, tt := range tests {
		matched, _ := score(tt.query, tt.target)
		if matched != tt.matched {
			t.Errorf("score(%q, %q): got matched=%v, want %v",
				tt.query, tt.target, matched, tt.matched)
		}
	}
}

func TestScoreRanking(t *testing.T) {
	// Higher score = better match
	tests := []struct {
		query  string
		better string
		worse  string
	}{
		// Consecutive matches beat scattered
		{"cfg", "config", "c-f-g"},
		// Match at path boundary beats mid-word
		{"cfg", "/home/cfg", "/home/xcfg"},
		// Match at start beats mid-string
		{"abc", "abcdef", "xabcdef"},
	}

	for _, tt := range tests {
		_, scoreBetter := score(tt.query, tt.better)
		_, scoreWorse := score(tt.query, tt.worse)
		if scoreBetter <= scoreWorse {
			t.Errorf("score(%q): expected %q (score=%d) > %q (score=%d)",
				tt.query, tt.better, scoreBetter, tt.worse, scoreWorse)
		}
	}
}

func TestScoreCaseInsensitive(t *testing.T) {
	matched1, score1 := score("cfg", "CONFIG")
	matched2, score2 := score("CFG", "config")
	matched3, score3 := score("CfG", "CoNfIg")

	if !matched1 || !matched2 || !matched3 {
		t.Error("case insensitive matching failed")
	}

	if score1 != score2 || score2 != score3 {
		t.Errorf("case insensitive scores should be equal: %d, %d, %d",
			score1, score2, score3)
	}
}

func TestRankerEmptyQuery(t *testing.T) {
	r := NewRanker()

	// Create mock entries with different distances
	entries := []*entry.PathEntry{
		{AbsPath: "/home/user/far", Distance: 3},
		{AbsPath: "/home/close", Distance: 1},
		{AbsPath: "/home/user/middle", Distance: 2},
	}

	for _, e := range entries {
		r.AddEntry(e)
	}

	r.SetQuery("")
	results := r.Results()

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Should be sorted by distance ascending
	if results[0].Entry.Distance != 1 {
		t.Errorf("expected closest first, got distance=%d", results[0].Entry.Distance)
	}
	if results[2].Entry.Distance != 3 {
		t.Errorf("expected farthest last, got distance=%d", results[2].Entry.Distance)
	}
}

func TestRankerFiltersAndSorts(t *testing.T) {
	r := NewRanker()

	entries := []*entry.PathEntry{
		{AbsPath: "/home/user/config", Distance: 2},
		{AbsPath: "/home/cfg", Distance: 1},
		{AbsPath: "/home/user/nomatch", Distance: 1},
		{AbsPath: "/etc/config", Distance: 3},
	}

	for _, e := range entries {
		r.AddEntry(e)
	}

	r.SetQuery("cfg")
	results := r.Results()

	// Should filter out "nomatch"
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// All results should match
	for _, res := range results {
		matched, _ := score("cfg", res.Entry.AbsPath)
		if !matched {
			t.Errorf("result %q should not be in results", res.Entry.AbsPath)
		}
	}

	// First result should have highest score or lowest distance on tie
	for i := 1; i < len(results); i++ {
		prev := results[i-1]
		curr := results[i]
		if prev.Score < curr.Score {
			t.Errorf("results not sorted by score: %d < %d", prev.Score, curr.Score)
		}
		if prev.Score == curr.Score && prev.Entry.Distance > curr.Entry.Distance {
			t.Errorf("results not sorted by distance on tie")
		}
	}
}

func TestRankerUpdateQuery(t *testing.T) {
	r := NewRanker()

	entries := []*entry.PathEntry{
		{AbsPath: "/home/user/config", Distance: 1},
		{AbsPath: "/home/user/cache", Distance: 1},
		{AbsPath: "/home/user/code", Distance: 1},
	}

	for _, e := range entries {
		r.AddEntry(e)
	}

	r.SetQuery("cfg")
	results1 := r.Results()

	r.SetQuery("cache")
	results2 := r.Results()

	if len(results1) == len(results2) {
		// They might have different lengths
		t.Log("results may vary based on query")
	}

	// "cache" should match exactly one entry
	found := false
	for _, res := range results2 {
		if res.Entry.AbsPath == "/home/user/cache" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected /home/user/cache in results for query 'cache'")
	}
}

func TestDebugFourthMatch(t *testing.T) {
	targets := []string{
		"/home/user/config",
		"/home/cfg",
		"/home/user/nomatch",
		"/etc/config",
	}

	for _, target := range targets {
		matched, s := score("cfg", target)
		t.Logf("score(%q, %q) = matched:%v score:%d", "cfg", target, matched, s)
	}
}
