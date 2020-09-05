// Package checks contains routines for processing check numbers.
package checks

import (
	"sort"
)

// Range represents a range of contiguous check numbers
type Range struct {

	// First check number of range inclusive
	First int

	// Last check number of range inclusive
	Last int
}

// RangeWithHoles is used to report missing checks.
type RangeWithHoles struct {

	// The first check number inclusive
	First int

	// The last check number inclusive
	Last int

	// The missing checks in between the first and last check number.
	Holes []Range
}

// Missing takes a sequence of check numbers and reports missing checks.
// Duplicate check numbers in checkNos are treated as if the check number
// appeared once. If checkNos is empty, Missing returns nil.
func Missing(checkNos []int) *RangeWithHoles {
	if len(checkNos) == 0 {
		return nil
	}
	sortedCheckNos := make([]int, len(checkNos))
	copy(sortedCheckNos, checkNos)
	sort.Ints(sortedCheckNos)
	length := len(sortedCheckNos)
	min := sortedCheckNos[0]
	max := sortedCheckNos[length-1]
	var holes []Range
	for i := 1; i < length; i++ {
		if sortedCheckNos[i] == sortedCheckNos[i-1] {
			continue
		}
		if sortedCheckNos[i] == sortedCheckNos[i-1]+1 {
			continue
		}
		holes = append(
			holes,
			Range{First: sortedCheckNos[i-1] + 1, Last: sortedCheckNos[i] - 1})
	}
	return &RangeWithHoles{First: min, Last: max, Holes: holes}
}
