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

const (
	kChPasswd = "chpasswd"
)

var (
	kTemplateSpec = `
<html>
<head>
  <title>{{.Global.Title}}</title>
  {{if .Global.Icon}}
    <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
  {{end}}
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
{{.LeftNav}}
<div class="main">
<h2>Changing password for {{.Name}}</h2>
{{if .Message}}
  {{if .Success}}
    <font color="#006600"><b>{{.Message}}</b></font>
  {{else}}
    <span class="error">{{.Message}}</span>
  {{end}}
{{end}}
<form method="post">
<input type="hidden" name="xsrf" value="{{.Xsrf}}">
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
</div>
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
	Store  UserStore
	Doer   db.Doer
	LN     *common.LeftNav
	Global *common.Global
}

// Changes user's password. This page must do nothing but change user's
// password. Caution: This page gets full access to user data store
// regardless of logged in user's permissions.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	leftnav := h.LN.Generate(w, r, common.SelectChpasswd())
	if leftnav == "" {
		return
	}
	session := common.GetUserSession(r)
	v := &view{
		Name:    session.User.Name,
		Xsrf:    common.NewXsrfToken(r, kChPasswd),
		LeftNav: leftnav,
		Global:  h.Global}
	if r.Method == "GET" {
		writeTemplate(w, v, "")
	} else {
		r.ParseForm()
		if !common.VerifyXsrfToken(r, kChPasswd) {
			writeTemplate(w, v, common.ErrXsrf.Error())
			return
		}
		old := r.Form.Get("old")
		new := r.Form.Get("new")
		verify := r.Form.Get("verify")
		if new != verify {
			writeTemplate(w, v, "Password re-typed incorrectly.")
			return
		}
		if len(new) < kMinPasswordLength {
			writeTemplate(
				w,
				v,
				fmt.Sprintf(
					"Password must be at least %d characters.",
					kMinPasswordLength))
			return
		}
		if !session.User.Verify(old) {
			writeTemplate(w, v, "Old password wrong.")
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
		v.Success = true
		writeTemplate(w, v, "Password changed successfully.")
	}
}

func writeTemplate(w http.ResponseWriter, v *view, message string) {
	v.Message = message
	http_util.WriteTemplate(w, kTemplate, v)
}

type view struct {
	Name    string
	Message string
	Xsrf    string
	Success bool
	LeftNav template.HTML
	Global  *common.Global
}

func init() {
	kTemplate = common.NewTemplate("chpasswd", kTemplateSpec)
}
