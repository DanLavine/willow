package btreeonetomany

import "github.com/DanLavine/willow/pkg/models/datatypes"

type OneToManyItem interface {
	// return the original value for the item saved in the tree
	Value() any

	// OneID returns the One Relation the paginated item belongs to
	OneID() string

	// ManyID returns the ID for the oobject saved in the Many tree. To quickly access this particualr
	// item again, this can be used as the _associated_id in a query
	ManyID() string

	// returns the KeyValues that define the Many Relationship
	ManyKeyValues() datatypes.KeyValues
}

type oneToManyItem struct {
	value any

	oneID string

	manyID string

	manyKeyValues datatypes.KeyValues
}

func (otmi *oneToManyItem) Value() any {
	return otmi.value
}

func (otmi *oneToManyItem) OneID() string {
	return otmi.oneID
}

func (otmi *oneToManyItem) ManyID() string {
	return otmi.manyID
}

func (otmi *oneToManyItem) ManyKeyValues() datatypes.KeyValues {
	newKeyValues := datatypes.KeyValues{}
	for key, value := range otmi.manyKeyValues {
		newKeyValues[key] = value
	}

	return newKeyValues
}
