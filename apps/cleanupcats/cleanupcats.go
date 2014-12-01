package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "flag"
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/finance/fin"
  for_csqlite "github.com/keep94/finance/fin/categories/categoriesdb/for_sqlite"
  "github.com/keep94/finance/fin/consumers"
  "github.com/keep94/finance/fin/findb/for_sqlite"
  "github.com/keep94/gofunctional3/functional"
)

var (
  fDb string
  fDryRun bool
)

func main() {
  flag.Parse()
  if fDb == "" {
    fmt.Println("Need to specify at least -db flag.")
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
  store := for_sqlite.New(dbase)
  cache := for_csqlite.New(dbase)
  doer := sqlite_db.NewDoer(dbase)
  err = doer.Do(func(t db.Transaction) error {
    totals := make(fin.CatTotals)
    allAccounts := make(fin.AccountSet)
    err := store.Entries(
        t,
        nil,
        functional.CompositeConsumer(
            new(fin.Entry),
            nil,
            consumers.FromCatPaymentAggregator(totals),
            consumers.FromCatPaymentAggregator(allAccounts)))
    if err != nil {
      return err
    }
    cds, err := cache.Get(t)
    if err != nil {
      return err
    }
    cats := cds.PurgeableCats(totals)
    accounts := cds.PurgeableAccounts(allAccounts)
    if cats == nil && accounts == nil {
      fmt.Println("No unused inactive categories.")
      return nil
    }
    if fDryRun {
      fmt.Println("Would purge the following categories: ")
    } else {
      fmt.Println("Purging the following categories: ")
    }
    fmt.Println()
    for _, detail := range cds.DetailsByIds(cats) {
      fmt.Println(detail.FullName())
    }
    for id := range accounts {
      if accounts[id] {
        catId := fin.Cat{Id: id, Type: fin.AccountCat}
        fmt.Println(cds.DetailById(catId).FullName())
      }
    }
    if !fDryRun {
      for id := range accounts {
        if accounts[id] {
          if err = store.RemoveAccount(t, id); err != nil {
            return err
          }
        }
      }
      // We do this last because it invalidates the cache.
      if err = cache.Purge(t, cats); err != nil {
        return err
      }
    }
    return nil
  })
  if err != nil {
    fmt.Printf("Got database error: %v\n", err)
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to database file")
  flag.BoolVar(&fDryRun, "dryrun", false, "Dry run only")
}
