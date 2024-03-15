package limter_integration_tests

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Limiter_Counters_Update(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	createRule := func(g *GomegaWithT, limiterClient limiterclient.LimiterClient) {
		// create rule
		rule := &v1.Rule{
			Name: "rule1",
			GroupByKeyValues: datatypes.KeyValues{
				"key0": datatypes.Any(),
			},
			Limit: 5,
		}

		err := limiterClient.CreateRule(rule, nil)
		g.Expect(err).ToNot(HaveOccurred())
	}

	t.Run("Incrementing counters", func(t *testing.T) {
		t.Run("It can increment a counter untill a rule limit is reached", func(t *testing.T) {
			t.Parallel()

			lockerTestConstruct := StartLocker(g)
			defer lockerTestConstruct.Shutdown(g)

			limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
			defer limiterTestConstruct.Shutdown(g)

			// setup client
			limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

			// create a default rule
			createRule(g, limiterClient)

			// increment the tags untill the rule limit is reached
			//// the first 5 rules should go fine
			for i := 0; i < 5; i++ {
				counter := &v1.Counter{
					KeyValues: datatypes.KeyValues{
						"key0":                    datatypes.Int(0),
						fmt.Sprintf("key%d", i+1): datatypes.Int(i),
					},
					Counters: 1,
				}

				g.Expect(limiterClient.UpdateCounter(counter, nil)).ToNot(HaveOccurred(), fmt.Sprintf("failed on counter %d", i))
			}

			// the 6th value should be an error
			counter := &v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key6": datatypes.Int(6),
				},
				Counters: 1,
			}
			err := limiterClient.UpdateCounter(counter, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule 'rule1'"))
		})
	})

	t.Run("Context decrementing counters", func(t *testing.T) {
		t.Run("It can decrement a counter even if one doesn't exist", func(t *testing.T) {
			t.Parallel()

			lockerTestConstruct := StartLocker(g)
			defer lockerTestConstruct.Shutdown(g)

			limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
			defer limiterTestConstruct.Shutdown(g)

			// setup client
			limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

			// decrement
			counter := &v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key6": datatypes.Int(6),
				},
				Counters: -1,
			}
			err := limiterClient.UpdateCounter(counter, nil)
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It can decrement a counter and allow a rule to start processing again", func(t *testing.T) {
			t.Parallel()

			lockerTestConstruct := StartLocker(g)
			defer lockerTestConstruct.Shutdown(g)

			limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
			defer limiterTestConstruct.Shutdown(g)

			// setup client
			limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

			// create a default rule
			createRule(g, limiterClient)

			// increment the tags untill the rule limit is reached
			//// the first 5 rules should go fine
			for i := 0; i < 5; i++ {
				counter := &v1.Counter{
					KeyValues: datatypes.KeyValues{
						"key0":                    datatypes.Int(0),
						fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1),
					},
					Counters: 1,
				}

				g.Expect(limiterClient.UpdateCounter(counter, nil)).ToNot(HaveOccurred(), fmt.Sprintf("failed on counter %d", i))
			}

			// the try incrementing a vlue that already exists
			incrementCounter := &v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key5": datatypes.Int(5),
				},
				Counters: 1,
			}
			err := limiterClient.UpdateCounter(incrementCounter, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule 'rule1'"))

			// perform a decrement
			decrementCounter := &v1.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.Int(0),
					"key5": datatypes.Int(5),
				},
				Counters: -1,
			}
			err = limiterClient.UpdateCounter(decrementCounter, nil)
			g.Expect(err).ToNot(HaveOccurred())

			// increment should now pass again
			err = limiterClient.UpdateCounter(incrementCounter, nil)
			g.Expect(err).ToNot(HaveOccurred())
		})
	})
}

func Test_Limiter_Counters_Query(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can list a number of counters that match the query", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a number of counters
		kv1 := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter1 := &v1.Counter{
			KeyValues: kv1,
			Counters:  1,
		}
		g.Expect(limiterClient.UpdateCounter(counter1, nil)).ToNot(HaveOccurred())

		kv2 := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
			"key2": datatypes.String("2"),
		}
		counter2 := &v1.Counter{
			KeyValues: kv2,
			Counters:  1,
		}
		g.Expect(limiterClient.UpdateCounter(counter2, nil)).ToNot(HaveOccurred())
		g.Expect(limiterClient.UpdateCounter(counter2, nil)).ToNot(HaveOccurred())

		counter3 := &v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
			},
			Counters: 1,
		}
		g.Expect(limiterClient.UpdateCounter(counter3, nil)).ToNot(HaveOccurred())

		counter4 := &v1.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int8(0),
			},
			Counters: 1,
		}
		g.Expect(limiterClient.UpdateCounter(counter4, nil)).ToNot(HaveOccurred())

		// query the counters
		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"key1": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}

		counterResp1 := &v1.Counter{
			KeyValues: kv1,
			Counters:  1,
		}
		countersResp2 := &v1.Counter{
			KeyValues: kv2,
			Counters:  2,
		}

		counters, err := limiterClient.QueryCounters(query, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(2))
		g.Expect(counters).To(ContainElements(counterResp1, countersResp2))
	})
}

func Test_Limiter_Counters_Set(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can set a number of counters regardless of the rules", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a restrictive rule
		rule := &v1.Rule{
			Name: "rule1",
			GroupByKeyValues: datatypes.KeyValues{
				"key1": datatypes.Any(),
				"key2": datatypes.Any(),
			},
			Limit: 5,
		}
		g.Expect(limiterClient.CreateRule(rule, nil)).ToNot(HaveOccurred())

		// set a counter for the rule thats above the count
		kv1 := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter1 := &v1.Counter{
			KeyValues: kv1,
			Counters:  32,
		}
		g.Expect(limiterClient.SetCounters(counter1, nil)).ToNot(HaveOccurred())

		// query the counters
		countersResp1 := &v1.Counter{
			KeyValues: kv1,
			Counters:  32,
		}

		counters, err := limiterClient.QueryCounters(&queryassociatedaction.AssociatedActionQuery{}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(1))
		g.Expect(counters).To(ContainElements(countersResp1))
	})

	t.Run("It removes the counter when set to <= 0", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, limiterTestConstruct.ServerURL)

		// create a restrictive rule
		rule := &v1.Rule{
			Name: "rule1",
			GroupByKeyValues: datatypes.KeyValues{
				"key1": datatypes.Any(),
				"key2": datatypes.Any(),
			},
			Limit: 5,
		}
		g.Expect(limiterClient.CreateRule(rule, nil)).ToNot(HaveOccurred())

		// set a counter for the rule thats above the count
		kv1 := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter1 := &v1.Counter{
			KeyValues: kv1,
			Counters:  32,
		}
		g.Expect(limiterClient.SetCounters(counter1, nil)).ToNot(HaveOccurred())

		// reset the counters to 0 to remove the item
		counter3 := &v1.Counter{
			KeyValues: kv1,
			Counters:  0,
		}
		g.Expect(limiterClient.SetCounters(counter3, nil)).ToNot(HaveOccurred())

		// query the counters to ensure it is removed
		query := &queryassociatedaction.AssociatedActionQuery{}
		counters, err := limiterClient.QueryCounters(query, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(0))
	})
}
