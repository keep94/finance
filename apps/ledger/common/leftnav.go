package common

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "html/template"
  "net/http"
  "net/url"
  "strings"
)

var (
  kLeftNavTemplateSpec = `
<div class="leftnav">
<b>{{.UserName}}</b><br>
{{.LastLogin}}<br>
<br>
Accounts:
<ul>
{{with $top := .}}
  {{range .ActiveAccountDetails}}
    {{if $top.Account .Id}}
      <li><span class="selected">{{.Name}}</span></li>
    {{else}}
      <li><a href="{{$top.AccountLink .Id}}">{{.Name}}</a></li>
    {{end}}
  {{end}}
{{end}}
</ul>
<br>
{{if .Reports}}
  <span class="selected">Reports</span><br>
{{else}}
  <a href="{{.ReportUrl}}">Reports</a><br>
{{end}}
{{if .Trends}}
  <span class="selected">Trends</span><br>
{{else}}
  <a href="{{.TrendUrl}}">Trends</a><br>
{{end}}
{{if .Totals}}
  <span class="selected">Totals</span><br>
{{else}}
  <a href="/fin/totals">Totals</a><br>
{{end}}
<br>
{{if .Search}}
  <span class="selected">Search</span><br>
{{else}}
  <a href="/fin/list">Search</a><br>
{{end}}
{{if .Unreviewed}}
  <span class="selected">Review</span><br>
{{else}}
  <a href="/fin/unreviewed">Review</a><br>
{{end}}
{{if .Manage}}
  <span class="selected">Manage Categories</span><br>
{{else}}
  <a href="/fin/catedit">Manage Categories</a><br>
{{end}}
{{if .Recurring}}
  <span class="selected">Recurring</span><br>
{{else}}
  <a href="/fin/recurringlist">Recurring</a><br>
{{end}}
{{if .Export}}
  <span class="selected">Export</span><br>
{{else}}
  <a href="/fin/export">Export</a><br>
{{end}}
<br>
{{if .Chpasswd}}
   <span class="selected">Change Password</span><br>
{{else}}
   <a href="/fin/chpasswd">Change Password</a><br>
{{end}}
<a href="/fin/logout">Sign out</a>
<br><br>
</div>`
)

var (
  kLeftNavTemplate *template.Template
)

// Selecter indicates the item to be selected in the left navigation bar.
type Selecter selecter
type selecter interface{}

type accounts int64
type reports int
type trends int
type totals int
type search int
type unreviewed int
type manage int
type recurring int
type export int
type chpasswd int

func SelectAccount(id int64) Selecter { return accounts(id) }
func SelectReports() Selecter { return reports(0) }
func SelectTrends() Selecter { return trends(0) }
func SelectTotals() Selecter { return totals(0) }
func SelectSearch() Selecter { return search(0) }
func SelectUnreviewed() Selecter { return unreviewed(0) }
func SelectManage() Selecter { return manage(0) }
func SelectRecurring() Selecter { return recurring(0) }
func SelectExport() Selecter { return export(0) }
func SelectChpasswd() Selecter { return chpasswd(0) }
func SelectNone() Selecter { return nil }

// LeftNav is for creating the left navigation bar.
type LeftNav struct {
  Cdc categoriesdb.Getter
  Clock date_util.Clock
}

// Generate generates the html for the left navigation bar including the div
// tags. sel indicates which item in the left navigation bar will be selected.
// If Generate can't generate the html, it returns the empty string, writes
// an error message to w, and writes the error to stderr.
func (l *LeftNav) Generate(
    w http.ResponseWriter, r *http.Request, sel Selecter) template.HTML {
  session := GetUserSession(r)
  lastLoginStr := "--"
  lastLogin, ok := session.LastLogin()
  if ok {
    lastLoginStr = lastLogin.Local().Format("Mon 01/02/2006 15:04")
  }
  cds, err := l.Cdc.Get(nil)
  if err != nil {
    http_util.ReportError(w, "Database error", err)
    return ""
  }
  now := date_util.TimeToDate(l.Clock.Now())
  // Include today!
  now = now.AddDate(0, 0, 1)
  oneMonthAgo := now.AddDate(0, -1, 0)
  oneYearAgo := now.AddDate(-1, 0, 0)
  var sb strings.Builder
  http_util.WriteTemplate(&sb, kLeftNavTemplate, &view{
      CatDetailStore: cds,
      ReportUrl: http_util.NewUrl(
          "/fin/report",
          "sd", oneMonthAgo.Format(date_util.YMDFormat),
          "ed", now.Format(date_util.YMDFormat)),
      TrendUrl: http_util.NewUrl(
          "/fin/trends",
          "sd", oneYearAgo.Format(date_util.YMDFormat),
          "ed", now.Format(date_util.YMDFormat)),
      UserName: session.User.Name,
      LastLogin: lastLoginStr,
      sel: sel})
  return template.HTML(sb.String())
}

type view struct {
  AccountLinker
  categories.CatDetailStore
  ReportUrl *url.URL
  TrendUrl *url.URL
  UserName string
  LastLogin string
  sel Selecter
}

func (v *view) Account(id int64) bool { return v.sel == accounts(id) }
func (v *view) Reports() bool { return v.sel == reports(0) }
func (v *view) Trends() bool { return v.sel == trends(0) }
func (v *view) Totals() bool { return v.sel == totals(0) }
func (v *view) Search() bool { return v.sel == search(0) }
func (v *view) Unreviewed() bool { return v.sel == unreviewed(0) }
func (v *view) Manage() bool { return v.sel == manage(0) }
func (v *view) Recurring() bool { return v.sel == recurring(0) }
func (v *view) Export() bool { return v.sel == export(0) }
func (v *view) Chpasswd() bool { return v.sel == chpasswd(0) }

func init() {
  kLeftNavTemplate = NewTemplate("leftnav", kLeftNavTemplateSpec)
}
