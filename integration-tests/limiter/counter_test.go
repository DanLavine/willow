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
