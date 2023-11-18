package limter_integration_tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/pkg/clients"
	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

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
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can create a rule", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func Test_Limiter_Rules_Get(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can retrieve a rule that exists with no overrides", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(uint64(5)))
		g.Expect(ruleResp.Overrides).To(BeNil())
	})

	t.Run("It can retrieve a rule that exists with all overrides", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create a number of overrides
		for i := 0; i < 100; i++ {
			override := v1.Override{
				Name:  fmt.Sprintf("override%d", i),
				Limit: 32,
				KeyValues: datatypes.KeyValues{
					fmt.Sprintf("other%d", i): datatypes.Float32(32),
				},
			}
			g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())
		}

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(uint64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(100))
	})

	t.Run("It can retrieve a rule's specific overrides that match the query", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create a number of overrides
		for i := 0; i < 100; i++ {
			override := v1.Override{
				Name:  fmt.Sprintf("override%d", i),
				Limit: 32,
				KeyValues: datatypes.KeyValues{
					fmt.Sprintf("other%d", i): datatypes.Float32(32),
				},
			}
			g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())
		}

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{
			KeyValues: &datatypes.KeyValues{
				"other1":  datatypes.Float32(32),
				"other32": datatypes.Float32(32),
			},
			OverrideQuery: v1.Match,
		})

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(uint64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(2))
	})
}

func Test_Limiter_Rules_List(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can retrieve a single rule", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		ruleResp, err := limiterClient.ListRules(v1.RuleQuery{OverrideQuery: v1.None})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(1))
		g.Expect(ruleResp[0].Name).To(Equal("rule1"))
		g.Expect(ruleResp[0].GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp[0].Limit).To(Equal(uint64(5)))
		g.Expect(ruleResp[0].Overrides).To(BeNil())
	})

	t.Run("It can retrieve multiple rules", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleRequest{}
		respRules := v1.Rules{}
		for i := 0; i < 5; i++ {
			rules = append(rules, v1.RuleRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
				Limit:   5,
			})

			err := limiterClient.CreateRule(rules[i])
			g.Expect(err).ToNot(HaveOccurred())

			respRules = append(respRules, &v1.RuleResponse{
				Name:      fmt.Sprintf("%d", i),
				GroupBy:   []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
				Limit:     5,
				Overrides: nil,
			})
		}

		// get the rules
		ruleResp, err := limiterClient.ListRules(v1.RuleQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(5))
		g.Expect(ruleResp).To(ContainElements(respRules))
	})

	t.Run("It can retrieve any rules that match the key values", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleRequest{}
		respRules := v1.Rules{}
		for i := 0; i < 5; i++ {
			rules = append(rules, v1.RuleRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
				Limit:   5,
			})

			err := limiterClient.CreateRule(rules[i])
			g.Expect(err).ToNot(HaveOccurred())

			if i == 1 || i == 2 {
				respRules = append(respRules, &v1.RuleResponse{
					Name:      fmt.Sprintf("%d", i),
					GroupBy:   []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
					Limit:     5,
					Overrides: nil,
				})
			}
		}

		// get the rules
		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
			"key3": datatypes.Int(3),
		}
		ruleResp, err := limiterClient.ListRules(v1.RuleQuery{KeyValues: &keyValues})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(2))
		g.Expect(ruleResp).To(ContainElements(respRules))
	})

	t.Run("It can retrieve any rules that match the key values and the overrides that match the key values", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleRequest{}
		respRules := v1.Rules{}
		for i := 0; i < 5; i++ {
			keyValues := datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}
			overrideKeyValues := datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1), fmt.Sprintf("key%d", i+2): datatypes.Int(i + 2)}

			rules = append(rules, v1.RuleRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: keyValues.Keys(),
				Limit:   5,
			})
			err := limiterClient.CreateRule(rules[i])
			g.Expect(err).ToNot(HaveOccurred())

			for index, overrideKV := range overrideKeyValues.GenerateTagPairs() {
				overrideReq := v1.Override{
					Name:      fmt.Sprintf("override%d", index),
					KeyValues: overrideKV,
					Limit:     32,
				}
				err := limiterClient.CreateOverride(fmt.Sprintf("%d", i), overrideReq)
				g.Expect(err).ToNot(HaveOccurred())
			}

			if i == 1 {
				respRules = append(respRules, &v1.RuleResponse{
					Name:      fmt.Sprintf("%d", i),
					GroupBy:   keyValues.Keys(),
					Limit:     5,
					Overrides: []v1.Override{},
				})

				// NOTE: this is only the keyValues not the overrideKeyValues. This is because overrideKeyValues contain
				// an extra key that should not be contained
				for index, overrideKV := range keyValues.GenerateTagPairs() {
					respRules[0].Overrides = append(respRules[0].Overrides, v1.Override{
						Name:      fmt.Sprintf("override%d", index),
						KeyValues: overrideKV,
						Limit:     32,
					})
				}
			}
		}

		// get the rules
		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		ruleResp, err := limiterClient.ListRules(v1.RuleQuery{KeyValues: &keyValues, OverrideQuery: v1.Match})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(1))
		g.Expect(ruleResp[0].Name).To(Equal("1"))
		g.Expect(ruleResp[0].GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp[0].Limit).To(Equal(uint64(5)))
		g.Expect(len(respRules[0].Overrides)).To(Equal(3))
		g.Expect(ruleResp[0].Overrides[0]).To(Equal(respRules[0].Overrides[0]))
		g.Expect(ruleResp[0].Overrides[1]).To(Equal(respRules[0].Overrides[1]))
		g.Expect(ruleResp[0].Overrides[2].KeyValues).To(Equal(respRules[0].Overrides[2].KeyValues))

	})
}

func Test_Limiter_Rules_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can update a rule that already exists", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the basic defaults
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(uint64(5)))
		g.Expect(ruleResp.Overrides).To(BeNil())

		// update the rule
		updateRule := v1.RuleUpdate{
			Limit: 231,
		}
		err = limiterClient.UpdateRule("rule1", updateRule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the update took
		ruleResp, err = limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(uint64(231)))
		g.Expect(ruleResp.Overrides).To(BeNil())
	})
}

func Test_Limiter_Rules_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can delete a rule that already exists", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the rule exists
		ruleResp, err := limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))

		// delete the rule
		err = limiterClient.DeleteRule("rule1")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule no longer exists
		deleteRule, err := limiterClient.GetRule("rule1", v1.RuleQuery{OverrideQuery: v1.All})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("rule with name 'rule1' could not be found"))
		g.Expect(deleteRule).To(BeNil())
	})
}
