package common

import (
	"github.com/gorilla/sessions"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/autoimport"
	"github.com/keep94/toolbox/db"
	"testing"
)

func TestSessionBatch(t *testing.T) {
	s := CreateUserSession(&sessions.Session{Values: make(map[interface{}]interface{})})
	batch5 := batchForTesting{5}
	batch7 := batchForTesting{7}
	s.SetBatch(5, batch5)
	s.SetBatch(7, batch7)
	if s.Batch(5) != batch5 {
		t.Error("Expected batch5")
	}
	if s.Batch(7) != batch7 {
		t.Error("Expected batch7")
	}
	if s.Batch(8) != nil {
		t.Error("Expected nil")
	}
	s.SetBatch(8, nil)
	if s.Batch(8) != nil {
		t.Error("Expected nil")
	}
	s.SetBatch(7, nil)
	if s.Batch(7) != nil {
		t.Error("Expected nil")
	}
	if s.Batch(5) != batch5 {
		t.Error("Expected batch5")
	}
}

type batchForTesting struct {
	acctId int64
}

func (b batchForTesting) Entries() []*fin.Entry {
	return nil
}

func (b batchForTesting) SkipProcessed(t db.Transaction) (
	result autoimport.Batch, err error) {
	return
}

func (b batchForTesting) MarkProcessed(t db.Transaction) error {
	return nil
}

func (b batchForTesting) Len() int {
	return 0
}
