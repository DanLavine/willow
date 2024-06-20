package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Counter is the full api model that is returned as part of the query operations
type Counter struct {
	Spec *CounterSpec `json:"Spec,omitempty"`

	State *CounterState `json:"State,omitempty"`
}

func (counter *Counter) Validate() *errors.ModelError {
	if err := counter.ValidateSpecOnly(); err != nil {
		return err
	}

	if counter.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("received a null value")}
	} else {
		if err := counter.State.Validate(); err != nil {
			return &errors.ModelError{Field: "State", Child: err}
		}
	}

	return nil
}

func (counter *Counter) ValidateSpecOnly() *errors.ModelError {
	if counter.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received a null value")}
	} else {
		if err := counter.Spec.Validate(); err != nil {
			return &errors.ModelError{Field: "Spec", Child: err}
		}
	}

	return nil
}

type CounterSpec struct {
	DBDefinition *CounterDBDefinition `json:"DBDefinition"`

	Properties *CounteProperties `json:"Properties"`
}

func (counterSpec *CounterSpec) Validate() *errors.ModelError {
	if counterSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := counterSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
	}

	if counterSpec.Properties == nil {
		return &errors.ModelError{Field: "Properties", Err: fmt.Errorf("received a null value")}
	} else {
		if err := counterSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

type CounterDBDefinition struct {
	KeyValues datatypes.TypedKeyValues `json:"KeyValues"`
}

func (counterDBDefinition *CounterDBDefinition) Validate() *errors.ModelError {
	if err := counterDBDefinition.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}

type CounteProperties struct {
	// Total number of counters for the particular KeyValues
	Counters *int64 `json:"Counters,omitempty"`
}

func (counteProperties *CounteProperties) Validate() *errors.ModelError {
	if counteProperties.Counters == nil {
		return &errors.ModelError{Field: "Counters", Err: fmt.Errorf("received a null value")}
	}

	return nil
}

type CounterState struct {
	Deleting bool `json:"Deleting"`
}

func (counterState *CounterState) Validate() *errors.ModelError {
	return nil
}
