package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/onsi/gomega"
)

func Test_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)

	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		enqueueBody := v1.EnqueueItemRequest{
			BrokerInfo: v1.BrokerInfo{
				Name:       "test queue",
				BrokerType: v1.Queue,
				Tags:       v1.Strings{"some tag"},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}

		createResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusBadRequest))
	})

	t.Run("enqueus a message on the matching tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			Name:                   "test queue",
			QueueMaxSize:           5,
			ItemRetryAttempts:      0,
			DeadLetterQueueMaxSize: 0,
		}
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := v1.EnqueueItemRequest{
			BrokerInfo: v1.BrokerInfo{
				Name:       "test queue",
				BrokerType: v1.Queue,
				Tags:       v1.Strings{"some tag"},
			},
			Data:       []byte(`hello world`),
			Updateable: false,
		}
		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(v1.String("test queue")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(v1.Strings{"some tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the same tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := testhelpers.DefaultCreate()
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := testhelpers.DefaultEnqueueItemRequestNotUpdateable()

		// enqueue 4 times
		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(v1.String("test queue")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(4)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(v1.Strings{"some tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("enqueus multiple messages with the different tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := testhelpers.DefaultCreate()
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := testhelpers.DefaultEnqueueItemRequestNotUpdateable()

		// enqueue multiple tags
		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		enqueueBody.BrokerInfo.Tags = v1.Strings{"new tag", "of course"}
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(v1.String("test queue")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(2)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(2))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1.TagMetricsResponse{Tags: v1.Strings{"some tag"}, Ready: 1, Processing: 0}))
		g.Expect(metrics.Queues[0].Tags).To(ContainElement(&v1.TagMetricsResponse{Tags: v1.Strings{"new tag", "of course"}, Ready: 1, Processing: 0}))
	})

	t.Run("updateable messages can collapse", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := testhelpers.DefaultCreate()
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		enqueueBody := testhelpers.DefaultEnqueueItemRequestUpdateable()

		// enqueue 4 times
		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))

		// check the metrics
		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Name).To(Equal(v1.String("test queue")))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(v1.Strings{"some tag"}))
		g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("limit guards", func(t *testing.T) {
		t.Run("it returns an error when trying to enqueue an item on a full queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			createBody := testhelpers.DefaultCreate()
			createBody.QueueMaxSize = 1
			createResponse := testConstruct.Create(g, createBody)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			enqueueBody := testhelpers.DefaultEnqueueItemRequestNotUpdateable()

			// enqueue 4 times
			enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
			g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
			enqueurResponse = testConstruct.Enqueue(g, enqueueBody)

			//should case an error when enquing 2nd item
			g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusTooManyRequests))
			enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
		})
	})
}
