package fin

import (
  "github.com/keep94/appcommon/date_util"
  "testing"
  "time"
)

func TestRecurringUnit(t *testing.T) {
 if r, ok := ToRecurringUnit(Months.ToInt()); r != Months || !ok {
    t.Error("Round trip failed for Months")
  }
  if r, ok := ToRecurringUnit(Weeks.ToInt()); r != Weeks || !ok {
    t.Error("Round trip failed for Weeks")
  }
  if r, ok := ToRecurringUnit(Days.ToInt()); r != Days || !ok {
    t.Error("Round trip failed for Days")
  }
  if r, ok := ToRecurringUnit(Years.ToInt()); r != Years || !ok {
    t.Error("Round trip failed for Years")
  }
  if r, ok := ToRecurringUnit(-1); r != RecurringUnitCount || ok {
    t.Error("Failure with illegal recurring unit")
  }
  if r, ok := ToRecurringUnit(RecurringUnitCount.ToInt()); r != RecurringUnitCount || ok {
    t.Error("Failure with illegal recurring unit")
  }
  if Years.String() != "years" {
    t.Error("expected 'years'")
  }
  if Months.String() != "months" {
    t.Error("expected 'months'")
  }
  if Weeks.String() != "weeks" {
    t.Error("expected 'weeks'")
  }
  if Days.String() != "days" {
    t.Error("expected 'days'")
  }
  if RecurringUnitCount.String() != "unknown" {
    t.Error("expected 'unknown'")
  }
}

func TestPeriod(t *testing.T) {
  var period RecurringPeriod
  verifyDate(
      t,
      date_util.YMD(2015, 6, 12),
      period.AddTo(date_util.YMD(2015, 5, 12)))
  period.Count = 6
  period.Unit = Days
  verifyDate(
      t,
      date_util.YMD(2015, 5, 19),
      period.AddTo(date_util.YMD(2015, 5, 13)))
  period.Count = 5
  period.Unit = Weeks
  verifyDate(
      t,
      date_util.YMD(2015, 7, 11),
      period.AddTo(date_util.YMD(2015, 6, 6)))
  period.Count = 3
  period.Unit = Months
  verifyDate(
      t,
      date_util.YMD(2015, 10, 4),
      period.AddTo(date_util.YMD(2015, 7, 4)))

  // time.AddDate doesn't always honor the day of month.
  verifyDate(
      t,
      date_util.YMD(2015, 3, 2),
      period.AddTo(date_util.YMD(2014, 11, 30)))

  period.Count = 2
  period.Unit = Years
  verifyDate(
      t,
      date_util.YMD(2017, 8, 20),
      period.AddTo(date_util.YMD(2015, 8, 20)))
}

func TestDayOfMonth(t *testing.T) {
  var period RecurringPeriod
  period.Unit = Months
  period.DayOfMonth = 31
  verifyDate(
      t,
      date_util.YMD(2016, 2, 29),
      period.AddTo(date_util.YMD(2016, 1, 31)))
  verifyDate(
      t,
      date_util.YMD(2016, 3, 31),
      period.AddTo(date_util.YMD(2016, 2, 29)))
  verifyDate(
      t,
      date_util.YMD(2016, 4, 30),
      period.AddTo(date_util.YMD(2016, 3, 31)))
  verifyDate(
      t,
      date_util.YMD(2016, 5, 31),
      period.AddTo(date_util.YMD(2016, 4, 30)))
  verifyDate(
      t,
      date_util.YMD(2016, 8, 31),
      period.AddTo(date_util.YMD(2016, 7, 31)))
  verifyDate(
      t,
      date_util.YMD(2017, 1, 31),
      period.AddTo(date_util.YMD(2016, 12, 31)))

  period.Count = 3
  period.DayOfMonth = 29

  verifyDate(
      t,
      date_util.YMD(2015, 2, 28),
      period.AddTo(date_util.YMD(2014, 11, 1)))
  verifyDate(
      t,
      date_util.YMD(2015, 2, 28),
      period.AddTo(date_util.YMD(2014, 11, 30)))
  verifyDate(
      t,
      date_util.YMD(2015, 5, 29),
      period.AddTo(date_util.YMD(2015, 2, 28)))
}

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
  r.Period.Count = 6
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

func verifyDate(t *testing.T, expected, actual time.Time) {
  if expected != actual {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

