package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (itc *IntegrationTestConstruct) Enqueue(g *gomega.WithT, enqueueBody v1.EnqueMessage) *http.Response {
	body, err := json.Marshal(enqueueBody)
	g.Expect(err).ToNot(HaveOccurred())

	enqueueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/message/add", itc.serverAddress), enqueueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.ServerClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
}

func (itc *IntegrationTestConstruct) GetMessage(g *gomega.WithT, readyBody v1.Ready) *http.Response {
	body, err := json.Marshal(readyBody)
	g.Expect(err).ToNot(HaveOccurred())

	enqueueBuffer := bytes.NewBuffer(body)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/message/ready", itc.serverAddress), enqueueBuffer)
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.ServerClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())

	return response
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

func (itc *IntegrationTestConstruct) Metrics(g *gomega.WithT) v1.Metrics {
	matchQuery := v1.MatchQuery{MatchRestriction: v1.ALL}

	matchBody, err := json.Marshal(matchQuery)
	g.Expect(err).ToNot(HaveOccurred())

	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/metrics", itc.metricsAddress), bytes.NewBuffer(matchBody))
	g.Expect(err).ToNot(HaveOccurred())

	response, err := itc.MetricsClient.Do(request)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(http.StatusOK))

	body := response.Body
	defer body.Close()

	metricsData, err := io.ReadAll(body)
	g.Expect(err).ToNot(HaveOccurred())

	metrics := v1.Metrics{}
	g.Expect(json.Unmarshal(metricsData, &metrics)).ToNot(HaveOccurred())

	return metrics
}
