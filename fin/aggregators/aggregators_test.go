package aggregators

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/finance/fin"
  "testing"
  "time"
)

func TestByPeriodZeroTime(t *testing.T) {
  bpt := NewByPeriodTotaler(
      date_util.YMD(2013, 1, 15), date_util.YMD(2013, 1, 15), Monthly())
  var pt PeriodTotal
  if bpt.Iterator().Next(&pt) {
    t.Error("Did not expect totals")
  }
}

func TestByPeriodSamePeriod(t *testing.T) {
  bpt := NewByPeriodTotaler(
      date_util.YMD(2013, 1, 15), date_util.YMD(2013, 1, 17), Monthly())
  aggregate(
      date_util.YMD(2013, 1, 1),
      date_util.YMD(2013, 1, 31),
      bpt)
  expected := []PeriodTotal{
      {
          PeriodStart: date_util.YMD(2013, 1, 1),
          Start: date_util.YMD(2013, 1, 15),
          End: date_util.YMD(2013, 1, 17),
          Total: -31}}
  verify(t, expected, bpt)
}

func TestByPeriodAdjacentPeriods(t *testing.T) {
  bpt := NewByPeriodTotaler(
      date_util.YMD(2013, 1, 31), date_util.YMD(2013, 2, 2), Monthly())
  aggregate(
      date_util.YMD(2013, 1, 1),
      date_util.YMD(2013, 3, 1),
      bpt)
  expected := []PeriodTotal{
      {
          PeriodStart: date_util.YMD(2013, 1, 1),
          Start: date_util.YMD(2013, 1, 31),
          End: date_util.YMD(2013, 2, 1),
          Total: -31},
      {
          PeriodStart: date_util.YMD(2013, 2, 1),
          Start: date_util.YMD(2013, 2, 1),
          End: date_util.YMD(2013, 2, 2),
          Total: -32}}
  verify(t, expected, bpt)
}

func TestByPeriodSingleFullPeriod(t *testing.T) {
  bpt := NewByPeriodTotaler(
      date_util.YMD(2013, 1, 1), date_util.YMD(2013, 2, 1), Monthly())
  aggregate(
      date_util.YMD(2013, 1, 1),
      date_util.YMD(2013, 3, 1),
      bpt)
  expected := []PeriodTotal{
      {
          PeriodStart: date_util.YMD(2013, 1, 1),
          Start: date_util.YMD(2013, 1, 1),
          End: date_util.YMD(2013, 2, 1),
          Total: -496}}
  verify(t, expected, bpt)
}

func TestByPeriodYearly(t *testing.T) {
  bpt := NewByPeriodTotaler(
      date_util.YMD(2012, 12, 31), date_util.YMD(2014, 1, 1), Yearly())
  aggregate(
      date_util.YMD(2013, 1, 1),
      date_util.YMD(2013, 2, 2),
      bpt)
  expected := []PeriodTotal{
      {
          PeriodStart: date_util.YMD(2012, 1, 1),
          Start: date_util.YMD(2012, 12, 31),
          End: date_util.YMD(2013, 1, 1),
          Total: 0},
      {
          PeriodStart: date_util.YMD(2013, 1, 1),
          Start: date_util.YMD(2013, 1, 1),
          End: date_util.YMD(2014, 1, 1),
          Total: -528}}
  verify(t, expected, bpt)
}

func aggregate(start, end time.Time, bpt *ByPeriodTotaler) {
  entry := fin.Entry{}
  var amount int64 = 1
  for date := start; date.Before(end); date = date.AddDate(0, 0, 1) {
    entry.Date = date
    entry.CatPayment = fin.NewCatPayment(fin.Expense, amount, false, 0)
    bpt.Include(&entry)
    amount++
  }
}

func verify(t *testing.T, expected []PeriodTotal, bpt *ByPeriodTotaler) {
  iter := bpt.Iterator()
  var actual PeriodTotal
  for idx := range expected {
    if !iter.Next(&actual) {
      t.Error("Expected to get a PeriodTotal.")
      return
    }
    if expected[idx] != actual {
      t.Errorf("Expected %v, got %v", &expected[idx], &actual)
    }
  }
  if iter.Next(&actual) {
    t.Error("Didn't expect a PeriodTotal.")
  }
}
