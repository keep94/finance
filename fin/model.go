// Package fin declares basic types used in personal finance.
package fin

import (
  "bytes"
  "encoding/gob"
  "errors"
  "fmt"
  "github.com/keep94/appcommon/passwords"
  "github.com/keep94/gofunctional3/functional"
  "math"
  "sort"
  "strconv"
  "strings"
  "time"
)

// A CatType specifies expense, income, or account category
type CatType int

const (
  ExpenseCat CatType = iota
  IncomeCat
  AccountCat
)

var (
  // Expense is the top level expense category.
  Expense = Cat{Type: ExpenseCat}
  // Income is the top level income category.
  Income = Cat{Type: IncomeCat}
)

var (
  missingColonError = errors.New("fin: String must contain colon.")
)

// A Cat specifies a raw category.
type Cat struct {
  // Id is unique over same type, but not different types.
  Id int64
  Type CatType
}

// NewCat creates a Cat value from a string e.g 0:4. If string cannot be parsed,
// returns the zero value.
func NewCat(s string) Cat {
  result, _ := CatFromString(s)
  return result
}

// CatFromString converts a string e.g 0:4 to a Cat instance
func CatFromString(s string) (cat Cat, err error) {
  ss := strings.SplitN(s, ":", 2)
  if len(ss) < 2 {
    err = missingColonError
    return
  }
  id, err := strconv.ParseInt(ss[1], 10, 64)
  if err != nil {
    return
  }
  i64, err := strconv.ParseInt(ss[0], 10, 0)
  if err != nil {
    return
  }
  cat = Cat{Type: CatType(i64), Id: id}
  return
}

func (c Cat) String() string {
  return c.ToString()
}

// ToString converts this value to a String.
func (c Cat) ToString() string {
  return fmt.Sprintf("%d:%d", c.Type, c.Id)
}

// IsTop returns true if this value represents a top-level category.
func (c Cat) IsTop() bool {
  return c.Id == 0 || c.Type == AccountCat
}

// CatSet represents a set of Cat instances
type CatSet map[Cat]bool

// AddSet adds the Cat instances in rhs to this instance.
// AddSet returns this instance for chaining.
func (c CatSet) AddSet(rhs CatSet) CatSet {
  for cat, ok := range rhs {
    if ok {
      c[cat] = true
    }
  }
  return c
}

// CatRec specifies a category, amount, and reconciled flag.
type CatRec struct {
  // C is the category.
  C Cat
  // A is amount in one cent increments. Positive means expense; negative
  // means income.
  A int64
  // R is reconcile flag, only applicable for categories of AccountCat type.
  R bool
}

// Sets this CatRec to other
func (c *CatRec) Set(other ROCatRec) {
  c.C = other.Id()
  c.A = other.Amount()
  c.R = other.Reconciled()
}

func (c *CatRec) String() string {
  return fmt.Sprintf("%v", *c)
}

// ROCatRec represents a read-only CatRec
type ROCatRec struct {
  ptr *CatRec
}

// NewROCatRec creates a ROCatRec that wraps c.
func NewROCatRec(c *CatRec) ROCatRec {
  return ROCatRec{c}
}

// Id returns the C field.
func (c ROCatRec) Id() Cat {
  return c.ptr.C
}

// Amount returns the A field.
func (c ROCatRec) Amount() int64 {
  return c.ptr.A
}

// Reconciled returns the R field.
func (c ROCatRec) Reconciled() bool {
  return c.ptr.R
}

// Unmarshaller builds components of CatPayment from database columns.
// ptr represents database columns; cr is where the new CatRec slice is
// to be stored; id is where type payment ID is to be stored; r is where
// the reconcile flag for the payment is to be stored.
type Unmarshaller func(ptr interface{}, cr *[]CatRec, id *int64, r *bool) error

// Marshaller marshalls a CatPayment to database columns.
// ptr represents database columns; cr, id, and r are the CatRec slice,
// payment ID, and reconcile flag respectively. Functions of this type
// must not modify cr in-place in any way.
type Marshaller func(cr []CatRec, id int64, r bool, ptr interface{})

// CatFilter filters categories. c is the category. Returns true if c
// should be included or false otherwise.
type CatFilter func(c Cat) bool

// CatPayment specifies category, amount, and payment information for a
// single Entry. CatPayment consist of one payment method and zero or more
// CatRecs. The zero value of CatPayment has no CatRecs, a payment ID of
// zero, and is not reconciled. CatPayment works like a value type with the
// assignment operator, but to test for equality use reflect.DeepEqual.
type CatPayment struct {
  cr []CatRec
  id int64
  r bool
}

// NewCatPayment returns a new CatPayment having payment of paymentId and
// single category of cat with given amount. reconciled indicates
// whether or not new instance should be reconciled under paymentId. Using this
// function is faster than using CatPaymentBuilder, but this function can
// only create CatPayment instances with a single category. Note that Total()
// on returned CatPayment will return -amount.
func NewCatPayment(cat Cat, amount int64, reconciled bool, paymentId int64) CatPayment {
  if cat.Type == AccountCat && cat.Id == paymentId {
    panic("cat cannot match paymentId.")
  }
  return CatPayment{
      cr: []CatRec{{C: cat, A: amount}},
      id: paymentId,
      r: reconciled}
}

// Unmarshall sets this value to what is in the database row.
// ptr is the database row.
func (c *CatPayment) Unmarshall(ptr interface{}, u Unmarshaller) error {
  return u(ptr, &c.cr, &c.id, &c.r)
}

// Marshall writes this value to a database row. ptr is the database row.
func (c *CatPayment) Marshall(m Marshaller, ptr interface{}) {
  m(c.cr, c.id, c.r, ptr)
}

// CatRecCount returns the number of CatRecs
func (c *CatPayment) CatRecCount() int {
  return len(c.cr)
}

func (c *CatPayment) CatRecByIndex(idx int) ROCatRec {
  return NewROCatRec(&c.cr[idx])
}

func (c *CatPayment) CatRecs() []ROCatRec {
  result := make([]ROCatRec, len(c.cr))
  for i := range result {
    result[i] = NewROCatRec(&c.cr[i])
  }
  return result
}

func (c *CatPayment) PaymentId() int64 {
  return c.id
}

// Reconciled returns the reconcile flag.
func (c *CatPayment) Reconciled() bool {
  return c.r
}

// Marks as reconciled. id is a payment Id. Returns true on success or
// false if id does not match payment ID or any of the CatRecs.
func (c *CatPayment) Reconcile(id int64) bool {
  if c.id == id {
    c.r = true
    return true
  }
  pc := Cat{Id: id, Type: AccountCat}
  for i := range c.cr {
    if c.cr[i].C == pc {
      ncr := make([]CatRec, len(c.cr))
      copy(ncr, c.cr)
      ncr[i].R = true
      c.cr = ncr
      return true
    }
  }
  return false
}

// Total returns the total. Negative means debit, positive means credit for
// the payment ID.
func (c *CatPayment) Total() int64 {
  var result int64 = 0
  for i := range c.cr {
    result -= c.cr[i].A
  }
  return result
}

// WithPayment changes this CatPayment so that the payment id matches id.
// If id is not the payment Id and it does not correspond to any payment ids
// in the CatRecs, then WithPayment returns false and leaves this value
// unchanged.
func (c *CatPayment) WithPayment(id int64) bool {
  if c.id == id {
    return true
  }
  pc := Cat{Id: id, Type: AccountCat}
  for i := range c.cr {
    if c.cr[i].C == pc {
      ncr := make([]CatRec, 1)
      ncr[0].C = Cat{Id: c.id, Type: AccountCat}
      ncr[0].A = -c.cr[i].A
      ncr[0].R = c.r
      c.id = c.cr[i].C.Id
      c.r = c.cr[i].R
      c.cr = ncr
      return true
    }
  }
  return false
}

// WithCat changes this value so that it contains only the CatRecs that
// match f. If there are no such CatRecs, WithCat leaves this value unchanged
// and returns false.
func (c *CatPayment) WithCat(f CatFilter) bool {
  idxs := make([]int, len(c.cr))
  idxslen := 0
  for i := range c.cr {
    if f(c.cr[i].C) {
      idxs[idxslen] = i
      idxslen++
    }
  }
  if idxslen == 0 {
    return false
  }
  if idxslen == len(c.cr) {
    return true
  }
  ncr := make([]CatRec, idxslen)
  for i := range ncr {
    ncr[i] = c.cr[idxs[i]]
  }
  c.cr = ncr
  return true
}

// SetSingleCat changes this instance to a single category clearing any
// previous split transaction. SetSingleCat does not change the value Total()
// returns. Returns true on success. If cat represents the paymentId of this
// instance, then SetSingleCat makes no change and returns false.
func (c *CatPayment) SetSingleCat(cat Cat) bool {
  if cat.Type == AccountCat && cat.Id == c.id {
    return false
  }
  *c = NewCatPayment(cat, -c.Total(), c.r, c.id)
  return true
}

// CatPaymentBuilder builds the specifications for a CatPayment value.
type CatPaymentBuilder struct {
  m map[Cat]CatRec
  pc Cat
  r bool
}

// Set sets this CatPaymentBuilder to cp so that calling Build on it will
// return a value equivalent to cp.
func (c *CatPaymentBuilder) Set(cp *CatPayment) *CatPaymentBuilder {
  c.initialize()
  c.ClearCatRecs()
  for idx := range cp.cr {
    c.AddCatRec(&cp.cr[idx])
  }
  c.SetPaymentId(cp.PaymentId())
  c.SetReconciled(cp.Reconciled())
  return c
}

// Clears all the added CatRecs from this builder
func (c *CatPaymentBuilder) ClearCatRecs() *CatPaymentBuilder {
  c.initialize()
  if len(c.m) > 0 {
    c.m = make(map[Cat]CatRec)
  }
  return c
}

// Build returns a new CatPayment instance, and resets this builder.
func (c *CatPaymentBuilder) Build() CatPayment {
  c.initialize()
  var catRecs []CatRec
  if len(c.m) != 0 {
    catRecs = c.newCatRecSlice()
  }
  c.m = nil
  return CatPayment{cr: catRecs, id: c.pc.Id, r: c.r}
}

// AddCatRec Adds a CatRec. It merges CatRecs having the same category.
func (c *CatPaymentBuilder) AddCatRec(cr *CatRec) *CatPaymentBuilder {
  return c.AddROCatRec(NewROCatRec(cr))
}

// AddROCatRec Adds a ROCatRec. It merges CatRecs having the same category.
func (c *CatPaymentBuilder) AddROCatRec(cr ROCatRec) *CatPaymentBuilder {
  c.initialize()
  ocr := c.m[cr.Id()]
  ocr.C = cr.Id()
  ocr.A += cr.Amount()
  if cr.Reconciled() {
    ocr.R = true
  }
  c.m[cr.Id()] = ocr
  return c
}

// SetPaymentId sets the payment id.
func (c *CatPaymentBuilder) SetPaymentId(x int64) *CatPaymentBuilder {
  c.initialize()
  c.pc.Id = x
  return c
}

// SetReconciled sets the reconciled flag.
func (c *CatPaymentBuilder) SetReconciled(x bool) *CatPaymentBuilder {
  c.initialize()
  c.r = x
  return c
}

func (c *CatPaymentBuilder) initialize() {
  if c.m == nil {
    c.m = make(map[Cat]CatRec)
    c.pc = Cat{Type: AccountCat} 
    c.r = false
  }
}

func (c *CatPaymentBuilder) newCatRecSlice() []CatRec {
  keys := make([]Cat, len(c.m))
  i := 0
  for k := range c.m {
    if k != c.pc {
      keys[i] = k
      i++
    }
  }
  keys = keys[:i]
  sort.Sort(catSlice(keys))
  ncr := make([]CatRec, len(keys))
  for i, k := range keys {
    ncr[i] = c.m[k]
  }
  return ncr
}

// ReviewStatus is the review status of an Entry.
type ReviewStatus int

const (
  Reviewed ReviewStatus = 1
  NotReviewed = 0
  ReviewInProgress = -1
)

// Entry represents a transaction entry.
type Entry struct {
  // Unique Id
  Id int64
  Date time.Time
  Name string
  Desc string
  CheckNo string
  CatPayment
  Status ReviewStatus
}

func (e *Entry) String() string {
  return fmt.Sprintf("%v", *e)
}

// GobEncode is included so that Entry values work with the etags package
func (e *Entry) GobEncode() (b []byte, err error) {
  buffer := bytes.NewBuffer(make([]byte, 0, 512))
  encoder := gob.NewEncoder(buffer)
  if err = encoder.Encode(e.Id); err != nil {
    return
  }
  if err = encoder.Encode(e.Date); err != nil {
    return
  }
  if err = encoder.Encode(e.Name); err != nil {
    return
  }
  if err = encoder.Encode(e.Desc); err != nil {
    return
  }
  if err = encoder.Encode(e.CheckNo); err != nil {
    return
  }
  if err = encoder.Encode(e.cr); err != nil {
    return
  }
  if err = encoder.Encode(e.id); err != nil {
    return
  }
  if err = encoder.Encode(e.r); err != nil {
    return
  }
  if err = encoder.Encode(e.Status); err != nil {
    return
  }
  b = buffer.Bytes()
  return
}

// EntryBalance is an Entry with an ending balance
type EntryBalance struct {
  Entry
  Balance int64
}

// Account represents an account for payment.
type Account struct {
  // Unique Id
  Id int64
  // Name
  Name string
  Active bool
  // Ending balance
  Balance int64
  // Reconciled balance
  RBalance int64
  // Count of all transactions
  Count int
  // Count of reconciled transactions
  RCount int
  // Auto import should ignore transactions before this date.
  ImportSD time.Time
}

func (a *Account) String() string {
  return fmt.Sprintf("%v", *a)
}

// AccountDelta represents changes in a single account.
type AccountDelta struct {
  // Balance is change in overall balance in cents.
  Balance int64
  // RBalance is change in overall reconciled balance in cents.
  RBalance int64
  // Count is the change in number of transactions.
  Count int
  // RCount is the change in number of reconciled transactions.
  RCount int
}

func (a *AccountDelta) String() string {
  return fmt.Sprintf("%v", *a)
}

func (a *AccountDelta) isZero() bool {
  return a.Balance == 0 && a.RBalance == 0 && a.Count == 0 && a.RCount == 0
}

func (a *AccountDelta) add(amount int64, reconciled bool, multiplier int) {
  amount *= int64(multiplier)
  a.Balance += amount
  a.Count += multiplier
  if reconciled {
    a.RBalance += amount
    a.RCount += multiplier
  }
}

// AccountDeltas represents changes in multiple accounts.
type AccountDeltas map[int64]*AccountDelta

// Include includes catPayment in these deltas.
func (a AccountDeltas) Include(catPayment *CatPayment) {
  a.add(catPayment, 1)
}

// Exclude excludes catPayment in these deltas.
func (a AccountDeltas) Exclude(catPayment *CatPayment) {
  a.add(catPayment, -1)
}

func (a AccountDeltas) add(catPayment *CatPayment, multiplier int) {
  var total int64
  for i := range catPayment.cr {
    catrec := &catPayment.cr[i]
    if catrec.C.Type == AccountCat {
      a._add(catrec.C.Id, catrec.A, catrec.R, multiplier)
    }
    total -= catrec.A
  }
  a._add(catPayment.id, total, catPayment.r, multiplier)
}

func (a AccountDeltas) _add(id int64, amount int64, reconciled bool, multiplier int) {
  delta := a[id]
  if delta == nil {
    delta = new(AccountDelta)
    a[id] = delta
  }
  delta.add(amount, reconciled, multiplier)
  if delta.isZero() {
    delete(a, id)
  }
}

// CatTotals represents category totals
type CatTotals map[Cat]int64

func (c CatTotals) Include(catPayment *CatPayment) {
  for i := range catPayment.cr {
    catrec := &catPayment.cr[i]
    if catrec.C.Type != AccountCat {
      c[catrec.C] += catrec.A
    }
  }
}

// AccountSet represents a set of account ids.
type AccountSet map[int64]bool

func (a AccountSet) Include(catPayment *CatPayment) {
  for i := range catPayment.cr {
    catrec := &catPayment.cr[i]
    if catrec.C.Type == AccountCat {
      a[catrec.C.Id] = true
    }
  }
  a[catPayment.id] = true
}

// Stream of EntryBalance adding ending balance to a Stream of Entry.
type AddBalanceStream struct {
  // ending balance
  Balance int64
  // original Stream of Entry
  functional.Stream
}

func (a *AddBalanceStream) Next(ptr interface{}) error {
  p := ptr.(*EntryBalance)
  err := a.Stream.Next(&p.Entry)
  if err == nil {
    p.Balance = a.Balance
    a.Balance -= p.Total()
  }
  return err
}

// User represents a user.
type User struct {
  Id int64
  Name string
  passwords.Password
}

// FormatUSD returns amount as dollars and cents.
// 347 -> "3.47"
func FormatUSD(x int64) string {
  return fmt.Sprintf("%.2f", float64(x) / 100.0)
}

// ParseUSD is the inverse of FormatUSD.
// "3.47" -> 347
func ParseUSD(s string) (v int64, e error) {
  f, e := strconv.ParseFloat(s, 64)
  if e != nil {
    return
  }
  v = int64(math.Floor(f * 100.0 + 0.5))
  return
}

type catSlice []Cat

func (s catSlice) Less(i, j int) bool {
  if s[i].Type < s[j].Type {
    return true
  }
  if s[i].Type == s[j].Type && s[i].Id < s[j].Id {
    return true
  }
  return false
}

func (s catSlice) Swap(i, j int) {
  s[j], s[i] = s[i], s[j]
}

func (s catSlice) Len() int {
  return len(s)
}

