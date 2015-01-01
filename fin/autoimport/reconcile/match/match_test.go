package match_test

import (
  "github.com/keep94/finance/fin/autoimport/reconcile/match"
  "reflect"
  "testing"
)

func TestMatch(t *testing.T) {
  verifyMatch(
      t,
      []int{5},
      []int{3},
      0,
      []int{0})
  verifyMatch(
      t,
      []int{5},
      []int{3},
      2,
      []int{-1})
  verifyMatch(
      t,
      []int{5},
      []int{4},
      2,
      []int{0})
  verifyMatch(
      t,
      []int{5},
      []int{5},
      2,
      []int{0})
  verifyMatch(
      t,
      []int{5},
      []int{6},
      2,
      []int{-1})
  verifyMatch(
      t,
      []int{5},
      []int{6},
      0,
      []int{-1})
  verifyMatch(
      t,
      []int{30, 30, 30},
      []int{29, 30, 31, 32},
      0,
      []int{0, 1, -1})
  verifyMatch(
      t,
      []int{-2, -1, 1, 2},
      []int{0, 0, 0},
      0,
      []int{-1, -1, 1, 2})
  verifyMatch(
      t,
      nil,
      []int{3, 4, 5},
      0,
      nil)
  verifyMatch(
      t,
      nil,
      []int{3, 4, 5},
      5,
      nil)
  verifyMatch(
      t,
      []int{3, 4, 5},
      nil,
      0,
      []int{-1, -1, -1})
  verifyMatch(
      t,
      []int{3, 4, 5},
      nil,
      5,
      []int{-1, -1, -1})
  verifyMatch(
      t,
      nil,
      nil,
      0,
      nil)
  verifyMatch(
      t,
      nil,
      nil,
      5,
      nil)
  verifyMatch(
      t,
      []int{23, 25, 26, 32, 33, 33, 40, 40, 40},
      []int{25, 25, 26, 26, 34, 34},
      7,
      []int{-1, 1, 2, 3, -1, -1, 4, 5, -1})
  verifyMatch(
      t,
      []int{23, 23, 39, 39},
      []int{10, 11, 17, 18, 25, 26, 32, 33, 40},
      7,
      []int{2, 3, 7, -1})
}

func verifyMatch(
    t *testing.T,
    bank, unrec []int,
    maxDiff int,
    expectedMatches []int) {
  var buffer []int
  actual := match.Match(bank, unrec, maxDiff, &buffer)
  if !reflect.DeepEqual(expectedMatches, actual) {
    t.Errorf("Expected %v, got %v", expectedMatches, actual)
  }
}
