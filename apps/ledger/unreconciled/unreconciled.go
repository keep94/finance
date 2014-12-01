package unreconciled

import (
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "html/template"
  "net/http"
  "strconv"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>{{.Account.Name}}</h2>    
<a href="#" onclick="document.forms[0].edit_id.value=-1; document.forms[0].submit()">New Entry</a>&nbsp;
<a href="#" onclick="document.forms[0].edit_id.value=-2; document.forms[0].submit()">Normal View</a>
<br><br>
Balance: {{FormatUSD .Account.Balance}}&nbsp;&nbsp;&nbsp;&nbsp;Reconciled: {{FormatUSD .Account.RBalance}}
<br><br>
<form method="post">
<input type="hidden" name="edit_id" value="">
{{if .Values}}
<input type="submit" value="Reconcile">
  <table>
    <tr>
      <td>Reconciled</td>
      <td>Date</td>
      <td>Category</td>
      <td>Name</td>
      <td>Amount</td>
    </tr>
{{with $top := .}}
  {{range .Values}}
    <tr class="lineitem">
      <td><input type="checkbox" name="id" value="{{.Id}}"></td>
      <td>{{FormatDate .Date}}</td>
      <td>{{$top.CatName .CatPayment}}</a></td>
      <td><a href="#" onclick="document.forms[0].edit_id.value={{.Id}}; document.forms[0].submit()">{{.Name}}</td>
      <td align=right>{{FormatUSD .Total}}</td>
    </tr>
    <tr>
      <td>{{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}</td>
      <td colspan=4>{{.Desc}}</td>
    </tr>
  {{end}}
{{end}}
  </table>
<input type="submit" value="Reconcile">
{{else}}
No unreconciled entries.
{{end}}
</form>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Store interface {
  findb.UnreconciledEntriesRunner
  findb.DoEntryChangesRunner
}

type Handler struct {
  Store Store
  Cdc categoriesdb.Getter
  Doer db.Doer
  PageSize int
}


func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  acctId, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
  if r.Method == "POST" {
    editId, _ := strconv.ParseInt(r.Form.Get("edit_id"), 10, 64)
    reconciler := functional.NewFilterer(func(ptr interface{}) error {
      p := ptr.(*fin.Entry)
      if p.Reconcile(acctId) {
        return nil
      }
      return functional.Skipped
    })
    ids := r.Form["id"]
    updates := make(map[int64]functional.Filterer, len(ids))
    for _, idStr := range ids {
      id, _ := strconv.ParseInt(idStr, 10, 64)
      updates[id] = reconciler
    }
    h.Store.DoEntryChanges(nil, &findb.EntryChanges{Updates: updates})
    if editId != 0 {
      entryLinker := common.EntryLinker{r.URL}
      accountLinker := common.AccountLinker{}
      if editId == -2 {
        http_util.Redirect(
            w,
            r,
            accountLinker.AccountLink(acctId).String())
      }else if editId == -1 {
        http_util.Redirect(w, r, entryLinker.NewEntryLink(acctId).String())
      } else {
        http_util.Redirect(w, r, entryLinker.EntryLink(editId).String())
      }
      return
    }
  }
  cds := categories.CatDetailStore{}
  buffer := consumers.NewEntryBuffer(h.PageSize)
  account := fin.Account{}
  err := h.Doer.Do(func(t db.Transaction) (err error) {
    cds, _ = h.Cdc.Get(t)
    return h.Store.UnreconciledEntries(t, acctId, &account, buffer)
  })
  if err == findb.NoSuchId {
    fmt.Fprintln(w, "No such account.")
    return
  }
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          buffer.Buffer,
          common.CatDisplayer{cds},
          common.EntryLinker{r.URL},
          &account})
}

type view struct {
  *consume.Buffer
  common.CatDisplayer
  common.EntryLinker
  Account *fin.Account
}

func init() {
  kTemplate = common.NewTemplate("unreconciled", kTemplateSpec)
}
