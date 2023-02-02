package encoder

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

func EncodeString(data string) string {
	encodedData := EncodeByte([]byte(data))
	return string(encodedData)
}

func EncodeStrings(data []string) string {
	if len(data) == 0 {
		return ""
	}

	var encodedDatas = []string{}
	for _, rawData := range data {
		encodedDatas = append(encodedDatas, EncodeString(rawData))
	}

	return EncodeString(strings.Join(encodedDatas, "@"))
}

// RETURNS:
// * int - prefix length
// * int - suffix length
// * []byte - encoded data
func EncodeByteWithEnding(data []byte, id uint64) (int, int, []byte) {
	// total lenght for additional data
	prefix := []byte(fmt.Sprintf("%d@", id))
	prefixLen := len(prefix)

	// encoded data len
	encodedLen := base64.StdEncoding.EncodedLen(len(data))

	// suffix
	suffix := []byte(`..`)
	suffixLen := len(suffix)

	finalEncoding := make([]byte, prefixLen+encodedLen+suffixLen)

	// copy prefix
	copy(finalEncoding, prefix)

	// encoded data
	base64.StdEncoding.Encode(finalEncoding[prefixLen:], data)

	// copy prefix
	copy(finalEncoding[prefixLen+encodedLen:], suffix)

	// return encoded data, and size of new byte array
	return prefixLen, 2, finalEncoding
}

func EncodeByteWithSeperator(data []byte) []byte {
	bytesLength := base64.StdEncoding.EncodedLen(len(data))
	encoded := make([]byte, bytesLength+1)
	base64.StdEncoding.Encode(encoded, data)

	// set the final 2 chars to '..'
	encoded[bytesLength] = '.' // write the seperator index

	// return encoded data, and size of new byte array
	return encoded
}

func EncodeByte(data []byte) []byte {
	bytesLength := base64.StdEncoding.EncodedLen(len(data))
	encoded := make([]byte, bytesLength)
	base64.StdEncoding.Encode(encoded, data)

	return encoded
}

func DecodeByte(data []byte) ([]byte, *v1.Error) {
	bytesLength := base64.StdEncoding.DecodedLen(len(data))
	decoded := make([]byte, bytesLength)

	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return nil, errors.DecodeFailed.With("", err.Error())
	}

	// strip out white space
	return decoded[:n], nil
}
