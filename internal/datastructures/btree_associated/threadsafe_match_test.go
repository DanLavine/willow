package btreeassociated

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestAssociated_Match(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"key1": datatypes.String("")})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"key2": datatypes.String("")})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")})

		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"key1": datatypes.String("one")})
		keyValues5 := ConverDatatypesKeyValues(datatypes.KeyValues{"key1": datatypes.String("one"), "key2": datatypes.String("two")})

		keyValues6 := ConverDatatypesKeyValues(datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")})
		keyValues7 := ConverDatatypesKeyValues(datatypes.KeyValues{"name": datatypes.String("rule7"), "key1": datatypes.String("")})
		keyValues8 := ConverDatatypesKeyValues(datatypes.KeyValues{"name": datatypes.String("rule8"), "key1": datatypes.String(""), "key2": datatypes.String("")})
		keyValues9 := ConverDatatypesKeyValues(datatypes.KeyValues{"name": datatypes.String("rule9"), "key1": datatypes.String(""), "key2": datatypes.String("")})

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

	t.Run("Context a single KeyValue", func(t *testing.T) {
		t.Run("It matchese only exact entries", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			keyValues := KeyValues{datatypes.String("key1"): datatypes.String("")}

			onQueryCount := 0
			onQueryPagination := func(item *AssociatedKeyValues) bool {
				onQueryCount++
				g.Expect(item.KeyValues()[datatypes.String("key1")]).To(Equal(datatypes.String("")))
				g.Expect(item.value).To(Equal("1"))
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

			keyValues := KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String("")}

			onQueryCount := 0
			foundKeyValues := []KeyValues{}
			onQueryPagination := func(item *AssociatedKeyValues) bool {
				onQueryCount++
				foundKeyValues = append(foundKeyValues, item.keyValues.StripAssociatedID())
				return true
			}

			g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
			g.Expect(onQueryCount).To(Equal(3))
			g.Expect(len(foundKeyValues)).To(Equal(3))
			g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key2"): datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String("")}))
		})

		t.Run("Context when there are keys for values that are not saved in the tree", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				keyValues := KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String(""), datatypes.String("unkown"): datatypes.String("")}

				onQueryCount := 0
				foundKeyValues := []KeyValues{}
				onQueryPagination := func(item *AssociatedKeyValues) bool {
					onQueryCount++
					foundKeyValues = append(foundKeyValues, item.keyValues.StripAssociatedID())
					return true
				}

				g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
				fmt.Println(foundKeyValues)
				g.Expect(onQueryCount).To(Equal(3))
				g.Expect(len(foundKeyValues)).To(Equal(3))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key2"): datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String("")}))
			})
		})

		t.Run("Context when there are keys for values that are a combination of single and multi value entries", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				keyValues := KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String(""), datatypes.String("name"): datatypes.String("rule6")}

				onQueryCount := 0
				foundKeyValues := []KeyValues{}
				onQueryPagination := func(item *AssociatedKeyValues) bool {
					onQueryCount++
					foundKeyValues = append(foundKeyValues, item.keyValues.StripAssociatedID())
					return true
				}

				g.Expect(asscoiatedTree.MatchPermutations(keyValues, onQueryPagination))
				g.Expect(onQueryCount).To(Equal(4))
				g.Expect(len(foundKeyValues)).To(Equal(4))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("name"): datatypes.String("rule6"), datatypes.String("key1"): datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key2"): datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(KeyValues{datatypes.String("key1"): datatypes.String(""), datatypes.String("key2"): datatypes.String("")}))
			})
		})
	})
}
