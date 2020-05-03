package single

import (
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "html/template"
  "net/http"
  "net/url"
  "strconv"
  "time"
)

const (
  kSingle = "single"
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
<input type="hidden" name="xsrf" value="{{.Xsrf}}">
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
)

// Store methods are from fin.Store
type Store interface {
  findb.DoEntryChangesRunner
  findb.EntryByIdRunner
  findb.EntriesRunner
}

type Handler struct {
  Doer db.Doer
  Clock date_util.Clock
  Global *common.Global
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  store := session.Store.(Store)
  catPopularity := session.CatPopularity()
  id, _ := strconv.ParseInt(r.Form.Get("id"), 10, 64)
  paymentId, _ := strconv.ParseInt(r.Form.Get("aid"), 10, 64)
  if r.Method == "GET" {
    h.doGet(w, r, id, paymentId, store, session.Cache, catPopularity)
  } else {
    h.doPost(w, r, id, store, session.Cache, catPopularity)
  }
}

func (h *Handler) doPost(
    w http.ResponseWriter, r *http.Request, id int64,
    store Store, cdc categoriesdb.Getter, catPopularity fin.CatPopularity) {
  var err error
  if !common.VerifyXsrfToken(r, kSingle) {
    err = common.ErrXsrf
  } else if http_util.HasParam(r.Form, "delete") {
    if isIdValid(id) {
      err = deleteId(id, store)
    }
  } else if http_util.HasParam(r.Form, "cancel") {
    // Do nothing
  } else {
    // Save button
    var mutation fin.EntryUpdater
    mutation, err = common.EntryMutation(r.Form)
    if err == nil {
      if isIdValid(id) {
        tag, _ := strconv.ParseUint(r.Form.Get("etag"), 10, 64)
        err = updateId(id, tag, mutation, store)
      } else {
        entry := fin.Entry{}
        mutation(&entry)
        // If user changed date since last submission check if that date is
        // reasonable.
        if r.Form.Get("last_date") != r.Form.Get("date") {
          if h.isDateReasonable(entry.Date) {
            err = add(&entry, store)
          } else {
            err = common.ErrDateMayBeWrong
          }
        } else {
          err = add(&entry, store)
        }
      }
    }
  }
  if err != nil {
    if err == findb.ConcurrentUpdate {
      err = common.ErrConcurrentModification
    }
    cds, _ := cdc.Get(nil)
    http_util.WriteTemplate(
        w,
        kTemplate,
        common.ToSingleEntryViewFromForm(
            isIdValid(id),
            r.Form,
            common.NewXsrfToken(r, kSingle),
            cds,
            catPopularity,
            h.Global,
            err))
  } else {
    http_util.Redirect(w, r, r.Form.Get("prev"))
  }
}

func (h *Handler) doGet(
    w http.ResponseWriter,
    r *http.Request,
    id, paymentId int64,
    store findb.EntryByIdRunner,
    cdc categoriesdb.Getter,
    catPopularity fin.CatPopularity) {
  var v *common.SingleEntryView
  if isIdValid(id) {
    var entryWithEtag fin.Entry
    var cds categories.CatDetailStore
    err := h.Doer.Do(func(t db.Transaction) (err error) {
      cds, err = cdc.Get(t)
      if err != nil {
        return
      }
      return store.EntryById(t, id, &entryWithEtag)
    })
    if err == findb.NoSuchId {
      fmt.Fprintln(w, "No entry found.")
      return
    }
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    v = common.ToSingleEntryView(
        &entryWithEtag,
        common.NewXsrfToken(r, kSingle),
        cds,
        catPopularity,
        h.Global)
  } else {
    cds, _ := cdc.Get(nil)
    values := make(url.Values)
    if paymentId > 0 {
      values.Set("payment", strconv.FormatInt(paymentId, 10))
    }
    v = common.ToSingleEntryViewFromForm(
        false,
        values,
        common.NewXsrfToken(r, kSingle),
        cds,
        catPopularity,
        h.Global,
        nil)
  }
  http_util.WriteTemplate(w, kTemplate, v)
}

func (h *Handler) isDateReasonable(date time.Time) bool {
  currentDate := date_util.TimeToDate(h.Clock.Now())
  oneMonthBefore := currentDate.AddDate(0, -1, 0)
  return !(date.Before(oneMonthBefore) || date.After(currentDate))
}

func isIdValid(id int64) bool {
  return id > 0
}

func deleteId(id int64, store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{Deletes: []int64{id}}
  return store.DoEntryChanges(nil, &changes)
}

func updateId(
    id int64,
    tag uint64,
    mutation fin.EntryUpdater,
    store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{
      Updates: map[int64]fin.EntryUpdater{ id: mutation},
      Etags: map[int64]uint64{ id: tag}}
  return store.DoEntryChanges(nil, &changes)
}

func add(entry *fin.Entry, store findb.DoEntryChangesRunner) error {
  changes := findb.EntryChanges{Adds: []*fin.Entry{entry}}
  return store.DoEntryChanges(nil, &changes)
}

func init() {
  kTemplate = common.NewTemplate("single", kTemplateSpec)
}
