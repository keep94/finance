// Package findb contains the persistence layer for the fin package.
package findb

import (
  "errors"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/functional"
  "time"
)

var (
  ConcurrentUpdate = errors.New("findb: Concurrent update.")
  NoSuchId = errors.New("findb: No Such Id.")
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

type UnreconciledEntriesRunner interface {
  // UnreconciledEntries gets unreconciled entries by account from most to least
  // recent.
  // acctId is the account ID; account, which can be nil, is where
  // Account object is stored; consumer consumes the Stream of Entry values
  UnreconciledEntries(t db.Transaction, acctId int64,
      account *fin.Account, consumer functional.Consumer) error
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
  Etags map[int64]uint32
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

