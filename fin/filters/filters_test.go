package filters

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/gofunctional3/functional"
  "testing"
)

func TestCompileAdvanceSearchSpec(t *testing.T) {
  if output := runFilter(CompileAdvanceSearchSpec(
      &AdvanceSearchSpec{
          Name: "Name"})); output != 3 {
    t.Errorf("Expected 3, got %v", output)
  }
  if output := runFilter(CompileAdvanceSearchSpec(
      &AdvanceSearchSpec{
          Name: "Name",
          Desc: "Desc"})); output != 2 {
    t.Errorf("Expected 2, got %v", output)
  }
  if output := runFilter(CompileAdvanceSearchSpec(
      &AdvanceSearchSpec{
          Name: "Name",
          Desc: "Desc",
          CF: func(c fin.Cat) bool { return c == fin.NewCat("0:7")}})); output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
  if output := runFilter(CompileAdvanceSearchSpec(
      &AdvanceSearchSpec{
          AF: func(amt int64) bool { return amt == -200 }})); output != 1 {
    t.Errorf("Expected 1, got %v", output)
  }
  if output := runFilter(CompileAdvanceSearchSpec(
      &AdvanceSearchSpec{
          AF: func(amt int64) bool { return amt == -201 }})); output != 0 {
    t.Errorf("Expected 0, got %v", output)
  }
}

func runFilter(f functional.Filterer) int {
  cpb := fin.CatPaymentBuilder{}
  result := 0
  if f.Filter(&fin.Entry{Name: "Name 1", Desc: "Desc 1"}) == nil {
    result++
  }
  if f.Filter(&fin.Entry{Name: "Name 2", Desc: "Other"}) == nil {
    result++
  }
  if f.Filter(&fin.Entry{Name: "Other", Desc: "Other"}) == nil {
    result++
  }
  if f.Filter(&fin.Entry{
      Name: "Name 3",
      Desc: "Desc 3",
      CatPayment: cpb.AddCatRec(&fin.CatRec{fin.NewCat("0:7"), 200, false}).Build()}) == nil {
    result++
  }
  return result
}
