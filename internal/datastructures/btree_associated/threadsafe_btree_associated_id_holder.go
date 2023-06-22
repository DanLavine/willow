package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
)

type idHolder struct {
	IDs []uint64
}

// can be used to create a new tree item
func newIDHolder(onFind datastructures.OnFind) func() any {
	return func() any {
		idHolder := &idHolder{
			IDs: []uint64{},
		}

		onFind(idHolder)

		return idHolder
	}
}

// can be used to clear the set on creation of a new tree item
func newIDHolderClearSet(set set.Set[uint64]) func() any {
	return func() any {
		set.Clear()

		return &idHolder{
			IDs: []uint64{},
		}
	}
}

func (idHolder *idHolder) add(id uint64) {
	idHolder.IDs = append(idHolder.IDs, id)
}

func (idHolder *idHolder) remove(idToRemove uint64) {
	for index, id := range idHolder.IDs {
		if id == idToRemove {
			idHolder.IDs[index] = idHolder.IDs[len(idHolder.IDs)-1]
			idHolder.IDs = idHolder.IDs[:len(idHolder.IDs)-1]
			return
		}
	}
}

//// onFindAdd adds any number of values to a passed in set
//func onFindIDHolderAdd(set set.Set[uint64]) func(item any) {
//	return func(item any) {
//		idHolder := item.(*idHolder)
//		set.Add(idHolder.IDs)
//	}
//}
//
//// onFindKeep takes a union of what is already in a set and the values found
//func onFindIDHolderKeep(set set.Set[uint64]) func(item any) {
//	return func(item any) {
//		idHolder := item.(*idHolder)
//		set.Keep(idHolder.IDs)
//	}
//}
//
//// onFindIDHolderRemoveID removes the ID holder
//func onFindIDHolderRemoveID(idToRemove uint64) func(item any) {
//	return func(item any) {
//		idHolder := item.(*idHolder)
//
//		for index, value := range idHolder.IDs {
//			if value == idToRemove {
//				idHolder.IDs[index] = idHolder.IDs[len(idHolder.IDs)-1]
//				idHolder.IDs = idHolder.IDs[:len(idHolder.IDs)-1]
//				return
//			}
//		}
//	}
//}
//
//// onFindRemove removes any values to a provided set
//func onFindRemove(set set.Set[uint64]) func(item any) {
//	return func(item any) {
//		idHolder := item.(*idHolder)
//		set.Remove(idHolder.IDs)
//	}
//}
//
//// CanRemove will return true iff there are no values
//func canRemoveIDHolder(item any) bool {
//	idHolder := item.(*idHolder)
//	return len(idHolder.IDs) == 0
//}
