package totals

import (
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
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
<h2>Totals</h2>
Total: {{FormatUSD .Total}}<br><br>
<table border=1>
  <tr>
    <td>Account</td>
    <td>Total</td>
  </td>
{{with $top := .}}
  {{range .Accounts}}
    <tr>
      <td><a href="{{$top.AccountLink .Id}}">{{.Name}}</a></td>
      <td align="right">{{FormatUSD .Balance}}</td>
    </tr>
  {{end}}
{{end}}
</table>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Store findb.ActiveAccountsRunner
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  accounts, err := h.Store.ActiveAccounts(nil)
  if err != nil {
    http_util.ReportError(w, "Database error", err)
    return
  }
  var total int64
  for _, account := range accounts {
    total += account.Balance
  }
  http_util.WriteTemplate(w, kTemplate, &view{
      Accounts: accounts,
      Total: total,
  })
}

type view struct {
  common.AccountLinker
  Accounts []*fin.Account
  Total int64
}

func init() {
  kTemplate = common.NewTemplate("totals", kTemplateSpec)
}
