package api

// All pkg models need to suport the encoding types
type APIObject interface {
	Validate() error

	// Encode an APIObject into the type specified by the contentType
	EncodeJSON() ([]byte, error)

	// Decode an APIObject into the type specified by the contentType
	DecodeJSON(data []byte) error
}
