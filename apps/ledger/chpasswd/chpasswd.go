package chpasswd

import (
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/appcommon/passwords"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin/findb"
  "html/template"
  "net/http"
)

const (
  kMinPasswordLength = 6
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>Changing password for {{.Name}}</h2>
{{if .Message}}
  {{if .Success}}
    <font color="#006600"><b>{{.Message}}</b></font>
  {{else}}
    <span class="error">{{.Message}}</span>
  {{end}}
{{end}}
<form method="post">
  <table>
    <tr>
      <td>Old Password: </td>
      <td><input type="password" name="old"></td>
    </tr>
    <tr>
      <td>New Password: </td>
      <td><input type="password" name="new"></td>
    </tr>
    <tr>
      <td>Verify: </td>
      <td><input type="password" name="verify"></td>
    </tr>
  </table>
  <br>
  <input type="submit" value="Change password">
</form>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type UserStore interface {
  findb.UserByIdRunner
  findb.UpdateUserRunner
}

type Handler struct {
  Store UserStore
  Doer db.Doer
}

// Changes user's password. This page must do nothing but change user's
// password. Caution: This page gets full access to user data store
// regardless of logged in user's permissions.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  session := common.GetUserSession(r)
  if r.Method == "GET" {
    http_util.WriteTemplate(w, kTemplate, &view{Name: session.User.Name})
  } else {
    r.ParseForm()
    old := r.Form.Get("old")
    new := r.Form.Get("new")
    verify := r.Form.Get("verify")
    if new != verify {
      http_util.WriteTemplate(
          w,
          kTemplate,
          &view{Name: session.User.Name, Message: "Password re-typed incorrectly."})
      return
    }
    if len(new) < kMinPasswordLength {
      http_util.WriteTemplate(
          w,
          kTemplate,
          &view{
              Name: session.User.Name,
              Message: fmt.Sprintf(
                  "Password must be at least %d characters.",
                  kMinPasswordLength)})
      return
    }
    if !session.User.Verify(old) {
      http_util.WriteTemplate(
          w,
          kTemplate,
          &view{Name: session.User.Name, Message: "Old password wrong."})
      return
    }
    err := h.Doer.Do(func(t db.Transaction) error {
      err := h.Store.UserById(t, session.User.Id, session.User)
      if err != nil {
        return err
      }
      session.User.Password = passwords.New(new)
      return h.Store.UpdateUser(t, session.User)
    })
    if err != nil && err != findb.NoSuchId {
      http_util.ReportError(w, "Error updating database", err)
      return
    }
    http_util.WriteTemplate(
        w,
        kTemplate,
        &view{
            Name: session.User.Name,
            Message: "Password changed successfully.",
            Success: true})
  }
}

type view struct {
  Name string
  Message string
  Success bool
}

func init() {
  kTemplate = common.NewTemplate("chpasswd", kTemplateSpec)
}
