package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Override can be thought of as a "sub query" for a Rule's KeyValues. Any request that matches all the
// given tags for an override will use the new override value. If multiple overrides match a particular set of tags,
// then each override will be validated for their KeyValue group.
type Override struct {
	Spec *OverrideSpec `json:"Spec"`

	State *OverrideState `json:"State,omitempty"`
}

func (override *Override) Validate() *errors.ModelError {
	if err := override.ValidateSpecOnly(); err != nil {
		return err
	}

	if override.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("received a null value")}
	} else {
		if err := override.State.Validate(); err != nil {
			return &errors.ModelError{Field: "State", Child: err}
		}
	}

	return nil
}

func (override *Override) ValidateSpecOnly() *errors.ModelError {
	if override.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received a null value")}
	} else {
		if err := override.Spec.Validate(); err != nil {
			return &errors.ModelError{Field: "Spec", Child: err}
		}
	}

	return nil
}

type OverrideSpec struct {
	// DBDefinition defines how to save the item in the Database
	DBDefinition *OverrideDBDefinition `json:"DBDefinition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	Properties *OverrideProperties `json:"Properties,omitempty"`
}

func (overrideSpec *OverrideSpec) Validate() *errors.ModelError {
	if overrideSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := overrideSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
	}

	if overrideSpec.Properties == nil {
		return &errors.ModelError{Field: "Properties", Err: fmt.Errorf("received a null value")}
	} else {
		if err := overrideSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

type OverrideDBDefinition struct {
	// The name of the override
	Name *string `json:"Name,omitempty"`

	// GroupByKeyValues match against the Counter's KeyValues to ensure that they are under the limit
	GroupByKeyValues datatypes.AnyKeyValues `json:"GroupByKeyValues,omitempty"`
}

func (overrideDBDefinition *OverrideDBDefinition) Validate() *errors.ModelError {
	if overrideDBDefinition.Name == nil {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("received a null value")}
	} else {
		if *overrideDBDefinition.Name == "" {
			return &errors.ModelError{Field: "Name", Err: fmt.Errorf("received an empty string")}
		}
	}

	if err := overrideDBDefinition.GroupByKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "GroupByKeyValues", Child: err}
	}

	return nil
}

type OverrideProperties struct {
	Limit *int64 `json:"Limit,omitempty"`
}

func (overrideProperties *OverrideProperties) Validate() *errors.ModelError {
	if overrideProperties.Limit == nil {
		return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("received a null value")}
	} else {
		if *overrideProperties.Limit < -1 {
			return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("is set below the minimum value of -1. Value must be [-1 (ulimited) | 0+ (zero or more specific limit)")}
		}
	}

	return nil
}

type OverrideState struct {
	Deleting bool `json:"Deleting"`
}

func (overrideState *OverrideState) Validate() *errors.ModelError {
	return nil
}
