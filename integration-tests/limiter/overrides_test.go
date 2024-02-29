package limter_integration_tests

import (
	"fmt"
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Limiter_Overrides_Create(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can create an override for a rule", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
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

func Test_Limiter_Overrides_Get(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can get an override by name", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
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

		// create a few override
		for i := 0; i < 17; i++ {
			override := &v1.Override{
				Name:  fmt.Sprintf("override%d", i),
				Limit: int64(i),
				KeyValues: datatypes.KeyValues{
					"key1":  datatypes.Int(1),
					"key2":  datatypes.Int(2),
					"other": datatypes.Int(i),
				},
			}

			err = limiterClient.CreateOverride("rule1", override)
			g.Expect(err).ToNot(HaveOccurred())
		}

		// get an overrid by name
		foundOverride, err := limiterClient.GetOverride("rule1", "override12")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundOverride).ToNot(BeNil())
		g.Expect(foundOverride.Name).To(Equal("override12"))
		g.Expect(foundOverride.Limit).To(Equal(int64(12)))
		g.Expect(foundOverride.KeyValues).To(Equal(datatypes.KeyValues{
			"key1":  datatypes.Int(1),
			"key2":  datatypes.Int(2),
			"other": datatypes.Int(12),
		}))
	})
}

func Test_Limiter_Overrides_Update(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can update an override by name", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
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

		// update override
		overrideUpdate := &v1.OverrideUpdate{
			Limit: 18,
		}
		err = limiterClient.UpdateOverride("rule1", "override1", overrideUpdate)
		g.Expect(err).ToNot(HaveOccurred())

		// check the overrid
		foundOverride, err := limiterClient.GetOverride("rule1", "override1")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundOverride).ToNot(BeNil())
		g.Expect(foundOverride.Limit).To(Equal(int64(18)))
	})
}

func Test_Limiter_Overrides_Delete(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can delete an override for a rule", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
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
		overrideResp, err := limiterClient.GetOverride("rule1", "override1")

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Override 'override1' not found"))
		g.Expect(overrideResp).To(BeNil())
	})
}

func Test_Limiter_Overrides_Match(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can list the overrides that match the query", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
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
		query := &v1common.MatchQuery{
			KeyValues: &datatypes.KeyValues{
				"key1":           datatypes.Int(1),
				"key2":           datatypes.Int(2),
				"other":          datatypes.Float32(32),
				"doesn't matter": datatypes.String("hoopla"),
			},
		}

		overrides, err := limiterClient.MatchOverrides("rule1", query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(1))
	})
}
