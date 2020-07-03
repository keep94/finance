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
  missing := checks.Missing([]int{1003, 1001, 1007, 1004, 1008, 1009, 1001, 1004})
  assert.Equal(1001, missing.First)
  assert.Equal(1009, missing.Last)
  assert.Equal(missing.Holes, []checks.Range{{1002, 1002}, {1005, 1006}})
}

func TestNoHoles(t *testing.T) {
  assert := assert.New(t)
  missing := checks.Missing([]int{101})
  assert.Equal(101, missing.First)
  assert.Equal(101, missing.Last)
  assert.Equal(0, len(missing.Holes))
}
