package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
)

// Query can be used to find a single or collection of items that match specific criteria. This
// is thread safe to call with any of the other functions on this object
//
//	PARAMETERS:
//	- selection - the selection for the specified items to find. See the Select docs for this param specifically
//	- onQueryPagination - is the callback used for an items found in the tree. It will recive the objects' value saved in the tree (what were originally provided)
//
//	RETURNS:
//	- error - any errors encountered with the parameters or destroying errors
//	        - 1. fmt.Errorf(...) // if query is not valid
//	        - 2. datastructures.ErrorsOnQueryPaginateNil
//	        - 3. datastructures.ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) QueryAction(query *queryassociatedaction.AssociatedActionQuery, onQueryPagination BTreeAssociatedIterate) error {
	// param validation
	if err := query.Validate(); err != nil {
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

	// select all
	if query.Selection == nil && query.And == nil && query.Or == nil {
		wrappedPagination := func(_ datatypes.EncapsulatedValue, item any) bool {
			return onQueryPagination(item.(AssociatedKeyValues))
		}

		if err := tsat.associatedIDs.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, wrappedPagination); err != nil {
			panic(err)
		}

		return nil
	}

	// need to filter for key values
	shouldContinue := true
	validIDs := tsat.query(query)

	for _, id := range validIDs.Values() {
		err := tsat.associatedIDs.Find(datatypes.String(id), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
			shouldContinue = onQueryPagination(item.(AssociatedKeyValues))
			return false
		})

		if err != nil {
			switch err {
			case btree.ErrorKeyDestroying:
				// this is fine, just skip this associated id
			default:
				// if we get here, ther was an error with the tree. need to fix the tree
				panic(err)
			}
		}

		// break querying more items
		if !shouldContinue {
			break
		}
	}

	return nil
}

// query is used to construct a list of valid ID's to search for
//
//	PARAMS:
//	- query - They user provided query to parse through
func (tsat *threadsafeAssociatedTree) query(query *queryassociatedaction.AssociatedActionQuery) set.Set[string] {
	// need to filter for key values
	var validIDs set.Set[string]
	validCounter := 0

	// find IDs based off the Where clause
	if query.Selection != nil {
		validIDs = tsat.findIDs(query.Selection)
		validCounter++
	}

	// intersect all the ANDs together
	for _, andSelection := range query.And {
		// can exit early if we know that there are 0 ids. Don't need to search
		if validCounter >= 1 && validIDs.Size() == 0 {
			// must not have any ids, so can exit early and stop selecting everything
			break
		}

		subsetValidIDs := tsat.query(andSelection)

		switch validCounter {
		case 0:
			validIDs = subsetValidIDs
			validCounter++
		default:
			validIDs.Intersection(subsetValidIDs.Values())
		}
	}

	// Can add all the IDs found from an OR join
	for _, orSelection := range query.Or {
		subsetValidIDs := tsat.query(orSelection)

		// Need to join our OR's with the current values
		switch validCounter {
		case 0:
			validIDs = subsetValidIDs
			validCounter++
		default:
			// add the subset to know values that we are about to check
			validIDs.AddBulk(subsetValidIDs.Values())
		}
	}

	return validIDs
}

func (tsat *threadsafeAssociatedTree) findIDs(associatedSelection *queryassociatedaction.Selection) set.Set[string] {
	// if there are associated IDs, then we need to run the key values query against those values
	if len(associatedSelection.IDs) != 0 {
		combinedIDs := set.New[string]()

		for _, associatedID := range associatedSelection.IDs {
			onFind := func(key datatypes.EncapsulatedValue, item any) bool {
				associatedKeyValues := item.(*associatedKeyValues)

				// run the KeyValues on the found item, against the rest of the query to ensure it matches
				if associatedSelection.MatchKeyValues(associatedKeyValues.keyValues) {
					combinedIDs.Add(associatedID)
				}

				return true
			}

			if err := tsat.associatedIDs.Find(datatypes.String(associatedID), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
				switch err {
				case btree.ErrorKeyDestroying:
					// this is fine, just skip this value
					break
				default:
					panic(err)
				}
			}

		}

		// can break early since there are no IDs that matched the selection
		return combinedIDs
	}
	// Recursively walk all the IDs for a specific key
	getAllIDs := func(allIDs *[]string) func(_ datatypes.EncapsulatedValue, item any) bool {
		return func(_ datatypes.EncapsulatedValue, item any) bool {
			// always casting to an ID node when finding possible indexes
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			for index, ids := range idNode.ids {
				if associatedSelection.MinNumberOfKeyValues != nil {
					// Note the +1 here. if a single key makes up the whole slection, then that is at index 0. So we need to
					// adjust for slices/arrays being 0 indexed, where APIs are 1 indexed
					if index+1 < *associatedSelection.MinNumberOfKeyValues {
						continue
					}
				}

				if associatedSelection.MaxNumberOfKeyValues != nil {
					// Note the +1 here. if a single key makes up the whole slection, then that is at index 0. So we need to
					// adjust for slices/arrays being 0 indexed, where APIs are 1 indexed
					if index+1 > *associatedSelection.MaxNumberOfKeyValues {
						break
					}
				}

				// include all the ids
				*allIDs = append(*allIDs, ids...)
			}

			return true
		}
	}

	// first step can be to just grab all the threadsafe value nodes for all the possible keys
	validAdded := false
	invalidAdded := false

	validIDs := set.New[string]()
	invalidIDs := set.New[string]()

	for _, queryKey := range associatedSelection.SortedKeys() {
		allIDS := []string{}
		associatedValue := associatedSelection.KeyValues[queryKey]

		switch associatedValue.Comparison {
		case v1common.Equals:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.Find(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if validAdded {
				validIDs.Intersection(allIDS)
			} else {
				validAdded = true
				validIDs.AddBulk(allIDS)
			}
		case v1common.NotEquals:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.Find(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if invalidAdded {
				invalidIDs.Intersection(allIDS)
			} else {
				invalidAdded = true
				invalidIDs.AddBulk(allIDS)
			}
		case v1common.LessThan:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.FindLessThan(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if validAdded {
				validIDs.Intersection(allIDS)
			} else {
				validAdded = true
				validIDs.AddBulk(allIDS)
			}
		case v1common.LessThanOrEqual:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.FindLessThanOrEqual(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if validAdded {
				validIDs.Intersection(allIDS)
			} else {
				validAdded = true
				validIDs.AddBulk(allIDS)
			}
		case v1common.GreaterThan:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.FindGreaterThan(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if validAdded {
				validIDs.Intersection(allIDS)
			} else {
				validAdded = true
				validIDs.AddBulk(allIDS)
			}
		case v1common.GreaterThanOrEqual:
			_ = tsat.keys.Find(datatypes.String(queryKey), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, func(key datatypes.EncapsulatedValue, item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				_ = valuesNode.values.FindGreaterThanOrEqual(associatedValue.Value, associatedValue.TypeRestrictions, getAllIDs(&allIDS))
				return false
			})

			if validAdded {
				validIDs.Intersection(allIDS)
			} else {
				validAdded = true
				validIDs.AddBulk(allIDS)
			}
		default:
			panic(fmt.Errorf("Query key '%s' has an unknown comparison recevied '%s'", queryKey, associatedValue.Comparison))
		}

		// exit early
		if validAdded && validIDs.Size() == 0 {
			return validIDs
		}
	}

	if !validAdded {
		// Special case where the only query is a FALSE check. So need to also find all other IDs and to perform a union
		// NOTE: also important to not update the IDNodes query here as this will be a lot of items to store additionally. So not worth the extra memory usage

		allIDs := []string{}
		tsat.keys.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, func(key datatypes.EncapsulatedValue, item any) bool {
			valuesNode := item.(*threadsafeValuesNode)

			_ = valuesNode.values.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, getAllIDs(&allIDs))
			validIDs.AddBulk(allIDs)
			return true
		})
	}

	// filter the IDs to a list of what is only acceptable
	for _, id := range invalidIDs.Values() {
		validIDs.Remove(id)
	}

	return validIDs
}

/*
func (tsat *threadsafeAssociatedTree) generateKeyValuePermutations(associatedSelection *queryassociatedaction.AssociatedActionQuery, validIds, invalidIds map[string][][]string) set.Set[string] {
	validIDKeys := []string{}
	for key, _ := range validIds {
		validIDKeys = append(validIDKeys, key)
	}

	invalidIDKeys := []string{}
	for key, _ := range invalidIds {
		invalidIDKeys = append(invalidIDKeys, key)
	}

	minKeys := 0
	if associatedSelection.MinNumberOfPermutationKeyValues != nil {
		minKeys = *associatedSelection.MinNumberOfPermutationKeyValues
	}

	maxKeys := -1
	if associatedSelection.MaxNumberOfPermutationKeyValues != nil {
		maxKeys = *associatedSelection.MaxNumberOfPermutationKeyValues
	}

	keyCombinations, err := datamanipulation.GenerateStringPermutations(append(validIDKeys, invalidIDKeys...), minKeys, maxKeys)
	if err != nil {
		panic(err)
	}

	combinedIDs := set.New[string]()
	allIDs := []string{}
	for _, grouping := range keyCombinations {
		validGroupedIDs := set.New[string]()
		invalidGroupedIDs := set.New[string]()

		// for all the keys that make up a prticualr grouping, find the possible KeyValue pairs
		inclusive := false
		for index, key := range grouping {
			// when the grouping has a valid id
			if ids, ok := validIds[key]; ok {
				inclusive = true

				combinedIDs := []string{}
				for _, idsForKeyLen := range ids {
					combinedIDs = append(combinedIDs, idsForKeyLen...)
				}

				if index == 0 {
					validGroupedIDs.AddBulk(combinedIDs)
				} else {
					validGroupedIDs.Intersection(combinedIDs)
				}

				continue
			}

			// when the grouping has an invalid id
			if ids, ok := invalidIds[key]; ok {
				combinedIDs := []string{}
				for _, idsForKeyLen := range ids {
					combinedIDs = append(combinedIDs, idsForKeyLen...)
				}

				if index == 0 {
					invalidGroupedIDs.AddBulk(combinedIDs)
				} else {
					invalidGroupedIDs.Intersection(combinedIDs)
				}
			}
		}

		switch inclusive {
		case true:
			// nothing to do here, will use shared logic at the end of switch
		default:
			// special case for when the grouping was only NOT queries

			// Recursively walk all the IDs for a specific key
			getAllIDs := func(allIDs *[]string) func(_ datatypes.EncapsulatedValue, item any) bool {
				return func(_ datatypes.EncapsulatedValue, item any) bool {
					// always casting to an ID node when finding possible indexes
					idNode := item.(*threadsafeIDNode)
					idNode.lock.RLock()
					defer idNode.lock.RUnlock()

					for index, ids := range idNode.ids {
						if associatedSelection.MinNumberOfKeyValues != nil {
							// Note the +1 here. if a single key makes up the whole slection, then that is at index 0. So we need to
							// adjust for slices/arrays being 0 indexed, where APIs are 1 indexed
							if index+1 < *associatedSelection.MinNumberOfKeyValues {
								continue
							}
						}

						if associatedSelection.MaxNumberOfKeyValues != nil {
							// Note the +1 here. if a single key makes up the whole slection, then that is at index 0. So we need to
							// adjust for slices/arrays being 0 indexed, where APIs are 1 indexed
							if index+1 > *associatedSelection.MaxNumberOfKeyValues {
								break
							}
						}

						// include all the ids
						*allIDs = append(*allIDs, ids...)
					}

					return true
				}
			}

			// only need to walk the IDs once. Otherwise, just use the list we already have saved
			if len(allIDs) == 0 {
				tsat.keys.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, func(key datatypes.EncapsulatedValue, item any) bool {
					valuesNode := item.(*threadsafeValuesNode)

					_ = valuesNode.values.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, getAllIDs(&allIDs))

					validGroupedIDs.AddBulk(allIDs)
					return true
				})
			} else {
				validGroupedIDs.AddBulk(allIDs)
			}
		}

		// filter the IDs to a list of what is only acceptable
		for _, id := range invalidGroupedIDs.Values() {
			validGroupedIDs.Remove(id)
		}

		// add all the possible permutations
		combinedIDs.AddBulk(validGroupedIDs.Values())
	}

	return combinedIDs
}
*/
