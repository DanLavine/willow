package datatypes

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []KeyValues, val KeyValues) {
	count := 0
	for _, groupVal := range group {
		if reflect.DeepEqual(groupVal, val) {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %v", line, val))
}

func TestKeyValues_Keys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map", func(t *testing.T) {
		KeyValues := KeyValues{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := KeyValues.Keys()
		g.Expect(keys).To(ContainElements([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestKeyValues_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map in a sorted order", func(t *testing.T) {
		KeyValues := KeyValues{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := KeyValues.SoretedKeys()
		g.Expect(keys).To(Equal([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestKeyValues_GenerateTagPairs(t *testing.T) {
	g := NewGomegaWithT(t)

	setupKeyValues := func(g *GomegaWithT) KeyValues {
		return KeyValues{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": String("4"),
			"e": String("5"),
		}
	}

	t.Run("it returns all individual elements", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1")})
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2")})
		MatchOnce(g, generatedTagGroups, KeyValues{"c": String("3")})
		MatchOnce(g, generatedTagGroups, KeyValues{"d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"e": String("5")})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "e": String("5")})

		// c group
		MatchOnce(g, generatedTagGroups, KeyValues{"c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"c": String("3"), "e": String("5")})

		// d group
		MatchOnce(g, generatedTagGroups, KeyValues{"d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "d": String("4"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "d": String("4"), "e": String("5")})

		// c group
		MatchOnce(g, generatedTagGroups, KeyValues{"c": String("3"), "d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "b": String("2"), "d": String("4"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, KeyValues{"a": String("1"), "c": String("3"), "d": String("4"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, KeyValues{"b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 5 pair elements as the last element", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		// a group
		lastGroup := KeyValues{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")}
		MatchOnce(g, generatedTagGroups, lastGroup)
		g.Expect(generatedTagGroups[len(generatedTagGroups)-1]).To(Equal(lastGroup))
	})

	t.Run("it has the proper size", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		// also matches total number of tests above
		g.Expect(len(generatedTagGroups)).To(Equal(31))
	})

	t.Run("it returns all elements in a sorted order", func(t *testing.T) {
		setupKeyValues := setupKeyValues(g)
		generatedTagGroups := setupKeyValues.GenerateTagPairs()

		expectedTags := []KeyValues{
			{"a": String("1")},
			{"b": String("2")},
			{"c": String("3")},
			{"d": String("4")},
			{"e": String("5")},
			{"a": String("1"), "b": String("2")},
			{"a": String("1"), "c": String("3")},
			{"a": String("1"), "d": String("4")},
			{"a": String("1"), "e": String("5")},
			{"b": String("2"), "c": String("3")},
			{"b": String("2"), "d": String("4")},
			{"b": String("2"), "e": String("5")},
			{"c": String("3"), "d": String("4")},
			{"c": String("3"), "e": String("5")},
			{"d": String("4"), "e": String("5")},
			{"a": String("1"), "b": String("2"), "c": String("3")},
			{"a": String("1"), "b": String("2"), "d": String("4")},
			{"a": String("1"), "b": String("2"), "e": String("5")},
			{"a": String("1"), "c": String("3"), "d": String("4")},
			{"a": String("1"), "c": String("3"), "e": String("5")},
			{"a": String("1"), "d": String("4"), "e": String("5")},
			{"b": String("2"), "c": String("3"), "d": String("4")},
			{"b": String("2"), "c": String("3"), "e": String("5")},
			{"b": String("2"), "d": String("4"), "e": String("5")},
			{"c": String("3"), "d": String("4"), "e": String("5")},
			{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4")},
			{"a": String("1"), "b": String("2"), "c": String("3"), "e": String("5")},
			{"a": String("1"), "b": String("2"), "d": String("4"), "e": String("5")},
			{"a": String("1"), "c": String("3"), "d": String("4"), "e": String("5")},
			{"b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")},
			{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")},
		}

		// also matches total number of tests above
		for index, value := range expectedTags {
			g.Expect(generatedTagGroups[index]).To(Equal(value), fmt.Sprintf("index: %d", index))
		}
	})
}