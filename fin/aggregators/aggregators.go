// Package aggregators contains aggregators of fin.Entry and fin.CatPayment
// values. Each aggregator has an Include method and can be easily converted
// to a consumer via the appropriate method in the consumers package.
package aggregators

import (
	"github.com/keep94/appcommon/date_util"
	"github.com/keep94/appcommon/str_util"
	"github.com/keep94/finance/fin"
	"time"
)

// Totaler sums up the total in each fin.CatPayment instance.
type Totaler struct {
	// Total is the total so far
	Total int64
}

func (t *Totaler) Include(cp *fin.CatPayment) {
	t.Total += cp.Total()
}

// AutoCompleteAggregator makes list of auto complete candidates.
type AutoCompleteAggregator struct {
	str_util.AutoComplete
	// The field on which to find auto complete candidates
	Field func(e *fin.Entry) string
}

func (a *AutoCompleteAggregator) Include(e *fin.Entry) {
	a.Add(a.Field(e))
}

// Recurring is the interface for recurring time periods. e.g monthly,
// yearly.
type Recurring interface {
	// Normalize returns the beginning of a time period for a given time
	Normalize(t time.Time) time.Time
	// Add returns the result of adding numPeriods time periods to a start time
	Add(t time.Time, numPeriods int) time.Time
}

func Monthly() Recurring {
	return monthly{}
}

func Yearly() Recurring {
	return yearly{}
}

// PeriodTotal contains the total of all transactions for a given period.
type PeriodTotal struct {
	// The start of the period
	PeriodStart time.Time
	// The actual start inclusive. May differ from PeriodStart if this record
	// covers a partial period
	Start time.Time
	// The end exclusive. May differ from start of next period if this record
	// covers a partial period.
	End time.Time
	// The total for the period.
	Total int64
}

// ByPeriodTotaler sums totals by period
type ByPeriodTotaler struct {
	start     time.Time
	end       time.Time
	recurring Recurring
	totals    map[time.Time]int64
}

// NewByPeriodTotaler creates a new ByPeriodTotaler that collects statistics
// for transactions happening between start inclusive and end exclusive.
// The recurring perameter indicates the recurring period such as monthly or
// yearly.
func NewByPeriodTotaler(start, end time.Time, recurring Recurring) *ByPeriodTotaler {
	return &ByPeriodTotaler{
		date_util.TimeToDate(start),
		date_util.TimeToDate(end),
		recurring,
		make(map[time.Time]int64)}
}

func (b *ByPeriodTotaler) Include(entry *fin.Entry) {
	if entry.Date.Before(b.start) || !b.end.After(entry.Date) {
		return
	}
	b.totals[b.recurring.Normalize(entry.Date)] += entry.Total()
}

// Iterator is used to get the totals by period.
func (b *ByPeriodTotaler) Iterator() *PeriodTotalIterator {
	return &PeriodTotalIterator{b, b.recurring.Normalize(b.start), 0}
}

// PeriodTotalIterator iterates over period totals.
type PeriodTotalIterator struct {
	totaler     *ByPeriodTotaler
	firstPeriod time.Time
	idx         int
}

// Next stores the next period total at p and returns true. If there
// is no next period total, Next returns false.
func (pti *PeriodTotalIterator) Next(p *PeriodTotal) bool {
	periodStart := pti.totaler.recurring.Add(pti.firstPeriod, pti.idx)
	total := pti.totaler.totals[periodStart]
	start := periodStart
	if start.Before(pti.totaler.start) {
		start = pti.totaler.start
	}
	end := pti.totaler.recurring.Add(pti.firstPeriod, pti.idx+1)
	if end.After(pti.totaler.end) {
		end = pti.totaler.end
	}
	if end.After(start) {
		pti.idx++
		*p = PeriodTotal{
			PeriodStart: periodStart,
			Start:       start,
			End:         end,
			Total:       total}
		return true
	}
	return false
}

type monthly struct{}

func (m monthly) Normalize(t time.Time) time.Time {
	return date_util.YMD(t.Year(), int(t.Month()), 1)
}

func (m monthly) Add(t time.Time, numPeriods int) time.Time {
	return t.AddDate(0, numPeriods, 0)
}

type yearly struct{}

func (y yearly) Normalize(t time.Time) time.Time {
	return date_util.YMD(t.Year(), 1, 1)
}

func (y yearly) Add(t time.Time, numPeriods int) time.Time {
	return t.AddDate(numPeriods, 0, 0)
}
