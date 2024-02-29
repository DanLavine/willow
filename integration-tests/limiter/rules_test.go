package limter_integration_tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/pkg/clients"
	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func setupClient(g *GomegaWithT, url string) limiterclient.LimiterClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:             url,
		ContentEncoding: api.ContentTypeJSON,
		CAFile:          filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
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

		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
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
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(0))
	})

	t.Run("It can retrieve a rule that exists with all overrides", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create a number of overrides
		for i := 0; i < 100; i++ {
			override := &v1.Override{
				Name:  fmt.Sprintf("override%d", i),
				Limit: 32,
				KeyValues: datatypes.KeyValues{
					"key1":                    datatypes.Int(1),
					"key2":                    datatypes.Int(2),
					fmt.Sprintf("other%d", i): datatypes.Float32(32),
				},
			}
			g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())
		}

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(100))
	})

	t.Run("It can retrieve a rule's specific overrides that match the request", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create a number of overrides
		for i := 0; i < 100; i++ {
			override := &v1.Override{
				Name:  fmt.Sprintf("override%d", i),
				Limit: 32,
				KeyValues: datatypes.KeyValues{
					"key1":                    datatypes.Int(1),
					"key2":                    datatypes.Int(2),
					fmt.Sprintf("other%d", i): datatypes.Float32(32),
				},
			}
			g.Expect(limiterClient.CreateOverride("rule1", override)).ToNot(HaveOccurred())
		}

		// get the rule
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleGet{
			OverridesToMatch: &v1common.MatchQuery{
				KeyValues: &datatypes.KeyValues{
					"key1":    datatypes.Int(1),
					"key2":    datatypes.Int(2),
					"other1":  datatypes.Float32(32),
					"other32": datatypes.Float32(32),
				},
			},
		})

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(2))
	})
}

func Test_Limiter_Rules_Match(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can retrieve a single rule", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create the rule
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		ruleResp, err := limiterClient.MatchRules(&v1.RuleMatch{RulesToMatch: &v1common.MatchQuery{}, OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(1))
		g.Expect(ruleResp[0].Name).To(Equal("rule1"))
		g.Expect(ruleResp[0].GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp[0].Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp[0].Overrides)).To(Equal(0))
	})

	t.Run("It can retrieve multiple rules", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleCreateRequest{}
		for i := 0; i < 5; i++ {
			rules = append(rules, v1.RuleCreateRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
				Limit:   5,
			})

			err := limiterClient.CreateRule(&rules[i])
			g.Expect(err).ToNot(HaveOccurred())
		}

		// get the rules
		ruleResp, err := limiterClient.MatchRules(&v1.RuleMatch{RulesToMatch: &v1common.MatchQuery{}, OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(5))

		for i := 0; i < 5; i++ {
			checkRule := ruleResp[i]
			switch ruleResp[i].Name {
			case "0":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key0", "key1"}))
			case "1":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key1", "key2"}))
			case "2":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key2", "key3"}))
			case "3":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key3", "key4"}))
			case "4":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key4", "key5"}))
			default:
				g.Fail("unkown rule resp")
			}

			g.Expect(checkRule.Limit).To(Equal(int64(5)))
			g.Expect(len(checkRule.Overrides)).To(Equal(0))
		}
	})

	t.Run("It can retrieve any rules that match the key values", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleCreateRequest{}
		for i := 0; i < 5; i++ {
			rules = append(rules, v1.RuleCreateRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
				Limit:   5,
			})

			err := limiterClient.CreateRule(&rules[i])
			g.Expect(err).ToNot(HaveOccurred())
		}

		// get the rules
		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
			"key3": datatypes.Int(3),
		}
		ruleResp, err := limiterClient.MatchRules(&v1.RuleMatch{RulesToMatch: &v1common.MatchQuery{KeyValues: &keyValues}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(2))

		for i := 0; i < 2; i++ {
			checkRule := ruleResp[i]
			switch ruleResp[i].Name {
			case "1":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key1", "key2"}))
			case "2":
				g.Expect(checkRule.GroupBy).To(ContainElements([]string{"key2", "key3"}))
			default:
				g.Fail("unkown rule resp")
			}

			g.Expect(checkRule.Limit).To(Equal(int64(5)))
			g.Expect(len(checkRule.Overrides)).To(Equal(0))
		}
	})

	t.Run("It can retrieve any rules that match the key values and the overrides that match the key values", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of rules
		rules := []v1.RuleCreateRequest{}
		for i := 0; i < 5; i++ {
			keyValues := datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}
			overrideKeyValues := datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}

			// create the rule
			rules = append(rules, v1.RuleCreateRequest{
				Name:    fmt.Sprintf("%d", i),
				GroupBy: keyValues.Keys(),
				Limit:   5,
			})
			err := limiterClient.CreateRule(&rules[i])
			g.Expect(err).ToNot(HaveOccurred())

			// create 5 overrides for each rule
			for k := 5; k < 10; k++ {
				overrideKeyValues[fmt.Sprintf("key%d", k)] = datatypes.Int(k)

				overrideReq := &v1.Override{
					Name:      fmt.Sprintf("override%d", k),
					KeyValues: overrideKeyValues,
					Limit:     32,
				}
				err := limiterClient.CreateOverride(fmt.Sprintf("%d", i), overrideReq)
				g.Expect(err).ToNot(HaveOccurred())
			}
		}

		// setup the response expectations
		respRules := v1.Rules{
			&v1.Rule{
				Name:    "1",
				GroupBy: []string{"key1", "key2"},
				Limit:   5,
				Overrides: v1.Overrides{
					{
						Name: "override5",
						KeyValues: datatypes.KeyValues{
							"key1": datatypes.Int(1),
							"key2": datatypes.Int(2),
							"key5": datatypes.Int(5),
						},
						Limit: 32,
					},
					{
						Name: "override6",
						KeyValues: datatypes.KeyValues{
							"key1": datatypes.Int(1),
							"key2": datatypes.Int(2),
							"key5": datatypes.Int(5),
							"key6": datatypes.Int(6),
						},
						Limit: 32,
					},
				},
			},
		}

		// get the 1 rule + 2 overrides
		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
			"key5": datatypes.Int(5), // override 1
			"key6": datatypes.Int(6), // this + 'key6' are override 2
		}
		ruleResp, err := limiterClient.MatchRules(&v1.RuleMatch{RulesToMatch: &v1common.MatchQuery{KeyValues: &keyValues}, OverridesToMatch: &v1common.MatchQuery{KeyValues: &keyValues}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(ruleResp)).To(Equal(1))
		g.Expect(ruleResp[0].Name).To(Equal("1"))
		g.Expect(ruleResp[0].GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp[0].Limit).To(Equal(int64(5)))
		g.Expect(len(respRules[0].Overrides)).To(Equal(2))
		g.Expect(ruleResp[0].Overrides[0]).To(Equal(respRules[0].Overrides[0]))
		g.Expect(ruleResp[0].Overrides[1]).To(Equal(respRules[0].Overrides[1]))
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
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the basic defaults
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(int64(5)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(0))

		// update the rule
		updateRule := &v1.RuleUpdateRquest{
			Limit: 231,
		}
		err = limiterClient.UpdateRule("rule1", updateRule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the update took
		ruleResp, err = limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))
		g.Expect(ruleResp.GroupBy).To(ContainElements([]string{"key1", "key2"}))
		g.Expect(ruleResp.Limit).To(Equal(int64(231)))
		g.Expect(len(ruleResp.Overrides)).To(Equal(0))
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
		rule := &v1.RuleCreateRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the rule exists
		ruleResp, err := limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ruleResp.Name).To(Equal("rule1"))

		// delete the rule
		err = limiterClient.DeleteRule("rule1")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule no longer exists
		deleteRule, err := limiterClient.GetRule("rule1", &v1.RuleGet{OverridesToMatch: &v1common.MatchQuery{}})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to find rule 'rule1' by name"))
		g.Expect(deleteRule).To(BeNil())
	})
}
