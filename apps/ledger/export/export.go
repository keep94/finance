package export

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/keep94/appcommon/date_util"
	"github.com/keep94/appcommon/http_util"
	"github.com/keep94/finance/apps/ledger/common"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/categories"
	"github.com/keep94/finance/fin/categories/categoriesdb"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/goconsume"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	kMaxLines = 100
)

var (
	kTemplateSpec = `
<html>
<head>
  <title>{{.Global.Title}}</title>
  {{if .Global.Icon}}
    <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
  {{end}}
  <link rel="stylesheet" type="text/css" href="/static/theme.css">
</head>
<body>
{{.LeftNav}}
<div class="main">
{{if .Error}}
  <span class="error">{{.Error.Error}}</span>
{{end}}
<form method="post">
<table>
  </tr>
    <td align="right">Account: </td>
    <td>
      <select name="acctId" size=1>
{{with .GetSelection .AccountSelectModel "acctId"}}
        <option value="{{.Value}}">{{.Name}}</option>
{{else}}
        <option value="">--Pick one--</option>
{{end}}
{{range .ActiveAccountDetails}}
        <option value="{{.Id}}">{{.Name}}</option>
{{end}}
      <select>
    </td>
  </tr>
  <tr>
    <td align="right">Start Date: </td>
    <td><input type="text" name="sd" value="{{.Get "sd"}}"></td>
  </tr>
  <tr>
    <td align="right">End Date: </td>
    <td><input type="text" name="ed" value="{{.Get "ed"}}"></td>
  </tr>
</table>
<input type="submit" name="download" value="Download">
</form>
</div>
</body>
</html>`
)

var (
	kTemplate *template.Template
)

type Handler struct {
	Store  findb.EntriesRunner
	Cdc    categoriesdb.Getter
	Clock  date_util.Clock
	LN     *common.LeftNav
	Global *common.Global
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cds, _ := h.Cdc.Get(nil)
	if r.Method == "GET" {
		h.doGet(w, r, cds)
	} else {
		h.doPost(w, r, cds)
	}
}

func (h *Handler) doGet(
	w http.ResponseWriter, r *http.Request, cds categories.CatDetailStore) {
	now := date_util.TimeToDate(h.Clock.Now())
	oneMonthAgo := now.AddDate(0, -1, 0)
	values := make(url.Values)
	values.Set("sd", oneMonthAgo.Format(date_util.YMDFormat))
	leftnav := h.LN.Generate(w, r, common.SelectExport())
	if leftnav == "" {
		return
	}
	http_util.WriteTemplate(
		w,
		kTemplate,
		h.toViewFromForm(values, cds, leftnav, nil))
}

func (h *Handler) doPost(
	w http.ResponseWriter, r *http.Request, cds categories.CatDetailStore) {
	r.ParseForm()
	leftnav := h.LN.Generate(w, r, common.SelectExport())
	if leftnav == "" {
		return
	}
	acctId, elo, err := parseForm(r.Form)
	if err != nil {
		http_util.WriteTemplate(
			w,
			kTemplate,
			h.toViewFromForm(r.Form, cds, leftnav, err))
		return
	}
	buffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(buffer)
	var columns [5]string
	columns[0] = "Date"
	columns[1] = "CheckNo"
	columns[2] = "Name"
	columns[3] = "Desc"
	columns[4] = "Amount"
	csvWriter.Write(columns[:])
	var consumer goconsume.Consumer
	consumer = goconsume.ConsumerFunc(func(ptr interface{}) {
		entry := ptr.(*fin.Entry)
		columns[0] = entry.Date.Format("1/2/2006")
		columns[1] = entry.CheckNo
		columns[2] = entry.Name
		columns[3] = entry.Desc
		columns[4] = fin.FormatUSD(-entry.Total())
		csvWriter.Write(columns[:])
	})
	consumer = goconsume.Slice(consumer, 0, kMaxLines+1)
	consumer = goconsume.Filter(
		consumer,
		func(ptr interface{}) bool {
			p := ptr.(*fin.Entry)
			return p.WithPayment(acctId)
		})
	err = h.Store.Entries(nil, elo, consumer)
	if err == nil && !consumer.CanConsume() {
		err = errors.New("File too big. Try a smaller date range")
	}
	if err != nil {
		http_util.WriteTemplate(
			w,
			kTemplate,
			h.toViewFromForm(r.Form, cds, leftnav, err))
		return
	}
	csvWriter.Flush()
	header := w.Header()
	header.Add("Content-Type", "application/octet-stream")
	now := date_util.TimeToDate(h.Clock.Now())
	header.Add(
		"Content-Disposition",
		fmt.Sprintf(
			"attachment; filename=\"Account_%d_%s.csv\"",
			acctId,
			now.Format(date_util.YMDFormat)))
	buffer.WriteTo(w)
}

func (h *Handler) toViewFromForm(
	values url.Values,
	cds categories.CatDetailStore,
	leftnav template.HTML,
	err error) *view {
	return &view{
		http_util.Values{values},
		common.CatDisplayer{cds},
		leftnav,
		h.Global,
		err}
}

type view struct {
	http_util.Values
	common.CatDisplayer
	LeftNav template.HTML
	Global  *common.Global
	Error   error
}

func parseForm(values url.Values) (
	acctId int64, elo *findb.EntryListOptions, err error) {
	acctId, err = strconv.ParseInt(values.Get("acctId"), 10, 64)
	if err != nil {
		err = errors.New("Account required.")
		return
	}
	sdptr, sderr := getDateRelaxed(values, "sd")
	edptr, ederr := getDateRelaxed(values, "ed")
	if sderr != nil || ederr != nil {
		err = errors.New("Dates must be in yyyyMMdd format.")
		return
	}
	elo = &findb.EntryListOptions{Start: sdptr, End: edptr}
	return
}

func getDateRelaxed(values url.Values, key string) (*time.Time, error) {
	s := strings.TrimSpace(values.Get(key))
	if s == "" {
		return nil, nil
	}
	t, e := time.Parse(date_util.YMDFormat, common.NormalizeYMDStr(s))
	if e != nil {
		return nil, e
	}
	return &t, nil
}

func init() {
	kTemplate = common.NewTemplate("export", kTemplateSpec)
}
