package models

// This is used for the in memory queue
type QueueItem struct {
	ID uint64

	Data []byte

	Processing bool
}
