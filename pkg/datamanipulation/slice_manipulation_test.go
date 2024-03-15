package datamanipulation

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_AssociatedSelection_GenerateKeyPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	MatchOnce := func(g *GomegaWithT, group [][]string, val []string) {
		count := 0
		for _, groupVal := range group {
			if reflect.DeepEqual(groupVal, val) {
				count++
			}
		}

		_, _, line, _ := runtime.Caller(1)
		g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %v", line, val))
	}

	setupStringSlice := func() []string {
		return []string{"a", "b", "c", "d", "e"}
	}

	t.Run("it returns all individual keys in the query", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		MatchOnce(g, generatedStringPermutations, []string{"a"})
		MatchOnce(g, generatedStringPermutations, []string{"b"})
		MatchOnce(g, generatedStringPermutations, []string{"c"})
		MatchOnce(g, generatedStringPermutations, []string{"d"})
		MatchOnce(g, generatedStringPermutations, []string{"e"})
	})

	t.Run("it returns all 2 pair elements", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		// a group
		MatchOnce(g, generatedStringPermutations, []string{"a", "b"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "c"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "e"})

		// b group
		MatchOnce(g, generatedStringPermutations, []string{"b", "c"})
		MatchOnce(g, generatedStringPermutations, []string{"b", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"b", "e"})

		// c group
		MatchOnce(g, generatedStringPermutations, []string{"c", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"c", "e"})

		// d group
		MatchOnce(g, generatedStringPermutations, []string{"d", "e"})
	})

	t.Run("it returns all 3 pair elements", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		// a group
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "c"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "e"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "c", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "c", "e"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "d", "e"})

		// b group
		MatchOnce(g, generatedStringPermutations, []string{"b", "c", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"b", "c", "e"})
		MatchOnce(g, generatedStringPermutations, []string{"b", "d", "e"})

		// c group
		MatchOnce(g, generatedStringPermutations, []string{"c", "d", "e"})
	})

	t.Run("it returns all 4 pair elements", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		// a group
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "c", "d"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "c", "e"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "d", "e"})
		MatchOnce(g, generatedStringPermutations, []string{"a", "c", "d", "e"})

		// b group
		MatchOnce(g, generatedStringPermutations, []string{"b", "c", "d", "e"})
	})

	t.Run("it returns all 5 pair elements as the last element", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		// group of 5
		MatchOnce(g, generatedStringPermutations, []string{"a", "b", "c", "d", "e"})
	})

	t.Run("it has the proper size", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		// also matches total number of tests above
		g.Expect(len(generatedStringPermutations)).To(Equal(31))
	})

	t.Run("it returns all elements in a sorted order", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		expectedTags := [][]string{
			{"a"},
			{"b"},
			{"c"},
			{"d"},
			{"e"},
			{"a", "b"},
			{"a", "c"},
			{"a", "d"},
			{"a", "e"},
			{"b", "c"},
			{"b", "d"},
			{"b", "e"},
			{"c", "d"},
			{"c", "e"},
			{"d", "e"},
			{"a", "b", "c"},
			{"a", "b", "d"},
			{"a", "b", "e"},
			{"a", "c", "d"},
			{"a", "c", "e"},
			{"a", "d", "e"},
			{"b", "c", "d"},
			{"b", "c", "e"},
			{"b", "d", "e"},
			{"c", "d", "e"},
			{"a", "b", "c", "d"},
			{"a", "b", "c", "e"},
			{"a", "b", "d", "e"},
			{"a", "c", "d", "e"},
			{"b", "c", "d", "e"},
			{"a", "b", "c", "d", "e"},
		}

		// also matches total number of tests above
		for index, value := range expectedTags {
			g.Expect(generatedStringPermutations[index]).To(Equal(value), fmt.Sprintf("index: %d", index))
		}
	})

	t.Run("It restricts the key values for the MinNumberOfKeys", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 2, len(strings))
		g.Expect(err).ToNot(HaveOccurred())

		expectedTags := [][]string{
			{"a", "b"},
			{"a", "c"},
			{"a", "d"},
			{"a", "e"},
			{"b", "c"},
			{"b", "d"},
			{"b", "e"},
			{"c", "d"},
			{"c", "e"},
			{"d", "e"},
			{"a", "b", "c"},
			{"a", "b", "d"},
			{"a", "b", "e"},
			{"a", "c", "d"},
			{"a", "c", "e"},
			{"a", "d", "e"},
			{"b", "c", "d"},
			{"b", "c", "e"},
			{"b", "d", "e"},
			{"c", "d", "e"},
			{"a", "b", "c", "d"},
			{"a", "b", "c", "e"},
			{"a", "b", "d", "e"},
			{"a", "c", "d", "e"},
			{"b", "c", "d", "e"},
			{"a", "b", "c", "d", "e"},
		}

		g.Expect(len(generatedStringPermutations)).To(Equal(len(expectedTags)))

		// also matches total number of tests above
		for index, value := range expectedTags {
			g.Expect(generatedStringPermutations[index]).To(Equal(value), fmt.Sprintf("index: %d", index))
		}
	})

	t.Run("It restricts the key values for the MaxNumberOfKeys", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 0, 2)
		g.Expect(err).ToNot(HaveOccurred())

		expectedTags := [][]string{
			{"a"},
			{"b"},
			{"c"},
			{"d"},
			{"e"},
			{"a", "b"},
			{"a", "c"},
			{"a", "d"},
			{"a", "e"},
			{"b", "c"},
			{"b", "d"},
			{"b", "e"},
			{"c", "d"},
			{"c", "e"},
			{"d", "e"},
		}

		g.Expect(len(generatedStringPermutations)).To(Equal(len(expectedTags)))

		// also matches total number of tests above
		for index, value := range expectedTags {
			g.Expect(generatedStringPermutations[index]).To(Equal(value), fmt.Sprintf("index: %d", index))
		}
	})

	t.Run("It allows MazNumberOfKeys to be -1 for unlimited", func(t *testing.T) {
		strings := setupStringSlice()
		generatedStringPermutations, err := GenerateStringPermutations(strings, 2, -1)
		g.Expect(err).ToNot(HaveOccurred())

		expectedTags := [][]string{
			{"a", "b"},
			{"a", "c"},
			{"a", "d"},
			{"a", "e"},
			{"b", "c"},
			{"b", "d"},
			{"b", "e"},
			{"c", "d"},
			{"c", "e"},
			{"d", "e"},
			{"a", "b", "c"},
			{"a", "b", "d"},
			{"a", "b", "e"},
			{"a", "c", "d"},
			{"a", "c", "e"},
			{"a", "d", "e"},
			{"b", "c", "d"},
			{"b", "c", "e"},
			{"b", "d", "e"},
			{"c", "d", "e"},
			{"a", "b", "c", "d"},
			{"a", "b", "c", "e"},
			{"a", "b", "d", "e"},
			{"a", "c", "d", "e"},
			{"b", "c", "d", "e"},
			{"a", "b", "c", "d", "e"},
		}

		g.Expect(len(generatedStringPermutations)).To(Equal(len(expectedTags)))

		// also matches total number of tests above
		for index, value := range expectedTags {
			g.Expect(generatedStringPermutations[index]).To(Equal(value), fmt.Sprintf("index: %d", index))
		}
	})

}
