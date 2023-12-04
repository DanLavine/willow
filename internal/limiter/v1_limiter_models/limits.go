package v1limitermodels

import "github.com/DanLavine/willow/pkg/models/datatypes"

type Limits []Limit

type Limit struct {
	Name      string
	KeyValues datatypes.KeyValues
	Limit     uint64
}
