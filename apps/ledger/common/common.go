// Package common provides routines common to all handlers in the ledger
// webapp.
package common

import (
  "fmt"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/appcommon/session_util"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/autoimport"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "github.com/keep94/finance/fin/findb"
  "github.com/gorilla/sessions"
  "html/template"
  "net/http"
  "net/url"
  "strconv"
  "time"
)

const (
  kSessionCookieName = "session-cookie"
)

type RecurringUnitComboBoxType []fin.RecurringUnit

func (r RecurringUnitComboBoxType) ToSelection(s string) *http_util.Selection {
  ivalue, err := strconv.Atoi(s)
  if err != nil {
    return nil
  }
  value, ok := fin.ToRecurringUnit(ivalue)
  if !ok {
    return nil
  }
  return &http_util.Selection{Name: value.String(), Value: s}
}

var (
  // Represents the combo box for recurring units.
  // Implements http_util.SelectionModel
  RecurringUnitComboBox = RecurringUnitComboBoxType{
      fin.Months, fin.Years, fin.Days, fin.Weeks}
)

// NewGorillaSession creates a gorilla session for the finance app
func NewGorillaSession(
    sessionStore sessions.Store, r *http.Request) (*sessions.Session, error) {
  return sessionStore.Get(r, kSessionCookieName)
}

// UserSession represents a session where user is logged in.
type UserSession struct {
  session_util.UserIdSession
  *sessions.Session
  // User is the logged in user or nil if no user logged in
  User *fin.User

  // Main store for accessing entries, accounts, and users
  Store interface{}

  // The category cache
  Cache categoriesdb.Getter

  // Loads QFX files
  QFXLoader autoimport.Loader
}

// CreateUserSession creates a UserSession instance from a gorilla session
// but does not populate the user field of the returned session.
func CreateUserSession(s *sessions.Session) *UserSession {
  return &UserSession{
      UserIdSession: session_util.UserIdSession{s},
      Session: s,
  }
}
      
// NewUserSession creates a UserSession and associates it with the request
// instance. If user not logged in, the User field in returned UserSession is
// nil. Caller must call context.Clear with request instance.
func NewUserSession(
    store findb.UserByIdRunner,
    sessionStore sessions.Store,
    r *http.Request) (*UserSession, error) {
  us, err := session_util.NewUserSession(
      sessionStore,
      r,
      kSessionCookieName,
      func(s *sessions.Session) session_util.UserSession {
        return CreateUserSession(s)
      },
      userGetter{store},
      findb.NoSuchId)
  if err != nil {
    return nil, err
  }
  return us.(*UserSession), nil
}

// GetUserSession gets the UserSession associated with the request.
// It can only be called after successful completion of NewUserSession.
func GetUserSession(r *http.Request) *UserSession {
  return session_util.GetUserSession(r).(*UserSession)
}

func (s *UserSession) SetUser(userPtr interface{}) {
  s.User = userPtr.(*fin.User)
}

// Batch returns the uploaded batch for a particular account ID. Batch
// returns nil if there is no pending batch.
func (s *UserSession) Batch(acctId int64) autoimport.Batch {
  result := s.Values[sessionBatchKeyType(acctId)]
  if result == nil {
    return nil
  }
  return result.(autoimport.Batch)
}

// SetBatch stores a batch in the session under a particular account ID.
// Passing nil for batch indicates there is no batch for the given account ID.
func (s *UserSession) SetBatch(acctId int64, batch autoimport.Batch) {
  if batch == nil {
    delete(s.Values, sessionBatchKeyType(acctId))
  } else {
    s.Values[sessionBatchKeyType(acctId)] = batch
  }
}

// AccountLinker creates URLs to account pages
type AccountLinker struct {
}

// AccountLink returns a URL to the account page with given account Id.
func (a AccountLinker) AccountLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/account",
      "acctId", strconv.FormatInt(id, 10))
}

// UnreconiledLink returns a URL to the unreconciled page with given account Id.
func (a AccountLinker) UnreconciledLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/unreconciled",
      "acctId", strconv.FormatInt(id, 10))
}

// UploadLink returns a URL to the upload page for a given account Id.
func (a AccountLinker) UploadLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/upload",
      "acctId", strconv.FormatInt(id, 10))
}

// RecurringLink returns a URL to the recurring entries page for a given
// account Id.
func (a AccountLinker) RecurringLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/recurringlist",
      "acctId", strconv.FormatInt(id, 10))
}

// EntryLinker creates URLs to the edit entry page.
type EntryLinker struct {
  URL *url.URL
}

// EntryLink returns a URL to the edit entry page. id is the id of the
// entry to be edited.
func (e EntryLinker) EntryLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/single",
      "id", strconv.FormatInt(id, 10),
      "prev", e.URL.String())
}

// NewEntryLink returns a URL to the edit entry page for creating a
// new entry. paymentId is the account ID of the default payment type.
// Passing 0 or a negative number means that user will have to select
// payment in the edit entry page.
func (e EntryLinker) NewEntryLink(paymentId int64) *url.URL {
  return http_util.NewUrl(
      "/fin/single",
      "aid", strconv.FormatInt(paymentId, 10),
      "prev", e.URL.String())
}

// RecurringEntryLinker creates URLs to the edit recurring entry page.
type RecurringEntryLinker struct {
  URL *url.URL
}

// EntryLink returns a URL to the edit recurring entry page.
// id is the id of the recurring entry to be edited.
func (e RecurringEntryLinker) EntryLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/fin/recurringsingle",
      "id", strconv.FormatInt(id, 10),
      "prev", e.URL.String())
}

// NewEntryLink returns a URL to the edit recurring entry page for creating a
// new recurring entry. paymentId is the account ID of the default payment type.
// Passing 0 or a negative number means that user will have to select
// payment in the edit recurring entry page.
func (e RecurringEntryLinker) NewEntryLink(paymentId int64) *url.URL {
  return http_util.NewUrl(
      "/fin/recurringsingle",
      "aid", strconv.FormatInt(paymentId, 10),
      "prev", e.URL.String())
}

// CatDisplayer is used to display categories.
type CatDisplayer struct {
  categories.CatDetailStore
}

// Returns the category name to display for an entry. cp is the CatPayment
// record in the entry.
func (c CatDisplayer) CatName(cp *fin.CatPayment) string {
  return fullName(c.CatDetailStore, cp)
}

// Returns the account name to display for an entry. cp is the CatPayment
// record in the entry.
func (c CatDisplayer) AcctName(cp *fin.CatPayment) string {
  return c.AccountDetailById(cp.PaymentId()).Name()
}

// CatSelectModel returns the model for category dropdowns.
func (c CatDisplayer) CatSelectModel() http_util.SelectModel {
  return catSelectModel{c.CatDetailStore}
}

// AccountSelectModel returns the model for account dropdowns.
func (c CatDisplayer) AccountSelectModel() http_util.SelectModel {
  return accountSelectModel{c.CatDetailStore}
}

// NormalizeYMDStr normalizes a date string.
func NormalizeYMDStr(dateStr string) string {
  if len(dateStr) == 4 {
    return dateStr + "0101"
  }
  if len(dateStr) == 6 {
    return dateStr + "01"
  }
  return dateStr
}

// NewTemplate returns a new template instance. name is the name
// of the template; templateStr is the template string. Returned
// template has FormatDate and FormatUSD defined. 
func NewTemplate(name, templateStr string) *template.Template {
  return template.Must(template.New(name).Funcs(
      template.FuncMap{
          "FormatDate": formatDate,
          "FormatUSD": formatUSD,
          "FormatUSDRaw": fin.FormatUSD}).Parse(templateStr))
}

func formatUSD(amt int64) template.HTML {
  negTemplate := `
      <span class="negative">(%s)</span>`
  positiveTemplate := `
      <span class="positive">%s</span>`
  if amt < 0 {
    return template.HTML(fmt.Sprintf(negTemplate, fin.FormatUSD(-amt)))
  }
  return template.HTML(fmt.Sprintf(positiveTemplate, fin.FormatUSD(amt)))
}

func fullName(cds categories.CatDetailStore, cp *fin.CatPayment) string {
  count := cp.CatRecCount()
  if count > 1 {
    return "SPLIT"
  }
  if count == 0 {
     return "--Missing--"
  }
  return cds.DetailById(cp.CatRecByIndex(0).Id()).FullName()
}

func formatDate(t time.Time) string {
  return t.Format("01/02/2006")
}

type sessionBatchKeyType int64

type catSelectModel struct {
  categories.CatDetailStore
}

func (c catSelectModel) ToSelection(s string) *http_util.Selection {
  cat, err := fin.CatFromString(s)
  if err != nil {
    return nil
  }
  detail := c.DetailById(cat)
  return &http_util.Selection{Name: detail.FullName(), Value: s}
}

type accountSelectModel struct {
  categories.CatDetailStore
}

func (a accountSelectModel) ToSelection(s string) *http_util.Selection {
  id, err := strconv.ParseInt(s, 10, 64)
  if err != nil {
    return nil
  }
  detail := a.AccountDetailById(id)
  return &http_util.Selection{Name: detail.Name(), Value: s}
}

type userGetter struct {
  findb.UserByIdRunner
}

func (g userGetter) GetUser(id int64) (interface{}, error) {
  var user fin.User
  if err := g.UserById(nil, id, &user); err != nil {
    return nil, err
  }
  return &user, nil
}

