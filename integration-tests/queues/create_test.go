package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/pkg/models/api/v1willow"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("can create a queue with proper name", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1willow.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}

		createResponse := testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal("test queue"))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Tags).To(BeNil())
		g.Expect(metrics.Queues[0].DeadLetterQueueMetrics).To(BeNil())
	})

	t.Run("can create a multiple queues", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1willow.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}

		createResponse := testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		createBody = v1willow.Create{
			Name:                   "other queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}

		createResponse = testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(2))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "test queue", Total: 0, Max: 5, Tags: nil, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "other queue", Total: 0, Max: 5, Tags: nil, DeadLetterQueueMetrics: nil}))
	})

	t.Run("can create the same queue multple times", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1willow.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}

		// first create
		createResponse := testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// second create
		createResponse = testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "test queue", Total: 0, Max: 5, Tags: nil, DeadLetterQueueMetrics: nil}))
	})
}
