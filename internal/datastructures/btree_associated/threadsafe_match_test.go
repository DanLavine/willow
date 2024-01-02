package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestAssociated_Match(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.KeyValues{"key1": datatypes.String("")}
		keyValues2 := datatypes.KeyValues{"key2": datatypes.String("")}
		keyValues3 := datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}

		keyValues4 := datatypes.KeyValues{"key1": datatypes.String("one")}
		keyValues5 := datatypes.KeyValues{"key1": datatypes.String("one"), "key2": datatypes.String("two")}

		keyValues6 := datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}
		keyValues7 := datatypes.KeyValues{"name": datatypes.String("rule7"), "key1": datatypes.String("")}
		keyValues8 := datatypes.KeyValues{"name": datatypes.String("rule8"), "key1": datatypes.String(""), "key2": datatypes.String("")}
		keyValues9 := datatypes.KeyValues{"name": datatypes.String("rule9"), "key1": datatypes.String(""), "key2": datatypes.String("")}

		g.Expect(associatedTree.CreateWithID("1", keyValues1, func() any { return "1" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("2", keyValues2, func() any { return "2" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("3", keyValues3, func() any { return "3" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("4", keyValues4, func() any { return "4" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("5", keyValues5, func() any { return "5" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule6", keyValues6, func() any { return "6" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule7", keyValues7, func() any { return "6" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule8", keyValues8, func() any { return "6" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule9", keyValues9, func() any { return "6" })).ToNot(HaveOccurred())

		return associatedTree
	}

	// TODO: add tests about deletion of associated ids

	t.Run("Context a single KeyValue", func(t *testing.T) {
		t.Run("It matchese only exact entries", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			keyValues := datatypes.KeyValues{"key1": datatypes.String("")}

			onQueryCount := 0
			onQueryPagination := func(item AssociatedKeyValues) bool {
				onQueryCount++
				g.Expect(item.KeyValues()["key1"]).To(Equal(datatypes.String("")))
				g.Expect(item.Value()).To(Equal("1"))
				return true
			}

			g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
			g.Expect(onQueryCount).To(Equal(1))
		})
	})

	t.Run("Context multiple KeyValue", func(t *testing.T) {
		t.Run("It matchese all permutations of the key values", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			keyValues := datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}

			onQueryCount := 0
			foundKeyValues := []datatypes.KeyValues{}
			onQueryPagination := func(item AssociatedKeyValues) bool {
				onQueryCount++
				foundKeyValues = append(foundKeyValues, item.KeyValues())
				return true
			}

			g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
			g.Expect(onQueryCount).To(Equal(3))
			g.Expect(len(foundKeyValues)).To(Equal(3))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
		})

		t.Run("Context when there are keys for values that are not saved in the tree", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				keyValues := datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String(""), "unkown": datatypes.String("")}

				onQueryCount := 0
				foundKeyValues := []datatypes.KeyValues{}
				onQueryPagination := func(item AssociatedKeyValues) bool {
					onQueryCount++
					foundKeyValues = append(foundKeyValues, item.KeyValues())
					return true
				}

				g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
				g.Expect(onQueryCount).To(Equal(3))
				g.Expect(len(foundKeyValues)).To(Equal(3))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
			})
		})

		t.Run("Context when there are keys for values that are a combination of single and multi value entries", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				keyValues := datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String(""), "name": datatypes.String("rule6")}

				onQueryCount := 0
				foundKeyValues := []datatypes.KeyValues{}
				onQueryPagination := func(item AssociatedKeyValues) bool {
					onQueryCount++
					foundKeyValues = append(foundKeyValues, item.KeyValues())
					return true
				}

				g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
				g.Expect(onQueryCount).To(Equal(4))
				g.Expect(len(foundKeyValues)).To(Equal(4))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
			})
		})
	})
}
