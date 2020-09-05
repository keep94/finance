// Package fixture provides test suites to test implementations of the
// interfaces in the findb package.
package fixture

import (
	"errors"
	"fmt"
	"github.com/keep94/appcommon/date_util"
	"github.com/keep94/appcommon/db"
	"github.com/keep94/appcommon/passwords"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/findb"
	"github.com/keep94/goconsume"
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

type EntryByIdWithEtagStore interface {
	MinimalStore
	findb.EntryByIdRunner
	findb.AccountByIdRunner
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

type RecurringEntriesApplier interface {
	findb.RecurringEntriesApplier
	findb.AddRecurringEntryRunner
	findb.RecurringEntryByIdRunner
	findb.AddAccountRunner
	findb.EntriesByAccountIdRunner
	findb.RemoveRecurringEntryByIdRunner
}

func (f EntryAccountFixture) AccountUpdates(t *testing.T, store AccountByIdStore) {
	f.createAccounts(t, store)
	cpb := fin.CatPaymentBuilder{}
	entry1 := fin.Entry{
		CatPayment: cpb.AddCatRec(
			fin.CatRec{fin.NewCat("0:7"), 6000, false}).AddCatRec(
			fin.CatRec{fin.NewCat("2:2"), 2000, false}).SetPaymentId(
			1).SetReconciled(true).Build()}
	entry2 := fin.Entry{
		CatPayment: fin.NewCatPayment(fin.NewCat("0:7"), 3000, false, 1)}
	changes := findb.EntryChanges{Adds: []*fin.Entry{&entry1, &entry2}}
	changeEntries(t, store, &changes)
	verifyAccounts(
		t,
		store,
		&fin.Account{Id: 1, Name: "checking", Active: true, Balance: -11000, RBalance: -8000, Count: 2, RCount: 1, ImportSD: kCheckingSD},
		&fin.Account{Id: 2, Name: "savings", Active: true, Balance: 2000, RBalance: 0, Count: 1, RCount: 0})

	changes = findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			1:    reconcileFunc(2),
			2:    reconcileFunc(1),
			9999: reconcileFunc(1)}}
	changeEntries(t, store, &changes)
	verifyAccounts(
		t,
		store,
		&fin.Account{Id: 1, Name: "checking", Active: true, Balance: -11000, RBalance: -11000, Count: 2, RCount: 2, ImportSD: kCheckingSD},
		&fin.Account{Id: 2, Name: "savings", Active: true, Balance: 2000, RBalance: 2000, Count: 1, RCount: 1})
	changes = findb.EntryChanges{Deletes: []int64{1, 2, 9998}}
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
	entry := fin.Entry{
		Date:       date_util.YMD(2012, 12, 9),
		Name:       "Foo",
		Desc:       "A description",
		CheckNo:    "1356",
		CatPayment: fin.NewCatPayment(fin.NewCat("0:4"), 1234, false, 1),
		Status:     fin.Reviewed}
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
	new_entry := fin.Entry{
		Id:         entry.Id,
		Date:       date_util.YMD(2012, 12, 9),
		Name:       "Foo",
		Desc:       "A description",
		CheckNo:    "1356",
		CatPayment: fin.NewCatPayment(fin.NewCat("0:4"), 1234, false, 1),
		Status:     fin.Reviewed}
	ec = findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			entry.Id: changeTo(&new_entry, true)}}
	changeEntries(t, store, &ec)
	verifyEntries(t, store, &new_entry)
}

func (f EntryAccountFixture) UpdateEntrySkipped(
	t *testing.T, store EntryByIdStore) {
	f.createAccounts(t, store)
	entry := fin.Entry{}
	ec := findb.EntryChanges{Adds: []*fin.Entry{&entry}}
	changeEntries(t, store, &ec)
	new_entry := fin.Entry{
		Id:         entry.Id,
		Date:       date_util.YMD(2012, 12, 9),
		Name:       "Foo",
		Desc:       "A description",
		CheckNo:    "1356",
		CatPayment: fin.NewCatPayment(fin.NewCat("0:4"), 1234, false, 1),
		Status:     fin.Reviewed}
	ec = findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			entry.Id: changeTo(&new_entry, false)}}
	changeEntries(t, store, &ec)
	verifyEntries(t, store, &entry)
}

func (f EntryAccountFixture) ListEntries(t *testing.T, store EntriesStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	fetched_entries := fetchEntries(t, store, nil)
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
	fetched_entries := fetchEntries(t, store, nil)
	verifyEntriesSorted(t, fetched_entries)
	if output := len(fetched_entries); output != 3 {
		t.Errorf("Expected to fetch 3 entries, but fetched %v", output)
	}
}

func (f EntryAccountFixture) ListEntriesEmptyOptions(
	t *testing.T, store EntriesStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	fetched_entries := fetchEntries(t, store, &findb.EntryListOptions{})
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
		End:   ymdPtr(2012, 10, 16)}
	fetched_entries := fetchEntries(t, store, &elo)
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
		Start:      ymdPtr(2012, 10, 15),
		End:        ymdPtr(2012, 10, 16),
		Unreviewed: true}
	fetched_entries := fetchEntries(t, store, &elo)
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
	fetched_entries := fetchEntries(t, store, &elo)
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
	fetched_entries := fetchEntries(t, store, &elo)
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
	fetched_entries := fetchEntries(t, store, &elo)
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
	if err := entriesByAccountIdError(t, store, 9999); err != findb.NoSuchId {
		t.Errorf("Expected NoSuchId, got %v", err)
	}
}

func (f EntryAccountFixture) EntriesByAccountIdNilPtr(
	t *testing.T, store EntriesByAccountIdStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	var entriesWithBalance []fin.EntryBalance
	err := store.EntriesByAccountId(
		nil, 1, nil, goconsume.AppendTo(&entriesWithBalance))
	if err != nil {
		t.Fatalf("Got error reading database: %v", err)
	}
	if output := len(entriesWithBalance); output != 2 {
		t.Fatalf("Expected size 2, got %v", output)
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
	if err := unreconciledEntriesError(t, store, 9999); err != findb.NoSuchId {
		t.Errorf("Expected NoSuchId, got %v", err)
	}
}

func (f EntryAccountFixture) UnreconciledEntriesNoAccount(
	t *testing.T, store UnreconciledEntriesStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	var entries []fin.Entry
	store.UnreconciledEntries(nil, 1, nil, goconsume.AppendTo(&entries))
}

func (f EntryAccountFixture) ConcurrentUpdateDetection(
	t *testing.T, store EntryByIdWithEtagStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	var entryWithEtag fin.Entry
	err := store.EntryById(nil, 2, &entryWithEtag)
	if err != nil {
		t.Errorf("Error reading entry from database: %v", err)
		return
	}
	etag := entryWithEtag.Etag
	var oldAccount1 fin.Account
	var oldAccount2 fin.Account
	err = store.AccountById(nil, 1, &oldAccount1)
	if err != nil {
		t.Errorf("Error reading account 1: %v", err)
		return
	}
	err = store.AccountById(nil, 2, &oldAccount2)
	if err != nil {
		t.Errorf("Error reading account 2: %v", err)
		return
	}

	// entry 2 transfers money from account 2 to account 1
	oldAmt := -entryWithEtag.Total()
	newCp := fin.NewCatPayment(fin.NewCat("2:1"), oldAmt+132, true, 2)
	ec := findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			2: all(
				changeNameFunc("A new name."),
				changeCatPaymentFunc(&newCp))},
		Etags: map[int64]uint64{
			2: etag}}
	err = store.DoEntryChanges(nil, &ec)
	if err != nil {
		t.Errorf("Error updating database: %v", err)
	}
	ec = findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			2: changeNameFunc("Another new name.")},
		Etags: map[int64]uint64{
			2: etag}}
	err = store.DoEntryChanges(nil, &ec)
	if err != findb.ConcurrentUpdate {
		t.Errorf("Expected ConcurrentUpdate error, got %v", err)
	}
	var entry fin.Entry
	err = store.EntryById(nil, 2, &entry)
	if err != nil {
		t.Errorf("Error reading entry from database: %v", err)
	}
	if entry.Name != "A new name." {
		t.Errorf("Expected 'A new name.', got %v", entry.Name)
	}
	var newAccount1 fin.Account
	var newAccount2 fin.Account
	err = store.AccountById(nil, 1, &newAccount1)
	if err != nil {
		t.Errorf("Error reading account 1: %v", err)
		return
	}
	err = store.AccountById(nil, 2, &newAccount2)
	if err != nil {
		t.Errorf("Error reading account 2: %v", err)
		return
	}
	if diff := newAccount1.Balance - oldAccount1.Balance; diff != 132 {
		t.Errorf("expected difference of 132, got %d", diff)
	}
	if diff := newAccount2.Balance - oldAccount2.Balance; diff != -132 {
		t.Errorf("expected difference of -132, got %d", diff)
	}
}

func (f EntryAccountFixture) ConcurrentUpdateSkipped(
	t *testing.T, store EntryByIdStore) {
	f.createAccounts(t, store)
	createListEntries(t, store)
	ec := findb.EntryChanges{
		Updates: map[int64]fin.EntryUpdater{
			2: skipUpdate},
		Etags: map[int64]uint64{ // Etag doesn't match
			2: 9999}}
	err := store.DoEntryChanges(nil, &ec)
	if err != nil {
		t.Errorf("Error updating database: %v", err)
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
		Id:       1,
		Name:     "bobobo",
		Active:   true,
		Balance:  79433,
		RBalance: 75024,
		Count:    4,
		RCount:   3,
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

func (f EntryAccountFixture) ApplyRecurringEntries(
	t *testing.T,
	store RecurringEntriesApplier) {
	f.createAccounts(t, store)

	finiteId := addRecurringEntry(
		t, store, date_util.YMD(2015, 6, 20), 3100, 4)
	infiniteId := addRecurringEntry(
		t, store, date_util.YMD(2015, 8, 10), 4700, 20)
	christmasId := addRecurringEntry(
		t, store, date_util.YMD(2015, 12, 25), 2900, -1)
	newYearId := addRecurringEntry(
		t, store, date_util.YMD(2015, 1, 1), 4400, 0)
	everyTwoWeeksId := addRecurringEntryWithPeriod(
		t, store, date_util.YMD(2015, 9, 24),
		14, fin.Days, 5900, -1)
	var cpb fin.CatPaymentBuilder
	cpb.AddCatRec(fin.CatRec{Cat: fin.NewCat("0:10"), Amount: 10000})
	cpb.AddCatRec(fin.CatRec{Cat: fin.NewCat("2:2"), Amount: 7100})
	cpb.SetPaymentId(1)
	cp := cpb.Build()
	valentinesDayId := addRecurringEntryWithPeriodAndCatPayment(
		t, store, date_util.YMD(2015, 2, 14),
		4, fin.Months, &cp, -1)

	// Apply finiteId once advancing it by 1 month
	var applied bool
	err := f.Doer.Do(func(t db.Transaction) error {
		var err error
		applied, err = findb.ApplyRecurringEntry(t, store, finiteId)
		return err
	})
	if err != nil {
		t.Fatalf("Error applying recurring entry.")
	}
	if !applied {
		t.Error("Expected finiteId entry to be applied.")
	}

	// Skip finiteId once advancing it by 1 month
	var skipped bool
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		skipped, err = findb.SkipRecurringEntry(t, store, finiteId)
		return err
	})
	if err != nil {
		t.Fatalf("Error skipping recurring entry.")
	}
	if !skipped {
		t.Error("Expected finiteId entry to be skipped.")
	}

	// Applying newYearId should return false
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		applied, err = findb.ApplyRecurringEntry(t, store, newYearId)
		return err
	})
	if err != nil {
		t.Fatalf("Error applying recurring entry.")
	}
	if applied {
		t.Error("Expected newYearId entry not to be applied.")
	}

	// Skipping newYearId should return false
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		skipped, err = findb.SkipRecurringEntry(t, store, newYearId)
		return err
	})
	if err != nil {
		t.Fatalf("Error skipping recurring entry.")
	}
	if skipped {
		t.Error("Expected newYearId entry not to be skipped.")
	}

	// Make sure fetching entries is sorted by ID in descending order
	var addedEntries []*fin.RecurringEntry
	if err := store.RecurringEntries(
		nil, goconsume.AppendPtrsTo(&addedEntries)); err != nil {
		t.Fatalf("Error fetching recurring entries: %v", err)
	}
	verifyRecurringEntriesSortedByDate(t, addedEntries)

	// Do dry run
	count, err := findb.ApplyRecurringEntriesDryRun(nil, store, 0, date_util.YMD(2015, 11, 10))
	if err != nil {
		t.Errorf("Got database error doing dry run: %v", err)
	}
	if count != 13 {
		t.Errorf("Expected that 13 entries will be added, got %d", count)
	}

	// Do dry run with only account 2
	count, err = findb.ApplyRecurringEntriesDryRun(nil, store, 2, date_util.YMD(2015, 11, 10))
	if err != nil {
		t.Errorf("Got database error doing dry run: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected that 3 entries will be added, got %d", count)
	}

	// Apply recurring entries for account 2
	count = 0
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		count, err = findb.ApplyRecurringEntries(t, store, 2, date_util.YMD(2015, 11, 10))
		return err
	})
	if err != nil {
		t.Fatalf("Error applying recurring entries.")
	}
	if count != 3 {
		t.Errorf("Expected 3 entries to be added, got %d", count)
	}

	// Apply rest of recurring entries should be 10 = 13 - 3
	count = 0
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		count, err = findb.ApplyRecurringEntries(t, store, 0, date_util.YMD(2015, 11, 10))
		return err
	})
	if err != nil {
		t.Fatalf("Error applying recurring entries.")
	}
	if count != 10 {
		t.Errorf("Expected 10 entries to be added, got %d", count)
	}

	// Test idempotency

	// Do dry run
	count, err = findb.ApplyRecurringEntriesDryRun(nil, store, 0, date_util.YMD(2015, 11, 10))
	if err != nil {
		t.Errorf("Got database error doing dry run: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected that 0 entries will be added, got %d", count)
	}

	// Apply recurring entries
	count = 657
	err = f.Doer.Do(func(t db.Transaction) error {
		var err error
		count, err = findb.ApplyRecurringEntries(t, store, 0, date_util.YMD(2015, 11, 10))
		return err
	})
	if err != nil {
		t.Fatalf("Error applying recurring entries.")
	}
	if count != 0 {
		t.Errorf("Expected 0 entries to be added, got %d", count)
	}

	// Verify recurring entries
	verifyRecurringEntry(t, store, infiniteId, date_util.YMD(2015, 12, 10), 16)
	verifyRecurringEntry(t, store, finiteId, date_util.YMD(2015, 10, 20), 0)
	verifyRecurringEntry(t, store, christmasId, date_util.YMD(2015, 12, 25), -1)
	verifyRecurringEntry(t, store, newYearId, date_util.YMD(2015, 1, 1), 0)
	verifyRecurringEntry(t, store, everyTwoWeeksId, date_util.YMD(2015, 11, 19), -1)
	verifyRecurringEntry(t, store, valentinesDayId, date_util.YMD(2016, 2, 14), -1)

	// verify entries
	verifyEntryDates(t, store, 1,
		-103000, 14,
		date_util.YMD(2015, 11, 10), date_util.YMD(2015, 11, 5),
		date_util.YMD(2015, 10, 22), date_util.YMD(2015, 10, 14),
		date_util.YMD(2015, 10, 10), date_util.YMD(2015, 10, 8),
		date_util.YMD(2015, 9, 24), date_util.YMD(2015, 9, 20),
		date_util.YMD(2015, 9, 10), date_util.YMD(2015, 8, 20),
		date_util.YMD(2015, 8, 10), date_util.YMD(2015, 6, 20),
		date_util.YMD(2015, 6, 14), date_util.YMD(2015, 2, 14))

	// Test removing recurring entries
	deleteRecurringEntry(t, store, infiniteId)
	verifyNoRecurringEntry(t, store, infiniteId)
	deleteRecurringEntry(t, store, finiteId)
	verifyNoRecurringEntry(t, store, finiteId)
	deleteRecurringEntry(t, store, christmasId)
	verifyNoRecurringEntry(t, store, christmasId)
	deleteRecurringEntry(t, store, everyTwoWeeksId)
	verifyNoRecurringEntry(t, store, everyTwoWeeksId)
}

func (f EntryAccountFixture) createAccounts(t *testing.T, store findb.AddAccountRunner) {
	err := f.Doer.Do(func(t db.Transaction) error {
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
		Date:       date_util.YMD(2012, 10, 16),
		Status:     fin.Reviewed,
		CatPayment: fin.NewCatPayment(fin.NewCat("2:1"), 100, true, 2)}
	entry2 := fin.Entry{
		Date:   date_util.YMD(2012, 10, 15),
		Status: fin.Reviewed,
		CatPayment: cpb.AddCatRec(
			fin.CatRec{
				Cat:        fin.NewCat("2:1"),
				Amount:     200,
				Reconciled: true}).SetPaymentId(2).Build()}
	entry3 := fin.Entry{
		Date:   date_util.YMD(2012, 11, 12),
		Status: fin.NotReviewed,
		CatPayment: cpb.AddCatRec(
			fin.CatRec{
				Cat:        fin.NewCat("0:7"),
				Amount:     400,
				Reconciled: true}).SetPaymentId(2).Build()}
	entry4 := fin.Entry{
		Date:   date_util.YMD(2011, 11, 12),
		Status: fin.NotReviewed}
	entry5 := fin.Entry{
		Date:   date_util.YMD(2012, 10, 15),
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
	options *findb.EntryListOptions) []fin.Entry {
	var entries []fin.Entry
	err := store.Entries(nil, options, goconsume.AppendTo(&entries))
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	return entries
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

func initRecurringEntry(
	date time.Time, count int, unit fin.RecurringUnit,
	cp *fin.CatPayment, numLeft int, entry *fin.RecurringEntry) {
	entry.Date = date
	entry.Period.Count = count
	entry.Period.Unit = unit
	entry.NumLeft = numLeft
	entry.CatPayment = *cp
	if entry.Period.Unit == fin.Months {
		entry.Period.DayOfMonth = date.Day()
	}
}

func addRecurringEntry(
	t *testing.T,
	store findb.AddRecurringEntryRunner,
	date time.Time,
	amount int64,
	numLeft int) int64 {
	return addRecurringEntryWithPeriod(
		t, store, date, 1, fin.Months, amount, numLeft)
}

func addRecurringEntryWithPeriod(
	t *testing.T,
	store findb.AddRecurringEntryRunner,
	date time.Time,
	count int,
	unit fin.RecurringUnit,
	amount int64,
	numLeft int) int64 {
	cp := fin.NewCatPayment(fin.Expense, amount, false, 1)
	return addRecurringEntryWithPeriodAndCatPayment(
		t, store, date, count, unit, &cp, numLeft)
}

func addRecurringEntryWithPeriodAndCatPayment(
	t *testing.T,
	store findb.AddRecurringEntryRunner,
	date time.Time,
	count int,
	unit fin.RecurringUnit,
	cp *fin.CatPayment,
	numLeft int) int64 {
	var entry fin.RecurringEntry
	initRecurringEntry(
		date, count, unit, cp, numLeft, &entry)
	if err := store.AddRecurringEntry(nil, &entry); err != nil {
		t.Fatalf("Error creating recurring entries: %v", err)
	}
	return entry.Id
}

func deleteRecurringEntry(
	t *testing.T,
	store findb.RemoveRecurringEntryByIdRunner,
	id int64) {
	if err := store.RemoveRecurringEntryById(nil, id); err != nil {
		t.Fatalf("Error removing recurring entry: %v", err)
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
		expectedWithEtag := *expected
		expectedWithEtag.Etag = actual.Etag
		if !reflect.DeepEqual(&expectedWithEtag, &actual) {
			t.Errorf("Expected %v, got %v", expectedWithEtag, actual)
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
	var entries []fin.EntryBalance
	account := fin.Account{}
	err := store.EntriesByAccountId(
		nil, acct_id, &account, goconsume.AppendTo(&entries))
	if err != nil {
		t.Errorf("Got error reading database: %v", err)
		return
	}
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
	var entries []fin.EntryBalance
	account := fin.Account{}
	return store.EntriesByAccountId(
		nil, acct_id, &account, goconsume.AppendTo(&entries))
}

func verifyUnreconciledEntries(
	t *testing.T,
	store findb.UnreconciledEntriesRunner,
	acct_id int64,
	expected_number int) {
	var entries []fin.Entry
	account := fin.Account{}
	err := store.UnreconciledEntries(
		nil, acct_id, &account, goconsume.AppendTo(&entries))
	if err != nil {
		t.Errorf("Got error reading database: %v", err)
		return
	}
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
	var entries []fin.Entry
	account := fin.Account{}
	return store.UnreconciledEntries(
		nil, acct_id, &account, goconsume.AppendTo(&entries))
}

func verifyEntriesSorted(t *testing.T, entries []fin.Entry) {
	length := len(entries)
	for i := 1; i < length; i++ {
		if entries[i].Date.After(entries[i-1].Date) {
			t.Error("Entries not sorted correctly.")
		}
		if entries[i].Date == entries[i-1].Date && entries[i].Id > entries[i-1].Id {
			t.Error("Entries not sorted correctly.")
		}
	}
}

func verifyRecurringEntriesSortedByDate(
	t *testing.T, entries []*fin.RecurringEntry) {
	length := len(entries)
	for i := 1; i < length; i++ {
		if entries[i].Date.Before(entries[i-1].Date) {
			t.Error("Recurring entries not sorted correctly.")
		}
	}
}

func verifyRecurringEntry(
	t *testing.T,
	store findb.RecurringEntryByIdRunner,
	id int64,
	expectedDate time.Time,
	expectedNumLeft int) {
	var entry fin.RecurringEntry
	if err := store.RecurringEntryById(nil, id, &entry); err != nil {
		t.Fatalf("Error retrieving recurring entry %d: %v", id, err)
	}
	if entry.Date != expectedDate {
		t.Errorf("Expected date %v, got %v", expectedDate, entry.Date)
	}
	if entry.Period.Unit == fin.Months && entry.Period.DayOfMonth != expectedDate.Day() {
		t.Errorf("Expected dayOfMonth %d, got %d", expectedDate.Day(), entry.Period.DayOfMonth)
	}
	if entry.NumLeft != expectedNumLeft {
		t.Errorf("Expected NumLeft %d, got %d", expectedNumLeft, entry.NumLeft)
	}
}

func verifyNoRecurringEntry(
	t *testing.T,
	store findb.RecurringEntryByIdRunner,
	id int64) {
	var entry fin.RecurringEntry
	if err := store.RecurringEntryById(nil, id, &entry); err != findb.NoSuchId {
		t.Errorf("Expected error findb.NoSuchId, got %v", err)
	}
}

func verifyEntryDates(
	t *testing.T,
	store findb.EntriesByAccountIdRunner,
	accountId int64,
	expectedAccountBalance int64,
	expectedCount int,
	expectedDates ...time.Time) {
	var entries []*fin.EntryBalance
	var account fin.Account
	if err := store.EntriesByAccountId(
		nil,
		accountId,
		&account,
		goconsume.AppendPtrsTo(&entries)); err != nil {
		t.Fatalf("Error retrieving added entries: %v", err)
	}
	if len(expectedDates) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(expectedDates), len(entries))
		return
	}
	for i := range entries {
		if entries[i].Date != expectedDates[i] {
			t.Errorf("Expected %v, got %v", expectedDates[i], entries[i].Date)
		}
	}
	if expectedAccountBalance != account.Balance {
		t.Errorf("Expected %v, got %v", expectedAccountBalance, account.Balance)
	}
	if expectedCount != account.Count {
		t.Errorf("Expected %v, got %v", expectedCount, account.Count)
	}
}

func verifyEntryBalanceSorted(t *testing.T, entries []fin.EntryBalance) {
	length := len(entries)
	for i := 1; i < length; i++ {
		if entries[i].Date.After(entries[i-1].Date) {
			t.Error("Entries not sorted correctly.")
		}
		if entries[i].Date == entries[i-1].Date && entries[i].Id > entries[i-1].Id {
			t.Error("Entries not sorted correctly.")
		}
	}
}

func reconcileFunc(id int64) fin.EntryUpdater {
	return func(entry *fin.Entry) bool {
		return entry.Reconcile(id)
	}
}

func all(first, second fin.EntryUpdater) fin.EntryUpdater {
	return func(entry *fin.Entry) bool {
		return first(entry) && second(entry)
	}
}

func changeNameFunc(name string) fin.EntryUpdater {
	return func(entry *fin.Entry) bool {
		entry.Name = name
		return true
	}
}

func changeCatPaymentFunc(cp *fin.CatPayment) fin.EntryUpdater {
	return func(entry *fin.Entry) bool {
		entry.CatPayment = *cp
		return true
	}
}

func changeTo(new_entry *fin.Entry, result bool) fin.EntryUpdater {
	return func(p *fin.Entry) bool {
		id := p.Id
		*p = *new_entry
		p.Id = id
		return result
	}
}

func skipUpdate(entry *fin.Entry) bool {
	return false
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

type LoginStore interface {
	UserByNameStore
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
	var users []fin.User
	err := store.Users(nil, goconsume.AppendTo(&users))
	if err != nil {
		t.Fatalf("Got error reading database: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %v", len(users))
	}
	if users[0].Name != "name1" {
		t.Errorf("Expected name1, got %v", users[0].Name)
	}
	if users[1].Name != "name2" {
		t.Errorf("Expected name2, got %v", users[1].Name)
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

func LoginUser(t *testing.T, doer db.Doer, store LoginStore) {
	createUsersWithFunc(t, store, newUserWithPassword)
	aTime := time.Date(2016, 12, 13, 14, 15, 16, 0, time.UTC)
	var user fin.User
	err := doer.Do(func(t db.Transaction) error {
		return findb.LoginUser(t, store, "name1", "password1", aTime, &user)
	})
	if err != nil {
		t.Fatalf("Got error logging in user")
	}
	if !user.LastLogin.IsZero() {
		t.Error("Expected no last login")
	}

	err = store.UserByName(nil, "name1", &user)
	if err != nil {
		t.Fatalf("Got error reading database: %v", err)
	}
	if user.LastLogin != aTime {
		t.Errorf("Expected last login %v, got %v", aTime, user.LastLogin)
	}

	bTime := time.Date(2017, 1, 2, 3, 4, 5, 0, time.UTC)
	err = doer.Do(func(t db.Transaction) error {
		return findb.LoginUser(t, store, "name1", "wrong_password", bTime, &user)
	})
	if err != findb.WrongPassword {
		t.Errorf("Expected WrongPassword, got %v", err)
	}

	// Login failure should not update last login
	err = store.UserByName(nil, "name1", &user)
	if err != nil {
		t.Fatalf("Got error reading database: %v", err)
	}
	if user.LastLogin != aTime {
		t.Errorf("Expected last login %v, got %v", aTime, user.LastLogin)
	}

	err = doer.Do(func(t db.Transaction) error {
		return findb.LoginUser(t, store, "no_user", "no_password", bTime, &user)
	})
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
	user := fin.User{
		Id: 2, Name: "n2", Password: "p2", Permission: fin.ReadPermission}
	err := store.UpdateUser(nil, &user)
	if err != nil {
		t.Fatalf("Got error updating database: %v", err)
	}
	verifyUser(t, store, &user)
}

func createUsersWithFunc(
	t *testing.T,
	store findb.AddUserRunner,
	newFunc func(int64) *fin.User) {
	for i := 1; i < 3; i++ {
		user := newFunc(int64(i))
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

func createUsers(t *testing.T, store findb.AddUserRunner) {
	createUsersWithFunc(t, store, newUser)
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

func newUserRawPassword(id int64, pwd passwords.Password) *fin.User {
	return &fin.User{
		Id:         id,
		Name:       fmt.Sprintf("name%d", id),
		Password:   pwd,
		Permission: fin.Permission(id) % (fin.NonePermission + 1)}
}

func newUser(id int64) *fin.User {
	return newUserRawPassword(id, passwords.Password(fmt.Sprintf("password%d", id)))
}

func newUserWithPassword(id int64) *fin.User {
	return newUserRawPassword(id, passwords.New(fmt.Sprintf("password%d", id)))
}

func ymdPtr(year, month, day int) *time.Time {
	result := date_util.YMD(year, month, day)
	return &result
}
