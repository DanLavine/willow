package btreeassociated

type JoinTreeTester struct {
	OnFindCount int
	Value       string
}

func NewJoinTreeTester(Value string) func() any {
	return func() any {
		return &JoinTreeTester{
			OnFindCount: 0,
			Value:       Value,
		}
	}
}

func OnFindTest(item any) {
	t := item.(*JoinTreeTester)
	t.OnFindCount++
}
