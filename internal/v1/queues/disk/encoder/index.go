package encoder

import (
	"fmt"
	"os"

	"github.com/DanLavine/willow/internal/v1/models"
)

type index struct {
	file *os.File

	// endIndex is the last index in the file
	endIndex int

	// lastWriteIndex is the starting location for the last index
	lastWriteIndex int
}

func newIndex(fileName string) (*index, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed creating queue file: %w", err)
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
func (i *index) Write(id uint64, data []byte) (int, int, error) {
	prefix, suffix, encodedData := EncodeByteWithEnding(data, id)

	var writeIndex int
	if i.endIndex > 0 {
		// note, use -1 to strip off the last '.' for the ending char
		writeIndex = i.endIndex - 1
	}

	n, err := i.file.WriteAt(encodedData, int64(writeIndex))
	if err != nil {
		return writeIndex + prefix, n, err
	}

	// set the last index, to the location where we started writting at
	i.lastWriteIndex = writeIndex
	i.endIndex += n

	return i.lastWriteIndex + prefix, n - prefix - suffix, nil
}

// Read a specific location and return the original data decoded
func (i *index) Read(startIndex, size int) ([]byte, error) {
	data := make([]byte, size)

	_, err := i.file.ReadAt(data, int64(startIndex))
	if err != nil {
		return nil, fmt.Errorf("Failed to read file at start location: %d, for %d bytes: %w", startIndex, size, err)
	}

	return data, nil
}

// OverwriteLast updates the last index with new data. When calling this, the data should already be encoded
// with the proper endings '..' included
//
// RETURNS:
// * location - location state on disk
// * error - an error encounterd during the write
func (i *index) Overwrite(encodedData []byte, location *models.Location) error {
	n, err := i.file.WriteAt(encodedData, int64(i.lastWriteIndex))
	if err != nil {
		return err
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

	_, err := d.file.ReadAt(data, int64(d.lastWriteIndex))
	return data, err
}

func (f *index) Close() error {
	return f.file.Close()
}
