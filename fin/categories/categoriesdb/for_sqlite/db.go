// Package for_sqlite stores types in categories package in a sqlite database.
package for_sqlite

import (
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  fsqlite "github.com/keep94/finance/fin/findb/for_sqlite"
  "github.com/keep94/gofunctional3/functional"
  "github.com/keep94/gosqlite/sqlite"
)

// CatDetails populates a CatDetailStore object from the database.
func catDetails(conn *sqlite.Conn) (
    cds categories.CatDetailStore, err error) {
  cdsb := categories.CatDetailStoreBuilder{}
  cdc := categories.CatDetailConsumer{Builder: &cdsb, Type: fin.ExpenseCat}
  if err = expenseCategories(conn, &cdc); err != nil {
    return
  }
  cdc.Type = fin.IncomeCat
  if err = incomeCategories(conn, &cdc); err != nil {
    return
  }
  adc := categories.AccountDetailConsumer{Builder: &cdsb}
  if err = fsqlite.ConnNew(conn).Accounts(nil, &adc); err != nil {
    return
  }
  cds = cdsb.Build()
  return
}

// accountStoreUpdater updates a sqlite database on behalf of a
// fin.CatDetailStore value.
type accountStoreUpdater struct {
  C *sqlite.Conn
}

func (u accountStoreUpdater) Add(name string) (newId int64, err error) {
  account := fin.Account{
      Name: name,
      Active: true,
  }
  if err = fsqlite.ConnNew(u.C).AddAccount(nil, &account); err != nil {
    return
  }
  newId = account.Id
  return
}

func (u accountStoreUpdater) Update(id int64, newName string) error {
  store := fsqlite.ConnNew(u.C)
  var account fin.Account
  err := store.AccountById(nil, id, &account) 
  if err != nil {
    return err
  }
  account.Name = newName
  account.Active = true
  return store.UpdateAccount(nil, &account)
}

func (u accountStoreUpdater) Remove(id int64) error {
  store := fsqlite.ConnNew(u.C)
  var account fin.Account
  err := store.AccountById(nil, id, &account) 
  if err != nil {
    return err
  }
  account.Active = false
  return store.UpdateAccount(nil, &account)
}

// catDetailStoreUpdater updates a sqlite database on behalf of a
// fin.CatDetailStore value.
type catDetailStoreUpdater struct {
  C *sqlite.Conn
}

func (u catDetailStoreUpdater) Add(t fin.CatType, row *categories.CatDbRow) error {
  values, err := sqlite_db.InsertValues(&rawCatDbRow{}, row)
  if err != nil {
    return err
  }
  if t == fin.ExpenseCat {
    err = u.C.Exec("insert into expense_categories (name, is_active, parent_id) values (?, ?, ?)", values...)
  } else if t == fin.IncomeCat {
    err = u.C.Exec("insert into income_categories (name, is_active, parent_id) values (?, ?, ?)", values...)
  } else {
    panic("t must be either ExpenseCat or IncomeCat")
  }
  if err != nil {
    return err
  }
  row.Id, err = sqlite_db.LastRowId(u.C)
  return err
}

func (u catDetailStoreUpdater) Update(t fin.CatType, row *categories.CatDbRow) error {
  values, err := sqlite_db.UpdateValues(&rawCatDbRow{}, row)
  if err != nil {
    return err
  }
  if t == fin.ExpenseCat {
    return u.C.Exec("update expense_categories set name = ?, is_active = ?, parent_id = ? where id = ?", values...)
  } else if t == fin.IncomeCat {
    return u.C.Exec("update income_categories set name = ?, is_active = ?, parent_id = ? where id = ?", values...)
  } else {
    panic("t must be either ExpenseCat or IncomeCat")
  }
  return nil
}

func (u catDetailStoreUpdater) Remove(t fin.CatType, id int64) error {
  if t == fin.ExpenseCat {
    return u.C.Exec("update expense_categories set is_active = 0 where id = ?", id)
  } else if t == fin.IncomeCat {
    return u.C.Exec("update income_categories set is_active = 0 where id = ?", id)
  }
  return categories.NeedExpenseIncomeCategory
}

type rawCatDbRow struct {
  *categories.CatDbRow
  sqlite_db.SimpleRow
}

func (r *rawCatDbRow) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Name, &r.Active, &r.ParentId}
}

func (r *rawCatDbRow) Values() []interface{} {
  return []interface{} {r.Name, r.Active, r.ParentId, r.Id}
}

func (r *rawCatDbRow) Pair(ptr interface{}) {
  r.CatDbRow = ptr.(*categories.CatDbRow)
}

func expenseCategories(conn *sqlite.Conn, consumer functional.Consumer) error {
  stmt, err := conn.Prepare("select id, name, is_active, parent_id from expense_categories")
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  return consumer.Consume(sqlite_db.ReadRows(&rawCatDbRow{}, stmt))
}

func incomeCategories(conn *sqlite.Conn, consumer functional.Consumer) error {
  stmt, err := conn.Prepare("select id, name, is_active, parent_id from income_categories")
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  return consumer.Consume(sqlite_db.ReadRows(&rawCatDbRow{}, stmt))
}
