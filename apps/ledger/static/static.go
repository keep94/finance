package static

import (
  "github.com/keep94/appcommon/http_util"
  "net/http"
)

func New() http.Handler {
  result := http.NewServeMux()
  http_util.AddStatic(result, "/autocomplete.css", kAutoCompleteCss)
  http_util.AddStatic(result, "/autocomplete-min.js", kAutoCompleteJs)
  http_util.AddStatic(result, "/datasource-min.js", kDataSourceMinJs)
  http_util.AddStatic(result, "/json2.js", kJson2Js)
  http_util.AddStatic(result, "/ledger.js", kLedgerJs)
  http_util.AddStatic(result, "/theme.css", kThemeCss)
  http_util.AddStatic(result, "/yahoo-dom-event.js", kYahooDomEventJs)
  return result
}
