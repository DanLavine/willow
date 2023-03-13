package datastructures

type TreeKey interface {
	Less(compareKey TreeKey) bool
}

type StringTreeKey struct {
	key string
}

func NewStringTreeKey(key string) *StringTreeKey {
	return &StringTreeKey{
		key: key,
	}
}

func (stk *StringTreeKey) Less(compareKey TreeKey) bool {
	return stk.key < compareKey.(*StringTreeKey).key
}

type IntTreeKey struct {
	key int
}

func NewIntTreeKey(key int) *IntTreeKey {
	return &IntTreeKey{
		key: key,
	}
}

func (itk *IntTreeKey) Less(compareKey TreeKey) bool {
	return itk.key < compareKey.(*IntTreeKey).key
}
