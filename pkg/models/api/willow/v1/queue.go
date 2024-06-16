package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

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
	// DBDefinition defines how to save the item in the Database
	DBDefinition *QueueDBDefinition `json:"DBDefinition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	//
	// This is an optional field as all the properties are currently optional
	Properties *QueueProperties `json:"Properties,omitempty"`
}

func (queueSpec *QueueSpec) Validate() *errors.ModelError {
	if queueSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := queueSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
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
	Name *string `json:"Name"`
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
	// Max size of the queue's eneueued and running items combined
	// -1 means unlimited
	MaxItems *int64 `json:"MaxItems,omitempty"`
}

func (queueProperties *QueueProperties) Validate() *errors.ModelError {
	if queueProperties.MaxItems == nil {
		return &errors.ModelError{Field: "MaxItems", Err: fmt.Errorf("recevied a null value")}
	}

	return nil
}

type QueueState struct {
	Deleting bool
}

func (queueState *QueueState) Validate() *errors.ModelError {
	return nil
}
