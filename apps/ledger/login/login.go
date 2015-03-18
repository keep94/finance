package login

import (
  "github.com/gorilla/sessions"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
  "html/template"
  "net/http"
)

var (
  kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>Login</h2>
{{if .}}
  <span class="error">{{.}}</span>
{{end}}
<form method="post">
  <table>
    <tr>
      <td>Name: </td>
      <td><input type="text" name="name"></td>
    </tr>
    <tr>
      <td>Password: </td>
      <td><input type="password" name="password"></td>
    </tr>
  </table>
  <br>
  <input type="submit" value="login">
</form>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  SessionStore sessions.Store
  Store findb.UserByNameRunner
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" {
    http_util.WriteTemplate(w, kTemplate, nil)
  } else {
    r.ParseForm()
    userName := r.Form.Get("name")
    password := r.Form.Get("password")
    user := fin.User{}
    err := h.Store.UserByName(nil, userName, &user)
    if err != nil && err != findb.NoSuchId {
      http_util.ReportError(w, "Database error", err)
      return
    }
    if err == findb.NoSuchId || !user.Verify(password) || user.Permission == fin.NonePermission {
      http_util.WriteTemplate(w, kTemplate, "Login incorrect.")
      return
    }
    gs, err := common.NewGorillaSession(h.SessionStore, r)
    if err != nil {
      http_util.ReportError(w, "Error retrieving session", err)
      return
    }
    session := common.CreateUserSession(gs)
    session.SetUserId(user.Id)
    session.ID = ""  // For added security, force a new session ID
    session.Save(r, w)
    http_util.Redirect(w, r, r.Form.Get("prev"))
  }
}

func init() {
  kTemplate = common.NewTemplate("login", kTemplateSpec)
}
