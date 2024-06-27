package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var requiredUniqueKeys = []string{"name"}

type Queue struct {
	// Specification fields define the object details and how it is saved in the DB
	Spec *QueueSpec `json:"Spec"`

	// State fields defined the usage details for the item saved in the DB
	State *QueueState `json:"State,omitempty"`
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (queue Queue) Validate() *errors.ModelError {
	if queue.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty spcification")}
	} else {
		if err := queue.Spec.Validate(); err != nil {
			return err
		}
	}

	if queue.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("received an empty state")}
	} else {
		if err := queue.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (queue *Queue) ValidateSpecOnly() *errors.ModelError {
	if queue.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty specification")}
	} else {
		if err := queue.Spec.Validate(); err != nil {
			return &errors.ModelError{Field: "Spec", Child: err}
		}
	}

	if queue.State != nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("must be null")}
	}

	return nil
}

type QueueSpec struct {
	// Defintion of the Queue's unique tags
	Definition v1common.AnyKeyValues `json:"Definition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	Properties *QueueProperties `json:"Properties,omitempty"`
}

func (queueSpec *QueueSpec) Validate() *errors.ModelError {
	// validate the Definition's keys
	if err := queueSpec.Definition.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "Definition", Child: err}
	}

	// ensure the defintion has the required unique keys
	for _, requiredUniqueKey := range requiredUniqueKeys {
		if value, ok := queueSpec.Definition[requiredUniqueKey]; ok {
			if !value.Unique {
				return &errors.ModelError{Field: fmt.Sprintf("Definition[%s].Unique", requiredUniqueKey), Err: fmt.Errorf("expected the value to be true, received false ")}
			}
		} else {
			return &errors.ModelError{Field: "Definition", Err: fmt.Errorf("missing the required key '%s'", requiredUniqueKey)}
		}
	}

	if queueSpec.Properties != nil {
		if err := queueSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

type QueueDBDefinition struct {
	// Name of the specific queue
	Name *string `json:"Name,omitempty"`
}

func (queueDBDefinition *QueueDBDefinition) Validate() *errors.ModelError {
	if queueDBDefinition.Name == nil {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("recevied a null value")}
	}

	if *queueDBDefinition.Name == "" {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("recevied an empty string")}
	}

	return nil
}

type QueueProperties struct {
	// Metadta to record any arbitrary information the end user might find helpful
	Metadata datatypes.AnyKeyValues `json:"Metadata,omitempty"`

	// Max size of the queue's eneueued and running items combined, -1 means unlimited
	MaxItems *int64 `json:"MaxItems,omitempty"`
}

func (queueProperties *QueueProperties) Validate() *errors.ModelError {
	if queueProperties.MaxItems == nil {
		return &errors.ModelError{Field: "MaxItems", Err: fmt.Errorf("recevied a null value")}
	}

	return nil
}

type QueueState struct {
	// Enqueued is the total number of items enqueued, but not running on the service
	Enqueued *uint64 `json:"Enqueued,omitempty"`

	// How many items across all the queues are running
	Running *uint64 `json:"Running,omitempty"`

	// Record of all the channels being destroyed
	// DestroyingChannels []*v1common.AssociatedActionQuery
}

func (queueState *QueueState) Validate() *errors.ModelError {
	return nil
}
