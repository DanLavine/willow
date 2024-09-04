package v1

import (
	"fmt"
	"regexp"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var requiredUniqueKeys = []string{"name"}
var queueNameRexexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type Queue struct {
	// Defintion of the Queue's unique tags
	//
	// These are constructed with a Single Key + Value that must be in the form of
	// `{"name": {"Type": 13, "Data": "[queue name]"}}
	// This `name` parameter is the name of the queue that can be used in the API urls and having a single value
	// ensures that each queue is unique as the groupings of these key values define the Queue's uniqueness
	Definition datatypes.AnyKeyValues `json:"Definition"`

	// Specification fields define the object details and how it is saved in the DB
	Spec *QueueSpec `json:"Spec"`

	// State fields defined the usage details for the item saved in the DB
	State *QueueState `json:"State,omitempty"`
}

func (queue Queue) validateRequiredParameters() *errors.ModelError {
	// validate the Definition's keys
	if err := queue.Definition.Validate(datatypes.T_string, datatypes.T_string); err != nil {
		return &errors.ModelError{Field: "Definition", Child: err}
	}
	// ensure the defintion has the required unique keys
	for _, requiredUniqueKey := range requiredUniqueKeys {
		if name, ok := queue.Definition[requiredUniqueKey]; ok {
			if !queueNameRexexp.MatchString(name.Data.(string)) {
				return &errors.ModelError{Field: "Definition[name].Data", Err: fmt.Errorf("does not match the allowable characters regexp '^[a-zA-Z0-9_]+$'")}
			}
		} else {
			return &errors.ModelError{Field: "Definition", Err: fmt.Errorf("missing the required key '%s'", requiredUniqueKey)}
		}
	}

	// validate all the spec operations
	if err := queue.ValidateSpec(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (queue Queue) Validate() *errors.ModelError {
	if err := queue.validateRequiredParameters(); err != nil {
		return err
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

func (queue Queue) ValidateUpsert() *errors.ModelError {
	if err := queue.validateRequiredParameters(); err != nil {
		return err
	}

	// extra check to ensure there is only one key, which is 'name'
	if len(queue.Definition) != 1 {
		return &errors.ModelError{Field: "Definition", Err: fmt.Errorf("must have a length of 1")}
	}

	if queue.State != nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("must be empty on a create request")}
	}

	return nil
}

func (queue Queue) ValidateSpec() *errors.ModelError {
	if queue.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty spcification")}
	} else {
		if err := queue.Spec.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type QueueSpec struct {
	// (Optional) Metadata can be used to save any non-unique relevant information got the particular queue
	Metadata datatypes.AnyKeyValues `json:"Metadata,omitempty"`

	// MaxItems defines how many items can ueued and running items combined, -1 means unlimited
	MaxItems *int64 `json:"MaxItems,omitempty"`
}

func (queueSpec *QueueSpec) Validate() *errors.ModelError {
	if queueSpec.Metadata != nil {
		if err := queueSpec.Metadata.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
			return &errors.ModelError{Field: "Metadata", Child: err}
		}
	}

	if queueSpec.MaxItems == nil {
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
