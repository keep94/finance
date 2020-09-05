// Package categoriesdb contains the persistence layer for the categories package.
package categoriesdb

import (
	"errors"
	"github.com/keep94/appcommon/db"
	"github.com/keep94/finance/fin"
	"github.com/keep94/finance/fin/categories"
)

var (
	NoPermission = errors.New("categoriesdb: Insufficient permission.")
)

type Getter interface {
	// Get retrieves the CatDetailStore instance.
	Get(t db.Transaction) (cds categories.CatDetailStore, err error)
}

type Invalidater interface {
	// Invalidate invalidates this cache forcing the next call to Get
	// to read from the database.
	Invalidate(t db.Transaction) error
}

type Adder interface {
	// Add adds a new category in the database, updates this cache, and returns
	// the updated store. name is the full name of the new category.
	// On error, Add returns the most current version of store available.
	Add(t db.Transaction, name string) (
		cds categories.CatDetailStore, newId fin.Cat, err error)
}

type Remover interface {
	// Remove removes a category by marking it inactive, updates this cache,
	// and returns the updated store. id is the category id to be removed.
	// On error, Remove returns the most current version of store available.
	Remove(t db.Transaction, id fin.Cat) (cds categories.CatDetailStore, err error)
}

type Renamer interface {
	// Rename renames a category in the database, updates this cache, and returns
	// the updated store. id is the category id to be renamed;
	// newName is the new name.
	// On error, Rename returns the most current version of store available.
	// If id represents an inactive category, it is made active when renamed.
	Rename(t db.Transaction, id fin.Cat, newName string) (
		cds categories.CatDetailStore, err error)
}

type AccountAdder interface {
	// AccountAdd adds a new account in the database, updates this cache,
	// and returns the updated store. name is the name of the new account.
	// On error, AccountAdd returns the most current version of store
	// available.
	AccountAdd(t db.Transaction, name string) (
		cds categories.CatDetailStore, newId int64, err error)
}

type AccountRemover interface {
	// AccountRemove removes a account by marking it inactive, updates this
	// cache, and returns the updated store. id is the account id to be
	// removed. On error, AccountRemove returns the most current version of
	// store available.
	AccountRemove(t db.Transaction, id int64) (
		cds categories.CatDetailStore, err error)
}

type AccountRenamer interface {
	// AccountRename renames a account in the database, updates this cache,
	// and returns the updated store. id is the account id to be renamed;
	// newName is the new name.
	// On error, AccountRename returns the most current version of store
	// available. If id represents an inactive account, it is made active
	// when renamed.
	AccountRename(t db.Transaction, id int64, newName string) (
		cds categories.CatDetailStore, err error)
}

type Purger interface {
	// Because Remove does not physically remove categories from the database
	// but only marks them as inactive, Purge is needed to physically remove
	// unused, inactive categories.
	//
	// Purge physically removes the categories in cats from the database and
	// invalidates the cache. If cats contains anything other than expense or
	// income categories e.g accounts, it returns
	// categories.NeedExpenseIncomeCategory.
	//
	// Callers should first call PurgeableCats() on the CatDetailStore instance
	// and pass the result of that to Purge.
	Purge(t db.Transaction, cats fin.CatSet) error
}

// NoPermissionCache always returns NoPermission error.
type NoPermissionCache struct {
}

func (n NoPermissionCache) Get(t db.Transaction) (cds categories.CatDetailStore, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) Invalidate(t db.Transaction) error {
	return NoPermission
}

func (n NoPermissionCache) Add(t db.Transaction, name string) (
	cds categories.CatDetailStore, newId fin.Cat, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) Remove(t db.Transaction, id fin.Cat) (cds categories.CatDetailStore, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) Rename(t db.Transaction, id fin.Cat, newName string) (
	cds categories.CatDetailStore, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) AccountAdd(t db.Transaction, name string) (
	cds categories.CatDetailStore, newId int64, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) AccountRemove(t db.Transaction, id int64) (
	cds categories.CatDetailStore, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) AccountRename(t db.Transaction, id int64, newName string) (
	cds categories.CatDetailStore, err error) {
	err = NoPermission
	return
}

func (n NoPermissionCache) Purge(t db.Transaction, cats fin.CatSet) error {
	return NoPermission
}
