package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type Channels []*Channel

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (c Channels) Validate() *errors.ModelError {
	if len(c) == 0 {
		return nil
	}

	for index, singleChan := range c {
		if err := singleChan.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
