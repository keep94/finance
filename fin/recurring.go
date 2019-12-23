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
  // Specifies the desired day of the month. Relevant only when Unit is Months. 
  // If <= 0 when Unit is Months, then RecurringPeriod works just like
  // time.AddDate(0, Count, 0)
  DayOfMonth int
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
      if r.DayOfMonth <= 0 {
        return date.AddDate(0, count, 0)
      }
      return addMonths(date, count, r.DayOfMonth)
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

// AdvanceOnce advances this recurring entry exactly once storing the
// generated entry at newEntry.
// newEntry may be nil.
// Returns true if this instance advanced or false if it did not.
// This instance won't advance if NumLeft is already 0.
func (r *RecurringEntry) AdvanceOnce(newEntry *Entry) (advanced bool) {
  // Are we out?
  if r.NumLeft == 0 {
    return false
  }
  if newEntry != nil {
    *newEntry = r.Entry
    newEntry.Id = 0
  }
  r.Date = r.Period.AddTo(r.Date)
  if r.NumLeft > 0 {
    r.NumLeft--
  }
  return true
}
  
// Advance advances this recurring entry through the current date adding the
// generated entries to appendedNewEntries. After this returns, the date
// of this instance is after currentDate unless the NumLeft field reached
// zero before this instance could be advanced passed currentDate. Returns
// true if Advance generated new entries or false otherwise.
func (r *RecurringEntry) Advance(
    currentDate time.Time, appendedNewEntries *[]*Entry) (advanced bool) {
  for !r.Date.After(currentDate) {
    var newEntry Entry
    if !r.AdvanceOnce(&newEntry) {
      break
    }
    *appendedNewEntries = append(*appendedNewEntries, &newEntry)
    advanced = true
  }
  return
}

func withDayOfMonth(date time.Time, dayOfMonth int) time.Time {
  return time.Date(
      date.Year(), date.Month(), dayOfMonth, date.Hour(), date.Minute(),
      date.Second(), date.Nanosecond(), date.Location())
}

func addMonths(date time.Time, months, dayOfMonth int) time.Time {
  // dayOfMonth cannot exceed 31
  if dayOfMonth > 31 {
    dayOfMonth = 31
  }
  firstDayOfOriginalMonth := withDayOfMonth(date, 1)
  firstDayOfNewMonth := firstDayOfOriginalMonth.AddDate(0, months, 0)
  newMonthWithCorrectDayOfMonth := withDayOfMonth(
      firstDayOfNewMonth, dayOfMonth)
  // If our month has too few days, use the last day of the month
  if newMonthWithCorrectDayOfMonth.Month() != firstDayOfNewMonth.Month() {
    return newMonthWithCorrectDayOfMonth.AddDate(
        0, 0, -newMonthWithCorrectDayOfMonth.Day())
  }
  return newMonthWithCorrectDayOfMonth
}
