package common

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/gofunctional3/functional"
  "net/url"
  "strconv"
  "strings"
  "time"
)

const (
  kMaxSplits = 10
)

var (
  // Error for date being wrong when adding an entry.
  ErrDateMayBeWrong = errors.New("Date may be wrong, proceed anyway?")
  ErrConcurrentModification = errors.New(
      "Someone else already updated this entry. Click cancel and try again.")
)

var (
  entrySplits []EntrySplitType
)

// View for single entry pages.
type SingleEntryView struct {
  http_util.Values
  CatDisplayer
  Splits []EntrySplitType
  Error error
  ExistingEntry bool
}

// DateMayBeWrong returns true if and only if the error for this view is
// ErrDateMayBeWrong.
func (v SingleEntryView) DateMayBeWrong() bool {
  return v.Error == ErrDateMayBeWrong
}

// EntrySplitType represents the display for a single split of an entry.
type EntrySplitType int

// CatParam returns the name of the category parameter for this split
func (s EntrySplitType) CatParam() string {
  return fmt.Sprintf("cat-%d", int(s))
}

// AmountParam returns the name of the amount parameter for this split
func (s EntrySplitType) AmountParam() string {
  return fmt.Sprintf("amount-%d", int(s))
}

// ReconcileParam returns the name of the reconciled parameter for this split
func (s EntrySplitType) ReconcileParam() string {
  return fmt.Sprintf("reconciled-%d", int(s))
}

// ToSingleEntryView creates a view from a particular entry.
// The caller may safely add additional name value pairs to the Values field
// of returned view.
// The caller must refrain from making other types of changes to returned view.
func ToSingleEntryView(
    entry *fin.Entry,
    tag uint64,
    cds categories.CatDetailStore) *SingleEntryView {
  result := &SingleEntryView{
      Values: http_util.Values{make(url.Values)}, CatDisplayer: CatDisplayer{cds}, Splits: entrySplits, Error: nil, ExistingEntry: true}
  result.Set("etag", strconv.FormatUint(tag, 10))
  result.Set("name", entry.Name)
  result.Set("desc", entry.Desc)
  result.Set("checkno", entry.CheckNo)
  result.Set("date", entry.Date.Format(date_util.YMDFormat))
  result.Set("payment", strconv.FormatInt(entry.PaymentId(), 10))
  if entry.Reconciled() {
    result.Set("reconciled", "on")
  }
  if entry.Status != fin.Reviewed {
    result.Set("need_review", "on")
  }
  catrecs := cds.SortedCatRecs(entry.CatRecs())
  for idx, split := range result.Splits {
    if idx < len(catrecs) {
      result.Set(split.CatParam(), catrecs[idx].Id().String())
      result.Set(split.AmountParam(), fin.FormatUSD(catrecs[idx].Amount()))
      if catrecs[idx].Reconciled() {
        result.Set(split.ReconcileParam(), "on")
      }
    }
  }
  return result
}

// ToSingleEntryViewFromForm creates a view from form data.
// existingEntry is true if the form data represents an existing entry or
// false if it represents a brand new entry.
// values are the form values.
// cds is the category detail store.
// err is the error from the form submission or nil if no error.
func ToSingleEntryViewFromForm(
    existingEntry bool,
    values url.Values,
    cds categories.CatDetailStore,
    err error) *SingleEntryView {
  return &SingleEntryView{
      Values: http_util.Values{values},
      CatDisplayer: CatDisplayer{cds},
      Splits: entrySplits,
      Error: err,
      ExistingEntry: existingEntry}
}

// EntryMutation converts the form values from a single entry page into
// a mutation and returns that mutation or an error if the form values were
// invalid.
// Calling Filter on the returned functional.Filterer always returns nil.
func EntryMutation(values url.Values) (mutation functional.Filterer, err error) {
  date, err := time.Parse(date_util.YMDFormat, values.Get("date"))
  if err != nil {
    err = errors.New("Date in wrong format.")
    return
  }
  name := values.Get("name")
  if strings.TrimSpace(name) == "" {
    err = errors.New("Name required.")
    return
  }
  desc := values.Get("desc")
  checkno := values.Get("checkno")
  paymentId, _ := strconv.ParseInt(values.Get("payment"), 10, 64)
  if paymentId == 0 {
    err = errors.New("Missing payment.")
    return
  }
  cpb := fin.CatPaymentBuilder{}
  cpb.SetPaymentId(paymentId).SetReconciled(values.Get("reconciled") != "")
  catrec := fin.CatRec{}
  for _, split := range entrySplits {
    cat := fin.NewCat(values.Get(split.CatParam()))
    amountStr := values.Get(split.AmountParam())
    if amountStr == "" {
      break
    }
    var amount int64
    amount, err = fin.ParseUSD(amountStr)
    if err != nil {
      err = errors.New(fmt.Sprintf("Invalid amount: %s", amountStr))
      return
    }
    catrec = fin.CatRec{C: cat, A: amount, R: values.Get(split.ReconcileParam()) != ""}
    cpb.AddCatRec(&catrec)
  }
  cp := cpb.Build()
  needReview := values.Get("need_review") != ""
  mutation = functional.NewFilterer(func(ptr interface{}) error {
    p := ptr.(*fin.Entry)
    p.Date = date
    p.Name = name
    p.Desc = desc
    p.CheckNo = checkno
    p.CatPayment = cp
    if needReview {
      if p.Status == fin.Reviewed {
        p.Status = fin.NotReviewed
      }
    } else {
      if p.Status != fin.Reviewed {
        p.Status = fin.Reviewed
      }
    }
    return nil
  })
  return
}

func init() {
  entrySplits = make([]EntrySplitType, kMaxSplits)
  for i := range entrySplits {
    entrySplits[i] = EntrySplitType(i)
  }
}
