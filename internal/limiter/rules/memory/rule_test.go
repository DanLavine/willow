package memory

import (
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/onsi/gomega"
)

func Test_Limit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It sets the limit properly", func(t *testing.T) {
		ruleProperties := &v1limiter.RuleProperties{Limit: helpers.PointerOf[int64](56)}
		g.Expect(ruleProperties.Validate()).ToNot(HaveOccurred())

		rule := New(ruleProperties)
		g.Expect(rule.Limit()).To(Equal(int64(56)))
	})
}

func Test_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It updates the limit properly", func(t *testing.T) {
		// create initial rule
		ruleProperties := &v1limiter.RuleProperties{Limit: helpers.PointerOf[int64](56)}
		g.Expect(ruleProperties.Validate()).ToNot(HaveOccurred())

		rule := New(ruleProperties)
		g.Expect(rule.Limit()).To(Equal(int64(56)))

		// update rule
		ruleUpdateRequest := &v1limiter.RuleProperties{
			Limit: helpers.PointerOf[int64](12),
		}
		g.Expect(ruleUpdateRequest.Validate()).ToNot(HaveOccurred())

		g.Expect(rule.Update(ruleUpdateRequest)).ToNot(HaveOccurred())
		g.Expect(rule.Limit()).To(Equal(int64(12)))
	})
}

func Test_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op", func(t *testing.T) {
		// create initial rule
		ruleProperties := &v1limiter.RuleProperties{Limit: helpers.PointerOf[int64](56)}
		g.Expect(ruleProperties.Validate()).ToNot(HaveOccurred())

		rule := New(ruleProperties)
		g.Expect(rule.Limit()).To(Equal(int64(56)))

		// delete
		g.Expect(rule.Delete()).ToNot(HaveOccurred())
	})
}
