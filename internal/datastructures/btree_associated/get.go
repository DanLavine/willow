package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (at *associatedTree) Get(keyValuePairs datatypes.StringMap, onFind datastructures.OnFind) (any, error) {
	if len(keyValuePairs) == 0 {
		return nil, fmt.Errorf("keyValuePairs requires a length of at least 1")
	}

	// NOTE: The read lock for the entire association. Since maps are unordered we need the entire grouped to be locked
	castableAssociatedKeyValues := at.groupedKeyValueAssociation.Find(datatypes.Int(len(keyValuePairs)), keyValuesReadLock)
	if castableAssociatedKeyValues == nil {
		return nil, nil
	}
	associatedKeyValues := castableAssociatedKeyValues.(*keyValues)
	defer associatedKeyValues.lock.RUnlock()

	firstLoop := true
	idSet := set.New()

	for searchKey, searchValue := range keyValuePairs {
		// create or find the values for a particular key
		castableValuesForKey := associatedKeyValues.values.Find(searchKey, nil)
		if castableValuesForKey == nil {
			return nil, nil
		}
		valuesForKey := castableValuesForKey.(*keyValues)

		// filter the IDs associated with the particualr value
		item := valuesForKey.values.Find(searchValue, nil)
		if item == nil {
			return nil, nil
		}

		idHolder := item.(*idHolder)

		if firstLoop {
			idSet.Add(idHolder.IDs)
			firstLoop = false
		} else {
			idSet.Keep(idHolder.IDs)
		}
	}

	switch len(idSet.Values()) {
	case 1:
		// if we reach this point, we are guranteed to only have 1 value
		returnValue := at.idTree.Get(idSet.Values()[0])
		if onFind != nil {
			onFind(returnValue)
		}
		return returnValue, nil
	default:
		return nil, nil
	}
}
