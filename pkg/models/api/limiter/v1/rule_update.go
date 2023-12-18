package v1

import (
	"encoding/json"
)

type RuleUpdate struct {
	// update the limits for a rule
	Limit uint64
}

// Convert the RuleUpdate logic to bytes that both the Client and Server understand
func (ru RuleUpdate) ToBytes() []byte {
	data, _ := json.Marshal(ru)
	return data
}
