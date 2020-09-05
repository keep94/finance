package aggregators

import (
	"github.com/keep94/appcommon/str_util"
	"github.com/keep94/finance/fin"
	"regexp"
)

var (
	kPattern = regexp.MustCompile(`[0-9#]{3,}`)
)

// Interface Categorizer assigns categories to entries just added from
// a bank statement.
type Categorizer interface {
	// Assigns a category to entry modifying it in-place.
	// If it can't pick a category for entry, it sets the category to
	// fin.Expense. Returns true if it set the category or false otherwise.
	Categorize(entry *fin.Entry) bool
}

// ByNameCategorizerBuilder builds a Categorizer that assigns categories
// based on the name in the entry. Feed a ByNameCategorizerBuilder entries
// from most recent to least recent to train it to create a Categorizer.
type ByNameCategorizerBuilder struct {
	trainingData map[string]*nameData
	n            int
	k            int
}

// The created object looks at the latest n entries with a similar name. If
// k of those entries have the same category then new entries with similar
// name get assigned that category. NewByNameCategorizerBuilder panics if
// n < k or k < 1.
func NewByNameCategorizerBuilder(n, k int) *ByNameCategorizerBuilder {
	if n < k || k < 1 {
		panic("n must be >= k and k >= 1")
	}
	return &ByNameCategorizerBuilder{make(map[string]*nameData), n, k}
}

// Include trains this instance with a particular entry.
func (b *ByNameCategorizerBuilder) Include(entry *fin.Entry) {
	normalizedName := categorizerNormalizeName(entry.Name)
	data := b.trainingData[normalizedName]
	cat := extractSingleCat(&entry.CatPayment)
	if data == nil {
		if b.k > 1 {
			data = &nameData{map[fin.Cat]int{cat: 1}, fin.Expense, 1}
		} else {
			data = &nameData{nil, cat, 1}
		}
		b.trainingData[normalizedName] = data
	} else {
		data.add(cat, b.n, b.k)
	}
}

// Build returns a Categorizer based on Entries it has observed so far.
func (b *ByNameCategorizerBuilder) Build() Categorizer {
	byNameMap := make(map[string]fin.Cat)
	for k, v := range b.trainingData {
		if v.cat != fin.Expense {
			byNameMap[k] = v.cat
		}
	}
	return byNameCategorizer(byNameMap)
}

type nameData struct {
	catFrequency map[fin.Cat]int
	cat          fin.Cat
	sampleSize   int
}

func (d *nameData) add(cat fin.Cat, n, k int) {
	if d.catFrequency == nil {
		return
	}
	frequency := d.catFrequency[cat] + 1
	if frequency == k {
		d.cat = cat
		d.catFrequency = nil
		return
	}
	d.catFrequency[cat] = frequency
	d.sampleSize++
	if d.sampleSize == n {
		d.catFrequency = nil
	}
}

type byNameCategorizer map[string]fin.Cat

func (b byNameCategorizer) Categorize(entry *fin.Entry) bool {
	return entry.SetSingleCat(b[categorizerNormalizeName(entry.Name)])
}

func categorizerNormalizeName(name string) string {
	return str_util.Normalize(kPattern.ReplaceAllString(name, ""))
}

func extractSingleCat(cp *fin.CatPayment) fin.Cat {
	if cp.CatRecCount() != 1 {
		return fin.Expense
	}
	return cp.CatRecByIndex(0).Cat
}
