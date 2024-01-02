package rules

/*
import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DanLavine/willow/internal/limiter/rules/rulefakes"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/onsi/gomega"
)

func setupServerHttp(serverMux *http.ServeMux) *httptest.Server {
	return httptest.NewServer(serverMux)
}

// mock constructor and rule
func setupMocks(t *testing.T) (*gomock.Controller, *rulefakes.MockRuleConstructor, *rulefakes.MockRule) {
	mockController := gomock.NewController(t)

	fakeRule := rulefakes.NewMockRule(mockController)
	fakeRuleConstructor := rulefakes.NewMockRuleConstructor(mockController)

	fakeRuleConstructor.EXPECT().New(gomock.Any()).DoAndReturn(func(createParams *v1limiter.RuleCreateRequest) rules.Rule {
		return fakeRule
	}).AnyTimes()

	return mockController, fakeRuleConstructor, fakeRule
}

func TestRulesManager_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when successfully creating a new rule", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		createRequest := &v1limiter.RuleCreateRequest{
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
		rulesManager := NewRulesManger(constructor)

		createRequest := &v1limiter.RuleCreateRequest{
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
		rulesManager := NewRulesManger(constructor)

		createRequest := &v1limiter.RuleCreateRequest{
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
		rulesManager := NewRulesManger(constructor)

		query := &v1limiter.RuleQuery{
			OverridesToInclude: v1limiter.All,
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		rule := rulesManager.Get(zap.NewNop(), "doesn't exist", query)
		g.Expect(rule).To(BeNil())
	})

	t.Run("Context when the rules exists", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		// create the rule
		createRequest := &v1limiter.RuleCreateRequest{
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
				"key1":  datatypes.Int(1),
				"key2":  datatypes.Int(2),
				"three": datatypes.Float64(52.123),
			},
			Limit: 87,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test", &overrideRequest)).ToNot(HaveOccurred())

		t.Run("It respects the key values query", func(t *testing.T) {
			query := &v1limiter.RuleQuery{OverridesToInclude: v1limiter.All}
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
		rulesManager := NewRulesManger(constructor)

		ruleQuery := &v1limiter.RuleQuery{
			KeyValues: &datatypes.KeyValues{
				"key1": datatypes.Int(1),
			},
			OverridesToInclude: v1limiter.All,
		}
		g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

		rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(rules)).To(Equal(0))
	})

	t.Run("Context when there are a number of rules", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		// create 5 rules with single group by keys
		for i := 0; i < 5; i++ {
			// single instance rule group by
			// group by: {[key0], [key1], [key2], [key3], [key4]}
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    fmt.Sprintf("test%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i)},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// up to KeyN -> Where N is the number of overrides
			// overrides: { {"keyi":i, "keyi+1": i+1}, {"keyi+1":i+1, "keyi+2":i+2}, ...}
			for k := i + 1; k <= i+5; k++ {
				// create number of overrides
				overrideRequest := v1limiter.Override{
					Name: fmt.Sprintf("override%d", k),
					KeyValues: datatypes.KeyValues{
						fmt.Sprintf("key%d", i): datatypes.Int(i),
						fmt.Sprintf("key%d", k): datatypes.Int(k),
					},
					Limit: 10,
				}
				g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.CreateOverride(zap.NewNop(), fmt.Sprintf("test%d", i), &overrideRequest)).ToNot(HaveOccurred())
			}
		}

		// create 4 rules with multiple key values
		keyValues := datatypes.KeyValues{"key0": datatypes.Int(0)}
		for i := 1; i < 5; i++ {
			// multi instance group by
			// group by: {[key0, key1], [key0, key1, key2], ...}
			keyValues[fmt.Sprintf("key%d", i)] = datatypes.Int(i)
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    fmt.Sprintf("multi_test%d", i),
				GroupBy: keyValues.Keys(),
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// add a number of overrides
			newKeyValues := datatypes.KeyValues{}
			for key, value := range keyValues {
				newKeyValues[key] = value
			}

			// create override with the same key values
			overrideRequest := v1limiter.Override{
				Name:      "override0",
				KeyValues: newKeyValues,
				Limit:     10,
			}
			g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.CreateOverride(zap.NewNop(), fmt.Sprintf("multi_test%d", i), &overrideRequest)).ToNot(HaveOccurred())

			for k := i + 1; k <= i+4; k++ {
				newKeyValues[fmt.Sprintf("key%d", k)] = datatypes.Int(k)

				// create number of overrides
				overrideRequest := v1limiter.Override{
					Name:      fmt.Sprintf("override%d", k),
					KeyValues: newKeyValues,
					Limit:     10,
				}
				g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.CreateOverride(zap.NewNop(), fmt.Sprintf("multi_test%d", i), &overrideRequest)).ToNot(HaveOccurred())
			}
		}

		t.Run("It can list all rules", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				OverridesToInclude: v1limiter.None,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(rules)).To(Equal(9))
		})

		t.Run("It can list all rules and their overrides", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				OverridesToInclude: v1limiter.All,
			}
			g.Expect(ruleQuery.Validate()).ToNot(HaveOccurred())

			rules, err := rulesManager.List(zap.NewNop(), ruleQuery)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(rules)).To(Equal(9))
			for i := 0; i < 9; i++ {
				g.Expect(len(rules[i].Overrides)).To(Equal(5))
			}
		})

		t.Run("It can match a nummber of key values", func(t *testing.T) {
			ruleQuery := &v1limiter.RuleQuery{
				KeyValues: &datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key1": datatypes.Int(1),
				},
				OverridesToInclude: v1limiter.None,
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
				OverridesToInclude: v1limiter.Match,
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
					g.Expect(len(rules[i].Overrides)).To(Equal(0))
				} else {
					g.Expect(len(rules[i].Overrides)).To(Equal(1))
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
		rulesManager := NewRulesManger(constructor)

		ruleUpdate := &v1limiter.RuleUpdateRquest{
			Limit: 12,
		}

		err = rulesManager.Update(zap.NewNop(), "doesn't exist", ruleUpdate)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("failed to find rule 'doesn't exist' by name"))
	})

	t.Run("It can update a rule by name", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		// create the rule
		createRequest := &v1limiter.RuleCreateRequest{
			Name:    "test",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// update the rule
		ruleUpdate := &v1limiter.RuleUpdateRquest{
			Limit: 12,
		}
		err = rulesManager.Update(zap.NewNop(), "test", ruleUpdate)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the rule was updated
		rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverridesToInclude: v1limiter.All})
		g.Expect(rule.Limit).To(Equal(int64(12)))
	})
}

func TestRulesManager_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil if the rule does not exist", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		err = rulesManager.Delete(zap.NewNop(), "not found")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It deletes the rule if it exists", func(t *testing.T) {
		constructor, err := rules.NewRuleConstructor("memory")
		g.Expect(err).ToNot(HaveOccurred())
		rulesManager := NewRulesManger(constructor)

		// create the rule
		createRequest := &v1limiter.RuleCreateRequest{
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
		rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverridesToInclude: v1limiter.All})
		g.Expect(rule).To(BeNil())
	})

	t.Run("Context when the cascade delete operation passes", func(t *testing.T) {
		t.Run("It also deletes all the overrides for the rule", func(t *testing.T) {
			mockController, mockConstructor, mockRule := setupMocks(t)
			defer mockController.Finish()

			// ensure cascade delete is called
			mockRule.EXPECT().CascadeDeletion(gomock.Any()).DoAndReturn(func(logger *zap.Logger) *errors.ServerError {
				return nil
			}).Times(1)

			rulesManager := NewRulesManger(mockConstructor)

			// create the rule
			createRequest := &v1limiter.RuleCreateRequest{
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
			rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverridesToInclude: v1limiter.All})
			g.Expect(rule).To(BeNil())
		})
	})

	t.Run("Context when the cascade delete operation fails", func(t *testing.T) {
		t.Run("It does not deletes the rule and reports the error", func(t *testing.T) {
			mockController, mockConstructor, mockRule := setupMocks(t)
			defer mockController.Finish()

			// ensure cascade delete and Get are called
			mockRule.EXPECT().CascadeDeletion(gomock.Any()).DoAndReturn(func(logger *zap.Logger) *errors.ServerError {
				return &errors.ServerError{Message: "failed to cascade delete", StatusCode: http.StatusInternalServerError}
			}).Times(1)
			mockRule.EXPECT().Get(gomock.Any()).DoAndReturn(func(includeOverrides *v1limiter.RuleQuery) *v1limiter.Rule {
				return &v1limiter.Rule{}
			}).Times(1)

			rulesManager := NewRulesManger(mockConstructor)

			// create the rule
			createRequest := &v1limiter.RuleCreateRequest{
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
			rule := rulesManager.Get(zap.NewNop(), "test", &v1limiter.RuleQuery{OverridesToInclude: v1limiter.All})
			g.Expect(rule).ToNot(BeNil())
		})
	})
}

func TestRulesManager_ListOverrides(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns an error if the rule name cannot be found", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		notExists := false
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key4": datatypes.Value{Exists: &notExists},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		overrides, err := rulesManager.ListOverrides(zap.NewNop(), "test1", query)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Rule test1 not found"))
		g.Expect(len(overrides)).To(Equal(0))
	})

	t.Run("It returns the overrides for the query", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		// create rule
		createRule := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   int64(12),
		}
		g.Expect(createRule.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRule)).ToNot(HaveOccurred())

		// create override
		overrideRequest := v1limiter.Override{
			Name: "override0",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Int(0),
			},
			Limit: 1,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test1", &overrideRequest)).ToNot(HaveOccurred())

		// query
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}, // select all
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		overrides, err := rulesManager.ListOverrides(zap.NewNop(), "test1", query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(1))
		g.Expect(overrides[0].Name).To(Equal("override0"))
	})
}

func TestRulesManager_CreateOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns an error if the rule name cannot be found", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		overrideRequest := v1limiter.Override{
			Name: "override0",
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
			},
			Limit: 1,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.CreateOverride(zap.NewNop(), "test1", &overrideRequest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Rule test1 not found"))
	})

	t.Run("It returns nil if the rule was successfully created", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		// create rule
		createRule := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   int64(12),
		}
		g.Expect(createRule.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRule)).ToNot(HaveOccurred())

		// create override
		overrideRequest := v1limiter.Override{
			Name: "override0",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Int(0),
			},
			Limit: 1,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.CreateOverride(zap.NewNop(), "test1", &overrideRequest)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestRulesManager_DeleteOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns an error if the rule name cannot be found", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		overrideRequest := v1limiter.Override{
			Name: "override0",
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
			},
			Limit: 1,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.DeleteOverride(zap.NewNop(), "test1", "override0")
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Rule test1 not found"))
	})

	t.Run("It returns an error if the override name cannot be found", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		// create rule
		createRule := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   int64(12),
		}
		g.Expect(createRule.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRule)).ToNot(HaveOccurred())

		// delete override
		err := rulesManager.DeleteOverride(zap.NewNop(), "test1", "override0")
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Override override0 not found"))
	})

	t.Run("It returns nil if the override is deleted successfully", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		// create rule
		createRule := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   int64(12),
		}
		g.Expect(createRule.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRule)).ToNot(HaveOccurred())

		// create override
		overrideRequest := v1limiter.Override{
			Name: "override0",
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Int(0),
			},
			Limit: 1,
		}
		g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())

		err := rulesManager.CreateOverride(zap.NewNop(), "test1", &overrideRequest)
		g.Expect(err).ToNot(HaveOccurred())

		// delete override
		err = rulesManager.DeleteOverride(zap.NewNop(), "test1", "override0")
		g.Expect(err).ToNot(HaveOccurred())
	})
}
*/
