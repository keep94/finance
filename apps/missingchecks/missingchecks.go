package main

import (
  "flag"
  "fmt"
  "log"
  "os"
  "strconv"

  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin"
  csqlite "github.com/keep94/finance/fin/categories/categoriesdb/for_sqlite"
  "github.com/keep94/finance/fin/checks"
  "github.com/keep94/finance/fin/findb/for_sqlite"
  "github.com/keep94/goconsume"
  "github.com/keep94/gosqlite/sqlite"
)

var (
  fDb string
  fAccount string
)

func main() {
  flag.Parse()
  if fDb == "" || fAccount == "" {
    fmt.Println("Need to specify db and account name")
    flag.Usage()
    os.Exit(1)
  }

  // fDb
  conn, err := sqlite.Open(fDb)
  if err != nil {
    log.Fatal(err)
  }
  dbase := sqlite_db.New(conn)
  defer dbase.Close()
  cache := csqlite.New(dbase)
  store := for_sqlite.New(dbase)
  cds, _ := cache.Get(nil)

  accountDetail, ok := cds.AccountDetailByName(fAccount)
  if !ok {
    fmt.Printf("Unknown account: %s\n", fAccount)
    os.Exit(1)
  }
  var account fin.Account
  var checkNos []int
  store.EntriesByAccountId(
      nil,
      accountDetail.Id(),
      &account,
      goconsume.Map(
          goconsume.AppendTo(&checkNos),
          func(srcPtr, destPtr interface{}) bool {
            entryPtr := srcPtr.(*fin.EntryBalance)

            // It can't be a valid check if it is a credit
            if entryPtr.Total() > 0 {
              return false
            }
            checkNo, err := strconv.Atoi(entryPtr.CheckNo)
            if err != nil {
              return false
            }
            *destPtr.(*int) = checkNo
            return true
          },
          (*int)(nil)))
  missing := checks.Missing(checkNos)  
  if missing == nil {
    fmt.Println("No checks found in account.")
    return
  }
  fmt.Printf("First check: %d\n", missing.First)
  fmt.Printf("Last check: %d\n", missing.Last)
  fmt.Println("Missing checks:")
  for _, hole := range missing.Holes {
    if hole.First == hole.Last {
      fmt.Printf("  %d\n", hole.First)
    } else {
      fmt.Printf("  %d-%d\n", hole.First, hole.Last)
    }
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to database file.")
  flag.StringVar(&fAccount, "account", "", "Name of account")
}
