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
	// How the object should be saved in the DB. This also drives the query and lookup apis
	DBDeifinition *dbdefinition.TypedKeyValues `json:"DBDefinition,omitempty"`

	// Timeout for the lock
	Timeout *time.Duration `json:"Timeout,omitempty"`
}

func (lockSpec *LockSpec) Validate() *errors.ModelError {
	if err := lockSpec.DBDeifinition.Validate(); err != nil {
		return &errors.ModelError{Field: "DBDeifinition", Child: err}
	}

	if lockSpec.Timeout != nil {
		if *lockSpec.Timeout == 0 {
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
