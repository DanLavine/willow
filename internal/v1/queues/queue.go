package queues

import v1 "github.com/DanLavine/willow/pkg/models/v1"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type Queue interface {
	// Initialize any background processes or setup a queue mght need
	Init() *v1.Error

	// Enqueue an item onto the queue
	Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error
}
