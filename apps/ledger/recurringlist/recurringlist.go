package recurringlist

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/goconsume"
  "html/template"
  "net/http"
  "strconv"
  "time"
)

const (
  kRecurringList = "recurringlist"
)

var (
  kTemplateSpec = `
<html>
<head>
  <title>{{.Global.Title}}</title>
  {{if .Global.Icon}}
    <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
  {{end}}
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
{{.LeftNav}}
<div class="main">
<h2>{{.AccountName}}</h2>
{{if .EntriesToAddCount}}
  <form method="POST">
    <input type="hidden" name="xsrf" value="{{.Xsrf}}">
    Today is <b>{{FormatDate .Today}}</b>.<br>
    Apply ALL recurring entries which will create {{.EntriesToAddCount}} new entries?
    <input type="submit" value="YES">
  </form>
{{end}}
<a href="{{.NewEntryLink .AccountId}}">New Recurring Entry</a>
{{if .AccountId}}
<a href="{{.AccountLink .AccountId}}">Normal View</a><br><br>
{{end}}
{{if .Error}}
  <span class="error">{{.Error.Error}}</span>
{{end}}
{{if .Message}}
  <font color="#006600"><b>{{.Message}}</b></font>
{{end}}
{{with $top := .}}
  <table>
    <tr>
      <td>Date</td>
      <td>Category</td>
      <td>Name</td>
      <td>Amount</td>
      <td>Account</td>
      <td>&nbsp;</td>
    </tr>
  {{range .Values}}
      <tr class="lineitem">
        <td>{{FormatDate .Date}}</td>
        <td>{{$top.CatName .CatPayment}}</td>
        <td><a href="{{$top.EntryLink .Id}}">{{.Name}}</a></td>
        <td align=right>{{FormatUSD .Total}}</td>
        <td>{{$top.AcctName .CatPayment}}</td>
        <td rowspan="2" bgcolor="#FFFFFF">
          <form method="POST">
            <input type="hidden" name="xsrf" value="{{$top.Xsrf}}">
            <input type="hidden" name="rid" value="{{.Id}}">
            <input type="submit" name="skip" value="Skip">
            <input type="submit" name="apply" value="Apply">
          </form>
        </td>
      </tr>
      <tr>
        <td>
          {{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}
        </td>
        <td>Every {{.Period.Count}} {{.Period.Unit}}</td>
        <td>{{$top.NumLeft .NumLeft}}</td>
        <td colspan="2">{{.Desc}}</td>
      </tr>
  {{end}}
  </table>
{{end}}
  <br>
  </div>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Store interface {
  findb.RecurringEntryByIdRunner
  findb.RecurringEntriesApplier
}

type Handler struct {
  Cdc categoriesdb.Getter
  Doer db.Doer
  Clock date_util.Clock
  LN *common.LeftNav
  Global *common.Global
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  acctId, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
  selecter := common.SelectRecurring()
  if acctId > 0 {
    selecter = common.SelectAccount(acctId)
  }
  leftnav := h.LN.Generate(w, r, selecter)
  if leftnav == "" {
    return
  }
  session := common.GetUserSession(r)
  store := session.Store.(Store)
  var postErr error
  var message string
  if r.Method == "POST" {
    rid, _ := strconv.ParseInt(r.Form.Get("rid"), 10, 64)
    if !common.VerifyXsrfToken(r, kRecurringList) {
      postErr = common.ErrXsrf
    } else if rid == 0 {
      message, postErr = h.applyRecurringEntries(store, acctId)
    } else if http_util.HasParam(r.Form, "skip") {
      message, postErr = h.skipEntry(store, rid)
    } else if http_util.HasParam(r.Form, "apply") {
      message, postErr = h.applyEntry(store, rid)
    }
  }
  cds, _ := h.Cdc.Get(nil)
  var entries []*fin.RecurringEntry
  consumer := goconsume.AppendPtrsTo(&entries)
  if acctId > 0 {
    consumer = goconsume.Filter(
        consumer,
        func(ptr interface{}) bool {
          p := ptr.(*fin.RecurringEntry)
          return p.WithPayment(acctId)
        })
  }
  err := store.RecurringEntries(nil, consumer)
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  currentDate := date_util.TimeToDate(h.Clock.Now())
  count, err := findb.ApplyRecurringEntriesDryRun(
      nil, store, acctId, currentDate)
  if err != nil {
    http_util.ReportError(
        w, "Error doing apply recurring entries dry run.", err)
    return
  }
  var accountName string
  if acctId != 0 {
    accountName = cds.AccountDetailById(acctId).Name()
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          CatDisplayer: common.CatDisplayer{cds},
          RecurringEntryLinker: common.RecurringEntryLinker{
              URL: r.URL,
              Sel: selecter},
          Values: entries,
          AccountId: acctId,
          Today: currentDate,
          EntriesToAddCount: count,
          AccountName: accountName,
          Error: postErr,
          Xsrf: common.NewXsrfToken(r, kRecurringList),
          Message: message,
          LeftNav: leftnav,
          Global: h.Global})
}

func (h *Handler) applyRecurringEntries(
    store findb.RecurringEntriesApplier,
    acctId int64) (message string, err error) {
  var count int
  err = h.Doer.Do(func(t db.Transaction) error {
    var err error
    count, err = findb.ApplyRecurringEntries(
        t, store, acctId, date_util.TimeToDate(h.Clock.Now()))
    return err
  })
  if err == nil {
    message = fmt.Sprintf("%d new entries added.", count)
  }
  return
}

func (h *Handler) skipEntry(
    store findb.RecurringEntrySkipper,
    rid int64) (message string, err error) {
  var skipped bool
  err = h.Doer.Do(func(t db.Transaction) error {
    var err error
    skipped, err = findb.SkipRecurringEntry(t, store, rid)
    return err
  })
  if skipped {
    message = "Recurring entry skipped."
  } else if err == nil {
    err = errors.New("Cannot advance. None left.")
  }
  return
}

func (h *Handler) applyEntry(
    store findb.RecurringEntryApplier,
    rid int64) (message string, err error) {
  var applied bool
  err = h.Doer.Do(func(t db.Transaction) error {
    var err error
    applied, err = findb.ApplyRecurringEntry(t, store, rid)
    return err
  })
  if applied {
    message = "Recurring entry aplied."
  } else if err == nil {
    err = errors.New("Cannot advance. None left.")
  }
  return
}

type view struct {
  common.CatDisplayer
  common.RecurringEntryLinker
  common.AccountLinker
  Values []*fin.RecurringEntry
  AccountId int64
  Today time.Time
  EntriesToAddCount int
  Error error
  Message string
  Xsrf string
  AccountName string
  LeftNav template.HTML
  Global *common.Global
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
