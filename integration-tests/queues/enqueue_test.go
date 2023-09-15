package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		enqueueBody := v1willow.EnqueueItemRequest{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test queue",
				Tags: datatypes.StringMap{"some": datatypes.String("tag")},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}

		createResponse := testConstruct.ServerClient.WillowEnqueue(g, enqueueBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("enqueus a message on the matching tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1willow.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}
		createResponse := testConstruct.ServerClient.WillowCreate(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := v1willow.EnqueueItemRequest{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test queue",
				Tags: datatypes.StringMap{"some": datatypes.String("tag")},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, enqueueBody)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal("test queue"))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the same tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue 4 times
		item := Queue1UpdateableEnqueue
		item.Updateable = false
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(4)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the different tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue multiple tags
		item := Queue1UpdateableEnqueue
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		item.BrokerInfo.Tags = datatypes.StringMap{"new": datatypes.String("tag"), "of": datatypes.String("course")}
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(2)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(2))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1willow.TagMetricsResponse{Tags: datatypes.StringMap{"some": datatypes.String("tag")}, Ready: 1, Processing: 0}))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1willow.TagMetricsResponse{Tags: datatypes.StringMap{"new": datatypes.String("tag"), "of": datatypes.String("course")}, Ready: 1, Processing: 0}))
	})

	t.Run("updateable messages can collapse", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue 4 times
		item := Queue1UpdateableEnqueue
		item.Updateable = true
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.ServerClient.WillowMetrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("limit guards", func(t *testing.T) {
		t.Run("it returns an error when trying to enqueue an item on a full queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			createBody := Queue1
			createBody.QueueMaxSize = 1
			createResponse := testConstruct.ServerClient.WillowCreate(g, createBody)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue
			item := Queue1UpdateableEnqueue
			item.Updateable = false
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, item)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			//should case an error when enquing 2nd item
			enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, item)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusTooManyRequests))
		})
	})
}
