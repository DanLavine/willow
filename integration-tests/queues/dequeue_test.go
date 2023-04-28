package queues_integration_tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/onsi/gomega"
)

func Test_Dequeue(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)

	t.Run("when match restrictins is STRICT", func(t *testing.T) {
		t.Run("returns a message from the desired queue", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			// create 2 queues
			createResponse := testConstruct.Create(g, testhelpers.Queue1)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			createResponse = testConstruct.Create(g, testhelpers.Queue2)
			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

			// enqueu an 2 item on each queue with different tags
			//// queue 1
			item1 := testhelpers.Queue1UpdateableEnqueue
			enqueueResponse := testConstruct.Enqueue(g, item1)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			item1.BrokerInfo.Tags = datatypes.Strings{"a", "b", "c"}
			item1.Data = []byte(`retrieve me`)
			enqueueResponse = testConstruct.Enqueue(g, item1)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			//// queue 2
			item2 := testhelpers.Queue2UpdateableEnqueue
			enqueueResponse = testConstruct.Enqueue(g, item2)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			item2.BrokerInfo.Tags = datatypes.Strings{"a", "b", "c"}
			item2.Data = []byte(`dont retrieve me`)
			enqueueResponse = testConstruct.Enqueue(g, item2)
			g.Expect(enqueueResponse.StatusCode).To(Equal(http.StatusOK))

			// retrieve a message
			matchBody := v1.MatchQuery{
				BrokerName:            datatypes.String("queue1"),
				MatchTagsRestrictions: v1.STRICT,
				Tags:                  datatypes.Strings{"a", "b", "c"},
			}

			messageResponse := testConstruct.GetItem(g, matchBody)
			g.Expect(messageResponse.StatusCode).To(Equal(http.StatusOK))

			queueItemBody, err := io.ReadAll(messageResponse.Body)
			defer messageResponse.Body.Close()
			g.Expect(err).ToNot(HaveOccurred())

			var queueItem v1.DequeueItemResponse
			g.Expect(json.Unmarshal(queueItemBody, &queueItem)).ToNot(HaveOccurred())
			g.Expect(queueItem.ID).To(Equal(uint64(1)))
			g.Expect(queueItem.Data).To(Equal([]byte(`retrieve me`)))
			g.Expect(queueItem.BrokerInfo.Tags).To(Equal(datatypes.Strings{"a", "b", "c"}))
		})
	})

	t.Run("when match restrictins is 'SUBSET'", func(t *testing.T) {

	})

	t.Run("when match restrictins is 'ANY'", func(t *testing.T) {

	})

	t.Run("when match restrictins is 'ALL'", func(t *testing.T) {

	})
}
