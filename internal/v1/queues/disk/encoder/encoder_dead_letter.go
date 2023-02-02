package encoder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var deadLetterFileName = "deadLetter.idx"

type EncoderDeadLetter struct {
	file           *os.File
	lastWriteIndex int
}

func NewEncoderDeadLetter(baseDir string, queueTags []string) (*EncoderDeadLetter, *v1.Error) {
	deadLetterQueueDir, err := FilePath(baseDir, queueTags)
	if err != nil {
		return nil, err
	}

	file, openErr := os.OpenFile(filepath.Join(deadLetterQueueDir, deadLetterFileName), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, errors.FailedToCreateDeadLetterFile.With("", openErr.Error())
	}

	return &EncoderDeadLetter{
		file:           file,
		lastWriteIndex: 0,
	}, nil
}

// Writes data to disk at the end of the file and updates the last index counter
//
// PARAMS:
// * data - encoded data to write to disk. Will always be appended at the end
//
// RETURNS:
// * int - start location on disk where the write happend
// * int - size of encoded data written. Does not include the path seperator
// * error - an error encounterd during the write
func (edl *EncoderDeadLetter) Write(data []byte) (*models.DiskLocation, *v1.Error) {
	n, err := edl.file.Write(AddSeperator(data))

	if err != nil {
		return nil, errors.WriteFailed.With("dead letter queue write to succeed", err.Error())
	}

	// set the last index, to the location where we started writting at
	lastWriteIndex := edl.lastWriteIndex
	edl.lastWriteIndex += n

	return &models.DiskLocation{StartIndex: lastWriteIndex, Size: n - 1}, nil
}

// Read data from dead letter file.
//
// PARAMS:
// * startIndex - location to start reading from disk
// * size - how many bytes to read from disk
//
// RETURNS:
// * []byte - decoded data from disk (original queue item that failed)
// * error - an error encounterd during the read or decode operation
func (edl *EncoderDeadLetter) Read(startIndex, size int) ([]byte, *v1.Error) {
	data := make([]byte, size)
	_, err := edl.file.ReadAt(data, int64(startIndex))
	if err != nil {
		return nil, errors.ReadFailed.With(fmt.Sprintf("start location %d for %d bytes to be valid", startIndex, size), err.Error())
	}

	decodedData, decodeErr := DecodeByte(data)
	if decodeErr != nil {
		return nil, errors.DecodeFailed.With(fmt.Sprintf("decode at start location: %d, for %d bytes", startIndex, size), decodeErr.Error())
	}

	return decodedData, nil
}

// Read a specific location and return the original data decoded
func (edl *EncoderDeadLetter) Clear() *v1.Error {
	if err := edl.file.Truncate(0); err != nil {
		return errors.TruncateError.With("truncate of index dead letter to succeed", err.Error())
	}

	edl.lastWriteIndex = 0

	return nil
}

func (edl *EncoderDeadLetter) Close() {
	edl.file.Close()
}
