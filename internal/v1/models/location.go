package models

import (
	"github.com/DanLavine/willow/internal/errors"
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
	RetryCount uint64

	// used to know if an item has actually started processing
	processing bool

	// Callback to set processing for the element
	process func(id uint64, sartIndex, size int) (*v1.DequeueMessage, *v1.Error)
}

func NewLocation(processCallback func(id uint64, startIndex, size int) (*v1.DequeueMessage, *v1.Error)) *Location {
	return &Location{
		process: processCallback,
	}
}

func (l *Location) Process() (*v1.DequeueMessage, *v1.Error) {
	if l.process == nil {
		return nil, errors.ProcessNotSet
	}

	l.processing = true
	return l.process(l.ID, l.StartIndex, l.Size)
}

func (l *Location) Processing() bool {
	return l.processing
}
