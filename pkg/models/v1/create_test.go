package v1_test

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func TestCreate_ParseCreate_SortsTags(t *testing.T) {
	g := NewGomegaWithT(t)

	create, err := v1.ParseCreate([]byte(`{"BrokerType":0,"BrokerTags":["b", "a"], "QueueParams":{"MaxSize": 5, "RetryCount":2}}`))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(create.BrokerTags).To(Equal([]string{"a", "b"}))
}
