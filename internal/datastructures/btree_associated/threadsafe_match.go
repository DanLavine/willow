package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/datamanipulation"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
)

// MatchPermutations can be used to find any combination of key values.
// I.E. suppose we we have keyValues = {"one": 1, 2: "something else", "float num": 3.42}
// to find all combinations, we can think of would be [{"one":1},{2:"something else"},{"float num": 3.42},{"one":1, 2: "something else"}, {"one":1, "float num": 3.42} ...]
//
// In the case of a large collection of KeyValues, a similar query would be massive to join all the possible combinations together.
// This is an optimization of performing those workflows in a reasonable way. It is important to note that any single associated id
// being destroyed will simply be skipped over
//
//	PARAMETERS:
//	- keyValues - the possible key values to join together when searching for items in a tree
//	- onQueryPagination - is the callback used for an items found in the tree. It will recive the objects' value saved in the tree (what were originally provided)
//
//	RETURNS:
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorsOnIterateNil
//	          3. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) MatchAction(matchActionQuery *querymatchaction.MatchActionQuery, onQueryPagination BTreeAssociatedIterate) error {
	// check parameters
	if err := matchActionQuery.Validate(); err != nil {
		return err
	}
	if onQueryPagination == nil {
		return ErrorsOnIterateNil
	}

	// tree destroying check
	if !tsat.destroySyncer.GuardOperation() {
		return ErrorTreeDestroying
	}
	defer tsat.destroySyncer.ClearOperation()

	// 1. find all the Keys
	keyIDs := map[string][][]string{}
	for _, key := range matchActionQuery.KeyValues.SortedKeys() {
		// NOTE: don't need to check datastructures.ErrorKeyDestroying here as that is only on the associated IDs
		findValue := func(dbKey datatypes.EncapsulatedValue, item any) bool {
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			for index, ids := range idNode.ids {
				// skip the key value pairs that don't have enough keys
				if matchActionQuery.MinNumberOfPermutationKeyValues != nil {
					if index+1 < *matchActionQuery.MinNumberOfPermutationKeyValues {
						keyIDs[key] = append(keyIDs[key], []string{})
						continue
					}
				}

				// stop the key value pairs that are greater than the allowed pairing
				if matchActionQuery.MaxNumberOfPermutationKeyValues != nil {
					if index+1 > *matchActionQuery.MaxNumberOfPermutationKeyValues {
						break
					}
				}

				if len(keyIDs[key])-1 >= index {
					// copy any values currently saved
					keyIDs[key][index] = append(keyIDs[key][index], ids...)
				} else {
					// add the new values
					keyIDs[key] = append(keyIDs[key], []string{})
					keyIDs[key][index] = append(keyIDs[key][index], ids...)
				}
			}

			return true
		}

		findKey := func(_ datatypes.EncapsulatedValue, valueNodeInterface any) bool {
			valuesNode := valueNodeInterface.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(matchActionQuery.KeyValues[key].Value, matchActionQuery.KeyValues[key].TypeRestrictions, findValue); err != nil {
				panic(err)
			}

			return true
		}

		// find the keys which are always strings
		if err := tsat.keys.Find(datatypes.String(key), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, findKey); err != nil {
			panic(err)
		}
	}

	// 2. For all the ID nodes, we need to group them together and find the values. Then run the callback on each found value.
	validIDKeys := []string{}
	for key, _ := range keyIDs {
		validIDKeys = append(validIDKeys, key)
	}

	minKeys := 0
	if matchActionQuery.MinNumberOfPermutationKeyValues != nil {
		minKeys = *matchActionQuery.MinNumberOfPermutationKeyValues
	}

	maxKeys := -1
	if matchActionQuery.MaxNumberOfPermutationKeyValues != nil {
		maxKeys = *matchActionQuery.MaxNumberOfPermutationKeyValues
	}

	keyCombinations, err := datamanipulation.GenerateStringPermutations(validIDKeys, minKeys, maxKeys)
	if err != nil {
		panic(err)
	}

	for _, grouping := range keyCombinations {
		validGroupedIDs := set.New[string]()

		// for all keys that make a grouping, we need to find the combined values
		for index, key := range grouping {
			// have a key with 0 ids
			if len(keyIDs[key]) < len(grouping) {
				validGroupedIDs.Clear()
				break
			}

			if index == 0 {
				validGroupedIDs.AddBulk(keyIDs[key][len(grouping)-1])
			} else {
				validGroupedIDs.Intersection(keyIDs[key][len(grouping)-1])
			}
		}

		// run the callback for all the found IDs
		for _, id := range validGroupedIDs.Values() {
			shouldContinue := false
			tsat.associatedIDs.Find(datatypes.String(id), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				shouldContinue = onQueryPagination(item.(AssociatedKeyValues))
				return shouldContinue
			})

			// break if the caller wants to stop
			if !shouldContinue {
				return nil
			}
		}
	}

	return nil
}
