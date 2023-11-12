package memory

import (
	"testing"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func defaultLimiterTestRule(g *GomegaWithT) *v1limiter.Rule {
	existFalse := false
	maxKeys := 5

	rule := &v1limiter.Rule{
		Name:    "test",
		GroupBy: []string{"key1", "key2"},
		QueryFilter: datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"key3": datatypes.Value{Exists: &existFalse},
				},
				Limits: &datatypes.KeyLimits{
					NumberOfKeys: &maxKeys,
				},
			},
		},
		Limit: 5,
	}

	g.Expect(rule.ValidateRequest()).ToNot(HaveOccurred())

	return rule
}

func defaultValidKeyValues(g *GomegaWithT) datatypes.KeyValues {
	keyValues := datatypes.KeyValues{
		"key1": datatypes.String("key1 value"),
		"key2": datatypes.String("key2 value"),
	}

	g.Expect(keyValues.Validate()).ToNot(HaveOccurred())

	return keyValues
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

func TestRule_GetRuleResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	// create the initial rule
	rule := NewRule(defaultLimiterTestRule(g))

	// set an override for the rule
	overrideRequest := v1limiter.Override{
		Name: "override1",
		KeyValues: datatypes.KeyValues{
			"three": datatypes.Float64(52.123),
		},
		Limit: 87,
	}
	g.Expect(overrideRequest.ValidateRequest()).ToNot(HaveOccurred())
	g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest)).ToNot(HaveOccurred())

	t.Run("Context includeOverrides == false", func(t *testing.T) {
		t.Run("It only returns the Rule ", func(t *testing.T) {
			foundRule := rule.Get(false)
			g.Expect(foundRule.Name).To(Equal("test"))
			g.Expect(len(foundRule.Overrides)).To(Equal(0))

		})
	})

	t.Run("Context includeOverrides == true", func(t *testing.T) {
		t.Run("It includes all rule overrides ", func(t *testing.T) {
			foundRule := rule.Get(true)
			g.Expect(foundRule.Name).To(Equal("test"))
			g.Expect(len(foundRule.Overrides)).To(Equal(1))
			g.Expect(foundRule.Overrides[0].Name).To(Equal("override1"))
			g.Expect(foundRule.Overrides[0].KeyValues).To(Equal(
				datatypes.KeyValues{
					"three": datatypes.Float64(52.123),
				},
			))
			g.Expect(foundRule.Overrides[0].Limit).To(Equal(uint64(87)))
		})
	})
}

func TestRule_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can update a rule's limits", func(t *testing.T) {
		ruleRequest := defaultLimiterTestRule(g)
		rule := NewRule(ruleRequest)

		// ensure defaults
		g.Expect(rule.limit).To(Equal(uint64(5)))

		// update the limit to 1
		rule.Update(zap.NewNop(), &v1limiter.RuleUpdate{Limit: 1})

		// check new limits
		g.Expect(rule.limit).To(Equal(uint64(1)))
	})
}

func TestRule_SetOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can create a rule override", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		override := &v1limiter.Override{
			Name:      "override name",
			KeyValues: defaultValidKeyValues(g),
			Limit:     72,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error if the query filter already ignores the override", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		override := &v1limiter.Override{
			Name: "different name",
			KeyValues: datatypes.KeyValues{
				"key3": datatypes.String("key3 value"),
			},
			Limit: 72,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Invalid request. Expected: the provided keys values to match the rule query. Actual: provided will never be found by the rule query."))
	})

	t.Run("It returns an error if a rule override already exists by a certain name", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		override := &v1limiter.Override{
			Name:      "override name",
			KeyValues: defaultValidKeyValues(g),
			Limit:     72,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).ToNot(HaveOccurred())

		// add a new key
		override.KeyValues["key4"] = datatypes.Float32(3.2)

		// 2nd create with the same name should be a problem
		err = rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("name to not be in use"))
	})

	t.Run("It returns an error if a rule override already exists with the same key values", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		override := &v1limiter.Override{
			Name:      "different name",
			KeyValues: defaultValidKeyValues(g),
			Limit:     72,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).ToNot(HaveOccurred())

		// 2nd create with the same name should be a problem
		err = rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key values to not have an override already"))
	})
}

/*
// func TestRule_DeleteOverride(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	t.Run("It performs a no-op if the override does not exist", func(t *testing.T) {
// 		defaultLimiterRule := defaultLimiterTestRule(g)
// 		rule := NewRule(defaultLimiterRule)

// 		override := &v1limiter.RuleOverrideRequest{
// 			RuleName: "test",
// 			KeyValues: datatypes.KeyValues{
// 				"key1": datatypes.Int(3),
// 			},
// 			Limit: 32,
// 		}

// 		err := rule.DeleteOverride(zap.NewNop(), override)
// 		g.Expect(err).ToNot(HaveOccurred())
// 	})

// 	t.Run("Context when an override does exist", func(t *testing.T) {
// 		t.Run("It deletes the override", func(t *testing.T) {
// 			defaultLimiterRule := defaultLimiterTestRule(g)
// 			rule := NewRule(defaultLimiterRule)

// 			override := &v1limiter.RuleOverrideRequest{
// 				KeyValues: defaultValidKeyValues(g),
// 				Limit:     32,
// 			}
// 			g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())

// 			// check the override limit
// 			limit := rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
// 			g.Expect(limit).To(Equal(uint64(32)))

// 			rule.DeleteOverride(zap.NewNop(), override)

// 			// check the limit again
// 			limit = rule.FindLimit(zap.NewNop(), defaultValidKeyValues(g))
// 			g.Expect(limit).To(Equal(uint64(5)))
// 		})
// 	})
// }

func TestRule_TagsMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if any 'group_by' tags are missing", func(t *testing.T) {
		defaultLimiterRule := defaultLimiterTestRule(g)
		rule := NewRule(defaultLimiterRule)

		matched := rule.TagsMatch(zap.NewNop(), datatypes.KeyValues{"key1": datatypes.Float64(3.2)})
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
		defaultLimiterRule.QueryFilter = datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"key3": datatypes.Value{Exists: &falsePtr},
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
		g.Expect(query.KeyValueSelection).ToNot(BeNil())
		g.Expect(len(query.KeyValueSelection.KeyValues)).To(Equal(2))
		g.Expect(*(query.KeyValueSelection.KeyValues["key1"].Value)).To(Equal(testKeyValues["key1"]))
		g.Expect(*(query.KeyValueSelection.KeyValues["key2"].Value)).To(Equal(testKeyValues["key2"]))
	})
}
*/
