package consumers

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/goconsume"
  "testing"
)

func TestAddBalance(t *testing.T) {
  var entryBalances []fin.EntryBalance
  entries := []fin.Entry {
      {CatPayment: makeTotal(-400)},
      {CatPayment: makeTotal(-700)},
  }
  consumer := &AddBalance{
    Balance: 347,
    EntryBalanceConsumer: goconsume.AppendTo(&entryBalances),
  }
  entry := entries[0]
  consumer.Consume(&entry)
  entry = entries[1]
  consumer.Consume(&entry)
  if entryBalances[0].Balance != 347 {
    t.Errorf("Expected 347, got %v", entryBalances[0].Balance)
  }
  if entryBalances[1].Balance != 747 {
    t.Errorf("Expected 747, got %v", entryBalances[1].Balance)
  }
  if consumer.Balance != 1447 {
    t.Errorf("Expected 1447, got %v", consumer.Balance)
  }
}

func TestFromCatPaymentAggregator(t *testing.T) {
  entries := []fin.Entry {
      {CatPayment: makeTotal(400)},
      {CatPayment: makeTotal(700)},
  }
  aggregator := catPaymentTotaler{}
  consumer := FromCatPaymentAggregator(&aggregator)
  entry := entries[0]
  consumer.Consume(&entry)
  entry = entries[1]
  consumer.Consume(&entry)
  if aggregator.total != 1100 {
    t.Errorf("Expected 1100, got %v", aggregator.total)
  }
}

func TestFromEntryAggregator(t *testing.T) {
  entries := []fin.Entry {
      {CatPayment: makeTotal(400)},
      {CatPayment: makeTotal(700)},
  }
  aggregator := entryTotaler{}
  consumer := FromEntryAggregator(&aggregator)
  entry := entries[0]
  consumer.Consume(&entry)
  entry = entries[1]
  consumer.Consume(&entry)
  if aggregator.total != 1100 {
    t.Errorf("Expected 1100, got %v", aggregator.total)
  }
}

func makeTotal(total int64) fin.CatPayment {
  return fin.NewCatPayment(fin.NewCat("0.7"), -total, false, 0)
}

type entryTotaler struct {
  total int64
}

func (e *entryTotaler) Include(entry *fin.Entry) {
  e.total += entry.Total()
}

type catPaymentTotaler struct {
  total int64
}

func (c *catPaymentTotaler) Include(cp *fin.CatPayment) {
  c.total += cp.Total()
}
