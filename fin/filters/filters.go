// Package filters contains useful search filters.
package filters

import (
  "github.com/keep94/appcommon/str_util"
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/functional"
  "strings"
)

// AmountFilter filters by amount. Returns true if amt should be included or
// false otherwise.
type AmountFilter func(amt int64) bool

// AdvanceSearchSpec specifies what entries to search for.
// searches ignore case and whitespace.
type AdvanceSearchSpec struct {
  Name string
  Desc string
  // If present, include only entries with line items that match CF.
  CF fin.CatFilter
  // If present, include only entries whose total matches AF.
  AF AmountFilter
}

// CompileAdvanceSearchSpec compiles a search specification into a
// functional.Filterer object.
func CompileAdvanceSearchSpec(spec *AdvanceSearchSpec) functional.Filterer {
  filters := []functional.Filterer{}
  if spec.CF != nil {
    filters = append(filters, byCatFilterer{spec.CF})
  }
  if spec.AF != nil {
    filters = append(filters, byAmountFilterer{spec.AF})
  }
  if spec.Name != "" {
    filters = append(filters, byNameFilterer{str_util.Normalize(spec.Name)})
  }
  if spec.Desc != "" {
    filters = append(filters, byDescFilterer{str_util.Normalize(spec.Desc)})
  }
  return functional.All(filters...)
}

type byCatFilterer struct {
  f fin.CatFilter
}

func (cf byCatFilterer) Filter(ptr interface{}) error {
  p := ptr.(*fin.Entry)
  if !p.WithCat(cf.f) {
    return functional.Skipped
  }
  return nil
}

type byAmountFilterer struct {
  f AmountFilter
}

func (af byAmountFilterer) Filter(ptr interface{}) error {
  p := ptr.(*fin.Entry)
  if !af.f(p.Total()) {
    return functional.Skipped
  }
  return nil
}

type byNameFilterer struct {
  name string
}

func (nf byNameFilterer) Filter(ptr interface{}) error {
  p := ptr.(*fin.Entry)
  if strings.Index(str_util.Normalize(p.Name), nf.name) == -1 {
    return functional.Skipped
  }
  return nil
}

type byDescFilterer struct {
  desc string
}

func (df byDescFilterer) Filter(ptr interface{}) error {
  p := ptr.(*fin.Entry)
  if strings.Index(str_util.Normalize(p.Desc), df.desc) == -1 {
    return functional.Skipped
  }
  return nil
}

