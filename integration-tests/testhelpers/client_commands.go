package testhelpers

import (
	"encoding/json"
	"io"
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"github.com/onsi/gomega"

	. "github.com/onsi/gomega"
)

func (itc *IntegrationTestConstruct) CreateAndGetMessage(g *gomega.WithT, data []byte, tags []string, updateable bool) *v1.DequeueMessage {
	itc.CreateAndEnqueue(g, data, tags, updateable)

	readyBody := v1.Ready{
		BrokerType: v1.Queue,
		MatchQuery: v1.MatchQuery{
			MatchRestriction: v1.STRICT,
			BrokerTags:       []string{"a", "b", "c"},
		},
	}

	messageResponse := itc.GetMessage(g, readyBody)
	g.Expect(messageResponse.StatusCode).To(Equal(http.StatusOK))

	queueItemBody, err := io.ReadAll(messageResponse.Body)
	defer messageResponse.Body.Close()
	g.Expect(err).ToNot(HaveOccurred())

	queueItem := &v1.DequeueMessage{}
	g.Expect(json.Unmarshal(queueItemBody, queueItem)).ToNot(HaveOccurred())

	return queueItem
}

func (itc *IntegrationTestConstruct) CreateAndEnqueue(g *gomega.WithT, data []byte, tags []string, updateable bool) {
	queue1 := v1.Create{
		BrokerType: v1.Queue,
		BrokerTags: tags,
	}
	createResponse := itc.Create(g, queue1)
	g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

	enqueueBody := v1.EnqueMessage{
		BrokerType: v1.Queue,
		BrokerTags: tags,
		Data:       data,
		Updateable: updateable,
	}

	enqueurResponse := itc.Enqueue(g, enqueueBody)
	g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
}
