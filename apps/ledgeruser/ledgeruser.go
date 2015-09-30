package main

import (
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
  "github.com/keep94/gosqlite/sqlite"
  "strings"
)

var (
  fDb string
  fName string
  fPassword string
  fPermission string
  fAdd bool
  fRemove bool
  fReset bool
  fChangePerm bool
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
      fmt.Println("Need to specify one of -add, -remove, -reset, -list or -changeperm")
      return
    case fAdd:
      perm, ok := getPermission()
      if !checkName() || !checkPermission(ok) {
        return
      }
      user := fin.User{
          Name: fName,
          Password: passwords.New(fPassword),
          Permission: perm}
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
    case fChangePerm:
      perm, ok := getPermission()
      if !checkName() || !checkPermission(ok) {
        return
      }
      err := doer.Do(func(t db.Transaction) (err error) {
        user := fin.User{}
        if err = store.UserByName(t, fName, &user); err != nil {
          return
        }
        user.Permission = perm
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

func checkPermission(ok bool) bool {
  if !ok {
    fmt.Println("Need to specify read, all, or none for the -perm flag.")
    flag.Usage()
  }
  return ok
}

func getPermission() (fin.Permission, bool) {
  permStr := strings.ToLower(fPermission)
  switch permStr {
  case "read":
    return fin.ReadPermission, true
  case "all":
    return fin.AllPermission, true
  case "none":
    return fin.NonePermission, true
  default:
    return fin.NonePermission, false
  }
}

func listUsers(store findb.UsersRunner) {
  var users []*fin.User
  err := store.Users(nil, consume.AppendPtrsTo(&users, nil))
  if err != nil {
    fmt.Printf("An error happened listing users - %v\n", err)
  }
  for _, u := range users {
    fmt.Printf("%s %s\n", u.Name, u.Permission)
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to database file")
  flag.StringVar(&fName, "name", "", "User name")
  flag.StringVar(&fPassword, "password", "password", "User password")
  flag.StringVar(&fPermission, "perm", "all", "User permission: read | all | none")
  flag.BoolVar(&fAdd, "add", false, "Adds user")
  flag.BoolVar(&fRemove, "remove", false, "Removes user")
  flag.BoolVar(&fChangePerm, "changeperm", false, "Changes permission")
  flag.BoolVar(&fReset, "reset", false, "Reset user's password to a particular value")
  flag.BoolVar(&fList, "list", false, "List users")
}
