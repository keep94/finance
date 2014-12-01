package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "flag"
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/appcommon/passwords"
  "github.com/keep94/finance/fin"
  "github.com/keep94/finance/fin/findb"
  "github.com/keep94/finance/fin/findb/for_sqlite"
  "github.com/keep94/finance/fin/findb/sqlite_setup"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
)

var (
  fDb string
  fName string
  fPassword string
  fAdd bool
  fRemove bool
  fReset bool
  fList bool
)

func main() {
  flag.Parse()
  if fDb == "" {
    fmt.Println("Need to specify -db flag.")
    flag.Usage()
    return
  }
  conn, err := sqlite.Open(fDb)
  if err != nil {
    fmt.Printf("Unable to open database - %s\n", fDb)
    return
  }
  dbase := sqlite_db.New(conn)
  defer dbase.Close()
  err = dbase.Do(func(conn *sqlite.Conn) error {
    return sqlite_setup.SetUpTables(conn)
  })
  if err != nil {
    fmt.Printf("Unable to create tables - %v\n", err)
    return
  }
  store := for_sqlite.New(dbase)
  doer := sqlite_db.NewDoer(dbase)
  switch {
    default:
      fmt.Println("Need to specify one of -add, -remove, -reset or -list")
      return
    case fAdd:
      if !checkName() {
        return
      }
      user := fin.User{Name: fName, Password: passwords.New(fPassword)}
      if err := store.AddUser(nil, &user); err != nil {
        fmt.Printf("An error happened adding user - %v\n", err)
      }
    case fReset:
      if !checkName() {
        return
      }
      err := doer.Do(func(t db.Transaction) (err error) {
        user := fin.User{}
        if err = store.UserByName(t, fName, &user); err != nil {
          return
        }
        user.Password = passwords.New(fPassword)
        return store.UpdateUser(t, &user)
      })
      if err == findb.NoSuchId {
        fmt.Printf("No such user - %s\n", fName)
      } else if err != nil {
        fmt.Printf("An error happened reseting user password - %v\n", err)
      }
    case fRemove:
      if !checkName() {
        return
      }
      if err := store.RemoveUserByName(nil, fName); err != nil {
        fmt.Printf("An error happened removing user - %v\n", err)
      }
    case fList:
      listUsers(store)
  }
}

func checkName() bool {
  if fName == "" {
    fmt.Println("Need to specify the -name flag.")
    flag.Usage()
    return false
  }
  return true
}

func listUsers(store findb.UsersRunner) {
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
    fmt.Printf("An error happened listing users - %v\n", err)
  }
  for _, name := range names {
    fmt.Println(name)
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to database file")
  flag.StringVar(&fName, "name", "", "User name")
  flag.StringVar(&fPassword, "password", "password", "User password")
  flag.BoolVar(&fAdd, "add", false, "Adds user")
  flag.BoolVar(&fRemove, "remove", false, "Removes user")
  flag.BoolVar(&fReset, "reset", false, "Reset user's password to a particular value")
  flag.BoolVar(&fList, "list", false, "List users")
}
