package limter_integration_tests

import (
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

	testConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can create a rule", func(t *testing.T) {
		testConstruct.StartLimiter(g)
		defer testConstruct.Shutdown(g)

		limiterClient := setupClient(g, testConstruct.ServerURL)

		rule := &v1.Rule{
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

	testConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can retrieve a rule that exists", func(t *testing.T) {
		testConstruct.StartLimiter(g)
		defer testConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, testConstruct.ServerURL)

		// create the rule
		rule := &v1.Rule{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule
		rule, err = limiterClient.GetRule("rule1", false)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rule.Name).To(Equal("rule1"))
		g.Expect(rule.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(rule.Limit).To(Equal(uint64(5)))
		g.Expect(rule.Overrides).To(BeNil())
		g.Expect(rule.QueryFilter).To(Equal(datatypes.AssociatedKeyValuesQuery{}))
	})
}

func Test_Limiter_Rules_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can update a rule that already exists", func(t *testing.T) {
		testConstruct.StartLimiter(g)
		defer testConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, testConstruct.ServerURL)

		// create the rule
		rule := &v1.Rule{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the basic defaults
		rule, err = limiterClient.GetRule("rule1", false)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rule.Name).To(Equal("rule1"))
		g.Expect(rule.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(rule.Limit).To(Equal(uint64(5)))
		g.Expect(rule.Overrides).To(BeNil())
		g.Expect(rule.QueryFilter).To(Equal(datatypes.AssociatedKeyValuesQuery{}))

		// update the rule
		updateRule := &v1.RuleUpdate{
			Limit: 231,
		}
		err = limiterClient.UpdateRule("rule1", updateRule)
		g.Expect(err).ToNot(HaveOccurred())

		// get the rule and ensure the update took
		rule, err = limiterClient.GetRule("rule1", false)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rule.Name).To(Equal("rule1"))
		g.Expect(rule.GroupBy).To(Equal([]string{"key1", "key2"}))
		g.Expect(rule.Limit).To(Equal(uint64(231)))
		g.Expect(rule.Overrides).To(BeNil())
		g.Expect(rule.QueryFilter).To(Equal(datatypes.AssociatedKeyValuesQuery{}))
	})
}

func Test_Limiter_Rules_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can delete a rule that already exists", func(t *testing.T) {
		testConstruct.StartLimiter(g)
		defer testConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, testConstruct.ServerURL)

		// create the rule
		rule := &v1.Rule{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the rule exists
		rule, err = limiterClient.GetRule("rule1", false)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rule.Name).To(Equal("rule1"))

		// delete the rule
		err = limiterClient.DeleteRule("rule1")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule no longer exists
		deleteRule, err := limiterClient.GetRule("rule1", false)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("rule with name 'rule1' could not be found"))
		g.Expect(deleteRule).To(BeNil())
	})
}
