package encoder

import (
	"fmt"
	"os"
)

type updateFile struct {
	file *os.File

	nextIndex int
}

func newUpdateFile(fileName string) (*updateFile, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed creating update file: %w", err)
	}

	return &updateFile{
		file: file,
	}, nil
}

// Write to the update file a record of what we are updating.
func (f *updateFile) Write(currentData, nextData []byte) error {
	// the currentData is already base64 encoded so just write it directly
	n, err := f.file.Write(currentData)
	if err != nil {
		return err
	}

	// write the seperator
	m, err := f.file.Write([]byte(`.`))
	if err != nil {
		return err
	}

	// record the seperator location
	f.nextIndex = n + m

	// next data should also already be encoded.
	_, err = f.file.Write(nextData)
	return err
}

// Clear is called after the update successfully processes
func (f *updateFile) Clear() error {
	return f.file.Truncate(0)
}

// Close the file
func (f *updateFile) Close() error {
	return f.file.Close()
}
