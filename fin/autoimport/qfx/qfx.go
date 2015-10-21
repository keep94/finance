// Package qfx provides processing of QFX files
package qfx

import (
  "bytes"
  "errors"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
  "github.com/keep94/gofunctional3/functional"
  "io"
  "regexp"
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

var (
  kQFXHeaderPattern = regexp.MustCompile(`^\s*[A-Z]+:[A-Z0-9]+\s*$`)
  kXMLTagPattern = regexp.MustCompile(`</?[A-Z]+>`)
)

// byXMLToken returns a stream of [2]string given a QFX XML body.
// The first string is the XML open or close tag; the second string is
// the contents. The caller must call Close() on the returned stream.
func byXMLToken(contents []byte) functional.Stream {
  return functional.NewGenerator(func(e functional.Emitter) error {
    var ptr interface{}
    var opened bool
    if ptr, opened = e.EmitPtr(); !opened {
      return nil
    }
    allTagIndexes := kXMLTagPattern.FindAllIndex(contents, -1)
    tagCount := len(allTagIndexes)
    if tagCount == 0 {
      functional.WaitForClose(e)
      return nil
    }
    for i := 0; i < tagCount - 1; i++ {
      tagSlice := ptr.([]string)
      tagSlice[0] = string(contents[allTagIndexes[i][0]:allTagIndexes[i][1]])
      tagSlice[1] = string(contents[allTagIndexes[i][1]:allTagIndexes[i+1][0]])
      if ptr, opened = e.Return(nil); !opened {
        return nil
      }
    }
    tagSlice := ptr.([]string)
    tagSlice[0] = string(contents[allTagIndexes[tagCount - 1][0]:allTagIndexes[tagCount - 1][1]])
    tagSlice[1] = string(contents[allTagIndexes[tagCount - 1][1]:])
    e.Return(nil)
    functional.WaitForClose(e)
    return nil
  })
}

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

  // skip over QFX headers for now
  for err = fileStream.Next(&line); err == nil; err = fileStream.Next(&line) {
    if !kQFXHeaderPattern.MatchString(line) {
      break
    }
  }

  // Load the XML body into this buffer
  var qfxContents bytes.Buffer

  for ; err == nil; err = fileStream.Next(&line) {
    line = strings.TrimSpace(line)
    qfxContents.Write([]byte(line))
  }

  // Return any errors from reading the file.
  if (err != functional.Done) {
    return nil, err
  }

  // We break the XML body into a stream of tags and contents.
  tagStream := byXMLToken(qfxContents.Bytes())
  defer tagStream.Close()

  qe := &qfxEntry{}
  var result []*qfxEntry
  var tagAndContents [2]string
  for err = tagStream.Next(tagAndContents[:]); err == nil; err = tagStream.Next(tagAndContents[:]) {
    tag := tagAndContents[0]
    contents := tagAndContents[1]
    if tag == kDtPosted {
      qe.entry.Date, err = parseQFXDate(contents)
      if err != nil {
        return nil, err
      }
    } else if tag == kName {
      qe.entry.Name = strings.Replace(contents, "&amp;", "&", -1)
    } else if tag == kCheckNum {
      qe.entry.CheckNo = contents
    } else if tag == kTrnAmt {
      var amt int64
      amt, err = fin.ParseUSD(contents)
      if err != nil {
        return nil, err
      }
      qe.entry.CatPayment = fin.NewCatPayment(fin.Expense, -amt, true, accountId)
    } else if tag == kFitId {
      qe.fitId = contents
    } else if tag == kStmtTrnClose {
      // No meaningful contents with this closing tag. This closing tag
      // means that we are done with an entry.
      if !qe.entry.Date.Before(startDate) {
        result = append(result, qe)
      }
      qe = &qfxEntry{}
    }
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
