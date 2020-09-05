package common_test

import (
	"fmt"
	"github.com/keep94/finance/apps/ledger/common"
)

func ExampleNormalizeYMDStr() {
	fmt.Printf("%s %s %s\n",
		common.NormalizeYMDStr("2012"),
		common.NormalizeYMDStr("201203"),
		common.NormalizeYMDStr("20120305"))
	// Output: 20120101 20120301 20120305
}
