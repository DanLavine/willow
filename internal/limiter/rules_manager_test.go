package limiter

import (
	"testing"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestRulesManager_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when successfully creating a new rule", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())

		err := rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error when trying to create rule with the same name", func(t *testing.T) {
		rulesManager := NewRulesManger()

		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())

		err := rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).ToNot(HaveOccurred())

		err = rulesManager.Create(zap.NewNop(), createRequest)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("rule already exists"))
	})
}

func TestRulesManager_Get(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns nil when a rule doesn't exist", func(t *testing.T) {
		rulesManager := NewRulesManger()

		rule := rulesManager.Get(zap.NewNop(), "doesn't exist", false)
		g.Expect(rule).To(BeNil())
	})

	t.Run("Context override lookups", func(t *testing.T) {
		rulesManager := NewRulesManger()

		// create the rule
		createRequest := &v1limiter.Rule{
			Name:        "test",
			GroupBy:     []string{"key1", "key2"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       5,
		}
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())
		g.Expect(rulesManager.Create(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// create an override
		overrideRequest := v1limiter.Override{
			Name: "override1",
			KeyValues: datatypes.KeyValues{
				"three": datatypes.Float64(52.123),
			},
			Limit: 87,
		}
		g.Expect(overrideRequest.ValidateRequest()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateOverride(zap.NewNop(), "test", &overrideRequest)).ToNot(HaveOccurred())

		t.Run("It does not include any overrides when FALSE", func(t *testing.T) {
			rule := rulesManager.Get(zap.NewNop(), "test", false)
			g.Expect(rule).ToNot(BeNil())
			g.Expect(len(rule.Overrides)).To(Equal(0))
		})

		t.Run("It includes any overrides when TRUE", func(t *testing.T) {
			rule := rulesManager.Get(zap.NewNop(), "test", true)
			g.Expect(rule).ToNot(BeNil())
			g.Expect(len(rule.Overrides)).To(Equal(1))
			g.Expect(rule.Overrides[0].Name).To(Equal("override1"))
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
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())

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
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())

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
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())
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
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())
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
		g.Expect(createRequest1.ValidateRequest()).ToNot(HaveOccurred())
		g.Expect(rulesManager.CreateGroupRule(zap.NewNop(), createRequest1)).ToNot(HaveOccurred())

		createRequest2 := &v1limiter.Rule{
			Name:        "test2",
			GroupBy:     []string{"key1"},
			QueryFilter: datatypes.AssociatedKeyValuesQuery{},
			Limit:       1,
		}
		g.Expect(createRequest2.ValidateRequest()).ToNot(HaveOccurred())
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
		g.Expect(createRequest.ValidateRequest()).ToNot(HaveOccurred())
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
