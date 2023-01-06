package encoder

import (
	"fmt"
	"os"
)

type indexState struct {
	file *os.File
}

func newIndexState(fileName string) (*indexState, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed creating queue file: %w", err)
	}

	return &indexState{
		file: file,
	}, nil
}

func (is *indexState) Processing(id uint64) error {
	return is.write(fmt.Sprintf("P%d.", id))
}

func (is *indexState) Retry(id uint64) error {
	return is.write(fmt.Sprintf("R%d.", id))
}

func (is *indexState) Delete(id uint64) error {
	return is.write(fmt.Sprintf("D%d.", id))
}

func (is *indexState) SentToDeadLetter(id uint64) error {
	return is.write(fmt.Sprintf("S%d.", id))
}

func (is *indexState) write(data string) error {
	if _, err := is.file.WriteString(data); err != nil {
		return fmt.Errorf("Failed to record index state for location: %w", err)
	}

	return nil
}

func (is *indexState) Close() error {
	return is.file.Close()
}
