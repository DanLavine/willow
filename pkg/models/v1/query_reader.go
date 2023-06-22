package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// This is really all I need for the readers. Its very simple, but is 100% inclusive only.
// There is no way to exclude something from a reaturn value. If thats the case I would suggest
// setting up a new 'Restrictive Queue' and there each client should pull Exactly what they need
//
// Revisiting how to do exclusions might be a fun exercise in the future, but I wasted a lot of time
// already tring to do that. With how my channel readers work, it makes 0 sense to do for now. But
// it will be required when I have my queue limiter in the future.

type ReaderSelect struct {
	// required broker name to search
	BrokerName string

	// Find all values from any of the provided queries
	// NOTE: setting this to nil, means to use the global reader
	Queries []ReaderQuery
}

type ReaderType string

const (
	ReaderExactly ReaderType = "exactly"
	ReaderMatches ReaderType = "matches"
)

type ReaderQuery struct {
	Type ReaderType

	Tags datatypes.StringMap
}

func ParseReaderSelect(reader io.ReadCloser) (*ReaderSelect, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	obj := &ReaderSelect{}
	if err := json.Unmarshal(body, obj); err != nil {
		return nil, ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); err != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (rs *ReaderSelect) Validate() *Error {
	if rs.BrokerName == "" {
		return InvalidRequestBody.With("BrokerName needs to be provided", "Name is the empty string")
	}

	if rs.Queries != nil {
		for index, query := range rs.Queries {
			switch query.Type {
			case ReaderExactly, ReaderMatches:
				// nothing to do here, this is valid
			default:
				return InvalidRequestBody.With(fmt.Sprintf("Query index %d has an invalid Type", index), "the type to be [exactly | matches]")
			}

			if len(query.Tags) == 0 {
				return InvalidRequestBody.With(fmt.Sprintf("Query index %d has an empty Tags field", index), "the Tags field to have at least 1 key value pair")
			}
		}
	}

	return nil
}
