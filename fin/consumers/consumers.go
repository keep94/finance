// Package streams contains useful consumers of basic types
package consumers

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
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
    aggregator CatPaymentAggregator) functional.Consumer {
  return entryAggregatorConsumer{aggregator: catPaymentToEntryAggregator{aggregator}}
}

// FromEntryAggregator converts a EntryAggregator to a Consumer of
// fin.Entry values.
func FromEntryAggregator(
    aggregator EntryAggregator) functional.Consumer {
  return entryAggregatorConsumer{aggregator: aggregator}
}

// Compose creates a new Consumer of fin.Entry values out of each Consumer in
// consumers.
func Compose(consumers ...functional.Consumer) functional.Consumer {
  return functional.CompositeConsumer(&fin.Entry{}, nil, consumers...)
}

// EntryBuffer stores fin.Entry instances fetched from database
type EntryBuffer struct {
  *consume.Buffer
}

// NewEntryBuffer creates an EntryBuffer that can store up to capacity fin.Entry
// instances.
func NewEntryBuffer(capacity int) EntryBuffer {
  return EntryBuffer{consume.NewBuffer(make([]fin.Entry, capacity))}
}

// Entries returns the entries gathered from last database fetch.
// Returned array valid until next call to Consume.
func (e EntryBuffer) Entries() []fin.Entry {
  return e.Values().([]fin.Entry)
}

// EntryPageBuffer is used to fetch a specific fixed-length page of fin.Entries.
type EntryPageBuffer struct {
  *consume.PageBuffer
}

// NewEntryPageBuffer creates an EntryPageBuffer.
// pageSize is the number of fin.Entries in each page; desiredPageNo is
// the 0-based desired page number.
func NewEntryPageBuffer(pageSize int, desiredPageNo int) EntryPageBuffer {
  return EntryPageBuffer{consume.NewPageBuffer(make([]fin.Entry, 2 * pageSize), desiredPageNo)}
}

// Entries returns the entries in the fetched page.
// Returned array valid until next call to Consume.
func (e EntryPageBuffer) Entries() []fin.Entry {
  return e.Values().([]fin.Entry)
}

// EntryBalanceBuffer stores fin.EntryBalance instances fetched from database
type EntryBalanceBuffer struct {
  *consume.Buffer
}

// NewEntryBalanceBuffer creates an EntryBalanceBuffer that can store up
// to capacity fin.EntryBalance instances.
func NewEntryBalanceBuffer(capacity int) EntryBalanceBuffer {
  return EntryBalanceBuffer{consume.NewBuffer(make([]fin.EntryBalance, capacity))}
}

// EntriesWithBalance returns the entries with balance
// gathered from last database fetch.
// Returned array valid until next call to Consume.
func (e EntryBalanceBuffer) EntriesWithBalance() []fin.EntryBalance {
  return e.Values().([]fin.EntryBalance)
}

// EntryBalancePageBuffer is used to fetch a specific fixed-length page
// of fin.EntryBalance values.
type EntryBalancePageBuffer struct {
  *consume.PageBuffer
}

// NewEntryBalancePageBuffer creates an EntryBalancePageBuffer.
// pageSize is the number of fin.EntryBalance values  in each page;
// desiredPageNo is the 0-based desired page number.
func NewEntryBalancePageBuffer(pageSize int, desiredPageNo int) EntryBalancePageBuffer {
  return EntryBalancePageBuffer{consume.NewPageBuffer(make([]fin.EntryBalance, 2 * pageSize), desiredPageNo)}
}

// EntriesWithBalance returns the entries with balance in the fetched page.
// Returned array valid until next call to Consume.
func (e EntryBalancePageBuffer) EntriesWithBalance() []fin.EntryBalance {
  return e.Values().([]fin.EntryBalance)
}

type entryAggregatorConsumer struct {
  aggregator EntryAggregator
}

func (e entryAggregatorConsumer) Consume(s functional.Stream) (err error) {
  var entry fin.Entry
  for err = s.Next(&entry); err == nil; err = s.Next(&entry) {
    e.aggregator.Include(&entry)
  }
  if err == functional.Done {
    err = nil
  }
  return
}

type catPaymentToEntryAggregator struct {
  cpa CatPaymentAggregator
}

func (c catPaymentToEntryAggregator) Include(entry *fin.Entry) {
  c.cpa.Include(&entry.CatPayment)
}
