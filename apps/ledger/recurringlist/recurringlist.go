package recurringlist

import (
  "fmt"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/consume"
  "html/template"
  "net/http"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<a href="{{.NewEntryLink}}">New Recurring Entry</a>
<br><br>   
{{with $top := .}}
  <table>
    <tr>
      <td>Date</td>
      <td>Category</td>
      <td>Name</td>
      <td>Amount</td>
      <td>Account</td>
    </tr>
  {{range .Values}}
      <tr class="lineitem">
        <td>{{FormatDate .Date}}</td>
        <td>{{$top.CatName .CatPayment}}</td>
        <td><a href="{{$top.EntryLink .Id}}">{{.Name}}</a></td>
        <td align=right>{{FormatUSD .Total}}</td>
        <td>{{$top.AcctName .CatPayment}}</td>
      </tr>
      <tr>
        <td>
          {{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}
        </td>
        <td>Every {{.Period}} months</td>
        <td>{{$top.NumLeft .NumLeft}}</td>
        <td colspan=2>{{.Desc}}</td>
      </tr>
  {{end}}
  </table>
  <br>
</body>
</html>
{{end}}`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Cdc categoriesdb.Getter
  Store findb.RecurringEntriesRunner
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  cds, _ := h.Cdc.Get(nil)
  var entries []*fin.RecurringEntry
  err := h.Store.RecurringEntries(nil, consume.AppendPtrsTo(&entries, nil))
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          common.CatDisplayer{cds},
          common.RecurringEntryLinker{r.URL},
          entries})
}

type view struct {
  common.CatDisplayer
  common.RecurringEntryLinker
  Values []*fin.RecurringEntry
}

func (v *view) NumLeft(numLeft int) string {
  if numLeft < 0 {
    return "unlimited"
  }
  return fmt.Sprintf("%d left", numLeft)
}

func init() {
  kTemplate = common.NewTemplate("recurringlist", kTemplateSpec)
}
