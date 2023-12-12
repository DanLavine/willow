package limter_integration_tests

import (
	"fmt"
	"testing"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Limiter_Increment(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	createRule := func(g *GomegaWithT, limiterClient limiterclient.LimiterClient) {
		// create rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key0"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())
	}

	t.Run("It can increment a counter untill a rule limit is reached", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a default rule
		createRule(g, limiterClient)

		// increment the tags untill the rule limit is reached
		//// the first 5 rules should go fine
		for i := 0; i < 5; i++ {
			counter := v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0":                    datatypes.Int(0),
					fmt.Sprintf("key%d", i+1): datatypes.Int(i),
				},
			}

			g.Expect(limiterClient.IncrementCounter(counter)).ToNot(HaveOccurred(), fmt.Sprintf("failed on counter %d", i))
		}

		// the 6th value should be an error
		counter := v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
				"key6": datatypes.Int(6),
			},
		}
		err := limiterClient.IncrementCounter(counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule 'rule1'"))
	})
}

func Test_Limiter_Decrement(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	createRule := func(g *GomegaWithT, limiterClient limiterclient.LimiterClient) {
		// create rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key0"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())
	}

	t.Run("It can decrement a counter even if one doesn't exist", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// decrement
		counter := v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
				"key6": datatypes.Int(6),
			},
		}
		err := limiterClient.DecrementCounter(counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It can decrement a counter and allow a rule to start processing again", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a default rule
		createRule(g, limiterClient)

		// increment the tags untill the rule limit is reached
		//// the first 5 rules should go fine
		for i := 0; i < 5; i++ {
			counter := v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0":                    datatypes.Int(0),
					fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1),
				},
			}

			g.Expect(limiterClient.IncrementCounter(counter)).ToNot(HaveOccurred(), fmt.Sprintf("failed on counter %d", i))
		}

		// the try incrementing a vlue that already exists
		counter := v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
				"key5": datatypes.Int(5),
			},
		}
		err := limiterClient.IncrementCounter(counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule 'rule1'"))

		// perform a decrement
		err = limiterClient.DecrementCounter(counter)
		g.Expect(err).ToNot(HaveOccurred())

		// increment should now pass again
		err = limiterClient.IncrementCounter(counter)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func Test_Limiter_CountersList(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can list a number of counters that match the query", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of counters
		kv1 := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter1 := v1.Counter{
			KeyValues: kv1,
		}
		g.Expect(limiterClient.IncrementCounter(counter1)).ToNot(HaveOccurred())

		kv2 := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
			"key2": datatypes.String("2"),
		}
		counter2 := v1.Counter{
			KeyValues: kv2,
		}
		g.Expect(limiterClient.IncrementCounter(counter2)).ToNot(HaveOccurred())
		g.Expect(limiterClient.IncrementCounter(counter2)).ToNot(HaveOccurred())

		counter3 := v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
			},
		}
		g.Expect(limiterClient.IncrementCounter(counter3)).ToNot(HaveOccurred())

		counter4 := v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int8(0),
			},
		}
		g.Expect(limiterClient.IncrementCounter(counter4)).ToNot(HaveOccurred())

		// query the counters
		trueCheck := true
		query := v1.Query{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key1": datatypes.Value{Exists: &trueCheck},
					},
				},
			},
		}

		counterResp1 := v1.CounterResponse{
			KeyValues: kv1,
			Counters:  1,
		}
		countersResp2 := v1.CounterResponse{
			KeyValues: kv2,
			Counters:  2,
		}

		counters, err := limiterClient.ListCounters(query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(2))
		g.Expect(counters).To(ContainElements(counterResp1, countersResp2))
	})
}

func Test_Limiter_SetCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	lockerTestConstruct := NewIntrgrationLockerTestConstruct(g)
	defer lockerTestConstruct.Cleanup(g)

	limiterTestConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer limiterTestConstruct.Cleanup(g)

	t.Run("It can set a number of counters regardless of the rules", func(t *testing.T) {
		lockerTestConstruct.StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct.StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a restrictive rule
		rule := v1.RuleRequest{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}
		g.Expect(limiterClient.CreateRule(rule)).ToNot(HaveOccurred())

		// set a counter for the rule thats above the count
		kv1 := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter1 := v1.CounterSet{
			KeyValues: kv1,
			Count:     32,
		}
		g.Expect(limiterClient.SetCounters(counter1)).ToNot(HaveOccurred())

		// query the counters
		trueCheck := true
		query := v1.Query{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key1": datatypes.Value{Exists: &trueCheck},
					},
				},
			},
		}

		countersResp1 := v1.CounterResponse{
			KeyValues: kv1,
			Counters:  32,
		}

		counters, err := limiterClient.ListCounters(query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(1))
		g.Expect(counters).To(ContainElements(countersResp1))
	})
}
