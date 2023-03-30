package queues_integration_tests

//import (
//	"net/http"
//	"testing"
//
//	"github.com/DanLavine/willow/integration-tests/testhelpers"
//	v1 "github.com/DanLavine/willow/pkg/models/v1"
//
//	. "github.com/onsi/gomega"
//)
//
//func Test_Enqueue(t *testing.T) {
//	g := NewGomegaWithT(t)
//	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)
//
//	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
//		testConstruct.Start(g)
//		defer testConstruct.Shutdown(g)
//
//		enqueueBody := v1.EnqueMessage{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a"},
//			Data:       []byte(`hello world`),
//			Updateable: false,
//		}
//
//		createResponse := testConstruct.Enqueue(g, enqueueBody)
//		g.Expect(createResponse.StatusCode).To(Equal(http.StatusBadRequest))
//	})
//
//	t.Run("enqueus a message on the matching tags", func(t *testing.T) {
//		testConstruct.Start(g)
//		defer testConstruct.Shutdown(g)
//
//		queue1 := v1.Create{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a", "b", "c"},
//		}
//		createResponse := testConstruct.Create(g, queue1)
//		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))
//
//		queue2 := v1.Create{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a", "b", "c", "d"},
//		}
//		createResponse = testConstruct.Create(g, queue2)
//		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))
//
//		enqueueBody := v1.EnqueMessage{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a", "b", "c"},
//			Data:       []byte(`hello world`),
//			Updateable: false,
//		}
//
//		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
//		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
//
//		metrics := testConstruct.Metrics(g)
//		g.Expect(len(metrics.Queues)).To(Equal(2))
//		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"a", "b", "c"}))
//		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(1)))
//		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
//	})
//
//	t.Run("enqueus multiple messages", func(t *testing.T) {
//		testConstruct.Start(g)
//		defer testConstruct.Shutdown(g)
//
//		queue1 := v1.Create{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a", "b", "c"},
//		}
//		createResponse := testConstruct.Create(g, queue1)
//		g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))
//
//		enqueueBody := v1.EnqueMessage{
//			BrokerType: v1.Queue,
//			BrokerTags: []string{"a", "b", "c"},
//			Data:       []byte(`hello world`),
//			Updateable: false,
//		}
//
//		enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
//		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
//
//		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
//		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
//
//		enqueurResponse = testConstruct.Enqueue(g, enqueueBody)
//		g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
//
//		metrics := testConstruct.Metrics(g)
//		g.Expect(len(metrics.Queues)).To(Equal(1))
//		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"a", "b", "c"}))
//		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(3)))
//		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
//	})
//}
