package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/aggregators"
	"github.com/keep94/finance/fin/categories"
	csqlite "github.com/keep94/finance/fin/categories/categoriesdb/for_sqlite"
	"github.com/keep94/finance/fin/consumers"
	"github.com/keep94/finance/fin/filters"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/finance/fin/findb/for_sqlite"
	"github.com/keep94/goconsume"
	"github.com/keep94/gofunctional3/functional"
	"github.com/keep94/gosqlite/sqlite"
	"github.com/keep94/toolbox/date_util"
	"github.com/keep94/toolbox/db/sqlite_db"
	"github.com/keep94/toolbox/google_graph"
	"html/template"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	kTemplateStr = `<html>
<body>
<table>
  <tr>
    <td>{{.MonthStr}} Income:</td>
    <td align="right">{{.MonthIncome}}</td>
  </tr>
  <tr>
    <td>{{.MonthStr}} Spending:</td>
    <td align="right">{{.MonthExpense}}</td>
  </tr>
  <tr>
    <td><b>{{.MonthStr}} Net:</b></td>
    <td align="right"><b>{{.MonthNet}}</b></td>
  </tr>
</table>
<br>
<table>
  <tr>
    <td>YTD Income:</td>
    <td align="right">{{.YTDIncome}}</td>
  </tr>
  <tr>
    <td>YTD Spending:</td>
    <td align="right">{{.YTDExpense}}</td>
  </tr>
  <tr>
    <td><b>YTD Net:</b></td>
    <td align="right"><b>{{.YTDNet}}</b></td>
  </tr>
</table>
<br>
<table>
{{range $rowIdx, $row := .Data}}
  <tr>   
  {{range $colIdx, $col := .}}
    {{if $rowIdx}}
      {{if $colIdx}}
        <td align="right">{{.}}</td>
      {{else}}
        <td>{{.}}</td>
      {{end}}
    {{else}}
      <td>{{.}}</td>
    {{end}}
  {{end}}
  </tr>
{{end}}
</table>
<br>
<img src="{{.Link}}" />
</body>
</html>
`
)

var (
	kTemplate      *template.Template
	fRecipients    string
	fConfig        string
	fDb            string
	fDate          string
	fGmailId       string
	fGmailPassword string
)

type balanceInfo struct {
	Expense int64
	Income  int64
}

func (b *balanceInfo) Net() int64 {
	return b.Income - b.Expense
}

type graphSpec struct {
	Title  string
	Filter goconsume.FilterFunc
}

type graphData struct {
	Titles []string
	Spec   []*graphSpec
	Totals [][]*aggregators.Totaler
}

func (g *graphData) XLen() int {
	return len(g.Spec)
}

func (g *graphData) YLen() int {
	return len(g.Totals)
}

func (g *graphData) XLabel(idx int) string {
	return g.Spec[idx].Title
}

func (g *graphData) YLabel(idx int) string {
	return g.Titles[idx]
}

func (g *graphData) Value(x, y int) int64 {
	return -g.Totals[y][x].Total
}

type view struct {
	Data         [][]string
	Link         *url.URL
	MonthStr     string
	MonthIncome  string
	MonthExpense string
	MonthNet     string
	YTDIncome    string
	YTDExpense   string
	YTDNet       string
}

func newDateFilter(start, end time.Time) goconsume.FilterFunc {
	return func(ptr interface{}) bool {
		p := ptr.(*fin.Entry)
		return !p.Date.Before(start) && p.Date.Before(end)
	}
}

func newConsumer(
	filter goconsume.FilterFunc,
	total *aggregators.Totaler) goconsume.Consumer {
	return goconsume.Filter(
		consumers.FromCatPaymentAggregator(total),
		filter)
}

func byCatFilterer(
	cds categories.CatDetailStore, name string) goconsume.FilterFunc {
	detail, ok := cds.DetailByFullName(name)
	if !ok {
		return nil
	}
	return filters.CompileAdvanceSearchSpec(
		&filters.AdvanceSearchSpec{CF: cds.Filter(detail.Id(), true)})
}

func byCatIdFilterer(
	cds categories.CatDetailStore, id fin.Cat) goconsume.FilterFunc {
	return filters.CompileAdvanceSearchSpec(
		&filters.AdvanceSearchSpec{CF: cds.Filter(id, true)})
}

func readGraphSpec(cds categories.CatDetailStore, path string) []*graphSpec {
	var result []*graphSpec
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	s := functional.ReadLines(f)
	var line string
	for s.Next(&line) == nil {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			log.Fatal("Config file must have 3 columns.")
		}
		result = append(result, &graphSpec{parts[2], byCatFilterer(cds, parts[1])})
	}
	return result
}

type reporter struct {
	takers []goconsume.Consumer
}

func (r *reporter) ComputeTotal(
	filter goconsume.FilterFunc) *aggregators.Totaler {
	result := &aggregators.Totaler{}
	r.takers = append(
		r.takers,
		newConsumer(filter, result))
	return result
}

func (r *reporter) ComputeTotals(spec []*graphSpec, start, end time.Time) []*aggregators.Totaler {
	result := make([]*aggregators.Totaler, len(spec))
	dateFilter := newDateFilter(start, end)
	for i := range result {
		result[i] = &aggregators.Totaler{}
		if spec[i].Filter == nil {
			continue
		}
		r.takers = append(
			r.takers,
			newConsumer(goconsume.All(dateFilter, spec[i].Filter), result[i]))
	}
	return result
}

func (r *reporter) ToConsumer() goconsume.Consumer {
	return goconsume.ComposeWithCopy(r.takers, (*fin.Entry)(nil))
}

func toTable(gd google_graph.GraphData2D) [][]string {
	xlen := gd.XLen()
	ylen := gd.YLen()
	result := make([][]string, xlen+1)
	for i := range result {
		result[i] = make([]string, ylen+1)
	}
	for i := 0; i < ylen; i++ {
		result[0][i+1] = gd.YLabel(i)
	}
	for i := 0; i < xlen; i++ {
		result[i+1][0] = gd.XLabel(i)
		for j := 0; j < ylen; j++ {
			result[i+1][j+1] = fin.FormatUSD(gd.Value(i, j))
		}
	}
	return result
}

func buildMessageHtml(
	subject string,
	gd google_graph.GraphData2D,
	grapher google_graph.Grapher2D,
	currentMonthName string,
	monthlyBalance, yearlyBalance *balanceInfo,
	recipients []string) []byte {
	var buffer bytes.Buffer
	var buffer1 bytes.Buffer
	w := multipart.NewWriter(&buffer1)
	fmt.Fprintf(&buffer, "To: %s\n", strings.Join(recipients, ", "))
	fmt.Fprintf(&buffer, "Subject: %s\n", subject)
	fmt.Fprintf(&buffer, "MIME-version: 1.0\n")
	fmt.Fprintf(&buffer, "Content-Type: multipart/mixed; boundary=%s\n", w.Boundary())
	part, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(part, "Below is the graph of expenses.\n\n")
	part, err = w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html"}})
	if err != nil {
		log.Fatal(err)
	}
	err = kTemplate.Execute(part, &view{
		Data:         toTable(gd),
		Link:         grapher.GraphURL2D(gd),
		MonthStr:     currentMonthName,
		MonthIncome:  fin.FormatUSD(monthlyBalance.Income),
		MonthExpense: fin.FormatUSD(monthlyBalance.Expense),
		MonthNet:     fin.FormatUSD(monthlyBalance.Net()),
		YTDIncome:    fin.FormatUSD(yearlyBalance.Income),
		YTDExpense:   fin.FormatUSD(yearlyBalance.Expense),
		YTDNet:       fin.FormatUSD(yearlyBalance.Net())})
	if err != nil {
		log.Fatal(err)
	}
	w.Close()
	buffer1.WriteTo(&buffer)
	return buffer.Bytes()
}

func toRecipients(s string) []string {
	temp := strings.Split(s, ",")
	result := make([]string, len(temp))
	for i := range temp {
		result[i] = strings.TrimSpace(temp[i])
	}
	return result
}

func main() {
	flag.Parse()
	if fRecipients == "" || fConfig == "" || fDb == "" {
		fmt.Println("Need to specify recipients, config, and db")
		flag.Usage()
		return
	}

	// fRecipients
	recipients := toRecipients(fRecipients)

	// fDb
	conn, err := sqlite.Open(fDb)
	if err != nil {
		log.Fatal(err)
	}
	dbase := sqlite_db.New(conn)
	defer dbase.Close()
	cache := csqlite.New(dbase)
	store := for_sqlite.New(dbase)
	cds, _ := cache.Get(nil)

	// fConfig
	data := readGraphSpec(cds, fConfig)

	// Set up reporting dates
	var theDate time.Time
	if fDate != "" {
		var err error
		theDate, err = time.Parse("20060102", fDate)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		theDate = date_util.TimeToDate(time.Now())
	}
	monthly := aggregators.Monthly()
	yearly := aggregators.Yearly()
	nextMonth := monthly.Normalize(theDate)
	currentMonth := monthly.Add(nextMonth, -1)
	prevMonth := monthly.Add(nextMonth, -2)
	currentYear := yearly.Normalize(currentMonth)

	// Set up reporting
	var r reporter
	prevTotals := r.ComputeTotals(data, prevMonth, currentMonth)
	totals := r.ComputeTotals(data, currentMonth, nextMonth)
	lastMonthFilter := newDateFilter(currentMonth, nextMonth)
	ytdFilter := newDateFilter(currentYear, nextMonth)
	expenseFilter := byCatIdFilterer(cds, fin.Expense)
	incomeFilter := byCatIdFilterer(cds, fin.Income)
	monthExpense := r.ComputeTotal(
		goconsume.All(lastMonthFilter, expenseFilter))
	monthIncome := r.ComputeTotal(
		goconsume.All(lastMonthFilter, incomeFilter))
	ytdExpense := r.ComputeTotal(
		goconsume.All(ytdFilter, expenseFilter))
	ytdIncome := r.ComputeTotal(
		goconsume.All(ytdFilter, incomeFilter))

	startTime := currentYear
	if prevMonth.Before(startTime) {
		startTime = prevMonth
	}
	err = store.Entries(nil, &findb.EntryListOptions{Start: &startTime, End: &nextMonth}, r.ToConsumer())
	if err != nil {
		log.Fatal(err)
	}
	barGraph := &google_graph.BarGraph{Palette: []string{"000099", "006600"}, Scale: 2}
	gd := &graphData{
		Titles: []string{
			prevMonth.Format("Jan 2006"),
			currentMonth.Format("Jan 2006")},
		Spec:   data,
		Totals: [][]*aggregators.Totaler{prevTotals, totals}}
	auth := smtp.PlainAuth(
		"", fGmailId, fGmailPassword, "smtp.gmail.com")
	subject := fmt.Sprintf(
		"Expense report for %s", currentMonth.Format("Jan 2006"))
	message := buildMessageHtml(
		subject,
		gd,
		barGraph,
		currentMonth.Format("Jan 2006"),
		&balanceInfo{
			Expense: -monthExpense.Total,
			Income:  monthIncome.Total},
		&balanceInfo{
			Expense: -ytdExpense.Total,
			Income:  ytdIncome.Total},
		recipients)
	err = smtp.SendMail("smtp.gmail.com:587", auth, fGmailId+"@gmail.com", recipients, message)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.StringVar(&fRecipients, "recipients", "", "Email Recipients")
	flag.StringVar(&fConfig, "config", "", "Configuration File")
	flag.StringVar(&fDb, "db", "", "Path to database file.")
	flag.StringVar(&fDate, "date", "", "Optional: Current date in yyyyMMdd format.")
	flag.StringVar(&fGmailId, "gmailid", "", "GMail ID")
	flag.StringVar(&fGmailPassword, "gmailpassword", "", "GMail Password")
	kTemplate = template.Must(template.New("email").Parse(kTemplateStr))
}
