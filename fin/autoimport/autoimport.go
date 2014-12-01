// Package autoimport provides support for importing transactions from banks
package autoimport

import (
  "github.com/keep94/appcommon/db"
  "github.com/keep94/finance/fin"
  "io"
  "time"
)

// Loader wraps the Load method which reads a batch of transactions from a bank.
// Banks provide transactions as QFX files, MS money files, and other formats.
// In a typical application, there will be one Loader instance for each type
// of file supported.
//
// In the Load method, accountId identifies the account for which transactions
// are being loaded; bankAccountId is the account ID from the bank. It should
// be left empty when the file contains transactions for just one account;
// however, some banks will provide a file containing activity from all
// accounts the customer has. In this case, the caller must provide a non-empty
// bankAccountId to get the transactions for the right account.
// r is where the file is read. Only transactions posted on or after startDate
// are read into the batch.  Load returns the read transactions as a Batch
// instance.
type Loader interface {
  Load(
      accountId int64,
      bankAccountId string,
      r io.Reader,
      startDate time.Time) (Batch, error)
}

// Batch represents a group of transactions read from a file using a Loader
// instance. Batch instances are immutable.
type Batch interface {
  // Entries returns the entries in the batch. The returned entries are
  // copies that the caller can safely modify.
  Entries() []*fin.Entry

  // SkippProcessed returns a new Batch like this one that contains only
  // entries that have not already been processed. t is the database
  // transaction; nil means run in a separate transaction.
  SkipProcessed(t db.Transaction) (Batch, error)

  // MarkProcessed marks all the entries in this Batch as processed. t is the
  // database transaction; nil means run in a separate transaction.
  MarkProcessed(t db.Transaction) error

  // Len returns the number of entries in this batch.
  Len() int
}
