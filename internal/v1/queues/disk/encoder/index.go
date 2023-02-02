package encoder

import (
	"fmt"
	"os"

	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type index struct {
	file *os.File

	// endIndex is the last index in the file
	endIndex int

	// lastWriteIndex is the starting location for the last index
	lastWriteIndex int
}

func newIndex(fileName string) (*index, *v1.Error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, errors.FailedToCreateQueueFile.With("", err.Error())
	}

	return &index{
		file:           file,
		endIndex:       0,
		lastWriteIndex: 0,
	}, nil
}

// Write writes data to disk at the end of the file and updates the last index counters
//
// RETURNS:
// * int - start location on disk where the write happend
// * int - size of encoded data written. Does not include things like path seperators or other info
// * error - an error encounterd during the write
func (i *index) Write(id uint64, data []byte) (int, int, *v1.Error) {
	prefix, suffix, encodedData := EncodeByteWithEnding(data, id)

	var writeIndex int
	if i.endIndex > 0 {
		// note, use -1 to strip off the last '.' for the ending char
		writeIndex = i.endIndex - 1
	}

	n, err := i.file.WriteAt(encodedData, int64(writeIndex))
	if err != nil {
		return writeIndex + prefix, n, errors.WriteFailed.With("", err.Error())
	}

	// set the last index, to the location where we started writting at
	i.lastWriteIndex = writeIndex
	i.endIndex += n

	return i.lastWriteIndex + prefix, n - prefix - suffix, nil
}

// Read a specific location and return the original data decoded
func (i *index) Read(startIndex, size int) ([]byte, *v1.Error) {
	data := make([]byte, size)

	_, err := i.file.ReadAt(data, int64(startIndex))
	if err != nil {
		return nil, errors.ReadFailed.With(fmt.Sprintf("start location %d for %d bytes to be valid", startIndex, size), err.Error())
	}

	return data, nil
}

// OverwriteLast updates the last index with new data. When calling this, the data should already be encoded
// with the proper endings '..' included
//
// RETURNS:
// * location - location state on disk
// * error - an error encounterd during the write
func (i *index) Overwrite(encodedData []byte, location *models.Location) *v1.Error {
	n, err := i.file.WriteAt(encodedData, int64(i.lastWriteIndex))
	if err != nil {
		return errors.WriteFailed.With(fmt.Sprintf("start location %d for %d bytes to be valid", i.lastWriteIndex, len(encodedData)), err.Error())
	}

	// set the last index, to the location where we started writting at
	i.endIndex = i.lastWriteIndex + n

	return nil
}

func (d *index) ReadLast() ([]byte, error) {
	if d.endIndex == 0 {
		return nil, nil
	}

	data := make([]byte, d.endIndex-d.lastWriteIndex)
	if _, err := d.file.ReadAt(data, int64(d.lastWriteIndex)); err != nil {
		return nil, errors.ReadFailed.With(fmt.Sprintf("start location %d for %d bytes to be valid", d.lastWriteIndex, len(data)), err.Error())
	}

	return data, nil
}

func (f *index) Close() error {
	return f.file.Close()
}
