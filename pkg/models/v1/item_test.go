package v1

//import (
//	"bytes"
//	"io"
//	"testing"
//
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//	. "github.com/onsi/gomega"
//)

//func TestEnqueueItemRequest_ParseEnqueueItemRequest(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it has properly sorted Tags", func(t *testing.T) {
//		item, err := ParseEnqueueItemRequest(io.NopCloser(bytes.NewBufferString(`{"BrokerInfo": {"Name":"test","Tags": ["b", "a"]}}`)))
//		g.Expect(err).ToNot(HaveOccurred())
//
//		g.Expect(item.BrokerInfo.Tags).To(Equal(datatypes.Strings{"a", "b"}))
//	})
//}
