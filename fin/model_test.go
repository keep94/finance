package fin

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/etag"
  "reflect"
  "testing"
)

func TestBuildCatPayment(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cpb.SetPaymentId(5).SetReconciled(true)

  // 0:9 should not disappear even though its total amount is 0
  cpb.AddCatRec(&CatRec{NewCat("0:9"), 4009, false})
  cpb.AddCatRec(&CatRec{NewCat("0:9"), -4009, false})

  cpb.AddCatRec(&CatRec{NewCat("0:5"), 2324, false})
  cpb.AddCatRec(&CatRec{NewCat("0:6"), 9002, false})

  // 2:5 should be ignored since it is the payment type
  cpb.AddCatRec(&CatRec{NewCat("2:5"), 3535, false})
  cpb.AddCatRec(&CatRec{NewCat("2:6"), 5003, false})

  // This 0:5 should be merged with first one
  cpb.AddCatRec(&CatRec{NewCat("0:5"), 1076, false})

  cp := cpb.Build()
  if cp.WithPayment(7) {
    t.Error("WithPayment(7) should have failed.")
  }
  if cp.Reconcile(7) {
    t.Error("Reconcile(7) should have failed.")
  }
  if cp.WithCat(func(c Cat) bool { return false }) {
    t.Error("WithCat() should have failed.")
  }
  if verifyCatPayment(t, &cp, -17405, 4, 5, true) {
    verifyCatRec(t, &cp, 0, "0:5", 3400, false)
    verifyCatRec(t, &cp, 1, "0:6", 9002, false)
    verifyCatRec(t, &cp, 2, "0:9", 0, false)
    verifyCatRec(t, &cp, 3, "2:6", 5003, false)
  }

  sixPayment := cp
  if !sixPayment.WithPayment(6) {
    t.Error("Expected WithPayment(6) to succeed.")
  }
  if verifyCatPayment(t, &sixPayment, 5003, 1, 6, false) {
    verifyCatRec(t, &sixPayment, 0, "2:5", -5003, true)
  }

  fiveReconciled := cp
  if !fiveReconciled.Reconcile(5) {
    t.Error("Expected Reconcil(5) to succeed")
  }
  if verifyCatPayment(t, &fiveReconciled, -17405, 4, 5, true) {
    verifyCatRec(t, &fiveReconciled, 0, "0:5", 3400, false)
    verifyCatRec(t, &fiveReconciled, 1, "0:6", 9002, false)
    verifyCatRec(t, &fiveReconciled, 2, "0:9", 0, false)
    verifyCatRec(t, &fiveReconciled, 3, "2:6", 5003, false)
  }

  sixReconciled := cp
  if !sixReconciled.Reconcile(6) {
    t.Error("Expected Reconcile(6) to succeed")
  }
  if verifyCatPayment(t, &sixReconciled, -17405, 4, 5, true) {
    verifyCatRec(t, &sixReconciled, 0, "0:5", 3400, false)
    verifyCatRec(t, &sixReconciled, 1, "0:6", 9002, false)
    verifyCatRec(t, &sixReconciled, 2, "0:9", 0, false)
    verifyCatRec(t, &sixReconciled, 3, "2:6", 5003, true)
  }

  expenseOnly := cp
  if !expenseOnly.WithCat(func(c Cat) bool { return c.Type == ExpenseCat }) {
    t.Error("Expected WithCat() to succeed.")
  }
  if verifyCatPayment(t, &expenseOnly, -12402, 3, 5, true) {
    verifyCatRec(t, &expenseOnly, 0, "0:5", 3400, false)
    verifyCatRec(t, &expenseOnly, 1, "0:6", 9002, false)
    verifyCatRec(t, &expenseOnly, 2, "0:9", 0, false)
  }
  if verifyCatPayment(t, &cp, -17405, 4, 5, true) {
    verifyCatRec(t, &cp, 0, "0:5", 3400, false)
    verifyCatRec(t, &cp, 1, "0:6", 9002, false)
    verifyCatRec(t, &cp, 2, "0:9", 0, false)
    verifyCatRec(t, &cp, 3, "2:6", 5003, false)
  }
}

func TestBuildCatPaymentSetPaymentLast(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cpb.AddCatRec(&CatRec{NewCat("2:5"), 3456, false})
  cpb.AddCatRec(&CatRec{NewCat("0:1"), 1234, false})
  cpb.SetPaymentId(5)
  cp := cpb.Build()
  if verifyCatPayment(t, &cp, -1234, 1, 5, false) {
    verifyCatRec(t, &cp, 0, "0:1", 1234, false)
  }
  fiveReconciled := cp
  if !fiveReconciled.Reconcile(5) {
    t.Error("Expected Reconcile(5) to succeed.")
  }
  if verifyCatPayment(t, &fiveReconciled, -1234, 1, 5, true) {
    verifyCatRec(t, &fiveReconciled, 0, "0:1", 1234, false)
  }
  if verifyCatPayment(t, &cp, -1234, 1, 5, false) {
    verifyCatRec(t, &cp, 0, "0:1", 1234, false)
  }
}

func TestZeroCatPaymentAndBuilder(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cp := CatPayment{}
  verifyCatPayment(t, &cp, 0, 0, 0, false)
  cp = cpb.Build()
  verifyCatPayment(t, &cp, 0, 0, 0, false)
}

func TestMergeReconcileInCatRec(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cpb.AddCatRec(&CatRec{NewCat("0:5"), 10000, false})
  cpb.SetPaymentId(9).SetReconciled(true)
  cp := cpb.Build()
  cpb.AddCatRec(&CatRec{NewCat("0:7"), 3000, true})
  cpb.AddCatRec(&CatRec{NewCat("0:7"), 1000, false})
  cp2 := cpb.Build()
  cpb.AddCatRec(&CatRec{NewCat("0:7"), 2000, false})
  cp3 := cpb.Build()
  verifyCatRec(t, &cp, 0, "0:5", 10000, false)
  verifyCatPayment(t, &cp, -10000, 1, 9, true)
  verifyCatRec(t, &cp2, 0, "0:7", 4000, true)
  verifyCatPayment(t, &cp2, -4000, 1, 0, false)
  verifyCatRec(t, &cp3, 0, "0:7", 2000, false)
  verifyCatPayment(t, &cp3, -2000, 1, 0, false)
}

func TestSetSingleCat(t *testing.T) {
  cp := NewCatPayment(NewCat("0:5"), 1234, true, 3)
  if !cp.SetSingleCat(NewCat("0:8")) {
    t.Error("Expected SetSingleCat to succeed.")
  }
  verifyCatPayment(t, &cp, -1234, 1, 3, true)
  verifyCatRec(t, &cp, 0, "0:8", 1234, false)

  cp = NewCatPayment(NewCat("0:5"), 1234, false, 3)
  if !cp.SetSingleCat(NewCat("2:8")) {
    t.Error("Expected SetSingleCat to succeed.")
  }
  verifyCatPayment(t, &cp, -1234, 1, 3, false)
  verifyCatRec(t, &cp, 0, "2:8", 1234, false)

  if cp.SetSingleCat(NewCat("2:3")) {
    t.Error("Expected SetSingleCat to fail.")
  }
  verifyCatPayment(t, &cp, -1234, 1, 3, false)
  verifyCatRec(t, &cp, 0, "2:8", 1234, false)
}

func TestChangeCat(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cp := cpb.AddCatRec(
      &CatRec{NewCat("0:5"), 1000, false}).AddCatRec(
      &CatRec{NewCat("0:7"), 2000, false}).AddCatRec(
      &CatRec{NewCat("0:10"), 4000, false}).SetPaymentId(
      9).SetReconciled(true).Build()
  // Change 0:5 to 0:7
  modifyCat(NewCat("0:5"), NewCat("0:7"), &cp)
  verifyCatPayment(t, &cp, -7000, 2, 9, true)
  verifyCatRec(t, &cp, 0, "0:7", 3000, false)
  verifyCatRec(t, &cp, 1, "0:10", 4000, false)
}

func TestCatPaymentBuilderSet(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cp := cpb.AddCatRec(
      &CatRec{NewCat("0:5"), 1000, false}).AddCatRec(
      &CatRec{NewCat("0:7"), 2000, false}).AddCatRec(
      &CatRec{NewCat("0:10"), 4000, false}).SetPaymentId(
      9).SetReconciled(true).Build()
  newCpb := CatPaymentBuilder{}
  newCpb.AddCatRec(&CatRec{NewCat("0:11"), 5500, true})
  newCpb.Set(&cp)
  newCp := newCpb.Build()
  verifyCatPayment(t, &newCp, -7000, 3, 9, true)
  verifyCatRec(t, &newCp, 0, "0:5", 1000, false)
  verifyCatRec(t, &newCp, 1, "0:7", 2000, false)
  verifyCatRec(t, &newCp, 2, "0:10", 4000, false)
}

  

func TestCatTotals(t *testing.T) {
  var ct CatTotals = make(map[Cat]int64)
  cpb := CatPaymentBuilder{}
  cp := cpb.AddCatRec(
      &CatRec{NewCat("0:7"), 6000, false}).AddCatRec(
      &CatRec{NewCat("1:5"), -3000, false}).AddCatRec(
      &CatRec{NewCat("0:3"), 2000, true}).AddCatRec(
      &CatRec{NewCat("2:2"), 1000, false}).SetPaymentId(
      1).SetReconciled(false).Build()
  ct.Include(&cp)
  var expected CatTotals = map[Cat]int64 {
      Cat{3, ExpenseCat}: 2000,
      Cat{7, ExpenseCat}: 6000,
      Cat{5, IncomeCat}: -3000}
  if !reflect.DeepEqual(expected, ct) {
    t.Errorf("Expected %v, got %v", expected, ct)
  }
  cp = cpb.AddCatRec(
      &CatRec{NewCat("0:7"), 1000, false}).AddCatRec(
      &CatRec{NewCat("1:5"), 3000, false}).AddCatRec(
      &CatRec{NewCat("1:7"), 0, false}).AddCatRec(
      &CatRec{NewCat("0:4"), 1500, true}).AddCatRec(
      &CatRec{NewCat("2:2"), 1000, false}).SetPaymentId(
      1).SetReconciled(false).Build()
  ct.Include(&cp)
  expected = map[Cat]int64 {
      Cat{3, ExpenseCat}: 2000,
      Cat{4, ExpenseCat}: 1500,
      Cat{5, IncomeCat}: 0,
      Cat{7, ExpenseCat}: 7000,
      Cat{7, IncomeCat}: 0}
  if !reflect.DeepEqual(expected, ct) {
    t.Errorf("Expected %v, got %v", expected, ct)
  }
}

func TestAccountSet(t *testing.T) {
  as := make(AccountSet)
  cpb := CatPaymentBuilder{}
  cp := cpb.AddCatRec(
      &CatRec{NewCat("0:7"), 6000, false}).AddCatRec(
      &CatRec{NewCat("1:5"), -3000, false}).AddCatRec(
      &CatRec{NewCat("2:4"), 2000, true}).AddCatRec(
      &CatRec{NewCat("2:2"), 1000, false}).SetPaymentId(
      1).SetReconciled(false).Build()
  as.Include(&cp)
  var expected AccountSet = AccountSet{1: true, 2: true, 4:true}
  if !reflect.DeepEqual(expected, as) {
    t.Errorf("Expected %v, got %v", expected, as)
  }
}

func TestAccountDeltas(t *testing.T) {
  var d AccountDeltas = make(map[int64]*AccountDelta)
  cpb := CatPaymentBuilder{}
  cp := cpb.AddCatRec(
      &CatRec{NewCat("0:7"), 6000, false}).AddCatRec(
      &CatRec{NewCat("2:3"), 1100, true}).AddCatRec(
      &CatRec{NewCat("2:3"), 900, false}).AddCatRec(
      &CatRec{NewCat("2:2"), 700, false}).AddCatRec(
      &CatRec{NewCat("2:2"), 300, false}).AddCatRec(
      &CatRec{NewCat("2:1"), 5800, true}).SetPaymentId(
      1).SetReconciled(false).Build()
  d.Include(&cp)
  var expected AccountDeltas = map[int64]*AccountDelta{
      1: {-9000, 0, 1, 0}, 2: {1000, 0, 1, 0}, 3: {2000, 2000, 1, 1}}
  if !reflect.DeepEqual(expected, d) {
    t.Errorf("Expected %v, got %v", expected, d)
  }
  cp2 := cpb.AddCatRec(
      &CatRec{NewCat("0:7"), 2000, false}).SetPaymentId(
      3).SetReconciled(true).Build()
  d.Include(&cp2)
  expected = map[int64]*AccountDelta{
      1: {-9000, 0, 1, 0}, 2: {1000, 0, 1, 0}, 3: {0, 0, 2, 2}}
  if !reflect.DeepEqual(expected, d) {
    t.Errorf("Expected %v, got %v", expected, d)
  }
  d.Exclude(&cp)
  d.Exclude(&cp2)
  expected = map[int64]*AccountDelta{}
  if !reflect.DeepEqual(expected, d) {
    t.Errorf("Expected %v, got %v", expected, d)
  }
}

func TestZeroCatPaymentsEqual(t *testing.T) {
  zero := CatPayment{}
  cpb := CatPaymentBuilder{}
  setToZero := cpb.Build()
  if !reflect.DeepEqual(zero, setToZero) {
    t.Error("Expected zero CatPayments to be equal.")
  }
}

func TestEtagEmpty(t *testing.T) {
  cpb := CatPaymentBuilder{}
  entry := Entry{}
  tag, err := etag.Etag32(&entry)
  if err != nil {
    t.Errorf("Got error computing etag: %v", err)
  }
  entry_copy := Entry{
      Date: date_util.YMD(1, 1, 1),
      CatPayment: cpb.Build()}
  tag_copy, err := etag.Etag32(&entry_copy)
  if err != nil {
    t.Errorf("Got error computing etag: %v", err)
  }
  if tag != tag_copy {
    t.Errorf("Etags should be equal, got %v and %v", tag, tag_copy)
  }
}

func TestEtag(t *testing.T) {
  cpb := CatPaymentBuilder{}
  entry := Entry{
     Id: 35,
     Date: date_util.YMD(2012, 12, 25),
     Name: "Name",
     Desc: "Desc",
     CheckNo: "1234",
     Status: Reviewed,
     CatPayment: cpb.AddCatRec(
         &CatRec{NewCat("0:7"), 3543, false}).AddCatRec(
         &CatRec{NewCat("0:3"), 2312, false}).SetPaymentId(
         1).SetReconciled(false).Build()}
  entry_copy := Entry{
     Id: 35,
     Date: date_util.YMD(2012, 12, 25),
     Name: "Name",
     Desc: "Desc",
     CheckNo: "1234",
     Status: Reviewed,
     CatPayment: cpb.AddCatRec(
         &CatRec{NewCat("0:7"), 3543, false}).AddCatRec(
         &CatRec{NewCat("0:3"), 2312, false}).SetPaymentId(
         1).SetReconciled(false).Build()}
  entry_diff := Entry{
     Id: 35,
     Date: date_util.YMD(2012, 12, 25),
     Name: "Name",
     Desc: "Desc",
     CheckNo: "1234",
     Status: Reviewed,
     CatPayment: cpb.AddCatRec(
         &CatRec{NewCat("0:7"), 3543, false}).AddCatRec(
         &CatRec{NewCat("0:3"), 2312, true}).SetPaymentId(
         1).SetReconciled(false).Build()}
  tag, err := etag.Etag32(&entry)
  if err != nil {
    t.Errorf("Got error computing etag: %v", err)
  }
  tag_copy, err := etag.Etag32(&entry_copy)
  if err != nil {
    t.Errorf("Got error computing etag: %v", err)
  }
  tag_diff, err := etag.Etag32(&entry_diff)
  if err != nil {
    t.Errorf("Got error computing etag: %v", err)
  }
  if tag != tag_copy {
    t.Errorf("Expected ETags to be equal, got %v and %v", tag, tag_copy)
  }
  if tag == tag_diff {
    t.Error("Expected ETags to be unequal")
  }
}

func TestCatRecs(t *testing.T) {
  cpb := CatPaymentBuilder{}
  cpb.AddCatRec(&CatRec{C: NewCat("0:1"), A: 1})
  cpb.AddCatRec(&CatRec{C: NewCat("0:2"), A: 1})
  cpb.AddCatRec(&CatRec{C: NewCat("0:3"), A: 1})
  cp := cpb.Build()
  catrecs := cp.CatRecs()
  if len(catrecs) != 3 {
    t.Fatalf("Expect lenth to be 3, got %v", len(catrecs))
  }
  for i := 0; i < 3; i++ {
    expected := Cat{Id: int64(i) +1, Type: ExpenseCat}
    if catrecs[i].Id() != expected {
      t.Errorf("Expected %v, got %v", expected, catrecs[i].Id())
    }
  }
}

func TestFormatUSD(t *testing.T) {
  if output := FormatUSD(347); output != "3.47" {
    t.Errorf("Expected 3.47, got %v", output)
  }
  if output := FormatUSD(-347); output != "-3.47" {
    t.Errorf("Expected -3.47, got %v", output)
  }
  if output := FormatUSD(0); output != "0.00" {
    t.Errorf("Expected 0.00, got %v", output)
  }
  if output := FormatUSD(1); output != "0.01" {
    t.Errorf("Expected 0.01 got %v", output)
  }
}

func TestParseUSD(t *testing.T) {
  verifyParseUSD(t, "3.476", 348)
  verifyParseUSD(t, "3.474", 347)
  verifyParseUSD(t, "-3.476", -348)
  verifyParseUSD(t, "-3.474", -347)
  verifyParseUSDError(t, "abc")
}

func TestCatSet(t *testing.T) {
  catSet := make(CatSet)
  catSet.AddSet(
      CatSet{NewCat("0:3"): true, NewCat("0:4"): true}).AddSet(
      CatSet{NewCat("0:4"): true, NewCat("0:5"): true, NewCat("0:6"): false})
  expected := CatSet{NewCat("0:3"): true, NewCat("0:4"): true, NewCat("0:5"): true}
  if !reflect.DeepEqual(expected, catSet) {
    t.Errorf("Expected %v, got %v", expected, catSet)
  }
}

func TestPermission(t *testing.T) {
  if p, ok := ToPermission(AllPermission.ToInt()); p != AllPermission || !ok {
    t.Error("Round trip failed for AllPermission")
  }
  if p, ok := ToPermission(ReadPermission.ToInt()); p != ReadPermission || !ok {
    t.Error("Round trip failed for ReadPermission")
  }
  if p, ok := ToPermission(NonePermission.ToInt()); p != NonePermission || !ok {
    t.Error("Round trip failed for NonePermission")
  }
  // -1 is smallest allowed int
  if p, ok := ToPermission(-2); p != NonePermission || ok {
    t.Error("Failure with illegal permission value")
  }
  // just larger than largest allowed int
  if p, ok := ToPermission(int(NonePermission)); p != NonePermission || ok {
    t.Error("Failure with illegal permission value")
  }
  if AllPermission.String() != "All" {
    t.Error("expected 'All'")
  }
  if ReadPermission.String() != "Read" {
    t.Error("expected 'Read'")
  }
  if NonePermission.String() != "None" {
    t.Error("expected 'None'")
  }
}

func verifyParseUSD(t *testing.T, s string, expected int64) {
  output, e := ParseUSD(s)
  if e != nil {
    t.Errorf("Got error parsing: %v", e)
    return
  }
  if output != expected {
    t.Errorf("Expected %v, got %v", expected, output)
  }
}

func verifyParseUSDError(t *testing.T, s string) {
  _, e := ParseUSD(s)
  if e == nil {
    t.Error("Expected error parsing.")
  }
}

func verifyCatRec(t *testing.T, cp *CatPayment, idx int, catId string, amount int64, reconciled bool) {
  catrec := cp.CatRecByIndex(idx)
  if catrec.Reconciled() != reconciled {
    t.Errorf("Expected %v, got %v", reconciled, catrec.Reconciled())
  }
  if catrec.Amount() != amount {
    t.Errorf("Expected %v, got %v", amount, catrec.Amount())
  }
  if output := catrec.Id().ToString(); output != catId {
    t.Errorf("Expected %v, got %v", catId, output)
  }
}

func verifyCatPayment(t *testing.T, cp *CatPayment, total int64, count int, paymentId int64, reconciled bool) bool {
  if output := cp.CatRecCount(); output != count {
    t.Errorf("Expected %v, got %v", count, output)
    return false
  }
  if output := cp.Total(); output != total {
    t.Errorf("Expected %v, got %v", total, output)
    return false
  }
  if output := cp.PaymentId(); output != paymentId {
    t.Errorf("Expected %v, got %v", paymentId, output)
    return false
  }
  if output := cp.Reconciled(); output != reconciled {
    t.Errorf("Expected %v, got %v", reconciled, output)
    return false
  }
  return true
}

func modifyCat(oldCat, newCat Cat, cp *CatPayment) {
  cpb := CatPaymentBuilder{}
  cpb.Set(cp).ClearCatRecs()
  newCatRec := CatRec{}
  oldCatRecs := cp.CatRecs()
  for _, oldCatRec := range oldCatRecs {
    if oldCatRec.Id() == oldCat {
      newCatRec.Set(oldCatRec)
      newCatRec.C = newCat
      cpb.AddCatRec(&newCatRec)
    } else {
      cpb.AddROCatRec(oldCatRec)
    }
  }
  *cp = cpb.Build()
}
