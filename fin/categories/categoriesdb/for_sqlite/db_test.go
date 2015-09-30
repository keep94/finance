package for_sqlite

import (
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "github.com/keep94/finance/fin/categories/categoriesdb/fixture"
  fsqlite "github.com/keep94/finance/fin/findb/for_sqlite"
  "github.com/keep94/finance/fin/findb/sqlite_setup"
  "github.com/keep94/gosqlite/sqlite"
  "testing"
)

func TestCatDetails(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetails(t)
}

func TestCatDetailGoodAdd(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailGoodAdd(t)
}

func TestCatDetailsBadAdds(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailsBadAdds(t)
}
  
func TestCatDetailsRename(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailsRename(t)
}

func TestCatDetailsRename2(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailsRename2(t)
}

func TestCatDetailsRenameSame(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailsRenameSame(t)
}

func TestCatDetailsRenameBad(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailsRenameBad(t)
}

func TestRemoveCatDetail(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).RemoveCatDetail(t)
}

func TestRemoveCatDetail2(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).RemoveCatDetail2(t)
}

func TestRemoveCatDetailMissing(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).RemoveCatDetailMissing(t)
}

func TestRemoveCatDetailError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).RemoveCatDetailError(t)
}

func TestCacheGet(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheGet(t, New(db))
}

func TestCatDetailInvalidate(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CatDetailInvalidate(t, New(db))
}

func TestCacheAdd(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAdd(t, New(db))
}

func TestCacheAddError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAddError(t, New(db))
}

func TestCacheRename(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheRename(t, New(db))
}
  
func TestCacheRenameError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheRenameError(t, New(db))
}

func TestCacheRemove(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheRemove(t, New(db))
}
  
func TestCacheRemoveError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheRemoveError(t, New(db))
}
  
func TestCacheAccountAdd(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountAdd(t, New(db))
}

func TestCacheAccountAddError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountAddError(t, New(db))
}

func TestCacheAccountAddMalformed(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountAddMalformed(t, New(db))
}

func TestCacheAccountRename(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRename(t, New(db))
}

func TestCacheAccountRenameSame(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRenameSame(t, New(db))
}

func TestCacheAccountRenameError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRenameError(t, New(db))
}

func TestCacheAccountRenameError2(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRenameError2(t, New(db))
}

func TestCacheAccountRenameMalformed(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRenameMalformed(t, New(db))
}

func TestCacheAccountRemove(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRemove(t, New(db))
}

func TestCacheAccountRemoveError(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CacheAccountRemoveError(t, New(db))
}

func TestCachePurge(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  newFixture(db).CachePurge(t, New(db))
}
  
func newFixture(db *sqlite_db.Db) *fixture.Fixture {
  return &fixture.Fixture{
      Store: fsqlite.New(db),
      Doer: sqlite_db.NewDoer(db),
      Db: dbstubb{db}}
}

type dbstubb struct {
  db *sqlite_db.Db
}

func (d dbstubb) Read(t db.Transaction) (
    cds categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(d.db, t).Do(func (conn *sqlite.Conn) (err error) {
    cds, err = catDetails(conn)
    return
  })
  return
}

func (d dbstubb) Add(
    t db.Transaction, cds categories.CatDetailStore, name string) (
      newStore categories.CatDetailStore, newId fin.Cat, err error) {
  err = sqlite_db.ToDoer(d.db, t).Do(func (conn *sqlite.Conn) (err error) {
    newStore, newId, err = cds.Add(name, catDetailStoreUpdater{C: conn})
    return
  })
  return
}

func (d dbstubb) Rename(
    t db.Transaction, cds categories.CatDetailStore, id fin.Cat, name string) (
    newStore categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(d.db, t).Do(func (conn *sqlite.Conn) (err error) {
    newStore, err = cds.Rename(id, name, catDetailStoreUpdater{C: conn})
    return
  })
  return
}

func (d dbstubb) Remove(
    t db.Transaction, cds categories.CatDetailStore, id fin.Cat) (
    newStore categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(d.db, t).Do(func (conn *sqlite.Conn) (err error) {
    newStore, err = cds.Remove(id, catDetailStoreUpdater{C: conn})
    return
  })
  return
}

func openDb(t *testing.T) *sqlite_db.Db {
  conn, err := sqlite.Open(":memory:")
  if err != nil {
    t.Fatalf("Error opening database: %v", err)
  }
  dbase := sqlite_db.New(conn)
  err = dbase.Do(func(conn *sqlite.Conn) error {
    return sqlite_setup.SetUpTables(conn)
  })
  if err != nil {
    t.Fatalf("Error creating tables: %v", err)
  }
  return dbase
}
  
func closeDb(t *testing.T, db *sqlite_db.Db) {
  if err := db.Close(); err != nil {
    t.Errorf("Error closing database: %v", err)
  }
}
