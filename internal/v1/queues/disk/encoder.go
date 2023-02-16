package disk

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

func CreateOrOpenFile(baseDir string, queueName string, fileName string) (*os.File, *v1.Error) {
	encodeDir, err := FilePath(baseDir, queueName)
	if err != nil {
		return nil, err
	}

	file, openFileErr := os.OpenFile(filepath.Join(encodeDir, fileName), os.O_CREATE|os.O_RDWR, 0755)
	if openFileErr != nil {
		return nil, errors.FailedToCreateFile.With("", openFileErr.Error())
	}

	return file, nil
}

// Generate a safe file path based on any number of strings and
// check that is is valid for use.
//
// PARAMS:
// * baseDir - base direcory thats not encoded (generally top level mount or data dir where everything is saved to)
// * queueName - name of the queue that will be base64 encoded
//
// RETURNS:
// * string - single filepath for any os
func FilePath(baseDir, queueName string) (string, *v1.Error) {
	encodeDir := filepath.Join(baseDir, EncodeString(queueName))

	filePath, err := os.Stat(encodeDir)
	if os.IsPermission(err) || os.IsNotExist(err) {
		// create the dir
		if err = os.MkdirAll(encodeDir, 0755); err != nil {
			return "", errors.FailedToCreateDir.With("", err.Error())
		}
	} else if err != nil {
		// some other error encountered
		return "", errors.FailedToStatDir.With("", err.Error())
	} else {
		// path already exists and is not dir
		if !filePath.IsDir() {
			return "", errors.PathAlreadyExists.With(filePath.Name(), "to be a dir")
		}
	}

	return encodeDir, nil
}

// Encode a generic string
//
// PARAMS:
// * data - string to encode
//
// RETURNS:
// * string - encoded version of the string
func EncodeString(data string) string {
	encodedData := EncodeByte([]byte(data))
	return string(encodedData)
}

// Encode a slice of strings
//
// PARAMS:
// * data - slice of strings to encode
//
// RETURNS:
// * []string - slice of all strings encoded in the same order
func EncodeStrings(data []string) []string {
	if len(data) == 0 {
		return nil
	}

	encodedStrings := []string{}
	for _, rawData := range data {
		encodedStrings = append(encodedStrings, EncodeString(rawData))
	}

	return encodedStrings
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

func AddSeperator(data []byte) []byte {
	return append(data, '.')
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
