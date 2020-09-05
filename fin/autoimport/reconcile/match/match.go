// package match matches two sorted arrays of int values together: matchTo
// and matchFrom. The values in matchTo and matchFrom are paired such that
// the value in matchFrom is as close as possible to the value in matchTo but
// is never greater than the value in matchTo. Optionally, a maxDiff value
// can be specified such that the difference between paired values is always
// less than maxDiff. Values are paired so that the number of matches
// is maximized.
package match

// match returns an []int that is the same size of matchTo.
// Each element in returned []int is either -1 if the corresponding
// element in matchTo does not match any in matchFrom or is the matching
// index in matchFrom. Each element in matchFrom may be associated with
// at most one element from matchTo. That is the returned array will not
// contain any duplicate indexes. Both matchTo and matchFrom must be
// sorted in ascending order. For index j in matchFrom to match index i in
// matchTo matchTo[i] >= matchFrom[j]. Moreover if maxDiff is positive,
// then matchTo[i] - matchFrom[i] < maxDiff. Caller should pass
// a pointer to a slice that is initally nil for buffer. Returned array
// will always be a slice of buffer (buffer grows as necessary). Returned
// array is only valid until the next call to this method.
func Match(matchTo, matchFrom []int, maxDiff int, buffer *[]int) []int {
	params := matchParams{
		MatchTo: matchTo, MatchFrom: matchFrom, MaxDiff: maxDiff}
	if len(matchTo) > len(*buffer) {
		*buffer = make([]int, len(matchTo))
	}
	result := (*buffer)[:len(matchTo)]
	bottomIdx := -1
	for toIdx := range matchTo {
		matchIdx := params.FindMatch(toIdx)
		if matchIdx == -1 {
			result[toIdx] = matchIdx
			continue
		}
		if params.CanAddMatch(result[:toIdx], bottomIdx, matchIdx) {
			result[toIdx] = matchIdx
		} else {
			result[toIdx] = -1
			bottomIdx = matchIdx
		}
	}
	return result
}

// matchParams represents the parameters passed to the match function
type matchParams struct {
	MatchTo   []int
	MatchFrom []int
	MaxDiff   int
	fromIdx   int
}

// Given an index into p.MatchTo, FindMatch returns the index of the element
// in p.MatchFrom that is closest to the element in matchTo such that
// p.MatchFrom[resultIdx] <= p.MatchTo[toIdx] and if MaxDiff is positive,
// p.MatchTo[toIdx] - p.MatchFrom[resultIdx] < MaxDiff. Returns -1 if there
// is no such matching element in p.MatchFrom. FindMatch must be called with
// ever increasing toIdx values for the lifetime of this instance.
func (p *matchParams) FindMatch(toIdx int) int {
	matchFromLen := len(p.MatchFrom)
	for p.fromIdx < matchFromLen && p.MatchTo[toIdx] >= p.MatchFrom[p.fromIdx] {
		p.fromIdx++
	}
	result := p.fromIdx - 1
	if result < 0 || !p.inRange(toIdx, result) {
		return -1
	}
	return result
}

// CanAddMatch returns true if idxToAdd can be added to the end of
// matchesSoFar. If CanAddMatch returns true it may decrement elements within
// matchesSoFar to make the add possible. If idxToAdd cannot be added then
// CanAddMatch returns false leaving matchesSoFar unchanged. All adjusted
// indexes will be greater than lowerBound (otherwise CanAddMatch returns
// false leaving matchesSoFar unchanged). In addition, CanAddMatch returns
// false if idxToAdd == bottom. -1 <= bottom <= idxToAdd. idxToAdd >= 0 and
// idxToAdd cannot be smaller than last element in matchesSoFar.
func (p *matchParams) CanAddMatch(
	matchesSoFar []int, bottom, idxToAdd int) bool {
	if bottom == idxToAdd {
		return false
	}
	top := idxToAdd
	matchesSoFarLen := len(matchesSoFar)
	idx := matchesSoFarLen - 1
	for ; idx >= 0 && matchesSoFar[idx] == top; idx-- {
		top = matchesSoFar[idx] - 1
		if top == bottom || !p.inRange(idx, top) {
			return false
		}
	}
	for idx2 := idx + 1; idx2 < matchesSoFarLen; idx2++ {
		matchesSoFar[idx2]--
	}
	return true
}

func (p *matchParams) inRange(toIdx, fromIdx int) bool {
	if p.MaxDiff <= 0 {
		return true
	}
	return p.MatchTo[toIdx]-p.MatchFrom[fromIdx] < p.MaxDiff
}
