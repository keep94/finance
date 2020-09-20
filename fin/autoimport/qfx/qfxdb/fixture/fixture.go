// Package fixture provides test suites to test implementations of the
// qfxdb.Store interface.
package fixture

import (
	"github.com/keep94/finance/fin/autoimport/qfx/qfxdb"
	"github.com/keep94/toolbox/db"
	"reflect"
	"testing"
)

type Fixture struct {
	Store qfxdb.Store
	Doer  db.Doer
}

func (f *Fixture) Find(t *testing.T) {
	setOne := qfxdb.FitIdSet{"FitId1_1": true, "FitId1_2": true}
	setTwo := qfxdb.FitIdSet{"FitId2_1": true, "FitId2_2": true}
	err := f.Doer.Do(func(t db.Transaction) error {
		if err := f.Store.Add(t, 1, setOne); err != nil {
			return err
		}
		return f.Store.Add(t, 2, setTwo)
	})
	if err != nil {
		t.Errorf("Error adding fitIds: %v", err)
		return
	}
	set := qfxdb.FitIdSet{"FitId1_1": true, "FitId1_2": true, "FitId1_3": true}
	inSet, err := f.Store.Find(nil, 1, set)
	if err != nil {
		t.Errorf("Error accessing database: %v", err)
		return
	}
	expected := qfxdb.FitIdSet{"FitId1_1": true, "FitId1_2": true}
	if !reflect.DeepEqual(inSet, expected) {
		t.Errorf("Expected %v, got %v", expected, inSet)
	}
	inSet, err = f.Store.Find(nil, 2, set)
	if err != nil {
		t.Errorf("Error accessing database: %v", err)
		return
	}
	if inSet != nil {
		t.Error("Expected empty set.")
	}
}
