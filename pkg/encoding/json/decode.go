package json

import (
	"encoding/json"
	"io"
)

func (jsonEncoder JsonEncoder) Decode(reader io.ReadCloser, obj any) error {
	return json.NewDecoder(reader).Decode(obj)
}
