// Package for_sqlite provides a sqlite implementation for storing processed
// QFX file fitIds.
package for_sqlite

import (
  "code.google.com/p/gosqlite/sqlite"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
)

const (
    kSQLByAcctIdFitId = "select acct_id from qfx_fitids where acct_id = ? and fit_id = ?"
    kSQLInsertAcctIdFitId = "insert into qfx_fitids (acct_id, fit_id) values (?, ?)"
)

// New creates sqlite implementation of qfxdb.Store interface
func New(db *sqlite_db.Db) qfxdb.Store {
  return sqliteStore{db}
}

func add(conn *sqlite.Conn, accountId int64, fitIds qfxdb.FitIdSet) error {
  addStmt, err := conn.Prepare(kSQLInsertAcctIdFitId)
  if err != nil {
    return err
  }
  defer addStmt.Finalize()
  for fitId, ok := range fitIds {
    if ok {
      err := addStmt.Exec(accountId, fitId)
      if err != nil {
        return err
      }
      addStmt.Next()
    }
  }
  return nil
}

func find(conn *sqlite.Conn, accountId int64, fitIds qfxdb.FitIdSet) (qfxdb.FitIdSet, error) {
  stmt, err := conn.Prepare(kSQLByAcctIdFitId)
  if err != nil {
    return nil, err
  }
  defer stmt.Finalize()
  var result qfxdb.FitIdSet
  for fitId, ok := range fitIds {
    if ok {
      err := stmt.Exec(accountId, fitId)
      if err != nil {
        return nil, err
      }
      if stmt.Next() {
        if result == nil {
          result = make(qfxdb.FitIdSet)
        }
        result[fitId] = true
      }
    }
  }
  return result, nil
}

type sqliteStore struct {
  db sqlite_db.Doer
}

func (s sqliteStore) Add(
    t db.Transaction, accountId int64, fitIds qfxdb.FitIdSet) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return add(conn, accountId, fitIds)
  })
}

func (s sqliteStore) Find(
    t db.Transaction, accountId int64, fitIds qfxdb.FitIdSet) (found qfxdb.FitIdSet, err error) {
  err = sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) (err error) {
    found, err = find(conn, accountId, fitIds)
    return
  })
  return
}

