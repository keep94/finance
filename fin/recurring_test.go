package fin

import (
  "github.com/keep94/appcommon/date_util"
  "testing"
  "time"
)

func TestInfiniteRecurring(t *testing.T) {
  var r RecurringEntry
  r.Id = 7
  r.Date = date_util.YMD(2015, 7, 20)
  r.NumLeft = -1
  var entries []*Entry
  
  // Shouldn't advance too early
  if output := r.Advance(date_util.YMD(2015, 7, 19), &entries); output {
    t.Error("Did not expect advancement")
  }
  verifyEntries(t, entries)

  // advance 1
  if output := r.Advance(date_util.YMD(2015, 7, 20), &entries); !output {
    t.Error("expected advancement")
  }
  verifyEntries(t, entries, date_util.YMD(2015, 7, 20))

  // We already advanced for this date, shouldn't advance again
  if output := r.Advance(date_util.YMD(2015, 7, 20), &entries); output {
    t.Error("Did not expect advancement")
  }
  verifyEntries(t, entries, date_util.YMD(2015, 7, 20))

  // Advance 2
  if output := r.Advance(date_util.YMD(2015, 9, 20), &entries); !output {
    t.Error("expected advancement")
  }
  verifyEntries(t, entries,
      date_util.YMD(2015, 7, 20),
      date_util.YMD(2015, 8, 20),
      date_util.YMD(2015, 9, 20))

  if output := r.Id; output != 7 {
    t.Error("Expected Id of recurring entry to stay the same")
  }
  if output := r.NumLeft; output != -1 {
    t.Error("Expected num left to stay -1")
  }
}

func TestFiniteRecurringEverySixMonths(t *testing.T) {
  var r RecurringEntry
  r.Id = 12
  r.Date = date_util.YMD(2015, 8, 1)
  r.NumLeft = 3
  r.Period = 6
  var entries []*Entry
  
  // Shouldn't advance too early
  if output := r.Advance(date_util.YMD(2015, 7, 31), &entries); output {
    t.Error("Did not expect advancement")
  }
  verifyEntries(t, entries)

  // advance 2
  if output := r.Advance(date_util.YMD(2016, 2, 1), &entries); !output {
    t.Error("expected advancement")
  }
  verifyEntries(t, entries,
      date_util.YMD(2015, 8, 1),
      date_util.YMD(2016, 2, 1))

  // Advance 2 more but we only advance 1 because we ran out
  if output := r.Advance(date_util.YMD(2017, 2, 1), &entries); !output {
    t.Error("expected advancement")
  }
  verifyEntries(t, entries,
      date_util.YMD(2015, 8, 1),
      date_util.YMD(2016, 2, 1),
      date_util.YMD(2016, 8, 1))

  // Try to advance again but we can't because we are out
  if output := r.Advance(date_util.YMD(2017, 8, 1), &entries); output {
    t.Error("Did not expect advancement")
  }
  verifyEntries(t, entries,
      date_util.YMD(2015, 8, 1),
      date_util.YMD(2016, 2, 1),
      date_util.YMD(2016, 8, 1))

  if output := r.Id; output != 12 {
    t.Error("Expected Id of recurring entry to stay the same")
  }
  if output := r.NumLeft; output != 0 {
    t.Error("Expected num left to be 0")
  }
}

func TestRecurringLastDayOfMonth(t *testing.T) {
  var r RecurringEntry
  r.Id = 9
  r.Date = date_util.YMD(2015, 5, 31)
  r.NumLeft = -1
  var entries []*Entry
  
  // advance 1
  if output := r.Advance(date_util.YMD(2015, 7, 1), &entries); !output {
    t.Error("expected advancement")
  }
  verifyEntries(t, entries,
      date_util.YMD(2015, 5, 31),
      date_util.YMD(2015, 7, 1))
}

func verifyEntries(
    t *testing.T, entries []*Entry, times ...time.Time) {
  if len(entries) != len(times) {
    t.Errorf("Expected %d times, got %d", len(times), len(entries))
    return
  }
  for i := range entries {
    if entries[i].Date != times[i] {
      t.Errorf("Expected %v, got %v", times[i], entries[i].Date)
    }
    if entries[i].Id != 0 {
      t.Error("Expect Id of new entries to be zero.")
    }
  }
}

