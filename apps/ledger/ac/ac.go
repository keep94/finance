package ac

import (
	"encoding/json"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/aggregators"
	"github.com/keep94/finance/fin/consumers"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/goconsume"
	"github.com/keep94/toolbox/http_util"
	"net/http"
)

const (
	kMaxAutoComplete = 1000
)

type Handler struct {
	Store findb.EntriesRunner
	Field func(e *fin.Entry) string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	aca := &aggregators.AutoCompleteAggregator{Field: h.Field}
	acc := consumers.FromEntryAggregator(aca)
	acc = goconsume.Slice(acc, 0, kMaxAutoComplete)
	err := h.Store.Entries(nil, nil, acc)
	if err != nil {
		http_util.ReportError(w, "Error reading database.", err)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(aca.Items)
}
