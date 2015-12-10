package upload

import (
  "bytes"
  "errors"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/aggregators"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/autoimport/reconcile"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/functional"
  "html/template"
  "io"
  "net/http"
  "path"
  "strconv"
  "strings"
  "time"
)

const (
  kMaxUploadSize = 1024 * 1024
  kMaxDays = 7
  kAutoCategorizeLookBack = 1000
)

var (
  kUploadTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>{{.Account.Name}} Import Entries</h2>
{{if .Error}}
  <span class="error">{{.Error}}</span>
{{end}}
<form method="post" enctype="multipart/form-data">
  <table>
    <tr>
      <td>QFX file: </td>
      <td><input type="file" name="contents"></td>
    </tr>
    <tr>
      <td>Start Date (YYYYmmdd): </td>
      <td><input type="text" name="sd" value="{{.StartDate}}"></td>
    </tr>
  </table>
  <table>
    <tr>
      <td><input type="submit" name="upload" value="Upload"></td>
      <td><input type="submit" name="cancel" value="Cancel"></td>
    </tr>
  </table>
</form>
</body>
</html>`

kConfirmTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>{{.Account.Name}} Import Entries</h2>
<form method="post">
  <table>
    <tr>
      <td>New entries: </td>
      <td>{{.NewCount}}</td>
    </tr>
    <tr>
      <td>Existing entries: </td>
      <td>{{.ExistingCount}}</td>
    </tr>
    <tr>
      <td colspan=2>&nbsp;</td>
    </tr>
    <tr>
      <td>Balance: </td>
      <td>{{FormatUSD .Balance}}</td>
    </tr>
    <tr>
      <td>Reconciled Balance: </td>
      <td>{{FormatUSD .RBalance}}</td>
    </tr>
  </table>
  <table>
    <tr>
      <td><input type="submit" name="upload" value="Confirm"></td>
      <td><input type="submit" name="cancel" value="Cancel"></td>
    </tr>
  </table>
</form>
</body>
</html>`
)

var (
  kUploadTemplate *template.Template
  kConfirmTemplate *template.Template
)

type Store interface {
  findb.AccountByIdRunner
  findb.EntriesRunner
  findb.DoEntryChangesRunner
  findb.UnreconciledEntriesRunner
  findb.UpdateAccountImportSDRunner
}

type Handler struct {
  Doer db.Doer
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  store := session.Store.(Store)
  acctId, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
  userSession := common.GetUserSession(r)
  batch := userSession.Batch(acctId)
  if batch == nil {
    h.serveUploadPage(w, r, acctId, store, session.Uploaders)
  } else {
    h.serveConfirmPage(w, r, acctId, batch, store)
  }
}

func (h *Handler) serveConfirmPage(w http.ResponseWriter, r *http.Request, acctId int64, batch autoimport.Batch, store Store) {
  if r.Method == "GET" {
    account := fin.Account{}
    unreconciled := make(reconcile.ByAmountCheckNo)
    err := store.UnreconciledEntries(
        nil,
        acctId,
        &account, 
        consumers.FromEntryAggregator(unreconciled))
    if err != nil {
      http_util.ReportError(w, "A database error happened fetching unreconciled entries", err)
      return
    }
    batchEntries := batch.Entries()
    reconcile.New(batchEntries).Reconcile(unreconciled, kMaxDays)
    http_util.WriteTemplate(
        w,
        kConfirmTemplate,
        computeConfirmView(&account, batchEntries))
  } else {
    if !http_util.HasParam(r.Form, "cancel") {
      categorizerBuilder := aggregators.NewByNameCategorizerBuilder(4, 2)
      // If this fails, we can carry on. We just won't get autocategorization
      store.Entries(
          nil, nil, functional.ModifyConsumer(
              consumers.FromEntryAggregator(categorizerBuilder),
              func(s functional.Stream) functional.Stream {
                return functional.Slice(s, 0, kAutoCategorizeLookBack)
              }))
      categorizer := categorizerBuilder.Build()
      err := h.Doer.Do(func(t db.Transaction) (err error) {
        batch, err = batch.SkipProcessed(t)
        if err != nil {
          return
        }
        if batch.Len() == 0 {
          return
        }
        unreconciled := make(reconcile.ByAmountCheckNo)
        err = store.UnreconciledEntries(
            t,
            acctId,
            nil, 
            consumers.FromEntryAggregator(unreconciled))
        if err != nil {
          return
        }
        batchEntries := batch.Entries()
        for _, v := range batchEntries {
          categorizer.Categorize(v)
        }
        reconcile.New(batchEntries).Reconcile(unreconciled, kMaxDays)
        err = store.DoEntryChanges(t, reconcile.GetChanges(batchEntries))
        if err != nil {
          return
        }
        return batch.MarkProcessed(t)
      })
      if err != nil {
        http_util.ReportError(w, "A database error happened importing entries", err)
        return
      }
    }
    userSession := common.GetUserSession(r)
    userSession.SetBatch(acctId, nil)
    userSession.Save(r, w)
    accountLinker := common.AccountLinker{}
    http_util.Redirect(w, r, accountLinker.AccountLink(acctId).String())
  }
}

func (h *Handler) serveUploadPage(
    w http.ResponseWriter, r *http.Request, acctId int64,
    store Store, uploaders map[string]autoimport.Loader) {
  account := fin.Account{}
  err := store.AccountById(nil, acctId, &account)
  if err != nil {
    http_util.ReportError(w, "Error reading account from database.", err)
    return
  }
  if r.Method == "GET" {
    view := &view{Account: &account, StartDate: account.ImportSD.Format(date_util.YMDFormat)}
    showView(w, view, nil)
  } else {
    sdStr := ""
    qfxFile := bytes.Buffer{}
    var fileTooLarge bool
    var loader autoimport.Loader
    reader, err := r.MultipartReader()
    if err != nil {
      http_util.ReportError(w, "Error reading multipart form", err)
    }
    for part, err := reader.NextPart(); err == nil; part, err = reader.NextPart() {
      if part.FormName() == "sd" {
        buffer := bytes.Buffer{}
        _, err = buffer.ReadFrom(part)
        if err != nil {
          http_util.ReportError(w, "Error reading sd", err)
          return
        }
        sdStr = buffer.String()
      } else if part.FormName() == "contents" {
        loader = uploaders[fileExtension(part.FileName())]
        limitedReader := io.LimitedReader{R: part, N: kMaxUploadSize}
        qfxFile.ReadFrom(&limitedReader)
        fileTooLarge = limitedReader.N == 0
      } else if part.FormName() =="cancel" {
        accountLinker := common.AccountLinker{}
        http_util.Redirect(w, r, accountLinker.AccountLink(acctId).String())
        return
      }
    } 
    view := &view{Account: &account, StartDate: sdStr}
    sd, err := time.Parse(date_util.YMDFormat, sdStr)
    if err != nil {
      showView(w, view, errors.New("Start date must be in yyyyMMdd format."))
      return
    }
    store.UpdateAccountImportSD(nil, acctId, sd)
    if fileTooLarge {
      showView(w, view, errors.New("File too large."))
      return
    }
    if qfxFile.Len() == 0 {
      showView(w, view, errors.New("Please select a file."))
      return
    }
    if loader == nil {
      showView(w, view, errors.New("File extension not recognized."))
      return
    }
    batch, err := loader.Load(acctId, "", &qfxFile, sd)
    if err != nil {
      showView(w, view, err)
      return
    }
    batch, err = batch.SkipProcessed(nil)
    if err != nil {
      http_util.ReportError(w, "Error skipping already processed entries.", err)
      return
    }
    if batch.Len() == 0 {
      showView(w, view, errors.New("No new entries to process."))
      return
    }
    userSession := common.GetUserSession(r)
    userSession.SetBatch(acctId, batch)
    userSession.Save(r, w)
    http_util.Redirect(w, r, r.URL.String())
  }
}

func showView(w http.ResponseWriter, v *view, err error) {
  v.Error = err
  http_util.WriteTemplate(w, kUploadTemplate, v)
}

type view struct {
  Account *fin.Account
  StartDate string
  Error error
}

type confirmView struct {
  Account *fin.Account
  NewCount int
  ExistingCount int
  Balance int64
  RBalance int64
}

func computeConfirmView(
    account *fin.Account, batchEntries []*fin.Entry) *confirmView {
  result := &confirmView{Account: account, Balance: account.Balance, RBalance: account.RBalance}
  for _, v := range batchEntries {
    total := v.Total()
    if v.Id == 0 {
      result.NewCount++
      result.Balance += total
    } else {
      result.ExistingCount++
    }
    result.RBalance += v.Total()
  }
  return result
}

func fileExtension(filename string) string {
  return strings.ToLower(path.Ext(filename))
}

func init() {
  kUploadTemplate = common.NewTemplate("upload", kUploadTemplateSpec)
  kConfirmTemplate = common.NewTemplate("upload_confirm", kConfirmTemplateSpec)
}
