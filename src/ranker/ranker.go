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
)

type ScoredEntry struct {
	Entry *entry.PathEntry
	Score int
}

type Ranker struct {
	entries       []*entry.PathEntry
	query         string
	previousQuery string
	results       []ScoredEntry
}

func NewRanker() *Ranker {
	return &Ranker{
		entries: make([]*entry.PathEntry, 0),
	}
}

func (r *Ranker) AddEntry(e *entry.PathEntry) {
	r.entries = append(r.entries, e)
	// Don't score here - too expensive with FZF v2 DP algorithm
	// Results get recomputed when query changes (debounced)
}

func (r *Ranker) SetQuery(q string) {
	if q == r.query {
		return // Query unchanged, skip recomputation
	}

	r.previousQuery = r.query
	r.query = q

	// Detect incremental query (user typed more characters)
	if r.previousQuery != "" && strings.HasPrefix(q, r.previousQuery) && len(q) > len(r.previousQuery) {
		r.computeResultsIncremental()
	} else {
		// Decremental or completely different query: full recompute
		r.computeResults()
	}
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

func (r *Ranker) computeResultsIncremental() {
	// Instead of checking all entries, only filter/rescore previous results
	filtered := r.results[:0] // Reuse slice to avoid allocation

	for _, scored := range r.results {
		matched, s := score(r.query, scored.Entry.AbsPath)
		if matched {
			scored.Score = s
			filtered = append(filtered, scored)
		}
	}

	r.results = filtered

	// Re-sort since scores changed (gaps/bonuses differ)
	sort.Slice(r.results, func(i, j int) bool {
		if r.results[i].Score != r.results[j].Score {
			return r.results[i].Score > r.results[j].Score
		}
		return r.results[i].Entry.Distance < r.results[j].Entry.Distance
	})
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

	// FZF v2 algorithm: Dynamic programming to find optimal match positions
	qLen := len(queryRunes)
	tLen := len(targetRunes)

	const negInf = -100000

	// M[i][j] = best score when query[i-1] matches at target[j-1]
	// H[i][j] = best overall score for query[0..i-1] in target[0..j-1]
	M := make([][]int, qLen+1)
	H := make([][]int, qLen+1)
	for i := range M {
		M[i] = make([]int, tLen+1)
		H[i] = make([]int, tLen+1)
		for j := range M[i] {
			M[i][j] = negInf
			H[i][j] = negInf
		}
	}

	// Base case: matching 0 query chars = 0 score
	for j := 0; j <= tLen; j++ {
		H[0][j] = 0
	}

	// Fill DP table
	for i := 1; i <= qLen; i++ {
		for j := 1; j <= tLen; j++ {
			// Try to match query[i-1] with target[j-1]
			if queryRunes[i-1] == targetRunes[j-1] {
				bonus := scoreMatch

				// Position-based bonuses
				if j == 1 {
					bonus += bonusFirstChar
				} else if originalRunes[j-2] == '/' {
					bonus += bonusPathSeparator
				}

				if i == 1 {
					// First query character
					M[i][j] = bonus
				} else {
					// Option 1: consecutive match (previous query char matched at j-1)
					consecutiveScore := negInf
					if M[i-1][j-1] > negInf {
						consecutiveScore = M[i-1][j-1] + bonus + bonusConsecutive
					}

					// Option 2: gap (previous query char matched somewhere before j-1)
					gapScore := H[i-1][j-1] + bonus + scoreGapStart

					M[i][j] = max(consecutiveScore, gapScore)
				}
			}

			// H[i][j] = max(skip target[j-1], match at target[j-1])
			H[i][j] = max(H[i][j-1]+scoreGapExtension, M[i][j])
		}
	}

	return true, H[qLen][tLen]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
