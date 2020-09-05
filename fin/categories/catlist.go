package categories

import (
	"container/heap"

	"github.com/keep94/finance/fin"
)

// MostPopularFirst returns catDetails prepended with the most popular
// categories followed by the least popular categories. MostPopularFirst
// prepends len(catDetails)*mostPopularRatio categories, but will not prepend
// more than maxMostPopularCount categories if maxMostPopularCount is >= 0.
// Categories with a popularity of 0 are never prepended.
// If catPopularity is nil, then MostPopularFirst just returns a copy of
// catDetails.
func MostPopularFirst(
	catDetails []CatDetail,
	catPopularity fin.CatPopularity,
	mostPopularRatio float64,
	maxMostPopularCount int) []CatDetail {
	if mostPopularRatio < 0.0 || mostPopularRatio > 1.0 {
		panic("mostPopularRatio must be between 0.0 and 1.0")
	}
	if catPopularity == nil {

		// Just return a copy of catDetails
		result := make([]CatDetail, 0, len(catDetails))
		result = append(result, catDetails...)
		return result
	}
	mostPopularCount := int(float64(len(catDetails)) * mostPopularRatio)
	if maxMostPopularCount >= 0 && mostPopularCount > maxMostPopularCount {
		mostPopularCount = maxMostPopularCount
	}
	var h catDetailWithCountHeap
	for _, catDetail := range catDetails {
		count := catPopularity.Popularity(catDetail.Id())
		if count > 0 {
			if h.Len() == mostPopularCount {
				pushAndPop(&h, catDetailWithCount{cd: catDetail, count: count})
			} else {
				heap.Push(&h, catDetailWithCount{cd: catDetail, count: count})
			}
		}
	}
	result := make([]CatDetail, h.Len(), h.Len()+len(catDetails))
	for h.Len() > 0 {
		cdWithCount := heap.Pop(&h).(catDetailWithCount)
		result[h.Len()] = cdWithCount.cd
	}
	result = append(result, catDetails...)
	return result
}

func pushAndPop(h heap.Interface, val interface{}) {
	length := h.Len()
	h.Push(val)
	if h.Less(0, length) {
		heap.Fix(h, length)
		heap.Pop(h)
	} else {
		h.Pop()
	}
}

type catDetailWithCount struct {
	cd    CatDetail
	count int
}

type catDetailWithCountHeap []catDetailWithCount

func (h catDetailWithCountHeap) Len() int {
	return len(h)
}

func (h catDetailWithCountHeap) Less(i, j int) bool {
	return h[i].count < h[j].count
}

func (h catDetailWithCountHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *catDetailWithCountHeap) Push(x interface{}) {
	cdWithCount := x.(catDetailWithCount)
	*h = append(*h, cdWithCount)
}

func (h *catDetailWithCountHeap) Pop() interface{} {
	length := len(*h)
	result := (*h)[length-1]
	*h = (*h)[0 : length-1]
	return result
}
