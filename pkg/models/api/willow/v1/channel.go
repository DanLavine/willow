package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
)

type Channel struct {
	Spec *ChannelSpec `json:"Spec,omitempty"`

	State *ChannelState `json:"State,omitempty"`
}

func (channel *Channel) Validate() *errors.ModelError {
	if channel.Spec == nil {
		return &errors.ModelError{Field: "Spec", Err: fmt.Errorf("received an empty spcification")}
	} else {
		if err := channel.Spec.Validate(); err != nil {
			return err
		}
	}

	if channel.State == nil {
		return &errors.ModelError{Field: "State", Err: fmt.Errorf("received an empty state")}
	} else {
		if err := channel.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type ChannelSpec struct {
	DBDefinition *ChannelDBDefinition `json:"DBDefinition,omitempty"`
}

func (channelSpec *ChannelSpec) Validate() *errors.ModelError {
	if channelSpec.DBDefinition == nil {
		return &errors.ModelError{Field: "DBDefinition", Err: fmt.Errorf("received a null value")}
	} else {
		if err := channelSpec.DBDefinition.Validate(); err != nil {
			return &errors.ModelError{Field: "DBDefinition", Child: err}
		}
	}

	return nil
}

type ChannelDBDefinition struct {
	KeyValues dbdefinition.TypedKeyValues `json:"KeyValues,omitempty"`
}

func (channelDBDefinition *ChannelDBDefinition) Validate() *errors.ModelError {
	if err := channelDBDefinition.KeyValues.Validate(); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}

type ChannelState struct {
	// Total number of enqueued items that are not processing
	EnqueuedItems int64

	// Total numbe of items that are currently processing
	ProcessingItems int64
}

func (channelState *ChannelState) Validate() *errors.ModelError {
	if channelState.EnqueuedItems+channelState.ProcessingItems == 0 {
		return &errors.ModelError{Err: fmt.Errorf("total number of items equals 0")}
	}

	return nil
}
