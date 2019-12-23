package reconcile

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/finance/fin"
  "github.com/keep94/goconsume"
  "reflect"
  "testing"
  "time"
)

var (
  kNoCheckNo = AmountCheckNo{-123, ""}
)

func TestReconcileSingleYes(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 8))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
}

func TestReconcileSingleYes2(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 1))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
}

func TestReconcileSingleNo1(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 3, 31))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestReconcileSingleNo2(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestB1U0(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  bank.Reconcile(nil, 7)
  if output := b1.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestReconcileB2U1(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 3))
  b2 := newEntry(0, date_util.YMD(2013, 4, 1))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1, b2}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
  if output := b2.Id; output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
}

func TestReconcileB2U1_2(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 3))
  b2 := newEntry(0, date_util.YMD(2013, 3, 31))
  u1 := newEntry(1, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1, b2}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
  if output := b2.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestReconcileB1U2(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 3))
  u2 := newEntry(2, date_util.YMD(2013, 4, 3))
  u3 := newEntry(3, date_util.YMD(2013, 4, 1))
  bank := ByAmountCheckNo{kNoCheckNo: {b1}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u2, u3}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 2 {
    t.Errorf("Expected 2, got %v", output)
  }
}

func TestReconcileB2U2(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  b2 := newEntry(0, date_util.YMD(2013, 4, 7))
  u2 := newEntry(2, date_util.YMD(2013, 4, 8))
  u3 := newEntry(3, date_util.YMD(2013, 4, 6))
  bank := ByAmountCheckNo{kNoCheckNo: {b1, b2}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u2, u3}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 2 {
    t.Errorf("Expected 2, got %v", output)
  }
  if output := b2.Id; output != 3 {
    t.Errorf("Expected 3, got %v", output)
  }
}

func TestReconcileB4U4(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  b2 := newEntry(0, date_util.YMD(2013, 4, 4))
  b3 := newEntry(0, date_util.YMD(2013, 4, 2))
  b4 := newEntry(0, date_util.YMD(2013, 4, 1))

  u8 := newEntry(8, date_util.YMD(2013, 4, 7))
  u6 := newEntry(6, date_util.YMD(2013, 4, 2))
  u4 := newEntry(4, date_util.YMD(2013, 4, 1))
  u2 := newEntry(2, date_util.YMD(2013, 3, 24))

  bank := ByAmountCheckNo{kNoCheckNo: {b1, b2, b3, b4}}
  unreconciled := ByAmountCheckNo{kNoCheckNo: {u8, u6, u4, u2}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 8 {
    t.Errorf("Expected 8, got %v", output)
  }
  if output := b2.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
  if output := b3.Id; output != 6 {
     t.Errorf("Expected 6, got %v", output) 
  }
  if output := b4.Id; output != 4 {
     t.Errorf("Expected 4, got %v", output) 
  }
}

func TestReconcileCheckYes(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  u1 := newEntry(8, date_util.YMD(2013, 4, 1))

  bank := ByAmountCheckNo{AmountCheckNo{-123, "123"}: {b1}}
  unreconciled := ByAmountCheckNo{AmountCheckNo{-123, "123"}: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 8 {
    t.Errorf("Expected 8, got %v", output)
  }
}

func TestReconcileCheckNo(t *testing.T) {
  b1 := newEntry(0, date_util.YMD(2013, 4, 9))
  u1 := newEntry(8, date_util.YMD(2013, 4, 10))

  bank := ByAmountCheckNo{AmountCheckNo{-123, "123"}: {b1}}
  unreconciled := ByAmountCheckNo{AmountCheckNo{-123, "123"}: {u1}}
  bank.Reconcile(unreconciled, 7)
  if output := b1.Id; output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestNew(t *testing.T) {
  e1 := &fin.Entry{
      Date: date_util.YMD(2013, 4, 1),
      CheckNo: "123",
      CatPayment: fin.NewCatPayment(fin.Expense, 924, true, 0)}
  e2 := &fin.Entry{
      Date: date_util.YMD(2013, 4, 3),
      CheckNo: "123",
      CatPayment: fin.NewCatPayment(fin.Expense, 924, true, 0)}
  e3 := &fin.Entry{
      Date: date_util.YMD(2013, 4, 9),
      CheckNo: "123",
      CatPayment: fin.NewCatPayment(fin.Expense, 924, true, 0)}
  bank := New([]*fin.Entry {e1, e2, e3})
  entries := bank[AmountCheckNo{-924, "123"}]
  if entries[0] != e3 || entries[1] != e2 || entries[2] != e1 {
    t.Error("Entries added out of order.")
  }
}

func TestGetChanges(t *testing.T) {
  entries := []*fin.Entry{
      {Name: "Add1"},
      {Id: 924,
       Name: "Update1",
       CatPayment: fin.NewCatPayment(fin.NewCat("0:73"), 0, true, 3)},
      {Name: "Add2"},
      {Id: 1007,
       Name: "Update2",
       CatPayment: fin.NewCatPayment(fin.NewCat("0:89"), 0, true, 3)}}
  changes := GetChanges(entries)
  addNames := make(map[string]bool)
  for _, v := range changes.Adds {
    addNames[v.Name] = true
  }
  if !reflect.DeepEqual(addNames, map[string]bool {"Add1": true, "Add2": true}) {
    t.Errorf("Expected Add1, Add2 but got %v", addNames)
  }
  if changes.Updates[924] == nil || changes.Updates[1007] == nil || len(changes.Updates) != 2 {
    t.Error("Updates field of changes incorrect.")
  }
  filterer := changes.Updates[924]
  // Reviewed entry
  e := fin.Entry{
      Name: "Foo",
      CatPayment: fin.NewCatPayment(fin.NewCat("0:5"), 0, false, 3),
      Status: fin.Reviewed}
  verifyFilterer(t, filterer, &e, fin.NewCat("0:5"), "Foo")
  // Unreviewed entry but categorized
  e = fin.Entry{
      Name: "Foo",
      CatPayment: fin.NewCatPayment(fin.NewCat("0:5"), 0, false, 3),
      Status: fin.NotReviewed}
  verifyFilterer(t, filterer, &e, fin.NewCat("0:5"), "Update1")
  // Unreviewed entry and uncategorized
  e = fin.Entry{
      Name: "Foo",
      CatPayment: fin.NewCatPayment(fin.Expense, 0, false, 3),
      Status: fin.NotReviewed}
  verifyFilterer(t, filterer, &e, fin.NewCat("0:73"), "Update1")
}

func verifyFilterer(t *testing.T, f goconsume.FilterFunc, e *fin.Entry, cat fin.Cat, name string) {
  if !f(e) {
    t.Error("Expected filter to succeed.")
  }
  if !e.Reconciled() {
    t.Error("Expected it to be reconciled.")
  }
  if output := e.CatRecByIndex(0).Cat; output != cat {
    t.Errorf("Expected %v, got %v", cat, output)
  }
  if output := e.Name; output != name {
    t.Errorf("Expected %v, got %v", name, output)
  }
}

func newEntry(id int64, date time.Time) *fin.Entry {
  return &fin.Entry{Id: id, Date: date}
}
