package datastructures

type EnumerableTreeKeys interface {
	// generic callback function for enumerable tree keys. If the Callback function returns
	// false, they Enumerator is expect to stop processing and bail out
	Each(callback func(index int, key TreeKey) bool)

	// return the len(...) call of underlying type
	Len() int
}

type TreeKey interface {
	Less(compareKey any) bool
}

//type StringTreeKey struct {
//	key string
//}
//
//func NewStringTreeKey(key string) *StringTreeKey {
//	return &StringTreeKey{
//		key: key,
//	}
//}
//
//func (stk *StringTreeKey) Less(compareKey TreeKey) bool {
//	return stk.key < compareKey.(*StringTreeKey).key
//}
//

// used for testing BTree
type EnumerableIntTreeKey []IntTreeKey

func (eitk EnumerableIntTreeKey) Each(callback func(index int, key TreeKey) bool) {
	for index, key := range eitk {
		if !callback(index, key) {
			break
		}
	}
}

func (eitk EnumerableIntTreeKey) Len() int {
	return len(eitk)
}

type IntTreeKey int

func (itk IntTreeKey) Less(compareKey any) bool {
	return itk < compareKey.(IntTreeKey)
}
