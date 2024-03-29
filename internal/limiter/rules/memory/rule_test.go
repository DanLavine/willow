package memory

import (
	"testing"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"go.uber.org/zap"

	. "github.com/onsi/gomega"
)

func Test_New(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It sets the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.RuleCreateRequest{
			Name:    "test",
			GroupBy: []string{"key1"},
			Limit:   56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())

		rule := New(ruleCreateRequest)
		g.Expect(rule.limit.Load()).To(Equal(int64(56)))
	})
}

func Test_Limit(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.RuleCreateRequest{
			Name:    "test",
			GroupBy: []string{"key1"},
			Limit:   56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())

		rule := New(ruleCreateRequest)
		g.Expect(rule.Limit()).To(Equal(int64(56)))
	})
}

func Test_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It updates the limit properly", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.RuleCreateRequest{
			Name:    "test",
			GroupBy: []string{"key1"},
			Limit:   56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())
		rule := New(ruleCreateRequest)

		ruleUpdateRequest := &v1limiter.RuleUpdateRquest{
			Limit: 12,
		}
		g.Expect(ruleUpdateRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rule.Update(zap.NewNop(), ruleUpdateRequest)).ToNot(HaveOccurred())
		g.Expect(rule.Limit()).To(Equal(int64(12)))
	})
}

func Test_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op", func(t *testing.T) {
		ruleCreateRequest := &v1limiter.RuleCreateRequest{
			Name:    "test",
			GroupBy: []string{"key1"},
			Limit:   56,
		}
		g.Expect(ruleCreateRequest.Validate()).ToNot(HaveOccurred())
		rule := New(ruleCreateRequest)

		g.Expect(rule.Delete()).ToNot(HaveOccurred())
	})
}

// func defaultLimiterTestRule(g *GomegaWithT) *v1limiter.RuleCreateRequest {
// 	rule := &v1limiter.RuleCreateRequest{
// 		Name:    "test",
// 		GroupBy: []string{"key1", "key2"},
// 		Limit:   5,
// 	}
// 	g.Expect(rule.Validate()).ToNot(HaveOccurred())

// 	return rule
// }

// func defaultValidKeyValues(g *GomegaWithT) datatypes.KeyValues {
// 	keyValues := datatypes.KeyValues{
// 		"key1": datatypes.String("key1 value"),
// 		"key2": datatypes.String("key2 value"),
// 	}

// 	g.Expect(keyValues.Validate()).ToNot(HaveOccurred())

// 	return keyValues
// }

// func TestRule_Get(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	// create the initial rule
// 	rule := NewRule(defaultLimiterTestRule(g))

// 	// set an override for the rule
// 	overrideRequest := v1limiter.Override{
// 		Name: "override1",
// 		KeyValues: datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 		},
// 		Limit: 87,
// 	}
// 	g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
// 	g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest)).ToNot(HaveOccurred())

// 	t.Run("Context includeOverrides == ''", func(t *testing.T) {
// 		t.Run("It only returns the Rule", func(t *testing.T) {
// 			query := &v1limiter.RuleQuery{}
// 			g.Expect(query.Validate()).ToNot(HaveOccurred())

// 			foundRule := rule.Get(query)
// 			g.Expect(foundRule.Name).To(Equal("test"))
// 			g.Expect(len(foundRule.Overrides)).To(Equal(0))

// 		})
// 	})

// 	t.Run("Context includeOverrides == all", func(t *testing.T) {
// 		t.Run("It includes all rule overrides ", func(t *testing.T) {
// 			query := &v1limiter.RuleQuery{
// 				OverridesToInclude: v1limiter.All,
// 			}
// 			g.Expect(query.Validate()).ToNot(HaveOccurred())

// 			foundRule := rule.Get(query)
// 			g.Expect(foundRule.Name).To(Equal("test"))
// 			g.Expect(len(foundRule.Overrides)).To(Equal(1))
// 			g.Expect(foundRule.Overrides[0].Name).To(Equal("override1"))
// 			g.Expect(foundRule.Overrides[0].KeyValues).To(Equal(
// 				datatypes.KeyValues{
// 					"key1":  datatypes.String("1"),
// 					"key2":  datatypes.String("2"),
// 					"three": datatypes.Float64(52.123),
// 				},
// 			))
// 			g.Expect(foundRule.Overrides[0].Limit).To(Equal(int64(87)))
// 		})
// 	})

// 	t.Run("Context includeOverrides == match", func(t *testing.T) {
// 		// set an additional override for the rule
// 		overrideRequest := v1limiter.Override{
// 			Name: "override2",
// 			KeyValues: datatypes.KeyValues{
// 				"key1": datatypes.String("1"),
// 				"key2": datatypes.String("2"),
// 				"one":  datatypes.String("1"),
// 			},
// 			Limit: 3,
// 		}
// 		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
// 		g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest)).ToNot(HaveOccurred())

// 		t.Run("It includes only overrides that match the key value permutations", func(t *testing.T) {
// 			query := &v1limiter.RuleQuery{
// 				KeyValues: &datatypes.KeyValues{
// 					"key1": datatypes.String("1"),
// 					"key2": datatypes.String("2"),
// 					"one":  datatypes.String("1"),
// 				},
// 				OverridesToInclude: v1limiter.Match,
// 			}
// 			g.Expect(query.Validate()).ToNot(HaveOccurred())

// 			foundRule := rule.Get(query)
// 			g.Expect(foundRule.Name).To(Equal("test"))
// 			g.Expect(len(foundRule.Overrides)).To(Equal(1))
// 			g.Expect(foundRule.Overrides[0].Name).To(Equal("override2"))
// 			g.Expect(foundRule.Overrides[0].KeyValues).To(Equal(
// 				datatypes.KeyValues{
// 					"key1": datatypes.String("1"),
// 					"key2": datatypes.String("2"),
// 					"one":  datatypes.String("1"),
// 				},
// 			))
// 			g.Expect(foundRule.Overrides[0].Limit).To(Equal(int64(3)))
// 		})

// 		t.Run("It does not include the override if the limits are reched on the override", func(t *testing.T) {
// 			query := &v1limiter.RuleQuery{
// 				KeyValues: &datatypes.KeyValues{
// 					"one":   datatypes.String("1"),
// 					"two":   datatypes.Int(2),
// 					"three": datatypes.Float64(37.89),
// 				},
// 				OverridesToInclude: v1limiter.Match,
// 			}
// 			g.Expect(query.Validate()).ToNot(HaveOccurred())

// 			foundRule := rule.Get(query)
// 			g.Expect(foundRule.Name).To(Equal("test"))
// 			g.Expect(len(foundRule.Overrides)).To(Equal(0))
// 		})
// 	})
// }

// func TestRule_FindLimits(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	// create the initial rule
// 	rule := NewRule(defaultLimiterTestRule(g))

// 	// set a number of override for the rule
// 	overrideRequest1 := v1limiter.Override{
// 		Name: "override1",
// 		KeyValues: datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 		},
// 		Limit: 87,
// 	}
// 	g.Expect(overrideRequest1.Validate()).ToNot(HaveOccurred())
// 	g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest1)).ToNot(HaveOccurred())

// 	overrideRequest2 := v1limiter.Override{
// 		Name: "override2",
// 		KeyValues: datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 			"four":  datatypes.String("4"),
// 		},
// 		Limit: 1,
// 	}
// 	g.Expect(overrideRequest2.Validate()).ToNot(HaveOccurred())
// 	g.Expect(rule.SetOverride(zap.NewNop(), &overrideRequest2)).ToNot(HaveOccurred())

// 	t.Run("It reurns the rules limit if no overrides were found", func(t *testing.T) {
// 		keyValues := datatypes.KeyValues{
// 			"key1": datatypes.Float64(52.123),
// 			"key2": datatypes.String("2"),
// 		}

// 		limits, err := rule.FindLimits(zap.NewNop(), keyValues)
// 		g.Expect(err).ToNot(HaveOccurred())
// 		g.Expect(len(limits)).To(Equal(1))
// 		g.Expect(limits[0].KeyValues).To(Equal(datatypes.KeyValues{
// 			"key1": datatypes.Float64(52.123),
// 			"key2": datatypes.String("2"),
// 		}))
// 		g.Expect(limits[0].Limit).To(Equal(int64(5)))
// 	})

// 	t.Run("It reurns any override limit if just 1 was found", func(t *testing.T) {
// 		keyValues := datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 		}

// 		limits, err := rule.FindLimits(zap.NewNop(), keyValues)
// 		g.Expect(err).ToNot(HaveOccurred())
// 		g.Expect(len(limits)).To(Equal(1))
// 		g.Expect(limits[0].KeyValues).To(Equal(datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 		}))
// 		g.Expect(limits[0].Limit).To(Equal(int64(87)))
// 	})

// 	t.Run("It returns all overrides that match the key values", func(t *testing.T) {
// 		keyValues := datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 			"four":  datatypes.String("4"),
// 		}

// 		limits, err := rule.FindLimits(zap.NewNop(), keyValues)
// 		g.Expect(err).ToNot(HaveOccurred())
// 		g.Expect(len(limits)).To(Equal(2))
// 		g.Expect(limits[0].KeyValues).To(Equal(datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 		}))
// 		g.Expect(limits[0].Limit).To(Equal(int64(87)))
// 		g.Expect(limits[1].KeyValues).To(Equal(datatypes.KeyValues{
// 			"key1":  datatypes.String("1"),
// 			"key2":  datatypes.String("2"),
// 			"three": datatypes.Float64(52.123),
// 			"four":  datatypes.String("4"),
// 		}))
// 		g.Expect(limits[1].Limit).To(Equal(int64(1)))
// 	})
// }

// func TestRule_Update(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	t.Run("It can update a rule's limits", func(t *testing.T) {
// 		ruleRequest := defaultLimiterTestRule(g)
// 		rule := NewRule(ruleRequest)

// 		// ensure defaults
// 		g.Expect(rule.limit).To(Equal(int64(5)))

// 		// update the limit to 1
// 		rule.Update(zap.NewNop(), &v1limiter.RuleUpdateRquest{Limit: 1})

// 		// check new limits
// 		g.Expect(rule.limit).To(Equal(int64(1)))
// 	})
// }

// func TestRule_QueryOverrides(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	t.Run("It returns empty if there are no overrides", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		query := &v1common.AssociatedQuery{AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}}
// 		g.Expect(query.Validate()).ToNot(HaveOccurred())

// 		overrides, err := rule.QueryOverrides(zap.NewNop(), query)
// 		g.Expect(err).ToNot(HaveOccurred())
// 		g.Expect(len(overrides)).To(Equal(0))
// 	})

// 	t.Run("It returns only overrides sthat match the query", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		// create 2 overrides
// 		override1 := &v1limiter.Override{
// 			Name:      "override name 1",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}
// 		g.Expect(rule.SetOverride(zap.NewNop(), override1)).ToNot(HaveOccurred())

// 		override1.KeyValues["key4"] = datatypes.Float32(3.2)
// 		override2 := &v1limiter.Override{
// 			Name:      "override name 2",
// 			KeyValues: override1.KeyValues,
// 			Limit:     12,
// 		}
// 		g.Expect(rule.SetOverride(zap.NewNop(), override2)).ToNot(HaveOccurred())

// 		// run the query
// 		notExists := false
// 		query := &v1common.AssociatedQuery{
// 			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
// 				KeyValueSelection: &datatypes.KeyValueSelection{
// 					KeyValues: map[string]datatypes.Value{
// 						"key4": datatypes.Value{Exists: &notExists},
// 					},
// 				},
// 			},
// 		}
// 		g.Expect(query.Validate()).ToNot(HaveOccurred())

// 		overrides, err := rule.QueryOverrides(zap.NewNop(), query)
// 		g.Expect(err).ToNot(HaveOccurred())
// 		g.Expect(len(overrides)).To(Equal(1))
// 		g.Expect(overrides[0]).To(Equal(&v1limiter.Override{
// 			Name:      "override name 1",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}))
// 	})
// }

// func TestRule_SetOverride(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	t.Run("It can create a rule override", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		override := &v1limiter.Override{
// 			Name:      "override name",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}

// 		err := rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).ToNot(HaveOccurred())
// 	})

// 	t.Run("It returns an error if a rule override does not have the GroupBy keys", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		badKeyValues := datatypes.KeyValues{
// 			"nope": datatypes.Int(1),
// 		}
// 		g.Expect(badKeyValues.Validate()).ToNot(HaveOccurred())

// 		override := &v1limiter.Override{
// 			Name:      "bad key values",
// 			KeyValues: badKeyValues,
// 			Limit:     72,
// 		}

// 		err := rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).To(HaveOccurred())
// 		g.Expect(err.Error()).To(ContainSubstring("Missing Rule's GroubBy keys in the override"))
// 	})

// 	t.Run("It returns an error if a rule override already exists by a certain name", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		override := &v1limiter.Override{
// 			Name:      "override name",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}

// 		err := rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).ToNot(HaveOccurred())

// 		// add a new key
// 		override.KeyValues["key4"] = datatypes.Float32(3.2)

// 		// 2nd create with the same name should be a problem
// 		err = rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).To(HaveOccurred())
// 		g.Expect(err.Error()).To(ContainSubstring("failed to create rule override. Name already in use for another override"))
// 	})

// 	t.Run("It returns an error if a rule override already exists with the same key values", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		override := &v1limiter.Override{
// 			Name:      "different name",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}

// 		err := rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).ToNot(HaveOccurred())

// 		// 2nd create with the same name should be a problem
// 		err = rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).To(HaveOccurred())
// 		g.Expect(err.Error()).To(ContainSubstring("failed to create rule override. KeyValues already in use for another override"))
// 	})
// }

// func TestRule_DeleteOverride(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	t.Run("It returns an error if there override name was not found", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		err := rule.DeleteOverride(zap.NewNop(), "doesn't exist")
// 		g.Expect(err).To(HaveOccurred())
// 		g.Expect(err.Error()).To(ContainSubstring("Override doesn't exist not found"))
// 	})

// 	t.Run("It can delte an override by name", func(t *testing.T) {
// 		rule := NewRule(defaultLimiterTestRule(g))

// 		// create the override
// 		override := &v1limiter.Override{
// 			Name:      "override name",
// 			KeyValues: defaultValidKeyValues(g),
// 			Limit:     72,
// 		}

// 		err := rule.SetOverride(zap.NewNop(), override)
// 		g.Expect(err).ToNot(HaveOccurred())

// 		// delete the override
// 		err = rule.DeleteOverride(zap.NewNop(), "override name")
// 		g.Expect(err).ToNot(HaveOccurred())

// 		// ensure the override was deleted
// 		foundRule := rule.Get(&v1limiter.RuleQuery{OverridesToInclude: v1limiter.All})
// 		g.Expect(len(foundRule.Overrides)).To(Equal(0))
// 	})
// }
