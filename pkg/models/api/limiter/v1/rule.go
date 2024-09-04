package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Rule saved in the DB that actions can be performed against
type Rule struct {
	Spec *RuleSpec `json:"Spec"`

	// State contains Read-Only details about the object's current state
	State *RuleState `json:"State,omitempty"`
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rule has all required fields set
func (rule *Rule) Validate() *errors.ModelError {
	if err := rule.ValidateSpecOnly(); err != nil {
		return err
	}

	if rule.State != nil {
		if err := rule.State.Validate(); err != nil {
			return &errors.ModelError{Field: "State", Child: err}
		}
	}

	return nil
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// ValidateSpecOnly ensures the Rule.Spec field has all required fields set
func (rule *Rule) ValidateSpecOnly() *errors.ModelError {
	if rule.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received a null value")}
	} else {
		if err := rule.Spec.Validate(); err != nil {
			return &errors.ModelError{Field: "Spec", Child: err}
		}
	}

	return nil
}

// RuleSpec defines the DB specification and all the object's properties
type RuleSpec struct {
	// Definition defines how the Rule will match against all Counters
	//
	// This can save and return additional things to the end user that are managed via the service?
	//	_associated_id [id]
	//	_created_at [timestamp]
	//	_deleted_at [timestamp]
	//
	// In the case of being a one-to-many relation (one side refrences the many side), we could then also have
	//	_one_to_many_[resource name]_id
	// Then in the case of many-to-one relation (many side referances one side), we could then have
	//	_many_to_one_[resource name]_id
	// I hope this could easly spin into many-to-many relations as well?
	//	_many_to_many_[resource name]_id
	//
	// Lastly, I think it would be nice to also have constraints on particular keys to make
	// them 'unique', even if they are not the generated name. This would be particularly useful
	//
	// Side note, should this always be called `Definition` as a way of knowing what a user ca	 n query?
	// seems like its nice for a common "API", but it isn't the best for "naming convention"
	Definition datatypes.AnyKeyValues `json:"Definition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	Properties *RuleProperties `json:"Properties,omitempty"`
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleSpec has all required fields set
func (ruleSpec *RuleSpec) Validate() *errors.ModelError {
	if err := ruleSpec.Definition.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "DBDefinition", Child: err}
	}

	if ruleSpec.Properties == nil {
		return &errors.ModelError{Field: "Properties", Err: fmt.Errorf("received a null value")}
	} else {
		if err := ruleSpec.Properties.Validate(); err != nil {
			return &errors.ModelError{Field: "Properties", Child: err}
		}
	}

	return nil
}

// RuleProperties are all the properties any actinable APIs will use. These fields are updateable
type RuleProperties struct {
	// Optional metadata to record with rule properties. This can be anything like a 'name' Key Value
	// pair to display in a web browser which make things easier for people to view
	Metadata datatypes.AnyKeyValues `json:"Metadata"`

	// Limit dictates what value of grouped counter KeyValues to allow untill a limit is reached.
	// Setting this value to -1 means unlimited
	//
	// Updateable
	Limit *int64 `json:"Limit,omitempty"`
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleProperties has all required fields set and any optional fields will be set to a default value
func (ruleProperties *RuleProperties) Validate() *errors.ModelError {
	if ruleProperties.Limit == nil {
		return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("received a null value")}
	} else {
		if *ruleProperties.Limit < -1 {
			return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("is set below the minimum value of -1. Value must be [-1 (ulimited) | 0+ (zero or more specific limit)")}
		}
	}

	return nil
}

// RuleState has all the Actionable details that affect the Rule's state and these are Read-Only specifications
// set from the service directly
type RuleState struct {
	// Record details here instead of the Spec.Definition?
	// But now, these are not queryable
	//
	// 	ID string
	//	CreatedAt time.Datetime
	//	DeletedAt *time.Datetime

	// Overrides for the particular rule
	Overrides Overrides `json:"Overrides,omitempty"`

	// TODO:
	// When thinking about "destroy" opertaions, I think it would be great to record a few operations (can fit all models with children):
	//
	// 1. Record if the Rule + all Overrides are being destroyed
	// NOTE: works for all models
	// Destroying bool `json:"Destroying"`
	//
	// 2. Record specific queries of objects to destroy. Then any requests that come in about creating/updating an Override
	//    need to ensure they are not matchin the Destroying operations.
	// NOTE: works for modesl with children
	// Destroying []*v1common.AssociatedActionQuery
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleState has all required fields set from the Service
func (ruleState *RuleState) Validate() *errors.ModelError {
	if err := ruleState.Overrides.Validate(); err != nil {
		return &errors.ModelError{Field: "Overrides", Child: err}
	}

	return nil
}
