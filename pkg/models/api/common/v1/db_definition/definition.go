package dbdefinition

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// These are the possible types that make up a definition, but the APIs only want to specify
// the actuall types they care about...

type Definition struct {
	// Named is a single string that defines the unique name to use when saving an object in the DB
	Named *string `json:"Named,omitempty"`

	// Associated is a KeyValue definiton with an optional name that defines how to save the object in the DB
	Associated *Associated `json:"Associated,omitempty"`

	// Child of a parent object that defines the relation to the parent
	// TODO: when it is actually used
	// OneToManyRelation *OneToManyRelation `json:"OneToManyRelation,omitempty"`
}

// Pass in some sort of coniguration for validation?
func (definition *Definition) Validate() *errors.ModelError {
	if definition.Named == nil && definition.Associated == nil {
		return &errors.ModelError{Err: fmt.Errorf("no defintion defined")}
	}

	if definition.Named != nil && definition.Associated != nil {
		return &errors.ModelError{Err: fmt.Errorf("multiple definitions defined, but requires exactly 1")}
	}

	if definition.Named != nil {
		if *definition.Named == "" {
			return &errors.ModelError{Field: "Named", Err: fmt.Errorf("received the empty string")}
		}
	}

	if definition.Associated != nil {
		if err := definition.Associated.Validate(); err != nil {
			return &errors.ModelError{Field: "Associated", Child: err}
		}
	}

	return nil
}

// Validate the combinations specifically?
func (definition *Definition) ValidateNamed() *errors.ModelError {
	if definition.Named == nil {
		return &errors.ModelError{Field: "Named", Err: fmt.Errorf("received a null value")}
	}

	if definition.Named != nil {
		if *definition.Named == "" {
			return &errors.ModelError{Field: "Named", Err: fmt.Errorf("received the empty string")}
		}
	}

	return nil
}

func (defintion *Definition) ValidateAssociated() *errors.ModelError {
	return nil
}
