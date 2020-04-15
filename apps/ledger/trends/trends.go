package trends

import (
  "errors"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/google_graph"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/aggregators"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/filters"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/goconsume"
  "html/template"
  "net/http"
  "net/url"
  "time"
)

const (
  kPageParam = "pageNo"
  kMaxPointsInGraph = 24
)

var (
  kExpenseBarGraph = &google_graph.BarGraph{
      Palette: []string {"660000"},
      Scale: 2}
  kIncomeBarGraph = &google_graph.BarGraph{
      Palette: []string {"006600", "660000"},
      Scale: 2}
  kExpenseIncomeBarGraph = kIncomeBarGraph
)

var (
  kTemplateSpec = `
{{define "MultiGraph"}}
<table>
  <tr>
    <td>
      <table border=1>
        <tr>
          <td>Date</td>
          <td>Income</td>
          <td>Expense</td>
          <td>Report</td>
        </tr>
{{with $top := .}}
{{range .MultiItems}}
        <tr>
  {{if .Url}}
          <td><a href="{{.Url}}">{{.Date.Format $top.FormatStr}}</a></td>
  {{else}}
          <td>{{.Date.Format $top.FormatStr}}</td>
  {{end}}
          <td align="right">{{FormatUSDRaw .IncomeValue}}</td>
          <td align="right">{{FormatUSDRaw .ExpenseValue}}</td>
  {{if .ReportUrl}}
          <td><a href="{{.ReportUrl}}">report</a></td>
  {{else}}
          <td>&nbsp;</td>
  {{end}}
        </tr>
{{end}}
{{end}}
      </table>
    </td>
    <td>
{{if .GraphUrl}}
  <img src="{{.GraphUrl}}" alt="graph">
{{else}}
  &nbsp;
{{end}}
    </td>
  </tr>
</table>
{{end}}
{{define "Graph"}}
<table>
  <tr>
    <td>
      <table border=1>
        <tr>
          <td>Date</td>
          <td>Amount</td>
          <td>Report</td>
        </tr>
{{with $top := .}}
{{range .Items}}
        <tr>
  {{if .Url}}
          <td><a href="{{.Url}}">{{.Date.Format $top.FormatStr}}</a></td>
  {{else}}
          <td>{{.Date.Format $top.FormatStr}}</td>
  {{end}}
        <td align="right">{{FormatUSDRaw .Value}}</td>
  {{if .ReportUrl}}
          <td><a href="{{.ReportUrl}}">report</a></td>
  {{else}}
          <td>&nbsp;</td>
  {{end}}
        </tr>
{{end}}
{{end}}
      </table>
    </td>
    <td>
{{if .GraphUrl}}
  <img src="{{.GraphUrl}}" alt="graph">
{{else}}
  &nbsp;
{{end}}
    </td>
  </tr>
</table>
{{end}}
<html>
  <head>
    <link rel="stylesheet" type="text/css" href="/static/theme.css" />
  </head>
  <body>
{{if .Error}}
  <span class="error">{{.Error}}</span>
{{end}}
    <form method="get">
      <table>
        <tr>
          <td>Category: </td>
          <td>
            <select name="cat">
{{with .GetSelection .CatSelectModel "cat"}}
              <option value="{{.Value}}">{{.Name}}</option>
{{end}}
              <option value="">ALL</option>
{{range .CatDetails}}
              <option value="{{.Id}}">{{.FullName}}</option>
{{end}}
            </select>
          </td>
          <td>Start date: </td>
          <td><input type="text" name="sd" value="{{.Get "sd"}}"></td>
          <td>End date: </td>
          <td><input type="text" name="ed" value="{{.Get "ed"}}"></td>
        </tr>
        <tr>
          <td>Top level: </td>
          <td><input type="checkbox" name="top" {{if .Get "top"}}checked{{end}}></td>
          <td>Frequency: </td>
          <td colspan="3"><select name="freq">
            <option value="M" {{if .Equals "freq" "M"}}selected{{end}}>Monthly</option>
            <option value="Y" {{if .Equals "freq" "Y"}}selected{{end}}>Yearly</option>
          </select></td>
        </tr>
        <tr>
          <td colspan="6">
            <input type="submit" value="Generate report">
          </td>
        </tr>
      </table>
    </form>
{{if .Items}}
  {{template "Graph" .}}
{{else}}
  {{template "MultiGraph" .}}
{{end}}
  </body>
</html>`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Cdc categoriesdb.Getter
  Store findb.EntriesRunner
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  cds, _ := h.Cdc.Get(nil)
  cat, caterr := fin.CatFromString(r.Form.Get("cat"))
  start, end, err := getDateRange(r)
  if err != nil {
    v := &view{
        Values: http_util.Values{r.Form},
        CatDisplayer: common.CatDisplayer{cds},
        Error: errors.New("Dates must be in yyyyMMdd format."),
        CatDetails: cds.DetailsByIds(fin.CatSet{fin.Expense: true, fin.Income: true})}
    http_util.WriteTemplate(w, kTemplate, v)
    return
  }
  if caterr == nil {
    points, graphUrl, cats, err := h.singleCat(cds, r.URL, cat, r.Form.Get("top") != "", start, end, r.Form.Get("freq") == "Y")
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    v := &view{
        Values: http_util.Values{r.Form},
        CatDisplayer: common.CatDisplayer{cds},
        Items: points,
        CatDetails: cds.DetailsByIds(cats),
        GraphUrl: graphUrl,
        FormatStr: formatStringLong(r.Form.Get("freq") == "Y")}
    http_util.WriteTemplate(w, kTemplate, v)
  } else {
    points, graphUrl, cats, err := h.allCats(cds, r.URL, start, end, r.Form.Get("freq") == "Y")
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    v := &view{
        Values: http_util.Values{r.Form},
        CatDisplayer: common.CatDisplayer{cds},
        MultiItems: points,
        CatDetails: cds.DetailsByIds(cats),
        GraphUrl: graphUrl,
        FormatStr: formatStringLong(r.Form.Get("freq") == "Y")}
    http_util.WriteTemplate(w, kTemplate, v)
  }
       
}

func (h *Handler) singleCat(
    cds categories.CatDetailStore,
    thisUrl *url.URL,
    cat fin.Cat,
    topOnly bool,
    start, end time.Time,
    isYearly bool) (points []*dataPoint, graphUrl *url.URL, cats fin.CatSet, err error) {
  // Only to see what the child categories are
  ct := make(fin.CatTotals)
  totals := createByPeriodTotaler(start, end, isYearly)
  cr := goconsume.Filter(
      goconsume.Compose(
          consumers.FromCatPaymentAggregator(ct),
          consumers.FromEntryAggregator(totals)),
      filters.CompileAdvanceSearchSpec(
          &filters.AdvanceSearchSpec{CF: cds.Filter(cat, !topOnly)}))
  elo := findb.EntryListOptions{
      Start: &start,
      End: &end}
  err = h.Store.Entries(nil, &elo, cr)
  if err != nil {
    return
  }
  isIncome := cat.Type == fin.IncomeCat
  var listUrl *url.URL
  if topOnly {
    listUrl = http_util.NewUrl(
        "/fin/list",
        "cat", cat.String(),
        "top", "on")
  } else {
    listUrl = http_util.NewUrl(
        "/fin/list",
        "cat", cat.String())
  }
  var reportUrl *url.URL
  if isYearly {
    reportUrl = http_util.WithParams(thisUrl, "freq", "M")
  }
  builder := dataSetBuilder{
      ListUrl: listUrl,
      ReportUrl: reportUrl,
      Totals: totals,
      IsIncome: isIncome}
  points = builder.Build()
  if len(points) <= kMaxPointsInGraph {
    g := &graphable{
        Data: points,
        Fmt: formatString(isYearly)}
    if isIncome {
      graphUrl = kIncomeBarGraph.GraphURL(g)
    } else {
      graphUrl = kExpenseBarGraph.GraphURL(g)
    }
  }
  _, children := cds.RollUp(ct)
  cats = fin.CatSet{fin.Expense:true , fin.Income:true}
  cats.AddSet(children[cat])
  return
}

func (h *Handler) allCats(
    cds categories.CatDetailStore,
    thisUrl *url.URL,
    start, end time.Time,
    isYearly bool) (points []*multiDataPoint, graphUrl *url.URL, cats fin.CatSet, err error) {
  // Only to see what the child categories are
  ct := make(fin.CatTotals)
  expenseTotals := createByPeriodTotaler(start, end, isYearly)
  incomeTotals := createByPeriodTotaler(start, end, isYearly)
  cr := goconsume.ComposeWithCopy(
      []goconsume.Consumer{
          consumers.FromCatPaymentAggregator(ct),
          goconsume.Filter(
              consumers.FromEntryAggregator(expenseTotals),
              filters.CompileAdvanceSearchSpec(
                  &filters.AdvanceSearchSpec{
                      CF: cds.Filter(fin.Expense, true)})),
          goconsume.Filter(
              consumers.FromEntryAggregator(incomeTotals),
              filters.CompileAdvanceSearchSpec(
                  &filters.AdvanceSearchSpec{
                      CF: cds.Filter(fin.Income, true)})),
      },
      (*fin.Entry)(nil))
  elo := findb.EntryListOptions{
      Start: &start,
      End: &end}
  err = h.Store.Entries(nil, &elo, cr)
  if err != nil {
    return
  }
  listUrl := http_util.NewUrl("/fin/list")
  var reportUrl *url.URL
  if isYearly {
    reportUrl = http_util.WithParams(thisUrl, "freq", "M")
  }
  builder := multiDataSetBuilder{
      ListUrl: listUrl,
      ReportUrl: reportUrl,
      ExpenseTotals: expenseTotals,
      IncomeTotals: incomeTotals}
  points = builder.Build()
  if len(points) <= kMaxPointsInGraph {
    g := &multiGraphable{
        Data: points,
        Fmt: formatString(isYearly)}
    graphUrl = kExpenseIncomeBarGraph.GraphURL2D(g)
  }
  _, children := cds.RollUp(ct)
  cats = fin.CatSet{fin.Expense: true, fin.Income: true}
  cats.AddSet(children[fin.Expense]).AddSet(children[fin.Income])
  return
}

func formatString(isYearly bool) string {
  if isYearly {
    return "06"
  }
  return "01"
}

func formatStringLong(isYearly bool) string {
  if isYearly {
    return "2006"
  }
  return "01/2006"
}

func createByPeriodTotaler(start, end time.Time, isYearly bool) *aggregators.ByPeriodTotaler {
  if isYearly {
    return aggregators.NewByPeriodTotaler(start, end, aggregators.Yearly())
  }
  return aggregators.NewByPeriodTotaler(start, end, aggregators.Monthly())
}

func getDateRange(r *http.Request) (start, end time.Time, err error) {
  start, err = time.Parse(
      date_util.YMDFormat, common.NormalizeYMDStr(r.Form.Get("sd")))
  if err != nil {
    return
  }
  end, err = time.Parse(
      date_util.YMDFormat, common.NormalizeYMDStr(r.Form.Get("ed")))
  if err != nil {
    return
  }
  return
}

type dataPoint struct {
  Date time.Time
  Value int64
  Url *url.URL
  ReportUrl *url.URL
}

type graphable struct {
  Data []*dataPoint
  Fmt string
}

func (g *graphable) Len() int { return len(g.Data) }
func (g *graphable) Label(i int) string { return g.Data[i].Date.Format(g.Fmt) }
func (g *graphable) Value(i int) int64 { return g.Data[i].Value }
func (g *graphable) Title() string { return "" }

type multiDataPoint struct {
  Date time.Time
  IncomeValue int64
  ExpenseValue int64
  Url *url.URL
  ReportUrl *url.URL
}

type multiGraphable struct {
  Data []*multiDataPoint
  Fmt string
}

func (g *multiGraphable) XLen() int { return len(g.Data) }
func (g *multiGraphable) YLen() int { return 2 }
func (g *multiGraphable) XLabel(i int) string { return g.Data[i].Date.Format(g.Fmt) }

func (g *multiGraphable) Value(x, y int) int64 { 
  if y == 0 {
    return g.Data[x].IncomeValue
  }
  return g.Data[x].ExpenseValue
}

func (g *multiGraphable) YLabel(y int) string {
  if y == 0 {
     return "Income"
  }
  return "Expense"
}
  
type view struct {
  http_util.Values
  common.CatDisplayer
  Items []*dataPoint
  MultiItems []*multiDataPoint
  GraphUrl *url.URL
  CatDetails []categories.CatDetail
  Error error
  FormatStr string
}

type dataSetBuilder struct {
  ListUrl *url.URL
  ReportUrl *url.URL
  Totals *aggregators.ByPeriodTotaler
  IsIncome bool
}
  
func (b *dataSetBuilder) Build() (result []*dataPoint) {
  iter := b.Totals.Iterator()
  var pt aggregators.PeriodTotal
  for iter.Next(&pt) {
    if !b.IsIncome {
      pt.Total = -pt.Total
    }  
    var reportUrl *url.URL
    sd := pt.Start.Format(date_util.YMDFormat)
    ed := pt.End.Format(date_util.YMDFormat)
    if b.ReportUrl != nil {
      reportUrl = http_util.WithParams(b.ReportUrl,
          "sd", sd,
          "ed", ed)
    }
    item := &dataPoint{
        Date: pt.PeriodStart,
        Value: pt.Total,
        ReportUrl: reportUrl,
        Url: http_util.WithParams(
            b.ListUrl,
            "sd", sd,
            "ed", ed)}
    result = append(result, item)
  }
  return result
}

type multiDataSetBuilder struct {
  ListUrl *url.URL
  ReportUrl *url.URL
  ExpenseTotals *aggregators.ByPeriodTotaler
  IncomeTotals *aggregators.ByPeriodTotaler
}
  
func (b *multiDataSetBuilder) Build() (result []*multiDataPoint) {
  iter := b.IncomeTotals.Iterator()
  expenseIter := b.ExpenseTotals.Iterator()
  var pt, ept aggregators.PeriodTotal
  for iter.Next(&pt) {
    if !expenseIter.Next(&ept) {
      panic("expense totals shorter than income totals.")
    }
    var reportUrl *url.URL
    sd := pt.Start.Format(date_util.YMDFormat)
    ed := pt.End.Format(date_util.YMDFormat)
    if b.ReportUrl != nil {
      reportUrl = http_util.WithParams(b.ReportUrl,
          "sd", sd,
          "ed", ed)
    }
    item := &multiDataPoint{
        Date: pt.PeriodStart,
        IncomeValue: pt.Total,
        ExpenseValue: -ept.Total,
        ReportUrl: reportUrl,
        Url: http_util.WithParams(
            b.ListUrl,
            "sd", sd,
            "ed", ed)}
    result = append(result, item)
  }
  return result
}

func init() {
  kTemplate = common.NewTemplate("trends", kTemplateSpec)
}
