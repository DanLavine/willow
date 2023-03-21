package datastructures

// Any items that satisfy this interface can be stored in a BTree

//counterfeiter:generate . TreeItem
type TreeItem interface {
	// Generic callback functions that can be set on items stored in the BTree
	OnFind()
	//CanDelete() bool
}
