package categories

import (
  "github.com/keep94/finance/fin"
  "reflect"
  "testing"
)

func TestZeroValue(t *testing.T) {
  zero := CatDetailStore{}
  cdsb := CatDetailStoreBuilder{}
  expected := cdsb.Build()
  if !reflect.DeepEqual(expected.data(), zero.data()) {
    t.Error("Zero value is not zero.")
  }
}

func TestIsChildOf(t *testing.T) {
  cds := createCatDetailStore()
  if !cds.IsChildOf(toCat("0:0"), toCat("0:0")) {
    t.Error("0:0 should be a child of 0:0")
  }
  if !cds.IsChildOf(toCat("0:9983"), toCat("0:0")) {
    t.Error("0:9983 should be a child of 0:0")
  }
  if !cds.IsChildOf(toCat("0:9983"), toCat("0:9983")) {
    t.Error("0:9983 should be a child of 0:9983")
  }
  if cds.IsChildOf(toCat("0:9983"), toCat("0:9982")) {
    t.Error("0:9983 should not be a child of 0:9982")
  }
  if cds.IsChildOf(toCat("0:0"), toCat("0:9983")) {
    t.Error("0:0 should not be a child of 0:9983")
  }
  if cds.IsChildOf(toCat("0:9983"), toCat("0:1")) {
    t.Error("0:9983 should not be a child of 0:1")
  }
  if !cds.IsChildOf(toCat("0:1"), toCat("0:0")) {
    t.Error("0:1 should be a child of 0:0")
  }
  if !cds.IsChildOf(toCat("0:6"), toCat("0:0")) {
    t.Error("0:6 should be a child of 0:0")
  }
  if cds.IsChildOf(toCat("0:0"), toCat("0:1")) {
    t.Error("0:0 should not be a child of 0:1")
  }
  if !cds.IsChildOf(toCat("0:1"), toCat("0:1")) {
    t.Error("0:1 should be a child of 0:1")
  }
  if !cds.IsChildOf(toCat("0:5"), toCat("0:1")) {
    t.Error("0:5 should be a child of 0:1")
  }
  if cds.IsChildOf(toCat("0:1"), toCat("0:5")) {
    t.Error("0:1 should not be a child of 0:5")
  }
  if cds.IsChildOf(toCat("0:1"), toCat("0:5")) {
    t.Error("0:1 should not be a child of 0:5")
  }
  if !cds.IsChildOf(toCat("0:8"), toCat("0:0")) {
    t.Error("0:8 should be a child of 0:0")
  }
  if !cds.IsChildOf(toCat("0:8"), toCat("0:101")) {
    t.Error("0:8 should be a child of 0:101")
  }
  if cds.IsChildOf(toCat("0:101"), toCat("0:8")) {
    t.Error("0:101 should not be a child of 0:8")
  }
  if cds.IsChildOf(toCat("1:0"), toCat("0:0")) {
    t.Error("1:0 should not be a child of 0:0")
  }
  if cds.IsChildOf(toCat("0:0"), toCat("1:0")) {
    t.Error("0:0 should not be a child of 1:0")
  }
  if !cds.IsChildOf(toCat("1:1"), toCat("1:0")) {
    t.Error("1:1 should be a child of 1:0")
  }
  if !cds.IsChildOf(toCat("1:4"), toCat("1:1")) {
    t.Error("1:4 should be a child of 1:1")
  }
  if cds.IsChildOf(toCat("2:0"), toCat("0:0")) {
    t.Error("2:0 should not be a child of 0:0")
  }
  if cds.IsChildOf(toCat("2:2"), toCat("0:0")) {
    t.Error("2:2 should not be a child of 0:0")
  }
  if cds.IsChildOf(toCat("2:2"), toCat("2:0")) {
    t.Error("2:2 should not be a child of 2:0")
  }
  if !cds.IsChildOf(toCat("2:2"), toCat("2:2")) {
    t.Error("2:2 should be a child of 2:2")
  }
}

func TestPurgeableCats(t *testing.T) {
  cds := createCatDetailStore()
  actual := cds.PurgeableCats(fin.CatTotals{toCat("0:8"): 0})
  expected := fin.CatSet {
      toCat("0:3"): true,
      toCat("0:6"): true,
      toCat("0:7"): true,
      toCat("0:98"): true,
      toCat("0:99"): true}
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

func TestPurgeableCatsUsed(t *testing.T) {
  cds := createCatDetailStore()
  actual := cds.PurgeableCats(fin.CatTotals{
      toCat("0:7"): 0,
      toCat("0:6"): 0,
      toCat("0:8"): 0,
      toCat("0:98"): 0,
      toCat("0:99"): 0})
  if actual != nil {
    t.Errorf("Expected nil, got %v", actual)
  }
}

func TestPurgeableAccounts(t *testing.T) {
  cds := createCatDetailStore()
  actual := cds.PurgeableAccounts(fin.AccountSet{2: true})
  expected := fin.AccountSet{3: true, 4: true, 5: true}
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

func TestPurgeableAccountsUsed(t *testing.T) {
  cds := createCatDetailStore()
  actual := cds.PurgeableAccounts(
      fin.AccountSet{2: true, 3: true, 4: true, 5: true})
  if actual != nil {
    t.Errorf("Expected nil, got %v", actual)
  }
}

func TestAncestors(t *testing.T) {
  cds := createCatDetailStore()
  verifyAncestors(t, cds, toCat("0:0"),
      "0:0", "expense")
  verifyAncestors(t, cds, toCat("0:6"),
      "0:0", "expense",
      "0:2", "charity",
      "0:6", "inactive")
  verifyAncestors(t, cds, toCat("0:8"),
      "0:0", "expense",
      "0:101", "101",
      "0:8", "childagain")
  verifyAncestors(t, cds, toCat("1:0"),
      "1:0", "income")
  verifyAncestors(t, cds, toCat("1:1"),
      "1:0", "income",
      "1:1", "google")
}

func TestLeafNameById(t *testing.T) {
  cds := createCatDetailStore()
  verifyLeafName(t, cds, toCat("0:0"), "expense")
  verifyLeafName(t, cds, toCat("0:1"), "car")
  verifyLeafName(t, cds, toCat("0:2"), "charity")
  verifyLeafName(t, cds, toCat("0:3"), "inactive")
  verifyLeafName(t, cds, toCat("0:4"), "gas")
  verifyLeafName(t, cds, toCat("0:5"), "charity")
  verifyLeafName(t, cds, toCat("0:6"), "inactive")
  verifyLeafName(t, cds, toCat("0:7"), "child")
  verifyLeafName(t, cds, toCat("0:8"), "childagain")
  verifyLeafName(t, cds, toCat("1:0"), "income")
  verifyLeafName(t, cds, toCat("1:1"), "google")
  verifyLeafName(t, cds, toCat("1:2"), "mtv")
  verifyLeafName(t, cds, toCat("1:3"), "bonus")
  verifyLeafName(t, cds, toCat("2:1"), "checking")
  verifyLeafName(t, cds, toCat("2:2"), "savings")
  verifyLeafName(t, cds, toCat("2:3"), "inactive")
  verifyLeafName(t, cds, toCat("0:101"), "101")
  verifyLeafName(t, cds, toCat("1:101"), "101")
  verifyLeafName(t, cds, toCat("2:101"), "101")
}

func TestImmediateParent(t *testing.T) {
  cds := createCatDetailStore()
  verifyImmediateParent(t, cds, toCat("0:0"), "0:0")
  verifyImmediateParent(t, cds, toCat("0:1"), "0:0")
  verifyImmediateParent(t, cds, toCat("0:2"), "0:0")
  verifyImmediateParent(t, cds, toCat("0:4"), "0:1")
  verifyImmediateParent(t, cds, toCat("0:6"), "0:2")
  verifyImmediateParent(t, cds, toCat("0:7"), "0:3")
  verifyImmediateParent(t, cds, toCat("0:8"), "0:101")
  verifyImmediateParent(t, cds, toCat("1:0"), "1:0")
  verifyImmediateParent(t, cds, toCat("1:1"), "1:0")
  verifyImmediateParent(t, cds, toCat("1:3"), "1:1")
  verifyImmediateParent(t, cds, toCat("2:1"), "2:1")
  verifyImmediateParent(t, cds, toCat("0:101"), "0:0")
  verifyImmediateParent(t, cds, toCat("1:101"), "1:0")
  verifyImmediateParent(t, cds, toCat("2:101"), "2:101")
}

func TestDetailById(t *testing.T) {
  cds := createCatDetailStore()
  verifyCatDetail(t, cds, toCat("0:0"), "expense", true)
  verifyCatDetail(t, cds, toCat("0:1"), "expense:car", true)
  verifyCatDetail(t, cds, toCat("0:2"), "expense:charity", true)
  verifyCatDetail(t, cds, toCat("0:3"), "expense:inactive", false)
  verifyCatDetail(t, cds, toCat("0:4"), "expense:car:gas", true)
  verifyCatDetail(t, cds, toCat("0:5"), "expense:car:charity", true)
  verifyCatDetail(t, cds, toCat("0:6"), "expense:charity:inactive", false)
  verifyCatDetail(t, cds, toCat("0:7"), "expense:inactive:child", false)
  verifyCatDetail(t, cds, toCat("0:8"), "expense:101:childagain", false)
  verifyCatDetail(t, cds, toCat("1:0"), "income", true)
  verifyCatDetail(t, cds, toCat("1:1"), "income:google", true)
  verifyCatDetail(t, cds, toCat("1:2"), "income:mtv", true)
  verifyCatDetail(t, cds, toCat("1:3"), "income:google:bonus", true)
  verifyCatDetail(t, cds, toCat("2:1"), "account:checking", true)
  verifyCatDetail(t, cds, toCat("2:2"), "account:savings", true)
  verifyCatDetail(t, cds, toCat("2:3"), "account:inactive", false)
  verifyCatDetail(t, cds, toCat("0:101"), "expense:101", false)
  verifyCatDetail(t, cds, toCat("1:101"), "income:101", false)
  verifyCatDetail(t, cds, toCat("2:101"), "account:101", false)
}

func TestAccountDetailById(t *testing.T) {
  cds := createCatDetailStore()
  verifyAccountDetail(t, cds, 1, "checking", true)
  verifyAccountDetail(t, cds, 2, "savings", true)
  verifyAccountDetail(t, cds, 3, "inactive", false)
  verifyAccountDetail(t, cds, 101, "101", false)
}

func TestInvalidDetailByFullName(t *testing.T) {
  cds := createCatDetailStore()
  verifyInactiveDetailByFullName(t, cds, "expense:inactive", toCat("0:99"))
  verifyInactiveDetailByFullName(
      t, cds, "expense:charity:inactive", toCat("0:6"))
  verifyNoInactiveDetailByFullName(t, cds, "expense:inactive:child")
  verifyNoInactiveDetailByFullName(t, cds, "expense:car:gas")
  verifyNoInactiveDetailByFullName(t, cds, "expense:car")
  verifyInactiveDetailByFullName(t, cds, "account:inactive", toCat("2:4"))
  verifyNoInactiveDetailByFullName(t, cds, "account:checking")
  verifyNoInactiveDetailByFullName(t, cds, "account:savings")
}


func TestDetailByFullName(t *testing.T) {
  cds := createCatDetailStore()
  verifyDetailByFullName(t, cds, "expense", toCat("0:0"))
  verifyDetailByFullName(t, cds, "expense:car:gas", toCat("0:4"))
  verifyDetailByFullName(t, cds, "income", toCat("1:0"))
  verifyDetailByFullName(t, cds, "income:mtv", toCat("1:2"))
  verifyDetailByFullName(t, cds, "account:checking", toCat("2:1"))
  verifyNoDetailByFullName(t, cds, "expense:inactive")
  verifyNoDetailByFullName(t, cds, "expense:inactive:child")
  verifyNoDetailByFullName(t, cds, "expense:charity:inactive")
  verifyNoDetailByFullName(t, cds, "account:inactive")
}

func TestAccountDetailByName(t *testing.T) {
  cds := createCatDetailStore()
  verifyAccountDetailByName(t, cds, "checking", 1)
  verifyAccountDetailByName(t, cds, "savings", 2)
  verifyNoAccountDetailByName(t, cds, "inactive")
}

func TestActiveCatDetails(t *testing.T) {
  cds := createCatDetailStore()
  expected := []CatDetail {
    createCatDetail(toCat("0:0"), "expense", true),
    createCatDetail(toCat("0:1"), "expense:car", true),
    createCatDetail(toCat("0:5"), "expense:car:charity", true),
    createCatDetail(toCat("0:4"), "expense:car:gas", true),
    createCatDetail(toCat("0:2"), "expense:charity", true),
    createCatDetail(toCat("1:0"), "income", true),
    createCatDetail(toCat("1:1"), "income:google", true),
    createCatDetail(toCat("1:3"), "income:google:bonus", true),
    createCatDetail(toCat("1:4"), "income:google:bonus:bonus", true),
    createCatDetail(toCat("1:2"), "income:mtv", true),
    createCatDetail(toCat("2:1"), "account:checking", true),
    createCatDetail(toCat("2:2"), "account:savings", true)}
  details := cds.ActiveCatDetails(true)
  if !reflect.DeepEqual(details, expected) {
    t.Errorf("Expected %v, got %v", expected, details)
  }
  details = cds.ActiveCatDetails(false)
  l := len(expected)
  if !reflect.DeepEqual(details, expected[:l-2]) {
    t.Errorf("Expected %v, got %v", expected, details)
  }
}

func TestActiveAccountDetails(t *testing.T) {
  cds := createCatDetailStore()
  details := cds.ActiveAccountDetails()
  verifyAccountDetailIs(t, details[0], 1, "checking", true)
  verifyAccountDetailIs(t, details[1], 2, "savings", true)
  if output := len(details); output != 2 {
    t.Errorf("Expected 2 accounts, got %v", output)
  }
}

func TestSortedCatRecs(t *testing.T) {
  cds := createCatDetailStore()
  catrecs := []fin.ROCatRec {
      fin.NewROCatRec(&fin.CatRec{C: toCat("1:2")}),
      fin.NewROCatRec(&fin.CatRec{C: toCat("1:4")}),
      fin.NewROCatRec(&fin.CatRec{C: toCat("2:1")}),
      fin.NewROCatRec(&fin.CatRec{C: toCat("0:101")}),
      fin.NewROCatRec(&fin.CatRec{C: toCat("0:1")})}
  actual := cds.SortedCatRecs(catrecs)
  expected := []fin.ROCatRec {
      fin.NewROCatRec(&fin.CatRec{C:toCat("0:101")}),
      fin.NewROCatRec(&fin.CatRec{C:toCat("0:1")}),
      fin.NewROCatRec(&fin.CatRec{C:toCat("1:4")}),
      fin.NewROCatRec(&fin.CatRec{C:toCat("1:2")}),
      fin.NewROCatRec(&fin.CatRec{C: toCat("2:1")})}
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

func TestDetailsByIds(t *testing.T) {
  cds := createCatDetailStore()
  actual := cds.DetailsByIds(fin.CatSet {
      toCat("1:2"): true,
      toCat("1:4"): true,
      toCat("2:1"): true,
      toCat("0:101"): true,
      toCat("0:3"): true,
      toCat("0:1"): true,
      toCat("0:102"): false})
  expected := []CatDetail {
    createCatDetail(toCat("0:101"), "expense:101", false),
    createCatDetail(toCat("0:1"), "expense:car", true),
    createCatDetail(toCat("0:3"), "expense:inactive", false),
    createCatDetail(toCat("1:4"), "income:google:bonus:bonus", true),
    createCatDetail(toCat("1:2"), "income:mtv", true),
    createCatDetail(toCat("2:1"), "account:checking", true)}
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
} 

func TestRollUp(t *testing.T) {
  cds := createCatDetailStore()
  orig_totals := fin.CatTotals{
      toCat("0:8"): 100,
      toCat("0:101"): 200,
      toCat("0:9983"): 300,
      toCat("0:7"): 400,
      toCat("0:3"): 500,
      toCat("0:4"): 600,
      toCat("0:5"): 700,
      toCat("0:2"): 800,
      toCat("1:4"): 900,
      toCat("1:2"): 1000}
  totals, children := cds.RollUp(orig_totals)
  expected_totals := fin.CatTotals{
      toCat("0:8"): 100,
      toCat("0:101"): 300,
      toCat("0:9983"): 300,
      toCat("0:7"): 400,
      toCat("0:3"): 900,
      toCat("0:4"): 600,
      toCat("0:5"): 700,
      toCat("0:1"): 1300,
      toCat("0:2"): 800,
      toCat("1:4"): 900,
      toCat("1:3"): 900,
      toCat("1:1"): 900,
      toCat("1:2"): 1000,
      toCat("0:0"): 3600,
      toCat("1:0"): 1900}
  expected_children := map[fin.Cat]fin.CatSet {
      toCat("0:101"): {toCat("0:8"): true},
      toCat("0:3"): {toCat("0:7"): true},
      toCat("0:1"): {toCat("0:4"): true, toCat("0:5"): true},
      toCat("1:3"): {toCat("1:4"): true},
      toCat("1:1"): {toCat("1:3"): true},
      toCat("0:0"): {
          toCat("0:101"): true,
          toCat("0:9983"): true,
          toCat("0:3"): true,
          toCat("0:1"): true,
          toCat("0:2"): true},
      toCat("1:0"): {
          toCat("1:1"): true,
          toCat("1:2"): true}}
  if !reflect.DeepEqual(totals, expected_totals) {
    t.Errorf("Expected %v, got %v", expected_totals, totals)
  }
  if !reflect.DeepEqual(children, expected_children) {
    t.Errorf("Expected %v, got %v", expected_children, children)
  }
}

func TestFilterWithChildren(t *testing.T) {
  cds := createCatDetailStore()
  cat_filter := cds.Filter(toCat("0:1"), true)
  verifyFilterIncludes(t, cat_filter, toCat("0:1"), toCat("0:5"))
  verifyFilterExcludes(t, cat_filter, toCat("0:2"))
}

func TestFilterNoChildren(t *testing.T) {
  cds := createCatDetailStore()
  cat_filter := cds.Filter(toCat("0:1"), false)
  verifyFilterIncludes(t, cat_filter, toCat("0:1"))
  verifyFilterExcludes(t, cat_filter, toCat("0:5"))
}
         
func TestFilterTopLevel(t *testing.T) {
  cds := createCatDetailStore()
  cat_filter := cds.Filter(toCat("0:0"), true)
  verifyFilterIncludes(t, cat_filter, toCat("0:0"), toCat("0:9983"))
  verifyFilterExcludes(t, cat_filter, toCat("1:0"))
}
         
func TestFilterTopLevelNoChildren(t *testing.T) {
  cds := createCatDetailStore()
  cat_filter := cds.Filter(toCat("0:0"), false)
  verifyFilterIncludes(t, cat_filter, toCat("0:0"))
  verifyFilterExcludes(t, cat_filter, toCat("0:9983"))
}

func toCat(s string) fin.Cat {
  return fin.NewCat(s)
}

func createCatDetailStore() CatDetailStore {
  // 0:1 expense:car
  // 0:2 expense:charity
  // 0:3 expense:inactive
  // 0:4 expense:car:gas
  // 0:5 expense:car:charity
  // 0:6 expense:charity:inactive
  // 0:7 expense:inactive:child
  // 0:8 expense:101:childagain
  // 0:98 expense:car:gas
  // 0:99 expense:inactive
  // 1:1 income:google
  // 1:2 income:mtv
  // 1:3 income:google:bonus
  // 2:1 account:checking
  // 2:2 account:savings
  // 2:3 account:inactive
  // 2:4 account:inactive
  // 2:5 account:checking
  cdsb := CatDetailStoreBuilder{}
  account := fin.Account{
      Id: 1,
      Name: "checking",
      Active: true}
  cdsb.AddAccount(&account)
  account = fin.Account{
      Id: 2,
      Name: "savings",
      Active: true}
  cdsb.AddAccount(&account)
  account = fin.Account{
      Id: 3,
      Name: "inactive",
      Active: false}
  cdsb.AddAccount(&account)
  account = fin.Account{
      Id: 4,
      Name: "inactive",
      Active: false}
  cdsb.AddAccount(&account)
  account = fin.Account{
      Id: 5,
      Name: "checking",
      Active: false}
  cdsb.AddAccount(&account)
  row := CatDbRow{
      Id: 1,
      Name: "car",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 2,
      Name: "charity",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 3,
      Name: "inactive",
      Active: false}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 4,
      ParentId: 1,
      Name: "gas",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 5,
      ParentId: 1,
      Name: "charity",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 6,
      ParentId: 2,
      Name: "inactive",
      Active: false}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 7,
      ParentId: 3,
      Name: "child",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 8,
      ParentId: 101,
      Name: "childagain",
      Active: true}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 98,
      ParentId: 1,
      Name: "gas",
      Active: false}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 99,
      Name: "inactive",
      Active: false}
  cdsb.AddCatDbRow(fin.ExpenseCat, &row)
  row = CatDbRow{
      Id: 1,
      Name: "google",
      Active: true}
  cdsb.AddCatDbRow(fin.IncomeCat, &row)
  row = CatDbRow{
      Id: 2,
      Name: "mtv",
      Active: true}
  cdsb.AddCatDbRow(fin.IncomeCat, &row)
  row = CatDbRow{
      Id: 3,
      ParentId: 1,
      Name: "bonus",
      Active: true}
  cdsb.AddCatDbRow(fin.IncomeCat, &row)
  row = CatDbRow{
      Id: 4,
      ParentId: 3,
      Name: "bonus",
      Active: true}
  cdsb.AddCatDbRow(fin.IncomeCat, &row)
  return cdsb.Build()
}

func verifyCatDetail(t *testing.T, cds CatDetailStore, cat fin.Cat, name string, active bool) {
  cd := cds.DetailById(cat)
  if cd.Id() != cat {
    t.Errorf("Expected %v, got %v", cat, cd.Id())
  }
  if cd.FullName() != name {
    t.Errorf("Expected %v, got %v", name, cd.FullName())
  }
  if cd.Active() != active {
    t.Errorf("Expected %v, got %v", active, cd.Active())
  }
}

func verifyLeafName(t *testing.T, cds CatDetailStore, cat fin.Cat, name string) {
  an := cds.LeafNameById(cat)
  if an != name {
    t.Errorf("Expected %v, got %v", name, an)
  }
}

func verifyImmediateParent(
    t *testing.T, cds CatDetailStore, cat fin.Cat, parent string) {
  parentCat := cds.ImmediateParent(cat)
  if parentCat.String() != parent {
    t.Errorf("Expected %v, got %v", parent, parentCat.String())
  }
}

func verifyAncestors(
    t *testing.T, cds CatDetailStore, cat fin.Cat, catAndName ...string) {
  ancestors := Ancestors(cds, cat)
  if len(ancestors) != len(catAndName) / 2 {
    t.Errorf(
        "Expected %d ancestors, got %d", len(catAndName) / 2, len(ancestors))
  }
  for i := 0; i < len(ancestors); i++ {
    if ancestors[i].Id.String() != catAndName[2*i] {
      t.Errorf(
          "Index %d, expected %s, got %s",
           i,
           catAndName[2*i],
           ancestors[i].Id.String())
    }
    if ancestors[i].Name != catAndName[2*i + 1] {
      t.Errorf(
          "Index %d, expected %s, got %s",
           i,
           catAndName[2*i + 1],
           ancestors[i].Name)
    }
  }
}

func verifyAccountDetail(t *testing.T, cds CatDetailStore, id int64, name string, active bool) {
  ad := cds.AccountDetailById(id)
  verifyAccountDetailIs(t, ad, id, name, active)
}

func verifyAccountDetailIs(t *testing.T, ad AccountDetail, id int64, name string, active bool) {
  if ad.Id() != id {
    t.Errorf("Expected %v, got %v", id, ad.Id())
  }
  if ad.Name() != name {
    t.Errorf("Expected %v, got %v", name, ad.Name())
  }
  if ad.Active() != active {
    t.Errorf("Expected %v, got %v", active, ad.Active())
  }
}

func verifyDetailByFullName(t *testing.T, cds CatDetailStore, name string, cat fin.Cat) {
  cd, ok := cds.DetailByFullName(name)
  if !ok {
    t.Error("Error getting detail by full name.")
    return
  }
  if cd.Id() != cat {
    t.Errorf("Expected %v, got %v", cat, cd.Id())
  }
}

func verifyInactiveDetailByFullName(
    t *testing.T, cds CatDetailStore, name string, cat fin.Cat) {
  cd, ok := cds.InactiveDetailByFullName(name)
  if !ok {
    t.Error("Error getting inactive detail by full name.")
    return
  }
  if cd.Id() != cat {
    t.Errorf("Expected %v, got %v", cat, cd.Id())
  }
}

func verifyAccountDetailByName(t *testing.T, cds CatDetailStore, name string, id int64) {
  ad, ok := cds.AccountDetailByName(name)
  if !ok {
    t.Error("Error getting account detail by name.")
    return
  }
  if ad.Id() != id {
    t.Errorf("Expected %v, got %v", id, ad.Id())
  }
}

func verifyNoDetailByFullName(t *testing.T, cds CatDetailStore, name string) {
  _, ok := cds.DetailByFullName(name)
  if ok {
    t.Error("Got detail by full name, but did not expect to")
  }
}

func verifyNoInactiveDetailByFullName(
    t *testing.T, cds CatDetailStore, name string) {
  _, ok := cds.InactiveDetailByFullName(name)
  if ok {
    t.Error("Got inactive detail by full name, but did not expect to")
  }
}

func verifyNoAccountDetailByName(t *testing.T, cds CatDetailStore, name string) {
  _, ok := cds.AccountDetailByName(name)
  if ok {
    t.Error("Got account detail by name, but did not expect to")
  }
}

func createCatDetail(cat fin.Cat, full_name string, active bool) CatDetail {
  return CatDetail{&catDetail{id: cat, fullName: full_name, active: active}}
}

func verifyFilterIncludes(t *testing.T, filter fin.CatFilter, cats ...fin.Cat) {
  for _, cat := range cats {
    if !filter(cat) {
      t.Errorf("Expected %v to be included", cat)
    }
  }
}

func verifyFilterExcludes(t *testing.T, filter fin.CatFilter, cats ...fin.Cat) {
  for _, cat := range cats {
    if filter(cat) {
      t.Errorf("Expected %v to be excluded", cat)
    }
  }
}
