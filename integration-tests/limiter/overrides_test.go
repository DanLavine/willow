package limter_integration_tests

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Limiter_Overrides_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLimiterTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can create an override for a rule", func(t *testing.T) {
		testConstruct.StartLimiter(g)
		defer testConstruct.Shutdown(g)

		// setup client
		limiterClient := setupClient(g, testConstruct.ServerURL)

		// create rule
		rule := &v1.Rule{
			Name:    "rule1",
			GroupBy: []string{"key1", "key2"},
			Limit:   5,
		}

		err := limiterClient.CreateRule(rule)
		g.Expect(err).ToNot(HaveOccurred())

		// create override
		override := &v1.Override{
			Name:  "override1",
			Limit: 32,
			KeyValues: datatypes.KeyValues{
				"other": datatypes.Float64(32),
			},
		}

		err = limiterClient.CreateOverride("rule1", override)
		g.Expect(err).ToNot(HaveOccurred())
	})
}
