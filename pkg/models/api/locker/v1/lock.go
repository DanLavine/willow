package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
)

// Lock is the full representation of the Lock Object
type Lock struct {
	// Specification fields define the object details and how it is saved in the DB
	Spec *LockSpec `json:"Spec"`

	// State fields defined the usage details for the item saved in the DB
	State *LockState `json:"State,omitempty"`
}

// SetDefaultProperties is currently an "easy" way to set defaults for optional values from when the request is parsed.
// Otherwise, I think I would want to change everything from pointers to proper values after the Spec and State? But
// I think the Delete operation will become harder? Nother there atm for handling persitence and actual long running delete
// operations to test the patterns I like. Also This allows for properties to have a "zero/unset" value for different behaviors
func (lock *Lock) SetDefaultProperties(lockProperties *LockProperties) {
	if lock.Spec.Properties == nil {
		lock.Spec.Properties = lockProperties
		return
	}

	if lock.Spec.Properties.Timeout == nil {
		lock.Spec.Properties.Timeout = lockProperties.Timeout
	}
}

// Validate the entire Lock's Spec and State fields
func (lock *Lock) Validate() *errors.ModelError {
	if lock.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("recieved an empty spcification")}
	} else {
		if err := lock.Spec.Validate(); err != nil {
			return err
		}
	}

	if lock.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("recieved an empty state")}
	} else {
		if err := lock.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate the Lock's Spec fields
func (lock *Lock) ValidateSpecOnly() *errors.ModelError {
	if err := lock.Spec.Validate(); err != nil {
		return &errors.ModelError{Field: "Spec", Child: err}
	}

	if lock.State != nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("must be empty")}
	}

	return nil
}

// LockSpec defines the specific specifications for the Lock
type LockSpec struct {
	// DBDefinition defines how to save the item in the Database
	DBDefinition *LockDBDefinition `json:"DBDefinition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	//
	// This is an optional field as all the properties are currently optional
	Properties *LockProperties `json:"Properties,omitempty"`
}

func (lockSpec *LockSpec) Validate() *errors.ModelError {
	if lockSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := lockSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
	}

	if lockSpec.Properties != nil {
		if err := lockSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

type LockDBDefinition struct {
	KeyValues dbdefinition.TypedKeyValues `json:"KeyValues,omitempty"`
}

func (LockDBDefinition *LockDBDefinition) Validate() *errors.ModelError {
	if err := LockDBDefinition.KeyValues.Validate(); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}

type LockProperties struct {
	// Timeout for the lock
	//
	// Optional. If this is not set, then the default configured on the service will be used
	Timeout *time.Duration `json:"Timeout,omitempty"`
}

func (lockProperties *LockProperties) Validate() *errors.ModelError {
	if lockProperties.Timeout != nil {
		if *lockProperties.Timeout == 0 {
			return &errors.ModelError{Field: "Timeout", Err: fmt.Errorf("recieved a timeout duration of 0. Requires a valid time duration represented as int64 nanosecond")}
		}
	}

	return nil
}

// LockStatus defines the the current state of the object's usage in the DB and are ReadOnly operations
type LockState struct {
	// LockID is the unique identifier for the lock that can be used for quick API access
	LockID string

	// SessionID associated with the client currently holding the lock. Used to heartbeat or release a lock
	SessionID string

	// Time until the lock expires if no heartbeats are received
	TimeTillExipre time.Duration

	// LocksHeldOrWaiting show how many clients are all trying to obtain the same lock
	LocksHeldOrWaiting uint64
}

func (LockState *LockState) Validate() *errors.ModelError {
	if LockState.LockID == "" {
		return &errors.ModelError{Field: "LockID", Err: fmt.Errorf("received an empty string")}
	}

	if LockState.SessionID == "" {
		return &errors.ModelError{Field: "SessionID", Err: fmt.Errorf("received an empty string")}
	}

	if LockState.LocksHeldOrWaiting == 0 {
		return &errors.ModelError{Field: "LocksHeldOrWaiting", Err: fmt.Errorf("received a value of 0")}
	}

	return nil
}
