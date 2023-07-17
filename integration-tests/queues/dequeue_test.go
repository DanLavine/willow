package queues_integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Dequeue(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("it returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// dequeue the item
		dequeueRequest := v1.ReaderSelect{
			BrokerName: "queue1",
			Queries: []v1.ReaderQuery{
				{
					Type: v1.ReaderExactly,
					Tags: datatypes.StringMap{"some": datatypes.String("tag")},
				},
			},
		}

		dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("it returns a message that is waiting to be processed if they are available", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue an item
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// dequeue the item
		dequeueRequest := v1.ReaderSelect{
			BrokerName: "queue1",
			Queries: []v1.ReaderQuery{
				{
					Type: v1.ReaderExactly,
					Tags: datatypes.StringMap{"some": datatypes.String("tag")},
				},
			},
		}
		dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

		// parse response
		body, err := io.ReadAll(dequeueResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem := &v1.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, dequeueItem)).ToNot(HaveOccurred())

		// check item returned
		g.Expect(dequeueItem.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem.ID).To(Equal(uint64(1)))
		g.Expect(dequeueItem.Data).To(Equal([]byte(`hello world`)))
	})

	t.Run("it can recieve a message enqueued after the dequeue request", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// dequeue the item before the queue tags exists
		dequeueRequest := v1.ReaderSelect{
			BrokerName: "queue1",
			Queries: []v1.ReaderQuery{
				{
					Type: v1.ReaderExactly,
					Tags: datatypes.StringMap{"some": datatypes.String("tag")},
				},
			},
		}

		requestBody, err := json.Marshal(dequeueRequest)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueBuffer := bytes.NewBuffer(requestBody)
		request, err := http.NewRequest("GET", "https://127.0.0.1:8080/v1/brokers/item/dequeue", dequeueBuffer)
		g.Expect(err).ToNot(HaveOccurred())

		makingRequest := make(chan struct{})
		responseChan := make(chan *http.Response)
		errChan := make(chan error)
		go func() {
			close(makingRequest)
			response, err := testConstruct.ServerClient.Do(request)
			errChan <- err
			responseChan <- response
		}()

		g.Eventually(makingRequest).Should(BeClosed())
		time.Sleep(1 * time.Second) // make sure the request for a reader goes through first

		// enqueue an item
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// check response
		g.Eventually(errChan).Should(Receive(BeNil()))
		dequeueResponse := <-responseChan

		// parse response
		body, err := io.ReadAll(dequeueResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem := &v1.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, dequeueItem)).ToNot(HaveOccurred())

		// check item returned
		g.Expect(dequeueItem.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem.ID).To(Equal(uint64(1)))
		g.Expect(dequeueItem.Data).To(Equal([]byte(`hello world`)))
	})

	t.Run("it dequeues TagGroup messages in the proper order", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue multiple items
		message1 := Queue1NotUpdateableEnqueue
		message1.Data = []byte(`first`)
		message2 := Queue1NotUpdateableEnqueue
		message2.Data = []byte(`second`)
		message3 := Queue1NotUpdateableEnqueue
		message3.Data = []byte(`third`)

		g.Expect(testConstruct.ServerClient.WillowEnqueue(g, message1).StatusCode).To(Equal(http.StatusOK))
		g.Expect(testConstruct.ServerClient.WillowEnqueue(g, message2).StatusCode).To(Equal(http.StatusOK))
		g.Expect(testConstruct.ServerClient.WillowEnqueue(g, message3).StatusCode).To(Equal(http.StatusOK))

		// dequeue the item
		dequeueRequest := v1.ReaderSelect{
			BrokerName: "queue1",
			Queries: []v1.ReaderQuery{
				{
					Type: v1.ReaderExactly,
					Tags: datatypes.StringMap{"some": datatypes.String("tag")},
				},
			},
		}

		// 1st response check
		response1 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response1.StatusCode).To(Equal(http.StatusOK))

		body, err := io.ReadAll(response1.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem1 := v1.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem1)).ToNot(HaveOccurred())
		g.Expect(dequeueItem1.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem1.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem1.Data).To(Equal([]byte(`first`)))
		g.Expect(dequeueItem1.ID).To(Equal(uint64(1)))

		// 2nd response check
		response2 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response2.StatusCode).To(Equal(http.StatusOK))

		body, err = io.ReadAll(response2.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem2 := v1.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem2)).ToNot(HaveOccurred())
		g.Expect(dequeueItem2.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem2.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem2.Data).To(Equal([]byte(`second`)))
		g.Expect(dequeueItem2.ID).To(Equal(uint64(2)))

		// 3rd response check
		response3 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response3.StatusCode).To(Equal(http.StatusOK))

		body, err = io.ReadAll(response3.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem3 := v1.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem3)).ToNot(HaveOccurred())
		g.Expect(dequeueItem3.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem3.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem3.Data).To(Equal([]byte(`third`)))
		g.Expect(dequeueItem3.ID).To(Equal(uint64(3)))
	})

	t.Run("when the query criteria is 'Exact'", func(t *testing.T) {
		t.Run("it pulls from the proper queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			// create the queue
			createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue an item
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			// dequeue the item
			dequeueRequest := v1.ReaderSelect{
				BrokerName: "queue1",
				Queries: []v1.ReaderQuery{
					{
						Type: v1.ReaderExactly,
						Tags: datatypes.StringMap{"some": datatypes.String("tag")},
					},
				},
			}
			dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
			g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

			// parse response
			body, err := io.ReadAll(dequeueResponse.Body)
			g.Expect(err).ToNot(HaveOccurred())

			dequeueItem := &v1.DequeueItemResponse{}
			g.Expect(json.Unmarshal(body, dequeueItem)).ToNot(HaveOccurred())

			// check item returned
			g.Expect(dequeueItem.BrokerInfo.Name).To(Equal("queue1"))
			g.Expect(dequeueItem.BrokerInfo.Tags).To(Equal(datatypes.StringMap{"some": datatypes.String("tag")}))
			g.Expect(dequeueItem.ID).To(Equal(uint64(1)))
			g.Expect(dequeueItem.Data).To(Equal([]byte(`hello world`)))
		})
	})

	t.Run("when the query criteria is 'Matches'", func(t *testing.T) {
		t.Run("it pulls from the proper queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			// create the queue
			createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue a few items
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			enqueueCopy := Queue1UpdateableEnqueue
			enqueueCopy.BrokerInfo.Tags = datatypes.StringMap{"some": datatypes.String("tag"), "another": datatypes.String("tag")}
			enqueueCopy.Data = []byte(`some more data to find`)
			enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueCopy)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			enqueueNotFound := Queue1UpdateableEnqueue
			enqueueNotFound.BrokerInfo.Tags = datatypes.StringMap{"not found": datatypes.String("tag")}
			enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueNotFound)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			expectedItem := []v1.DequeueItemResponse{
				{
					BrokerInfo: v1.BrokerInfo{
						Name: "queue1",
						Tags: datatypes.StringMap{"some": datatypes.String("tag")},
					},
					Data: []byte(`hello world`),
					ID:   1,
				},
				{
					BrokerInfo: v1.BrokerInfo{
						Name: "queue1",
						Tags: datatypes.StringMap{"some": datatypes.String("tag"), "another": datatypes.String("tag")},
					},
					Data: []byte(`some more data to find`),
					ID:   1,
				},
			}

			foundOne, foundTwo := false, false
			for i := 1; i <= 2; i++ {
				dequeueRequest := v1.ReaderSelect{
					BrokerName: "queue1",
					Queries: []v1.ReaderQuery{
						{
							Type: v1.ReaderMatches,
							Tags: datatypes.StringMap{"some": datatypes.String("tag")},
						},
					},
				}
				dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
				g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

				// parse response
				body, err := io.ReadAll(dequeueResponse.Body)
				g.Expect(err).ToNot(HaveOccurred())

				dequeueItem := v1.DequeueItemResponse{}
				g.Expect(json.Unmarshal(body, &dequeueItem)).ToNot(HaveOccurred())

				if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[0].BrokerInfo.Tags) {
					g.Expect(dequeueItem).To(Equal(expectedItem[0]))
					foundOne = true
				} else if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[1].BrokerInfo.Tags) {
					g.Expect(dequeueItem).To(Equal(expectedItem[1]))
					foundTwo = true
				} else {
					g.Fail(fmt.Sprintf("found unexpected tags: %v", dequeueItem.BrokerInfo.Tags))
				}
			}

			// check both items returned
			g.Expect(foundOne).To(BeTrue())
			g.Expect(foundTwo).To(BeTrue())
		})
	})

	t.Run("when the query criteria is multiple lookups", func(t *testing.T) {
		t.Run("it pulls from any of the proper queue's tag groups", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			// create the queue
			createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueue a few items
			enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			enqueueCopy := Queue1UpdateableEnqueue
			enqueueCopy.BrokerInfo.Tags = datatypes.StringMap{"some": datatypes.String("tag"), "another": datatypes.String("tag")}
			enqueueCopy.Data = []byte(`some more data to find`)
			enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueCopy)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			enqueueMatchesFound := Queue1UpdateableEnqueue
			enqueueMatchesFound.BrokerInfo.Tags = datatypes.StringMap{"the other unique tag": datatypes.String("ok")}
			enqueueMatchesFound.Data = []byte(`this should be found`)
			enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueMatchesFound)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			expectedItem := []v1.DequeueItemResponse{
				{
					BrokerInfo: v1.BrokerInfo{
						Name: "queue1",
						Tags: datatypes.StringMap{"some": datatypes.String("tag"), "another": datatypes.String("tag")},
					},
					Data: []byte(`some more data to find`),
					ID:   1,
				},
				{
					BrokerInfo: v1.BrokerInfo{
						Name: "queue1",
						Tags: datatypes.StringMap{"the other unique tag": datatypes.String("ok")},
					},
					Data: []byte(`this should be found`),
					ID:   1,
				},
			}

			foundOne, foundTwo := false, false
			for i := 1; i <= 2; i++ {
				dequeueRequest := v1.ReaderSelect{
					BrokerName: "queue1",
					Queries: []v1.ReaderQuery{
						{
							Type: v1.ReaderExactly,
							Tags: datatypes.StringMap{"some": datatypes.String("tag"), "another": datatypes.String("tag")},
						},
						{
							Type: v1.ReaderMatches,
							Tags: datatypes.StringMap{"the other unique tag": datatypes.String("ok")},
						},
					},
				}
				dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
				g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

				// parse response
				body, err := io.ReadAll(dequeueResponse.Body)
				g.Expect(err).ToNot(HaveOccurred())

				dequeueItem := v1.DequeueItemResponse{}
				g.Expect(json.Unmarshal(body, &dequeueItem)).ToNot(HaveOccurred())

				if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[0].BrokerInfo.Tags) {
					g.Expect(dequeueItem).To(Equal(expectedItem[0]))
					foundOne = true
				} else if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[1].BrokerInfo.Tags) {
					g.Expect(dequeueItem).To(Equal(expectedItem[1]))
					foundTwo = true
				} else {
					g.Fail(fmt.Sprintf("found unexpected tags: %v", dequeueItem.BrokerInfo.Tags))
				}
			}

			// check both items returned
			g.Expect(foundOne).To(BeTrue())
			g.Expect(foundTwo).To(BeTrue())
		})
	})
}
