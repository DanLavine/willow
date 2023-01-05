package models

import "fmt"

type QueueMetrics struct {
	Ready      int
	Processing int
}

type Metrics struct {
	// [queue name][tag name]QueueMetrics
	Queues map[string]map[string]QueueMetrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		Queues: map[string]map[string]QueueMetrics{},
	}
}

func (m *Metrics) Add(queue string, tag string, metrics QueueMetrics) error {
	if queues, ok := m.Queues[queue]; ok {
		if _, found := queues[tag]; found {
			return fmt.Errorf("name and tag already exist")
		}

		m.Queues[queue][tag] = metrics
	} else {
		m.Queues[queue] = map[string]QueueMetrics{tag: metrics}
	}

	fmt.Println("queues:", m)

	return nil
}
