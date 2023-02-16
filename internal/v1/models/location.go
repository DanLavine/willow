package models

// this is the items.idx file representation.

// We can assume that each DATA file is exactl 8kb in size. if thats the case
// the we can figure out the actual data file for where to read from based off of the
// start index

type Location struct {
	// ID of the location
	ID uint64

	// place on disk to start reading from
	StartIndex int64

	// size of data to read from disk
	Size int64

	// number of times this location has been retried
	RetryCount uint64
}
