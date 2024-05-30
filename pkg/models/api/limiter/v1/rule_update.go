package v1

import "github.com/DanLavine/willow/pkg/models/api/common/errors"

// RuleUpdateRquest defines the fields of a Rule that can be updated
type RuleUpdateRquest struct {
	// Limit for a particual Rule
	// set this to -1 for unlimited
	Limit int64
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleUpdateRquest has all required fields set
func (req *RuleUpdateRquest) Validate() *errors.ModelError {
	// no-op atm
	return nil
}
