package encoder

import (
	"fmt"
	"os"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type indexState struct {
	file *os.File
}

func newIndexState(fileName string) (*indexState, *v1.Error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, errors.FailedToCreateStateFile.With("", err.Error())
	}

	return &indexState{
		file: file,
	}, nil
}

func (is *indexState) Processing(id uint64) *v1.Error {
	return is.write(fmt.Sprintf("P%d.", id))
}

func (is *indexState) Retry(id uint64) *v1.Error {
	return is.write(fmt.Sprintf("R%d.", id))
}

func (is *indexState) Delete(id uint64) *v1.Error {
	return is.write(fmt.Sprintf("D%d.", id))
}

func (is *indexState) SentToDeadLetter(id uint64) *v1.Error {
	return is.write(fmt.Sprintf("S%d.", id))
}

func (is *indexState) write(data string) *v1.Error {
	if _, err := is.file.WriteString(data); err != nil {
		return errors.WriteFailed.With("state file write to succeed", err.Error())
	}

	return nil
}

func (is *indexState) Close() error {
	return is.file.Close()
}
