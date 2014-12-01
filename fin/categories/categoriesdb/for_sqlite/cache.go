package for_sqlite

import (
  "code.google.com/p/gosqlite/sqlite"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/categories"
  "sync"
)

func New(db *sqlite_db.Db) *Cache {
  return &Cache{db: db}
}

type catDetailCache struct {
  mutex sync.Mutex
  data categories.CatDetailStore
  valid bool
}

func (c *catDetailCache) DbGet(db *sqlite_db.Db) (
    cds categories.CatDetailStore, err error) {
  cds, ok := c.getFromCache()
  if ok {
    return
  }
  err = db.Do(func(conn *sqlite.Conn) (err error) {
    cds, err = c.load(conn)
    return
  })
  return
}

func (c *catDetailCache) Get(conn *sqlite.Conn) (
    cds categories.CatDetailStore, err error) {
  cds, ok := c.getFromCache()
  if ok {
    return
  }
  return c.load(conn)
}

func (c *catDetailCache) Invalidate(conn *sqlite.Conn) error {
  c.mutex.Lock()
  defer c.mutex.Unlock()
  c.valid = false
  return nil
}

func (c *catDetailCache) AccountAdd(conn *sqlite.Conn, name string) (
    cds categories.CatDetailStore, newId int64, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, newId, err = cds.AccountAdd(name, accountStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) AccountRename(
    conn *sqlite.Conn, id int64, name string) (
    cds categories.CatDetailStore, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, err = cds.AccountRename(id, name, accountStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) AccountRemove(
    conn *sqlite.Conn, id int64) (
    cds categories.CatDetailStore, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, err = cds.AccountRemove(id, accountStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) Add(conn *sqlite.Conn, name string) (
    cds categories.CatDetailStore, newId fin.Cat, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, newId, err = cds.Add(name, catDetailStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) Remove(conn *sqlite.Conn, id fin.Cat) (
    cds categories.CatDetailStore, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, err = cds.Remove(id, catDetailStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) Purge(conn *sqlite.Conn, cats fin.CatSet) error {
  expenseStmt, err := conn.Prepare("delete from expense_categories where id = ?")
  if err != nil {
    return err
  }
  defer expenseStmt.Finalize()
  incomeStmt, err := conn.Prepare("delete from income_categories where id = ?")
  if err != nil {
    return err
  }
  defer incomeStmt.Finalize()
  for cat, ok := range cats {
    if ok {
      if cat.Type == fin.ExpenseCat {
        if err := expenseStmt.Exec(cat.Id); err != nil {
          return err
        }
        expenseStmt.Next()
      } else if cat.Type == fin.IncomeCat {
        if err := incomeStmt.Exec(cat.Id); err != nil {
          return err
        }
        incomeStmt.Next()
      } else {
        return categories.NeedExpenseIncomeCategory
      }
    }
  }
  return c.Invalidate(conn)
}

func (c *catDetailCache) Rename(conn *sqlite.Conn, id fin.Cat, newName string) (
    cds categories.CatDetailStore, err error) {
  if cds, err = catDetails(conn); err != nil {
    cds, _ = c.getFromCache()
    return
  }
  cds, err = cds.Rename(id, newName, catDetailStoreUpdater{conn})
  c.save(cds)
  return
}

func (c *catDetailCache) save(cds categories.CatDetailStore) {
  c.mutex.Lock()
  defer c.mutex.Unlock()
  c.data = cds
  c.valid = true
}

func (c *catDetailCache) load(conn *sqlite.Conn) (
    cds categories.CatDetailStore, err error) {
  if cds, err = catDetails(conn); err != nil {
    return
  }
  c.save(cds)
  return
}

func (c *catDetailCache) getFromCache() (cds categories.CatDetailStore, ok bool) {
  c.mutex.Lock()
  defer c.mutex.Unlock()
  if !c.valid {
    return
  }
  return c.data, true
}

type Cache struct {
  db *sqlite_db.Db
  c catDetailCache
}

func (c *Cache) AccountAdd(t db.Transaction, name string) (
    cds categories.CatDetailStore, newId int64, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, newId, err = c.c.AccountAdd(conn, name)
    return
  })
  return
}

func (c *Cache) AccountRename(t db.Transaction, id int64, name string) (
    cds categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, err = c.c.AccountRename(conn, id, name)
    return
  })
  return
}

func (c *Cache) AccountRemove(t db.Transaction, id int64) (
    cds categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, err = c.c.AccountRemove(conn, id)
    return
  })
  return
}

func (c *Cache) Add(t db.Transaction, name string) (
    cds categories.CatDetailStore, newId fin.Cat, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, newId, err = c.c.Add(conn, name)
    return
  })
  return
}

func (c *Cache) Get(t db.Transaction) (
    cds categories.CatDetailStore, err error) {
  if t != nil {
    err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
      cds, err = c.c.Get(conn)
      return
    })
    return
  }
  return c.c.DbGet(c.db)
}

func (c *Cache) Invalidate(t db.Transaction) error {
  return sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) error {
    return c.c.Invalidate(conn)
  })
}

func (c *Cache) Remove(t db.Transaction, id fin.Cat) (
    cds categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, err = c.c.Remove(conn, id)
    return
  })
  return
}

func (c *Cache) Purge(t db.Transaction, cats fin.CatSet) error {
  return sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) error {
    return c.c.Purge(conn, cats)
  })
}

func (c *Cache) Rename(t db.Transaction, id fin.Cat, newName string) (
    cds categories.CatDetailStore, err error) {
  err = sqlite_db.ToDoer(c.db, t).Do(func(conn *sqlite.Conn) (err error) {
    cds, err = c.c.Rename(conn, id, newName)
    return
  })
  return
}
