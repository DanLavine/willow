package disk

type location struct {
	// where to start reading disk from for the encoded data
	start int

	// size of data to read from disk
	size int

	// number of times an item was retried
	retryCount int
}
