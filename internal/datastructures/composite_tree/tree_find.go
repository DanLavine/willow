package compositetree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

// Find a value based on the Inclusive Query
func (ct *compositeTree) FindInclusive(inclusiveClause *v1.JoinedInclusiveWhereClause, onFind datastructures.OnFind) []any {
	var queryResults []any

	// find the single value with no key value pairs
	if inclusiveClause == nil {
		castableKeyValues := ct.compositeColumns.Find(datatypes.Int(0), KeyValuesReadLock)
		if castableKeyValues == nil {
			return nil
		}
		recordedKeyValues := castableKeyValues.(*KeyValues)
		defer recordedKeyValues.lock.RUnlock() // unlock here since the map of keyValues is in a random order
	}

	// TODO

	return queryResults
}
