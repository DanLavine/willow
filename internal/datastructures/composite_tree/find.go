package compositetree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type QueryResults []QueryResult

type QueryResult struct {
	keyValues map[datatypes.String]datatypes.String
	value     any
}

// Find a value based on the query
func (ct *compositeTree) Find(query v1.Query, onFind datastructures.OnFind) QueryResults {
	//ct.lock.RLock()
	//defer ct.lock.RUnlock()

	var queryResults QueryResults

	// TODO

	return queryResults
}
