package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"github.com/onsi/gomega"

	. "github.com/onsi/gomega"
)

func (itc *IntegrationTestConstruct) Create(g *gomega.WithT, createBody v1.Create) *http.Response {
	body, err := json.Marshal(createBody)
	g.Expect(err).ToNot(HaveOccurred())

	createBuffer := bytes.NewBuffer(body)
	createRequest, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/brokers/queues/create", itc.serverAddress), createBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.ServerClient.Do(createRequest)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (itc *IntegrationTestConstruct) Enqueue(g *gomega.WithT, enqueueBody v1.EnqueueItemRequest) *http.Response {
	body, err := json.Marshal(enqueueBody)
	g.Expect(err).ToNot(HaveOccurred())

	enqueueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/brokers/item/enqueue", itc.serverAddress), enqueueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.ServerClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (itc *IntegrationTestConstruct) Dequeue(g *gomega.WithT, readerQuery v1.ReaderSelect) *http.Response {
	body, err := json.Marshal(readerQuery)
	g.Expect(err).ToNot(HaveOccurred())

	dequeueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/brokers/item/dequeue", itc.serverAddress), dequeueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	responseChan := make(chan *http.Response)
	errChan := make(chan error)
	go func() {
		response, err := itc.ServerClient.Do(request)
		errChan <- err
		responseChan <- response
	}()

	g.Eventually(errChan).Should(Receive(BeNil()))

	select {
	case <-time.After(time.Second):
		g.Fail("didn't recieve a dequeue message")
		return nil
	case response := <-responseChan:
		return response
	}
}

func (itc *IntegrationTestConstruct) ACKMessage(g *gomega.WithT, ackBody v1.ACK) *http.Response {
	body, err := json.Marshal(ackBody)
	g.Expect(err).ToNot(HaveOccurred())

	enqueueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/message/ack", itc.serverAddress), enqueueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.ServerClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (itc *IntegrationTestConstruct) Metrics(g *gomega.WithT) v1.MetricsResponse {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/metrics", itc.metricsAddress), nil)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.MetricsClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(http.StatusOK))

	body := response.Body
	defer body.Close()

	metricsData, err := io.ReadAll(body)
	g.Expect(err).ToNot(HaveOccurred())

	metrics := v1.MetricsResponse{}
	g.Expect(json.Unmarshal(metricsData, &metrics)).ToNot(HaveOccurred())

	return metrics
}
