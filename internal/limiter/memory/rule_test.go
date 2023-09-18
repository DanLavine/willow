package memory

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/api/v1limiter"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func defaultLimiterTestRule(g *GomegaWithT) *v1limiter.RuleRequest {
	rule := &v1limiter.RuleRequest{
		Name:     "test",
		GroupBy:  []string{"key1", "key2"},
		Seletion: query.Select{},
		Limit:    5,
	}

	g.Expect(rule.Validate()).ToNot(HaveOccurred())

	return rule
}

func defaultValidKeyValues(g *GomegaWithT) datatypes.StringMap {
	return datatypes.StringMap{
		"key1": datatypes.String("key1 value"),
		"key2": datatypes.String("key2 value"),
	}
}

// Test that the defaults are valid
func Test_Defaults(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("The defaultValidKeyValues are valid", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))
		defaultValidKeyValues := defaultValidKeyValues(g)

		g.Expect(rule.TagsMatch(zap.NewNop(), defaultValidKeyValues)).To(BeTrue())
	})
}

func TestRule_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It updates to use the new default limit", func(t *testing.T) {
		// setup a rule where that is designed to block 100% of requests
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		// update the limit to 1
		rule.Update(zap.NewNop(), 1)

		g.Expect(rule.ruleModel.Limit).To(Equal(uint64(1)))
	})
}

func TestRule_SetOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an errpd of the tags don't match", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		stringKey := datatypes.String("2")
		defaultLimiterRule.Seletion = query.Select{
			Where: &query.Query{
				KeyValues: map[string]query.Value{
					"key1": query.Value{Value: &stringKey, ValueComparison: query.EqualsPtr()},
				},
			},
		}
		rule := NewRule(defaultLimiterRule)

		override := &v1limiter.RuleOverride{
			RuleName: "test",
			KeyValues: datatypes.StringMap{
				"key1": datatypes.Int(3),
			},
			Limit: 32,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("the provided keys values to match the rule query."))
	})

	t.Run("Context when using find", func(t *testing.T) {
		t.Run("It the override value for the limit", func(t *testing.T) {
			defaultLimiterRule := defaultLimiterTestRule(g)
			rule := NewRule(defaultLimiterRule)

			// check the initial limit
			limit := rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(5)))

			override := &v1limiter.RuleOverride{
				RuleName:  "test",
				KeyValues: defaultValidKeyValues(g),
				Limit:     32,
			}
			g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())

			// check the limit again
			limit = rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(32)))
		})

		t.Run("It allows for updating an override that already exists", func(t *testing.T) {
			defaultLimiterRule := defaultLimiterTestRule(g)
			rule := NewRule(defaultLimiterRule)

			// check the initial limit
			limit := rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(5)))

			override := &v1limiter.RuleOverride{
				RuleName:  "test",
				KeyValues: defaultValidKeyValues(g),
				Limit:     32,
			}
			g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())

			override.Limit = 22341
			g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())

			// check the limit again
			limit = rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(22341)))
		})
	})
}

func TestRule_DeleteOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op if the override does not exist", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		override := &v1limiter.RuleOverride{
			RuleName: "test",
			KeyValues: datatypes.StringMap{
				"key1": datatypes.Int(3),
			},
			Limit: 32,
		}

		err := rule.DeleteOverride(zap.NewNop(), override)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Context when an override does exist", func(t *testing.T) {
		t.Run("It deletes the override", func(t *testing.T) {
			defaultLimiterRule := defaultLimiterTestRule(g)
			rule := NewRule(defaultLimiterRule)

			override := &v1limiter.RuleOverride{
				RuleName:  "test",
				KeyValues: defaultValidKeyValues(g),
				Limit:     32,
			}
			g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())

			// check the override limit
			limit := rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(32)))

			rule.DeleteOverride(zap.NewNop(), override)

			// check the limit again
			limit = rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
			g.Expect(limit).To(Equal(uint64(5)))
		})
	})
}

func TestRule_TagsMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if any 'group_by' tags are missing", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		matched := rule.TagsMatch(zap.NewNop(), datatypes.StringMap{"key1": datatypes.Float64(3.2)})
		g.Expect(matched).To(BeFalse())
	})

	t.Run("It returns true if all 'group_by' tags are included", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		matched := rule.TagsMatch(zap.NewNop(), defaultValidKeyValues(g))
		g.Expect(matched).To(BeTrue())
	})

	t.Run("It returns true if all 'group_by' tags are included and there are extra key values", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		validKeys := defaultValidKeyValues(g)
		validKeys["key3"] = datatypes.String("other")

		matched := rule.TagsMatch(zap.NewNop(), validKeys)
		g.Expect(matched).To(BeTrue())
	})

	t.Run("It returns false the rule's selection filters out a group of tags", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		falsePtr := false
		defaultLimiterRule.Seletion = query.Select{
			And: []query.Select{
				{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"key3": query.Value{Exists: &falsePtr},
						},
					},
				},
			},
		}
		rule := NewRule(defaultLimiterRule)

		invalidKeys := defaultValidKeyValues(g)
		invalidKeys["key3"] = datatypes.String("other")

		matched := rule.TagsMatch(zap.NewNop(), invalidKeys)
		g.Expect(matched).To(BeFalse())
	})
}

func TestRule_Generate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns a query using the rule's group_by key values", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		testKeyValues := defaultValidKeyValues(g)
		testKeyValues["key3"] = datatypes.String("other") // this should not be in the final query

		query := rule.GenerateQuery(testKeyValues)
		g.Expect(query.Where).ToNot(BeNil())
		g.Expect(len(query.Where.KeyValues)).To(Equal(2))
		g.Expect(*(query.Where.KeyValues["key1"].Value)).To(Equal(testKeyValues["key1"]))
		g.Expect(*(query.Where.KeyValues["key2"].Value)).To(Equal(testKeyValues["key2"]))
	})
}
