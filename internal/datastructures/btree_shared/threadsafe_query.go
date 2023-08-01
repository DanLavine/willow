package btreeshared

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

func (tsat *threadsafeAssociatedTree) Query(selection query.Select, onFindSelection datastructures.OnFindSelection) error {
	if err := selection.Validate(); err != nil {
		return err
	}

	items := []any{}
	onFind := func(item any) {
		items = append(items, item)
	}

	if onFindSelection == nil {
		return fmt.Errorf("onFindSelection cannot be nil")
	}

	if selection.Where == nil && selection.And == nil && selection.Or == nil {
		// select all
		tsat.ids.Iterate(onFind)

		// only run the query if we have found items
		if len(items) != 0 {
			onFindSelection(items)
		}
	} else {
		// need to recurse through everything
		tsat.query(selection, onFind)

		// add all possible values to the end users callback
		finalItems := []any{}
		for _, id := range items {
			if item := tsat.ids.Get(id.(uint64)); item != nil {
				finalItems = append(finalItems, item)
			}
		}

		// only run the query if we have found items
		if len(finalItems) != 0 {
			onFindSelection(finalItems)
		}
	}

	return nil
}

func (tsat *threadsafeAssociatedTree) query(selection query.Select, onFind datastructures.OnFind) {
	validIDs := []set.Set[uint64]{}
	invalidIDs := []set.Set[uint64]{}

	// parse the where clause
	inclusive := false
	if selection.Where != nil {
		where := selection.Where

		addID := func(newSet set.Set[uint64]) func(item any) {
			return func(item any) {
				idNode := item.(*threadsafeIDNode)

				for index, ids := range idNode.ids {
					// break if we reach a key count that has more than the requested length
					if where.Limits != nil {
						if index >= *where.Limits.NumberOfKeys {
							return
						}
					}

					// include all the ids
					newSet.AddBulk(ids)
				}
			}
		}

		addBulkIDs := func(newSet set.Set[uint64]) func(item []any) {
			return func(items []any) {
				for _, item := range items {
					idNode := item.(*threadsafeIDNode)

					for index, ids := range idNode.ids {
						// break if we reach a key count that has more than the requested length
						if where.Limits != nil {
							if index >= *where.Limits.NumberOfKeys {
								return
							}
						}

						// include all the ids
						newSet.AddBulk(ids)
					}
				}
			}
		}

		for key, value := range where.KeyValues {
			newSet := set.New[uint64]()

			// existence check
			if value.Exists != nil {
				switch *value.Exists {
				case true:
					inclusive = true

					tsat.keys.Find(datatypes.String(key), func(item any) {
						valuesNode := item.(*threadsafeValuesNode)

						if value.ExistsType != nil {
							// add all values for the  desired type
							valuesNode.values.IterateMatchType(*value.ExistsType, addID(newSet))
						} else {
							// add all values
							valuesNode.values.Iterate(addID(newSet))
						}
					})

					validIDs = append(validIDs, newSet)
				case false:
					tsat.keys.Find(datatypes.String(key), func(item any) {
						valuesNode := item.(*threadsafeValuesNode)

						if value.ExistsType != nil {
							valuesNode.values.IterateMatchType(*value.ExistsType, addID(newSet))
						} else {
							valuesNode.values.Iterate(addID(newSet))
						}
					})

					invalidIDs = append(invalidIDs, newSet)
				}

				continue
			}

			// comparison check
			switch value.ValueComparison {
			case query.Equals():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.Find(*value.Value, addID(newSet))
				})

				validIDs = append(validIDs, newSet)
			case query.NotEquals():
				// find the key to strip all values
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// find the value to remove everything
					if value.ValueTypeMatch != nil && *value.ValueTypeMatch == true {
						inclusive = true
						valuesNode.values.FindNotEqualMatchType(*value.Value, addBulkIDs(newSet))
						validIDs = append(validIDs, newSet)
					} else {
						valuesNode.values.Find(*value.Value, addID(newSet))
						invalidIDs = append(invalidIDs, newSet)
					}
				})
			case query.LessThan():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch == true {
						valuesNode.values.FindLessThanMatchType(*value.Value, addBulkIDs(newSet))
					} else {
						valuesNode.values.FindLessThan(*value.Value, addBulkIDs(newSet))
					}
				})

				validIDs = append(validIDs, newSet)
			case query.LessThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch == true {
						valuesNode.values.FindLessThanOrEqualMatchType(*value.Value, addBulkIDs(newSet))
					} else {
						valuesNode.values.FindLessThanOrEqual(*value.Value, addBulkIDs(newSet))
					}
				})

				validIDs = append(validIDs, newSet)
			case query.GreaterThan():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch == true {
						valuesNode.values.FindGreaterThanMatchType(*value.Value, addBulkIDs(newSet))
					} else {
						valuesNode.values.FindGreaterThan(*value.Value, addBulkIDs(newSet))
					}
				})

				validIDs = append(validIDs, newSet)
			case query.GreaterThanOrEqual():
				inclusive = true

				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					if value.ValueTypeMatch != nil && *value.ValueTypeMatch == true {
						valuesNode.values.FindGreaterThanOrEqualMatchType(*value.Value, addBulkIDs(newSet))
					} else {
						valuesNode.values.FindGreaterThanOrEqual(*value.Value, addBulkIDs(newSet))
					}
				})

				validIDs = append(validIDs, newSet)
			}
		}

		// Special case where the only query is a FALSE check. So need to also find all other IDs
		if !inclusive {
			newSet := set.New[uint64]()

			tsat.keys.Iterate(func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Iterate(addID(newSet))
			})

			validIDs = append(validIDs, newSet)
		}
	}

	// save all the not IDs and join them?
	finalIds := set.New[uint64]()
	for index, includeIds := range validIDs {
		if index == 0 {
			finalIds.AddBulk(includeIds.Values())
		} else {
			finalIds.Intersection(includeIds.Values())
		}
	}

	// join exclude IDs
	finalExcludeIds := set.New[uint64]()
	for index, excludeIds := range invalidIDs {
		if index == 0 {
			finalExcludeIds.AddBulk(excludeIds.Values())
		} else {
			finalExcludeIds.Intersection(excludeIds.Values())
		}
	}

	// filter the IDs to a list of what is only acceptable
	for _, id := range finalExcludeIds.Values() {
		finalIds.Remove(id)
	}

	// intersect all the ANDs together
	for index, andSelection := range selection.And {
		// can exit early if we know that there are 0 ids. Don't need to search
		if finalIds.Size() == 0 {
			if selection.Where == nil && index == 0 {
				// nothing to do here. everything is joined in []And
			} else {
				// must not have any ids, so can exit early and stop selecting everything
				break
			}
		}

		andIDs := []uint64{}
		andOnFind := func(item any) {
			andIDs = append(andIDs, item.(uint64))
		}

		tsat.query(andSelection, andOnFind)

		if selection.Where == nil && index == 0 {
			finalIds.AddBulk(andIDs)
		} else {
			finalIds.Intersection(andIDs)
		}
	}

	// union all the ORs together
	for _, orSelection := range selection.Or {
		orIDs := []uint64{}
		orOnFind := func(item any) {
			orIDs = append(orIDs, item.(uint64))
		}

		tsat.query(orSelection, orOnFind)
		finalIds.AddBulk(orIDs)
	}

	for _, id := range finalIds.Values() {
		onFind(id)
	}
}
