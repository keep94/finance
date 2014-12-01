// Package qfx provides processing of QFX files
package qfx

import (
  "errors"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
  "github.com/keep94/gofunctional3/functional"
  "io"
  "strings"
  "time"
)

const (
  kDtPosted = "<DTPOSTED>"
  kTrnAmt = "<TRNAMT>"
  kName = "<NAME>"
  kCheckNum = "<CHECKNUM>"
  kStmtTrnClose = "</STMTTRN>"
  kFitId = "<FITID>"
)

// QFXLoader implements the autoimport.Loader interface for QFX files.
type QFXLoader struct {
  // Store stores which fitIds, unique identifier in QFX files,
  // have already been processed.
  Store qfxdb.Store
}

func (q QFXLoader) Load(
    accountId int64,
    bankAccountId string,
    r io.Reader,
    startDate time.Time) (autoimport.Batch, error) {
  fileStream := functional.ReadLines(r)
  var line string
  var err error
  qe := &qfxEntry{}
  var result []*qfxEntry
  for err = fileStream.Next(&line); err == nil; err = fileStream.Next(&line) {
    line = strings.TrimSpace(line)
    if line == "" {
      continue
    }
    if strings.HasPrefix(line, kDtPosted) {
      qe.entry.Date, err = parseQFXDate(line[len(kDtPosted):])
      if err != nil {
        return nil, err
      }
    } else if strings.HasPrefix(line, kName) {
      qe.entry.Name = strings.Replace(line[len(kName):], "&amp;", "&", -1)
    } else if strings.HasPrefix(line, kCheckNum) {
      qe.entry.CheckNo = line[len(kCheckNum):]
    } else if strings.HasPrefix(line, kTrnAmt) {
      var amt int64
      amt, err = fin.ParseUSD(line[len(kTrnAmt):])
      if err != nil {
        return nil, err
      }
      qe.entry.CatPayment = fin.NewCatPayment(fin.Expense, -amt, true, accountId)
    } else if strings.HasPrefix(line, kFitId) {
      qe.fitId = line[len(kFitId):]
    } else if strings.HasPrefix(line, kStmtTrnClose) {
      if !qe.entry.Date.Before(startDate) {
        result = append(result, qe)
      }
      qe = &qfxEntry{}
    }
  }
  if (err != functional.Done) {
    return nil, err
  }
  return &qfxBatch{store: q.Store, accountId: accountId, qfxEntries: result}, nil
}

type qfxBatch struct {
  store qfxdb.Store
  accountId int64
  qfxEntries []*qfxEntry
}

func (q *qfxBatch) Entries() []*fin.Entry {
  result := make([]*fin.Entry, len(q.qfxEntries))
  for i := range q.qfxEntries {
    e := q.qfxEntries[i].entry
    result[i] = &e
  }
  return result
}

func (q *qfxBatch) Len() int {
  return len(q.qfxEntries)
}

func (q *qfxBatch) SkipProcessed(t db.Transaction) (autoimport.Batch, error) {
  existingFitIds, err := q.store.Find(t, q.accountId, q.toFitIdSet())
  if err != nil {
    return nil, err
  }
  if existingFitIds == nil {
    return q, nil
  }
  result := make([]*qfxEntry, len(q.qfxEntries))
  idx := 0
  for _, qe := range q.qfxEntries {
    if !existingFitIds[qe.fitId] {
      result[idx] = qe
      idx++
    }
  }
  return &qfxBatch{store: q.store, accountId: q.accountId, qfxEntries: result[:idx]}, nil
}
 
func (q *qfxBatch) MarkProcessed(t db.Transaction) error {
  return q.store.Add(t, q.accountId, q.toFitIdSet())
}

func (q *qfxBatch) toFitIdSet() qfxdb.FitIdSet {
  fitIdSet := make(qfxdb.FitIdSet, len(q.qfxEntries))
  for _, qe := range q.qfxEntries {
    fitIdSet[qe.fitId] = true
  }
  return fitIdSet
}

type qfxEntry struct {
  fitId string
  entry fin.Entry
}

func parseQFXDate(s string) (time.Time, error) {
  if len(s) < 8 {
    return time.Time{}, errors.New("Invalid date field in qfx file.")
  }
  return time.Parse(date_util.YMDFormat, s[:8])
}
