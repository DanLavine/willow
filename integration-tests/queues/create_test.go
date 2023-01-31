package queues_integration_tests

import (
	"net/http"
	"testing"

	"github.com/DanLavine/willow/integration-tests/testhelpers"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func Test_Create(t *testing.T) {
	g := NewGomegaWithT(t)
	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)

	t.Run("can create a queue with no tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			BrokerType: v1.Queue,
			BrokerTags: nil,
		}

		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(BeNil())
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("can create a queue with empty tags", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			BrokerType: v1.Queue,
			BrokerTags: []string{},
		}

		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("will create a new queue", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			BrokerType: v1.Queue,
			BrokerTags: []string{"a", "b", "c"},
		}

		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"a", "b", "c"}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("can create the same queue multple times", func(t *testing.T) {
		testConstruct.Start(g)
		defer testConstruct.Shutdown(g)

		createBody := v1.Create{
			BrokerType: v1.Queue,
			BrokerTags: []string{"a", "b", "c"},
		}

		// first create
		createResponse := testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		// second create
		createResponse = testConstruct.Create(g, createBody)
		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))

		metrics := testConstruct.Metrics(g)
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"a", "b", "c"}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})
}
