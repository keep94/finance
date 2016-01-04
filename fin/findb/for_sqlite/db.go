// Package for_sqlite stores types in fin package in a sqlite database.
package for_sqlite

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/appcommon/passwords"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "github.com/keep94/gosqlite/sqlite"
  "strconv"
  "strings"
  "time"
)

const (
    kSQLEntryById = "select id, date, name, desc, check_no, cats, payment, reviewed from entries where id = ?"
    kSQLEntriesPrefix = "select id, date, name, desc, check_no, cats, payment, reviewed from entries"
    kSQLEntries = "select id, date, name, desc, check_no, cats, payment, reviewed from entries order by date desc, id desc"
    kSQLEntryOrderBy = " order by date desc, id desc"
    kSQLInsertEntry = "insert into entries (date, name, desc, check_no, cats, payment, reviewed) values (?, ?, ?, ?, ?, ?, ?)"
    kSQLUpdateEntry = "update entries set date = ?, name = ?, desc = ?, check_no = ?, cats = ?, payment = ?, reviewed = ? where id = ?"
    kSQLDeleteEntryById = "delete from entries where id = ?"
    kSQLRecurringEntryById = "select id, date, name, desc, check_no, cats, payment, reviewed, count, unit, num_left, day_of_month from recurring_entries where id = ?"
    kSQLRecurringEntries = "select id, date, name, desc, check_no, cats, payment, reviewed, count, unit, num_left, day_of_month from recurring_entries order by date, id"
    kSQLInsertRecurringEntry = "insert into recurring_entries (date, name, desc, check_no, cats, payment, reviewed, count, unit, num_left, day_of_month) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
    kSQLUpdateRecurringEntry = "update recurring_entries set date = ?, name = ?, desc = ?, check_no = ?, cats = ?, payment = ?, reviewed = ?, count = ?, unit = ?, num_left = ?, day_of_month = ? where id = ?"
    kSQLDeleteRecurringEntryById = "delete from recurring_entries where id = ?"
    kSQLAccountById = "select id, name, is_active, balance, reconciled, b_count, r_count, import_sd from accounts where id = ?"
    kSQLAccounts = "select id, name, is_active, balance, reconciled, b_count, r_count, import_sd from accounts"
    kSQLActiveAccounts = "select id, name, is_active, balance, reconciled, b_count, r_count, import_sd from accounts where is_active = 1 order by name"
    kSQLInsertAccount = "insert into accounts (name, is_active, balance, reconciled, b_count, r_count, import_sd) values (?, ?, ?, ?, ?, ?, ?)"
    kSQLUpdateAccountImportSD = "update accounts set import_sd = ? where id = ?"
    kSQLUpdateAccount = "update accounts set name = ?, is_active = ?, balance = ?, reconciled = ?, b_count = ?, r_count = ?, import_sd = ? where id = ?"
    kSQLRemoveAccount = "delete from accounts where id = ?"
    kSQLUserById = "select id, name, go_password, permission from users where id = ?"
    kSQLUsers = "select id, name, go_password, permission from users order by name"
    kSQLRemoveUserByName = "delete from users where name = ?"
    kSQLUserByName = "select id, name, go_password, permission from users where name = ?"
    kSQLInsertUser = "insert into users (name, go_password, permission) values (?, ?, ?)"
    kSQLUpdateUser = "update users set name = ?, go_password = ?, permission = ? where id = ?"
)

func New(db *sqlite_db.Db) Store {
  return Store{db}
}

func ConnNew(conn *sqlite.Conn) Store {
  return Store{sqlite_db.NewSqliteDoer(conn)}
}

func ReadOnlyWrapper(store Store) ReadOnlyStore {
  return ReadOnlyStore{store: store}
}

func entries(conn *sqlite.Conn, options *findb.EntryListOptions, consumer functional.Consumer) error {
  var sql string
  if options != nil {
    where_clauses := make([]string, 3)
    where_clause_count := 0
    if options.Start != nil {
      where_clauses[where_clause_count] = "date >= ?"
      where_clause_count++
    }
    if options.End != nil {
      where_clauses[where_clause_count] = "date < ?"
      where_clause_count++
    }
    if options.Unreviewed {
      where_clauses[where_clause_count] = "reviewed != 1"
      where_clause_count++
    }
    if where_clause_count > 0 {
      sql = kSQLEntriesPrefix + " where " + strings.Join(where_clauses[:where_clause_count], " and ") + kSQLEntryOrderBy
    } else {
      sql = kSQLEntriesPrefix + kSQLEntryOrderBy
    }
  } else {
    sql = kSQLEntriesPrefix + kSQLEntryOrderBy
  }
  stmt, err := conn.Prepare(sql)
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  if options != nil && (options.Start != nil || options.End != nil) {
    sql_params := make([]interface{}, 2)
    sql_param_count := 0
    if options.Start != nil {
      sql_params[sql_param_count] = sqlite_db.DateToString(*options.Start)
      sql_param_count++
    }
    if options.End != nil {
      sql_params[sql_param_count] = sqlite_db.DateToString(*options.End)
      sql_param_count++
    }
    err = stmt.Exec(sql_params[:sql_param_count]...)
    if err != nil {
      return err
    }
  }
  return consumer.Consume(sqlite_db.ReadRows(&rawEntry{}, stmt))
}

func entryById(conn *sqlite.Conn, id int64, entry interface{}) error {
  stmt, err := conn.Prepare(kSQLEntryById)
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  return _entryById(stmt, &rawEntry{}, id, entry)
}

func _entryById(stmt *sqlite.Stmt, r *rawEntry, id int64, entry interface{}) error {
  if err := stmt.Exec(id); err != nil {
    return err
  }
  return consume.FirstOnly(sqlite_db.ReadRows(r, stmt), findb.NoSuchId, entry)
}

func entriesByAccountId(conn *sqlite.Conn, acctId int64, account *fin.Account, consumer functional.Consumer) error {
  if account == nil {
    account = &fin.Account{}
  }
  if err := accountById(conn, acctId, account); err != nil {
    return err
  }
  stmt, err := conn.Prepare(kSQLEntries)
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  stream := sqlite_db.ReadRows(&rawEntry{}, stmt)
  stream = functional.Slice(
      functional.Filter(
          functional.NewFilterer(func(ptr interface{}) error {
            p := ptr.(*fin.Entry)
            if !p.WithPayment(acctId) {
              return functional.Skipped
            }
            return nil
          }), stream), 0, account.Count)
  stream = &fin.AddBalanceStream{Balance: account.Balance, Stream: stream}
  return consumer.Consume(stream)
}

func unreconciledEntries(conn *sqlite.Conn, acctId int64, account *fin.Account, consumer functional.Consumer) error {
  if account == nil {
    account = &fin.Account{}
  }
  if err := accountById(conn, acctId, account); err != nil {
    return err
  }
  stmt, err := conn.Prepare(kSQLEntries)
  if err != nil {
    return err
  }
  defer stmt.Finalize()
  stream := sqlite_db.ReadRows(&rawEntry{}, stmt)
  stream = functional.Slice(
      functional.Filter(
          functional.NewFilterer(func(ptr interface{}) error {
            p := ptr.(*fin.Entry)
            if !p.WithPayment(acctId) || p.Reconciled() {
              return functional.Skipped
            }
            return nil
          }), stream), 0, account.Count - account.RCount)
  return consumer.Consume(stream)
}

func doEntryChanges(conn *sqlite.Conn, changes *findb.EntryChanges) error {
  row := &rawEntry{}
  var err error
  var deltas fin.AccountDeltas = make(map[int64]*fin.AccountDelta)
  var lastRowIdStmt, getStmt, addStmt, deleteStmt, updateStmt *sqlite.Stmt
  if len(changes.Updates) > 0 || len(changes.Deletes) > 0 {
    getStmt, err = conn.Prepare(kSQLEntryById)
    if err != nil {
      return err
    }
    defer getStmt.Finalize()
  }
  if len(changes.Adds) > 0 {
    addStmt, err = conn.Prepare(kSQLInsertEntry)
    if err != nil {
      return err
    }
    defer addStmt.Finalize()
    lastRowIdStmt, err = conn.Prepare(sqlite_db.LastRowIdSQL)
    if err != nil {
      return err
    }
    defer lastRowIdStmt.Finalize()
  }
  if len(changes.Deletes) > 0 {
    deleteStmt, err = conn.Prepare(kSQLDeleteEntryById)
    if err != nil {
      return err
    }
    defer deleteStmt.Finalize()
  }
  if len(changes.Updates) > 0 {
    updateStmt, err = conn.Prepare(kSQLUpdateEntry)
    if err != nil {
      return err
    }
    defer updateStmt.Finalize()
  }
  var entry fin.Entry
  for _, id := range changes.Deletes {
    err = _entryById(getStmt, row, id, &entry)
    if err == findb.NoSuchId {
      continue
    }
    if err != nil {
      return err
    }
    deltas.Exclude(&entry.CatPayment)
    err = deleteStmt.Exec(id)
    if err != nil {
      return err
    }
    deleteStmt.Next()
  }
  for id, update := range changes.Updates {
    var entryWithEtag fin.EntryWithEtag
    var entryPtr *fin.Entry
    if changes.Etags != nil {
      entryPtr = &entryWithEtag.Entry
      err = _entryById(getStmt, row, id, &entryWithEtag)
    } else {
      entryPtr = &entry
      err = _entryById(getStmt, row, id, &entry)
    }
    if err == findb.NoSuchId {
      continue
    }
    if err != nil {
      return err
    }
    concurrent_update_detected := false
    if changes.Etags != nil {
      expected_etag, ok := changes.Etags[id]
      if !ok {
        panic("Etags field present, but does not contain etag for all updated entries.")
      }
      if expected_etag != entryWithEtag.Etag {
        concurrent_update_detected = true
      }
    }
    old_cat_payment := entryPtr.CatPayment
    err = update.Filter(entryPtr)
    if err == functional.Skipped {
      continue
    }
    if err != nil {
      return err
    }
    if concurrent_update_detected {
      return findb.ConcurrentUpdate
    }
    deltas.Exclude(&old_cat_payment)
    deltas.Include(&entryPtr.CatPayment)
    entryPtr.Id = id
    var updateValues []interface{}
    if updateValues, err = sqlite_db.UpdateValues(row, entryPtr); err != nil {
      return err
    }
    err = updateStmt.Exec(updateValues...)
    if err != nil {
      return err
    }
    updateStmt.Next()
  }
  for _, entry := range changes.Adds {
    deltas.Include(&entry.CatPayment)
    err = addEntry(addStmt, lastRowIdStmt, row, entry)
    if err != nil {
      return err
    }
  }
  return recordAccountDeltas(conn, deltas)
}

func accountById(conn *sqlite.Conn, acctId int64, account *fin.Account) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawAccount{},
        findb.NoSuchId,
        account,
        kSQLAccountById,
        acctId)
}

func activeAccounts(conn *sqlite.Conn) (accounts []*fin.Account, err error) {
  stmt, err := conn.Prepare(kSQLActiveAccounts)
  if err != nil {
    return
  }
  defer stmt.Finalize()
  s := sqlite_db.ReadRows(&rawAccount{}, stmt)
  for {
    p := &fin.Account{}
    err = s.Next(p)
    if  err == functional.Done {
      err = nil
      return
    }
    if err != nil {
      return
    }
    accounts = append(accounts, p)
  }
  return
}

func updateAccountImportSD(conn *sqlite.Conn, acctId int64, date time.Time) error {
  return conn.Exec(kSQLUpdateAccountImportSD, sqlite_db.DateToString(date), acctId)
}

func addEntry(stmt, lastRowIdStmt *sqlite.Stmt, r *rawEntry, entry *fin.Entry) error {
  values, err := sqlite_db.InsertValues(r, entry)
  if err != nil {
    return err
  }
  err = stmt.Exec(values...)
  if err != nil {
    return err
  }
  stmt.Next()
  entry.Id, err = sqlite_db.LastRowIdFromStmt(lastRowIdStmt)
  return err
}

func recordAccountDeltas(conn *sqlite.Conn, deltas fin.AccountDeltas) error {
  for id, delta := range deltas {
    err := conn.Exec("update accounts set balance = balance + ?, reconciled = reconciled + ?, b_count = b_count + ?, r_count = r_count + ? where id = ?", delta.Balance, delta.RBalance, delta.Count, delta.RCount, id)
    if err != nil {
      return err
    }
  }
  return nil
}

type rawEntry struct {
  *fin.Entry
  dateStr string
  cat string
  payment string
  status int
}

func (r *rawEntry) Ptrs() []interface{} {
  return []interface{}{&r.Id, &r.dateStr, &r.Name, &r.Desc, &r.CheckNo, &r.cat, &r.payment, &r.status}
}

func (r *rawEntry) Values() []interface{} {
  return []interface{}{r.dateStr, r.Name, r.Desc, r.CheckNo, r.cat, r.payment, r.status, r.Id}
}

func (r *rawEntry) Pair(ptr interface{}) {
  r.Entry = ptr.(*fin.Entry)
}

func (r *rawEntry) Unmarshall() error {
  var err error
  if r.Entry.Date, err = sqlite_db.StringToDate(r.dateStr); err != nil {
    return err
  }
  r.Status = fin.ReviewStatus(r.status)
  return r.Entry.Unmarshall(r, unmarshall)
}

func (r *rawEntry) Marshall() error {
  r.dateStr = sqlite_db.DateToString(r.Date)
  r.status = int(r.Status)
  r.Entry.Marshall(marshall, r)
  return nil
}


type rawRecurringEntry struct {
  *fin.RecurringEntry
  re rawEntry
  unit int
}

func (r *rawRecurringEntry) Ptrs() []interface{} {
  return []interface{}{&r.Id, &r.re.dateStr, &r.Name, &r.Desc, &r.CheckNo, &r.re.cat, &r.re.payment, &r.re.status, &r.Period.Count, &r.unit, &r.NumLeft, &r.Period.DayOfMonth}
}

func (r *rawRecurringEntry) Values() []interface{} {
  return []interface{}{r.re.dateStr, r.Name, r.Desc, r.CheckNo, r.re.cat, r.re.payment, r.re.status, r.Period.Count, r.unit, r.NumLeft, r.Period.DayOfMonth, r.Id}
}

func (r *rawRecurringEntry) Pair(ptr interface{}) {
  p := ptr.(*fin.RecurringEntry)
  r.RecurringEntry = p
  r.re.Pair(&p.Entry)
}

func (r *rawRecurringEntry) Unmarshall() (err error) {
  if err = r.re.Unmarshall(); err != nil {
    return
  }
  var valid bool
  if r.Period.Unit, valid = fin.ToRecurringUnit(r.unit); !valid {
    err = errors.New("Invalid recurring unit found in database.")
  }
  return
}

func (r *rawRecurringEntry) Marshall() (err error) {
  if err = r.re.Marshall(); err != nil {
    return
  }
  r.unit = r.Period.Unit.ToInt()
  return
}

type rawAccount struct {
  *fin.Account
  importSDStr string
} 

func (r *rawAccount) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Name, &r.Active, &r.Balance, &r.RBalance, &r.Count, &r.RCount, &r.importSDStr}
}

func (r *rawAccount) Values() []interface{} {
  return []interface{} {r.Name, r.Active, r.Balance, r.RBalance, r.Count, r.RCount, r.importSDStr, r.Id}
}

func (r *rawAccount) Pair(ptr interface{}) {
  r.Account = ptr.(*fin.Account)
}

func (r *rawAccount) Unmarshall() error {
  r.Account.ImportSD, _ = sqlite_db.StringToDate(r.importSDStr)
  return nil
}
  
func (r *rawAccount) Marshall() error {
  r.importSDStr = sqlite_db.DateToString(r.ImportSD)
  return nil
}

type rawUser struct {
  *fin.User
  rawPassword string
  rawPermission int
} 

func (r *rawUser) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Name, &r.rawPassword, &r.rawPermission}
}

func (r *rawUser) Values() []interface{} {
  return []interface{} {r.Name, r.rawPassword, r.rawPermission, r.Id}
}

func (r *rawUser) Pair(ptr interface{}) {
  r.User = ptr.(*fin.User)
}

func (r *rawUser) Unmarshall() error {
  r.Password = passwords.Password(r.rawPassword)
  // Defaults to fin.NonePermission if the raw permission is not recognized
  r.Permission, _ = fin.ToPermission(r.rawPermission)
  return nil
}

func (r *rawUser) Marshall() error {
  r.rawPassword = string(r.Password)
  r.rawPermission = r.Permission.ToInt()
  return nil
}

func unmarshall(ptr interface{}, cr *[]fin.CatRec, id *int64, reconciled *bool) error {
  p := ptr.(*rawEntry)
  var parts []string
  if p.cat != "" {
    parts = strings.Split(p.cat, "|")
  }
  partLen := len(parts)
  if partLen % 3 != 0 {
    return errors.New(fmt.Sprintf("for_sqlite: Category string invalid: %s", p.cat))
  }
  if partLen != 0 {
    (*cr) = make([]fin.CatRec, partLen / 3)
  } else {
    (*cr) = nil
  }
  for i := range (*cr) {
    a, err := strconv.ParseInt(parts[3 * i + 1], 10, 64)
    if err != nil {
      return err
    }
    r, err := strconv.ParseInt(parts[3 * i + 2], 10, 0)
    if err != nil {
      return err
    }
    if r > 0 {
      (*cr)[i] = fin.CatRec{A: a, R: true}
    } else {
      (*cr)[i] = fin.CatRec{A: a, R: false}
    }
    (*cr)[i].C, err = fin.CatFromString(parts[3 * i])
    if err != nil {
      return err
    }
  }
  parts = strings.SplitN(p.payment, "|", 2)
  partLen = len(parts)
  if partLen < 2 {
    return errors.New(fmt.Sprintf("for_sqlite: Payment string invalid: %s", p.payment))
  }
  r, err := strconv.ParseInt(parts[1], 10, 0)
  if err != nil {
    return err
  }
  if r > 0 {
    *reconciled = true
  } else {
    *reconciled = false
  }
  pc, err := fin.CatFromString(parts[0])
  if err != nil {
    return err
  }
  *id = pc.Id
  return nil
}

func marshall(cr []fin.CatRec, id int64, reconciled bool, ptr interface{}) {
  p := ptr.(*rawEntry)
  catStrs := make([]string, 3 * len(cr))
  for i := range cr {
    catStrs[3 * i] = cr[i].C.ToString()
    catStrs[3 * i + 1] = strconv.FormatInt(cr[i].A, 10)
    if cr[i].R {
      catStrs[3 * i + 2] = "1"
    } else {
      catStrs[3 *i + 2] = "0"
    }
  }
  paymentStrs := make([]string, 2)
  pc := fin.Cat{Id: id, Type: fin.AccountCat}
  paymentStrs[0] = pc.ToString()
  if reconciled {
    paymentStrs[1] = "1"
  } else {
    paymentStrs[1] = "0"
  }
  p.cat = strings.Join(catStrs, "|")
  p.payment = strings.Join(paymentStrs, "|")
}

type Store struct {
  db sqlite_db.Doer
}

func (s Store) AccountById(
    t db.Transaction, acctId int64, account *fin.Account) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return accountById(conn, acctId, account)
  })
}

func (s Store) Accounts(
    t db.Transaction, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(
        conn, &rawAccount{}, consumer, kSQLAccounts)
  })
}

func (s Store) ActiveAccounts(t db.Transaction) (
    accounts []*fin.Account, err error) {
  err = sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) (err error) {
    accounts, err = activeAccounts(conn)
    return
  })
  return
}

func (s Store) AddAccount(t db.Transaction, account *fin.Account) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(
        conn, &rawAccount{}, account, &account.Id, kSQLInsertAccount)
  })
}

func (s Store) DoEntryChanges(
    t db.Transaction, changes *findb.EntryChanges) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return doEntryChanges(conn, changes)
  })
}

func (s Store) Entries(
    t db.Transaction, options *findb.EntryListOptions,
    consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return entries(conn, options, consumer)
  })
}

func (s Store) EntriesByAccountId(
    t db.Transaction, acctId int64, account *fin.Account,
    consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return entriesByAccountId(conn, acctId, account, consumer)
  })
}

func (s Store) EntryById(
    t db.Transaction, id int64, entry *fin.Entry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return entryById(conn, id, entry)
  })
}

func (s Store) EntryByIdWithEtag(
    t db.Transaction, id int64, entry *fin.EntryWithEtag) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return entryById(conn, id, entry)
  })
}

func (s Store) UnreconciledEntries(
    t db.Transaction, acctId int64,
    account *fin.Account, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return unreconciledEntries(conn, acctId, account, consumer)
  })
}

func (s Store) UpdateAccountImportSD(
    t db.Transaction, acctId int64, date time.Time) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return updateAccountImportSD(conn, acctId, date)
  })
}

func (s Store) UpdateAccount(
    t db.Transaction, account *fin.Account) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.UpdateRow(conn, &rawAccount{}, account, kSQLUpdateAccount)
  })
}

func (s Store) RemoveAccount(
    t db.Transaction, id int64) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLRemoveAccount, id)
  })
}

func (s Store) AddUser(t db.Transaction, user *fin.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(conn, &rawUser{}, user, &user.Id, kSQLInsertUser)
  })
}

func (s Store) RemoveUserByName(t db.Transaction, name string) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLRemoveUserByName, name)
  })
}

func (s Store) UpdateUser(t db.Transaction, user *fin.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.UpdateRow(conn, &rawUser{}, user, kSQLUpdateUser)
  })
}

func (s Store) UserById(
    t db.Transaction, id int64, user *fin.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn, &rawUser{}, findb.NoSuchId, user, kSQLUserById, id)
  })
}

func (s Store) UserByName(
    t db.Transaction, name string, user *fin.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn, &rawUser{}, findb.NoSuchId, user, kSQLUserByName, name)
  })
}

func (s Store) Users(
    t db.Transaction, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(conn, &rawUser{}, consumer, kSQLUsers)
  })
}

func (s Store) AddRecurringEntry(
    t db.Transaction, entry *fin.RecurringEntry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(
        conn, &rawRecurringEntry{}, entry, &entry.Id, kSQLInsertRecurringEntry)
  })
}

func (s Store) UpdateRecurringEntry(
    t db.Transaction, entry *fin.RecurringEntry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.UpdateRow(
        conn, &rawRecurringEntry{}, entry, kSQLUpdateRecurringEntry)
  })
}

func (s Store) RecurringEntryById(
      t db.Transaction, id int64, entry *fin.RecurringEntry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawRecurringEntry{},
        findb.NoSuchId,
        entry,
        kSQLRecurringEntryById,
        id)
  })
}

func (s Store) RecurringEntryByIdWithEtag(
      t db.Transaction,
      id int64,
      entry *fin.RecurringEntryWithEtag) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawRecurringEntry{},
        findb.NoSuchId,
        entry,
        kSQLRecurringEntryById,
        id)
  })
}

func (s Store) RecurringEntries(
    t db.Transaction, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(
        conn, &rawRecurringEntry{}, consumer, kSQLRecurringEntries)
  })
}

func (s Store) RemoveRecurringEntryById(t db.Transaction, id int64) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLDeleteRecurringEntryById, id)
  })
}


type ReadOnlyStore struct {
  findb.NoPermissionStore
  store Store
}

func (s ReadOnlyStore) AccountById(
    t db.Transaction, acctId int64, account *fin.Account) error {
  return s.store.AccountById(t, acctId, account)
}

func (s ReadOnlyStore) Accounts(
    t db.Transaction, consumer functional.Consumer) error {
  return s.store.Accounts(t, consumer)
}

func (s ReadOnlyStore) ActiveAccounts(t db.Transaction) (
    accounts []*fin.Account, err error) {
  return s.store.ActiveAccounts(t)
}

func (s ReadOnlyStore) Entries(
    t db.Transaction, options *findb.EntryListOptions,
    consumer functional.Consumer) error {
  return s.store.Entries(t, options, consumer)
}

func (s ReadOnlyStore) EntriesByAccountId(
    t db.Transaction, acctId int64, account *fin.Account,
    consumer functional.Consumer) error {
  return s.store.EntriesByAccountId(t, acctId, account, consumer)
}

func (s ReadOnlyStore) EntryById(
    t db.Transaction, id int64, entry *fin.Entry) error {
  return s.store.EntryById(t, id, entry)
}

func (s ReadOnlyStore) EntryByIdWithEtag(
    t db.Transaction, id int64, entry *fin.EntryWithEtag) error {
  return s.store.EntryByIdWithEtag(t, id, entry)
}

func (s ReadOnlyStore) UnreconciledEntries(
    t db.Transaction, acctId int64,
    account *fin.Account, consumer functional.Consumer) error {
  return s.store.UnreconciledEntries(t, acctId, account, consumer)
}

func (s ReadOnlyStore) UserById(
    t db.Transaction, id int64, user *fin.User) error {
  return s.store.UserById(t, id, user)
}

func (s ReadOnlyStore) UserByName(
    t db.Transaction, name string, user *fin.User) error {
  return s.store.UserByName(t, name, user)
}

func (s ReadOnlyStore) Users(
    t db.Transaction, consumer functional.Consumer) error {
  return s.store.Users(t, consumer)
}

func (s ReadOnlyStore) RecurringEntryById(
      t db.Transaction, id int64, entry *fin.RecurringEntry) error {
  return s.store.RecurringEntryById(t, id, entry)
}

func (s ReadOnlyStore) RecurringEntryByIdWithEtag(
      t db.Transaction, id int64, entry *fin.RecurringEntryWithEtag) error {
  return s.store.RecurringEntryByIdWithEtag(t, id, entry)
}

func (s ReadOnlyStore) RecurringEntries(
    t db.Transaction, consumer functional.Consumer) error {
  return s.store.RecurringEntries(t, consumer)
}
