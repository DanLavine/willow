package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
)

type Item struct {
	// Specification fields define the object details and how it is saved in the DB
	Spec *ItemSpec `json:"Spec,omitempty"`

	// State fields defined the usage details for the item saved in the DB
	State *ItemState `json:"State,omitempty"`
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (item Item) Validate() *errors.ModelError {
	if item.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty spcification")}
	} else {
		if err := item.Spec.Validate(); err != nil {
			return err
		}
	}

	if item.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("received an empty state")}
	} else {
		if err := item.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (item Item) ValidateSpecOnly() *errors.ModelError {
	if item.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty specification")}
	} else {
		if err := item.Spec.Validate(); err != nil {
			return &errors.ModelError{Field: "Spec", Child: err}
		}
	}

	if item.State != nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("must be null")}
	}

	return nil
}

type ItemSpec struct {
	// DBDefinition defines how to save the item in the Database
	DBDefinition *ItemDBDefinition `json:"DBDefinition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	Properties *ItemProperties `json:"Properties,omitempty"`
}

func (itemSpec *ItemSpec) Validate() *errors.ModelError {
	if itemSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := itemSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
	}

	if itemSpec.Properties == nil {
		return &errors.ModelError{Field: "Properties", Err: fmt.Errorf("received a null value")}
	} else {
		if err := itemSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

type ItemDBDefinition struct {
	KeyValues dbdefinition.TypedKeyValues `json:"KeyValues,omitempty"`
}

func (itemDBDefinition *ItemDBDefinition) Validate() *errors.ModelError {
	if err := itemDBDefinition.KeyValues.Validate(); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}

type ItemProperties struct {
	// Raw data that the end user clients know how to parse
	Data []byte

	// If the item can be updated on another request
	Updateable *bool `json:"Updateable,omitempty"`

	// How many attempts to retry in the case of a failure
	RetryAttempts *uint64 `json:"RetryAttempts,omitempty"`

	// Where to enqueue the item on a failed attempt
	RetryPosition *string `json:"RetryPosition,omitempty"`

	// How long to wait for heartbeats untill the item is considered failed
	TimeoutDuration *time.Duration `json:"TimeoutDuration,omitempty"`
}

func (itemProperties *ItemProperties) Validate() *errors.ModelError {
	if len(itemProperties.Data) == 0 {
		return &errors.ModelError{Field: "Data", Err: fmt.Errorf("received a value of 0 bytes")}
	}

	if itemProperties.Updateable == nil {
		return &errors.ModelError{Field: "Updateable", Err: fmt.Errorf("received a null value")}
	}

	if itemProperties.RetryAttempts == nil {
		return &errors.ModelError{Field: "RetryAttempts", Err: fmt.Errorf("received a null value")}
	}

	if itemProperties.RetryPosition == nil {
		return &errors.ModelError{Field: "RetryPosition", Err: fmt.Errorf("received a null value")}
	} else {
		switch *itemProperties.RetryPosition {
		case "front", "back":
			// these are fine
		default:
			return &errors.ModelError{Field: "RetryPosition", Err: fmt.Errorf("must be either [front | back], but received '%s'", *itemProperties.RetryPosition)}
		}
	}

	if itemProperties.TimeoutDuration == nil {
		return &errors.ModelError{Field: "TimeoutDuration", Err: fmt.Errorf("received a null value")}
	}

	return nil
}

type ItemState struct {
	// ID of the item that needs to be heartbeat and acked
	ID string `json:"ID"`
}

func (itemState *ItemState) Validate() *errors.ModelError {
	if itemState.ID == "" {
		return &errors.ModelError{Field: "ID", Err: fmt.Errorf("is the empty string")}
	}

	return nil
}
