package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/onsi/gomega"
)

func Test_ACK(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)

	t.Run("when setting success to TRUE", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			testConstruct.Start(g)
			defer testConstruct.Shutdown(g)

			queueItem := testConstruct.CreateAndGetMessage(g, []byte(`hello world`), []string{"a", "b", "c"}, true)

			ackBody := v1.ACK{
				ID:         queueItem.ID,
				BrokerType: v1.Queue,
				BrokerTags: queueItem.BrokerTags,
				Passed:     true,
			}
			response := testConstruct.ACKMessage(g, ackBody)
			g.Expect(response.StatusCode).To(Equal(http.StatusOK))

			// we can look at metrics to see that the item has been removed
			metrics := testConstruct.Metrics(g)
			g.Expect(len(metrics.Queues)).To(Equal(1))
			g.Expect(metrics.Queues[0].Tags).To(Equal(queueItem.BrokerTags))
			g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
		})
	})

	t.Run("when setting success to FALSE", func(t *testing.T) {

	})
}
