package recurringsingle

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
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
  "time"
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
    <td align="right">Period: </td>
    <td><input type="text" name="count" value="{{.Get "count"}}">&nbsp;
      <select name="unit">
        {{with .GetSelection .RecurringUnitModel "unit"}}
          <option value="{{.Value}}">{{.Name}}
        {{end}}
        {{range .RecurringUnitModel}}
          <option value="{{.ToInt}}">{{.}}
        {{end}}
      </select>
    </td>
  <tr>
  <tr>
    <td align="right">Remaining: </td>
    <td><input type="text" name="remaining" value="{{.Get "remaining"}}">&nbsp;(blank for unlimited)</td>
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
  findb.AddRecurringEntryRunner
  findb.RecurringEntryByIdWithEtagRunner
  findb.UpdateRecurringEntryRunner
  findb.RemoveRecurringEntryByIdRunner
}

type UpdateRecurringEntryRunner interface {
  findb.RecurringEntryByIdWithEtagRunner
  findb.UpdateRecurringEntryRunner
}

type Handler struct {
  Doer db.Doer
  Clock date_util.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  store := session.Store.(Store)
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
        tag, _ := strconv.ParseUint(r.Form.Get("etag"), 10, 64)
        err = h.updateId(id, tag, mutation, store)
      } else {
        var entry fin.RecurringEntry
        mutation.Filter(&entry)
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
        w, kTemplate, toViewFromForm(isIdValid(id), r.Form, cds, err))
  } else {
    http_util.Redirect(w, r, r.Form.Get("prev"))
  }
}

func (h *Handler) doGet(
    w http.ResponseWriter, id, paymentId int64,
    store findb.RecurringEntryByIdWithEtagRunner, cdc categoriesdb.Getter) {
  var v *view
  if isIdValid(id) {
    var entryWithEtag fin.RecurringEntryWithEtag
    var cds categories.CatDetailStore
    err := h.Doer.Do(func(t db.Transaction) (err error) {
      cds, err = cdc.Get(t)
      if err != nil {
        return
      }
      return store.RecurringEntryByIdWithEtag(t, id, &entryWithEtag)
    })
    if err == findb.NoSuchId {
      fmt.Fprintln(w, "No entry found.")
      return
    }
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    v = toView(&entryWithEtag.RecurringEntry, entryWithEtag.Etag, cds)
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
  return date.After(currentDate)
}

func (h *Handler) updateId(
    id int64,
    tag uint64,
    mutation functional.Filterer,
    store UpdateRecurringEntryRunner) error {
  return h.Doer.Do(func(t db.Transaction) (err error) {
    var entryWithEtag fin.RecurringEntryWithEtag
    if err = store.RecurringEntryByIdWithEtag(t, id, &entryWithEtag); err != nil {
      return
    }
    if tag != entryWithEtag.Etag {
      err = findb.ConcurrentUpdate
      return
    }
    mutation.Filter(&entryWithEtag.RecurringEntry)
    return store.UpdateRecurringEntry(t, &entryWithEtag.RecurringEntry)
  })
}

type view struct {
  *common.SingleEntryView
  RecurringUnitModel common.RecurringUnitComboBoxType
}

func isIdValid(id int64) bool {
  return id > 0
}

func toView(
    entry *fin.RecurringEntry,
    tag uint64,
    cds categories.CatDetailStore) *view {
  result := &view{RecurringUnitModel: common.RecurringUnitComboBox}
  result.SingleEntryView = common.ToSingleEntryView(&entry.Entry, tag, cds)
  result.Set("count", strconv.Itoa(entry.Period.Count))
  result.Set("unit", strconv.Itoa(entry.Period.Unit.ToInt()))
  if entry.NumLeft >= 0 {
    result.Set("remaining", strconv.Itoa(entry.NumLeft))
  }
  return result
}

// ShowEntryFromForm shows an entry from form values.
func toViewFromForm(
    existingEntry bool,
    values url.Values,
    cds categories.CatDetailStore,
    err error) *view {
  result := &view{RecurringUnitModel: common.RecurringUnitComboBox}
  result.SingleEntryView = common.ToSingleEntryViewFromForm(
      existingEntry, values, cds, err)
  return result
}

func entryMutation(values url.Values) (mutation functional.Filterer, err error) {
  var entryFilterer functional.Filterer
  if entryFilterer, err = common.EntryMutation(values); err != nil {
    return
  }
  countStr := values.Get("count")
  var count int
  if countStr == "" {
    count = 1
  } else {
    if count, err = strconv.Atoi(countStr); err != nil {
      return
    }
    if count < 1 {
      err = errors.New("Period must be at least 1.")
      return
    }
  }
  unitStr := values.Get("unit")
  var iunit int
  if iunit, err = strconv.Atoi(unitStr); err != nil {
    return
  }
  unit, ok := fin.ToRecurringUnit(iunit)
  if !ok {
    err = errors.New("Invalid recurring unit.")
    return
  }

  numLeftStr := values.Get("remaining")
  var numLeft int
  if numLeftStr == "" {
    numLeft = -1
  } else {
    if numLeft, err = strconv.Atoi(numLeftStr); err != nil {
      return
    }
    if numLeft < 0 {
      err = errors.New("Remaining must be positive.")
      return
    }
  }
  mutation = functional.NewFilterer(func(ptr interface{}) error {
    p := ptr.(*fin.RecurringEntry)
    entryFilterer.Filter(&p.Entry)
    p.CheckNo = ""
    p.Period.Count = count
    p.Period.Unit = unit
    p.NumLeft = numLeft
    return nil
  })
  return
}

func deleteId(id int64, store findb.RemoveRecurringEntryByIdRunner) error {
  return store.RemoveRecurringEntryById(nil, id)
}

func add(
    entry *fin.RecurringEntry, store findb.AddRecurringEntryRunner) error {
  return store.AddRecurringEntry(nil, entry)
}

func init() {
  kTemplate = common.NewTemplate("recurringSingle", kTemplateSpec)
}
