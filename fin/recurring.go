// Package fin declares basic types used in personal finance.
package fin

import (
  "time"
)

// RecurringEntry represents a recurring entry.
type RecurringEntry struct {
  // The entry, the date field in this entry corresponds to the date of the
  // next entry that the Advance method generates.
  Entry

  // The number of months between entries that the Advance method generates
  Period int

  // The number of entries left to generate, a negative number means
  // unlimited.
  NumLeft int
}
  
// Advance advances this recurring entry through the current date adding the
// gnerated entries to appendedNewEntries. After this returns, the date
// of this instance is after currentDate unless the NumLeft field reached
// zero before this instance could be advanced passed currentDate. Returns
// true if Advance generated new entries or false otherwise.
func (r *RecurringEntry) Advance(
    currentDate time.Time, appendedNewEntries *[]*Entry) (advanced bool) {
  for !r.Date.After(currentDate) {
    // Are we out?
    if r.NumLeft == 0 {
      break
    }
    newEntry := r.Entry
    newEntry.Id = 0
    *appendedNewEntries = append(*appendedNewEntries, &newEntry)
    if r.Period > 0 {
      r.Date = r.Date.AddDate(0, r.Period, 0)
    } else {
      r.Date = r.Date.AddDate(0, 1, 0)
    }
    if r.NumLeft > 0 {
      r.NumLeft--
    }
    advanced = true
  }
  return
}
