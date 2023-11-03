package v1locker

import (
	"encoding/json"
)

type HeartbeatError struct {
	Session string `json:"Session"`
	Error   string `json:"Error"`
}

func (heartbeatError *HeartbeatError) ToBytes() []byte {
	data, _ := json.Marshal(heartbeatError)
	return data
}
