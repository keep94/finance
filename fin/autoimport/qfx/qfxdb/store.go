// Package qfxdb provides the data structures for storing which fitIds have
// been processed.
package qfxdb

import (
  "github.com/keep94/appcommon/db"
)

// FitIdSet represents a set of fitids.
type FitIdSet map[string]bool

// Interface Store handles storage and retrieval of fitIds from QFX files.
type Store interface {
  // Add adds a set of fitIds to the store for a particular account Id.
  Add(t db.Transaction, accountId int64, fitIds FitIdSet) error
  
  // Find finds fitIds for a particular account Id and returns them.
  // fitIds is the set of fitIds to look for. The returned set of fitIds will
  // always be a subset of the fitIds parameter or nil if Find cannot find any
  // of the fitIds.
  Find(t db.Transaction, accountId int64, fitIds FitIdSet) (FitIdSet, error)
}
