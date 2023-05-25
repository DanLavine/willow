package v1

// TODO: revisit this. I don't need this right now
//
//import (
//	"bytes"
//	"io/ioutil"
//	"testing"
//
//	. "github.com/onsi/gomega"
//)
//
//func TestQueryInclusive_ParseQuery(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it returns an error on an invalid json syntax", func(t *testing.T) {
//		buffer := ioutil.NopCloser(bytes.NewBufferString(`{"BrokerName":"test"`))
//
//		query, err := ParseQueryInclusive(buffer)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(ContainSubstring("Failed to parse request body"))
//		g.Expect(query).To(BeNil())
//	})
//}
//
//func TestQueryInclusive_Validate_QueryInclusive(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	setup := func() *QueryInclusive {
//		return &QueryInclusive{
//			BrokerName: "test",
//		}
//	}
//
//	t.Run("it returns an error when the BrokerName is not set", func(t *testing.T) {
//		query := setup()
//		query.BrokerName = ""
//
//		err := query.Validate()
//		g.Expect(err).ToNot(BeNil())
//		g.Expect(err.Error()).To(ContainSubstring("BrokerName cannot be empty"))
//	})
//}
//
//func TestQueryInclusive_Validate_InclusiveWhere(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("context validating where clause", func(t *testing.T) {
//		t.Run("it returns an error if both query types are set", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{}, AddWhere: KeyValues{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Only ExactWhere OR AddWhere can be inclueded at a time, not both"))
//		})
//
//		t.Run("it returns an error if no query types are set", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("ExactWhere OR AddWhere must be inclueded in a query"))
//		})
//
//		t.Run("it returns an error if the queries are empty", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Where clause is empty. Needs to be populated"))
//		})
//	})
//
//	t.Run("context when there is no join type", func(t *testing.T) {
//		t.Run("it returns an error JoinWhere is set without a join type", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, JoinWhere: &InclusiveWhere{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Have a join clause, but no join type. Need to have either ['and' | 'or']"))
//		})
//
//		t.Run("it returns an error JoinInclusiveWhere is set without a join type", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, JoinInclusiveWhere: &InclusiveWhere{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Have a join clause, but no join type. Need to have either ['and' | 'or']"))
//		})
//	})
//
//	t.Run("context when a 'Join' type set", func(t *testing.T) {
//		t.Run("it returns an error if Join is an unkown value", func(t *testing.T) {
//			badJoin := Join("other")
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, Join: &badJoin}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Invalid Join type. Must be ['and' | 'or']"))
//		})
//
//		t.Run("it returns an error if both JoinWhere and JoinInclusiveWhere are nil", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, Join: &WhereAnd}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Have a join type, but no join clause. Need to have either ['JoinWhere' | 'JoinInclusiveWhere']"))
//		})
//
//		t.Run("it returns an error if both JointWhere and JoinInclusiveWhere are both set", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, Join: &WhereAnd, JoinWhere: &InclusiveWhere{}, JoinInclusiveWhere: &InclusiveWhere{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("Have a join type, with multiple join clauses"))
//		})
//
//		t.Run("it validates 'JoinWhere'", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, Join: &WhereAnd, JoinWhere: &InclusiveWhere{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("ExactWhere OR AddWhere must be inclueded in a query"))
//		})
//
//		t.Run("it validates 'JoinInclusiveWhere'", func(t *testing.T) {
//			inclusiveWhere := &InclusiveWhere{ExactWhere: KeyValues{"one": "1"}, Join: &WhereAnd, JoinInclusiveWhere: &InclusiveWhere{}}
//
//			err := inclusiveWhere.Validate()
//			g.Expect(err).ToNot(BeNil())
//			g.Expect(err.Error()).To(ContainSubstring("ExactWhere OR AddWhere must be inclueded in a query"))
//		})
//	})
//}
