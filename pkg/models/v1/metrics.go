package v1

// Metrics for all queues by name
type Metrics struct {
	// Name of each queue and their metrics
	Queues []QueueMetrics
}

// Metrics for each individual queue
type QueueMetrics struct {
	Tags []string

	Ready      uint64
	Processing uint64
}
