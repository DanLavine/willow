package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"
	. "github.com/onsi/gomega"
)

func TestAssociated_MatchAction(t *testing.T) {
	g := NewGomegaWithT(t)

	setupTree := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues0 := datatypes.KeyValues{"key1": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"key1": datatypes.String("")}
		keyValues2 := datatypes.KeyValues{"key2": datatypes.String("")}
		keyValues3 := datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}

		keyValues4 := datatypes.KeyValues{"key1": datatypes.String("one")}
		keyValues5 := datatypes.KeyValues{"key1": datatypes.String("one"), "key2": datatypes.String("two")}

		keyValues6 := datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}
		keyValues7 := datatypes.KeyValues{"name": datatypes.String("rule7"), "key1": datatypes.String("")}
		keyValues8 := datatypes.KeyValues{"name": datatypes.String("rule8"), "key1": datatypes.String(""), "key2": datatypes.String("")}
		keyValues9 := datatypes.KeyValues{"name": datatypes.String("rule9"), "key1": datatypes.String(""), "key2": datatypes.String("")}

		g.Expect(associatedTree.CreateWithID("0", keyValues0, func() any { return "10" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("1", keyValues1, func() any { return "1" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("2", keyValues2, func() any { return "2" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("3", keyValues3, func() any { return "3" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("4", keyValues4, func() any { return "4" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("5", keyValues5, func() any { return "5" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule6", keyValues6, func() any { return "6" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule7", keyValues7, func() any { return "7" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule8", keyValues8, func() any { return "8" })).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("rule9", keyValues9, func() any { return "9" })).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("Context a single KeyValue", func(t *testing.T) {
		t.Run("It matchese only exact entries", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			matchActionQuery := &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"key1": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			}

			foundKeyValues := []datatypes.KeyValues{}
			onQueryPagination := func(item AssociatedKeyValues) bool {
				foundKeyValues = append(foundKeyValues, item.KeyValues())
				return true
			}

			g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
			g.Expect(len(foundKeyValues)).To(Equal(2))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
		})
	})

	t.Run("Context multiple KeyValue", func(t *testing.T) {
		t.Run("It matchese all permutations of the key values", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			matchActionQuery := &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"key1": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
					"key2": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			}

			foundKeyValues := []datatypes.KeyValues{}
			onQueryPagination := func(item AssociatedKeyValues) bool {
				foundKeyValues = append(foundKeyValues, item.KeyValues())
				return true
			}

			g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
			g.Expect(len(foundKeyValues)).To(Equal(4))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
		})

		t.Run("Context when there are keys for values that are not saved in the tree", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				matchActionQuery := &querymatchaction.MatchActionQuery{
					KeyValues: querymatchaction.MatchKeyValues{
						"key1": querymatchaction.MatchValue{
							Value:            datatypes.String(""),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"key2": querymatchaction.MatchValue{
							Value:            datatypes.String(""),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"unkown": querymatchaction.MatchValue{
							Value:            datatypes.String(""),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				}

				foundKeyValues := []datatypes.KeyValues{}
				onQueryPagination := func(item AssociatedKeyValues) bool {
					foundKeyValues = append(foundKeyValues, item.KeyValues())
					return true
				}

				g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
				g.Expect(len(foundKeyValues)).To(Equal(4))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
			})
		})

		t.Run("Context when there are keys for values that are a combination of single and multi value entries", func(t *testing.T) {
			t.Run("It matchese all permutations of the key values that are known", func(t *testing.T) {
				t.Parallel()
				asscoiatedTree := setupTree(g)

				matchActionQuery := &querymatchaction.MatchActionQuery{
					KeyValues: querymatchaction.MatchKeyValues{
						"key1": querymatchaction.MatchValue{
							Value:            datatypes.String(""),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"key2": querymatchaction.MatchValue{
							Value:            datatypes.String(""),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"name": querymatchaction.MatchValue{
							Value:            datatypes.String("rule6"),
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				}

				foundKeyValues := []datatypes.KeyValues{}
				onQueryPagination := func(item AssociatedKeyValues) bool {
					foundKeyValues = append(foundKeyValues, item.KeyValues())
					return true
				}

				g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
				g.Expect(len(foundKeyValues)).To(Equal(5))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
				g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
			})
		})
	})

	t.Run("Context type restrictions", func(t *testing.T) {
		t.Run("It enforces specific types", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			matchActionQueryStrict := &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"key1": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.MinDataType, datatypes.T_string),
					},
					"key2": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.MinDataType, datatypes.T_string),
					},
					"name": querymatchaction.MatchValue{
						Value:            datatypes.String("rule6"),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.MinDataType, datatypes.T_string),
					},
				},
			}

			foundKeyValues := []datatypes.KeyValues{}
			onQueryPagination := func(item AssociatedKeyValues) bool {
				foundKeyValues = append(foundKeyValues, item.KeyValues())
				return true
			}

			g.Expect(asscoiatedTree.MatchAction(matchActionQueryStrict, onQueryPagination))
			g.Expect(len(foundKeyValues)).To(Equal(4))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
		})

		t.Run("It enforces any type", func(t *testing.T) {
			t.Parallel()
			asscoiatedTree := setupTree(g)

			matchActionQueryStrict := &querymatchaction.MatchActionQuery{
				KeyValues: querymatchaction.MatchKeyValues{
					"key1": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_any, datatypes.T_any),
					},
					"key2": querymatchaction.MatchValue{
						Value:            datatypes.String(""),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_any, datatypes.T_any),
					},
					"name": querymatchaction.MatchValue{
						Value:            datatypes.String("rule6"),
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_any, datatypes.T_any),
					},
				},
			}

			foundKeyValues := []datatypes.KeyValues{}
			onQueryPagination := func(item AssociatedKeyValues) bool {
				foundKeyValues = append(foundKeyValues, item.KeyValues())
				return true
			}

			g.Expect(asscoiatedTree.MatchAction(matchActionQueryStrict, onQueryPagination))
			g.Expect(len(foundKeyValues)).To(Equal(1))
			g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
		})
	})

	t.Run("It respects the MinNumberOfPermutationKeyValues", func(t *testing.T) {
		t.Parallel()
		asscoiatedTree := setupTree(g)

		matchActionQuery := &querymatchaction.MatchActionQuery{
			MinNumberOfPermutationKeyValues: helpers.PointerOf(2),
			KeyValues: querymatchaction.MatchKeyValues{
				"key1": querymatchaction.MatchValue{
					Value:            datatypes.String(""),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
				"key2": querymatchaction.MatchValue{
					Value:            datatypes.String(""),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
				"name": querymatchaction.MatchValue{
					Value:            datatypes.String("rule6"),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
			},
		}

		foundKeyValues := []datatypes.KeyValues{}
		onQueryPagination := func(item AssociatedKeyValues) bool {
			foundKeyValues = append(foundKeyValues, item.KeyValues())
			return true
		}

		g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
		g.Expect(len(foundKeyValues)).To(Equal(2))
		g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"name": datatypes.String("rule6"), "key1": datatypes.String("")}))
		g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String(""), "key2": datatypes.String("")}))
	})

	t.Run("It respects the MaxNumberOfPermutationKeyValues", func(t *testing.T) {
		t.Parallel()
		asscoiatedTree := setupTree(g)

		matchActionQuery := &querymatchaction.MatchActionQuery{
			MaxNumberOfPermutationKeyValues: helpers.PointerOf(1),
			KeyValues: querymatchaction.MatchKeyValues{
				"key1": querymatchaction.MatchValue{
					Value:            datatypes.String(""),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
				"key2": querymatchaction.MatchValue{
					Value:            datatypes.String(""),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
				"name": querymatchaction.MatchValue{
					Value:            datatypes.String("rule6"),
					TypeRestrictions: testmodels.NoTypeRestrictions(g),
				},
			},
		}

		foundKeyValues := []datatypes.KeyValues{}
		onQueryPagination := func(item AssociatedKeyValues) bool {
			foundKeyValues = append(foundKeyValues, item.KeyValues())
			return true
		}

		g.Expect(asscoiatedTree.MatchAction(matchActionQuery, onQueryPagination))
		g.Expect(len(foundKeyValues)).To(Equal(3))
		g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.Any()}))
		g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key1": datatypes.String("")}))
		g.Expect(foundKeyValues).To(ContainElement(datatypes.KeyValues{"key2": datatypes.String("")}))
	})
}
