package btreeshared

//import (
//	"fmt"
//
//	"github.com/DanLavine/willow/internal/datastructures"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//)
//
//func (tst *threadsafeSharedTree) CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error {
//	if len(keyValuePairs) == 0 {
//		return fmt.Errorf("keyValuePairs cannot be empty")
//	}
//
//	if onCreate == nil {
//		return fmt.Errorf("onCreate cannot be nil")
//	}
//
//	if onFind == nil {
//		return fmt.Errorf("onFind cannot be nil")
//	}
//
//	//// always attempt a read first since these are read locked
//	//found := false
//	//wrappedOnFind := func(item any) {
//	//	found = true
//	//	onFind(item)
//	//}
//	//_ = at.Find(keyValuePairs, wrappedOnFind)
//	//if found == true {
//	//	return nil
//	//}
//
//	sortedKeys := keyValuePairs.SoretedKeys()
//
//	createKey := func() any {
//		return newValuesNode()
//	}
//
//	findKey := func(item any) {
//		valuesNode := item.(threadsafeValuesNode)
//	}
//
//	for _, key := range sortedKeys {
//		tst.keys.CreateOrFind(datatypes.String(key), createKey(findKey), findKey)
//	}
//
//	return nil
//}
