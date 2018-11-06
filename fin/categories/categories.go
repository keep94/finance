// Package categories maintains metadata on spending categories.
package categories

import (
  "errors"
  "fmt"
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/functional"
  "sort"
  "strconv"
  "strings"
)

var (
  MalformedCategoryName = errors.New("categories: Malformed catgory name")
  NoParentCategory = errors.New("categories: No parent category")
  NoSuchCategory = errors.New("categories: No such category")
  NeedExpenseIncomeCategory = errors.New("categories: Need expense/income category")
  InvalidRename = errors.New("categories: Invalid rename")
  InvalidRenameWouldCauseCycle = errors.New("categories: Rename would cause a cycle.")
  DuplicateCategoryName = errors.New("categories: Duplicate category name.")
)

var (
  kEmptyData = newCatDetailStore(newCatIdToDetail())._data
)

// CatDbRow represents a category database row
type CatDbRow struct {
  Id int64
  ParentId int64
  Name string
  Active bool
}

// AccountDetail represents account detail.
type AccountDetail struct {
  ptr *detail
}

// Name returns the account name.
func (a AccountDetail) Name() string {
  return a.ptr.name
}

// Id returns the account Id.
func (a AccountDetail) Id() int64 {
  return a.ptr.id.Id
}

// Active returns true if account is active.
func (a AccountDetail) Active() bool {
  return a.ptr.active
}


// CatDetail represents category detail.
type CatDetail struct {
  ptr *catDetail
}

// FullName returns the full name of the category.
func (c CatDetail) FullName() string {
  return c.ptr.fullName
}

// Id returns the category ID of the category.
func (c CatDetail) Id() fin.Cat {
  return c.ptr.id
}

// Active returns true if category is active.
func (c CatDetail) Active() bool {
  return c.ptr.active
}

type catDetail struct {
  fullName string
  id fin.Cat
  active bool
}

// CategoryAdder adds a new category to the database.
// t is fin.ExpenseCat or fin.IncomeCat. row is the category to be added.
type CategoryAdder interface {
  Add(t fin.CatType, row *CatDbRow) error
}

// CategoryUpdater updates a category in the database
// t is fin.ExpenseCat or fin.IncomeCat. row denotes the updated category.
type CategoryUpdater interface {
  Update(t fin.CatType, row *CatDbRow) error
}

// CategoryRemover removes a category in the database by marking it inactive.
// t is fin.ExpenseCat or fin.IncomeCat. id specifies what is to be deleted.
type CategoryRemover interface {
  Remove(t fin.CatType, id int64) error
}

// AccountAdder adds a new account to the database.
// name is the name of the account.
type AccountAdder interface {
  Add(name string) (newId int64, err error)
}

// AccountUpdater updates a account in the database
type AccountUpdater interface {
  Update(id int64, newName string) error
}

// AccountRemover removes a account in the database by marking it inactive.
type AccountRemover interface {
  Remove(id int64) error
}

// NamedCat represents a category Id and name
type NamedCat struct {
  Id fin.Cat
  Name string
}

// CatDetailStore contains all information on categories.
// The zero value of CatDetailStore contains only the
// trivial categories, "expense" and "income," and no accounts.
type CatDetailStore struct {
  _data *catDetailStore
}

type catDetailStore struct {
  catIdToDetail map[fin.Cat]*detail
  fullNameToDetail map[string]*detail
  fullNameToInactiveDetail map[string]*detail
  accountNameToDetail map[string]*detail
}

type detail struct {
  catDetail
  name string
  parentId int64
  origActive bool
}

// LeafNameById returns the category leaf name by category Id.
func (cds CatDetailStore) LeafNameById(cat fin.Cat) string {
  if cat.Id == 0 {
    return topLevelName(cat.Type)
  }
  d, ok := cds.data().catIdToDetail[cat]
  if !ok {
    return strconv.FormatInt(cat.Id, 10)
  }
  return d.name
}

// DetailById returns details by category Id.
func (cds CatDetailStore) DetailById(cat fin.Cat) CatDetail {
  d, ok := cds.data().catIdToDetail[cat]
  if !ok {
    return CatDetail{&catDetail{
        fullName: fmt.Sprintf("%s:%d", topLevelName(cat.Type), cat.Id),
        id: cat,
        active: false}}
  }
  return CatDetail{&d.catDetail}
}

// AccountDetailById returns account details by acount Id.
func (cds CatDetailStore) AccountDetailById(id int64) AccountDetail {
  cat := fin.Cat{Id: id, Type: fin.AccountCat}
  d, ok := cds.data().catIdToDetail[cat]
  if !ok {
    d = &detail{
        catDetail: catDetail{id: cat},
        name: strconv.FormatInt(id, 10)}
  }
  return AccountDetail{d}
}

// ActiveCatDetails returns all active category details sorted by full name.
// If accounts is true also includes account categories.
func (cds CatDetailStore) ActiveCatDetails(accounts bool) []CatDetail {
  fullNameToDetail := cds.data().fullNameToDetail
  result := make([]CatDetail, len(fullNameToDetail))
  cnt := 0
  for _, v := range fullNameToDetail {
    if accounts || v.id.Type != fin.AccountCat {
      result[cnt] = CatDetail{&v.catDetail}
      cnt++
    }
  }
  result = result[:cnt]
  sort.Sort(catDetails(result))
  return result
}

// PurgeableAccounts returns account ids eligible for purging. An account
// is eligible for purging it it is inactive and not used
// in any entry. accountSet should include the account ids used in all
// entries in the data store. 
func (cds CatDetailStore) PurgeableAccounts(
    accountSet fin.AccountSet) fin.AccountSet {
  var result fin.AccountSet
  for cat, d := range cds.data().catIdToDetail {
    if cat.Type != fin.AccountCat || d.active {
      continue
    }
    if accountSet[cat.Id] {
      continue
    }
    if result == nil {
      result = make(fin.AccountSet)
    }
    result[cat.Id] = true
  }
  return result
}

// PurgeableCats returns category ids eligible for purging. A category
// is eligible for purging it it is inactive and it nor any of its
// children are used in any entry. total should include totals for all entries
// in the data store. PurgeableCats calls RollUp on total internally.
// PurgeableCats will only return expense and income categories.
func (cds CatDetailStore) PurgeableCats(total fin.CatTotals) fin.CatSet {
  total, _ = cds.RollUp(total)
  var result fin.CatSet
  for cat, d := range cds.data().catIdToDetail {
    if cat.Type == fin.AccountCat || d.active {
      continue
    }
    if _, ok := total[cat]; ok {
      continue
    }
    if result == nil {
      result = make(fin.CatSet)
    }
    result[cat] = true
  }
  return result
}

// ActiveAccountDetails returns all active details sorted by name.
func (cds CatDetailStore) ActiveAccountDetails() []AccountDetail {
  accountNameToDetail := cds.data().accountNameToDetail
  result := make([]AccountDetail, len(accountNameToDetail))
  cnt := 0
  for _, v := range accountNameToDetail {
    result[cnt] = AccountDetail{v}
    cnt++
  }
  sort.Sort(accountDetails(result))
  return result
}

// DetailsByIds returns details sorted by full name for selected categories.
func (cds CatDetailStore) DetailsByIds(cats fin.CatSet) []CatDetail {
  result := make([]CatDetail, len(cats))
  idx := 0
  for cat, ok := range cats {
    if ok {
      result[idx] = cds.DetailById(cat)
      idx++
    }
  }
  result = result[:idx]
  sort.Sort(catDetails(result))
  return result
}

// InactiveDetailByFullName returns details for inactive categories
// by category full name that can be made active again.
// If no such category exists or is active, or can't be made active
// because it has some ancestor category that is inactive then exists is
// false.
func (cds CatDetailStore) InactiveDetailByFullName(fullName string) (
    catDetail CatDetail, exists bool) {
  if _, alreadyActive := cds.DetailByFullName(fullName); alreadyActive {
    return
  }
  d, exists := cds.data().fullNameToInactiveDetail[fullName]
  if !exists {
    return
  }
  catDetail = CatDetail{&d.catDetail}
  return
}

// DetailByFullName returns details by category full name. If no such
// category exists or is inactive, then exists is false.
func (cds CatDetailStore) DetailByFullName(fullName string) (catDetail CatDetail, exists bool) {
  d, exists := cds.data().fullNameToDetail[fullName]
  if !exists {
    return
  }
  catDetail = CatDetail{&d.catDetail}
  return
}

// AccountDetailByName returns account detail by account name. If no such
// account exists or is inactive, then exists is false.
func (cds CatDetailStore) AccountDetailByName(name string) (accountDetail AccountDetail, exists bool) {
  d, exists := cds.data().accountNameToDetail[name]
  if !exists {
    return
  }
  accountDetail = AccountDetail{d}
  return
}

// IsChildOf returns true if childCat is a child of parentCat.
func (cds CatDetailStore) IsChildOf(childCat, parentCat fin.Cat) bool {
  return cds.data().isChildOf(childCat, parentCat)
}

// ImmediateParent returns the immediate parent category of given
// category. If given category is already top level, returns the same
// top level category.
func (cds CatDetailStore) ImmediateParent(cat fin.Cat) fin.Cat {
  if cat.IsTop() {
    return cat
  }
  return cds.data().immediateParent(cat)
}

// Filter returns a CatFilter for finding entries under a specific category.
func (cds CatDetailStore) Filter(cat fin.Cat, includeChildren bool) fin.CatFilter {
  if !includeChildren {
    return func(c fin.Cat) bool {
      return cat == c
    }
  }
  if cat.Id == 0 {
    return func(c fin.Cat) bool {
      return cat.Type == c.Type
    }
  }
  data := cds.data()
  return func(c fin.Cat) bool {
    return data.isChildOf(c, cat)
  }
}

// Add adds a new category in the database and returns the updated store.
// Name is the full name of the new category; adder adds the category to
// the database. On error, returns the receiver unchanged.
func (cds CatDetailStore) Add(name string, adder CategoryAdder) (
    updatedStore CatDetailStore, newId fin.Cat, err error) {
  updatedStore = cds
  idx := strings.LastIndex(name, ":")
  if idx == -1 {
    err = MalformedCategoryName
    return
  }
  d, ok := cds.DetailByFullName(name[:idx])
  if !ok {
    err = NoParentCategory
    return
  }
  cat := d.Id()
  if cat.Type == fin.AccountCat {
    err = MalformedCategoryName
    return
  }
  _, ok = cds.DetailByFullName(name)
  if ok {
    err = DuplicateCategoryName
    return
  }
  catDbRow := CatDbRow{ParentId: cat.Id, Name: name[idx + 1:], Active: true}
  if err = adder.Add(cat.Type, &catDbRow); err != nil {
    return
  }
  catIdToDetail := copyRawInfo(cds.data().catIdToDetail)
  newDetail := catDbRowToDetail(cat.Type, &catDbRow)
  catIdToDetail[newDetail.id] = newDetail
  newId = newDetail.id
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// Rename renames a category in the database and returns the updated store.
// id is the id of the category to be renamed; newName is the new name;
// updater does the rename in the database.
// On error, returns the receiver unchanged.
// If id represents an inactive category, it is made active when renamed.
func (cds CatDetailStore) Rename(id fin.Cat, newName string, updater CategoryUpdater) (
      updatedStore CatDetailStore, err error) {
  updatedStore = cds
  if id.Type == fin.AccountCat {
    err = NeedExpenseIncomeCategory
    return
  }
  catIdToDetail := cds.data().catIdToDetail
  if catIdToDetail[id] == nil {
    err = NoSuchCategory
    return
  }
  idx := strings.LastIndex(newName, ":")
  if idx == -1 {
    err = MalformedCategoryName
    return
  }
  d, ok := cds.DetailByFullName(newName[:idx])
  if !ok {
    err = NoParentCategory
    return
  }
  new_parent_cat := d.Id()
  if new_parent_cat.Type != id.Type {
    err = InvalidRename
    return
  }
  if cds.IsChildOf(new_parent_cat, id) {
    err = InvalidRenameWouldCauseCycle
    return
  }
  new_cat, ok := cds.DetailByFullName(newName)
  if ok {
    if new_cat.Id() == id {
      return
    }
    err = DuplicateCategoryName
    return
  }
  catDbRow := CatDbRow{Id: id.Id, ParentId: new_parent_cat.Id, Name: newName[idx + 1:], Active: true}
  if err = updater.Update(new_parent_cat.Type, &catDbRow); err != nil {
    return
  }
  catIdToDetail = copyRawInfo(catIdToDetail)
  newDetail := catDbRowToDetail(new_parent_cat.Type, &catDbRow)
  catIdToDetail[newDetail.id] = newDetail
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// Remove removes a category form database and returns the updated store.
// cat specifies the category to be removed; remover does the actual remove
// in the database.
// On error, returns the receiver unchanged.
func (cds CatDetailStore) Remove(
    cat fin.Cat, remover CategoryRemover) (
    updatedStore CatDetailStore, err error) {
  updatedStore = cds
  if cat.Id == 0 {
    err = NoParentCategory
    return
  }
  catIdToDetail := cds.data().catIdToDetail
  if catIdToDetail[cat] == nil {
    err = NoSuchCategory
    return
  }
  if err = remover.Remove(cat.Type, cat.Id); err != nil {
    return
  }
  catIdToDetail = copyRawInfo(catIdToDetail)
  catIdToDetail[cat].origActive = false
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// AccountAdd adds a new account in the database and returns the updated store.
// Name is the name of the new account; adder adds the account to
// the database. On error, returns the receiver unchanged.
func (cds CatDetailStore) AccountAdd(name string, adder AccountAdder) (
    updatedStore CatDetailStore, newId int64, err error) {
  updatedStore = cds
  if strings.Index(name, ":") != -1 {
    err = MalformedCategoryName
    return
  }
  _, ok := cds.AccountDetailByName(name)
  if ok {
    err = DuplicateCategoryName
    return
  }
  newId, err = adder.Add(name)
  if err != nil {
    return
  }
  catIdToDetail := copyRawInfo(cds.data().catIdToDetail)
  account := fin.Account{
      Id: newId,
      Name: name,
      Active: true,
  }
  newDetail := accountToDetail(&account)
  catIdToDetail[newDetail.id] = newDetail
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// AccountRename renames an account in the database and returns the
// updated store. id is the id of the account to be renamed; newName is
// the new name; updater does the rename in the database.
// On error, returns the receiver unchanged.
// If id represents an inactive account, it is made active when renamed.
func (cds CatDetailStore) AccountRename(
    id int64, name string, updater AccountUpdater) (
    updatedStore CatDetailStore, err error) {
  updatedStore = cds
  if strings.Index(name, ":") != -1 {
    err = MalformedCategoryName
    return
  }
  catIdToDetail := cds.data().catIdToDetail
  catId := fin.Cat{Id: id, Type: fin.AccountCat}
  if catIdToDetail[catId] == nil {
    err = NoSuchCategory
    return
  }
  new_account, ok := cds.AccountDetailByName(name)
  if ok {
    if new_account.Id() == id {
      return
    }
    err = DuplicateCategoryName
    return
  }
  if err = updater.Update(id, name); err != nil {
    return
  }
  catIdToDetail = copyRawInfo(catIdToDetail)
  catIdToDetail[catId].name = name
  catIdToDetail[catId].origActive = true
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// AccountRemove removes a account from database and returns the updated
// store. id specifies the account to be removed; remover does the actual
// remove in the database.
// On error, returns the receiver unchanged.
func (cds CatDetailStore) AccountRemove(
    id int64, remover AccountRemover) (
    updatedStore CatDetailStore, err error) {
  updatedStore = cds
  catIdToDetail := cds.data().catIdToDetail
  catId := fin.Cat{Id: id, Type: fin.AccountCat}
  if catIdToDetail[catId] == nil {
    err = NoSuchCategory
    return
  }
  if err = remover.Remove(id); err != nil {
    return
  }
  catIdToDetail = copyRawInfo(catIdToDetail)
  catIdToDetail[catId].origActive = false
  updatedStore = newCatDetailStore(catIdToDetail)
  return
}

// SortedCatRecs sorts catrecs by category full name and returns catrecs.
func (cds CatDetailStore) SortedCatRecs(catrecs []fin.ROCatRec) []fin.ROCatRec {
  count := len(catrecs)
  cdacr := catDetailsAndCatRecs{
      catDetails: make(catDetails, count),
      catrecs: catrecs}
  for i := 0; i < count; i++ {
    cdacr.catDetails[i] = cds.DetailById(catrecs[i].Id())
  }
  sort.Sort(&cdacr)
  return catrecs
}

// RollUp rolls totals from child categories into parent categories.
// The children return value contains the immediate child categories that were
// rolled into each parent category.
func (cds CatDetailStore) RollUp(totals fin.CatTotals) (rolledUp fin.CatTotals, children map[fin.Cat]fin.CatSet) {
  rolledUp = make(fin.CatTotals, len(totals))
  children = make(map[fin.Cat]fin.CatSet)
  data := cds.data()
  for k, v := range totals {
    for !k.IsTop() {
      rolledUp[k] = rolledUp[k] + v
      parent := data.immediateParent(k)
      childs := children[parent]
      if childs == nil {
        childs = make(fin.CatSet)
        children[parent] = childs
      }
      childs[k] = true
      k = parent
    }
    rolledUp[k] = rolledUp[k] + v
  }
  return
}

// Ancestors returns all the ancestor categories of cat.
// The first item in returned slice is always one of the top level
// categories; the last item is always cat. The items in between are
// the other ancestor categories of cat with the most distant
// ones coming first. The name in each returned category is that category's
// leaf name.
func Ancestors(cds CatDetailStore, cat fin.Cat) []NamedCat {
  var result []NamedCat
  name := cds.LeafNameById(cat)
  result = append(result, NamedCat{Id: cat, Name: name})
  for !cat.IsTop() {
    cat = cds.ImmediateParent(cat)
    name := cds.LeafNameById(cat)
    result = append(result, NamedCat{Id: cat, Name: name})
  }
  // reverse slice
  left := 0
  right := len(result) - 1
  for left < right {
    result[left], result[right] = result[right], result[left]
    left++
    right--
  }
  return result
}

func (cds catDetailStore) isChildOf(childCat, parentCat fin.Cat) bool {
  if childCat.Type != parentCat.Type {
    return false
  }
  for !childCat.IsTop() {
    if childCat == parentCat {
      return true
    }
    childCat = cds.immediateParent(childCat)
  }
  return childCat == parentCat
}

func (cds catDetailStore) immediateParent(cat fin.Cat) fin.Cat {
  detail, ok := cds.catIdToDetail[cat]
  if !ok {
    return fin.Cat{Type: cat.Type}
  }
  return fin.Cat{Type: cat.Type, Id: detail.parentId}
}

func (cds CatDetailStore) data() *catDetailStore {
  if cds._data == nil {
    return kEmptyData
  }
  return cds._data
}

// CatDetailsStoreBuilder is used to initialize CatDetailStore values.
type CatDetailStoreBuilder struct {
  catIdToDetail map[fin.Cat]*detail
  noCopy *CatDetailStoreBuilder
}

// Build returns a new CatDetailStore and resets this builder.
func (cdsb *CatDetailStoreBuilder) Build() CatDetailStore {
  cdsb.initialize()
  catIdToDetail := cdsb.catIdToDetail
  cdsb.catIdToDetail = nil
  return newCatDetailStore(catIdToDetail)
}

// AddAccount adds an account.
func (cdsb *CatDetailStoreBuilder) AddAccount(
    account *fin.Account) *CatDetailStoreBuilder {
  cdsb.initialize()
  d := accountToDetail(account)
  cdsb.catIdToDetail[d.id] = d
  return cdsb
}

// AddCatDbRow adds expense or income category depending on t.
// t is either fin.ExpenseCat or fin.IncomeCat.
func (cdsb *CatDetailStoreBuilder) AddCatDbRow(
    t fin.CatType, row *CatDbRow) *CatDetailStoreBuilder {
  if t == fin.AccountCat {
    panic("t must be either fin.ExpenseCat or fin.IncomeCat.")
  }
  cdsb.initialize()
  d := catDbRowToDetail(t, row)
  cdsb.catIdToDetail[d.id] = d
  return cdsb
}

func (cdsb *CatDetailStoreBuilder) initialize() {
  if cdsb.catIdToDetail == nil {
    cdsb.catIdToDetail = newCatIdToDetail()
    cdsb.noCopy = cdsb
  } else if (cdsb.noCopy != cdsb) {
    panic("CatDetailStoreBuilder values cannot be copied.")
  }
}

// AccountDetailConsumer populates a CatDetailStoreBuilder value with accounts.
type AccountDetailConsumer struct {
  // Builder is what is being populated
  Builder *CatDetailStoreBuilder
}

// Consume does the populating. s is a Stream of Account values.
func (c *AccountDetailConsumer) Consume(s functional.Stream) (err error) {
  var account fin.Account
  for err = s.Next(&account); err == nil; err = s.Next(&account)  {
    c.Builder.AddAccount(&account)
  }
  if err == functional.Done {
    err = nil
  }
  return
}

// CatDetailConsumer populates a CatDetailStoreBuilder value with CatDbRow values.
type CatDetailConsumer struct {
  // Builder is what is being populated
  Builder *CatDetailStoreBuilder
  // Type is either fin.ExpenseCat or fin.IncomeCat.
  Type fin.CatType
}

// Consume does the populating. s is a Stream of CatDbRow values.
func (c *CatDetailConsumer) Consume(s functional.Stream) (err error) {
  var row CatDbRow
  for err = s.Next(&row); err == nil; err = s.Next(&row)  {
    c.Builder.AddCatDbRow(c.Type, &row)
  }
  if err == functional.Done {
    err = nil
  }
  return
}

type catDetails []CatDetail

func (s catDetails) Len() int {
  return len(s)
}

func (s catDetails) Swap(i, j int) {
  s[i], s[j] = s[j], s[i]
}

func (s catDetails) Less(i, j int) bool {
  if s[i].Id().Type < s[j].Id().Type {
    return true
  }
  if s[i].Id().Type > s[j].Id().Type {
    return false
  }
  return s[i].FullName() < s[j].FullName()
}

type accountDetails []AccountDetail

func (s accountDetails) Len() int {
  return len(s)
}

func (s accountDetails) Swap(i, j int) {
  s[i], s[j] = s[j], s[i]
}

func (s accountDetails) Less(i, j int) bool {
  return s[i].Name() < s[j].Name()
}

type catDetailsAndCatRecs struct {
  catDetails
  catrecs []fin.ROCatRec
}

func (s *catDetailsAndCatRecs) Swap(i, j int) {
  s.catDetails.Swap(i, j)
  s.catrecs[i], s.catrecs[j] = s.catrecs[j], s.catrecs[i]
}

func storeFullNameToInactiveDetail(
    fullNameToInactiveDetail map[string]*detail, d *detail) {
  oldVal, ok := fullNameToInactiveDetail[d.fullName]
  if !ok || d.id.Id > oldVal.id.Id {
    fullNameToInactiveDetail[d.fullName] = d
  }
}

func newCatDetailStore(catIdToDetail map[fin.Cat]*detail) CatDetailStore {
  fullNameToDetail := make(map[string]*detail)
  fullNameToInactiveDetail := make(map[string]*detail)
  accountNameToDetail := make(map[string]*detail)
  for _, v := range catIdToDetail {
    computeFullNameAndActive(v, catIdToDetail)
    if v.active {
      fullNameToDetail[v.fullName] = v
      if v.id.Type == fin.AccountCat {
        accountNameToDetail[v.name] = v
      }
    } else {
       if v.id.Type == fin.AccountCat {
         storeFullNameToInactiveDetail(fullNameToInactiveDetail, v)
       } else {
         // it is inactive
         parentCat := fin.Cat{Type: v.id.Type, Id: v.parentId}
         parent := catIdToDetail[parentCat]
         if parent != nil && parent.active {
           storeFullNameToInactiveDetail(fullNameToInactiveDetail, v)
         }
       }
    }
  }
  return CatDetailStore{&catDetailStore{
      catIdToDetail: catIdToDetail,
      fullNameToDetail: fullNameToDetail,
      fullNameToInactiveDetail: fullNameToInactiveDetail,
      accountNameToDetail: accountNameToDetail}}
}

func newCatIdToDetail() map[fin.Cat]*detail {
  expenseId := fin.Cat{Id: 0, Type: fin.ExpenseCat}
  incomeId := fin.Cat{Id: 0, Type: fin.IncomeCat}
  return map[fin.Cat]*detail {
      expenseId: &detail{
          catDetail: catDetail{id: expenseId},
          origActive: true},
      incomeId: &detail{
          catDetail: catDetail{id: incomeId},
          origActive: true}}
}

func computeFullNameAndActive(
    d *detail, catIdToDetail map[fin.Cat]*detail) {
  if d.fullName != "" {
    return
  }
  if d.id.Type == fin.AccountCat {
    d.fullName = fmt.Sprintf("%s:%s", topLevelName(fin.AccountCat), d.name)
    d.active = d.origActive
    return
  }
  if d.id.Id == 0 {
    d.fullName = topLevelName(d.id.Type)
    d.active = d.origActive
    return
  }
  parent_d := catIdToDetail[fin.Cat{Type: d.id.Type, Id: d.parentId}]
  if parent_d != nil {
    computeFullNameAndActive(parent_d, catIdToDetail)
    d.fullName = fmt.Sprintf(
        "%s:%s", parent_d.fullName, d.name)
    d.active = d.origActive && parent_d.active
    return
  }
  d.fullName = fmt.Sprintf("%s:%d:%s", topLevelName(d.id.Type), d.parentId, d.name)
  d.active = false
}

func copyRawInfo(catIdToDetail map[fin.Cat]*detail) map[fin.Cat]*detail {
  result := make(map[fin.Cat]*detail, len(catIdToDetail))
  for k, v := range catIdToDetail {
    newDetail := *v
    newDetail.fullName = ""
    result[k] = &newDetail
  }
  return result
}

func catDbRowToDetail(t fin.CatType, row *CatDbRow) *detail {
  return &detail{
      catDetail: catDetail{
          id: fin.Cat{
              Id: row.Id,
              Type: t}},
      name: row.Name,
      origActive: row.Active,
      parentId: row.ParentId}
}

func accountToDetail(account *fin.Account) *detail {
  return &detail{
      catDetail: catDetail{
          id: fin.Cat{
              Id: account.Id,
              Type: fin.AccountCat}},
      name: account.Name,
      origActive: account.Active}
}

func topLevelName(t fin.CatType) string {
  switch (t) {
    case fin.AccountCat:
      return "account"
    case fin.ExpenseCat:
      return "expense"
    case fin.IncomeCat:
      return "income"
    default:
      panic("No such CatType.")
  }
  return "expense"
}

