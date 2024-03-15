package json

import (
	"encoding/json"
)

func (jsonEncoder JsonEncoder) Encode(obj any) ([]byte, error) {
	return json.Marshal(obj)
}
