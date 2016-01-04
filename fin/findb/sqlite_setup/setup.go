// Package sqlite_setup sets up a sqlite database for personal finance.
package sqlite_setup

import (
  "github.com/keep94/gosqlite/sqlite"
)

// SetUpTables creates all needed tables in database.
func SetUpTables(conn *sqlite.Conn) error {
  err := conn.Exec("create table if not exists accounts (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, is_active INTEGER, balance INTEGER, reconciled INTEGER, b_count INTEGER, r_count INTEGER, import_sd TEXT)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists entries (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, name TEXT, cats TEXT, payment TEXT, desc TEXT, check_no TEXT, reviewed INTEGER)")
  if err != nil {
    return err
  }
  err = conn.Exec("create index if not exists entries_date_id_idx on entries (date, id)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists recurring_entries (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, name TEXT, cats TEXT, payment TEXT, desc TEXT, check_no TEXT, reviewed INTEGER, count INTEGER, unit INTEGER, num_left INTEGER, day_of_month INTEGER)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists expense_categories (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, is_active INTEGER, parent_id INTEGER)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists income_categories (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, is_active INTEGER, parent_id INTEGER)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, go_password TEXT, permission INTEGER)")
  if err != nil {
    return err
  }
  err = conn.Exec("create unique index if not exists users_name_idx on users (name)")
  if err != nil {
    return err
  }
  err = conn.Exec("create table if not exists qfx_fitids (acct_id INTEGER, fit_id TEXT)")
  if err != nil {
    return err
  }
  err = conn.Exec("create unique index if not exists qfx_fitids_acct_id_fit_id_idx on qfx_fitids (acct_id, fit_id)")
  if err != nil {
    return err
  }
  return nil
}


