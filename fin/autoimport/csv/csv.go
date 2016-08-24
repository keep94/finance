// Package csv provides processing of csv files
package csv

import (
  gocsv "encoding/csv"
  "errors"
  "fmt"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/qfx"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
  "hash/fnv"
  "io"
  "strconv"
  "time"
)

// CsvLoader implements the autoimport.Loader interface for csv files.
type CsvLoader struct {
  // Store stores which transactions have already been processed.
  Store qfxdb.Store
}

func (c CsvLoader) Load(
    accountId int64,
    bankAccountId string,
    r io.Reader,
    startDate time.Time) (autoimport.Batch, error) {
  reader := gocsv.NewReader(r)
  line, err := reader.Read()
  if err != nil {
    return nil, err
  }
  parseit := fromHeader(line)
  if parseit == nil {
    return nil, errors.New("Unrecognized csv header")
  }
  var result []*qfx.QfxEntry
  for line, err = reader.Read(); err == nil; line, err = reader.Read() {
    var qentry qfx.QfxEntry
    var ok bool
    ok, err = parseit(line, accountId, &qentry.Entry)
    if err != nil {
      return nil, err
    }
    if !ok || qentry.Date.Before(startDate) {
      continue
    }
    qentry.FitId, err = generateFitId(line)
    if err != nil {
      return nil, err
    }
    result = append(result, &qentry)
  }
  if err != io.EOF {
    return nil, err
  }
  return &qfx.QfxBatch{Store: c.Store, AccountId: accountId, QfxEntries: result}, nil
}

func fromNativeHeader(line []string, accountId int64, entry *fin.Entry) (ok bool, err error) {
  entry.Date, err = time.Parse("1/2/2006", line[0])
  if err != nil {
    return
  }
  entry.CheckNo = line[1]
  entry.Name = line[2]
  entry.Desc = line[3]
  var amt int64
  amt, err = fin.ParseUSD(line[4])
  if err != nil {
    return
  }
  entry.CatPayment = fin.NewCatPayment(fin.Expense, -amt, true, accountId)
  ok = true
  return
}

func fromPaypalHeader(line []string, accountId int64, entry *fin.Entry) (ok bool, err error) {
  entry.Date, err = time.Parse("1/2/2006", line[0])
  if err != nil {
    return
  }
  entry.Name = line[3]
  if entry.Name == "Bank Account" {
    return
  }
  var amt int64
  amt, err = fin.ParseUSD(line[6])
  if err != nil {
    return
  }
  entry.CatPayment = fin.NewCatPayment(fin.Expense, -amt, true, accountId)
  ok = true
  return
}

func fromHeader(line []string) func([]string, int64, *fin.Entry) (bool, error) {
  if len(line) == 10 && line[0] == "Date" && line[3] == " Name" && line[6] == " Amount" {
    return fromPaypalHeader
  }
  if len(line) == 5 && line[0] == "Date" && line[1] == "CheckNo" && line[2] == "Name" && line[3] == "Desc" && line[4] == "Amount" {
    return fromNativeHeader
  }
  return nil
}

func generateFitId(line []string) (string, error) {
  h := fnv.New64a()
  s := fmt.Sprintf("%v", line)
  _, err := h.Write(([]byte)(s))
  if err != nil {
    return "", err
  }
  return strconv.FormatUint(h.Sum64(), 10), nil
}
