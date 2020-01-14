// Package streams contains useful consumers of basic types
package consumers

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/goconsume"
)

// CatPaymentAggregator aggregates CatPayment values.
type CatPaymentAggregator interface {
  Include(cp *fin.CatPayment)
}

// EntryAggregator aggregates Entry values.
type EntryAggregator interface {
  Include(entry *fin.Entry)
}

// FromCatPaymentAggregator converts a CatPaymentAggregator to a Consumer of
// fin.Entry values.
func FromCatPaymentAggregator(
    aggregator CatPaymentAggregator) goconsume.Consumer {
  return entryAggregatorConsumer{aggregator: catPaymentToEntryAggregator{aggregator}}
}

// FromEntryAggregator converts a EntryAggregator to a Consumer of
// fin.Entry values.
func FromEntryAggregator(
    aggregator EntryAggregator) goconsume.Consumer {
  return entryAggregatorConsumer{aggregator: aggregator}
}

// AddBalance is a consumer of Entry values that passes on EntryBalance
// values. AddBalance consumes Entry values newest to oldest passing on
// each Entry value along with the current balance.
type AddBalance struct {
  // ending balance
  Balance int64
  // EntryBalance values passed on here
  EntryBalanceConsumer goconsume.Consumer
  entryBalance fin.EntryBalance
}

func (c *AddBalance) CanConsume() bool {
  return c.EntryBalanceConsumer.CanConsume()
}

func (c *AddBalance) Consume(ptr interface{}) {
  p := ptr.(*fin.Entry)
  c.entryBalance.Entry = *p
  c.entryBalance.Balance = c.Balance
  c.Balance -= p.Total()
  c.EntryBalanceConsumer.Consume(&c.entryBalance)
}

type entryAggregatorConsumer struct {
  aggregator EntryAggregator
}

func (e entryAggregatorConsumer) Consume(ptr interface{}) {
  entry := ptr.(*fin.Entry)
  e.aggregator.Include(entry)
}

func (e entryAggregatorConsumer) CanConsume() bool {
  return true
}

type catPaymentToEntryAggregator struct {
  cpa CatPaymentAggregator
}

func (c catPaymentToEntryAggregator) Include(entry *fin.Entry) {
  c.cpa.Include(&entry.CatPayment)
}
