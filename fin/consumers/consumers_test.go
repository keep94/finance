package consumers

import (
  "errors"
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/functional"
  "testing"
)

var (
  someError = errors.New("consumers: some error.")
)

func TestFromCatPaymentAggregator(t *testing.T) {
  entries := []fin.Entry {
      {CatPayment: makeTotal(400)},
      {CatPayment: makeTotal(700)},
  }
  aggregator := catPaymentTotaler{}
  consumer := FromCatPaymentAggregator(&aggregator)
  err := consumer.Consume(functional.NewStreamFromValues(entries, nil))
  if err != nil {
    t.Errorf("Expected no error, got %v", err)
  }
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
  err := consumer.Consume(functional.NewStreamFromValues(entries, nil))
  if err != nil {
    t.Errorf("Expected no error, got %v", err)
  }
  if aggregator.total != 1100 {
    t.Errorf("Expected 1100, got %v", aggregator.total)
  }
}

func TestFromCatPaymentAggregatorError(t *testing.T) {
  aggregator := catPaymentTotaler{}
  consumer := FromCatPaymentAggregator(&aggregator)
  es := errorStream{err: someError}
  err := consumer.Consume(es)
  if err != someError {
    t.Errorf("Expected someError, got %v", err)
  }
}

func TestFromEntryAggregatorError(t *testing.T) {
  aggregator := entryTotaler{}
  consumer := FromEntryAggregator(&aggregator)
  es := errorStream{err: someError}
  err := consumer.Consume(es)
  if err != someError {
    t.Errorf("Expected someError, got %v", err)
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

type errorStream struct {
  err error
}

func (e errorStream) Next(ptr interface{}) error {
  return e.err
}

func (e errorStream) Close() error {
  return nil
}
