package limiter

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/limiter/rules/rulefakes"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// mock constructor and rule
func setupMocks(t *testing.T) (*gomock.Controller, *rulefakes.MockRuleConstructor, *rulefakes.MockRule) {
	mockController := gomock.NewController(t)

	fakeRule := rulefakes.NewMockRule(mockController)
	fakeRuleConstructor := rulefakes.NewMockRuleConstructor(mockController)

	fakeRuleConstructor.EXPECT().New(gomock.Any()).DoAndReturn(func(createParams *v1limiter.RuleRequest) rules.Rule {
		return fakeRule
	}).AnyTimes()

	return mockController, fakeRuleConstructor, fakeRule
}

func TestRulesManager_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when successfully creating a new rule", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())

		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error when trying to create rule with the same name", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())

		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())

		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to create rule"))
	})

	t.Run("It returns an error when trying to create rule with the same group by keys", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())

		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())

		createRequest.Name = "test2"
		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to create rule"))
	})
}

func TestRulesManager_Get(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when a rule doesn't exist", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		query := &v1limiter.RuleQuery{
			OverrideQuery: v1limiter.All,
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		rule := rulesManager.Get(zap.NewNop(), "doesn't exist", query)
		g.Expect(rule).To(BeNil())
	})

	t.Run("Context when the rules exists", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		// create the rule
		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// create an override
		overrideRequest := v1limiter.Override{
			Name: "override1",
			KeyValues: datatypes.KeyValues{
				"three": datatypes.Float64(52.123),
			},
			Limit: 87,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test", &overrideRequest)).ToNot(HaveOccurred())

		t.Run("It respects the key values query", func(t *testing.T) {
			query := &v1limiter.RuleQuery{OverrideQuery: v1limiter.All}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			rule := rulesManager.Get(zap.NewNop(), "test", query)
			g.Expect(rule).ToNot(BeNil())
			g.Expect(len(rule.Overrides)).To(Equal(1))
			g.Expect(rule.Overrides[0].Name).To(Equal("override1"))
		})
	})
}

func TestRulesManager_List(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an empty list when no rules are found", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		ruleQuery := &v1limiter.RuleQuery{
			KeyValues: &datatypes.KeyValues{
				"key1": datatypes.Int(1),
			},
			OverrideQuery: v1limiter.All,
		}
		g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

		rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(rules).To(BeEmpty())
	})

	t.Run("Context when there are a number of rules", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		// create 5 rules with single group by keys
		for i := 0; i < 5; i++ {
			// single instance rule group by
			// group by: {[key0], [key1], [key2], [key3], [key4]}
			createRequest := &v1limiter.RuleRequest{
				Name:    fmt.Sprintf("test%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i)},
				Limit:   5, // todo. make this optinal...
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// up to KeyN -> Where N is the number of overrides
			// overrides: { {"key0":0}, {"key1":1}, ...}
			for k := 0; k <= i; k++ {
				// create number of overrides
				overrideRequest := v1limiter.Override{
					Name: fmt.Sprintf("override%d", k),
					KeyValues: datatypes.KeyValues{
						fmt.Sprintf("key%d", k): datatypes.Int(k),
					},
					Limit: 10,
				}
				g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.CreateOverride(zap.NewNop(), fmt.Sprintf("test%d", i), &overrideRequest)).ToNot(HaveOccurred())
			}
		}

		keyValues := datatypes.KeyValues{"key0": datatypes.Int(0)}
		for i := 1; i < 5; i++ {
			// multi instance group by
			// group by: {[key0, key1], [key0, key1, key2], ...}
			keyValues[fmt.Sprintf("key%d", i)] = datatypes.Int(i)
			createRequest := &v1limiter.RuleRequest{
				Name:    fmt.Sprintf("multi_test%d", i),
				GroupBy: keyValues.Keys(),
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// permutation for all key values as overrides
			for index, keyValue := range keyValues.GenerateTagPairs() {
				// create number of overrides
				overrideRequest := v1limiter.Override{
					Name:      fmt.Sprintf("override%d", index),
					KeyValues: keyValue,
					Limit:     10,
				}
				g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.CreateOverride(zap.NewNop(), fmt.Sprintf("multi_test%d", i), &overrideRequest)).ToNot(HaveOccurred())
			}
		}

		t.Run("It can list all rules", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				OverrideQuery: v1limiter.None,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(rules)).To(Equal(9))
		})

		t.Run("It can list all rules and their overrides", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				OverrideQuery: v1limiter.All,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(rules)).To(Equal(9))
			for i := 0; i < 9; i++ {
				if reflect.DeepEqual(rules[i].GroupBy, []string{"key0"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(1))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key1"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(2))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key2"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(3))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key3"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(4))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key4"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(5))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key0", "key1"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(3))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key0", "key1", "key2"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(7))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key0", "key1", "key2", "key3"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(15))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key0", "key1", "key2", "key3", "key4"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(31))
				}
			}
		})

		t.Run("It can match a nummber of key values", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				KeyValues: &datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key1": datatypes.Int(1),
				},
				OverrideQuery: v1limiter.None,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())

			// 1 for: [key0]
			// 1 for: [key1]
			// 1 for: [key0, key1]
			g.Expect(len(rules)).To(Equal(3))
			g.Expect(len(rules[0].Overrides)).To(Equal(0))
			g.Expect(len(rules[1].Overrides)).To(Equal(0))
			g.Expect(len(rules[2].Overrides)).To(Equal(0))
		})

		t.Run("It can match a nummber of key values and includes any overrides that match the key values", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				KeyValues: &datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key1": datatypes.Int(1),
				},
				OverrideQuery: v1limiter.All,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())

			// 1 for: [key0]
			// 1 for: [key1]
			// 1 for: [key0, key1]
			g.Expect(len(rules)).To(Equal(3))

			for i := 0; i < 3; i++ {
				if reflect.DeepEqual(rules[i].GroupBy, []string{"key0"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(1))
				} else if reflect.DeepEqual(rules[i].GroupBy, []string{"key1"}) {
					g.Expect(len(rules[i].Overrides)).To(Equal(2))
				} else {
					g.Expect(len(rules[i].Overrides)).To(Equal(3))
				}
			}
		})
	})
}

func TestRulesManager_Update(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error when failing to find the rule by name", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		ruleUpdate := &v1limiter.RuleUpdate{
			Limit: 12,
		}

		err = rulesManager.Update(zap.NewNop(), "doesn't exist", ruleUpdate)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("failed to find rule by name"))
	})

	t.Run("It can update a rule by name", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		// create the rule
		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// update the rule
		ruleUpdate := &v1limiter.RuleUpdate{
			Limit: 12,
		}
		err = rulesManager.Update(zap.NewNop(), "test", ruleUpdate)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule was updated
		rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverrideQuery: v1limiter.All})
		g.Expect(rule.Limit).To(Equal(uint64(12)))
	})
}

func TestRulesManager_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil if the rule does not exist", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		err = rulesManager.Delete(zap.NewNop(), "not found")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It deletes the rule if it exists", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor, nil)

		// create the rule
		createRequest := &v1limiter.RuleRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// delete the rule
		err = rulesManager.Delete(zap.NewNop(), "test")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule was deleted
		rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverrideQuery: v1limiter.All})
		g.Expect(rule).To(BeNil())
	})

	t.Run("Context when the cascade delete operation passes", func(t *testing.T) {
		t.Run("It also deletes all the overrides for the rule", func(t *testing.T) {
			mockController, mockConstructor, mockRule := setupMocks(t)
			defer mockController.Finish()

			// ensure cascade delete is called
			mockRule.EXPECT().CascadeDeletion(gomock.Any()).DoAndReturn(func(logger *zap.Logger) *api.Error {
				return nil
			}).Times(1)

			rulesManager := NewRulesManger(mockConstructor, nil)

			// create the rule
			createRequest := &v1limiter.RuleRequest{
				Name:    "test",
				GroupBy: []string{"key1", "key2"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// delete the rule
			err := rulesManager.Delete(zap.NewNop(), "test")
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the rule was deleted
			rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverrideQuery: v1limiter.All})
			g.Expect(rule).To(BeNil())
		})
	})

	t.Run("Context when the cascade delete operation fails", func(t *testing.T) {
		t.Run("It does not deletes the rule and reports the error", func(t *testing.T) {
			mockController, mockConstructor, mockRule := setupMocks(t)
			defer mockController.Finish()

			// ensure cascade delete and Get are called
			mockRule.EXPECT().CascadeDeletion(gomock.Any()).DoAndReturn(func(logger *zap.Logger) *api.Error {
				return &api.Error{Message: "failed to cascade delete", StatusCode: http.StatusInternalServerError}
			}).Times(1)
			mockRule.EXPECT().Get(gomock.Any()).DoAndReturn(func(includeOverrides *v1limiter.RuleQuery) *v1limiter.RuleResponse {
				return &v1limiter.RuleResponse{}
			}).Times(1)

			rulesManager := NewRulesManger(mockConstructor, nil)

			// create the rule
			createRequest := &v1limiter.RuleRequest{
				Name:    "test",
				GroupBy: []string{"key1", "key2"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// delete the rule
			err := rulesManager.Delete(zap.NewNop(), "test")
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to cascade delete"))

			// ensure the rule was not deleted
			rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverrideQuery: v1limiter.All})
			g.Expect(rule).ToNot(BeNil())
		})
	})
}

/*
func TestRulesManager_FindRule(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when a rule doesn't exist", func(t *testing.T) {
		rulesManager := NewRulesManger()

		rule := rulesManager.FindRule(zap.NewNop(), "doesn't exist")
		g.Expect(rule).To(BeNil())
	})

	t.Run("It returns the proper rule if it exists", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.CreateGroupRule(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())

		rule := rulesManager.FindRule(zap.NewNop(), "test")
		g.Expect(rule).ToNot(BeNil())
	})
}

func TestRulesManager_ListRules(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an empty list when no rules exist", func(t *testing.T) {
		rulesManager := NewRulesManger()

		rules := rulesManager.ListRules(zap.NewNop())
		g.Expect(len(rules)).To(Equal(0))
	})

	t.Run("It returns the proper rule if it exists", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.CreateGroupRule(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())

		rules := rulesManager.ListRules(zap.NewNop())
		g.Expect(len(rules)).To(Equal(1))
	})
}

func TestRulesManager_DeleteGroupRule(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It deletes a rule iff it exists by name", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		rules := rulesManager.ListRules(zap.NewNop())
		g.Expect(len(rules)).To(Equal(1))

		rulesManager.DeleteGroupRule(zap.NewNop(), "test")

		rules = rulesManager.ListRules(zap.NewNop())
		g.Expect(len(rules)).To(Equal(0))
	})
}

func TestRulesManager_Increment(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can increment a anything if no rules exist", func(t *testing.T) {
		rulesManager := NewRulesManger()

		increment := &v1limiter.RuleCounterRequest{
			KeyValues: datatypes.KeyValues{"key1": datatypes.String("first")},
		}

		err := rulesManager.Increment(zap.NewNop(), increment)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error if a rule has reached its limit", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		increment := &v1limiter.RuleCounterRequest{
			KeyValues: datatypes.KeyValues{"key1": datatypes.String("first"), "key2": datatypes.Float64(3.4)},
		}

		// setup to reach the limit of 5
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())

		// next call should error since the limit has been reached
		err := rulesManager.Increment(zap.NewNop(), increment)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Unable to process limit request. The limits are already reached"))
	})

	t.Run("It returns an error if any rule has reached its limit", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest1 := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest1.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest1)).ToNot(HaveOccurred())

		createRequest2 := &v1limiter.Rule{
			Name:        "test2",
			GroupBy:     []string{"key1"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       1,
		}
		g.Expect(createRequest2.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest2)).ToNot(HaveOccurred())

		increment := &v1limiter.RuleCounterRequest{
			KeyValues: datatypes.KeyValues{"key1": datatypes.String("first"), "key2": datatypes.Float64(3.4)},
		}

		// setup to reach the limit of 1 from rule 2 with the stricter set of keys
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())

		// next call should error since the limit has been reached
		err := rulesManager.Increment(zap.NewNop(), increment)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Unable to process limit request. The limits are already reached"))
	})

	t.Run("It returns an error if a rule is added after a group of key values are already past its limit", func(t *testing.T) {
		rulesManager := NewRulesManger()

		increment := &v1limiter.RuleCounterRequest{
			KeyValues: datatypes.KeyValues{"key1": datatypes.String("first"), "key2": datatypes.Float64(3.4), "key3": datatypes.Int(2)},
		}

		// setup to reach the limit of 5
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), increment)).ToNot(HaveOccurred())

		// create the rule
		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// next call should error since the limit has been reached
		err := rulesManager.Increment(zap.NewNop(), increment)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Unable to process limit request. The limits are already reached"))
	})
}

func TestRulesManager_Decrement(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It decreases the counter by 1 and removes an 'counters' if their limit hit 0", func(t *testing.T) {
		rulesManager := NewRulesManger()

		counter := &v1limiter.RuleCounterRequest{
			KeyValues: datatypes.KeyValues{"key1": datatypes.String("first"), "key2": datatypes.Float64(3.4)},
		}
		g.Expect(rulesManager.Increment(zap.NewNop(), counter)).ToNot(HaveOccurred())
		g.Expect(rulesManager.Increment(zap.NewNop(), counter)).ToNot(HaveOccurred())

		var counterValue uint64
		onFind := func(item any) bool {
			counterValue = item.(*btreeassociated.AssociatedKeyValues).Value().(*atomic.Uint64).Load()
			return true
		}

		// ensure we have a counter of 2
		rulesManager.counters.Query(datatypes.AssociatedKeyValuesQuery{}, onFind)
		g.Expect(counterValue).To(Equal(uint64(2)))

		rulesManager.Decrement(zap.NewNop(), counter)
		// ensure we have a counter of 1
		rulesManager.counters.Query(datatypes.AssociatedKeyValuesQuery{}, onFind)
		g.Expect(counterValue).To(Equal(uint64(1)))

		rulesManager.Decrement(zap.NewNop(), counter)
		// ensure we have a counter of 0
		counterValue = 0
		rulesManager.counters.Query(datatypes.AssociatedKeyValuesQuery{}, onFind)
		g.Expect(counterValue).To(Equal(uint64(0)))
	})
}
*/
