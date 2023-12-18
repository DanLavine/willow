package testclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	. "github.com/onsi/gomega"
)

func (c *Client) WillowCreate(g *WithT, createBody v1willow.Create) *http.Response {
	body, err := json.Marshal(createBody)
	g.Expect(err).ToNot(HaveOccurred())

	createBuffer := bytes.NewBuffer(body)
	createRequest, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/brokers/queues/create", c.address), createBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := c.httpClient.Do(createRequest)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (c *Client) WillowEnqueue(g *WithT, enqueueBody v1willow.EnqueueItemRequest) *http.Response {
	body, err := json.Marshal(enqueueBody)
	g.Expect(err).ToNot(HaveOccurred())

	enqueueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/brokers/item/enqueue", c.address), enqueueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := c.httpClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (c *Client) WillowDequeue(g *WithT, readerQuery v1willow.DequeueItemRequest) *http.Response {
	body, err := json.Marshal(readerQuery)
	g.Expect(err).ToNot(HaveOccurred())

	dequeueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/brokers/item/dequeue", c.address), dequeueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	responseChan := make(chan *http.Response)
	errChan := make(chan error)
	go func() {
		response, err := c.httpClient.Do(request)
		errChan <- err
		responseChan <- response
	}()

	for {
		select {
		case err := <-errChan:
			g.Expect(err).ToNot(HaveOccurred())
		case <-time.After(time.Second):
			g.Fail("didn't recieve a dequeue message")
			return nil
		case response := <-responseChan:
			return response
		}
	}
}

func (c *Client) WillowACK(g *WithT, ackBody v1willow.ACK) *http.Response {
	body, err := json.Marshal(ackBody)
	g.Expect(err).ToNot(HaveOccurred())

	requestBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/brokers/item/ack", c.address), requestBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := c.httpClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (c *Client) WillowMetrics(g *WithT) v1willow.MetricsResponse {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/metrics", c.metricsAddress), nil)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := c.metricsClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(http.StatusOK))

	body := response.Body
	defer body.Close()

	metricsData, err := io.ReadAll(body)
	g.Expect(err).ToNot(HaveOccurred())

	metrics := v1willow.MetricsResponse{}
	g.Expect(json.Unmarshal(metricsData, &metrics)).ToNot(HaveOccurred())

	return metrics
}

func (c *Client) CloseClient() {
	c.httpClient.CloseIdleConnections()
	c.metricsClient.CloseIdleConnections()
}
