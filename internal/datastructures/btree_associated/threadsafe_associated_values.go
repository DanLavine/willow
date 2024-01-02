package btreeassociated

import "github.com/DanLavine/willow/pkg/models/datatypes"

const (
	ReservedID = "_associated_id"
)

type AssociatedKeyValues interface {
	// obtain the original value for the item in the tree
	Value() any

	// key values that define the item saved in the tree
	KeyValues() datatypes.KeyValues

	// associated id for the item saved in the tree
	AssociatedID() string
}

// associatedKeyValues reocrds all the key values that were saved to create the associated grouping.
type associatedKeyValues struct {
	// key value pairs that make up the item
	// need to change this now.
	keyValues datatypes.KeyValues // This is the best I can come up with at this time.

	// data saved by the client
	value any
}

// get the Value that is saved in the AssociatedTree
func (associatedKeyValues *associatedKeyValues) Value() any {
	return associatedKeyValues.value
}

// return the list of key values in a copy so they cannot be changed by the caller
func (associatedKeyValues *associatedKeyValues) KeyValues() datatypes.KeyValues {
	newKeyValues := datatypes.KeyValues{}
	for key, value := range associatedKeyValues.keyValues {
		if key != ReservedID {
			newKeyValues[key] = value
		}
	}

	return newKeyValues
}

// get the AssociatedID for the item
func (associatedKeyValues *associatedKeyValues) AssociatedID() string {
	return associatedKeyValues.keyValues[ReservedID].Value().(string)
}
