package limiter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/limiter/rules/rulefakes"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/clients/locker_client/lockerclientfakes"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

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

func TestRulesManager_IncrementCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns nil if there are no rules to match against", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("1"),
			},
			Counters: 1,
		}

		err := rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a limit reached error if any matched rule has a limit of 0", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i := 0; i < 5; i++ {
			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    fmt.Sprintf("test%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i)},
				Limit:   int64(i),
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}

		err := rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(LimitReached))
	})

	t.Run("It returns a limit reached error if any matched overrides have a limit of 0", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// single instance rule group by
		createRequest := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   15,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// create 5 overrides
		for k := 2; k < 7; k++ {
			// create number of overrides
			overrideRequest := v1limiter.Override{
				Name: fmt.Sprintf("override%d", k),
				KeyValues: datatypes.KeyValues{
					"key1":                  datatypes.Int(1),
					fmt.Sprintf("key%d", k): datatypes.Int(k),
				},
				Limit: int64(k - 2),
			}
			g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test1", &overrideRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Int(1),
				"key2": datatypes.Int(2),
				"key3": datatypes.Int(3),
				"key4": datatypes.Int(4),
			},
			Counters: 1,
		}

		err := rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(LimitReached))
	})

	t.Run("Describe obtaining lock failures", func(t *testing.T) {
		t.Run("It locks and releases each key value pair when trying to increment a counter if a rule is found", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			//setup rules
			for i := 0; i < 5; i++ {
				// single instance rule group by
				createRequest := &v1limiter.RuleCreateRequest{
					Name:    fmt.Sprintf("test%d", i),
					GroupBy: []string{fmt.Sprintf("key%d", i)},
					Limit:   int64(i),
				}
				g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}

			mockController := gomock.NewController(t)
			defer mockController.Finish()

			fakeLock := lockerclientfakes.NewMockLock(mockController)
			fakeLock.EXPECT().Done().Times(2)
			fakeLock.EXPECT().Release().Times(2)

			fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
			fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Lock, error) {
				return fakeLock, nil
			}).Times(2)

			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("Context when obtaining the lock fails", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				rulesManager := NewRulesManger(constructor)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.RuleCreateRequest{
						Name:    fmt.Sprintf("test%d", i),
						GroupBy: []string{fmt.Sprintf("key%d", i)},
						Limit:   int64(i),
					}
					g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
					g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					KeyValues: datatypes.KeyValues{
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
					Counters: 1,
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				fakeLock := lockerclientfakes.NewMockLock(mockController)
				fakeLock.EXPECT().Release().Times(1)
				fakeLock.EXPECT().Done().MaxTimes(1)

				obtainCount := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Lock, error) {
					if obtainCount == 0 {
						obtainCount++
						return fakeLock, nil
					} else {
						return nil, fmt.Errorf("failed to obtain 2nd lock")
					}
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testLgger := zap.New(testZapCore)

				err := rulesManager.IncrementCounters(testLgger, ctx, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("failed to obtain a lock from the locker service"))
			})
		})

		t.Run("Context when a lock is lost that was already obtained", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				rulesManager := NewRulesManger(constructor)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.RuleCreateRequest{
						Name:    fmt.Sprintf("test%d", i),
						GroupBy: []string{fmt.Sprintf("key%d", i)},
						Limit:   int64(i),
					}
					g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
					g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					KeyValues: datatypes.KeyValues{
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
					Counters: 1,
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				donechan := make(chan struct{})
				close(donechan)

				fakeLock := lockerclientfakes.NewMockLock(mockController)
				fakeLock.EXPECT().Release().Times(2)
				fakeLock.EXPECT().Done().DoAndReturn(func() <-chan struct{} {
					return donechan
				}).MaxTimes(2)

				count := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Lock, error) {
					if count == 0 {
						count++
					} else {
						time.Sleep(100 * time.Millisecond)
					}

					return fakeLock, nil
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testLgger := zap.New(testZapCore)

				err := rulesManager.IncrementCounters(testLgger, ctx, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("a lock was released unexpedily"))
			})
		})
	})

	t.Run("Context rule limits", func(t *testing.T) {
		mockController := gomock.NewController(t)

		fakeLock := lockerclientfakes.NewMockLock(mockController)
		fakeLock.EXPECT().Release().AnyTimes()
		fakeLock.EXPECT().Done().AnyTimes()

		fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
		fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Lock, error) {
			return fakeLock, nil
		}).AnyTimes()

		t.Run("It adds the counter if no rules have reached their limit", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// counter shuold be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			found := false
			onFind := func(item any) {
				found = true
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue())
		})

		t.Run("It respects the unlimited rules", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   -1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 341,
			}

			// counter shuold be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			found := false
			onFind := func(item any) {
				found = true
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue())
		})

		t.Run("It can update the limit for counters that are below all the rules", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// counter shuold be added
			g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())

			// ensure the counter was added
			count := int64(0)
			onFind := func(item any) {
				count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(4)))
		})

		t.Run("It returns an error if the counter >= the limit", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item any) {
				count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if the counter >= the limit with any combination of different counters", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter1 := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}
			counter2 := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter1)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter2)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item any) {
				count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter1.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))

			count = 0 // reset the counter
			id, counterErr = rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter2.KeyValues), onFind)
			g.Expect(id).To(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(0)))
		})

		t.Run("It returns an error if any rule has hit the limit", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// restrictive rule
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// non restrictive rules
			for i := 1; i < 5; i++ {
				createRequest := &v1limiter.RuleCreateRequest{
					Name:    fmt.Sprintf("test%d", i),
					GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
					Limit:   int64(i + 10),
				}
				g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item any) {
				count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if any rule's overrides hit the limit", func(t *testing.T) {
			rulesManager := NewRulesManger(constructor)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// restrictive rule
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// set override to allow for more values
			override := &v1limiter.Override{
				Name: "override1",
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
				},
				Limit: 5,
			}
			g.Expect(override.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test0", override)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should be added as well
			err = rulesManager.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item any) {
				count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
			}

			id, counterErr := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
			g.Expect(id).ToNot(Equal(""))
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(2)))
		})
	})
}

func TestRulesManager_DecrementCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns nil if there are no counters to decrement against", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("1"),
			},
			Counters: 1,
		}

		err := rulesManager.DecrementCounters(zap.NewNop(), counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It decrements a counter when the count is above 1", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item any) {
			count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
		}

		id, err := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
		g.Expect(id).ToNot(Equal(""))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(4)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: -1,
		}
		err = rulesManager.DecrementCounters(zap.NewNop(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		id, err = rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
		g.Expect(id).ToNot(Equal(""))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(3)))
	})

	t.Run("It removes a counter when the count is at 1", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item any) {
			count = item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Load()
		}

		id, err := rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
		g.Expect(id).ToNot(Equal(""))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(1)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: -1,
		}
		err = rulesManager.DecrementCounters(zap.NewNop(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		id, err = rulesManager.counters.Find(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), onFind)
		g.Expect(id).To(Equal(""))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(0)))
	})
}

func TestRulesManager_ListCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	exists := true

	t.Run("It returns empty list if there are no counters that match the query", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"not found": datatypes.Value{
							Exists: &exists,
						},
					},
				},
			},
		}

		countersResponse, err := rulesManager.ListCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a list of counters that match the query", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create a number of various counters
		keyValuesOne := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
			"key2": datatypes.String("2"),
		}
		counter1 := &v1limiter.Counter{
			KeyValues: keyValuesOne,
			Counters:  1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())

		keyValuesTwo := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter2 := &v1limiter.Counter{
			KeyValues: keyValuesTwo,
			Counters:  1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter2)).ToNot(HaveOccurred())

		counter3 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
			},
			Counters: 1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter3)).ToNot(HaveOccurred())

		counter4 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("0"),
			},
			Counters: 1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter4)).ToNot(HaveOccurred())

		counter5 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
				"key1": datatypes.String("0"),
			},
			Counters: 1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter5)).ToNot(HaveOccurred())

		// run the query
		int0 := datatypes.Int(0)
		int1 := datatypes.Int(1)
		int2 := datatypes.Int(2)

		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key0": datatypes.Value{ // values 1
							Exists:     &exists,
							ExistsType: &datatypes.T_string,
						},
					},
				},
				Or: []datatypes.AssociatedKeyValuesQuery{ // values 2
					datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"key0": datatypes.Value{Value: &int0, ValueComparison: datatypes.EqualsPtr()},
								"key1": datatypes.Value{Value: &int1, ValueComparison: datatypes.EqualsPtr()},
								"key2": datatypes.Value{Value: &int2, ValueComparison: datatypes.EqualsPtr()},
							},
						},
					},
				},
			},
		}

		resp1 := &v1limiter.Counter{
			KeyValues: keyValuesOne,
			Counters:  3,
		}
		resp2 := &v1limiter.Counter{
			KeyValues: keyValuesTwo,
			Counters:  1,
		}

		countersResponse, err := rulesManager.ListCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(2))
		g.Expect(countersResponse).To(ConsistOf(resp1, resp2))
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestRulesManager_SetCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It sets the value for the particualr key values regardless of the value already stored", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)

		// set the counters
		kvs := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Float64(3.4),
		}
		countersSet := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  87,
		}

		err := rulesManager.SetCounters(zap.NewNop(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}, // select all
		}

		countersResponse, err := rulesManager.ListCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(1))
		g.Expect(countersResponse[0].KeyValues).To(Equal(kvs))
		g.Expect(countersResponse[0].Counters).To(Equal(int64(87)))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It removes the counters if set to 0", func(t *testing.T) {
		rulesManager := NewRulesManger(constructor)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create initial counters through increment
		kvs := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
		}
		counter := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  1,
		}
		g.Expect(rulesManager.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// set the counters
		countersSet := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  0,
		}

		err := rulesManager.SetCounters(zap.NewNop(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}, // select all
		}

		countersResponse, err := rulesManager.ListCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})
}
