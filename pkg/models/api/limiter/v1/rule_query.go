package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type OverridesToInclude string

const (
	None  OverridesToInclude = ""
	All   OverridesToInclude = "all"
	Match OverridesToInclude = "match"
)

// Rule query is used to find any Rules that match the provided KeyValues
type RuleQuery struct {
	// Find any rules that match the provided key values. If this is nil, then OverrideQuery can be None or All.
	// If this has a value, then OverrideQuery can be Match
	KeyValues *datatypes.KeyValues

	// type of overrrides to query for
	// 1. Empty string - returns nothing
	// 2. All - returns all overrides
	// 3. Match - returns all override that match the KeyValues if there are any
	OverridesToInclude OverridesToInclude
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (ruleQuery *RuleQuery) Validate() error {
	if ruleQuery.KeyValues != nil {
		if len(*ruleQuery.KeyValues) == 0 {
			return fmt.Errorf("'KeyValues' requres at least 1 key + value piar")
		}

		if err := ruleQuery.KeyValues.Validate(); err != nil {
			return err
		}
	}

	switch ruleQuery.OverridesToInclude {
	case None, All, Match:
		// these are all valid
	default:
		return fmt.Errorf("'OverrideQuery' is %s, but must be one of ['' | all | match]", string(ruleQuery.OverridesToInclude))
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (ruleQuery *RuleQuery) EncodeJSON() []byte {
	data, _ := json.Marshal(ruleQuery)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (ruleQuery *RuleQuery) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, ruleQuery); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return ruleQuery.Validate()
}

// Collection of rules
type Rules struct {
	Rules []*RuleResponse
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (rules *Rules) Validate() error {
	if len(rules.Rules) == 0 {
		return nil
	}

	for index, rule := range rules.Rules {
		if rule == nil {
			return fmt.Errorf("invalid index in rules at indexLoc: %d: rule cannot be nil", index)
		}

		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid index in rules %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (rules *Rules) EncodeJSON() []byte {
	data, _ := json.Marshal(rules)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (rules *Rules) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, rules); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return rules.Validate()
}
