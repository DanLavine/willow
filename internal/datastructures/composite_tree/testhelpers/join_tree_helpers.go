package testhelpers

import (
	"fmt"
)

type JoinTreeTester struct {
	OnFindCount int
	Value       string
}

func NewJoinTreeTester(Value string) func() (any, error) {
	return func() (any, error) {
		return &JoinTreeTester{
			OnFindCount: 0,
			Value:       Value,
		}, nil
	}
}

func NewJoinTreeTesterWithError() (any, error) {
	return nil, fmt.Errorf("failure")
}

func OnFindTest(item any) {
	t := item.(*JoinTreeTester)
	t.OnFindCount++
}
