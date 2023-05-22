package compositetree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

// Find a value based on the Inclusive Query
func (ct *compositeTree) FindInclusive(inclusiveClause *v1.JoinedInclusiveWhereClause, onFind datastructures.OnFind) []any {
	var queryResults []any

	// find the global value since there are no key value pairs
	if inclusiveClause == nil {
		castableGlobalValue := ct.compositeColumns.Find(datatypes.Int(0), onFind)
		if castableGlobalValue == nil {
			return nil
		}

		return append(queryResults, castableGlobalValue.(*globalValue).value)
	}

	// find the actual query we need
	whereClause := inclusiveClause.Where

	if whereClause.EqualsKeyValues != nil {
	} else if whereClause.MatchesKeyValues != nil {
	} else if whereClause.ContainsKeys != nil {
	}

}
