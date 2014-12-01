package catedit

import (
  "fmt"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "html/template"
  "net/http"
  "strings"
)

var (
  kTemplateSpec = `
<html>
  <head>
    <link rel="stylesheet" type="text/css" href="/static/theme.css" />
  </head>
<body onload="parent.leftnav.location.reload()">
{{if .Error}}
  <span class="error">{{.Error.Error}}</span>
{{end}}
{{if .Message}}
  <font color="#006600"><b>{{.Message}}</b></font>
{{end}}
<form method="post">
<table>
  <tr>
    <td>Existing category:</td>
    <td>
      <select name="cat" size=1>
{{with .GetSelection .CatSelectModel "cat"}}
        <option value="{{.Value}}">{{.Name}}</option>
{{end}}
{{range .ActiveCatDetails true}}
        <option value="{{.Id}}">{{.FullName}}</option>
{{end}}
      </select>
    </td>
  </tr>
  <tr>
    <td>New category:</td>
    <td><input type="text" name="name" value="{{.Get "name"}}" size="40"></td>
  </tr>
</table>
<br>
<input type="submit" name="add" value="Add">
<input type="submit" name="rename" value="Rename">
<input type="submit" name="remove" value="Remove" onclick="return confirm('Are you sure you want to remove this category?');">
</form>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Cache interface {
  categoriesdb.Adder
  categoriesdb.Renamer
  categoriesdb.Remover
  categoriesdb.AccountAdder
  categoriesdb.AccountRenamer
  categoriesdb.AccountRemover
  categoriesdb.Getter
}

type Handler struct {
  Cdc Cache
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  if r.Method == "GET" {
    cds, _ := h.Cdc.Get(nil)
    http_util.WriteTemplate(
        w,
        kTemplate,
        &view{CatDisplayer: common.CatDisplayer{cds}})
  } else {
    message := ""
    cds, _ := h.Cdc.Get(nil)
    cat := fin.NewCat(r.Form.Get("cat"))
    var err error
    if http_util.HasParam(r.Form, "add") {
      name := r.Form.Get("name")
      if strings.HasPrefix(name, "account:") {
        cds, _, err = h.Cdc.AccountAdd(nil, name[8:])
      } else {
        cds, _, err = h.Cdc.Add(nil, name)
      }
      message = fmt.Sprintf("Category %s added.", name)
    } else if http_util.HasParam(r.Form, "rename") {
      name := r.Form.Get("name")
      oldName := cds.DetailById(cat).FullName()
      if cat.Type == fin.AccountCat && strings.HasPrefix(name, "account:") {
        cds, err = h.Cdc.AccountRename(nil, cat.Id, name[8:])
      } else {
        cds, err = h.Cdc.Rename(nil, cat, name)
      }
      message = fmt.Sprintf(
          "Category %s renamed to %s.", oldName, name)
    } else if http_util.HasParam(r.Form, "remove") {
      oldName := cds.DetailById(cat).FullName()
      if cat.Type == fin.AccountCat {
        cds, err = h.Cdc.AccountRemove(nil, cat.Id)
      } else {
        cds, err = h.Cdc.Remove(nil, cat)
      }
      message = fmt.Sprintf(
          "Category %s removed.", oldName)
    }
    values := http_util.Values{}
    if err == nil {
      cat = fin.Cat{}
    } else {
      values = http_util.Values{r.Form}
      message = ""
    }
    http_util.WriteTemplate(w, kTemplate, &view{
        CatDisplayer: common.CatDisplayer{cds},
        Values: values,
        Error: err,
        Message: message})
  }
}

type view struct {
  common.CatDisplayer
  http_util.Values
  Error error
  Message string
}

func init() {
  kTemplate = common.NewTemplate("catedit", kTemplateSpec)
}
