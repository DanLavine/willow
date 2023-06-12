package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func Test_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		enqueueBody := v1.EnqueueItemRequest{
			BrokerInfo: v1.BrokerInfo{
				Name: "test queue",
				Tags: datatypes.StringMap{"some": "tag"},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}

		createResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("enqueus a message on the matching tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			DeadLetterQueueMaxSize: 0,
		}
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := v1.EnqueueItemRequest{
			BrokerInfo: v1.BrokerInfo{
				Name: "test queue",
				Tags: datatypes.StringMap{"some": "tag"},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}
		enqueueResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("test queue")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": "tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the same tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.Create(g, testhelpers.Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue 4 times
		item := testhelpers.Queue1UpdateableEnqueue
		item.Updateable = false
		enqueurResponse := testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("queue1")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(4)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": "tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the different tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.Create(g, testhelpers.Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue multiple tags
		item := testhelpers.Queue1UpdateableEnqueue
		enqueurResponse := testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		item.BrokerInfo.Tags = datatypes.StringMap{"new": "tag", "of": "course"}
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("queue1")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(2)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(2))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1.TagMetricsResponse{Tags: datatypes.StringMap{"some": "tag"}, Ready: 1, Processing: 0}))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1.TagMetricsResponse{Tags: datatypes.StringMap{"new": "tag", "of": "course"}, Ready: 1, Processing: 0}))
	})

	t.Run("updateable messages can collapse", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createResponse := testConstruct.Create(g, testhelpers.Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue 4 times
		item := testhelpers.Queue1UpdateableEnqueue
		item.Updateable = true
		enqueurResponse := testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, item)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("queue1")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": "tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("limit guards", func(t *testing.T) {
		t.Run("it returns an error when trying to enqueue an item on a full queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			createBody := testhelpers.Queue1
			createBody.QueueMaxSize = 1
			createResponse := testConstruct.Create(g, createBody)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue
			item := testhelpers.Queue1UpdateableEnqueue
			item.Updateable = false
			enqueurResponse := testConstruct.Enqueue(g, item)
			g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

			//should case an error when enquing 2nd item
			enqueurResponse = testConstruct.Enqueue(g, item)
			g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusTooManyRequests))
		})
	})
}
