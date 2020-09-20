// Package reconcile provides functionality for reconciling entries imported
// from a bank with existing entries that have not yet been reconciled.
package reconcile

import (
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/autoimport/reconcile/match"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/toolbox/date_util"
	"sort"
	"time"
)

var (
	kY2k = date_util.YMD(2000, 1, 1)
)

// AmountCheckNo is a key consisting of amount and check number. To be
// reconciled, entries must be organized by amount and check number.
type AmountCheckNo struct {
	Amount  int64
	CheckNo string
}

// The entries organized by amount and check number. Under each key, the
// entries are sorted by date in descending order. An empty instance of
// this type can be used as an aggregator. See the aggregators package.
// Note that methods of this type change the Id field of the fin.Entry
// values in place through the pointers.
type ByAmountCheckNo map[AmountCheckNo][]*fin.Entry

// New creates a new ByAmountCheckNo from existing entries.
// Note that the returned instance has methods that change the Id field
// of the fin.Entry structures of entries in place through the pointers
// as no defensive copying is done.
func New(entries []*fin.Entry) ByAmountCheckNo {
	sortedEntries := make([]*fin.Entry, len(entries))
	copy(sortedEntries, entries)
	sort.Sort(byDateDesc(sortedEntries))
	result := make(ByAmountCheckNo)
	for _, v := range sortedEntries {
		result.includePtr(v)
	}
	return result
}

func (b ByAmountCheckNo) Include(e *fin.Entry) {
	eCopy := *e
	b.includePtr(&eCopy)
}

// Reconcile reconciles the entries from the bank in this instance with the
// the existing, unreconciled entries in unreconciled. When Reconcile
// returns, the Id field of each entry in this instance matches the ID
// field of the entry it reconciles with in unreconciled. If an entry in
// this instance does not reconcile with any entry, then its ID field is set
// to zero. maxDays is the maximum days allowed between entries reconciled
// together that lack a check number.
func (b ByAmountCheckNo) Reconcile(unreconciled ByAmountCheckNo, maxDays int) {
	var bankIntArray []int
	var unrecIntArray []int
	var matchesIntArray []int
	for k, v := range b {
		if k.CheckNo != "" {
			reconcile(
				v, unreconciled[k], -1,
				&bankIntArray, &unrecIntArray, &matchesIntArray)
		} else {
			reconcile(
				v, unreconciled[k], maxDays,
				&bankIntArray, &unrecIntArray, &matchesIntArray)
		}
	}
}

func (b ByAmountCheckNo) includePtr(e *fin.Entry) {
	acn := AmountCheckNo{e.Total(), e.CheckNo}
	b[acn] = append(b[acn], e)
}

// GetChanges returns the changes needed to add / reconcile the entries from
// the bank. reconciled are the entries from the bank that have been
// reconciled. That is, the bank entries in reconciled that match an existing
// entry in the datastore will have a non-zero Id field
func GetChanges(reconciled []*fin.Entry) *findb.EntryChanges {
	entries := make([]*fin.Entry, len(reconciled))
	newIdx := 0
	existingIdx := len(reconciled) - 1
	for _, v := range reconciled {
		if v.Id == 0 {
			entries[newIdx] = v
			newIdx++
		} else {
			entries[existingIdx] = v
			existingIdx--
		}
	}
	updates := make(
		map[int64]fin.EntryUpdater, len(entries)-1-existingIdx)
	for idx := len(entries) - 1; idx > existingIdx; idx-- {
		updates[entries[idx].Id] = reconciler(entries[idx])
	}
	return &findb.EntryChanges{Adds: entries[:newIdx], Updates: updates}
}

type byDateDesc []*fin.Entry

func (b byDateDesc) Len() int {
	return len(b)
}

func (b byDateDesc) Less(i, j int) bool {
	return b[i].Date.After(b[j].Date)
}

func (b byDateDesc) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func reconciler(f *fin.Entry) fin.EntryUpdater {
	return func(p *fin.Entry) bool {
		if p.Status != fin.Reviewed {
			p.Name = f.Name
			if p.CatRecCount() == 1 && p.CatRecByIndex(0).Cat == fin.Expense {
				p.CatPayment = f.CatPayment
			} else {
				p.Reconcile(f.PaymentId())
			}
		} else {
			p.Reconcile(f.PaymentId())
		}
		return true
	}
}

func reconcile(
	bank, unreconciled []*fin.Entry, maxDays int,
	bankIntArray, unrecIntArray, matchesIntArray *[]int) {
	bankDates := toAscendingIntArray(bank, bankIntArray)
	unrecDates := toAscendingIntArray(unreconciled, unrecIntArray)
	// maxDays is exclusive in match package
	matches := match.Match(
		bankDates, unrecDates, maxDays+1, matchesIntArray)
	pairBankEntries(bank, unreconciled, matches)
}

func dayDiff(end, start time.Time) int {
	return int(end.Sub(start) / (24 * time.Hour))
}

func toAscendingIntArray(
	entriesByDateDesc []*fin.Entry, buffer *[]int) []int {
	if len(entriesByDateDesc) > len(*buffer) {
		*buffer = make([]int, len(entriesByDateDesc))
	}
	result := (*buffer)[:len(entriesByDateDesc)]
	revIdx := len(entriesByDateDesc) - 1
	for _, entry := range entriesByDateDesc {
		result[revIdx] = dayDiff(entry.Date, kY2k)
		revIdx--
	}
	return result
}

func pairBankEntries(bank, unreconciled []*fin.Entry, matches []int) {
	revIdx := len(matches) - 1
	unrecLen := len(unreconciled)
	for _, entry := range bank {
		match := matches[revIdx]
		revIdx--
		if match == -1 {
			entry.Id = 0
		} else {
			entry.Id = unreconciled[unrecLen-match-1].Id
		}
	}
}
