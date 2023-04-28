package testhelpers

import (
	"fmt"
)

type BTreeTester struct {
	OnFindCount int
	Value       string
}

func NewBTreeTester(Value string) func() (any, error) {
	return func() (any, error) {
		return &BTreeTester{
			OnFindCount: 0,
			Value:       Value,
		}, nil
	}
}
func NewBTreeTesterWithError() (any, error) {
	return nil, fmt.Errorf("failure")
}

func OnFindTest(item any) {
	btt := item.(*BTreeTester)
	btt.OnFindCount++
}
