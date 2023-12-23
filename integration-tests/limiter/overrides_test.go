package limter_integration_tests

import (
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"key1":  datatypes.Int(1),
				"key2":  datatypes.Int(2),
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
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(limiterClient.CreateRule(rule)).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"key1":  datatypes.Int(1),
				"key2":  datatypes.Int(2),
				"other": datatypes.Float32(32),
			},
		}
		g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())

		// delete override
		err := limiterClient.DeleteOverride("rule1", "override1")
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule with overrides to ensure it is deleted
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleQuery{OverridesToInclude: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.Overrides).To(BeNil())
	})
}

func Test_Limiter_Overrides_List(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can list the overrides that match the query", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)
		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create rule
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create overrides
		override1 := &v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"key1":  datatypes.Int(1),
				"key2":  datatypes.Int(2),
				"other": datatypes.Float32(32),
			},
		}
		g.Expect(limiterClient.CreateOverride("rule1", override1)).ToNot(HaveOccurred())

		override2 := &v1.Override{
			Name:  "override2",
			Limit: 18,
			KeyValues: datatypes.KeyValues{
				"key1":  datatypes.String("other"),
				"key2":  datatypes.Int(2),
				"other": datatypes.Float32(32),
			},
		}
		g.Expect(limiterClient.CreateOverride("rule1", override2)).ToNot(HaveOccurred())

		// query
		existsTrue := true
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key1": datatypes.Value{Exists: &existsTrue, ExistsType: &datatypes.T_string},
					},
				},
			},
		}

		overrides, err := limiterClient.ListOverrides("rule1", query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(1))
	})
}
