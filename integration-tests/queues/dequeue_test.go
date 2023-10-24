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

	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Dequeue(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := NewIntrgrationTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("it returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.StartWillow(g)
		defer testConstruct.Shutdown(g)

		// dequeue item request
		tagString := datatypes.String("tag")
		dequeueRequest := v1willow.DequeueItemRequest{
			Name: "queue1",
			Query: query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"some": query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
					},
				},
			},
		}

		dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusNotAcceptable))
	})

	t.Run("it returns a message that is waiting to be processed if they are available", func(t *testing.T) {
		testConstruct.StartWillow(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue an item
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		// dequeue item request
		tagString := datatypes.String("tag")
		oneKey := 1
		dequeueRequest := v1willow.DequeueItemRequest{
			Name: "queue1",
			Query: query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"some": query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
					},
					Limits: &query.KeyLimits{
						NumberOfKeys: &oneKey,
					},
				},
			},
		}

		dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

		// parse response
		body, err := io.ReadAll(dequeueResponse.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem := &v1willow.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, dequeueItem)).ToNot(HaveOccurred())

		// check item returned
		g.Expect(dequeueItem.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem.Data).To(Equal([]byte(`hello world`)))
	})

	t.Run("it can recieve a message enqueued after the dequeue request", func(t *testing.T) {
		testConstruct.StartWillow(g)
		defer testConstruct.Shutdown(g)

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// dequeue item request
		tagString := datatypes.String("tag")
		oneKey := 1
		dequeueRequest := v1willow.DequeueItemRequest{
			Name: "queue1",
			Query: query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"some": query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
					},
					Limits: &query.KeyLimits{
						NumberOfKeys: &oneKey,
					},
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

		dequeueItem := &v1willow.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, dequeueItem)).ToNot(HaveOccurred())

		// check item returned
		g.Expect(dequeueItem.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem.Data).To(Equal([]byte(`hello world`)))
	})

	t.Run("it dequeues TagGroup messages in the proper order", func(t *testing.T) {
		testConstruct.StartWillow(g)
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

		// dequeue item request
		tagString := datatypes.String("tag")
		oneKey := 1
		dequeueRequest := v1willow.DequeueItemRequest{
			Name: "queue1",
			Query: query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"some": query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
					},
					Limits: &query.KeyLimits{
						NumberOfKeys: &oneKey,
					},
				},
			},
		}

		// 1st response check
		response1 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response1.StatusCode).To(Equal(http.StatusOK))

		body, err := io.ReadAll(response1.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem1 := v1willow.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem1)).ToNot(HaveOccurred())
		g.Expect(dequeueItem1.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem1.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem1.Data).To(Equal([]byte(`first`)))

		// 2nd response check
		response2 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response2.StatusCode).To(Equal(http.StatusOK))

		body, err = io.ReadAll(response2.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem2 := v1willow.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem2)).ToNot(HaveOccurred())
		g.Expect(dequeueItem2.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem2.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem2.Data).To(Equal([]byte(`second`)))

		// 3rd response check
		response3 := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
		g.Expect(response3.StatusCode).To(Equal(http.StatusOK))

		body, err = io.ReadAll(response3.Body)
		g.Expect(err).ToNot(HaveOccurred())

		dequeueItem3 := v1willow.DequeueItemResponse{}
		g.Expect(json.Unmarshal(body, &dequeueItem3)).ToNot(HaveOccurred())
		g.Expect(dequeueItem3.BrokerInfo.Name).To(Equal("queue1"))
		g.Expect(dequeueItem3.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"some": datatypes.String("tag")}))
		g.Expect(dequeueItem3.Data).To(Equal([]byte(`third`)))
	})

	t.Run("it respects the query criterian and pulls from the proper tag groups", func(t *testing.T) {
		testConstruct.StartWillow(g)
		defer testConstruct.Shutdown(g)

		defer func() {
			fmt.Println(string(testConstruct.Session.Out.Contents()))
			fmt.Println(string(testConstruct.Session.Err.Contents()))
		}()

		// create the queue
		createResponse := testConstruct.ServerClient.WillowCreate(g, Queue1)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// enqueue a few items
		enqueueResponse := testConstruct.ServerClient.WillowEnqueue(g, Queue1UpdateableEnqueue)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		enqueueCopy := Queue1UpdateableEnqueue
		enqueueCopy.BrokerInfo.Tags = datatypes.KeyValues{"some": datatypes.String("tag"), "another": datatypes.String("tag")}
		enqueueCopy.Data = []byte(`some more data to find`)
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueCopy)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		enqueueMatchesFound := Queue1UpdateableEnqueue
		enqueueMatchesFound.BrokerInfo.Tags = datatypes.KeyValues{"the other unique tag": datatypes.String("ok")}
		enqueueMatchesFound.Data = []byte(`this should be found`)
		enqueueResponse = testConstruct.ServerClient.WillowEnqueue(g, enqueueMatchesFound)
		g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

		expectedItem := []v1willow.DequeueItemResponse{
			{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "queue1",
					Tags: datatypes.KeyValues{"some": datatypes.String("tag"), "another": datatypes.String("tag")},
				},
				Data: []byte(`some more data to find`),
			},
			{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "queue1",
					Tags: datatypes.KeyValues{"the other unique tag": datatypes.String("ok")},
				},
				Data: []byte(`this should be found`),
			},
		}

		foundOne, foundTwo := false, false
		for i := 1; i <= 2; i++ {
			tagString := datatypes.String("tag")
			notokString := datatypes.String("not ok")
			twoKey := 2
			trueValue := true
			dequeueRequest := v1willow.DequeueItemRequest{
				Name: "queue1",
				Query: query.AssociatedKeyValuesQuery{
					Or: []query.AssociatedKeyValuesQuery{
						{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"some":    query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
									"another": query.Value{Value: &tagString, ValueComparison: query.EqualsPtr()},
								},
								Limits: &query.KeyLimits{
									NumberOfKeys: &twoKey,
								},
							},
						},
						{
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{
											"the other unique tag": query.Value{Value: &notokString, ValueComparison: query.NotEqualsPtr()},
										},
									},
								},
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{
											"the other unique tag": query.Value{Exists: &trueValue},
										},
									},
								},
							},
						},
					},
				},
			}

			dequeueResponse := testConstruct.ServerClient.WillowDequeue(g, dequeueRequest)
			g.Expect(dequeueResponse.StatusCode).To(Equal(http.StatusOK))

			// parse response
			body, err := io.ReadAll(dequeueResponse.Body)
			g.Expect(err).ToNot(HaveOccurred())

			dequeueItem := v1willow.DequeueItemResponse{}
			g.Expect(json.Unmarshal(body, &dequeueItem)).ToNot(HaveOccurred())

			if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[0].BrokerInfo.Tags) {
				expectedItem[0].ID = dequeueItem.ID
				g.Expect(dequeueItem).To(Equal(expectedItem[0]))
				foundOne = true
			} else if reflect.DeepEqual(dequeueItem.BrokerInfo.Tags, expectedItem[1].BrokerInfo.Tags) {
				expectedItem[1].ID = dequeueItem.ID
				g.Expect(dequeueItem).To(Equal(expectedItem[1]))
				foundTwo = true
			} else {
				fmt.Printf("broker info: %#v\n", dequeueItem.BrokerInfo)
				g.Fail(fmt.Sprintf("found unexpected tags: %v", dequeueItem.BrokerInfo.Tags))
			}

			fmt.Printf("Dequeued Item: %#v\n", dequeueItem)
		}

		// check both items returned
		g.Expect(foundOne).To(BeTrue())
		g.Expect(foundTwo).To(BeTrue())
	})
}
