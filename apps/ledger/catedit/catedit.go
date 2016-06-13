package catedit

import (
  "fmt"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
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
{{with .AddConfirm}}
  <form method="post">
    <span class="error">An inactive category with that name alrady exists: </span>
    <input type="hidden" name="makeActiveCat" value="{{.Cat}}">
    <input type="hidden" name="name" value="{{.Name}}">
    <input type="submit" name="makeActive" value="Activate existing category">
    <input type="submit" name="addForSure" value="Create new category">
  </form>
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

type adder interface {
  categoriesdb.Adder
  categoriesdb.AccountAdder
}

type renamer interface {
  categoriesdb.Renamer
  categoriesdb.AccountRenamer
}

type addConfirmType struct {
  Cat fin.Cat
  Name string
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  session := common.GetUserSession(r)
  cache := session.Cache.(Cache)
  r.ParseForm()
  if r.Method == "GET" {
    cds, _ := cache.Get(nil)
    http_util.WriteTemplate(
        w,
        kTemplate,
        &view{CatDisplayer: common.CatDisplayer{cds}})
  } else {
    message := ""
    cds, _ := cache.Get(nil)
    cat := fin.NewCat(r.Form.Get("cat"))
    var err error
    var addConfirm *addConfirmType
    if http_util.HasParam(r.Form, "addForSure") {
      name := r.Form.Get("name")
      cds, err = addCategory(cache, name)
      message = fmt.Sprintf("Category %s added.", name)
    } else if http_util.HasParam(r.Form, "makeActive") {
      name := r.Form.Get("name")
      makeActiveCat := fin.NewCat(r.Form.Get("makeActiveCat"))
      cds, err = renameCategory(cache, makeActiveCat, name)
      message = fmt.Sprintf("Category %s activated.", name)
    } else if http_util.HasParam(r.Form, "add") {
      name := r.Form.Get("name")
      detail, exists := cds.InactiveDetailByFullName(name)
      if exists {
        addConfirm = &addConfirmType{
            Cat: detail.Id(), Name: name}
      } else {
        cds, err = addCategory(cache, name)
        message = fmt.Sprintf("Category %s added.", name)
      }
    } else if http_util.HasParam(r.Form, "rename") {
      name := r.Form.Get("name")
      oldName := cds.DetailById(cat).FullName()
      cds, err = renameCategory(cache, cat, name)
      message = fmt.Sprintf(
          "Category %s renamed to %s.", oldName, name)
    } else if http_util.HasParam(r.Form, "remove") {
      oldName := cds.DetailById(cat).FullName()
      if cat.Type == fin.AccountCat {
        cds, err = cache.AccountRemove(nil, cat.Id)
      } else {
        cds, err = cache.Remove(nil, cat)
      }
      message = fmt.Sprintf(
          "Category %s removed.", oldName)
    }
    values := http_util.Values{}
    if err != nil {
      values = http_util.Values{r.Form}
      message = ""
    }
    http_util.WriteTemplate(w, kTemplate, &view{
        CatDisplayer: common.CatDisplayer{cds},
        Values: values,
        Error: err,
        Message: message,
        AddConfirm: addConfirm})
  }
}

func addCategory(cache adder, fullName string) (
    cds categories.CatDetailStore, err error) {
  if strings.HasPrefix(fullName, "account:") {
    cds, _, err = cache.AccountAdd(nil, fullName[8:])
    return
  }
  cds, _, err = cache.Add(nil, fullName)
  return
}

func renameCategory(cache renamer, cat fin.Cat, fullName string) (
    cds categories.CatDetailStore, err error) {
  if cat.Type == fin.AccountCat && strings.HasPrefix(fullName, "account:") {
    cds, err = cache.AccountRename(nil, cat.Id, fullName[8:])
    return
  }
  cds, err = cache.Rename(nil, cat, fullName)
  return
}

type view struct {
  common.CatDisplayer
  http_util.Values
  Error error
  Message string
  AddConfirm *addConfirmType
}

func init() {
  kTemplate = common.NewTemplate("catedit", kTemplateSpec)
}
