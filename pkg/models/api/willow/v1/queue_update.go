package v1

type QueueUpdate struct {
	// Max size of the queue
	QueueMaxSize int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (qu QueueUpdate) Validate() error {
	return nil
}
