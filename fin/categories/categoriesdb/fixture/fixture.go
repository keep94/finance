// Package fixture provides test suites to test implementations of the
// interfaces in the categoriesdb package.
package fixture

import (
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb"
  "reflect"
  "testing"
)

var (
  kNilCds = categories.CatDetailStore{}
  kNonExistentCat = fin.Cat{Type: fin.ExpenseCat, Id: 9983}
)

// Actual tests must provide an implementation of this Db interface.
type Db interface {

  // Read reads category details from the database and returns them.
  Read(t db.Transaction) (cds categories.CatDetailStore, err error)

  // Add adds a new category to the database and returns the updated
  // category details and the Id of the new category.
  Add(t db.Transaction, cds categories.CatDetailStore, name string) (
      newStore categories.CatDetailStore, newId fin.Cat, err error)

  // Rename renames a category in the database and returns the updated
  // category details.
  Rename(
      t db.Transaction,
      cds categories.CatDetailStore,
      id fin.Cat,
      name string) (
      newStore categories.CatDetailStore, err error)

  // Remove removes a category from the database and returns the updated
  // category details.
  Remove(t db.Transaction, cds categories.CatDetailStore, id fin.Cat) (
      newStore categories.CatDetailStore, err error)
}

type Invalidater interface {
  categoriesdb.Getter
  categoriesdb.Invalidater
}

type Adder interface {
  Invalidater
  categoriesdb.Adder
}

type Renamer interface {
  Invalidater
  categoriesdb.Renamer
}

type Remover interface {
  Invalidater
  categoriesdb.Remover
}

type AccountAdder interface {
  Invalidater
  categoriesdb.AccountAdder
}

type AccountRenamer interface {
  Invalidater
  categoriesdb.AccountRenamer
}

type AccountRemover interface {
  Invalidater
  categoriesdb.AccountRemover
}

type Purger interface {
  categoriesdb.Getter
  categoriesdb.Purger
}

// Fixture tests implementations of interfaces in the categoriesdb package.
// Each exported method is one test.
type Fixture struct {
  Store findb.AddAccountRunner
  Doer db.Doer
  Db
}

func (f *Fixture) CatDetails(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  expense_subcat := detailByFullName(t, cds, "expense:cat:subcat")
  expense_cat := detailByFullName(t, cds, "expense:cat")
  income_subcat := detailByFullName(t, cds, "income:cat:subcat")
  detailByFullName(t, cds, "income:cat")
  if income_subcat.Id() == expense_subcat.Id() {
    t.Error("Income and expense categories should not share Ids")
  }
  if expense_subcat.Id() == expense_cat.Id() {
    t.Error("Different categories should not share the same Id.")
  }
  if !cds.IsChildOf(expense_subcat.Id(), expense_cat.Id()) {
    t.Error("expense:cat:subcat should be a child of expense:cat")
  }
  if output := len(cds.ActiveCatDetails(true)); output != 8 {
    t.Errorf("Expected 8 categories, got %v", output)
  }
}

func (f *Fixture) CatDetailGoodAdd(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  oldCds := cds
  cds, newId := f.addCatDetail(t, cds, "expense:newCat")
  f.verifySameAsDb(t, cds)
  if output := cds.DetailById(newId).FullName(); output != "expense:newCat" {
    t.Errorf("Expected 'expense:newCat', got %v", output)
  }
  if output := oldCds.DetailById(newId).FullName(); output == "expense:newCat" {
    t.Error("Dont expect 'expense:newCat' in original cds")
  }
}

func (f *Fixture) CatDetailsBadAdds(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  if err := f.addCatDetailError(t, cds, "expense"); err != categories.MalformedCategoryName {
    t.Errorf("Expected MalformedCategoryName, got %v", err)
  }
  if err := f.addCatDetailError(t, cds, "expense:noexist:bad"); err != categories.NoParentCategory {
    t.Errorf("Expected NoParentCategory, got %v", err)
  }
  if err := f.addCatDetailError(t, cds, "account:checking:new_account"); err != categories.MalformedCategoryName {
    t.Errorf("Expected MalformedCategoryName, got %v", err)
  }
  if err := f.addCatDetailError(t, cds, "account:new_account"); err != categories.NoParentCategory {
    t.Errorf("Expected NoParentCategory, got %v", err)
  }
  if err := f.addCatDetailError(t, cds, "expense:cat"); err != categories.DuplicateCategoryName {
    t.Errorf("Expected DuplicateCategoryName, got %v", err)
  }
}
  
func (f *Fixture) CatDetailsRename(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  oldCds := cds
  d := detailByFullName(t, cds, "expense:cat")
  cds = f.renameCatDetail(t, cds, d.Id(), "expense:new_cat")
  f.verifySameAsDb(t, cds)
  detailByFullName(t, cds, "expense:new_cat")
  detailByFullName(t, cds, "expense:new_cat:subcat")
  noDetailByFullName(t, cds, "expense:cat")
  noDetailByFullName(t, cds, "expense:cat:subcat")
  noDetailByFullName(t, oldCds, "expense:new_cat")
  noDetailByFullName(t, oldCds, "expense:new_cat:subcat")
  detailByFullName(t, oldCds, "expense:cat")
  detailByFullName(t, oldCds, "expense:cat:subcat")
}

func (f *Fixture) CatDetailsRename2(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  oldCds := cds
  d := detailByFullName(t, cds, "income:cat:subcat")
  cds = f.renameCatDetail(t, cds, d.Id(), "income:subcat")
  f.verifySameAsDb(t, cds)
  detailByFullName(t, cds, "income:cat")
  detailByFullName(t, cds, "income:subcat")
  noDetailByFullName(t, cds, "income:cat:subcat")
  noDetailByFullName(t, oldCds, "income:subcat")
  detailByFullName(t, oldCds, "income:cat:subcat")
}

func (f *Fixture) CatDetailsRenameSame(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  d := detailByFullName(t, cds, "income:cat:subcat")
  cds = f.renameCatDetail(t, cds, d.Id(), "income:cat:subcat")
  f.verifySameAsDb(t, cds)
  detailByFullName(t, cds, "income:cat:subcat")
}

func (f *Fixture) CatDetailsRenameBad(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  d := detailByFullName(t, cds, "expense:cat")
  if err := f.renameCatDetailError(t, cds, d.Id(), "income:foo"); err != categories.InvalidRename {
    t.Errorf("Expected categories.InvalidRename, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, d.Id(), "expense:cat:newCat"); err != categories.InvalidRenameWouldCauseCycle {
    t.Errorf("Expected categories.InvalidRenameWouldCauseCycle, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, fin.Cat{Type: fin.AccountCat, Id: 1}, "expense:cat:newCat"); err != categories.NeedExpenseIncomeCategory {
    t.Errorf("Expected categories.NeedExpenseIncomeCategory, got %v", err)
  }
  d = detailByFullName(t, cds, "income:cat:subcat")
  if err := f.renameCatDetailError(t, cds, d.Id(), "income:cat"); err != categories.DuplicateCategoryName {
    t.Errorf("Expected categories.DuplicateCategoryName, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, d.Id(), "expense:foo"); err != categories.InvalidRename {
    t.Errorf("Expected categories.InvalidRename, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, d.Id(), "income:missingcat:subcat"); err != categories.NoParentCategory {
    t.Errorf("Expected categories.NoParentCategory, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, fin.Cat{}, "expense:food"); err != categories.InvalidRenameWouldCauseCycle {
    t.Errorf("Expected categories.InvalidRenameWouldCauseCycle, got %v", err)
  }
  if err := f.renameCatDetailError(t, cds, fin.Cat{Type: fin.ExpenseCat, Id: 10469}, "expense:food"); err != categories.NoSuchCategory {
    t.Errorf("Expected categories.NoSuchCategory, got %v", err)
  }
  d = detailByFullName(t, cds, "income:cat:subcat")
  if err := f.renameCatDetailError(t, cds, d.Id(), "income"); err != categories.MalformedCategoryName {
    t.Errorf("Expected categories.MalformedCategoryName, got %v", err)
  }
}

func (f *Fixture) RemoveCatDetail(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  oldCds := cds
  d := detailByFullName(t, cds, "income:cat")
  cds = f.removeCatDetail(t, cds, d.Id())
  f.verifySameAsDb(t, cds)
  noDetailByFullName(t, cds, "income:cat")
  noDetailByFullName(t, cds, "income:cat:subcat")
  detailByFullName(t, cds, "expense:cat:subcat")
  detailByFullName(t, oldCds, "income:cat")
  detailByFullName(t, oldCds, "income:cat:subcat")
}

func (f *Fixture) RemoveCatDetail2(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  oldCds := cds
  d := detailByFullName(t, cds, "expense:cat")
  cds = f.removeCatDetail(t, cds, d.Id())
  f.verifySameAsDb(t, cds)
  noDetailByFullName(t, cds, "expense:cat")
  noDetailByFullName(t, cds, "expense:cat:subcat")
  detailByFullName(t, cds, "income:cat:subcat")
  detailByFullName(t, oldCds, "expense:cat")
  detailByFullName(t, oldCds, "expense:cat:subcat")
}

func (f *Fixture) RemoveCatDetailMissing(t *testing.T) {
  f.createAccounts(t)
  cds := categories.CatDetailStore{}
  if err := f.removeCatDetailError(t, cds, kNonExistentCat); err != categories.NoSuchCategory {
    t.Errorf("Expected categories.NoSuchCategory, got %v", err)
  }
}

func (f *Fixture) RemoveCatDetailError(t *testing.T) {
  f.createAccounts(t)
  cds := f.createCatDetails(t)
  if err := f.removeCatDetailError(t, cds, fin.Cat{Type: fin.AccountCat, Id: 1}); err != categories.NeedExpenseIncomeCategory {
    t.Errorf("Expected categories.NeedExpenseIncomeCategory, got %v", err)
  }
  if err := f.removeCatDetailError(t, cds, fin.Cat{}); err != categories.NoParentCategory {
    t.Errorf("Expected categories.NoParentCategory, got %v", err)
  }
}

func (f *Fixture) CacheGet(t *testing.T, cache categoriesdb.Getter) {
  f.createAccounts(t)
  f.createCatDetails(t)
  cds := cacheGet(t, cache)
  detailByFullName(t, cds, "expense:cat")
  verifyCached(t, cache, cds)
}

func (f *Fixture) CatDetailInvalidate(t *testing.T, cache Invalidater) {
  f.createAccounts(t)
  f.createCatDetails(t)
  cds := cacheGet(t, cache)
  cacheInvalidate(t, cache)
  verifyNotCached(t, cache, cds)
}

func (f *Fixture) CacheAdd(t *testing.T, cache Adder) {
  f.createAccounts(t)
  cacheGet(t, cache)
  f.createCatDetails(t)
  cds, cat  := cacheAdd(t, cache, "expense:cat:newOne")
  if detailByFullName(t, cds, "expense:cat:newOne").Id() != cat {
    t.Error("Wrong Id return for new category.")
  }
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheAddError(t *testing.T, cache Adder) {
  f.createAccounts(t)
  f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, _, _ := cache.Add(nil, "expense:asdf:newOne")
  if cds == kNilCds {
    t.Error("Add() should make best effort to refresh cache.")
  } else {
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CacheRename(t *testing.T, cache Renamer) {
  f.createAccounts(t)
  cacheGet(t, cache)
  oldCds := f.createCatDetails(t)
  cds := cacheRename(
      t, cache, 
      detailByFullName(t, oldCds, "expense:cat").Id(), "expense:newOne")
  detailByFullName(t, cds, "expense:newOne")
  noDetailByFullName(t, cds, "expense:cat")
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheRenameError(t *testing.T, cache Renamer) {
  f.createAccounts(t)
  f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, _ := cache.Rename(nil, kNonExistentCat, "expense:newOne")
  if cds == kNilCds {
    t.Error("Rename() should make best effort to refresh cache.")
  } else {
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CacheRemove(t *testing.T, cache Remover) {
  f.createAccounts(t)
  cacheGet(t, cache)
  oldCds := f.createCatDetails(t)
  cds := cacheRemove(t, cache, detailByFullName(t, oldCds, "expense:cat").Id())
  noDetailByFullName(t, cds, "expense:cat")
  verifyCached(t, cache, cds)
}
  
func (f *Fixture) CacheRemoveError(t *testing.T, cache Remover) {
  f.createAccounts(t)
  f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, _ := cache.Remove(nil, kNonExistentCat)
  if cds == kNilCds {
    t.Error("Remove() should make best effort to refresh cache.")
  } else {
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CacheAccountAdd(t *testing.T, cache AccountAdder) {
  f.createAccounts(t)
  cacheGet(t, cache)
  f.createCatDetails(t)
  cds, id := cacheAccountAdd(t, cache, "visa")
  expectedCat := fin.Cat{Id: id, Type: fin.AccountCat}
  if detailByFullName(t, cds, "account:visa").Id() != expectedCat {
    t.Error("Wrong cat Id return for new account.")
  }
  f.verifySameAsDb(t, cds)
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheAccountAddError(t *testing.T, cache AccountAdder) {
  f.createAccounts(t)
  oldCds := f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, _, err := cache.AccountAdd(nil, "checking")
  if err != categories.DuplicateCategoryName {
    t.Errorf("Expected categories.DuplicateCategoryName, got %v", err)
  }
  if cds == kNilCds {
    t.Error("AccountAdd() should make best effort to refresh cache.")
  } else {
    verifySameAsDb2(t, oldCds, cds)
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CacheAccountAddMalformed(
    t *testing.T, cache AccountAdder) {
  f.createAccounts(t)
  f.createCatDetails(t)
  _, _, err := cache.AccountAdd(nil, "account:checking")
  if err != categories.MalformedCategoryName {
    t.Errorf("Expected categories.MalformedCategoryName, got %v", err)
  }
}

func (f *Fixture) CacheAccountRename(t *testing.T, cache AccountRenamer) {
  f.createAccounts(t)
  cacheGet(t, cache)
  oldCds := f.createCatDetails(t)
  catId := detailByFullName(t, oldCds, "account:checking").Id()
  cds := cacheAccountRename(
      t, cache, catId.Id, "newOne")
  detailByFullName(t, cds, "account:newOne")
  noDetailByFullName(t, cds, "account:checking")
  f.verifySameAsDb(t, cds)
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheAccountRenameSame(
    t *testing.T, cache AccountRenamer) {
  f.createAccounts(t)
  cacheGet(t, cache)
  oldCds := f.createCatDetails(t)
  catId := detailByFullName(t, oldCds, "account:checking").Id()
  cds := cacheAccountRename(
      t, cache, catId.Id, "checking")
  verifySameAsDb2(t, oldCds, cds)
  f.verifySameAsDb(t, cds)
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheAccountRenameError(
    t *testing.T, cache AccountRenamer) {
  f.createAccounts(t)
  oldCds := f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, err := cache.AccountRename(nil, 9998, "newOne")
  if err != categories.NoSuchCategory {
    t.Errorf("Expected categories.NoSuchCategory, got %v", err)
  }
  if cds == kNilCds {
    t.Error("AccountRename() should make best effort to refresh cache.")
  } else {
    verifySameAsDb2(t, oldCds, cds)
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CacheAccountRenameError2(
    t *testing.T, cache AccountRenamer) {
  f.createAccounts(t)
  oldCds := f.createCatDetails(t)
  catId := detailByFullName(t, oldCds, "account:checking").Id()
  _, err := cache.AccountRename(nil, catId.Id, "savings")
  if err != categories.DuplicateCategoryName {
    t.Errorf("Expected categories.DuplicateCategoryName, got %v", err)
  }
}

func (f *Fixture) CacheAccountRenameMalformed(
    t *testing.T, cache AccountRenamer) {
  f.createAccounts(t)
  f.createCatDetails(t)
  _, err := cache.AccountRename(nil, 1, "account:checking")
  if err != categories.MalformedCategoryName {
    t.Errorf("Expected categories.MalformedCategoryName, got %v", err)
  }
}

func (f *Fixture) CacheAccountRemove(t *testing.T, cache AccountRemover) {
  f.createAccounts(t)
  cacheGet(t, cache)
  oldCds := f.createCatDetails(t)
  catId := detailByFullName(t, oldCds, "account:checking").Id()
  cds := cacheAccountRemove(t, cache, catId.Id)
  noDetailByFullName(t, cds, "account:checking")
  f.verifySameAsDb(t, cds)
  verifyCached(t, cache, cds)
}

func (f *Fixture) CacheAccountRemoveError(
    t *testing.T, cache AccountRemover) {
  f.createAccounts(t)
  oldCds := f.createCatDetails(t)
  cacheInvalidate(t, cache)
  cds, err := cache.AccountRemove(nil, 9998)
  if err != categories.NoSuchCategory {
    t.Errorf("Expected categories.NoSuchCategory, got %v", err)
  }
  if cds == kNilCds {
    t.Error("AccountRemove() should make best effort to refresh cache.")
  } else {
    verifySameAsDb2(t, oldCds, cds)
    verifyCached(t, cache, cds)
  }
}

func (f *Fixture) CachePurge(t *testing.T, cache Purger) {
  f.createAccounts(t)
  f.createCatDetails(t)
  oldCds := cacheGet(t, cache)
  cat1 := detailByFullName(t, oldCds, "expense:cat").Id()
  cat2 := detailByFullName(t, oldCds, "income:cat:subcat").Id()
  if err := cache.Purge(nil, fin.CatSet{cat1: true, cat2: true}); err != nil {
    t.Fatalf("Got error purging: %v", err)
  }
  cds := cacheGet(t, cache)
  if cds.DetailById(cat1).FullName() == "expense:cat" {
    t.Error("expense:cat should have been purged.")
  }
  if cds.DetailById(cat2).FullName() == "income:cat:subcat" {
    t.Error("income:cat:subcat should have been purged.")
  }
}

func (f *Fixture) verifySameAsDb(
    t *testing.T, cds categories.CatDetailStore) {
  dbcds := f.readCatDetails(t)
  verifySameAsDb2(t, dbcds, cds)
}

func verifySameAsDb2(
    t *testing.T, dbcds, cds categories.CatDetailStore) {
  if !reflect.DeepEqual(cds, dbcds) {
    t.Error("Expect CatDetailStore to be in sync with database.")
  }
}

func verifyCached(
    t *testing.T,
    cache categoriesdb.Getter,
    cds categories.CatDetailStore) {
  if cds != cacheGet(t, cache) {
    t.Error("CatDetailStore should have been cached.")
  }
}

func verifyNotCached(
    t *testing.T,
    cache categoriesdb.Getter,
    cds categories.CatDetailStore) {
  if cds == cacheGet(t, cache) {
    t.Error("Invalidate should force fresh data.")
  }
}

func cacheGet(
    t *testing.T, cache categoriesdb.Getter) categories.CatDetailStore {
  cds, err := cache.Get(nil)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return cds
}

func cacheAdd(
    t *testing.T,
    cache categoriesdb.Adder,
    name string) (cds categories.CatDetailStore, newId fin.Cat) {
  cds, newId, err := cache.Add(nil, name)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return
} 

func cacheRename(
    t *testing.T,
    cache categoriesdb.Renamer,
    id fin.Cat, name string) categories.CatDetailStore {
  cds, err := cache.Rename(nil, id, name)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return cds
} 

func cacheRemove(
    t *testing.T,
    cache categoriesdb.Remover,
    id fin.Cat) categories.CatDetailStore {
  cds, err := cache.Remove(nil, id)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return cds
} 

func cacheAccountAdd(
    t *testing.T,
    cache categoriesdb.AccountAdder,
    name string) (cds categories.CatDetailStore, newId int64) {
  cds, newId, err := cache.AccountAdd(nil, name)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return
}

func cacheAccountRename(
    t *testing.T,
    cache categoriesdb.AccountRenamer,
    id int64,
    name string) categories.CatDetailStore {
  cds, err := cache.AccountRename(nil, id, name)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return cds
}

func cacheAccountRemove(
    t *testing.T,
    cache categoriesdb.AccountRemover,
    id int64) categories.CatDetailStore {
  cds, err := cache.AccountRemove(nil, id)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  return cds
}

func cacheInvalidate(
    t *testing.T, cache categoriesdb.Invalidater) {
  err := cache.Invalidate(nil)
  if err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
} 

func (f *Fixture) createCatDetails(t *testing.T) (
    cds categories.CatDetailStore) {
  err := f.Doer.Do(func(t db.Transaction) (err error) {
    cds, err = f.Read(t)
    if err != nil {
      return
    }
    catNames := []string{
        "expense:cat", "expense:cat:subcat", "income:cat", "income:cat:subcat"}
    for _, catName := range catNames {
      cds, _, err = f.Add(t, cds, catName)
      if err != nil {
        return
      }
    }
    return
  })
  if err != nil {
    t.Fatalf("Got error creating categories: %v", err)
  }
  newCds := f.readCatDetails(t)
  if !reflect.DeepEqual(cds, newCds) {
    t.Fatal("Expected CatDetailStore to be in sync with database after adding categories.")
  }
  return
}

func (f *Fixture) addCatDetail(
    t *testing.T,
    cds categories.CatDetailStore, name string) (
    newStore categories.CatDetailStore, newId fin.Cat) {
  newStore, newId, err := f.Add(nil, cds, name)
  if err != nil {
    t.Fatalf("Got error adding cat details: %v", err)
  }
  return
}

func (f *Fixture) addCatDetailError(
    t *testing.T,
    cds categories.CatDetailStore, name string) error {
  newStore, _, err := f.Add(nil, cds, name)
  if cds != newStore {
    t.Error("newStore and receiver should be equal on error.")
  }
  return err
}

func (f *Fixture) renameCatDetail(
    t *testing.T,
    cds categories.CatDetailStore,
    id fin.Cat,
    name string) categories.CatDetailStore {
  newStore, err := f.Rename(nil, cds, id, name)
  if err != nil {
    t.Fatalf("Got error renaming cat detail: %v", err)
  }
  return newStore
}

func (f *Fixture) renameCatDetailError(
    t *testing.T,
    cds categories.CatDetailStore,
    id fin.Cat,
    name string) error {
  newStore, err := f.Rename(nil, cds, id, name)
  if newStore != cds {
    t.Error("newStore and receiver should be equal on error.")
  }
  return err
}

func (f *Fixture) removeCatDetail(
    t *testing.T,
    cds categories.CatDetailStore,
    id fin.Cat) categories.CatDetailStore {
  newStore, err := f.Remove(nil, cds, id)
  if err != nil {
    t.Fatalf("Got error deleting cat detail, %v", err)
  }
  return newStore
}

func (f *Fixture) removeCatDetailError(
    t *testing.T,
    cds categories.CatDetailStore,
    id fin.Cat) error {
  newStore, err := f.Remove(nil, cds, id)
  if newStore != cds {
    t.Error("newStore and receiver should be equal on error.")
  }
  return err
}

func (f *Fixture) readCatDetails(t *testing.T) categories.CatDetailStore {
  cds, err := f.Read(nil)
  if err != nil {
    t.Fatalf("To error reading cat details: %v", err)
  }
  return cds
}

func (f *Fixture) createAccounts(t *testing.T) {
  err := f.Doer.Do(func(t db.Transaction) error {
    err := f.Store.AddAccount(t, &fin.Account{Name: "checking", Active: true})
    if err != nil {
      return err
    }
    return f.Store.AddAccount(t, &fin.Account{Name: "savings", Active: true})
  })
  if err != nil {
    t.Fatalf("Error creating accounts: %v", err)
  }
}

func detailByFullName(
    t *testing.T, cds categories.CatDetailStore, name string) categories.CatDetail {
  cd, ok := cds.DetailByFullName(name)
  if !ok {
    t.Fatalf("Category does not exist: %v", name)
  }
  return cd
}

func noDetailByFullName(
    t *testing.T, cds categories.CatDetailStore, name string) {
  _, ok := cds.DetailByFullName(name)
  if ok {
    t.Errorf("Category should not exit: %v", name)
  }
}

