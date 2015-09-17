package unreviewed

import (
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/functional"
  "html/template"
  "net/http"
  "net/url"
  "strings"
  "strconv"
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
<script type="text/javascript">
  gActiveCategories = [{{range .ActiveCatDetails true}}"{{.Id}}", "{{.FullName}}",{{end}}];
</script>
<style type="text/css">
.descAutoComplete {
  width:30em;
  padding-bottom:1em;
}
</style>

</head>
<body class="yui-skin-sam">
{{if .ErrorMessage}}
  <span class="error">{{.ErrorMessage}}</span>
{{end}}
{{if .Entries}}
<form method="post">
  <input type="hidden" name="edit_id" value="">
  <input type="submit" name="draft" value="Save Draft">
  <input type="submit" name="final" value="Submit checked entries">
  <table>
    <tr>
      <td>&nbsp;</td>
      <td>Category</td>
      <td>Date</td>
      <td>Name</td>
      <td>Amount</td>
      <td>Account</td>
    </tr>
  {{with $top := .}}
    {{range .Entries}}
    <tr class="lineitem">
      <td>
        <input type="hidden" name="id" value="{{.Id}}">
        <input type="hidden" name="etag_{{.Id}}" value="{{.Etag}}">
        <input type="checkbox" id="checked_{{.Id}}" name="checked_{{.Id}}" {{if $top.InProgress .Status}}checked{{end}}></td>
      <td>
        <select id="cat_{{.Id}}" name="cat_{{.Id}}" onchange="this.form['checked_{{.Id}}'].checked=true">
          <option value="">{{$top.CatName .CatPayment}}</option>
        </select>
        <script type="text/javascript">populateSelect(document.getElementById("cat_{{.Id}}"), gActiveCategories)</script>
      </td>
      <td>{{FormatDate .Date}}</td>
      <td><a href="#" onclick="document.forms[0].edit_id.value={{.Id}}; document.forms[0].submit()">{{.Name}}</a></td>
      <td align=right>{{FormatUSD .Total}}</td>
      <td>{{$top.AcctName .CatPayment}}</td>
      </tr>
      <tr>
        <td>&nbsp;</td>
        <td>
          {{if .CheckNo}}{{.CheckNo}}{{else}}&nbsp;{{end}}
        </td>
        <td colspan=4>
           <div class="descAutoComplete">
            <input type="text" id="desc_{{.Id}}" name="desc_{{.Id}}" value="{{.Desc}}">
         <div id="descContainer_{{.Id}}" /></div>
       </td>
      </tr>
  {{end}}
  {{end}}
  </table>
  <input type="submit" name="draft" value="Save Draft">
  <input type="submit" name="final" value="Submit checked entries">
</form>
<script type="text/javascript">
  var descSuggester = new Suggester("/fin/acdesc");
  var descDs = new YAHOO.util.FunctionDataSource(function() {
    return descSuggester.getSuggestions();
  });
  var autoCompleteFields = [];
  var ids = {{.IdsAsJsArray .Entries}};
  for (var i = 0; ids[i]; i++) {
      var descAutoComplete = new YAHOO.widget.AutoComplete("desc_" + ids[i], "descContainer_" + ids[i], descDs);
      initAutoComplete(descAutoComplete);
      autoCompleteFields.push(descAutoComplete);
  }
</script>
{{else}}
No unreviewed entries.
{{end}}

</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Store interface {
  findb.EntriesRunner
  findb.DoEntryChangesRunner
}

type Handler struct {
  Doer db.Doer
  PageSize int
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  store := session.Store.(Store)
  cache := session.Cache
  message := ""
  if r.Method == "POST" {
    _, isFinal := r.Form["final"]
    ids := r.Form["id"]
    updates := make(map[int64]functional.Filterer, len(ids))
    etags := make(map[int64]uint64, len(ids))
    for _, idStr := range ids {
      id, _ := strconv.ParseInt(idStr, 10, 64)
      updates[id] = createMutation(r.Form, id, isFinal)
      etag, _ := strconv.ParseUint(r.Form.Get(fmt.Sprintf("etag_%d", id)), 10, 64)
      etags[id] = etag
    }
    err := store.DoEntryChanges(nil, &findb.EntryChanges{
        Updates: updates, Etags: etags})
    if err == findb.ConcurrentUpdate {
      message = "You changes were not saved because another user saved while you were editing."
    } else if err == findb.NoPermission {
      message = "Insufficient permission."
    } else if err != nil {
      http_util.ReportError(w, "Error writing to database.", err)
      return
    }
    redirectId, _ := strconv.ParseInt(r.Form.Get("edit_id"), 10, 64)
    if redirectId > 0 {
      entryLinker := common.EntryLinker{r.URL}
      http_util.Redirect(w, r, entryLinker.EntryLink(redirectId).String())
      return
    }
  }
  buffer := consumers.NewEntryBufferWithEtags(h.PageSize)
  cds := categories.CatDetailStore{}
  err := h.Doer.Do(func(t db.Transaction) error {
    cds, _ = cache.Get(t)
    return store.Entries(t, &findb.EntryListOptions{Unreviewed: true}, buffer)
  })
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          http_util.Values{r.Form},
          common.CatDisplayer{cds},
          common.EntryLinker{r.URL},
          buffer.EntriesWithEtags(),
          message})
}

type view struct {
  http_util.Values
  common.CatDisplayer
  common.EntryLinker
  Entries []fin.EntryWithEtag
  ErrorMessage string
}

func (v *view) InProgress(status fin.ReviewStatus) bool {
  return status == fin.ReviewInProgress
}

func (v *view) IdsAsJsArray(entries []fin.EntryWithEtag) template.JS {
  ids := make([]string, len(entries))
  for i := range entries {
    ids[i] = strconv.FormatInt(entries[i].Id, 10)
  }
  return template.JS(fmt.Sprintf("[%s]", strings.Join(ids, ", ")))
}

func createMutation(values url.Values, id int64, isFinal bool) functional.Filterer {
  cat, caterr := fin.CatFromString(values.Get(fmt.Sprintf("cat_%d", id)))
  desc := values.Get(fmt.Sprintf("desc_%d", id))
  var status fin.ReviewStatus = fin.NotReviewed
  if values.Get(fmt.Sprintf("checked_%d", id)) != "" {
    if isFinal {
      status = fin.Reviewed
    } else {
      status = fin.ReviewInProgress
    }
  }
  return functional.NewFilterer(func(ptr interface{}) error {
    p := ptr.(*fin.Entry)
    p.Desc = desc
    if caterr != nil || p.SetSingleCat(cat) {
      p.Status = status
    }
    return nil
  })
}

func init() {
  kTemplate = common.NewTemplate("unreviewed", kTemplateSpec)
}
