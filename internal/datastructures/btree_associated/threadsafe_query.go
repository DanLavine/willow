package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

// This is totally broken atm!
// needs a refactor after the pagination change rather than bulk

// Query can be used to find a single or collection of items that match specific criteria. This
// is thread safe to call with any of the other functions ono this object
//
// PARAMS:
// - selection - the selection for the specified items to find. See the Select docs for this param specifically
// - onFindPagination - is the callback used for an items found in the tree. It will recive the objects' value saved in the tree (what were originally provided)
//
// RETURNS:
// - error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) Query(selection query.Select, onFindPagination datastructures.OnFindPagination) error {
	if err := selection.Validate(); err != nil {
		return err
	}
	if onFindPagination == nil {
		return fmt.Errorf("onFindPagination cannot be nil")
	}

	if selection.Where == nil && selection.And == nil && selection.Or == nil {
		// select all
		tsat.ids.Iterate(onFindPagination)
	} else {
		// need to recurse through everything

		var validIDs set.Set[string]
		validCounter := 0

		// find the first where
		if selection.Where != nil {
			validIDs = tsat.queryWhere(selection)
			validCounter++
		}

		// intersect all the ANDs together
		for _, andSelection := range selection.And {
			// can exit early if we know that there are 0 ids. Don't need to search
			if validCounter >= 1 && validIDs.Size() == 0 {
				// must not have any ids, so can exit early and stop selecting everything
				break
			}

			subsetValidIDs := tsat.queryAnd(andSelection)

			switch validCounter {
			case 0:
				validIDs = subsetValidIDs
				validCounter++
			default:
				validIDs.Intersection(subsetValidIDs.Values())
			}
		}

		// have first set of valid IDs to lookup
		shouldContinue := true
		if validCounter >= 1 {
			for _, id := range validIDs.Values() {
				tsat.ids.Find(datatypes.String(id), func(item any) {
					shouldContinue = onFindPagination(item)
				})

				// break querying more items
				if !shouldContinue {
					return nil
				}
			}
		}

		// check the OR cases as well
		for _, orSelection := range selection.Or {
			subsetValidIDs := tsat.queryOr(orSelection)

			// Need to interset our OR's with the current values
			switch validCounter {
			case 0:
				validIDs = subsetValidIDs
				validCounter++
			default:
				validIDs.Intersection(subsetValidIDs.Values())

				// remve the values we already checked
				for _, alreadyFoundID := range validIDs.Values() {
					subsetValidIDs.Remove(alreadyFoundID)
				}

				// add the subset to know values that we are about to check
				validIDs.AddBulk(subsetValidIDs.Values())
			}

			for _, id := range subsetValidIDs.Values() {
				tsat.ids.Find(datatypes.String(id), func(item any) {
					shouldContinue = onFindPagination(item)
				})

				// break querying more items
				if !shouldContinue {
					return nil
				}
			}
		}
	}

	return nil
}

func (tsat *threadsafeAssociatedTree) queryAnd(selection query.Select) set.Set[string] {
	var validIDs set.Set[string]
	validCounter := 0

	// find the first where
	if selection.Where != nil {
		validIDs = tsat.queryWhere(selection)
		validCounter++
	}

	// intersect all the ANDs together
	for _, andSelection := range selection.And {
		// can exit early if we know that there are 0 ids. Don't need to search
		if validCounter >= 1 && validIDs.Size() == 0 {
			// must not have any ids, so can exit early and stop selecting everything
			break
		}

		subsetValidIDs := tsat.queryAnd(andSelection)
		switch validCounter {
		case 0:
			validIDs = subsetValidIDs
			validCounter++
		default:
			validIDs.Intersection(subsetValidIDs.Values())
		}
	}

	// add all the OR IDs that we find as well
	for _, orSelection := range selection.Or {
		validIDs.AddBulk(tsat.queryOr(orSelection).Values())
	}

	return validIDs
}

func (tsat *threadsafeAssociatedTree) queryOr(selection query.Select) set.Set[string] {
	var validIDs set.Set[string]
	validCounter := 0

	// find the first where
	if selection.Where != nil {
		validIDs = tsat.queryWhere(selection)
		validCounter++
	}

	// intersect all the ANDs together
	for _, andSelection := range selection.And {
		// can exit early if we know that there are 0 ids. Don't need to search
		if validCounter >= 1 && validIDs.Size() == 0 {
			// must not have any ids, so can exit early and stop selecting everything
			break
		}

		subsetValidIDs := tsat.queryAnd(andSelection)
		switch validCounter {
		case 0:
			validIDs = subsetValidIDs
			validCounter++
		default:
			validIDs.Intersection(subsetValidIDs.Values())
		}
	}

	// add all the OR IDs that we find as well
	for _, orSelection := range selection.Or {
		validIDs.AddBulk(tsat.queryOr(orSelection).Values())
	}

	return validIDs
}

func (tsat *threadsafeAssociatedTree) queryWhere(selection query.Select) set.Set[string] {
	validCounter := 0
	validIDs := set.New[string]()

	invalidCounter := 0
	invalidIDs := set.New[string]()

	// parse the where clause
	inclusive := false
	if selection.Where != nil {
		where := selection.Where

		getAllIDs := func(allIDs *[]string) func(item any) bool {
			return func(item any) bool {
				// always casting to an ID node when finding possible indexes
				idNode := item.(*threadsafeIDNode)
				idNode.lock.RLock()
				defer idNode.lock.RUnlock()

				for index, ids := range idNode.ids {
					// break if we reach a key count that has more than the requested length
					if where.Limits != nil {
						if index >= *where.Limits.NumberOfKeys {
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
			// quick check to know if we hit a query where there are no possible valid IDs
			if validCounter >= 1 && validIDs.Size() == 0 {
				return false
			}

			// always casting to an ID node when finding possible indexes
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			var possibleIDs set.Set[string]
			for index, ids := range idNode.ids {
				// break if we reach a key count that has more than the requested length
				if where.Limits != nil {
					if index >= *where.Limits.NumberOfKeys {
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

			fmt.Println("intersecting ids:", possibleIDs)

			if possibleIDs != nil {
				validIDs.Intersection(possibleIDs.Values())
			}

			validCounter++
			return true
		}

		addInvalidIDs := func(item any) bool {
			// quick check to know if we hit a query where there are no possible invalid IDs
			if invalidCounter >= 1 && invalidIDs.Size() == 0 {
				return false
			}

			// always casting to an ID node when finding possible indexes
			idNode := item.(*threadsafeIDNode)
			idNode.lock.RLock()
			defer idNode.lock.RUnlock()

			var possibleInvalidIDs set.Set[string]
			for index, ids := range idNode.ids {
				// break if we reach a key count that has more than the requested length
				if where.Limits != nil {
					if index >= *where.Limits.NumberOfKeys {
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

			invalidCounter++
			return true
		}

		sortedKeys := where.SortedKeys()
		for _, key := range sortedKeys {
			value := where.KeyValues[key]
			allIDS := []string{}

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
							validCounter++
						default:
							validIDs.Intersection(allIDS)
						}
					})
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
							invalidCounter++
						default:
							invalidIDs.Intersection(allIDS)
						}
					})
				}

				continue
			}

			// comparison check
			switch *value.ValueComparison {
			case query.Equals():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.Find(*value.Value, func(item any) { addValidIDs(item) })
				})
			case query.NotEquals():
				// find the key to strip all values
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// find the value to remove everything
					if value.ValueTypeMatch != nil && *value.ValueTypeMatch {
						inclusive = true
						valuesNode.values.FindNotEqualMatchType(*value.Value, getAllIDs(&allIDS))

						switch validCounter {
						case 0:
							validIDs.AddBulk(allIDS)
							validCounter++
						default:
							validIDs.Intersection(allIDS)
						}
					} else {
						valuesNode.values.Find(*value.Value, func(item any) { addInvalidIDs(item) })
					}
				})
			case query.LessThan():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch {
						valuesNode.values.FindLessThanMatchType(*value.Value, getAllIDs(&allIDS))
					} else {
						valuesNode.values.FindLessThan(*value.Value, getAllIDs(&allIDS))
					}

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
						validCounter++
					default:
						validIDs.Intersection(allIDS)
					}
				})
			case query.LessThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch {
						valuesNode.values.FindLessThanOrEqualMatchType(*value.Value, getAllIDs(&allIDS))
					} else {
						valuesNode.values.FindLessThanOrEqual(*value.Value, getAllIDs(&allIDS))
					}

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
						validCounter++
					default:
						validIDs.Intersection(allIDS)
					}
				})
			case query.GreaterThan():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch {
						valuesNode.values.FindGreaterThanMatchType(*value.Value, getAllIDs(&allIDS))
					} else {
						valuesNode.values.FindGreaterThan(*value.Value, getAllIDs(&allIDS))
					}

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
						validCounter++
					default:
						validIDs.Intersection(allIDS)
					}
				})
			case query.GreaterThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch {
						valuesNode.values.FindGreaterThanOrEqualMatchType(*value.Value, getAllIDs(&allIDS))
					} else {
						valuesNode.values.FindGreaterThanOrEqual(*value.Value, getAllIDs(&allIDS))
					}

					switch validCounter {
					case 0:
						validIDs.AddBulk(allIDS)
						validCounter++
					default:
						validIDs.Intersection(allIDS)
					}
				})
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
