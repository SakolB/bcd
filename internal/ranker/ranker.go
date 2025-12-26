package ranker

import (
	"container/heap"
	"strings"

	"github.com/sakolb/bcd/internal/entry"
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

// ResultsHeap implements heap.Interface for max-heap (highest scores first)
type ResultsHeap []ScoredEntry

func (h ResultsHeap) Len() int { return len(h) }

func (h ResultsHeap) Less(i, j int) bool {
	// Max-heap: higher scores come first
	if h[i].Score != h[j].Score {
		return h[i].Score > h[j].Score
	}
	// Tie-break by distance (closer is better)
	return h[i].Entry.Distance < h[j].Entry.Distance
}

func (h ResultsHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *ResultsHeap) Push(x any) {
	*h = append(*h, x.(ScoredEntry))
}

func (h *ResultsHeap) Pop() any {
	oldHeap := *h
	n := len(oldHeap)
	element := oldHeap[n-1]
	oldHeap[n-1] = ScoredEntry{}
	*h = oldHeap[:n-1]
	return element
}

type Ranker struct {
	entries       []*entry.PathEntry
	query         string
	previousQuery string
	resultsHeap   *ResultsHeap
}

func NewRanker() *Ranker {
	h := &ResultsHeap{}
	return &Ranker{
		entries:     make([]*entry.PathEntry, 0),
		resultsHeap: h,
	}
}

func (r *Ranker) AddEntry(e *entry.PathEntry) {
	r.entries = append(r.entries, e)
	// Don't score individual entries - use AddEntryBatch instead
}

func (r *Ranker) AddEntryBatch(batch []*entry.PathEntry) {
	r.entries = append(r.entries, batch...)

	// Score each entry in batch and insert into heap
	for _, e := range batch {
		if r.query != "" {
			matched, s := score(r.query, e.AbsPath)
			if matched {
				heap.Push(r.resultsHeap, ScoredEntry{Entry: e, Score: s})
			}
		} else {
			// No query = show all, sorted by distance
			heap.Push(r.resultsHeap, ScoredEntry{Entry: e, Score: 0})
		}
	}
}

func (r *Ranker) SetQuery(q string) {
	if q == r.query {
		return // Query unchanged, skip recomputation
	}

	r.previousQuery = r.query
	r.query = q

	// Rebuild heap with new query
	r.rebuildHeap()
}

func (r *Ranker) rebuildHeap() {
	// Detect incremental query for optimization
	if r.previousQuery != "" && strings.HasPrefix(r.query, r.previousQuery) && len(r.query) > len(r.previousQuery) {
		// Incremental: only rescore entries that matched previous query
		// The current heap already contains only matched entries!
		r.scoreMatchedEntries()
	} else {
		// Full rescore: clear heap and score all entries
		r.resultsHeap = &ResultsHeap{}
		heap.Init(r.resultsHeap)
		r.scoreAllEntries()
	}
}

func (r *Ranker) scoreAllEntries() {
	if r.query == "" {
		// No query: show all by distance
		for _, e := range r.entries {
			heap.Push(r.resultsHeap, ScoredEntry{Entry: e, Score: 0})
		}
		return
	}

	// Score all entries and push to heap
	for _, e := range r.entries {
		matched, s := score(r.query, e.AbsPath)
		if matched {
			heap.Push(r.resultsHeap, ScoredEntry{Entry: e, Score: s})
		}
	}
}

func (r *Ranker) scoreMatchedEntries() {
	// Extract entries from current heap (they all matched previous query)
	oldHeap := *r.resultsHeap
	matchedEntries := make([]*entry.PathEntry, len(oldHeap))
	for i, scored := range oldHeap {
		matchedEntries[i] = scored.Entry
	}

	// Clear heap and rebuild with new scores
	r.resultsHeap = &ResultsHeap{}
	heap.Init(r.resultsHeap)

	// Only rescore entries that matched before
	for _, e := range matchedEntries {
		matched, s := score(r.query, e.AbsPath)
		if matched {
			heap.Push(r.resultsHeap, ScoredEntry{Entry: e, Score: s})
		}
	}
}

func (r *Ranker) Results() []ScoredEntry {
	// Return heap as sorted slice (heap is already sorted)
	// Make a copy to avoid modifying the heap
	results := make([]ScoredEntry, len(*r.resultsHeap))
	copy(results, *r.resultsHeap)
	return results
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

	// Fast rejection: subsequence check using two pointers
	q_idx := 0
	for tg_idx := 0; tg_idx < len(targetRunes) && q_idx < len(queryRunes); tg_idx++ {
		if targetRunes[tg_idx] == queryRunes[q_idx] {
			q_idx++
		}
	}
	if q_idx < len(queryRunes) {
		return false, 0
	}

	// FZF v2 algorithm: Dynamic programming to find optimal match positions
	qLen := len(queryRunes)
	tgLen := len(targetRunes)

	const negInf = -100000

	// M[i][j] = best score when query[i-1] matches at target[j-1]
	// H[i][j] = best overall score for query[0..i-1] in target[0..j-1]
	M := make([][]int, qLen+1)
	H := make([][]int, qLen+1)
	for i := range M {
		M[i] = make([]int, tgLen+1)
		H[i] = make([]int, tgLen+1)
		for j := range M[i] {
			M[i][j] = negInf
			H[i][j] = negInf
		}
	}

	// Base case: matching 0 query chars = 0 score
	for j := 0; j <= tgLen; j++ {
		H[0][j] = 0
	}

	// Fill DP table
	for i := 1; i <= qLen; i++ {
		for j := 1; j <= tgLen; j++ {
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

	return true, H[qLen][tgLen]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
