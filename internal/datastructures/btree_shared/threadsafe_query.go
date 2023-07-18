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
	validIds := set.New[uint64]()
	invalidIds := set.New[uint64]()
	intersectIds := set.New[uint64]()

	// intersect any possible IDs for the query
	intersectValidIDs := func(query *query.Query) func(item any) {
		return func(item any) {
			idNode := item.(*threadsafeIDNode)

			for index, ids := range idNode.ids {
				if query.Limits != nil {
					if index >= *query.Limits.NumberOfKeys {
						// when limiting number of keys. only procede up to the key limit
						return
					}
				}

				intersectIds.AddBulk(ids)
			}
		}
	}

	// add any possible IDs for the query
	addValidIDs := func(query *query.Query) func(item any) {
		return func(item any) {
			idNode := item.(*threadsafeIDNode)

			for index, ids := range idNode.ids {
				if query.Limits != nil {
					if index >= *query.Limits.NumberOfKeys {
						// when limiting number of keys. only procede up to the key limit
						return
					}
				}

				validIds.AddBulk(ids)
			}
		}
	}

	// add any possible IDs that need to be stripped from the query
	addInvalidIDs := func(query *query.Query) func(item any) {
		return func(item any) {
			idNode := item.(*threadsafeIDNode)

			for index, ids := range idNode.ids {
				if query.Limits != nil {
					if index >= *query.Limits.NumberOfKeys {
						// when limiting number of keys. only procede up to the key limit
						return
					}
				}

				invalidIds.AddBulk(ids)
			}
		}
	}

	first := true
	inclusive := false

	// parse the where clause
	if selection.Where != nil {
		where := selection.Where

		for key, value := range where.KeyValues {
			if value.Exists != nil {
				// existence check
				switch *value.Exists {
				case true:
					switch first {
					case true:
						// add all values for the
						tsat.keys.Find(datatypes.String(key), func(item any) {
							valuesNode := item.(*threadsafeValuesNode)
							valuesNode.values.Iterate(addValidIDs(where))
						})
					default:
						// intersect all values for the desired key
						tsat.keys.Find(datatypes.String(key), func(item any) {
							valuesNode := item.(*threadsafeValuesNode)
							valuesNode.values.Iterate(intersectValidIDs(where))
						})

						// keep only the shared intersection of all values
						validIds.Intersection(intersectIds.Values())
						intersectIds.Clear()
					}

					inclusive = true
				case false:
					// record the invalid IDs that need to be stripped after all finds
					tsat.keys.Find(datatypes.String(key), func(item any) {
						valuesNode := item.(*threadsafeValuesNode)
						valuesNode.values.Iterate(addInvalidIDs(where))
					})
				}
			} else {
				inclusive = true

				// check an assigned value
				tsat.keys.Find(datatypes.String(key), func(item any) {
					valuesNode := item.(*threadsafeValuesNode)

					switch value.ValueComparison {
					case query.Equals():
						switch first {
						case true:
							valuesNode.values.Find(*value.Value, addValidIDs(where))
						default:
							valuesNode.values.Find(*value.Value, intersectValidIDs(where))

							// keep only the shared intersection of all values
							validIds.Intersection(intersectIds.Values())
							intersectIds.Clear()
						}
					case query.NotEquals():
					case query.LessThan():
					case query.LessThanOrEqual():
					case query.GreaterThan():
					case query.GreaterThanOrEqual():
					}
				})
			}

			first = false
		}

		// Special case where the only query is a FALSE check. So need to also find all other IDs
		if !inclusive {
			tsat.keys.Iterate(func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Iterate(addValidIDs(where))
			})
		}
	}

	// remove all invalid IDs from the set
	for _, invalidID := range invalidIds.Values() {
		validIds.Remove(invalidID)
	}

	// add all possible values to the end users callback
	for _, validID := range validIds.Values() {
		if item := tsat.ids.Get(validID); item != nil {
			onFind(item)
		}
	}
}
