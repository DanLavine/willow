package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	. "github.com/onsi/gomega"
)

func Test_Enqueu(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)

	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		enqueueBody := v1.EnqueMessage{
			BrokerType: v1.Queue,
			BrokerTags: []string{"a"},
			Data:       []byte(`hello world`),
			Updateable: false,
		}

		createResponse := testConstruct.Enqueue(g, enqueueBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusBadRequest))
	})
}
