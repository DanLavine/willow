package queues_integration_tests

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/onsi/gomega"
)

func Test_ACK(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("it returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: datatypes.String("queue1"),
				Tags: datatypes.StringMap{"not": "found"},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ACK(g, ackRequest)
		g.Expect(ackResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("it returns an error if the tags don't exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.Create(g, testhelpers.Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: datatypes.String("queue1"),
				Tags: datatypes.StringMap{"not": "found"},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ACK(g, ackRequest)
		g.Expect(ackResponse.StatusCode).To(Equal(http.StatusBadRequest))

		body, err := io.ReadAll(ackResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(body)).To(ContainSubstring("tag group not found"))
	})

	t.Run("it returns an error if the ID is not processing", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.Create(g, testhelpers.Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue an item
		enqueueResponse := testConstruct.Enqueue(g, testhelpers.Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: datatypes.String("queue1"),
				Tags: datatypes.StringMap{"some": "tag"},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ACK(g, ackRequest)
		g.Expect(ackResponse.StatusCode).To(Equal(http.StatusBadRequest))

		body, err := io.ReadAll(ackResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(body)).To(ContainSubstring("ID 1 is not processing"))
	})

	t.Run("when setting 'processed = true'", func(t *testing.T) {
		t.Run("it deletes the item", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			defer func() {
				fmt.Println(string(testConstruct.Session.Out.Contents()))
				fmt.Println(string(testConstruct.Session.Err.Contents()))
			}()

			// create the queue
			createResponse := testConstruct.Create(g, testhelpers.Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue an item
			enqueueResponse := testConstruct.Enqueue(g, testhelpers.Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			// dequeue an item
			dequeueResponse := testConstruct.Dequeue(g, testhelpers.Queue1Dequeue)
			g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

			// check metrics
			metrics := testConstruct.Metrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("queue1")))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
			g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": "tag"}))
			g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(1)))

			// ack the dequeued message
			ackRequest := v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: datatypes.String("queue1"),
					Tags: datatypes.StringMap{"some": "tag"},
				},
				ID:     1,
				Passed: true,
			}

			ackResponse := testConstruct.ACK(g, ackRequest)
			g.Expect(ackResponse.StatusCode).To(Equal(http.StatusOK))

			// ensure the item has been removed
			metrics = testConstruct.Metrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal(datatypes.String("queue1")))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(0)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(0))
		})
	})
}
