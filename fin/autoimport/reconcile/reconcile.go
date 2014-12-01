// Package reconcile provides functionality for reconciling entries imported
// from a bank with existing entries that have not yet been reconciled.
package reconcile

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/functional"
  "sort"
  "time"
)

// AmountCheckNo is a key consisting of amount and check number. To be
// reconciled, entries must be organized by amount and check number.
type AmountCheckNo struct {
  Amount int64
  CheckNo string
}

// The entries organized by amount and check number. Under each key, the
// entries are sorted by date in descending order. An empty instance of
// this type can be used as an aggregator. See the aggregators package.
type ByAmountCheckNo map[AmountCheckNo][]*fin.Entry

// New creates a new ByAmountCheckNo from existing entries.
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
  for k, v := range b {
    if k.CheckNo != "" {
      reconcile(v, unreconciled[k], -1)
    } else {
      reconcile(v, unreconciled[k], maxDays)
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
  updates := make(map[int64]functional.Filterer, len(entries) - 1 - existingIdx)
  for idx := len(entries) - 1; idx > existingIdx; idx-- {
    updates[entries[idx].Id] = filterer{entries[idx]}
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

type filterer struct {
  *fin.Entry
}

func (f filterer) Filter(ptr interface{}) error {
  p := ptr.(*fin.Entry)
  if p.Status != fin.Reviewed {
    p.Name = f.Name
    if p.CatRecCount() == 1 && p.CatRecByIndex(0).Id() == fin.Expense {
      p.CatPayment = f.CatPayment
    } else {
      p.Reconcile(f.PaymentId())
    }
  } else {
    p.Reconcile(f.PaymentId())
  }
  return nil
}

func reconcile(bank, unreconciled []*fin.Entry, maxDays int) {
  for _, v := range bank {
    v.Id = 0
  }
  bankIdx := 0
  unrecIdx := 0
  for bankIdx < len(bank) && unrecIdx < len(unreconciled) {
    if bank[bankIdx].Date.Before(unreconciled[unrecIdx].Date) {
      unrecIdx++
    } else if maxDays < 0 || dayDiff(bank[bankIdx].Date, unreconciled[unrecIdx].Date) <= maxDays {
      bank[bankIdx].Id = int64(unrecIdx + 1)
      bankIdx++
      unrecIdx++
    } else {
      bankIdx++
    }
  }
  bankIdx = len(bank) - 1
  unrecIdx = len(unreconciled) - 1
  for bankIdx >= 0 && unrecIdx >= 0 {
    if bank[bankIdx].Id != 0 {
      if bank[bankIdx].Id <= int64(unrecIdx + 1) {
        unrecIdx = int(bank[bankIdx].Id - 2)
        bankIdx--
        continue
      } else {
        bank[bankIdx].Id = 0
      }
    }
    if bank[bankIdx].Date.Before(unreconciled[unrecIdx].Date) {
      bankIdx--
    } else if maxDays < 0 || dayDiff(bank[bankIdx].Date, unreconciled[unrecIdx].Date) <= maxDays {
      bank[bankIdx].Id = int64(unrecIdx + 1)
      bankIdx--
      unrecIdx--
    } else {
      unrecIdx--
    }
  }
  for ;bankIdx >= 0; bankIdx-- {
    if bank[bankIdx].Id != 0 {
      bank[bankIdx].Id = 0
    }
  }
  for _, v := range bank {
    if v.Id != 0 {
      v.Id = unreconciled[v.Id -1].Id
    }
  }
}

func dayDiff(end, start time.Time) int {
  return int(end.Sub(start) / (24 * time.Hour))
}
