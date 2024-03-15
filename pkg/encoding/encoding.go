package encoding

import (
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/encoding/json"
)

const (
	ContentType = "Content-Type"

	ContentTypeJSON   = json.ContentTypeJSON
	ContentTypeUnkown = "unkown"
)

type Encoder interface {
	Encode(obj any) ([]byte, error)

	Decode(reader io.ReadCloser, obj any) error

	ContentType() string
}

func NewEncoder(contentType string) (Encoder, error) {
	switch contentType {
	case "", json.ContentTypeJSON:
		return json.JsonEncoder{}, nil
	default:
		return nil, fmt.Errorf("unknown content type for the encoder: %s", contentType)
	}
}
