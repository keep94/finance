package checks_test

import (
  "testing"

  "github.com/keep94/finance/fin/checks"
  "github.com/stretchr/testify/assert"
)

func TestNoChecks(t *testing.T) {
  assert := assert.New(t)
  assert.Nil(checks.Missing(nil))
}

func TestMissingChecks(t *testing.T) {
  assert := assert.New(t)
  missing := checks.Missing([]int{1003, 1001, 1007, 1004, 1009, 1001, 1004})
  assert.Equal(1001, missing.First)
  assert.Equal(1009, missing.Last)
  assert.Equal(
      []checks.Range{{1002, 1002}, {1005, 1006}, {1008, 1008}}, missing.Holes)
}

func TestSingle(t *testing.T) {
  assert := assert.New(t)
  missing := checks.Missing([]int{101})
  assert.Equal(101, missing.First)
  assert.Equal(101, missing.Last)
  assert.Empty(missing.Holes)
}

func TestOneBigHole(t *testing.T) {
  assert := assert.New(t)
  missing := checks.Missing([]int{3000, 2000})
  assert.Equal(2000, missing.First)
  assert.Equal(3000, missing.Last)
  assert.Equal([]checks.Range{{2001, 2999}}, missing.Holes)
}

func TestNoHoles(t *testing.T) {
  assert := assert.New(t)
  missing := checks.Missing([]int{101, 102, 103, 104})
  assert.Equal(101, missing.First)
  assert.Equal(104, missing.Last)
  assert.Empty(missing.Holes)
}
