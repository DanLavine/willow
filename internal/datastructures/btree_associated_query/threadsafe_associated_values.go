package btreeassociatedquery

import "github.com/DanLavine/willow/pkg/models/datatypes"

const (
	ReservedID = "_associated_id"
)

// AssociatedKeyValues reocrds all the key values that were saved to create the associated grouping.
type AssociatedKeyValueQuery struct {
	// record the original query that was used to setup the tree
	query datatypes.AssociatedKeyValuesQuery

	// data saved by the client
	value any
}
