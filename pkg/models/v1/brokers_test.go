package v1

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []datatypes.Strings, val datatypes.Strings) {
	count := 0
	for _, groupVal := range group {
		if reflect.DeepEqual(groupVal, val) {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %s", line, val))
}

func TestBrokerType_GenerateTagPairs(t *testing.T) {
	g := NewGomegaWithT(t)

	setupBrokerInfo := func(g *GomegaWithT) *BrokerInfo {
		broker := &BrokerInfo{
			Name: "test",
			Tags: map[datatypes.String]datatypes.String{
				"a": "1",
				"b": "2",
				"c": "3",
				"d": "4",
				"e": "5",
			},
		}

		g.Expect(broker.validate()).ToNot(HaveOccurred())
		return broker
	}

	t.Run("it uses the default tag when none are provided", func(t *testing.T) {
		brokerInfo := &BrokerInfo{
			Name: "test",
		}
		g.Expect(brokerInfo.validate()).ToNot(HaveOccurred())

		generatedTagGroups := brokerInfo.GenerateTagPairs()
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"default"})
		g.Expect(len(generatedTagGroups)).To(Equal(1))
	})

	t.Run("it returns all individual elements", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"c", "3"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"e", "5"})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "c", "3"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "e", "5"})

		// b group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "c", "3"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "e", "5"})

		// c group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"c", "3", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"c", "3", "e", "5"})

		// d group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"d", "4", "e", "5"})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "c", "3"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "e", "5"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "c", "3", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "c", "3", "e", "5"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "d", "4", "e", "5"})

		// b group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "c", "3", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "c", "3", "e", "5"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "d", "4", "e", "5"})

		// c group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"c", "3", "d", "4", "e", "5"})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "c", "3", "d", "4"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "c", "3", "e", "5"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "b", "2", "d", "4", "e", "5"})
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"a", "1", "c", "3", "d", "4", "e", "5"})

		// b group
		MatchOnce(g, generatedTagGroups, datatypes.Strings{"b", "2", "c", "3", "d", "4", "e", "5"})
	})

	t.Run("it returns all 5 pair elements as the last element", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		// a group
		lastGroup := datatypes.Strings{"a", "1", "b", "2", "c", "3", "d", "4", "e", "5"}
		MatchOnce(g, generatedTagGroups, lastGroup)
		g.Expect(generatedTagGroups[len(generatedTagGroups)-1]).To(Equal(lastGroup))
	})

	t.Run("it has the proper size", func(t *testing.T) {
		brokerInfo := setupBrokerInfo(g)
		generatedTagGroups := brokerInfo.GenerateTagPairs()

		// also matches total number of tests above
		g.Expect(len(generatedTagGroups)).To(Equal(31))
	})
}
