package filters

import (
  "github.com/keep94/finance/fin"
  "github.com/keep94/goconsume"
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

func runFilter(f goconsume.FilterFunc) int {
  result := 0
  if f(&fin.Entry{Name: "Name 1", Desc: "Desc 1"}){
    result++
  }
  if f(&fin.Entry{Name: "Name 2", Desc: "Other"}) {
    result++
  }
  if f(&fin.Entry{Name: "Other", Desc: "Other"}) {
    result++
  }
  if f(&fin.Entry{
      Name: "Name 3",
      Desc: "Desc 3",
      CatPayment: fin.NewCatPayment(fin.NewCat("0:7"), 200, false, 0)}) {
    result++
  }
  return result
}
