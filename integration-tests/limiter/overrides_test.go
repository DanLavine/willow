package limter_integration_tests

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Limiter_Overrides_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can create an override for a rule", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)
		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"other": datatypes.Float32(32),
			},
		}

		err = limiterClient.CreateOverride("rule1", override)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func Test_Limiter_Overrides_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can delete an override for a rule", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(limiterClient.CreateRule(rule)).ToNot(HaveOccurred())

		// create override
		override := v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"other": datatypes.Float32(32),
			},
		}
		g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())

		// delete override
		err := limiterClient.DeleteOverride("rule1", "override1")
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule with overrides to ensure it is deleted
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.Overrides).To(BeNil())
	})
}
