package fin

import (
  "github.com/keep94/goconsume"
)

// CatPopularity tells the popularity of each category
type CatPopularity interface {

  // Popularity returns the popularity of the category as a value greater than
  // or equal to zero. The higher the return value the more popular the
  // category.
  Popularity(cat Cat) int
}

// BuildCatPopularity returns a consumer that consumes Entry values to
// build a CatPopularity instance. The returned consumer consumes at most
// maxEntriesToRead values with categories other than the top level expense
// category and skips values that have only the top level expense category.
// Caller must call Finalize on returned consumer for the built CatPopularity
// instance to be stored at catPopularity.
func BuildCatPopularity(
    maxEntriesToRead int, 
    catPopularity *CatPopularity) goconsume.ConsumeFinalizer {
  popularities := make(catPopularityMap)
  consumer := goconsume.Slice(popularities, 0, maxEntriesToRead)
  consumer = goconsume.Filter(consumer, nonTrivialCategories)
  return &catPopularityConsumer{
    Consumer: consumer, popularities: popularities, result: catPopularity}
}

type catPopularityMap map[Cat]int

func (c catPopularityMap) Popularity(cat Cat) int {
  return c[cat]
}

func (c catPopularityMap) CanConsume() bool {
  return true
}

func (c catPopularityMap) Consume(ptr interface{}) {
  entry := ptr.(*Entry)
  for _, catrec := range entry.CatRecs() {
    c[catrec.Cat]++
  }
}

func nonTrivialCategories(ptr interface{}) bool {
  entry := ptr.(*Entry)
  if entry.CatRecCount() > 1 {
    return true
  }
  if entry.CatRecCount() == 0 {
    return false
  }
  return entry.CatRecByIndex(0).Cat != Expense
}

type catPopularityConsumer struct {
  goconsume.Consumer
  popularities catPopularityMap
  result *CatPopularity
  finalized bool
}

func (c *catPopularityConsumer) Finalize() {
  if c.finalized {
    return
  }
  c.finalized = true
  c.Consumer = goconsume.Nil()  
  *c.result = c.popularities
}
