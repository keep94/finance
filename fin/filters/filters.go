// Package filters contains useful search filters.
package filters

import (
	"github.com/keep94/finance/fin"
	"github.com/keep94/goconsume"
	"github.com/keep94/toolbox/str_util"
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
func CompileAdvanceSearchSpec(spec *AdvanceSearchSpec) goconsume.FilterFunc {
	var filters []goconsume.FilterFunc
	if spec.CF != nil {
		filters = append(filters, byCatFilterer(spec.CF))
	}
	if spec.AF != nil {
		filters = append(filters, byAmountFilterer(spec.AF))
	}
	if spec.Name != "" {
		filters = append(filters, byNameFilterer(str_util.Normalize(spec.Name)))
	}
	if spec.Desc != "" {
		filters = append(filters, byDescFilterer(str_util.Normalize(spec.Desc)))
	}
	return goconsume.All(filters...)
}

func byCatFilterer(f fin.CatFilter) goconsume.FilterFunc {
	return func(ptr interface{}) bool {
		p := ptr.(*fin.Entry)
		return p.WithCat(f)
	}
}

func byAmountFilterer(f AmountFilter) goconsume.FilterFunc {
	return func(ptr interface{}) bool {
		p := ptr.(*fin.Entry)
		return f(p.Total())
	}
}

func byNameFilterer(name string) goconsume.FilterFunc {
	return func(ptr interface{}) bool {
		p := ptr.(*fin.Entry)
		return strings.Index(str_util.Normalize(p.Name), name) != -1
	}
}

func byDescFilterer(desc string) goconsume.FilterFunc {
	return func(ptr interface{}) bool {
		p := ptr.(*fin.Entry)
		return strings.Index(str_util.Normalize(p.Desc), desc) != -1
	}
}
