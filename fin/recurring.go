// Package fin declares basic types used in personal finance.
package fin

import (
  "time"
)

// RecurringUnit represents a unit of time for recurring entries.
// The zero value is equivalent to 'months'.
type RecurringUnit int

const (
  Months RecurringUnit = iota
  Years
  Days
  Weeks
  // Placeholder for unit count. Does not represent an actual unit.
  // New units must be inserted right before this one.
  RecurringUnitCount
)

// ToRecurringUnit takes an int that ToInt returned and converts it back to a
// RecurringUnit. On success, returns the RecurringUnit and true. If x is out
// of range, returns RecurringUnitCount and false.
func ToRecurringUnit(x int) (RecurringUnit, bool) {
  if x < 0 || x >= int(RecurringUnitCount) {
    return RecurringUnitCount, false
  }
  return RecurringUnit(x), true
}

func (r RecurringUnit) String() string {
  switch r {
    case Days:
      return "days"
    case Weeks:
      return "weeks"
    case Months:
      return "months"
    case Years:
      return "years"
    default:
      return "unknown"
  }
}

// ToInt maps a RecurringUnit to an int in a way that is suitable for
// persistent storage.
func (r RecurringUnit) ToInt() int {
  return int(r)
}

// RecurringPeriod represents the time period between recurring entries.
type RecurringPeriod struct {
  // The count of units.
  Count int
  // The time unit.
  Unit RecurringUnit
}

// AddTo returns date + (this instance). If the Count field of this instance
// is < 1, AddTo treats it as 1. AddTo panics if the Unit field of this
// instance is not an actual unit.
func (r RecurringPeriod) AddTo(date time.Time) time.Time {
  count := r.Count
  if count < 1 {
    count = 1
  }
  switch r.Unit {
    case Days:
      return date.AddDate(0, 0, count)
    case Weeks:
      return date.AddDate(0, 0, 7 * count)
    case Months:
      return date.AddDate(0, count, 0)
    case Years:
      return date.AddDate(count, 0, 0)
    default:
      panic("Unit field not a valid RecurringUnit.")
  }
}

// RecurringEntry represents a recurring entry.
type RecurringEntry struct {
  // The entry, the date field in this entry corresponds to the date of the
  // next entry that the Advance method generates.
  Entry

  // The period of time between each generated entries
  Period RecurringPeriod

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
    r.Date = r.Period.AddTo(r.Date)
    if r.NumLeft > 0 {
      r.NumLeft--
    }
    advanced = true
  }
  return
}
