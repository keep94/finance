package single

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/functional"
  "html/template"
  "net/http"
  "net/url"
  "strconv"
  "strings"
  "time"
)

const (
  kMaxSplits = 10
)

var (
  kDateMayBeWrong = errors.New("Date may be wrong, proceed anyway?")
  kCantAccessDatastore = errors.New("Can't access datastore")
)

var (
kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css">

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
<script type="text/javascript">
  gActiveCategories = [{{range .ActiveCatDetails true}}"{{.Id}}", "{{.FullName}}",{{end}}];
</script>
<style type="text/css">
#nameAutoComplete {
  width:25em;
  padding-bottom:1em;
}

#descAutoComplete {
  width:25em;
  padding-bottom:1em;
}
</style>
</head>
<body class="yui-skin-sam">
{{if .Error}}
  <span class="error">{{.Error.Error}}</span>
{{end}}
<form method="post">
<input type="submit" name="save" value="Save">
<input type="submit" name="cancel" value="Cancel">
{{if .ExistingEntry}}
<input type="submit" name="delete" value="Delete" onclick="return confirm('Are you sure you want to delete this entry?');">
<input type="hidden" name="etag" value="{{.Get "etag"}}">
{{end}}
{{if .DateMayBeWrong}}
<input type="hidden" name="last_date" value="{{.Get "date"}}">
{{end}}
<input type="checkbox" name="need_review" {{if .Get "need_review"}}checked{{end}}>Under review
<table>
  <tr>
    <td align="right">Date: </td>
    <td><input type="text" name="date" value="{{.Get "date"}}"></td>
  </tr>
  <tr>
    <td align="right">Name: </td>
    <td>
      <div id="nameAutoComplete">
        <input type="text" id="nameField" name="name" value="{{.Get "name"}}">
        <div id="nameContainer"></div>
      </div>
    </td>
  </tr>
  <tr>
    <td align="right">Desc: </td>
    <td>
      <div id="descAutoComplete">
        <input type="text" id="descField" name="desc" value="{{.Get "desc"}}">
        <div id="descContainer"></div>
      </div>
    </td>
  </tr>
  <tr>
    <td align="right">Check #: </td>
    <td><input type="text" name="checkno" value="{{.Get "checkno"}}"></td>
  <tr>
</table>
<table>
  <tr>
    <td>Payment: </td>
    <td>
      <select name="payment" size=1>
{{with .GetSelection .AccountSelectModel "payment"}}
        <option value="{{.Value}}">{{.Name}}</option>
{{else}}
        <option value="">--Pick one--</option>
{{end}}
{{range .ActiveAccountDetails}}
        <option value="{{.Id}}">{{.Name}}</option>
{{end}}
      </select> 
      <input type="checkbox" name="reconciled" {{if .Get "reconciled"}}checked{{end}}>
    </td>
  </tr>
</table>
<table>
{{with $top := .}}
  {{range .Splits}}
  <tr>
    <td>
      <select id="{{.CatParam}}" name="{{.CatParam}}" size=1>
    {{with $top.GetSelection $top.CatSelectModel .CatParam}}
        <option value="{{.Value}}">{{.Name}}</option>
    {{end}}
      </select>
      <script type="text/javascript">populateSelect(document.getElementById("{{.CatParam}}"), gActiveCategories)</script>
    </td>
    <td>
      <input type="text" name="{{.AmountParam}}" value="{{$top.Get .AmountParam}}" size="12">
    </td>
    <td>
     <input type="checkbox" name="{{.ReconcileParam}}" {{if $top.Get .ReconcileParam}} checked {{end}}>
    </td>
  </tr>
  {{end}}
{{end}}
</table>
<input type="submit" name="save" value="Save">
<input type="submit" name="cancel" value="Cancel">
{{if .ExistingEntry}}
<input type="submit" name="delete" value="Delete" onclick="return confirm('Are you sure you want to delete this entry?');">
{{end}}
</form>

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
</html>`
)

var (
  kTemplate *template.Template
  kSplits []split
)

// Store methods are from fin.Store
type Store interface {
  findb.DoEntryChangesRunner
  findb.EntryByIdRunner
}

type Handler struct {
  Doer db.Doer
  Clock date_util.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  store, ok := session.Store.(Store)
  if !ok {
    http_util.ReportError(
        w, "Error accessing datastore", kCantAccessDatastore)
    return
  }
  id, _ := strconv.ParseInt(r.Form.Get("id"), 10, 64)
  paymentId, _ := strconv.ParseInt(r.Form.Get("aid"), 10, 64)
  if r.Method == "GET" {
    h.doGet(w, id, paymentId, store, session.Cache)
  } else {
    h.doPost(w, r, id, store, session.Cache)
  }
}

func (h *Handler) doPost(
    w http.ResponseWriter, r *http.Request, id int64,
    store Store, cdc categoriesdb.Getter) {
  var err error
  if http_util.HasParam(r.Form, "delete") {
    if isIdValid(id) {
      err = deleteId(id, store)
    }
  } else if http_util.HasParam(r.Form, "cancel") {
    // Do nothing
  } else {
    // Save button
    var mutation functional.Filterer
    mutation, err = entryMutation(r.Form)
    if err == nil {
      if isIdValid(id) {
        tag, _ := strconv.ParseUint(r.Form.Get("etag"), 10, 32)
        err = updateId(id, uint32(tag), mutation, store)
      } else {
        entry := fin.Entry{}
        mutation.Filter(&entry)
        // If user changed date since last submission check if that date is
        // reasonable.
        if r.Form.Get("last_date") != r.Form.Get("date") {
          if h.isDateReasonable(entry.Date) {
            err = add(&entry, store)
          } else {
            err = kDateMayBeWrong
          }
        } else {
          err = add(&entry, store)
        }
      }
    }
  }
  if err != nil {
    if err == findb.ConcurrentUpdate {
      err = errors.New("Someone else already updated this entry. Hit cancel and try again.") 
    }
    cds, _ := cdc.Get(nil)
    http_util.WriteTemplate(
        w, kTemplate, toViewFromForm(isIdValid(id), r.Form, cds, err))
  } else {
    http_util.Redirect(w, r, r.Form.Get("prev"))
  }
}

func (h *Handler) doGet(
    w http.ResponseWriter, id, paymentId int64,
    store findb.EntryByIdRunner, cdc categoriesdb.Getter) {
  var v *view
  if isIdValid(id) {
    entry := fin.Entry{}
    cds := categories.CatDetailStore{}
    err := h.Doer.Do(func(t db.Transaction) (err error) {
      cds, err = cdc.Get(t)
      if err != nil {
        return
      }
      return store.EntryById(t, id, &entry)
    })
    if err == findb.NoSuchId {
      fmt.Fprintln(w, "No entry found.")
      return
    }
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    v = toView(&entry, cds)
  } else {
    cds, _ := cdc.Get(nil)
    values := make(url.Values)
    if paymentId > 0 {
      values.Set("payment", strconv.FormatInt(paymentId, 10))
    }
    v = toViewFromForm(false, values, cds, nil)
  }
  http_util.WriteTemplate(w, kTemplate, v)
}

func (h *Handler) isDateReasonable(date time.Time) bool {
  currentDate := date_util.TimeToDate(h.Clock.Now())
  oneMonthBefore := currentDate.AddDate(0, -1, 0)
  return !(date.Before(oneMonthBefore) || date.After(currentDate))
}

type view struct {
  http_util.Values
  common.CatDisplayer
  Splits []split
  Error error
  ExistingEntry bool
}

func (v view) DateMayBeWrong() bool {
  return v.Error == kDateMayBeWrong
}

type split int

func (s split) CatParam() string {
  return fmt.Sprintf("cat-%d", int(s))
}

func (s split) AmountParam() string {
  return fmt.Sprintf("amount-%d", int(s))
}

func (s split) ReconcileParam() string {
  return fmt.Sprintf("reconciled-%d", int(s))
}

func isIdValid(id int64) bool {
  return id > 0
}

func toView(entry *fin.Entry, cds categories.CatDetailStore) *view {
  result := &view{
      Values: http_util.Values{make(url.Values)}, CatDisplayer: common.CatDisplayer{cds}, Splits: kSplits, Error: nil, ExistingEntry: true}
  tag, _ := etag.Etag32(entry)
  result.Set("etag", strconv.FormatInt(int64(tag), 10))
  result.Set("name", entry.Name)
  result.Set("desc", entry.Desc)
  result.Set("checkno", entry.CheckNo)
  result.Set("date", entry.Date.Format(date_util.YMDFormat))
  result.Set("payment", strconv.FormatInt(entry.PaymentId(), 10))
  if entry.Reconciled() {
    result.Set("reconciled", "on")
  }
  if entry.Status != fin.Reviewed {
    result.Set("need_review", "on")
  }
  catrecs := cds.SortedCatRecs(entry.CatRecs())
  for idx, split := range result.Splits {
    if idx < len(catrecs) {
      result.Set(split.CatParam(), catrecs[idx].Id().String())
      result.Set(split.AmountParam(), fin.FormatUSD(catrecs[idx].Amount()))
      if catrecs[idx].Reconciled() {
        result.Set(split.ReconcileParam(), "on")
      }
    }
  }
  return result
}

// ShowEntryFromForm shows an entry from form values.
func toViewFromForm(
    existingEntry bool, values url.Values, cds categories.CatDetailStore, err error) *view {
  return &view{
      Values: http_util.Values{values},
      CatDisplayer: common.CatDisplayer{cds},
      Splits: kSplits,
      Error: err,
      ExistingEntry: existingEntry}
}

func entryMutation(values url.Values) (mutation functional.Filterer, err error) {
  date, err := time.Parse(date_util.YMDFormat, values.Get("date"))
  if err != nil {
    err = errors.New("Date in wrong format.")
    return
  }
  name := values.Get("name")
  if strings.TrimSpace(name) == "" {
    err = errors.New("Name required.")
    return
  }
  desc := values.Get("desc")
  checkno := values.Get("checkno")
  paymentId, _ := strconv.ParseInt(values.Get("payment"), 10, 64)
  if paymentId == 0 {
    err = errors.New("Missing payment.")
    return
  }
  cpb := fin.CatPaymentBuilder{}
  cpb.SetPaymentId(paymentId).SetReconciled(values.Get("reconciled") != "")
  catrec := fin.CatRec{}
  for _, split := range kSplits {
    cat := fin.NewCat(values.Get(split.CatParam()))
    amountStr := values.Get(split.AmountParam())
    if amountStr == "" {
      break
    }
    var amount int64
    amount, err = fin.ParseUSD(amountStr)
    if err != nil {
      err = errors.New(fmt.Sprintf("Invalid amount: %s", amountStr))
      return
    }
    catrec = fin.CatRec{C: cat, A: amount, R: values.Get(split.ReconcileParam()) != ""}
    cpb.AddCatRec(&catrec)
  }
  cp := cpb.Build()
  needReview := values.Get("need_review") != ""
  mutation = functional.NewFilterer(func(ptr interface{}) error {
    p := ptr.(*fin.Entry)
    p.Date = date
    p.Name = name
    p.Desc = desc
    p.CheckNo = checkno
    p.CatPayment = cp
    if needReview {
      if p.Status == fin.Reviewed {
        p.Status = fin.NotReviewed
      }
    } else {
      if p.Status != fin.Reviewed {
        p.Status = fin.Reviewed
      }
    }
    return nil
  })
  return
}

func deleteId(id int64, store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{Deletes: []int64{id}}
  return store.DoEntryChanges(nil, &changes)
}

func updateId(
    id int64,
    tag uint32,
    mutation functional.Filterer,
    store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{
      Updates: map[int64]functional.Filterer{ id: mutation},
      Etags: map[int64]uint32{ id: tag}}
  return store.DoEntryChanges(nil, &changes)
}

func add(entry *fin.Entry, store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{Adds: []*fin.Entry{entry}}
  return store.DoEntryChanges(nil, &changes)
}

func init() {
  kTemplate = common.NewTemplate("single", kTemplateSpec)
  kSplits = make([]split, kMaxSplits)
  for i := range kSplits {
    kSplits[i] = split(i)
  }
}
