package totals

import (
	"github.com/keep94/finance/apps/ledger/common"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/toolbox/http_util"
	"html/template"
	"net/http"
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
</div>
</body>
</html>`
)

var (
	kTemplate *template.Template
)

type Handler struct {
	Store  findb.ActiveAccountsRunner
	LN     *common.LeftNav
	Global *common.Global
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	leftnav := h.LN.Generate(w, r, common.SelectTotals())
	if leftnav == "" {
		return
	}
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
		Total:    total,
		LeftNav:  leftnav,
		Global:   h.Global,
	})
}

type view struct {
	common.AccountLinker
	Accounts []*fin.Account
	Total    int64
	LeftNav  template.HTML
	Global   *common.Global
}

func init() {
	kTemplate = common.NewTemplate("totals", kTemplateSpec)
}
