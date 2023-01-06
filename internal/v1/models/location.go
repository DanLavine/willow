package models

import (
	"fmt"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Location struct {
	// ID of the location
	ID uint64

	// place on disk to start reading from
	StartIndex int

	// size of data to read from disk
	Size int

	// number of times this location has been retried
	RetryCount int

	// Callback to set processing for the element
	process func(id uint64, sartIndex, size int) (*v1.DequeueMessage, error)
}

func NewLocation(processCallback func(id uint64, startIndex, size int) (*v1.DequeueMessage, error)) *Location {
	return &Location{
		process: processCallback,
	}
}

func (l *Location) Process() (*v1.DequeueMessage, error) {
	if l.process == nil {
		return nil, fmt.Errorf("process is nil")
	}

	return l.process(l.ID, l.StartIndex, l.Size)
}
