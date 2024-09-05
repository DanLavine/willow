package limter_integration_tests

import (
	"context"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

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
		rule := &v1.Rule{
			Spec: &v1.RuleSpec{
				DBDefinition: &v1.RuleDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Any(),
						"key2": datatypes.Any(),
					},
				},
				Properties: &v1.RuleProperties{
					Limit: helpers.PointerOf[int64](5),
				},
			},
		}

		ruleResp, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.Int(1),
						"key2":  datatypes.Int(2),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](32),
				},
			},
		}

		override, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(override).ToNot(BeNil())
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
		rule := &v1.Rule{
			Spec: &v1.RuleSpec{
				DBDefinition: &v1.RuleDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Any(),
						"key2": datatypes.Any(),
					},
				},
				Properties: &v1.RuleProperties{
					Limit: helpers.PointerOf[int64](5),
				},
			},
		}

		ruleResp, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create a few override
		overrides := []*v1.Override{}
		for i := 0; i < 17; i++ {
			override := &v1.Override{
				Spec: &v1.OverrideSpec{
					DBDefinition: &v1.OverrideDBDefinition{
						GroupByKeyValues: datatypes.KeyValues{
							"key1":  datatypes.Int(1),
							"key2":  datatypes.Int(2),
							"other": datatypes.Int(i),
						},
					},
					Properties: &v1.OverrideProperties{
						Limit: helpers.PointerOf[int64](int64(i)),
					},
				},
			}

			override, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override)
			g.Expect(err).ToNot(HaveOccurred())
			overrides = append(overrides, override)
		}

		// get an overrid by name
		foundOverride, err := limiterClient.GetOverride(context.Background(), ruleResp.State.ID, overrides[12].State.ID)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundOverride).ToNot(BeNil())
		g.Expect(*foundOverride.Spec.Properties.Limit).To(Equal(int64(12)))
		g.Expect(foundOverride.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{
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
		rule := &v1.Rule{
			Spec: &v1.RuleSpec{
				DBDefinition: &v1.RuleDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Any(),
						"key2": datatypes.Any(),
					},
				},
				Properties: &v1.RuleProperties{
					Limit: helpers.PointerOf[int64](5),
				},
			},
		}

		ruleResp, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.Int(1),
						"key2":  datatypes.Int(2),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](32),
				},
			},
		}
		override, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override)
		g.Expect(err).ToNot(HaveOccurred())

		// update override
		overrideUpdate := &v1.OverrideProperties{
			Limit: helpers.PointerOf[int64](18),
		}
		err = limiterClient.UpdateOverride(context.Background(), ruleResp.State.ID, override.State.ID, overrideUpdate)
		g.Expect(err).ToNot(HaveOccurred())

		// check the overrid
		foundOverride, err := limiterClient.GetOverride(context.Background(), ruleResp.State.ID, override.State.ID)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundOverride).ToNot(BeNil())
		g.Expect(*foundOverride.Spec.Properties.Limit).To(Equal(int64(18)))
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
		rule := &v1.Rule{
			Spec: &v1.RuleSpec{
				DBDefinition: &v1.RuleDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Any(),
						"key2": datatypes.Any(),
					},
				},
				Properties: &v1.RuleProperties{
					Limit: helpers.PointerOf[int64](5),
				},
			},
		}
		ruleResp, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.Int(1),
						"key2":  datatypes.Int(2),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](32),
				},
			},
		}
		overrideResp, err := limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override)
		g.Expect(err).ToNot(HaveOccurred())

		// delete override
		err = limiterClient.DeleteOverride(context.Background(), ruleResp.State.ID, overrideResp.State.ID)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule with overrides to ensure it is deleted
		overrideResp, err = limiterClient.GetOverride(context.Background(), ruleResp.State.ID, overrideResp.State.ID)
		g.Expect(err).To(HaveOccurred())
		g.Expect(overrideResp).To(BeNil())
	})
}

func Test_Limiter_Overrides_Query(t *testing.T) {
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
		rule := &v1.Rule{
			Spec: &v1.RuleSpec{
				DBDefinition: &v1.RuleDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Any(),
						"key2": datatypes.Any(),
					},
				},
				Properties: &v1.RuleProperties{
					Limit: helpers.PointerOf[int64](5),
				},
			},
		}

		ruleResp, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create overrides
		override1 := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.Int(1),
						"key2":  datatypes.Int(2),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](32),
				},
			},
		}
		_, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override1)
		g.Expect(err).ToNot(HaveOccurred())

		override2 := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.String("other"),
						"key2":  datatypes.Int(2),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](18),
				},
			},
		}
		_, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override2)
		g.Expect(err).ToNot(HaveOccurred())

		override3 := &v1.Override{
			Spec: &v1.OverrideSpec{
				DBDefinition: &v1.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1":  datatypes.String("other"),
						"key2":  datatypes.Int(3),
						"other": datatypes.Float32(32),
					},
				},
				Properties: &v1.OverrideProperties{
					Limit: helpers.PointerOf[int64](18),
				},
			},
		}
		_, err = limiterClient.CreateOverride(context.Background(), ruleResp.State.ID, override3)
		g.Expect(err).ToNot(HaveOccurred())

		// query
		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"key1": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
					"key2": {
						Value:            datatypes.Int(2),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
					},
					"other": {
						Value:            datatypes.Float32(32),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}

		overrides, err := limiterClient.QueryOverrides(context.Background(), ruleResp.State.ID, query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(2))
	})
}
