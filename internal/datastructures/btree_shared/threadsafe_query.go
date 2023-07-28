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
	} else {
		// need to recurse through everything
		tsat.query(selection, onFind)
	}

	// only run the query if we have found items
	if len(items) != 0 {
		onFindSelection(items)
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
					valuesNode.values.Find(*value.Value, addID(newSet))
				})

				invalidIDs = append(invalidIDs, newSet)
			case query.LessThan():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.FindLessThan(*value.Value, addBulkIDs(newSet))
				})

				validIDs = append(validIDs, newSet)
			case query.LessThanOrEqual():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.FindLessThanOrEqual(*value.Value, addBulkIDs(newSet))
				})

				validIDs = append(validIDs, newSet)
			case query.GreaterThan():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.FindGreaterThan(*value.Value, addBulkIDs(newSet))
				})

				validIDs = append(validIDs, newSet)
			case query.GreaterThanOrEqual():
				inclusive = true

				// only need to match one key
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					// only need to match the one value
					valuesNode.values.FindGreaterThanOrEqual(*value.Value, addBulkIDs(newSet))
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

	// add all possible values to the end users callback
	for _, validID := range finalIds.Values() {
		if item := tsat.ids.Get(validID); item != nil {
			onFind(item)
		}
	}
}
