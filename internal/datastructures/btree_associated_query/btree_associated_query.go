package btreeassociatedquery

import (
	"fmt"
	"sort"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// bTreeAssociatedQuery is used to save a query in a tree and have arbitrary KeyValues match against those saved queries.
type BTreeAssociatedQuery interface {
	// Create an item in the associated tree.
	// Returns an error if
	// 1. if the query already exists when creating the associated item in the tree
	Create(query datatypes.AssociatedKeyValuesQuery, onCreate datastructures.OnCreate) (string, error)

	// CreateWithID an item in the associated tree.
	// Returns an error if
	// 1. the associatedID already exists
	// 2. if the query already exists
	CreateWithID(associatedID string, query datatypes.AssociatedKeyValuesQuery, onCreate datastructures.OnCreate) error

	// Create or Find an item in the association tree
	CreateOrFind(query datatypes.AssociatedKeyValuesQuery, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (string, error)

	// Serch for any number of items in the assoociation tree
	Match(keyValues datatypes.KeyValues, onFindPagination datastructures.OnFindPagination) error

	// Delete an item in the association tree
	Delete(query datatypes.AssociatedKeyValuesQuery, canDelete datastructures.CanDelete) error

	// Delete an item in the association tree by the AssociatedID
	DeleteByAssociatedID(associatedID string, canDelete datastructures.CanDelete) error

	// delete an number of items that match the KeyValues
	//DeleteByKeyValues(keyValues datatypes.KeyValues, canDelete datastructures.CanDelete) error
}

// DSL TODO: Open question. what would these look like? On doing the comparison of matching values to queries I don't think these
// can be used.
const (
	T_equals     datatypes.DataType = 100
	T_not_equals datatypes.DataType = 101
)

type queryValue struct {
	Type datatypes.DataType

	Data any
	//Or []insertableValues
}

func (qv queryValue) Validate() error {
	return nil
}

func (qv queryValue) DataType() datatypes.DataType {
	return qv.Type
}

func (qv queryValue) Value() any {
	return qv.Data
}

func (qv queryValue) Less(comparableObj datatypes.EncapsulatedData) bool {
	// know data type is less
	if qv.Type < comparableObj.DataType() {
		return true
	}

	// know data type is greater
	if qv.Type > comparableObj.DataType() {
		return false
	}

	switch qv.Type {
	case T_equals:
		// always return false here. On the check for equals, 2 false checks in either deirection means true
		return false
	case T_not_equals:
		// always return false here. On the check for equals, 2 false checks in either deirection means true
		return false
	default:
		panic(fmt.Errorf("unkonw type: %d", qv.Type))
	}
}

func (qv queryValue) LessType(comparableObj datatypes.EncapsulatedData) bool {
	return qv.Type < comparableObj.(queryValue).Type
}

func (qv queryValue) LessValue(comparableObj datatypes.EncapsulatedData) bool {
	return false
}

type insertableValues map[string]queryValue

// SortedKeys is needed to ensure all requests process keys in the same order
func (iv insertableValues) SortedKeys() []string {
	keys := []string{}

	for key, _ := range iv {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

type insertableKeyValues []insertableValues

func convertAssociatedKeyValuesQuery(query datatypes.AssociatedKeyValuesQuery) insertableKeyValues {
	newInserableKeyValues := insertableKeyValues{}

	// 1. Join all the values for the KeyValueSelection
	if query.KeyValueSelection != nil {
		newInsertableValues := insertableValues{}

		for key, value := range query.KeyValueSelection.KeyValues {
			var newQueryValue queryValue

			// TODO: check _associated_id

			if value.Exists != nil {
				switch *value.Exists {
				case true:
					newQueryValue = queryValue{Type: T_equals, Data: value}
				default: //false
					newQueryValue = queryValue{Type: T_not_equals, Data: value}
				}
			}

			newInsertableValues[key] = newQueryValue
		}

		newInserableKeyValues = append(newInserableKeyValues, newInsertableValues)
	}

	return newInserableKeyValues
}
