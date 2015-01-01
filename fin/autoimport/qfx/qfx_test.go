package qfx

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
  "reflect"
  "strings"
  "testing"
)

const kSampleQfx = `
OFXHEADER:100
DATA:OFXSGML
VERSION:102
SECURITY:NONE
ENCODING:USASCII
CHARSET:1252
COMPRESSION:NONE
OLDFILEUID:NONE
NEWFILEUID:NONE
<OFX>
<SIGNONMSGSRSV1>
<SONRS>
<STATUS>
<CODE>0
<SEVERITY>INFO
</STATUS>
<DTSERVER>20121115120000[0:GMT]
<LANGUAGE>ENG
<FI>
<ORG>ISC
<FID>10898
</FI>
<INTU.BID>10898
</SONRS>
</SIGNONMSGSRSV1>
<CREDITCARDMSGSRSV1>
<CCSTMTTRNRS>
<TRNUID>1
<STATUS>
<CODE>0
<SEVERITY>INFO
<MESSAGE>Success
</STATUS>
<CCSTMTRS>
<CURDEF>USD
<CCACCTFROM>
<ACCTID>4147202080404005
</CCACCTFROM>
<BANKTRANLIST>
<DTSTART>20121115120000[0:GMT]
<DTEND>20121115120000[0:GMT]
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20121113120000[0:GMT]
<TRNAMT>-109.01
<FITID>10200
<NAME>WHOLEFDS LAT 10155
</STMTTRN>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20121114120000[0:GMT]
<TRNAMT>-100.75
<FITID>10201
<NAME>WHOLEFDS LAT 10155
</STMTTRN>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20121114120000[0:GMT]
<TRNAMT>-57.14
<FITID>10202
<NAME>Amazon.com
</STMTTRN>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20121115120000[0:GMT]
<TRNAMT>-12.12
<FITID>10203
<NAME>safeway
</STMTTRN>
</BANKTRANLIST>
<LEDGERBAL>
<BALAMT>-3392.62
<DTASOF>20121115120000[0:GMT]
</LEDGERBAL>
<AVAILBAL>
<BALAMT>21714.00
<DTASOF>20121115120000[0:GMT]
</AVAILBAL>
</CCSTMTRS>
</CCSTMTTRNRS>
</CREDITCARDMSGSRSV1>
</OFX>`

func TestReadQFX(t *testing.T) {
  r := strings.NewReader(kSampleQfx)
  var loader autoimport.Loader
  loader = QFXLoader{make(storeType)}
  batch, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  entries := batch.Entries()
  cp := fin.CatPaymentBuilder{}
  expectedEntries := []*fin.Entry {
      {
          Date: date_util.YMD(2012, 11, 14),
          Name: "WHOLEFDS LAT 10155",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 10075}).Build()},
      {
          Date: date_util.YMD(2012, 11, 14),
          Name: "Amazon.com",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 5714}).Build()},
      {
          Date: date_util.YMD(2012, 11, 15),
          Name: "safeway",
          CatPayment: cp.SetPaymentId(3).SetReconciled(true).AddCatRec(&fin.CatRec{A: 1212}).Build()}}
  if !reflect.DeepEqual(expectedEntries, entries) {
    t.Errorf("Expected %v, got %v", expectedEntries, entries)
  }
}

func TestSkipProcessed(t *testing.T) {
  r := strings.NewReader(kSampleQfx)
  store := make(storeType)
  var loader autoimport.Loader
  loader = QFXLoader{store}
  batch, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  if output := batch.Len(); output != 3 {
    t.Errorf("Expected 3, got %v", output)
  }
  if output := len(batch.Entries()); output != 3 {
    t.Errorf("Expected 3, got %v", output)
  }
  // Pretend a fitId that happens to match one of our entries gets processed in
  // another account. This should not affect our batch.
  store.Add(nil, 4, qfxdb.FitIdSet{"10201": true})

  // SkipProcessed should return the same batch 
  newBatch, _ := batch.SkipProcessed(nil)
  if newBatch != batch {
    t.Error("Expected newBatch to be the same as batch.")
  }

  // Pretend one of the entries in the batch got processed in another thread.
  // Our batch should have one fewer entries.
  store.Add(nil, 3, qfxdb.FitIdSet{"10201": true})
  newBatch, _ = batch.SkipProcessed(nil)
  if output := len(newBatch.Entries()); output != 2 {
    t.Errorf("Expected 2, got %v", output)
  }
  // But batches should be immutable.
  if output := len(batch.Entries()); output != 3 {
    t.Errorf("Expected 3, got %v", output)
  }
}

func TestMarkProcessed(t *testing.T) {
  r := strings.NewReader(kSampleQfx)
  loader := QFXLoader{make(storeType)}
  batch, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  batch.MarkProcessed(nil)
  r = strings.NewReader(kSampleQfx)
  newBatch, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  newBatch, _ = newBatch.SkipProcessed(nil)
  if output := len(newBatch.Entries()); output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func TestBatchImmutability(t *testing.T) {
  r := strings.NewReader(kSampleQfx)
  loader := QFXLoader{make(storeType)}
  batch, err := loader.Load(3, "", r, date_util.YMD(2012, 11, 14))
  if err != nil {
    t.Errorf("Got error %v", err)
    return
  }
  entries1 := batch.Entries()
  entries2 := batch.Entries()
  if entries1[0] == entries2[0] {
    t.Errorf("Expect Entries() to return copies of fin.Entry")
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