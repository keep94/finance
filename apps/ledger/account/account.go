package account

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
  "html/template"
  "net/http"
  "net/url"
  "strconv"
)

const (
  kPageParam = "pageNo"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>{{.Account.Name}}</h2>    
{{with $top := .}}
<a href="{{.NewEntryLink .Account.Id}}">New Entry</a>&nbsp;
<a href="{{.UploadLink .Account.Id}}">Import Entries</a>&nbsp;
<a href="{{.RecurringLink .Account.Id}}">Recurring Entries</a>&nbsp;
{{if .Account.HasUnreconciled}}
<a href="{{.UnreconciledLink .Account.Id}}">Unreconciled</a>
{{end}}
<br><br>
Page: {{.DisplayPageNo}}
{{if .PageNo}}<a href="{{.PrevPageLink}}">&lt;</a>{{end}}
{{if .End}}&nbsp;{{else}}<a href="{{.NextPageLink}}">&gt;</a>{{end}}
<br><br>
  <table>
    <tr>
      <td>Date</td>
      <td>Category</td>
      <td>Name</td>
      <td>Amount</td>
      <td>Balance</td>
    </tr>
  {{range .Values}}
      <tr class="lineitem">
        <td>{{FormatDate .Date}}</td>
        <td>{{range $top.CatLink .CatPayment}}{{if .Link}}<a href="{{.Link}}">{{.Text}}</a>{{else}}{{.Text}}{{end}}{{end}}</td>
        <td><a href="{{$top.EntryLink .Id}}">{{.Name}}</td>
        <td align=right>{{FormatUSD .Total}}</td>
        <td align=right>{{FormatUSD .Balance}}</td>
      </tr>
      <tr>
        <td>{{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}</td>
        <td colspan=4>{{.Desc}}</td>
      </tr>
  {{end}}
  </table>
  <br>
Page: {{.DisplayPageNo}}
{{if .PageNo}}<a href="{{.PrevPageLink}}">&lt;</a>{{end}}
{{if .End}}&nbsp;{{else}}<a href="{{.NextPageLink}}">&gt;</a>{{end}}
</body>
</html>
{{end}}`
)

var (
  kTemplate *template.Template
)

var (
  kListEntriesUrl = http_util.NewUrl("/fin/list")
)

type Handler struct {
  Doer db.Doer
  Store findb.EntriesByAccountIdRunner
  Cdc categoriesdb.Getter
  PageSize int
  Links bool
}


func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  id, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
  pageNo, _ := strconv.Atoi(r.Form.Get(kPageParam))
  cds := categories.CatDetailStore{}
  ebpb := consumers.NewEntryBalancePageBuffer(h.PageSize, pageNo)
  account := fin.Account{}
  err := h.Doer.Do(func(t db.Transaction) (err error) {
    cds, err = h.Cdc.Get(t)
    if err != nil {
      return
    }
    return h.Store.EntriesByAccountId(t, id, &account, ebpb)
  })
  if err == findb.NoSuchId {
    fmt.Fprintln(w, "No such account.")
    return
  }
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  var listEntriesUrl *url.URL
  if h.Links {
    listEntriesUrl = kListEntriesUrl
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          Pager: http_util.Pager{
              PageBuffer: ebpb.PageBuffer,
              URL: r.URL,
              PageNoParam: kPageParam},
          CatLinker: common.CatLinker{Cds: cds, ListEntries: listEntriesUrl},
          EntryLinker: common.EntryLinker{r.URL},
          Account: accountWrapper{&account}})
}

type view struct {
  http_util.Pager
  common.CatLinker
  common.AccountLinker
  common.EntryLinker
  Account accountWrapper
}

type accountWrapper struct {
  *fin.Account
}

func (a accountWrapper) HasUnreconciled() bool {
  return a.RCount < a.Count
}

func init() {
  kTemplate = common.NewTemplate("account", kTemplateSpec)
}
