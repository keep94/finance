package recurringlist

import (
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "html/template"
  "net/http"
  "strconv"
  "time"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<a href="{{.NewEntryLink .AccountId}}">New Recurring Entry</a><br><br>
{{if .EntriesToAddCount}}
  <form method="POST" action="/fin/applyrecurring">
    Today is <b>{{FormatDate .Today}}</b>.<br>
    Apply ALL recurring entries which will create {{.EntriesToAddCount}} new entries?
    <input type="submit" value="YES">
  </form>
{{end}}
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
        <td>Every {{.Period.Count}} {{.Period.Unit}}</td>
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
  Clock date_util.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  acctId, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
  cds, _ := h.Cdc.Get(nil)
  var entries []*fin.RecurringEntry
  consumer := consume.AppendPtrsTo(&entries, nil)
  if acctId > 0 {
    consumer = functional.FilterConsumer(
        consumer,
        functional.NewFilterer(func(ptr interface{}) error {
          p := ptr.(*fin.RecurringEntry)
          if !p.WithPayment(acctId) {
            return functional.Skipped
          }
          return nil
        }))
  }
  err := h.Store.RecurringEntries(nil, consumer)
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  currentDate := date_util.TimeToDate(h.Clock.Now())
  count, err := findb.ApplyRecurringEntriesDryRun(nil, h.Store, currentDate)
  if err != nil {
    http_util.ReportError(
        w, "Error doing apply recurring entries dry run.", err)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          CatDisplayer: common.CatDisplayer{cds},
          RecurringEntryLinker: common.RecurringEntryLinker{r.URL},
          Values: entries,
          AccountId: acctId,
          Today: currentDate,
          EntriesToAddCount: count})
}

type view struct {
  common.CatDisplayer
  common.RecurringEntryLinker
  Values []*fin.RecurringEntry
  AccountId int64
  Today time.Time
  EntriesToAddCount int
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
