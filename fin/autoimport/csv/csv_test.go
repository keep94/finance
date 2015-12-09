package csv_test

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/csv"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
  "reflect"
  "strings"
  "testing"
)

const kPaypalCsv = `
Date, Time, Time Zone, Name, Type, Status, Amount, Receipt ID, Balance,
"12/6/2015","07:59:04","PST","TrackR, Inc","Express Checkout Payment Sent","Completed","-87.00","","0.00",
"12/6/2015","07:59:04","PST","Bank Account","Add Funds from a Bank Account","Completed","87.00","","87.00",
"9/5/2015","09:15:47","PST","Starbucks Coffee Company","Express Checkout Payment Sent","Completed","-48.10","","0.00",
"9/5/2015","09:15:47","PST","Bank Account","Add Funds from a Bank Account","Completed","48.10","","48.10",
"09/03/2015","17:55:28","PST","Disney Online","Express Checkout Payment Sent","Completed","-46.41","","0.00",
"09/03/2015","17:55:28","PST","Bank Account","Add Funds from a Bank Account","Completed","46.41","","46.41",
"9/2/2015","09:27:09","PST","Disney Online","Express Checkout Payment Sent","Completed","-18.43","","0.00",
"9/2/2015","09:27:09","PST","Bank Account","Add Funds from a Bank Account","Completed","18.43","","18.43",
`

func TestReadBadCsvFile(t *testing.T) {
  r := strings.NewReader("A bad file\nNo CSV things in here\n")
  var loader autoimport.Loader
  loader = csv.CsvLoader{make(storeType)}
  _, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err == nil {
    t.Error("Expected error")
  }
}

func TestReadPaypalCsv(t *testing.T) {
  r := strings.NewReader(kPaypalCsv)
  var loader autoimport.Loader
  loader = csv.CsvLoader{make(storeType)}
  batch, err := loader.Load(3, "", r, date_util.YMD(2015, 9, 3))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  entries := batch.Entries()
  cp := fin.CatPaymentBuilder{}
  expectedEntries := []*fin.Entry {
      {
          Date: date_util.YMD(2015, 12, 6),
          Name: "TrackR, Inc",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 8700}).Build()},
      {
          Date: date_util.YMD(2015, 9, 5),
          Name: "Starbucks Coffee Company",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 4810}).Build()},
      {
          Date: date_util.YMD(2015, 9, 3),
          Name: "Disney Online",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 4641}).Build()}}
  if !reflect.DeepEqual(expectedEntries, entries) {
    t.Errorf("Expected %v, got %v", expectedEntries, entries)
  }
}

func TestMarkProcessed(t *testing.T) {
  r := strings.NewReader(kPaypalCsv)
  loader := csv.CsvLoader{make(storeType)}
  batch, err := loader.Load(3, "", r, date_util.YMD(2015, 9, 3))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  batch.MarkProcessed(nil)
  r = strings.NewReader(kPaypalCsv)
  newBatch, err := loader.Load(3, "", r, date_util.YMD(2015, 9, 2))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  newBatch, _ = newBatch.SkipProcessed(nil)
  if output := len(newBatch.Entries()); output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
}

type storeType map[int64]map[string]bool

func (s storeType) Add(t db.Transaction, accountId int64, fitIds qfxdb.FitIdSet) error {
  if s[accountId] == nil {
    s[accountId] = make(map[string]bool)
  }
  for fitId, ok := range fitIds {
    if ok {
      s[accountId][fitId] = true
    }
  }
  return nil
}

func (s storeType) Find(t db.Transaction, accountId int64, fitIds qfxdb.FitIdSet) (qfxdb.FitIdSet, error) {
  var result qfxdb.FitIdSet
  for fitId, ok := range fitIds {
    if ok {
      if s[accountId][fitId] {
        if result == nil {
          result = make(qfxdb.FitIdSet)
        }
        result[fitId] = true
      }
    }
  }
  return result, nil
}
