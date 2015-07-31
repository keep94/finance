package applyrecurring

import (
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/finance/apps/ledger/common"
  "github.com/keep94/finance/fin/findb"
  "net/http"
)

type Handler struct {
  Doer db.Doer
  Clock date_util.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" {
    fmt.Fprintf(w, "This page does not allow GET requests.")
    return
  }
  session := common.GetUserSession(r)
  store := session.Store.(findb.RecurringEntriesApplier)
  var count int
  err := h.Doer.Do(func(t db.Transaction) error {
    var err error
    count, err = findb.ApplyRecurringEntries(
        t, store, date_util.TimeToDate(h.Clock.Now()))
    return err
  })
  if err != nil {
    http_util.ReportError(
        w, "A database error happened applying recurring transactions.", err)
    return
  }
  fmt.Fprintf(w, "%d new entries added.", count)
}
