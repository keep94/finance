package list

import (
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/aggregators"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/filters"
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
  kPageParam = "pageNo"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />

<style type="text/css">
/*margin and padding on body element
  can introduce errors in determining
  element position and are not recommended;
  we turn them off as a foundation for YUI
  CSS treatments. */
body {
        margin:0;
        padding:0;
}
</style>

<!--CSS file (default YUI Sam Skin) -->
<link type="text/css" rel="stylesheet" href="/static/autocomplete.css">

<!-- Dependencies -->
<script src="/static/yahoo-dom-event.js"></script>
<script src="/static/datasource-min.js"></script>

<!-- Source file -->
<script src="/static/autocomplete-min.js"></script>

<script type="text/javascript" src="/static/json2.js"></script>
<script type="text/javascript" src="/static/ledger.js"></script>

<style type="text/css">
#nameAutoComplete {
  width:15em;
  padding-bottom:1em;
}

#descAutoComplete {
  width:15em;
  padding-bottom:1em;
}
</style>

</head>
<body class="yui-skin-sam">
{{if .ErrorMessage}}
  <span class="error">{{.ErrorMessage}}</span>
{{end}}
<form>
  <table>
    <tr>
      <td>Category: </td>
      <td>
        <select name="cat" size=1>
{{with .GetSelection .CatSelectModel "cat"}}
          <option value="{{.Value}}">{{.Name}}</option>
{{end}}
          <option value="">ALL</option>
{{range .ActiveCatDetails false}}
          <option value="{{.Id}}">{{.FullName}}</option>
{{end}}
        </select>
        <br>
        Top level only: <input type="checkbox" name="top" {{if .Get "top"}}checked{{end}}>
      </td>
    </tr>
    <tr>
      <td>Start Date (yyyyMMdd): </td>
      <td><input type="text" name="sd" value="{{.Get "sd"}}"></td>
      <td>End Date (yyyyMMdd): </td>
      <td><input type="text" name="ed" value="{{.Get "ed"}}"></td>
    </tr>
    <tr>
      <td>Name: </td>
      <td>
        <div id="nameAutoComplete">
          <input type="text" id="nameField" name="name" value="{{.Get "name"}}">
          <div id="nameContainer"></div>
        </div>
      </td>
      <td>Range: </td>
      <td><input type="text" name="range" value="{{.Get "range"}}"></td>
    </tr>
    <tr>
      <td>Desc: </td>
      <td>
        <div id="descAutoComplete">
          <input type="desc" id="descField" name="desc" value="{{.Get "desc"}}">
          <div id="descContainer"></div>
        </div>
      </td>
    </tr>
  </table>
<input type="submit" value="Search">
</form>
<hr>
{{if .Totaler}}
<b>Total: {{FormatUSD .Total}}</b>&nbsp;&nbsp;
{{end}}
<a href="{{.NewEntryLink 0}}">New Entry</a>
<br><br>   
{{with $top := .}}
Page: {{.DisplayPageNo}}&nbsp;
{{if .PageNo}}<a href="{{.PrevPageLink}}">&lt;</a>{{end}}
{{if .End}}&nbsp;{{else}}<a href="{{.NextPageLink}}">&gt;</a>{{end}}
<br><br>
  <table>
    <tr>
      <td>Date</td>
      <td>Category</td>
      <td>Name</td>
      <td>Amount</td>
      <td>Account</td>
    </tr>
  {{range .Entries}}
      <tr class="lineitem">
        <td>{{FormatDate .Date}}</td>
        <td>{{range $top.CatLink .CatPayment}}{{if .Link}}<a href="{{.Link}}">{{.Text}}</a>{{else}}{{.Text}}{{end}}{{end}}</td>
        <td><a href="{{$top.EntryLink .Id}}">{{.Name}}</a></td>
        <td align=right>{{FormatUSD .Total}}</td>
        <td>{{with $top.AccountNameLink .CatPayment}}{{if .Link}}<a href="{{.Link}}">{{.Text}}</a>{{else}}{{.Text}}{{end}}{{end}}</td>
      </tr>
      <tr>
        <td>
          {{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}
        </td>
        <td colspan=4>{{.Desc}}</td>
      </tr>
  {{end}}
  </table>
  <br>
Page: {{.DisplayPageNo}}&nbsp;
{{if .PageNo}}<a href="{{.PrevPageLink}}">&lt;</a>{{end}}
{{if .End}}&nbsp;{{else}}<a href="{{.NextPageLink}}">&gt;</a>{{end}}
<script type="text/javascript">
  var nameSuggester = new Suggester("/fin/acname");
  var descSuggester = new Suggester("/fin/acdesc");
  var nameDs = new YAHOO.util.FunctionDataSource(function() {
    return nameSuggester.getSuggestions();
  });
  var descDs = new YAHOO.util.FunctionDataSource(function() {
    return descSuggester.getSuggestions();
  });
  var nameAutoComplete = new YAHOO.widget.AutoComplete("nameField", "nameContainer", nameDs);
  initAutoComplete(nameAutoComplete);
  var descAutoComplete = new YAHOO.widget.AutoComplete("descField", "descContainer", descDs);
  initAutoComplete(descAutoComplete);
</script>
</body>
</html>
{{end}}`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Cdc categoriesdb.Getter
  Store findb.EntriesRunner
  PageSize int
  Links bool
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  pageNo, _ := strconv.Atoi(r.Form.Get(kPageParam))
  cds, _ := h.Cdc.Get(nil)
  var filt fin.CatFilter
  cat, caterr := fin.CatFromString(r.Form.Get("cat"))
  if caterr == nil {
    filt = cds.Filter(cat, r.Form.Get("top") == "")
  }
  var amtFilter filters.AmountFilter
  errorMessage := ""
  rangeStr := r.Form.Get("range")
  if rangeStr != "" {
    amtFilter = compileRangeFilter(rangeStr)
    if amtFilter == nil {
      errorMessage = "Range must be of form 12.34 to 56.78."
    }
  }
  var filter goconsume.FilterFunc
  if amtFilter != nil || filt != nil || r.Form.Get("name") != "" || r.Form.Get("desc") != "" {
    filter = filters.CompileAdvanceSearchSpec(&filters.AdvanceSearchSpec{
        CF: filt,
        AF: amtFilter,
        Name: r.Form.Get("name"),
        Desc: r.Form.Get("desc")})
  }
  var totaler *aggregators.Totaler
  var entries []fin.Entry
  var morePages bool
  epb := goconsume.Page(pageNo, h.PageSize, &entries, &morePages)
  var cr goconsume.Consumer = epb
  sdPtr, sderr := getDateRelaxed(r.Form, "sd")
  edPtr, ederr := getDateRelaxed(r.Form, "ed")
  if filter != nil {
    if sdPtr != nil {
      totaler = &aggregators.Totaler{}
      cr = goconsume.Compose(consumers.FromCatPaymentAggregator(totaler), cr)
    }
    cr = goconsume.ModFilter(cr, filter, (*fin.Entry)(nil))
  }
  var elo *findb.EntryListOptions
  if sderr != nil || ederr != nil {
    errorMessage = "Start and end date must be in yyyyMMdd format."
  } else {
    elo = &findb.EntryListOptions{Start: sdPtr, End: edPtr}
  }
  err := h.Store.Entries(nil, elo, cr)
  epb.Finalize()
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  var listEntriesUrl *url.URL
  if h.Links {
    listEntriesUrl = r.URL
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          http_util.PageBreadCrumb{
              URL: r.URL,
              PageNoParam: kPageParam,
              PageNo: pageNo,
              End: !morePages},
          entries,
          totaler,
          http_util.Values{r.Form},
          common.CatDisplayer{cds},
          common.CatLinker{ListEntries: listEntriesUrl, Cds: cds},
          common.EntryLinker{r.URL},
          errorMessage})
}

type view struct {
  http_util.PageBreadCrumb
  Entries []fin.Entry
  *aggregators.Totaler
  http_util.Values
  common.CatDisplayer
  common.CatLinker
  common.EntryLinker
  ErrorMessage string
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

func compileRangeFilter(expr string) filters.AmountFilter {
  expr = strings.ToLower(expr)
  parts := strings.SplitN(expr, "to", 2)
  for i := range parts {
    parts[i] = strings.TrimSpace(parts[i])
  }
  if len(parts) == 1 {
    neededAmount, err := fin.ParseUSD(parts[0])
    if err != nil {
      return nil
    }
    return func(amt int64) bool {
      return amt == -neededAmount
    }
  }
  if parts[0] != "" && parts[1] != "" {
    lower, err := fin.ParseUSD(parts[0])
    if err != nil {
      return nil
    }
    upper, err := fin.ParseUSD(parts[1])
    if err != nil {
      return nil
    }
    return func(amt int64) bool {
      return amt >= -upper && amt <= -lower
    }
  }
  if parts[0] != "" {
    lower, err := fin.ParseUSD(parts[0])
    if err != nil {
      return nil
    }
    return func(amt int64) bool {
      return amt <= -lower
    }
  }
  if parts[1] != "" {
    upper, err := fin.ParseUSD(parts[1])
    if err != nil {
      return nil
    }
    return func(amt int64) bool {
      return amt >= -upper
    }
  }
  return nil
}

func init() {
  kTemplate = common.NewTemplate("list", kTemplateSpec)
}
