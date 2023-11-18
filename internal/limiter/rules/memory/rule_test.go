package memory

import (
	"testing"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func defaultLimiterTestRule(g *GomegaWithT) *v1limiter.RuleRequest {
	rule := &v1limiter.RuleRequest{
		Name:    "test",
		GroupBy: []string{"key1", "key2"},
		Limit:   5,
	}
	g.Expect(rule.Validate()).ToNot(HaveOccurred())

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

func TestRule_Get(t *testing.T) {
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
	g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
	g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest)).ToNot(HaveOccurred())

	t.Run("Context includeOverrides == ''", func(t *testing.T) {
		t.Run("It only returns the Rule", func(t *testing.T) {
			query := &v1limiter.RuleQuery{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			foundRule := rule.Get(query)
			g.Expect(foundRule.Name).To(Equal("test"))
			g.Expect(len(foundRule.Overrides)).To(Equal(0))

		})
	})

	t.Run("Context includeOverrides == all", func(t *testing.T) {
		t.Run("It includes all rule overrides ", func(t *testing.T) {
			query := &v1limiter.RuleQuery{
				OverrideQuery: v1limiter.All,
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			foundRule := rule.Get(query)
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

	t.Run("Context includeOverrides == match", func(t *testing.T) {
		// set an additional override for the rule
		overrideRequest := v1limiter.Override{
			Name: "override2",
			KeyValues: datatypes.KeyValues{
				"one": datatypes.String("1"),
			},
			Limit: 2,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest)).ToNot(HaveOccurred())

		t.Run("It includes only overrides that match the key value permutations", func(t *testing.T) {
			query := &v1limiter.RuleQuery{
				KeyValues: &datatypes.KeyValues{
					"one": datatypes.String("1"),
				},
				OverrideQuery: v1limiter.Match,
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			foundRule := rule.Get(query)
			g.Expect(foundRule.Name).To(Equal("test"))
			g.Expect(len(foundRule.Overrides)).To(Equal(1))
			g.Expect(foundRule.Overrides[0].Name).To(Equal("override2"))
			g.Expect(foundRule.Overrides[0].KeyValues).To(Equal(
				datatypes.KeyValues{
					"one": datatypes.String("1"),
				},
			))
			g.Expect(foundRule.Overrides[0].Limit).To(Equal(uint64(2)))
		})

		t.Run("It does not include the override if the limits are reched on the override", func(t *testing.T) {
			query := &v1limiter.RuleQuery{
				KeyValues: &datatypes.KeyValues{
					"one":   datatypes.String("1"),
					"two":   datatypes.Int(2),
					"three": datatypes.Float64(37.89),
				},
				OverrideQuery: v1limiter.Match,
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			foundRule := rule.Get(query)
			g.Expect(foundRule.Name).To(Equal("test"))
			g.Expect(len(foundRule.Overrides)).To(Equal(0))
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

func TestRule_DeleteOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil if there was no item to delete", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		err := rule.DeleteOverride(zap.NewNop(), "doesn't exist")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It can delte an override by name", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		// create the override
		override := &v1limiter.Override{
			Name:      "override name",
			KeyValues: defaultValidKeyValues(g),
			Limit:     72,
		}

		err := rule.SetOverride(zap.NewNop(), override)
		g.Expect(err).ToNot(HaveOccurred())

		// delete the override
		err = rule.DeleteOverride(zap.NewNop(), "override name")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the override was deleted
		foundRule := rule.Get(&v1limiter.RuleQuery{OverrideQuery: v1limiter.All})
		g.Expect(len(foundRule.Overrides)).To(Equal(0))
	})
}

func TestRule_FindLimit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns the Rule limit if an override cannot be found", func(t *testing.T) {
		rule := NewRule(defaultLimiterTestRule(g))

		keyValues := datatypes.KeyValues{
			"one": datatypes.String("1"),
		}

		limit, err := rule.FindLimit(zap.NewNop(), keyValues)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(limit).To(Equal(rule.limit))
	})

	//t.Run("It returns the Override limit if an override can be found", func(t *testing.T) {
	//	rule := NewRule(defaultLimiterTestRule(g))
	//
	//	// set one overide that matches
	//	override := &v1limiter.Override{
	//		Name: "name",
	//		KeyValues: datatypes.KeyValues{
	//			"key4": datatypes.String("key4 value"),
	//		},
	//		Limit: 72,
	//	}
	//	g.Expect(override.Validate()).ToNot(HaveOccurred())
	//	g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())
	//
	//	// set one overide that does not match
	//	override = &v1limiter.Override{
	//		Name: "different name",
	//		KeyValues: datatypes.KeyValues{
	//			"key4": datatypes.String("key4 value"),
	//			"key5": datatypes.String("key5 value"),
	//		},
	//		Limit: 13,
	//	}
	//	g.Expect(override.Validate()).ToNot(HaveOccurred())
	//	g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())
	//
	//	// check limit
	//	keyValues := datatypes.KeyValues{
	//		"key4": datatypes.String("key4 value"),
	//	}
	//
	//	limit, err := rule.FindLimit(zap.NewNop(), keyValues)
	//	g.Expect(err).ToNot(HaveOccurred())
	//	g.Expect(limit).To(Equal(uint64(72)))
	//})

	//t.Run("Context when multiple overrides match", func(t *testing.T) {
	//	t.Run("It selects the lowest override limit", func(t *testing.T) {
	//		rule := NewRule(defaultLimiterTestRule(g))
	//
	//		// set one overide that matches
	//		override := &v1limiter.Override{
	//			Name: "name",
	//			KeyValues: datatypes.KeyValues{
	//				"key4": datatypes.String("key4 value"),
	//			},
	//			Limit: 72,
	//		}
	//		g.Expect(override.Validate()).ToNot(HaveOccurred())
	//		g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())
	//
	//		// set one overide that does not match
	//		override = &v1limiter.Override{
	//			Name:      "different name",
	//			KeyValues: datatypes.KeyValues{},
	//			Limit:     1,
	//		}
	//		g.Expect(override.Validate()).ToNot(HaveOccurred())
	//		g.Expect(rule.SetOverride(zap.NewNop(), override)).ToNot(HaveOccurred())
	//
	//		// check limit
	//		keyValues := datatypes.KeyValues{
	//			"key4": datatypes.String("key4 value"),
	//		}
	//
	//		limit, err := rule.FindLimit(zap.NewNop(), keyValues)
	//		g.Expect(err).ToNot(HaveOccurred())
	//		g.Expect(limit).To(Equal(uint64(1)))
	//	})
	//})
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
