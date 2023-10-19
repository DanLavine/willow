package btreeassociated

import "github.com/DanLavine/willow/pkg/models/datatypes"

const (
	ReservedID = "_associated_id"
)

// AssociatedKeyValues reocrds all the key values that were saved to create the associated grouping.
type AssociatedKeyValues struct {
	// key value pairs that make up the item
	keyValues datatypes.StringMap // This is the best I can come up with at this time.

	// data saved by the client
	value any
}

func (associatedKeyValues *AssociatedKeyValues) ID() string {
	return associatedKeyValues.keyValues[ReservedID].Value.(string)
}

func (associatedKeyValues *AssociatedKeyValues) KeyValues() datatypes.StringMap {
	return associatedKeyValues.keyValues
}

func (associatedKeyValues *AssociatedKeyValues) Value() any {
	return associatedKeyValues.value
}