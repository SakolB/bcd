package ranker

import (
	"github.com/sakolb/bcd/src/entry"
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
}

func (r *Ranker) Results() []ScoredEntry {
	return nil
}
