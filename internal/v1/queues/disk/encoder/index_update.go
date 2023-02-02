package encoder

import (
	"os"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type updateFile struct {
	file *os.File

	nextIndex int
}

func newUpdateFile(fileName string) (*updateFile, *v1.Error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, errors.FailedToCreateUpdateFile.With("", err.Error())
	}

	return &updateFile{
		file: file,
	}, nil
}

// Write to the update file a record of what we are updating.
func (f *updateFile) Write(currentData, nextData []byte) *v1.Error {
	// the currentData is already base64 encoded so just write it directly
	n, err := f.file.Write(currentData)
	if err != nil {
		return errors.WriteFailed.With("update file current data to succeed", err.Error())
	}

	// write the seperator
	m, err := f.file.Write([]byte(`.`))
	if err != nil {
		return errors.WriteFailed.With("update file seperator write to succeed", err.Error())
	}

	// record the seperator location
	f.nextIndex = n + m

	// next data should also already be encoded.
	if _, err = f.file.Write(nextData); err != nil {
		return errors.WriteFailed.With("update file update data to succeed", err.Error())
	}

	return nil
}

// Clear is called after the update successfully processes
func (f *updateFile) Clear() error {
	return f.file.Truncate(0)
}

// Close the file
func (f *updateFile) Close() error {
	return f.file.Close()
}
