package memory

import (
	"testing"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	. "github.com/onsi/gomega"
)

func Test_New(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It sets the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.Rule{
			Name:             "test",
			GroupByKeyValues: datatypes.KeyValues{"key1": datatypes.Any()},
			Limit:            56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())

		rule := New(ruleCreateRequest)
		g.Expect(rule.limit.Load()).To(Equal(int64(56)))
	})
}

func Test_Limit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.Rule{
			Name:             "test",
			GroupByKeyValues: datatypes.KeyValues{"key1": datatypes.Any()},
			Limit:            56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())

		rule := New(ruleCreateRequest)
		g.Expect(rule.Limit()).To(Equal(int64(56)))
	})
}

func Test_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It updates the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.Rule{
			Name:             "test",
			GroupByKeyValues: datatypes.KeyValues{"key1": datatypes.Any()},
			Limit:            56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())
		rule := New(ruleCreateRequest)

		ruleUpdateRequest := &v1limiter.RuleUpdateRquest{
			Limit: 12,
		}
		g.Expect(ruleUpdateRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rule.Update(zap.NewNop(), ruleUpdateRequest)).ToNot(HaveOccurred())
		g.Expect(rule.Limit()).To(Equal(int64(12)))
	})
}

func Test_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.Rule{
			Name:             "test",
			GroupByKeyValues: datatypes.KeyValues{"key1": datatypes.Any()},
			Limit:            56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())
		rule := New(ruleCreateRequest)

		g.Expect(rule.Delete()).ToNot(HaveOccurred())
	})
}
