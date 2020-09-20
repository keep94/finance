package account

import (
	"fmt"
	"github.com/keep94/finance/apps/ledger/common"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/categories"
	"github.com/keep94/finance/fin/categories/categoriesdb"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/goconsume"
	"github.com/keep94/toolbox/db"
	"github.com/keep94/toolbox/http_util"
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
  <title>{{.Global.Title}}</title>
  {{if .Global.Icon}}
    <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
  {{end}}
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
{{.LeftNav}}
<div class="main">
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
{{end}}
</div>
</body>
</html>`
)

var (
	kTemplate *template.Template
)

var (
	kListEntriesUrl = http_util.NewUrl("/fin/list")
)

type Handler struct {
	Doer     db.Doer
	Store    findb.EntriesByAccountIdRunner
	Cdc      categoriesdb.Getter
	PageSize int
	Links    bool
	LN       *common.LeftNav
	Global   *common.Global
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.ParseInt(r.Form.Get("acctId"), 10, 64)
	selecter := common.SelectAccount(id)
	leftnav := h.LN.Generate(w, r, selecter)
	if leftnav == "" {
		return
	}
	pageNo, _ := strconv.Atoi(r.Form.Get(kPageParam))
	cds := categories.CatDetailStore{}
	var entryBalances []fin.EntryBalance
	var morePages bool
	consumer := goconsume.Page(pageNo, h.PageSize, &entryBalances, &morePages)
	account := fin.Account{}
	err := h.Doer.Do(func(t db.Transaction) (err error) {
		cds, err = h.Cdc.Get(t)
		if err != nil {
			return
		}
		err = h.Store.EntriesByAccountId(t, id, &account, consumer)
		if err != nil {
			return
		}
		consumer.Finalize()
		return
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
			PageBreadCrumb: http_util.PageBreadCrumb{
				URL:         r.URL,
				PageNoParam: kPageParam,
				PageNo:      pageNo,
				End:         !morePages,
			},
			Values:      entryBalances,
			CatLinker:   common.CatLinker{Cds: cds, ListEntries: listEntriesUrl},
			EntryLinker: common.EntryLinker{URL: r.URL, Sel: selecter},
			Account:     accountWrapper{&account},
			LeftNav:     leftnav,
			Global:      h.Global})
}

type view struct {
	http_util.PageBreadCrumb
	common.CatLinker
	common.AccountLinker
	common.EntryLinker
	Account accountWrapper
	Values  []fin.EntryBalance
	LeftNav template.HTML
	Global  *common.Global
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
