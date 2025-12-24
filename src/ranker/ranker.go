package ranker

import (
	"sort"
	"strings"

	"github.com/sakolb/bcd/src/entry"
)

const (
	scoreMatch        = 16
	scoreGapStart     = -3
	scoreGapExtension = -1

	bonusPathSeparator = 8
	bonusFirstChar     = 16
	bonusConsecutive   = 12
	bonusWordMatch     = 4 // bonus when gap contains only word chars (not delimiters)
)

type ScoredEntry struct {
	Entry *entry.PathEntry
	Score int
}

type Ranker struct {
	entries []*entry.PathEntry
	query   string
	results []ScoredEntry
}

func NewRanker() *Ranker {
	return &Ranker{
		entries: make([]*entry.PathEntry, 0),
	}
}

func (r *Ranker) AddEntry(e *entry.PathEntry) {
	r.entries = append(r.entries, e)
}

func (r *Ranker) SetQuery(q string) {
	r.query = q
	r.computeResults()
}

func (r *Ranker) Results() []ScoredEntry {
	return r.results
}

func (r *Ranker) computeResults() {
	r.results = r.results[:0]

	if r.query == "" {
		for _, e := range r.entries {
			r.results = append(r.results, ScoredEntry{Entry: e, Score: 0})
		}
		sort.Slice(r.results, func(i, j int) bool {
			return r.results[i].Entry.Distance < r.results[j].Entry.Distance
		})
		return
	}

	for _, e := range r.entries {
		matched, s := score(r.query, e.AbsPath)
		if matched {
			r.results = append(r.results, ScoredEntry{Entry: e, Score: s})
		}
	}

	sort.Slice(r.results, func(i, j int) bool {
		if r.results[i].Score != r.results[j].Score {
			return r.results[i].Score > r.results[j].Score
		}
		return r.results[i].Entry.Distance < r.results[j].Entry.Distance
	})
}

func isDelimiter(r rune) bool {
	return r == '/' || r == '-' || r == '_' || r == '.' || r == ' '
}

func score(query, target string) (bool, int) {
	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)

	if len(queryLower) == 0 {
		return false, 0
	}

	queryRunes := []rune(queryLower)
	targetRunes := []rune(targetLower)
	originalRunes := []rune(target)

	// Fast rejection: subsequence check
	qi := 0
	for ti := 0; ti < len(targetRunes) && qi < len(queryRunes); ti++ {
		if targetRunes[ti] == queryRunes[qi] {
			qi++
		}
	}
	if qi < len(queryRunes) {
		return false, 0
	}

	// Compute score
	qi = 0
	totalScore := 0
	prevMatchIdx := -1

	for ti := 0; ti < len(targetRunes) && qi < len(queryRunes); ti++ {
		if targetRunes[ti] == queryRunes[qi] {
			totalScore += scoreMatch

			if ti == 0 {
				totalScore += bonusFirstChar
			}

			if ti > 0 && originalRunes[ti-1] == '/' {
				totalScore += bonusPathSeparator
			}

			if prevMatchIdx >= 0 {
				gap := ti - prevMatchIdx - 1
				if gap == 0 {
					totalScore += bonusConsecutive
				} else {
					totalScore += scoreGapStart

					// Check if gap contains only word characters (no delimiters)
					// This rewards "config" over "c-f-g"
					hasDelimiter := false
					for gi := prevMatchIdx + 1; gi < ti; gi++ {
						if isDelimiter(originalRunes[gi]) {
							hasDelimiter = true
							break
						}
					}
					if !hasDelimiter {
						totalScore += bonusWordMatch
					}
				}
			}

			prevMatchIdx = ti
			qi++
		}
	}

	return true, totalScore
}
