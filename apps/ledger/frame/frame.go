package frame

import (
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "html/template"
  "net/http"
)

var (
  kTemplateSpec = `
<html>
<head>
  <title>{{.}}</title>
  <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
</head>
<frameset cols="200, *">
  <frame src="/fin/leftnav" name="leftnav" />
  <frame src="/fin/list" name="main" />
</frameset>
</html>`
)

var (
  kTemplate *template.Template
)

type Handler struct {
  Title string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  http_util.WriteTemplate(w, kTemplate, h.Title)
}

func init() {
  kTemplate = common.NewTemplate("frame", kTemplateSpec)
}
