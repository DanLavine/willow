package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Rule saved in the DB that actions can be performed against
type Rule struct {
	Spec *RuleSpec `json:"Spec"`

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
	// DBDefinition defines how to save the item in the Database
	DBDefinition *RuleDBDefinition `json:"DBDefinition,omitempty"`

	// Properties are the configurable/updateable fields for the Rule
	Properties *RuleProperties `json:"Properties,omitempty"`
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleSpec has all required fields set
func (ruleSpec *RuleSpec) Validate() *errors.ModelError {
	if ruleSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := ruleSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
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

type RuleDBDefinition struct {
	// Name to store in the DB for the Rule. This will be used in the API urls for a quick lookup
	Name *string `json:"ID,omitempty"`

	// KeyValues that define the Rule and match against all Counters
	GroupByKeyValues datatypes.AnyKeyValues `json:"GroupByKeyValues"`
}

func (ruleDBDefinition *RuleDBDefinition) Validate() *errors.ModelError {
	if ruleDBDefinition.Name == nil {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("received a null value")}
	} else {
		if *ruleDBDefinition.Name == "" {
			return &errors.ModelError{Field: "Name", Err: fmt.Errorf("received an empty string")}
		}
	}

	if err := ruleDBDefinition.GroupByKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "GroupByKeyValues", Child: err}
	}

	return nil
}

// RuleProperties are all the properties any actinable APIs will use. These fields are updateable
type RuleProperties struct {
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
