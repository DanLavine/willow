package btreeassociated

import "github.com/DanLavine/willow/pkg/models/datatypes"

const (
	ReservedID = "_associated_id"
)

// AssociatedKeyValues reocrds all the key values that were saved to create the associated grouping.
type AssociatedKeyValues struct {
	// key value pairs that make up the item
	// need to change this now.
	keyValues KeyValues // This is the best I can come up with at this time.

	// data saved by the client
	value any
}

// get the AssociatedID for the item
func (associatedKeyValues *AssociatedKeyValues) AssociatedID() string {
	return associatedKeyValues.keyValues[datatypes.String(ReservedID)].Value().(string)
}

// return the list of key values in a copy so they cannot be changed by the caller
func (associatedKeyValues *AssociatedKeyValues) KeyValues() KeyValues {
	newKeyValues := KeyValues{}
	for key, value := range associatedKeyValues.keyValues {
		newKeyValues[key] = value
	}

	return newKeyValues
}

// get the Value that is saved in the AssociatedTree
func (associatedKeyValues *AssociatedKeyValues) Value() any {
	return associatedKeyValues.value
}
