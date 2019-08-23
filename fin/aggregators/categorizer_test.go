package aggregators

import (
  "github.com/keep94/finance/fin"
  "testing"
)

func TestByNameCategorizerNormal(t *testing.T) {
  builder := NewByNameCategorizerBuilder(6, 3)
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  verifyNoMatch(t, builder, "a")
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  verifyMatch(t, builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  verifyMatch(t, builder, "a", fin.NewCat("0:1"))
  verifyNoMatch(t, builder, "b")
}

func TestByNameCategorizerNoPrevelant(t *testing.T) {
  builder := NewByNameCategorizerBuilder(6, 3)
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:3"))
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:3"))
  addToBuilder(builder, "a", fin.NewCat("0:3"))
  verifyNoMatch(t, builder, "a")
}

func TestByNameCategorizerMixUp(t *testing.T) {
  builder := NewByNameCategorizerBuilder(6, 3)
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:3"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:3"))
  addToBuilder(builder, "a", fin.NewCat("0:2"))
  verifyMatch(t, builder, "a", fin.NewCat("0:2"))
}

func TestByNameCategorizerSplitTrx(t *testing.T) {
  builder := NewByNameCategorizerBuilder(6, 3)
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:1"), fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:1"), fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "a", fin.NewCat("0:1"), fin.NewCat("0:2"))
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  verifyNoMatch(t, builder, "a")
}

func TestByNameCategorizerNameVariations(t *testing.T) {
  builder := NewByNameCategorizerBuilder(6, 3)
  addToBuilder(builder, "  Safeway Store001", fin.NewCat("0:1"))
  addToBuilder(builder, "SAFEWAY STORE 003", fin.NewCat("0:1"))
  addToBuilder(builder, "SAFEWAY #06 STORE", fin.NewCat("0:1"))
  verifyMatch(t, builder, "Safeway Store", fin.NewCat("0:1"))
}

func TestByNameCategorizerMultNames(t *testing.T) {
  builder := NewByNameCategorizerBuilder(1, 1)
  addToBuilder(builder, "a", fin.NewCat("0:1"))
  addToBuilder(builder, "b", fin.NewCat("0:2"))
  addToBuilder(builder, "c", fin.NewCat("0:3"))
  verifyMatch(t, builder, "a", fin.NewCat("0:1"))
  verifyMatch(t, builder, "b", fin.NewCat("0:2"))
  verifyMatch(t, builder, "c", fin.NewCat("0:3"))
  verifyNoMatch(t, builder, "d")
}

func TestByNameCategorizerCategorizeAccounts(t *testing.T) {
  builder := NewByNameCategorizerBuilder(1, 1)
  addToBuilder(builder, "a", fin.NewCat("2:1"))
  categorizer := builder.Build()
  entry := fin.Entry{
      Name: "a",
      CatPayment: fin.NewCatPayment(fin.Expense, 357, false, 0)}
  if !categorizer.Categorize(&entry) {
    t.Error("Expected a match.")
  }
  entry = fin.Entry{
      Name: "a",
      CatPayment: fin.NewCatPayment(fin.Expense, 357, false, 1)}
  if categorizer.Categorize(&entry) {
    t.Error("Did not expect a match.")
  }

}

func addToBuilder(builder *ByNameCategorizerBuilder, name string, cats ...fin.Cat) {
  cp := fin.CatPaymentBuilder{}
  for _, cat := range cats {
    cp.AddCatRec(fin.CatRec{Cat: cat, Amount: 100})
  }
  entry := fin.Entry{Name: name, CatPayment: cp.Build()}
  builder.Include(&entry)
}

func verifyNoMatch(t *testing.T, builder *ByNameCategorizerBuilder, name string) {
  verifyMatch(t, builder, name, fin.Expense)
}

func verifyMatch(t *testing.T, builder *ByNameCategorizerBuilder, name string, cat fin.Cat) {
  categorizer := builder.Build()
  entry := fin.Entry{
      Name: name,
      CatPayment: fin.NewCatPayment(fin.Expense, 357, false, 0)}
  if !categorizer.Categorize(&entry) {
    t.Error("Expected a match.")
    return
  }
  if output := entry.CatRecCount(); output != 1 {
    t.Errorf("Expected 1 category, got %d", output)
    return
  }
  catRec := entry.CatRecByIndex(0)
  if catRec.Cat != cat {
    t.Errorf("Expected %v, got %v", cat, catRec.Cat)
  }
}

 
