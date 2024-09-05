package limter_integration_tests

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/clients"
	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func setupClient(g *GomegaWithT, url string) limiterclient.LimiterClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:           url,
		CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	limiterClient, err := limiterclient.NewLimiterClient(cfg)
	g.Expect(err).ToNot(HaveOccurred())

	return limiterClient
}

func Test_Limiter_Rules_Create(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can create a rule", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

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

		_, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func Test_Limiter_Rules_Get(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can retrieve a rule that exists with no overrides", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
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
		g.Expect(ruleResp.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Any(), "key2": datatypes.Any()}))
		g.Expect(*ruleResp.Spec.Properties.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.State.Overrides)).To(Equal(0))
	})
}

func Test_Limiter_Rules_List(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can retrieve all rules", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
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
		_, err := limiterClient.CreateRule(context.Background(), rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		ruleResp, err := limiterClient.QueryRules(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(1))
		g.Expect(ruleResp[0].Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Any(), "key2": datatypes.Any()}))
		g.Expect(*ruleResp[0].Spec.Properties.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp[0].State.Overrides)).To(Equal(0))
	})

	t.Run("It can retrieve rules that match the query", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.Rule{}
		for i := 0; i < 5; i++ {
			rules = append(rules, v1.Rule{
				Spec: &v1.RuleSpec{
					DBDefinition: &v1.RuleDBDefinition{
						GroupByKeyValues: datatypes.KeyValues{
							fmt.Sprintf("key%d", i):   datatypes.Any(),
							fmt.Sprintf("key%d", i+1): datatypes.Any(),
						},
					},
					Properties: &v1.RuleProperties{
						Limit: helpers.PointerOf[int64](int64(i)),
					},
				},
			})

			_, err := limiterClient.CreateRule(context.Background(), &rules[i])
			g.Expect(err).ToNot(HaveOccurred())
		}

		// get the rules
		ruleResp, err := limiterClient.QueryRules(context.Background(), &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"key1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(2))

		for i := 0; i < 2; i++ {
			checkRule := ruleResp[i]
			switch *ruleResp[i].Spec.Properties.Limit {
			case 0:
				g.Expect(checkRule.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key0": datatypes.Any(), "key1": datatypes.Any()}))
			case 1:
				g.Expect(checkRule.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Any(), "key2": datatypes.Any()}))
			default:
				g.Fail("unkown rule resp")
			}
		}
	})
}

func Test_Limiter_Rules_Update(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can update a rule that already exists", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
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
		g.Expect(ruleResp.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Any(), "key2": datatypes.Any()}))
		g.Expect(*ruleResp.Spec.Properties.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.State.Overrides)).To(Equal(0))

		// update the rule
		updateRule := &v1.RuleProperties{
			Limit: helpers.PointerOf[int64](231),
		}
		err = limiterClient.UpdateRule(context.Background(), ruleResp.State.ID, updateRule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the update took
		ruleResp, err = limiterClient.GetRule(context.Background(), ruleResp.State.ID)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Any(), "key2": datatypes.Any()}))
		g.Expect(*ruleResp.Spec.Properties.Limit).To(Equal(int64(231)))
		g.Expect(len(ruleResp.State.Overrides)).To(Equal(0))
	})
}

func Test_Limiter_Rules_Delete(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can delete a rule that already exists", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
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
		g.Expect(ruleResp.State.ID).ToNot(Equal(""))

		// delete the rule
		err = limiterClient.DeleteRule(context.Background(), ruleResp.State.ID)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule no longer exists
		deleteRule, err := limiterClient.GetRule(context.Background(), ruleResp.State.ID)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to find rule"))
		g.Expect(deleteRule).To(BeNil())
	})
}
