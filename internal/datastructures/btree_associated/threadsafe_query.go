package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Query can be used to find a single or collection of items that match specific criteria. This
// is thread safe to call with any of the other functions ono this object
//
// PARAMS:
// - selection - the selection for the specified items to find. See the Select docs for this param specifically
// - onFindPagination - is the callback used for an items found in the tree. It will recive the objects' value saved in the tree (what were originally provided)
//
// RETURNS:
// - error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) Query(query datatypes.AssociatedKeyValuesQuery, onFindPagination datastructures.OnFindPagination) error {
	if err := query.Validate(); err != nil {
		return err
	}
	if onFindPagination == nil {
		return fmt.Errorf("onFindPagination cannot be nil")
	}

	if query.KeyValueSelection == nil && query.And == nil && query.Or == nil {
		// select all
		tsat.associatedIDs.Iterate(onFindPagination)
	} else {
		// need to filter for key values
		validIDs := tsat.query(query)

		shouldContinue := true
		for _, id := range validIDs.Values() {
			tsat.associatedIDs.Find(datatypes.String(id), func(item any) {
				shouldContinue = onFindPagination(item)
			})

			// break querying more items
			if !shouldContinue {
				break
			}
		}
	}

	return nil
}

func (tsat *threadsafeAssociatedTree) query(query datatypes.AssociatedKeyValuesQuery) set.Set[string] {
	// need to filter for key values
	var validIDs set.Set[string]
	validCounter := 0

	// find the first where
	if query.KeyValueSelection != nil {
		validIDs = tsat.findIDs(query)
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

	// check the OR cases as well
	for _, orSelection := range query.Or {
		subsetValidIDs := tsat.query(orSelection)

		// Need to interset our OR's with the current values
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

func (tsat *threadsafeAssociatedTree) findIDs(dbQuery datatypes.AssociatedKeyValuesQuery) set.Set[string] {
	validCounter := 0
	validIDs := set.New[string]()

	invalidCounter := 0
	invalidIDs := set.New[string]()

	// parse the where clause
	inclusive := false
	if dbQuery.KeyValueSelection != nil {
		keyValuesSelection := dbQuery.KeyValueSelection

		getAllIDs := func(allIDs *[]string) func(item any) bool {
			return func(item any) bool {
				// always casting to an ID node when finding possible indexes
				idNode := item.(*threadsafeIDNode)
				idNode.lock.RLock()
				defer idNode.lock.RUnlock()

				for index, ids := range idNode.ids {
					// break if we reach a key count that has more than the requested length
					if keyValuesSelection.Limits != nil {
						if index >= *keyValuesSelection.Limits.NumberOfKeys {
							//return false
							break
						}
					}

					// include all the ids
					*allIDs = append(*allIDs, ids...)
				}

				return true
			}
		}

		addValidIDs := func(item any) bool {
			// always casting to an ID node when finding possible indexes
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			var possibleIDs set.Set[string]
			for index, ids := range idNode.ids {
				// break if we reach a key count that has more than the requested length
				if keyValuesSelection.Limits != nil {
					if index >= *keyValuesSelection.Limits.NumberOfKeys {
						return false
					}
				}

				// include all the ids
				switch validCounter {
				case 0:
					validIDs.AddBulk(ids)
				default:
					if possibleIDs == nil {
						possibleIDs = set.New[string]()
					}

					possibleIDs.AddBulk(ids)
				}
			}

			if possibleIDs != nil {
				validIDs.Intersection(possibleIDs.Values())
			}

			return true
		}

		addInvalidIDs := func(item any) bool {
			// always casting to an ID node when finding possible indexes
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			if invalidCounter >= 1 && invalidIDs.Size() == 0 {
				return true
			}

			var possibleInvalidIDs set.Set[string]
			for index, ids := range idNode.ids {
				// break if we reach a key count that has more than the requested length
				if keyValuesSelection.Limits != nil {
					if index >= *keyValuesSelection.Limits.NumberOfKeys {
						return false
					}
				}

				// include all the ids
				switch invalidCounter {
				case 0:
					invalidIDs.AddBulk(ids)
				default:
					if possibleInvalidIDs == nil {
						possibleInvalidIDs = set.New[string]()
					}

					possibleInvalidIDs.AddBulk(ids)
				}
			}

			if possibleInvalidIDs != nil {
				invalidIDs.Intersection(possibleInvalidIDs.Values())
			}

			return true
		}

		sortedKeys := keyValuesSelection.SortedKeys()
		for _, key := range sortedKeys {
			value := keyValuesSelection.KeyValues[key]
			allIDS := []string{}

			// special case for the _reserved_id
			if key == ReservedID {
				onFind := func(item any) {
					strValue := value.Value.Value.(string)

					switch validCounter {
					case 0:
						if keyValuesSelection.Limits != nil {
							if len(item.(*AssociatedKeyValues).keyValues) <= *keyValuesSelection.Limits.NumberOfKeys+1 {
								validIDs.Add(strValue)
							}
						} else {
							validIDs.Add(strValue)
						}
					default:
						if keyValuesSelection.Limits != nil {
							if len(item.(*AssociatedKeyValues).keyValues) <= *keyValuesSelection.Limits.NumberOfKeys+1 {
								validIDs.Intersection([]string{strValue})
							} else {
								validIDs.Remove(strValue)
							}
						} else {
							validIDs.Intersection([]string{strValue})
						}
					}

					inclusive = true
					validCounter++
				}

				_ = tsat.associatedIDs.Find(datatypes.String(value.Value.Value.(string)), onFind)

				continue
			}

			// existence check
			if value.Exists != nil {
				switch *value.Exists {
				case true:
					inclusive = true

					tsat.keys.Find(datatypes.String(key), func(item any) {
						valuesNode := item.(*threadsafeValuesNode)

						if value.ExistsType != nil {
							// add all values for the  desired type
							valuesNode.values.IterateMatchType(*value.ExistsType, getAllIDs(&allIDS))
						} else {
							// add all values
							valuesNode.values.Iterate(getAllIDs(&allIDS))
						}

						switch validCounter {
						case 0:
							validIDs.AddBulk(allIDS)
						default:
							validIDs.Intersection(allIDS)
						}
					})

					validCounter++
				case false:
					tsat.keys.Find(datatypes.String(key), func(item any) {
						valuesNode := item.(*threadsafeValuesNode)

						if value.ExistsType != nil {
							valuesNode.values.IterateMatchType(*value.ExistsType, getAllIDs(&allIDS))
						} else {
							valuesNode.values.Iterate(getAllIDs(&allIDS))
						}

						switch invalidCounter {
						case 0:
							invalidIDs.AddBulk(allIDS)
						default:
							invalidIDs.Intersection(allIDS)
						}
					})

					invalidCounter++
				}

				continue
			}

			// comparison check
			switch *value.ValueComparison {
			case datatypes.Equals():
				inclusive = true
				found := false

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.Find(*value.Value, func(item any) {
						found = true
						addValidIDs(item)
					})
				})

				validCounter++

				// didn't find an expected key, so clear out the results
				if !found {
					validIDs.Clear()
				}
			case datatypes.NotEquals():
				// if this is the first request to come through, then we need to record an invalid ID.
				// if it is the n+ request tto ome through, then I think we can use the "invalid id" lookups...
				// if it is only != requests, then we still need to iterate over all values...
				found := false

				// find the key to strip all values
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					valuesNode.values.Find(*value.Value, func(item any) {
						found = true
						addInvalidIDs(item)
					})
				})

				invalidCounter++

				if !found {
					// if this is the case, then we can increate the invalid counters. AKA everything is valid
					invalidIDs.Clear()
				}

			case datatypes.LessThan():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindLessThan(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.LessThanMatchType():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindLessThanMatchType(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.LessThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindLessThanOrEqual(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.LessThanOrEqualMatchType():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindLessThanOrEqualMatchType(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.GreaterThan():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindGreaterThan(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.GreaterThanMatchType():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindGreaterThanMatchType(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.GreaterThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindGreaterThanOrEqual(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			case datatypes.GreaterThanOrEqualMatchType():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)
					valuesNode.values.FindGreaterThanOrEqualMatchType(*value.Value, getAllIDs(&allIDS))

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
					default:
						validIDs.Intersection(allIDS)
					}
				})

				validCounter++
			}

			// stop querying when we know that there are 0 ids we are interessted in
			if validCounter >= 1 && validIDs.Size() == 0 {
				break
			}
		}

		// Special case where the only query is a FALSE check. So need to also find all other IDs and to perform a union
		if !inclusive {
			allIDs := []string{}
			tsat.keys.Iterate(func(item any) bool {
				valuesNode := item.(*threadsafeValuesNode)

				valuesNode.values.Iterate(func(item any) bool {
					// always return true here. the false is only for reaching limits, but we need to iterate over all values
					_ = getAllIDs(&allIDs)(item)
					return true
				})

				return true
			})

			validIDs.AddBulk(allIDs)
		}
	}

	// filter the IDs to a list of what is only acceptable
	for _, id := range invalidIDs.Values() {
		validIDs.Remove(id)
	}

	return validIDs
}
