package treeitems

import "github.com/DanLavine/willow/internal/datastructures/set"

type IDHolders struct {
	Values []uint64
}

// can be used to create a new tree item
func NewIDHolder() (any, error) {
	return &IDHolders{
		Values: []uint64{},
	}, nil
}

// can be used to clear the set on creation of a new tree item
func NewIDHolderClearSet(set set.Set) func() (any, error) {
	set.Clear()

	return func() (any, error) {
		return &IDHolders{
			Values: []uint64{},
		}, nil
	}
}

// OnFindAdd adds any number of values to a passed in set
func OnFindAdd(set set.Set) func(item any) {
	return func(item any) {
		idHolders := item.(*IDHolders)
		set.Add(idHolders.Values)
	}
}

// OnFindKeep takes a union of what is already in a set and the values found
func OnFindKeep(set set.Set) func(item any) {
	return func(item any) {
		idHolders := item.(*IDHolders)
		set.Keep(idHolders.Values)
	}
}

func OnFindRemoveID(idToRemove uint64) func(item any) {
	return func(item any) {
		idHolders := item.(*IDHolders)

		for index, value := range idHolders.Values {
			if value == idToRemove {
				idHolders.Values[index] = idHolders.Values[len(idHolders.Values)-1]
				idHolders.Values = idHolders.Values[:len(idHolders.Values)-1]
				return
			}
		}
	}
}

// OnFindRemove removes any values to a provided set
func OnFindRemove(set set.Set) func(item any) {
	return func(item any) {
		idHolders := item.(*IDHolders)
		set.Keep(idHolders.Values)
	}
}

// CanRemove will return true iff there are no values
func CanRemove(item any) bool {
	idHolders := item.(*IDHolders)
	return len(idHolders.Values) == 0
}
