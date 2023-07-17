package queues_integration_tests

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_ACK(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("it returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: "queue1",
				Tags: datatypes.StringMap{"not": datatypes.String("found")},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ServerClient.WillowACK(g, ackRequest)
		g.Expect(ackResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("it returns an error if the tags don't exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: "queue1",
				Tags: datatypes.StringMap{"not": datatypes.String("found")},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ServerClient.WillowACK(g, ackRequest)
		g.Expect(ackResponse.StatusCode).To(Equal(http.StatusBadRequest))

		body, err := io.ReadAll(ackResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(body)).To(ContainSubstring("tag group not found"))
	})

	t.Run("it returns an error if the ID is not processing", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue an item
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		ackRequest := v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: "queue1",
				Tags: datatypes.StringMap{"some": datatypes.String("tag")},
			},
			ID:     1,
			Passed: true,
		}

		ackResponse := testConstruct.ServerClient.WillowACK(g, ackRequest)
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
			createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue an item
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			// dequeue an item
			dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, Queue1Dequeue)
			g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

			// check metrics
			metrics := testConstruct.ServerClient.WillowMetrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
			g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
			g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(1)))

			// ack the dequeued message
			ackRequest := v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: "queue1",
					Tags: datatypes.StringMap{"some": datatypes.String("tag")},
				},
				ID:     1,
				Passed: true,
			}

			ackResponse := testConstruct.ServerClient.WillowACK(g, ackRequest)
			g.Expect(ackResponse.StatusCode).To(Equal(http.StatusOK))

			// ensure the item has been removed
			metrics = testConstruct.ServerClient.WillowMetrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(0)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(0))
		})
	})

	t.Run("when the client disconnects", func(t *testing.T) {
		t.Run("it fails a pending item", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			defer func() {
				fmt.Println(string(testConstruct.Session.Out.Contents()))
				fmt.Println(string(testConstruct.Session.Err.Contents()))
			}()

			// create the queue
			createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue an item
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			// dequeue an item
			dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, Queue1Dequeue)
			g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

			// check metrics
			metrics := testConstruct.ServerClient.WillowMetrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(1))
			g.Expect(metrics.Queues[0].Tags[0].Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
			g.Expect(metrics.Queues[0].Tags[0].Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Queues[0].Tags[0].Processing).To(Equal(uint64(1)))

			// have the client disconnect
			testConstruct.ServerClient.CloseClient()
			g.Eventually(func() string {
				return testConstruct.ServerStdout.String()
			}).Should(ContainSubstring("Client disconnect"))

			// ensure that the item was properly processed server side
			metrics = testConstruct.ServerClient.WillowMetrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Name).To(Equal("queue1"))
			g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
			g.Expect(metrics.Queues[0].Total).To(Equal(uint64(0)))
			g.Expect(len(metrics.Queues[0].Tags)).To(Equal(0))
		})
	})
}
