package v1

import "github.com/DanLavine/willow/pkg/models/api/common/errors"

type RequeueLocation uint

const (
	RequeueFront RequeueLocation = iota
	RequeueEnd
	RequeueNone
)

type ACK struct {
	// common broker info
	BrokerInfo

	// ID of the original message being acknowledged
	ID string

	// Indicate a success or failure of the message
	Passed          bool
	RequeueLocation RequeueLocation // only used when set to false
}

func (ack *ACK) Validate() *errors.Error {
	if err := ack.BrokerInfo.validate(); err != nil {
		return err
	}

	if ack.ID == "" {
		return errors.InvalidRequestBody.With("ID cannot be an empty string", "")
	}

	return nil
}
