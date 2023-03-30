package queues_integration_tests

//import (
//	"encoding/json"
//	"io"
//	"net/http"
//	"testing"
//
//	"github.com/DanLavine/willow/integration-tests/testhelpers"
//	v1 "github.com/DanLavine/willow/pkg/models/v1"
//
//	. "github.com/onsi/gomega"
//)
//
//func Test_Messages(t *testing.T) {
//	g := NewGomegaWithT(t)
//	testConstruct := testhelpers.NewIntrgrationTestConstruct(g)
//
//	t.Run("when match restrictins is STRICT", func(t *testing.T) {
//		t.Run("returns an error if the queue does not exist", func(t *testing.T) {
//			testConstruct.Start(g)
//			defer testConstruct.Shutdown(g)
//
//			readyBody := v1.Ready{
//				BrokerType: v1.Queue,
//				MatchQuery: v1.MatchQuery{
//					MatchRestriction: v1.STRICT,
//					BrokerTags:       []string{"a", "b"},
//				},
//			}
//
//			messageResponse := testConstruct.GetMessage(g, readyBody)
//			g.Expect(messageResponse.StatusCode).To(Equal(http.StatusBadRequest))
//		})
//
//		t.Run("returns a message from the desired queue", func(t *testing.T) {
//			testConstruct.Start(g)
//			defer testConstruct.Shutdown(g)
//
//			queue1 := v1.Create{
//				BrokerType: v1.Queue,
//				BrokerTags: []string{"a", "b", "c"},
//			}
//			createResponse := testConstruct.Create(g, queue1)
//			g.Expect(createResponse.StatusCode).To(Equal(http.StatusCreated))
//
//			enqueueBody := v1.EnqueMessage{
//				BrokerType: v1.Queue,
//				BrokerTags: []string{"a", "b", "c"},
//				Data:       []byte(`hello world`),
//				Updateable: false,
//			}
//
//			enqueurResponse := testConstruct.Enqueue(g, enqueueBody)
//			g.Expect(enqueurResponse.StatusCode).To(Equal(http.StatusOK))
//
//			readyBody := v1.Ready{
//				BrokerType: v1.Queue,
//				MatchQuery: v1.MatchQuery{
//					MatchRestriction: v1.STRICT,
//					BrokerTags:       []string{"a", "b", "c"},
//				},
//			}
//
//			messageResponse := testConstruct.GetMessage(g, readyBody)
//			g.Expect(messageResponse.StatusCode).To(Equal(http.StatusOK))
//
//			queueItemBody, err := io.ReadAll(messageResponse.Body)
//			defer messageResponse.Body.Close()
//			g.Expect(err).ToNot(HaveOccurred())
//
//			var queueItem v1.DequeueMessage
//			g.Expect(json.Unmarshal(queueItemBody, &queueItem)).ToNot(HaveOccurred())
//			g.Expect(queueItem.ID).To(Equal(uint64(1)))
//			g.Expect(queueItem.Data).To(Equal([]byte(`hello world`)))
//			g.Expect(queueItem.BrokerTags).To(Equal([]string{"a", "b", "c"}))
//
//			metrics := testConstruct.Metrics(g)
//			g.Expect(len(metrics.Queues)).To(Equal(1))
//			g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"a", "b", "c"}))
//			g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
//			g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(1)))
//		})
//	})
//
//	t.Run("when match restrictins is 'SUBSET'", func(t *testing.T) {
//
//	})
//
//	t.Run("when match restrictins is 'ANY'", func(t *testing.T) {
//
//	})
//
//	t.Run("when match restrictins is 'ALL'", func(t *testing.T) {
//
//	})
//}
