package leftnav

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "html/template"
  "net/http"
  "net/url"
)

var (
  kTemplateSpec = `
<html>
<body>
Accounts:
<ul>
{{with $top := .}}
  {{range .ActiveAccountDetails}}
    <li><a href="{{$top.AccountLink .Id}}" target="main">{{.Name}}</a></li>
  {{end}}
{{end}}
</ul>
<br>
<a href="{{.ReportUrl}}" target="main">Reports</a><br>
<a href="{{.TrendUrl}}" target="main">Trends</a><br>
<br>
<a href="/fin/list" target="main">Search</a><br>
<a href="/fin/unreviewed" target="main">Review</a><br>
<a href="/fin/catedit" target="main">Manage Categories</a><br>
<br>
<a href="/fin/chpasswd" target="main">Change Password</a><br>
<a href="/fin/logout" target="_top">Sign out</a>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Cdc categoriesdb.Getter
  Clock date_util.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  cds, err := h.Cdc.Get(nil)
  if err != nil {
    http_util.ReportError(w, "Database error", err)
    return
  }
  now := date_util.TimeToDate(h.Clock.Now())
  // Include today!
  now = now.AddDate(0, 0, 1)
  oneMonthAgo := now.AddDate(0, -1, 0)
  oneYearAgo := now.AddDate(-1, 0, 0)
  http_util.WriteTemplate(w, kTemplate, &view{
      CatDetailStore: cds,
      ReportUrl: http_util.NewUrl(
          "/fin/report",
          "sd", oneMonthAgo.Format(date_util.YMDFormat),
          "ed", now.Format(date_util.YMDFormat)),
      TrendUrl: http_util.NewUrl(
          "/fin/trends",
          "sd", oneYearAgo.Format(date_util.YMDFormat),
          "ed", now.Format(date_util.YMDFormat))})
      
}

type view struct {
  common.AccountLinker
  categories.CatDetailStore
  ReportUrl *url.URL
  TrendUrl *url.URL
}

func init() {
  kTemplate = common.NewTemplate("leftnav", kTemplateSpec)
}