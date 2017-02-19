// Package findb contains the persistence layer for the fin package.
package findb

import (
  "errors"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "time"
)

var (
  ConcurrentUpdate = errors.New("findb: Concurrent update.")
  NoSuchId = errors.New("findb: No Such Id.")
  WrongPassword = errors.New("findb: Wrong password.")
  NoPermission = errors.New("findb: Insufficient permission.")
)

type AccountByIdRunner interface {
  // AccountById fetches an account by Id.
  AccountById(t db.Transaction, acctId int64, account *fin.Account) error
}

type AccountsRunner interface {
  // Accounts fetches all accounts.
  Accounts(t db.Transaction, consumer functional.Consumer) error
}

type ActiveAccountsRunner interface {
  // ActiveAccounts fetches all active accounts sorted by name.
  ActiveAccounts(t db.Transaction) (accounts []*fin.Account, err error)
}

type AddAccountRunner interface {
  // AddAccount adds a new account.
  AddAccount(t db.Transaction, Account *fin.Account) error
}

type UpdateAccountImportSDRunner interface {
  // UpdateAccountImportSD updates the import start date of an account.
  UpdateAccountImportSD(
      t db.Transaction, accountId int64, date time.Time) error
}

type UpdateAccountRunner interface {
  // UpdateAccount updates an account.
  UpdateAccount(
      t db.Transaction, account *fin.Account) error
}

type RemoveAccountRunner interface {
  // RemoveAccount removes an account.
  RemoveAccount(t db.Transaction, accountId int64) error
}

type DoEntryChangesRunner interface {
  // DoEntryChanges adds, updates, and deletes entries in bulk.
  DoEntryChanges(t db.Transaction, changes *EntryChanges) error
}

type EntriesRunner interface {
  // Entries gets entries from most to least recent.
  // options is additional options for getting entries, may be nil;
  // consumer consumes the Stream of fetched entries.
  Entries(t db.Transaction, options *EntryListOptions,
      consumer functional.Consumer) error
}

type EntriesByAccountIdRunner interface {
  // EntryByAccountId gets entries by account from most to least recent.
  // acctId is the account ID; account is where
  // Account object is stored; consumer consumes the Stream of EntryBalance
  // values.
  EntriesByAccountId(t db.Transaction, acctId int64,
      account *fin.Account, consumer functional.Consumer) error
}

type EntryByIdRunner interface {
  // EntryById fetches an Entry by id.
  EntryById(t db.Transaction, id int64, entry *fin.Entry) error
}

type EntryByIdWithEtagRunner interface {
  // EntryByIdWithEtag fetches an Entry by id along with its etag.
  EntryByIdWithEtag(
      t db.Transaction, id int64, entry *fin.EntryWithEtag) error
}

type UnreconciledEntriesRunner interface {
  // UnreconciledEntries gets unreconciled entries by account from most to least
  // recent.
  // acctId is the account ID; account, which can be nil, is where
  // Account object is stored; consumer consumes the Stream of Entry values
  UnreconciledEntries(t db.Transaction, acctId int64,
      account *fin.Account, consumer functional.Consumer) error
}

type AddRecurringEntryRunner interface {
  // AddRecurringEntry adds a new recurring entry.
  AddRecurringEntry(t db.Transaction, entry *fin.RecurringEntry) error
}

type UpdateRecurringEntryRunner interface {
  // UpdateRecurringEntry updates a recurring entry.
  UpdateRecurringEntry(t db.Transaction, entry *fin.RecurringEntry) error
}

type RecurringEntryByIdRunner interface {
  // RecurringEntryById gets a recurring entry by id.
  RecurringEntryById(
      t db.Transaction, id int64, entry *fin.RecurringEntry) error
}

type RecurringEntryByIdWithEtagRunner interface {
  // RecurringEntryByIdWithEtag gets a recurring entry by id along with its
  // etag.
  RecurringEntryByIdWithEtag(
      t db.Transaction, id int64, entry *fin.RecurringEntryWithEtag) error
}

type RecurringEntriesRunner interface {
  // RecurringEntries gets all the recurring entries sorted by date
  // in ascending order.
  RecurringEntries(t db.Transaction, consumer functional.Consumer) error
}

type RemoveRecurringEntryByIdRunner interface {
  // RemoveRecurringEntryById removes a recurring entry by id.
  RemoveRecurringEntryById(t db.Transaction, id int64) error
}

type AddUserRunner interface {
  // AddUser adds a new user.
  AddUser(t db.Transaction, user *fin.User) error
}

type UpdateUserRunner interface {
  // UpdateUser updates a user.
  UpdateUser(t db.Transaction, user *fin.User) error
}

type UserByIdRunner interface {
  // UserById gets a user by id.
  UserById(t db.Transaction, id int64, user *fin.User) error
}

type UserByNameRunner interface {
  // UserByName gets a user by name.
  UserByName(t db.Transaction, name string, user *fin.User) error
}

type UsersRunner interface {
  //Users gets all the users sorted by user name.
  Users(t db.Transaction, consumer functional.Consumer) error
}

type RemoveUserByNameRunner interface {
  // RemoveUserByName removes a user by name.
  RemoveUserByName(t db.Transaction, name string) error
}

// EntryChanges represents changes to entries.
type EntryChanges struct {
  // Adds is entries to add
  Adds []*fin.Entry
  // The key is the entry id; the value does the update in-place.
  Updates map[int64]functional.Filterer
  // Deletes is the ids of the entries to delete.
  Deletes []int64
  // Etags contains the etags of the entries being updated.
  // It is used to detect concurrent updates.
  // The key is the entry id; the value is the etag of the original entry.
  // This field is optional, but if present it must contain the etag of
  // each entry being updated.
  Etags map[int64]uint64
}

// EntryListOptions represents options to list entries.
type EntryListOptions struct {
  // If set, entries listed are on or after this date.
  Start *time.Time
  // If set, entries listed are before this date
  End *time.Time
  // If true, show only unreviewed entries
  Unreviewed bool
}

// NoPermissionStore always returns NoPermissionError
type NoPermissionStore struct {
}

func (n NoPermissionStore) AccountById(
    t db.Transaction, acctId int64, account *fin.Account) error {
  return NoPermission
}

func (n NoPermissionStore) Accounts(
    t db.Transaction, consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) ActiveAccounts(
    t db.Transaction) (accounts []*fin.Account, err error) {
  return nil, NoPermission
}

func (n NoPermissionStore) AddAccount(
    t db.Transaction, Account *fin.Account) error {
  return NoPermission
}

func (n NoPermissionStore) UpdateAccountImportSD(
      t db.Transaction, accountId int64, date time.Time) error {
  return NoPermission
}

func (n NoPermissionStore) UpdateAccount(
      t db.Transaction, account *fin.Account) error {
  return NoPermission
}

func (n NoPermissionStore) RemoveAccount(
    t db.Transaction, accountId int64) error {
  return NoPermission
}

func (n NoPermissionStore) DoEntryChanges(
    t db.Transaction, changes *EntryChanges) error {
  return NoPermission
}

func (n NoPermissionStore) Entries( t db.Transaction, options *EntryListOptions,
    consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) EntriesByAccountId(t db.Transaction, acctId int64,
    account *fin.Account, consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) EntryById(
    t db.Transaction, id int64, entry *fin.Entry) error {
  return NoPermission
}

func (n NoPermissionStore) EntryByIdWithEtag(
    t db.Transaction, id int64, entry *fin.EntryWithEtag) error {
  return NoPermission
}

func (n NoPermissionStore) UnreconciledEntries(
    t db.Transaction, acctId int64, account *fin.Account,
    consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) AddRecurringEntry(
    t db.Transaction, entry *fin.RecurringEntry) error {
  return NoPermission
}

func (n NoPermissionStore) UpdateRecurringEntry(
    t db.Transaction, entry *fin.RecurringEntry) error {
  return NoPermission
}

func (n NoPermissionStore) RecurringEntryById(
    t db.Transaction, id int64, entry *fin.RecurringEntry) error {
  return NoPermission
}

func (n NoPermissionStore) RecurringEntryByIdWithEtag(
    t db.Transaction, id int64, entry *fin.RecurringEntryWithEtag) error {
  return NoPermission
}

func (n NoPermissionStore) RecurringEntries(
    t db.Transaction, consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) RemoveRecurringEntryById(
    t db.Transaction, id int64) error {
  return NoPermission
}

func (n NoPermissionStore) AddUser(t db.Transaction, user *fin.User) error {
  return NoPermission
}

func (n NoPermissionStore) UpdateUser(t db.Transaction, user *fin.User) error {
  return NoPermission
}

func (n NoPermissionStore) UserById(t db.Transaction, id int64, user *fin.User) error {
  return NoPermission
}

func (n NoPermissionStore) UserByName(t db.Transaction, name string, user *fin.User) error {
  return NoPermission
}

func (n NoPermissionStore) Users(t db.Transaction, consumer functional.Consumer) error {
  return NoPermission
}

func (n NoPermissionStore) RemoveUserByName(t db.Transaction, name string) error {
  return NoPermission
}

type RecurringEntriesApplier interface {
  DoEntryChangesRunner
  UpdateRecurringEntryRunner
  RecurringEntriesRunner
}

type RecurringEntryApplier interface {
  DoEntryChangesRunner
  UpdateRecurringEntryRunner
  RecurringEntryByIdRunner
}

type RecurringEntrySkipper interface {
  RecurringEntryByIdRunner
  UpdateRecurringEntryRunner
}

// SkipRecurringEntry advances the recurring entry with given id without
// creating a new entry for it.
// t is the database transaction and must be non-nil.
// Returns true if the entry was skipped or false if the NumLeft field
// has already reached 0.
func SkipRecurringEntry(
    t db.Transaction,
    store RecurringEntrySkipper,
    id int64) (bool, error) {
  if t == nil {
    panic("non nil transaction required.")
  }
  var entry fin.RecurringEntry
  if err := store.RecurringEntryById(t, id, &entry); err != nil {
    return false, err
  }
  // If we didn't advance we are done
  if !entry.AdvanceOnce(nil) {
    return false, nil
  }
  if err := store.UpdateRecurringEntry(t, &entry); err != nil {
    return false, err
  }
  return true, nil
}

// ApplyRecurringEntry advances the recurring entry with given id
// creating one new entry for it.
// t is the database transaction and must be non-nil.
// Returns true if the entry was applied or false if the NumLeft field
// has already reached 0.
func ApplyRecurringEntry(
    t db.Transaction,
    store RecurringEntryApplier,
    id int64) (bool, error) {
  if t == nil {
    panic("non nil transaction required.")
  }
  var entry fin.RecurringEntry
  if err := store.RecurringEntryById(t, id, &entry); err != nil {
    return false, err
  }
  var newEntry fin.Entry
  // If we didn't advance we are done
  if !entry.AdvanceOnce(&newEntry) {
    return false, nil
  }
  if err := store.UpdateRecurringEntry(t, &entry); err != nil {
    return false, err
  }
  changes := &EntryChanges{Adds: []*fin.Entry{&newEntry}}
  if err := store.DoEntryChanges(t, changes); err != nil {
    return false, err
  }
  return true, nil
}

// ApplyRecurringEntriesDryRun returns out how many new entries would be
// added to the database if ApplyRecurringEntries were run.
// t is the database transaction.
// store is the database store.
// If acctId is non-zero, ApplyRecurringEntriesDryRun considers only
// the outstanding recurring entries pertaining to that account.
// currentDate is the current date.
func ApplyRecurringEntriesDryRun(
    t db.Transaction,
    store RecurringEntriesRunner,
    acctId int64,
    currentDate time.Time) (int, error) {
  _, entriesToAdd, err := applyRecurringEntriesDryRun(
      t, store, acctId, currentDate)
  return len(entriesToAdd), err
}


// ApplyRecurringEntries applies all outstanding recurring entries
// and returns how many new entries were added to the database as a result.
// If there are no outstanding recurring entries, this function does
// nothing and returns 0. Note that ApplyRecurringEntries is idempotent.
// t is the database transaction and must be non-nil.
// store is the database store.
// If acctId is non-zero, ApplyRecurringEntries applies only the outstanding
// recurring entries pertaining to that account.
// currentDate is the current date.
func ApplyRecurringEntries(
    t db.Transaction,
    store RecurringEntriesApplier,
    acctId int64,
    currentDate time.Time) (int, error) {
  if t == nil {
    panic("non nil transaction required.")
  }
  recurringEntries, entries, err := applyRecurringEntriesDryRun(
      t, store, acctId, currentDate)
  if err != nil {
    return 0, err
  }
  for i := range recurringEntries {
    if err := store.UpdateRecurringEntry(t, recurringEntries[i]); err != nil {
      return 0, err
    }
  }
  changes := &EntryChanges{Adds: entries}
  if err := store.DoEntryChanges(t, changes); err != nil {
    return 0, err
  }
  return len(entries), nil
}

type UpdateUserByNameRunner interface {
  UserByNameRunner
  UpdateUserRunner
}

// LoginUser logs in a user. Caller responsible for setting any cookies
// resulting from login. On success, LoginUser sets logged in user at
// user. If password or userName is wrong, LoginUser returns NoSuchId.
func LoginUser(
    t db.Transaction,
    store UpdateUserByNameRunner,
    userName string,
    password string,
    currentTime time.Time,
    user *fin.User) error {
  if t == nil {
    panic("non nil transaction required.")
  }
  if err := store.UserByName(t, userName, user); err != nil {
    return err
  }
  if !user.Verify(password) {
    return WrongPassword
  }
  if user.Permission == fin.NonePermission {
    return NoSuchId
  }
  newUser := *user
  newUser.LastLogin = currentTime
  if err := store.UpdateUser(t, &newUser); err != nil {
    return err
  }
  return nil
}

func applyRecurringEntriesDryRun(
    t db.Transaction,
    store RecurringEntriesRunner,
    acctId int64,
    currentDate time.Time) (
        recurringEntriesToUpdate []*fin.RecurringEntry,
        entriesToAdd []*fin.Entry,
        err error) {
  consumer := consume.AppendPtrsTo(&recurringEntriesToUpdate, nil)
  if acctId != 0 {
    consumer = functional.FilterConsumer(consumer, accountFilter(acctId))
  }
  if err = store.RecurringEntries(t, consumer); err != nil {
    return
  }
  idx := 0
  for i := range recurringEntriesToUpdate {
    if recurringEntriesToUpdate[i].Advance(currentDate, &entriesToAdd) {
      recurringEntriesToUpdate[idx] = recurringEntriesToUpdate[i]
      idx++
    }
  }
  recurringEntriesToUpdate = recurringEntriesToUpdate[:idx]
  return
}

func accountFilter(acctId int64) functional.Filterer {
  return functional.NewFilterer(func(ptr interface{}) error {
    re := ptr.(*fin.RecurringEntry)
    cp := re.CatPayment
    if !cp.WithPayment(acctId) {
      return functional.Skipped
    }
    return nil
  })
}
