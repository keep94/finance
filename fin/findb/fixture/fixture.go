// Package fixture provides test suites to test implementations of the
// interfaces in the findb package.
package fixture

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/date_util"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/passwords"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "reflect"
  "testing"
  "time"
)

var (
 changeError = errors.New("fixture: Error while changing.")
 kCheckingSD = date_util.YMD(2010, 8, 1)
)

// EntryAccountFixture tests implementations of interfaces in the findb
// package that access entries and accounts. Each exported method is one test.
type EntryAccountFixture struct {
  Doer db.Doer
}

type MinimalStore interface {
  findb.AddAccountRunner
  findb.DoEntryChangesRunner
}

type AccountByIdStore interface {
  MinimalStore
  findb.AccountByIdRunner
}

type EntryByIdStore interface {
  MinimalStore
  findb.EntryByIdRunner
}

type EntriesStore interface {
  MinimalStore
  findb.EntriesRunner
}

type EntriesByAccountIdStore interface {
  MinimalStore
  findb.EntriesByAccountIdRunner
}

type UnreconciledEntriesStore interface {
  MinimalStore
  findb.UnreconciledEntriesRunner
}

type ActiveAccountsStore interface {
  MinimalStore
  findb.ActiveAccountsRunner
}

type UpdateAccountImportSDStore interface {
  MinimalStore
  findb.AccountByIdRunner
  findb.UpdateAccountImportSDRunner
}

type UpdateAccountStore interface {
  MinimalStore
  findb.AccountByIdRunner
  findb.UpdateAccountRunner
}

type RemoveAccountStore interface {
  MinimalStore
  findb.AccountByIdRunner
  findb.RemoveAccountRunner
}

func (f EntryAccountFixture) AccountUpdates(t *testing.T, store AccountByIdStore) {
  f.createAccounts(t, store)
  cpb := fin.CatPaymentBuilder{}
  entry1 := fin.Entry{
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:7"), 6000, false}).AddCatRec(
          &fin.CatRec{fin.NewCat("2:2"), 2000, false}).SetPaymentId(
          1).SetReconciled(true).Build()}
  entry2 := fin.Entry{
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:7"), 3000, false}).SetPaymentId(
          1).SetReconciled(false).Build()}
  changes := findb.EntryChanges{Adds: []*fin.Entry{&entry1, &entry2}}
  changeEntries(t, store, &changes)
  verifyAccounts(
      t,
      store,
      &fin.Account{Id: 1, Name: "checking", Active: true, Balance: -11000, RBalance: -8000, Count: 2, RCount: 1, ImportSD: kCheckingSD}, 
      &fin.Account{Id: 2, Name: "savings", Active: true, Balance: 2000, RBalance: 0, Count: 1, RCount: 0})

  changes = findb.EntryChanges{
      Updates: map[int64]functional.Filterer {
          1: reconcileFunc(2),
          2: reconcileFunc(1),
          9999: reconcileFunc(1)}}
  changeEntries(t, store, &changes)
  verifyAccounts(
      t,
      store,
      &fin.Account{Id: 1, Name: "checking", Active: true, Balance: -11000, RBalance: -11000, Count: 2, RCount: 2, ImportSD: kCheckingSD}, 
      &fin.Account{Id: 2, Name: "savings", Active: true, Balance: 2000, RBalance: 2000, Count: 1, RCount: 1})
  changes = findb.EntryChanges{Deletes: []int64 {1, 2, 9998}}
  changeEntries(t, store, &changes)
  verifyAccounts(
      t,
      store,
      &fin.Account{Id: 1, Name: "checking", Active: true, Balance: 0, RBalance: 0, Count: 0, RCount: 0, ImportSD: kCheckingSD}, 
      &fin.Account{Id: 2, Name: "savings", Active: true, Balance: 0, RBalance: 0, Count: 0, RCount: 0})
}

func (f EntryAccountFixture) SaveAndLoadEntry(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  cpb := fin.CatPaymentBuilder{}
  entry := fin.Entry{
      Date: date_util.YMD(2012, 12, 9),
      Name: "Foo",
      Desc: "A description",
      CheckNo: "1356",
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:4"), 1234, false}).SetPaymentId(1).Build(),
      Status: fin.Reviewed}
  ec := findb.EntryChanges{Adds: []*fin.Entry{&entry}}
  changeEntries(t, store, &ec)
  verifyEntries(t, store, &entry)
  verifyNoEntry(t, store, 9999)
}

func (f EntryAccountFixture) UpdateEntry(t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  entry := fin.Entry{}
  ec := findb.EntryChanges{Adds: []*fin.Entry{&entry}}
  changeEntries(t, store, &ec)
  cpb := fin.CatPaymentBuilder{}
  new_entry := fin.Entry{
      Id: entry.Id,
      Date: date_util.YMD(2012, 12, 9),
      Name: "Foo",
      Desc: "A description",
      CheckNo: "1356",
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:4"), 1234, false}).SetPaymentId(1).Build(),
      Status: fin.Reviewed}
  ec = findb.EntryChanges{
      Updates: map[int64]functional.Filterer{
          entry.Id: changeTo(&new_entry)}}
  changeEntries(t, store, &ec)
  verifyEntries(t, store, &new_entry)
}

func (f EntryAccountFixture) UpdateEntryError(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  entry := fin.Entry{}
  ec := findb.EntryChanges{Adds: []*fin.Entry{&entry}}
  changeEntries(t, store, &ec)
  cpb := fin.CatPaymentBuilder{}
  new_entry := fin.Entry{
      Id: entry.Id,
      Date: date_util.YMD(2012, 12, 9),
      Name: "Foo",
      Desc: "A description",
      CheckNo: "1356",
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:4"), 1234, false}).SetPaymentId(1).Build(),
      Status: fin.Reviewed}
  ec = findb.EntryChanges{
      Updates: map[int64]functional.Filterer{
          entry.Id: throwError(changeTo(&new_entry), changeError)}}
  err := store.DoEntryChanges(nil, &ec)
  if err != changeError {
    t.Fatalf("Expected changeError, got %v", err)
  }
  verifyEntries(t, store, &entry)
}

func (f EntryAccountFixture) UpdateEntrySkipped(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  entry := fin.Entry{}
  ec := findb.EntryChanges{Adds: []*fin.Entry{&entry}}
  changeEntries(t, store, &ec)
  cpb := fin.CatPaymentBuilder{}
  new_entry := fin.Entry{
      Id: entry.Id,
      Date: date_util.YMD(2012, 12, 9),
      Name: "Foo",
      Desc: "A description",
      CheckNo: "1356",
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{fin.NewCat("0:4"), 1234, false}).SetPaymentId(1).Build(),
      Status: fin.Reviewed}
  ec = findb.EntryChanges{
      Updates: map[int64]functional.Filterer{
          entry.Id: throwError(changeTo(&new_entry), functional.Skipped)}}
  changeEntries(t, store, &ec)
  verifyEntries(t, store, &entry)
}

func (f EntryAccountFixture) ListEntries(t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  fetched_entries := fetchEntries(t, store, nil, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 5 {
    t.Errorf("Expected to fetch 5 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) DeleteEntries(t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  changes := findb.EntryChanges{Deletes: []int64{1, 9998, 2}}
  changeEntries(t, store, &changes)
  fetched_entries := fetchEntries(t, store, nil, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 3 {
    t.Errorf("Expected to fetch 3 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesEmptyOptions(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  fetched_entries := fetchEntries(t, store, &findb.EntryListOptions{}, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 5 {
    t.Errorf("Expected to fetch 5 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesDateRange(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  elo := findb.EntryListOptions{
      Start: ymdPtr(2012, 10, 15),
      End: ymdPtr(2012, 10, 16)}
  fetched_entries := fetchEntries(t, store, &elo, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 2 {
    t.Errorf("Expected to fetch 2 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesDateRangeAndUnreviewed(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  elo := findb.EntryListOptions{
      Start: ymdPtr(2012, 10, 15),
      End: ymdPtr(2012, 10, 16),
      Unreviewed: true}
  fetched_entries := fetchEntries(t, store, &elo, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 1 {
    t.Errorf("Expected to fetch 1 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesJustStartDate(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  elo := findb.EntryListOptions{
      Start: ymdPtr(2012, 10, 15)}
  fetched_entries := fetchEntries(t, store, &elo, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 4 {
    t.Errorf("Expected to fetch 4 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesJustEndDate(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  elo := findb.EntryListOptions{
      End: ymdPtr(2012, 10, 15)}
  fetched_entries := fetchEntries(t, store, &elo, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 1 {
    t.Errorf("Expected to fetch 1 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) ListEntriesUnreviewed(
    t *testing.T, store EntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  elo := findb.EntryListOptions{
      Unreviewed: true}
  fetched_entries := fetchEntries(t, store, &elo, 10)
  verifyEntriesSorted(t, fetched_entries)
  if output := len(fetched_entries); output != 3 {
    t.Errorf("Expected to fetch 3 entries, but fetched %v", output)
  }
}

func (f EntryAccountFixture) EntriesByAccountId(
    t *testing.T, store EntriesByAccountIdStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  verifyEntriesByAccountId(t, store, 1, 2)
  verifyEntriesByAccountId(t, store, 2, 3)
  if err := entriesByAccountIdError(t, store, 9999); err != findb.NoSuchId  {
    t.Errorf("Expected NoSuchId, got %v", err)
  }
}

func (f EntryAccountFixture) EntriesByAccountIdNilPtr(
    t *testing.T, store EntriesByAccountIdStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  ebb := consumers.NewEntryBalanceBuffer(5)
  err := store.EntriesByAccountId(nil, 1, nil, ebb)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  entriesWithBalance := ebb.EntriesWithBalance()
  if output := len(entriesWithBalance); output != 2 {
    t.Errorf("Expected size 2, got %v", output)
  }
  verifyEntryBalanceSorted(t, entriesWithBalance)
  verifyEntryBalances(
      t, 1, entriesWithBalance[0].Balance, entriesWithBalance)
}

func (f EntryAccountFixture) UnreconciledEntries(
    t *testing.T, store UnreconciledEntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  verifyUnreconciledEntries(t, store, 1, 1)
  verifyUnreconciledEntries(t, store, 2, 2)
  if err := unreconciledEntriesError(t, store, 9999); err != findb.NoSuchId  {
    t.Errorf("Expected NoSuchId, got %v", err)
  }
}

func (f EntryAccountFixture) UnreconciledEntriesNoAccount(
    t *testing.T, store UnreconciledEntriesStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  store.UnreconciledEntries(nil, 1, nil, consumers.NewEntryBuffer(1))
}

func (f EntryAccountFixture) ConcurrentUpdateDetection(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  var entry fin.Entry
  err := store.EntryById(nil, 2, &entry)
  if err != nil {
    t.Errorf("Error reading entry from database: %v", err)
  }
  etag, err := etag.Etag32(&entry)
  if err != nil {
    t.Errorf("Error computing etag: %v", err)
  }
  ec := findb.EntryChanges{
    Updates: map[int64]functional.Filterer {
        2: changeNameFunc("A new name.")},
    Etags: map[int64]uint32 {
        2: etag}}
  err = store.DoEntryChanges(nil, &ec)
  if err != nil {
    t.Errorf("Error updating database: %v", err)
  }
  ec = findb.EntryChanges{
    Updates: map[int64]functional.Filterer {
        2: changeNameFunc("Another new name.")},
    Etags: map[int64]uint32 {
        2: etag}}
  err = store.DoEntryChanges(nil, &ec)
  if err != findb.ConcurrentUpdate {
    t.Errorf("Expected ConcurrentUpdate error, got %v", err)
  }
  err = store.EntryById(nil, 2, &entry)
  if err != nil {
    t.Errorf("Error reading entry from database: %v", err)
  }
  if entry.Name != "A new name." {
    t.Errorf("Expected 'A new name.', got %v", entry.Name)
  }
}

func (f EntryAccountFixture) ConcurrentUpdateSkipped(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  ec := findb.EntryChanges{
    Updates: map[int64]functional.Filterer {
        2: throwError(nil, functional.Skipped)},
    Etags: map[int64]uint32 {  // Etag doesn't match
        2: 9999}}
  err := store.DoEntryChanges(nil, &ec)
  if err != nil {
    t.Errorf("Error updating database: %v", err)
  }
}

func (f EntryAccountFixture) ConcurrentUpdateError(
    t *testing.T, store EntryByIdStore) {
  f.createAccounts(t, store)
  createListEntries(t, store)
  ec := findb.EntryChanges{
    Updates: map[int64]functional.Filterer {
        2: throwError(nil, changeError)},
    Etags: map[int64]uint32 {  // Etag doesn't match
        2: 9999}}
  err := store.DoEntryChanges(nil, &ec)
  if err != changeError {
    t.Errorf("Expected changeError error, got %v", err)
  }
}

func (f EntryAccountFixture) ActiveAccounts(
    t *testing.T, store ActiveAccountsStore) {
  f.createAccounts(t, store)
  accounts, err := store.ActiveAccounts(nil)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  expected := []*fin.Account{
      {Id: 1, Name: "checking", Active: true, ImportSD: kCheckingSD},
      {Id: 2, Name: "savings", Active: true}}
  if !reflect.DeepEqual(expected, accounts) {
    t.Errorf("Expected %v, got %v", expected, accounts)
  }
}

func (f EntryAccountFixture) UpdateAccountImportSD(
    t *testing.T, store UpdateAccountImportSDStore) {
  f.createAccounts(t, store)
  expected := date_util.YMD(2013, 4, 10)
  if output := store.UpdateAccountImportSD(nil, 1, expected); output != nil {
    t.Errorf("Got error updating database, %v", output)
  }
  account := fin.Account{}
  store.AccountById(nil, 1, &account)
  if account.ImportSD != expected {
    t.Errorf("Expected %v, got %v", expected, account.ImportSD)
  }
}

func (f EntryAccountFixture) UpdateAccount(
    t *testing.T, store UpdateAccountStore) {
  f.createAccounts(t, store)
  account := fin.Account{
      Id: 1,
      Name: "bobobo",
      Active: true,
      Balance: 79433,
      RBalance: 75024,
      Count: 4,
      RCount: 3,
      ImportSD: date_util.YMD(2014, 5, 26)}
  if output := store.UpdateAccount(nil, &account); output != nil {
    t.Errorf("Got error updating database, %v", output)
  }
  var account2 fin.Account
  store.AccountById(nil, 1, &account2)
  if account2 != account {
    t.Errorf("Expected %v, got %v", account, account2)
  }
}

func (f EntryAccountFixture) RemoveAccount(
    t *testing.T, store RemoveAccountStore) {
  f.createAccounts(t, store)
  if output := store.RemoveAccount(nil, 1); output != nil {
    t.Errorf("Got error updating database, %v", output)
  }
  var account2 fin.Account
  err := store.AccountById(nil, 1, &account2)
  if err != findb.NoSuchId {
    t.Error("Expected account to be removed.")
  }
}

func (f EntryAccountFixture) createAccounts(t *testing.T, store findb.AddAccountRunner) {
  err := f.Doer.Do(func (t db.Transaction) error {
    err := store.AddAccount(t, &fin.Account{
        Name: "checking", Active: true, ImportSD: kCheckingSD})
    if err != nil {
      return err
    }
    err = store.AddAccount(t, &fin.Account{Name: "savings", Active: true})
    if err != nil {
      return err
    }
    return store.AddAccount(t, &fin.Account{Name: "inactive", Active: false})
  })
  if err != nil {
    t.Fatalf("Error creating accounts: %v", err)
  }
}

func createListEntries(t *testing.T, store findb.DoEntryChangesRunner) {
  cpb := fin.CatPaymentBuilder{}
  entry1 := fin.Entry{
      Date: date_util.YMD(2012, 10, 16),
      Status: fin.Reviewed,
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{C: fin.NewCat("2:1"), A: 100, R: false}).SetPaymentId(
          2).SetReconciled(true).Build()}
  entry2 := fin.Entry{
      Date: date_util.YMD(2012, 10, 15),
      Status: fin.Reviewed,
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{C: fin.NewCat("2:1"), A: 200, R: true}).SetPaymentId(
          2).Build()}
  entry3 := fin.Entry{
      Date: date_util.YMD(2012, 11, 12),
      Status: fin.NotReviewed,
      CatPayment: cpb.AddCatRec(
          &fin.CatRec{C: fin.NewCat("0:7"), A: 400, R: true}).SetPaymentId(
          2).Build()}
  entry4 := fin.Entry{
      Date: date_util.YMD(2011, 11, 12),
      Status: fin.NotReviewed}
  entry5 := fin.Entry{
      Date: date_util.YMD(2012, 10, 15),
      Status: fin.ReviewInProgress}
  ec := findb.EntryChanges{
      Adds: []*fin.Entry{&entry1, &entry2, &entry3, &entry4, &entry5}}
  err := store.DoEntryChanges(nil, &ec)
  if err != nil {
    t.Fatalf("Got error adding entries: %v", err)
  }
}

func fetchEntries(
    t *testing.T,
    store findb.EntriesRunner,
    options *findb.EntryListOptions,
    capacity int) []fin.Entry {
  eb := consumers.NewEntryBuffer(capacity)
  err := store.Entries(nil, options, eb)
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  return eb.Entries()
}

func changeEntries(
    t *testing.T,
    store findb.DoEntryChangesRunner,
    ec *findb.EntryChanges) {
  err := store.DoEntryChanges(nil, ec)
  if err != nil {
    t.Fatalf("Got error changing entries %v", err)
  }
}

func verifyAccounts(
    t *testing.T,
    store findb.AccountByIdRunner,
    accounts ...*fin.Account) {
  var actual fin.Account
  for _, expected := range accounts {
    err := store.AccountById(nil, expected.Id, &actual)
    if err != nil {
      t.Errorf("Expected no error, got %v", err)
      return
    }
    if *expected != actual {
      t.Errorf("Expected %v, got %v", *expected, actual)
    }
  }
}

func verifyEntries(
    t *testing.T,
    store findb.EntryByIdRunner,
    entries ...*fin.Entry) {
  var actual fin.Entry
  for _, expected := range entries {
    err := store.EntryById(nil, expected.Id, &actual)
    if err != nil {
      t.Errorf("Expected no error, got %v", err)
      return
    }
    if !reflect.DeepEqual(expected, &actual) {
      t.Errorf("Expected %v, got %v", *expected, actual)
    }
  }
}

func verifyNoEntry(t *testing.T, store findb.EntryByIdRunner, id int64) {
  actual := fin.Entry{}
  err := store.EntryById(nil, id, &actual)
  if err != findb.NoSuchId {
    t.Errorf("Expected NoSuchId, got %v", err)
  }
}

func verifyEntriesByAccountId(
    t *testing.T,
    store findb.EntriesByAccountIdRunner,
    acct_id int64,
    expected_number int) {
  ebb := consumers.NewEntryBalanceBuffer(2 * expected_number + 1)
  account := fin.Account{}
  err := store.EntriesByAccountId(nil, acct_id, &account, ebb)
  if err != nil {
    t.Errorf("Got error reading database: %v", err)
    return
  }
  entries := ebb.EntriesWithBalance()
  if len(entries) != expected_number {
    t.Errorf("Expected %v entries, got %v", expected_number, len(entries))
  }
  verifyEntryBalanceSorted(t, entries)
  verifyEntryBalances(t, acct_id, account.Balance, entries)
}

func entriesByAccountIdError(
    t *testing.T,
    store findb.EntriesByAccountIdRunner,
    acct_id int64) error {
  ebb := consumers.NewEntryBalanceBuffer(10)
  account := fin.Account{}
  return store.EntriesByAccountId(nil, acct_id, &account, ebb)
}
    
func verifyUnreconciledEntries(
    t *testing.T,
    store findb.UnreconciledEntriesRunner,
    acct_id int64,
    expected_number int) {
  eb := consumers.NewEntryBuffer(2 * expected_number + 1)
  account := fin.Account{}
  err := store.UnreconciledEntries(nil, acct_id, &account, eb)
  if err != nil {
    t.Errorf("Got error reading database: %v", err)
    return
  }
  entries := eb.Entries()
  if len(entries) != expected_number {
    t.Errorf("Expected %v entries, got %v", expected_number, len(entries))
  }
  verifyEntriesSorted(t, entries)
  var sum int64
  for i := range entries {
    entry := &entries[i]
    if entry.PaymentId() != acct_id {
      t.Errorf("Expected account %v, got %v", acct_id, entry.PaymentId())
    }
    if entry.Reconciled() {
      t.Error("Expected only unreconciled entries.")
    }
    sum += entry.Total()
  }
  if output := account.RBalance + sum; output != account.Balance {
    t.Errorf("Expected balance %v, got %v", account.Balance, output)
  }
}
    
func unreconciledEntriesError(
    t *testing.T,
    store findb.UnreconciledEntriesRunner,
    acct_id int64) error {
  eb := consumers.NewEntryBuffer(10)
  account := fin.Account{}
  return store.UnreconciledEntries(nil, acct_id, &account, eb)
}

func verifyEntriesSorted(t *testing.T, entries []fin.Entry) {
  length := len(entries)
  for i := 1; i < length; i++ {
    if entries[i].Date.After(entries[i - 1].Date) {
      t.Error("Entries not sorted correctly.")
    }
    if entries[i].Date == entries[i - 1].Date && entries[i].Id > entries[i - 1].Id {
      t.Error("Entries not sorted correctly.")
    }
  }
}

func verifyEntryBalanceSorted(t *testing.T, entries []fin.EntryBalance) {
  length := len(entries)
  for i := 1; i < length; i++ {
    if entries[i].Date.After(entries[i - 1].Date) {
      t.Error("Entries not sorted correctly.")
    }
    if entries[i].Date == entries[i - 1].Date && entries[i].Id > entries[i - 1].Id {
      t.Error("Entries not sorted correctly.")
    }
  }
}

func reconcileFunc(id int64) functional.Filterer {
  return functional.NewFilterer(
    func(ptr interface{}) error {
      entry := ptr.(*fin.Entry)
      if !entry.Reconcile(id) {
        return functional.Skipped
      }
      return nil
    })
}

func changeNameFunc(name string) functional.Filterer {
  return functional.NewFilterer(
    func(ptr interface{}) error {
      entry := ptr.(*fin.Entry)
      entry.Name = name
      return nil
    })
}

func changeTo(new_entry *fin.Entry) functional.Filterer {
  return functional.NewFilterer(
    func(ptr interface{}) error {
      p := ptr.(*fin.Entry)
      id := p.Id
      *p = *new_entry
      p.Id = id
      return nil
    })
}

func throwError(f functional.Filterer, e error) functional.Filterer {
  return functional.NewFilterer(
    func(ptr interface{}) error {
      if f != nil {
        f.Filter(ptr)
      }
      return e
    })
}

func verifyEntryBalances(
    t *testing.T,
    acct_id int64,
    expected_balance int64,
    entries []fin.EntryBalance) {
  for i := range entries {
    entry := &entries[i]
    if entry.PaymentId() != acct_id {
      t.Errorf("Expected account %v, got %v", acct_id, entry.PaymentId())
    }
    if entry.Balance != expected_balance {
      t.Errorf("Expected balance %v, got %v", expected_balance, entry.Balance)
    }
    expected_balance -= entry.Total()
  }
  if expected_balance != 0 {
    t.Errorf("Expected zero balance initially, got %v", expected_balance)
  }
}

type UserByIdStore interface {
  findb.AddUserRunner
  findb.UserByIdRunner
}
    
type UserByNameStore interface {
  findb.AddUserRunner
  findb.UserByNameRunner
}

type UsersStore interface {
  findb.AddUserRunner
  findb.UsersRunner
}

type RemoveUserByNameStore interface {
  UserByNameStore
  findb.RemoveUserByNameRunner
}

type UpdateUserStore interface {
  UserByIdStore
  findb.UpdateUserRunner
}
    
func UserById(t *testing.T, store UserByIdStore) {
  createUsers(t, store)
  verifyUser(t, store, newUser(1))
  verifyUser(t, store, newUser(2))
  verifyNoUser(t, store, 9999)
}

func UserByName(t *testing.T, store UserByNameStore) {
  createUsers(t, store)
  user := fin.User{}
  err := store.UserByName(nil, "name1", &user)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  expected := *newUser(1)
  if user != expected {
    t.Errorf("Expected %v, got %v", expected, user)
  }
}

func Users(t *testing.T, store UsersStore) {
  createUsers(t, store)
  var names []string
  err := store.Users(nil, functional.MapConsumer(
      consume.AppendTo(&names),
      functional.NewMapper(func(srcPtr, destPtr interface{}) error {
        s := srcPtr.(*fin.User)
        d := destPtr.(*string)
        *d = s.Name
        return nil
      }),
      new(fin.User)))
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  expected := []string {"name1", "name2"}
  if !reflect.DeepEqual(expected, names) {
    t.Errorf("Expected %v, got %v", expected, names)
  }
}

func RemoveUserByName(t *testing.T, store RemoveUserByNameStore) {
  createUsers(t, store)
  user := fin.User{}
  err := store.UserByName(nil, "name1", &user)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  err = store.RemoveUserByName(nil, "name1")
  if err != nil {
    t.Fatalf("Got error removing user: %v", err)
  }
  err = store.UserByName(nil, "name1", &user)
  if err != findb.NoSuchId {
    t.Errorf("Expected NoSuchId, got %v", err)
  }
}

func NoUserByName(t *testing.T, store UserByNameStore) {
  createUsers(t, store)
  user := fin.User{}
  err := store.UserByName(nil, "user9999", &user)
  if err != findb.NoSuchId {
    t.Fatalf("Expected NoSuchId, got %v", err)
  }
}

func UpdateUser(t *testing.T, store UpdateUserStore) {
  createUsers(t, store)
  user := fin.User{Id: 2, Name: "n2", Password: "p2"}
  err := store.UpdateUser(nil, &user)
  if err != nil {
    t.Fatalf("Got error updating database: %v", err)
  }
  verifyUser(t, store, &user)
}

func createUsers(t *testing.T, store findb.AddUserRunner) {
  for i := 1; i < 3; i++ {
    user := newUser(int64(i))
    user.Id = 0
    err := store.AddUser(nil, user)
    if err != nil {
      t.Fatalf("Got error writing to database: %v", err)
    }
    if user.Id == 0 {
      t.Error("Expected user.Id to be set.")
    }
  }
}

func verifyUser(t *testing.T, store findb.UserByIdRunner, expected *fin.User) {
  actual := fin.User{}
  err := store.UserById(nil, expected.Id, &actual)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  if *expected != actual {
    t.Errorf("Expected %v, got %v", *expected, actual)
  }
}
    
func verifyNoUser(t *testing.T, store findb.UserByIdRunner, id int64) {
  actual := fin.User{}
  err := store.UserById(nil, id, &actual)
  if err != findb.NoSuchId {
    t.Errorf("Expected NoSuchId, got: %v", err)
  }
}
    
func newUser(id int64) *fin.User {
  return &fin.User{Id: id, Name: fmt.Sprintf("name%d", id), Password: passwords.Password(fmt.Sprintf("password%d", id))}
}

func ymdPtr(year, month, day int) *time.Time {
  result := date_util.YMD(year, month, day)
  return &result
}
