package for_sqlite

import (
	"github.com/keep94/finance/fin/autoimport/qfx/qfxdb/fixture"
	"github.com/keep94/finance/fin/findb/sqlite_setup"
	"github.com/keep94/gosqlite/sqlite"
	"github.com/keep94/toolbox/db/sqlite_db"
	"testing"
)

func TestFind(t *testing.T) {
	db := openDb(t)
	defer closeDb(t, db)
	newFixture(db).Find(t)
}

func newFixture(db *sqlite_db.Db) *fixture.Fixture {
	return &fixture.Fixture{Store: New(db), Doer: sqlite_db.NewDoer(db)}
}

func closeDb(t *testing.T, db *sqlite_db.Db) {
	if err := db.Close(); err != nil {
		t.Errorf("Error closing database: %v", err)
	}
}

func openDb(t *testing.T) *sqlite_db.Db {
	conn, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	db := sqlite_db.New(conn)
	err = db.Do(func(conn *sqlite.Conn) error {
		return sqlite_setup.SetUpTables(conn)
	})
	if err != nil {
		t.Fatalf("Error creating tables: %v", err)
	}
	return db
}
