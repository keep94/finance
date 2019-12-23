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
  "github.com/keep94/goconsume"
  "github.com/keep94/gosqlite/sqlite"
  "os"
  "strings"
)

const (
  kDbFlag = "db"
  kNameFlag = "name"
  kPermFlag = "perm"
)

func main() {
  if len(os.Args) == 1 {
    fmt.Println("usage: ledgeruser <command> [<args>]")
    fmt.Println("  list   list the users")
    fmt.Println("  add    add a user")
    fmt.Println("  remove remove user")
    fmt.Println("  update update user")
    return
  }
  switch os.Args[1] {
    case "list":
      if !doList(os.Args[2:]) {
        os.Exit(1)
      }
    case "add":
      if !doAdd(os.Args[2:]) {
        os.Exit(1)
      }
    case "remove":
      if !doRemove(os.Args[2:]) {
        os.Exit(1)
      }
    case "update":
      if !doUpdate(os.Args[2:]) {
        os.Exit(1)
      }
    default:
      fmt.Printf("%q is not a valid command.\n", os.Args[1])
      os.Exit(2)
  }
}

func checkStrFlag(f *flag.FlagSet, flagName, flagValue string) {
  if flagValue == "" {
    fmt.Fprintf(f.Output(), "Need to specify -%s flag.\n", flagName)
    os.Exit(2)
  }
}

func checkDbAndName(f *flag.FlagSet, db, name string) {
  checkStrFlag(f, kDbFlag, db)
  checkStrFlag(f, kNameFlag, name)
}

func checkPermission(f *flag.FlagSet, p string) fin.Permission {
  result, ok := getPermission(p)
  if !ok {
    fmt.Fprintf(f.Output(), "Need to specify read, all, or none for the -%s flag.\n", kPermFlag)
    os.Exit(2)
  }
  return result
}

func addDbFlag(f *flag.FlagSet) *string {
  return f.String(kDbFlag, "", "Path to database file")
}

func addNameFlag(f *flag.FlagSet) *string {
  return f.String(kNameFlag, "", "User name")
}

func addPasswordFlag(f *flag.FlagSet, defaultValue string) *string {
  return f.String("password", defaultValue, "User password")
}

func addPermFlag(f *flag.FlagSet, defaultValue string) *string {
  return f.String(kPermFlag, defaultValue, "User permission: read | all | none")
}

func doList(args []string) bool {
  flags := flag.NewFlagSet("list", flag.ExitOnError)
  dbPath := addDbFlag(flags)
  flags.Parse(args)
  checkStrFlag(flags, kDbFlag, *dbPath)
  dbase := openDb(*dbPath)
  defer dbase.Close()
  store, _, ok := initDb(dbase)
  if !ok {
    return false
  }
  return listUsers(store)
}

func doAdd(args []string) bool {
  flags := flag.NewFlagSet("add", flag.ExitOnError)
  dbPath := addDbFlag(flags)
  name := addNameFlag(flags)
  password := addPasswordFlag(flags, "password")
  permission := addPermFlag(flags, "read")
  flags.Parse(args)
  checkDbAndName(flags, *dbPath, *name)
  perm := checkPermission(flags, *permission)

  dbase := openDb(*dbPath)
  defer dbase.Close()
  store, _, ok := initDb(dbase)
  if !ok {
    return false
  }
  user := fin.User{
      Name: *name,
      Password: passwords.New(*password),
      Permission: perm}
  if err := store.AddUser(nil, &user); err != nil {
    fmt.Printf("An error happened adding user - %v\n", err)
    return false
  }
  return true
}

func doUpdate(args []string) bool {
  flags := flag.NewFlagSet("update", flag.ExitOnError)
  dbPath := addDbFlag(flags)
  name := addNameFlag(flags)
  password := addPasswordFlag(flags, "")
  permission := addPermFlag(flags, "")
  flags.Parse(args)
  checkDbAndName(flags, *dbPath, *name)
  perm := fin.NonePermission
  if *permission != "" {
    perm = checkPermission(flags, *permission)
  }
  dbase := openDb(*dbPath)
  defer dbase.Close()
  store, doer, ok := initDb(dbase)
  if !ok {
    return false
  }
  err := doer.Do(func(t db.Transaction) (err error) {
      var user fin.User
      if err = store.UserByName(t, *name, &user); err != nil {
        return
      }
      if *password != "" {
        user.Password = passwords.New(*password)
      }
      if *permission != "" {
        user.Permission = perm
      }
      return store.UpdateUser(t, &user)
  })
  if err == findb.NoSuchId {
    fmt.Printf("No such user - %s\n", *name)
    return false
  } else if err != nil {
    fmt.Printf("An error happened updating user - %v\n", err)
    return false
  }
  return true
}

func doRemove(args []string) bool {
  flags := flag.NewFlagSet("remove", flag.ExitOnError)
  dbPath := addDbFlag(flags)
  name := addNameFlag(flags)
  flags.Parse(args)
  checkDbAndName(flags, *dbPath, *name)
  dbase := openDb(*dbPath)
  defer dbase.Close()
  store, _, ok := initDb(dbase)
  if !ok {
    return false
  }
  if err := store.RemoveUserByName(nil, *name); err != nil {
    fmt.Printf("An error happened removing user - %v\n", err)
    return false
  }
  return true
}

func openDb(dbPath string) *sqlite_db.Db {
  conn, err := sqlite.Open(dbPath)
  if err != nil {
    fmt.Printf("Unable to open database - %s\n", dbPath)
    os.Exit(1)
  }
  return sqlite_db.New(conn)
}

func initDb(dbase *sqlite_db.Db) (
    store for_sqlite.Store, doer db.Doer, ok bool) {
  err := dbase.Do(func(conn *sqlite.Conn) error {
    return sqlite_setup.SetUpTables(conn)
  })
  if err != nil {
    fmt.Printf("Unable to create tables - %v\n", err)
    return
  }
  return for_sqlite.New(dbase), sqlite_db.NewDoer(dbase), true
}

func getPermission(perm string) (fin.Permission, bool) {
  permStr := strings.ToLower(perm)
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

func listUsers(store findb.UsersRunner) bool {
  var users []*fin.User
  err := store.Users(nil, goconsume.AppendPtrsTo(&users))
  if err != nil {
    fmt.Printf("An error happened listing users - %v\n", err)
    return false
  }
  for _, u := range users {
    lastLoginStr := "--"
    if !u.LastLogin.IsZero() {
      lastLoginStr = u.LastLogin.Local().Format("Mon 01/02/2006 15:04")
    }
    fmt.Printf(
        "%-12s %-20s %s\n",
        u.Name,
        lastLoginStr,
        u.Permission)
  }
  return true
}
