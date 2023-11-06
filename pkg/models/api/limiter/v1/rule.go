package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

/*
The name should be ignored on the "find queries"...
OR, we could check Name EXISTS? That would solve the problem since
it would be true for everything?s
var K = datatypes.KeyValues{
	"name":     datatypes.String("name"),
	"groupBy1": datatypes.Nil(),
	"groupBy2": datatypes.Nil(),
	//...
}

// This would be the query to search for all possible rules that match above Find for Key Values
var beTrue = true
var two = 2
var three = 3
var TQuery = datatypes.AssociatedKeyValuesQuery{
	Or: []datatypes.AssociatedKeyValuesQuery{
		{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"groupBy1": datatypes.Value{Exists: &beTrue},
				},
				Limits: &datatypes.KeyLimits{
					NumberOfKeys: &two, // this is the number + to account for the 'name'. Will always be this way
				},
			},
			Or: []datatypes.AssociatedKeyValuesQuery{
				{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"groupBy1": datatypes.Value{Exists: &beTrue},
							"groupBy2": datatypes.Value{Exists: &beTrue},
						},
						Limits: &datatypes.KeyLimits{
							NumberOfKeys: &three,
						},
					},
				},
			},
		},
		{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"groupBy2": datatypes.Value{Exists: &beTrue},
				},
				Limits: &datatypes.KeyLimits{
					NumberOfKeys: &two,
				},
			},
		},
	},
}

Another option would be to have a new query option where where I lookup all the keys 1x since this query above is looking up
each key N times which isn't great. Then save all the nodes /w the IDs and merge the combinations after as a "pure existene"
check only.

*/

type Rule struct {
	// Name of the rule
	Name string // save this as the _associated_id in the the tree?

	// These can be used to create a rule groupiing that any tags will have to match agains
	GroupBy []string // now these can just be normal "key values"

	// When comparing tags, use this selection to figure out if a rule applies to them
	QueryFilter datatypes.AssociatedKeyValuesQuery

	// Limit Key is an optional param that can be used to dictate what value of the tags to use as a limiter
	Limit uint64

	// This is a "Read Only" parameter and will be ignored on create operations
	Overrides []Override
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleRequest(reader io.ReadCloser) (*Rule, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &Rule{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.ValidateRequest(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

// Client side logic to parse a Rule
func ParseRuleResponse(reader io.ReadCloser) (*Rule, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &Rule{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return obj, nil
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rreq *Rule) ValidateRequest() *api.Error {
	if rreq.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	}

	if len(rreq.GroupBy) == 0 {
		return api.InvalidRequestBody.With("GroupBy tags to be provided", "recieved empty tag grouping")
	}

	if err := rreq.QueryFilter.Validate(); err != nil {
		return api.InvalidRequestBody.With("Selection query to be a valid expression", err.Error())
	}

	if len(rreq.Overrides) != 0 {
		return api.InvalidRequestBody.With("Overrides must be empty", "recieved an override")
	}

	return nil
}

func (rule *Rule) ToBytes() []byte {
	data, _ := json.Marshal(rule)
	return data
}
