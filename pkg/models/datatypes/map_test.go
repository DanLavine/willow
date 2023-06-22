package datatypes

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []StringMap, val StringMap) {
	count := 0
	for _, groupVal := range group {
		if reflect.DeepEqual(groupVal, val) {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %v", line, val))
}

func TestStringMap_Keys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map", func(t *testing.T) {
		stringMap := StringMap{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := stringMap.Keys()
		g.Expect(keys).To(ContainElements([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestStringMap_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map in a sorted order", func(t *testing.T) {
		stringMap := StringMap{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := stringMap.SoretedKeys()
		g.Expect(keys).To(Equal([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestStringMap_GenerateTagPairs(t *testing.T) {
	g := NewGomegaWithT(t)

	setupStringMap := func(g *GomegaWithT) StringMap {
		return StringMap{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": String("4"),
			"e": String("5"),
		}
	}

	t.Run("it returns all individual elements", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1")})
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2")})
		MatchOnce(g, generatedTagGroups, StringMap{"c": String("3")})
		MatchOnce(g, generatedTagGroups, StringMap{"d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"e": String("5")})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "e": String("5")})

		// c group
		MatchOnce(g, generatedTagGroups, StringMap{"c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"c": String("3"), "e": String("5")})

		// d group
		MatchOnce(g, generatedTagGroups, StringMap{"d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		fmt.Printf("generated string map len: %#v\n", len(generatedTagGroups))
		fmt.Printf("generated string map: %#v\n", generatedTagGroups)

		// a group
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "c": String("3")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "d": String("4"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "d": String("4"), "e": String("5")})

		// c group
		MatchOnce(g, generatedTagGroups, StringMap{"c": String("3"), "d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		// a group
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "c": String("3"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "b": String("2"), "d": String("4"), "e": String("5")})
		MatchOnce(g, generatedTagGroups, StringMap{"a": String("1"), "c": String("3"), "d": String("4"), "e": String("5")})

		// b group
		MatchOnce(g, generatedTagGroups, StringMap{"b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")})
	})

	t.Run("it returns all 5 pair elements as the last element", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		// a group
		lastGroup := StringMap{"a": String("1"), "b": String("2"), "c": String("3"), "d": String("4"), "e": String("5")}
		MatchOnce(g, generatedTagGroups, lastGroup)
		g.Expect(generatedTagGroups[len(generatedTagGroups)-1]).To(Equal(lastGroup))
	})

	t.Run("it has the proper size", func(t *testing.T) {
		stringMap := setupStringMap(g)
		generatedTagGroups := stringMap.GenerateTagPairs()

		// also matches total number of tests above
		g.Expect(len(generatedTagGroups)).To(Equal(31))
	})
}
