package btreeassociated

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Query_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	validSelection := &queryassociatedaction.AssociatedActionQuery{}
	invalidSelection := &queryassociatedaction.AssociatedActionQuery{Selection: &queryassociatedaction.Selection{}}
	onFindPagination := func(value AssociatedKeyValues) bool { return true }

	t.Run("it returns an error if the select query is invalid", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.QueryAction(invalidSelection, onFindPagination)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Selection: requires 'IDs', 'KeyValues', 'MinNumberOfKeyValues' or 'MaxNumberOfKeyValues' to be specified, but received nothing"))
	})

	t.Run("it returns an error with nil onFindPagination", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.QueryAction(validSelection, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})
}

func setupTestQueryTree(g *GomegaWithT) ([]string, *threadsafeAssociatedTree) {
	associatedTree := NewThreadSafe()
	ids := make([]string, 39)

	noOpOnFind := func(item AssociatedKeyValues) { panic("shouldn't find me during creation") }

	// create a number of entries in the associated tree

	// int values
	keyValues0 := datatypes.KeyValues{"1": datatypes.Int(1)}
	keyValues1 := datatypes.KeyValues{"2": datatypes.Int(2)}
	keyValues2 := datatypes.KeyValues{"3": datatypes.Int(3)}

	// string values
	keyValues3 := datatypes.KeyValues{"1": datatypes.String("1")}
	keyValues4 := datatypes.KeyValues{"2": datatypes.String("2")}
	keyValues5 := datatypes.KeyValues{"3": datatypes.String("3")}

	// any values
	keyValues6 := datatypes.KeyValues{"1": datatypes.Any()}
	keyValues7 := datatypes.KeyValues{"2": datatypes.Any()}
	keyValues8 := datatypes.KeyValues{"3": datatypes.Any()}

	// combination of the key and string values starting at 1
	//// 1 as int
	keyValues9 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}
	keyValues10 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}
	keyValues11 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Any()}
	keyValues12 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)}
	keyValues13 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")}
	keyValues14 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Any()}
	//// 1 as string
	keyValues15 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.Int(2)}
	keyValues16 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.String("2")}
	keyValues17 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.Any()}
	keyValues18 := datatypes.KeyValues{"1": datatypes.String("1"), "3": datatypes.Int(3)}
	keyValues19 := datatypes.KeyValues{"1": datatypes.String("1"), "3": datatypes.String("3")}
	keyValues20 := datatypes.KeyValues{"1": datatypes.String("1"), "3": datatypes.Any()}
	//// 1 as any
	keyValues21 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Int(2)}
	keyValues22 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.String("2")}
	keyValues23 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Any()}
	keyValues24 := datatypes.KeyValues{"1": datatypes.Any(), "3": datatypes.Int(3)}
	keyValues25 := datatypes.KeyValues{"1": datatypes.Any(), "3": datatypes.String("3")}
	keyValues26 := datatypes.KeyValues{"1": datatypes.Any(), "3": datatypes.Any()}

	// combination of the key and string values starting at 2
	//// 2 as int
	keyValues27 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)}
	keyValues28 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")}
	keyValues29 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Any()}
	//// 2 as string
	keyValues30 := datatypes.KeyValues{"2": datatypes.String("2"), "3": datatypes.Int(3)}
	keyValues31 := datatypes.KeyValues{"2": datatypes.String("2"), "3": datatypes.String("3")}
	keyValues32 := datatypes.KeyValues{"2": datatypes.String("2"), "3": datatypes.Any()}
	//Any()
	keyValues33 := datatypes.KeyValues{"2": datatypes.Any(), "3": datatypes.Int(3)}
	keyValues34 := datatypes.KeyValues{"2": datatypes.Any(), "3": datatypes.String("3")}
	keyValues35 := datatypes.KeyValues{"2": datatypes.Any(), "3": datatypes.Any()}

	// combination of all pure types for now. can maybe add all combinations, but think what I currently
	// have can test all queries
	keyValues36 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
	keyValues37 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.String("3")}
	keyValues38 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Any(), "3": datatypes.Any()}

	// create all the values and save the IDs. Don't need to check any errors here as those should be tested in the creation
	// tests to ensure everything is working properly
	ids[0], _ = associatedTree.CreateOrFind(keyValues0, func() any { return "0" }, noOpOnFind)
	ids[1], _ = associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
	ids[2], _ = associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
	ids[3], _ = associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
	ids[4], _ = associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)
	ids[5], _ = associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)
	ids[6], _ = associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)
	ids[7], _ = associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)
	ids[8], _ = associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)
	ids[9], _ = associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)
	ids[10], _ = associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)
	ids[11], _ = associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)
	ids[12], _ = associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)
	ids[13], _ = associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)
	ids[14], _ = associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)
	ids[15], _ = associatedTree.CreateOrFind(keyValues15, func() any { return "15" }, noOpOnFind)
	ids[16], _ = associatedTree.CreateOrFind(keyValues16, func() any { return "16" }, noOpOnFind)
	ids[17], _ = associatedTree.CreateOrFind(keyValues17, func() any { return "17" }, noOpOnFind)
	ids[18], _ = associatedTree.CreateOrFind(keyValues18, func() any { return "18" }, noOpOnFind)
	ids[19], _ = associatedTree.CreateOrFind(keyValues19, func() any { return "19" }, noOpOnFind)
	ids[20], _ = associatedTree.CreateOrFind(keyValues20, func() any { return "20" }, noOpOnFind)
	ids[21], _ = associatedTree.CreateOrFind(keyValues21, func() any { return "21" }, noOpOnFind)
	ids[22], _ = associatedTree.CreateOrFind(keyValues22, func() any { return "22" }, noOpOnFind)
	ids[23], _ = associatedTree.CreateOrFind(keyValues23, func() any { return "23" }, noOpOnFind)
	ids[24], _ = associatedTree.CreateOrFind(keyValues24, func() any { return "24" }, noOpOnFind)
	ids[25], _ = associatedTree.CreateOrFind(keyValues25, func() any { return "25" }, noOpOnFind)
	ids[26], _ = associatedTree.CreateOrFind(keyValues26, func() any { return "26" }, noOpOnFind)
	ids[27], _ = associatedTree.CreateOrFind(keyValues27, func() any { return "27" }, noOpOnFind)
	ids[28], _ = associatedTree.CreateOrFind(keyValues28, func() any { return "28" }, noOpOnFind)
	ids[29], _ = associatedTree.CreateOrFind(keyValues29, func() any { return "29" }, noOpOnFind)
	ids[30], _ = associatedTree.CreateOrFind(keyValues30, func() any { return "30" }, noOpOnFind)
	ids[31], _ = associatedTree.CreateOrFind(keyValues31, func() any { return "31" }, noOpOnFind)
	ids[32], _ = associatedTree.CreateOrFind(keyValues32, func() any { return "32" }, noOpOnFind)
	ids[33], _ = associatedTree.CreateOrFind(keyValues33, func() any { return "33" }, noOpOnFind)
	ids[34], _ = associatedTree.CreateOrFind(keyValues34, func() any { return "34" }, noOpOnFind)
	ids[35], _ = associatedTree.CreateOrFind(keyValues35, func() any { return "35" }, noOpOnFind)
	ids[36], _ = associatedTree.CreateOrFind(keyValues36, func() any { return "36" }, noOpOnFind)
	ids[37], _ = associatedTree.CreateOrFind(keyValues37, func() any { return "37" }, noOpOnFind)
	ids[38], _ = associatedTree.CreateOrFind(keyValues38, func() any { return "38" }, noOpOnFind)

	return ids, associatedTree
}

func TestAssociatedTree_Query_SelectAll(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It runs select all if each AssociatedQuery field is nil", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		foundValues := []string{}
		expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It stops pagination if the callback stop", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		count := 0
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			count++
			return false
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(1))
	})
}

func TestAssociatedTree_Query_Basic(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It does not run the callback if there are no keys that match a single query", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"not found": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		count := 0
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			count++
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(0))
	})

	t.Run("It stops pagination if the callback stop", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				MinNumberOfKeyValues: helpers.PointerOf(3),
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
					"2": {
						Value:            datatypes.Int(2),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		count := 0
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			count++
			return false
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(1))
	})

	t.Run("It respects the limits.MinNumberOfKeyValues if they are provided", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				MinNumberOfKeyValues: helpers.PointerOf(3),
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
					"2": {
						Value:            datatypes.Int(2),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It respects the limits.MaxNumberOfKeyValues if they are provided", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				MaxNumberOfKeyValues: helpers.PointerOf(2),
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
					"2": {
						Value:            datatypes.Int(2),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"9", "11", "21", "23"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("Describe EQUALS queries", func(t *testing.T) {
		t.Run("It can find all fields with a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "6", "9", "10", "11", "12", "13", "14", "21", "22", "23", "24", "25", "26", "36", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds with multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "11", "21", "23", "36", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "36"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("Context when the query uses T_any", func(t *testing.T) {
			t.Run("It respects the type restrictions", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
							"2": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"9", "36"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("It allows for all values if there are no type restrictions", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"9", "10", "11", "15", "16", "17", "21", "22", "23", "36", "37", "38"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})
		})
	})

	t.Run("Describe NOT EQUALS queries", func(t *testing.T) {
		t.Run("It can find all fields that do not contain a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.NotEquals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"1", "2", "3", "4", "5", "7", "8", "15", "16", "17", "18", "19", "20", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds not containing multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.NotEquals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.NotEquals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "12", "13", "14", "15", "16", "17", "18", "19", "20", "22", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.NotEquals,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.NotEquals,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("Context when the query uses T_any", func(t *testing.T) {
			t.Run("It respects the type restrictions", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Any(),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
							"2": {
								Value:            datatypes.Any(),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37", "38"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("It allows for all values if there are no type restrictions", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Any(),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Any(),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "12", "13", "14", "18", "19", "20", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})
		})
	})

	t.Run("Describe LESS THAN queries", func(t *testing.T) {
		t.Run("It can find all fields with a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "6", "9", "10", "11", "12", "13", "14", "21", "22", "23", "24", "25", "26", "36", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds with multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.String("2"),
							Comparison:       v1common.LessThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "11", "21", "23", "36", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThan,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
						"2": {
							Value:            datatypes.String("2"),
							Comparison:       v1common.LessThan,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "36"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})
	})

	t.Run("Describe LESS THAN OR EQUAL queries", func(t *testing.T) {
		t.Run("It can find all fields with a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "3", "6", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "36", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds with multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.String("2"),
							Comparison:       v1common.LessThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "10", "11", "15", "16", "17", "21", "22", "23", "36", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.String("1"),
							Comparison:       v1common.LessThanOrEqual,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
						"2": {
							Value:            datatypes.String("2"),
							Comparison:       v1common.LessThanOrEqual,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "10", "15", "16", "36", "37"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})
	})

	t.Run("Describe GREAT THAN queries", func(t *testing.T) {
		t.Run("It can find all fields with a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"3", "6", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds with multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.GreaterThan,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"16", "17", "22", "23", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThan,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.GreaterThan,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"16", "37"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})
	})

	t.Run("Describe GREAT THAN OR EQUAL queries", func(t *testing.T) {
		t.Run("It can find all fields with a single KeyValue", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"0", "3", "6", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "36", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It can find all fileds with multiple KeyValues", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.GreaterThanOrEqual,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "10", "11", "15", "16", "17", "21", "22", "23", "36", "37", "38"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})

		t.Run("It respects the type restrictions", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: map[string]queryassociatedaction.ValueQuery{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.GreaterThanOrEqual,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
						"2": {
							Value:            datatypes.Int(2),
							Comparison:       v1common.GreaterThanOrEqual,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_string),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			expectedValues := []string{"9", "10", "15", "16", "36", "37"}
			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
			for _, expectedValue := range expectedValues {
				g.Expect(foundValues).To(ContainElement(expectedValue))
			}
		})
	})

	t.Run("Destribe using associatedIDs in the query", func(t *testing.T) {
		t.Run("It does not run the query if no IDs were found", func(t *testing.T) {
			_, tree := setupTestQueryTree(g)

			query := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					IDs: []string{"nopes"},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(0))
		})

		t.Run("It matches the query against any found IDs", func(t *testing.T) {
			ids, tree := setupTestQueryTree(g)

			queryMatch := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					IDs: []string{ids[0]},
					KeyValues: queryassociatedaction.SelectionKeyValues{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(queryMatch.Validate()).ToNot(HaveOccurred())

			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&queryMatch, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(1))
			g.Expect(foundValues).To(Equal([]string{"0"}))
		})

		t.Run("It ignores the found IDs if the wuery does not match their key values", func(t *testing.T) {
			ids, tree := setupTestQueryTree(g)

			noMatchType := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					IDs: []string{ids[0]},
					KeyValues: queryassociatedaction.SelectionKeyValues{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_any, datatypes.T_any),
						},
					},
				},
			}
			g.Expect(noMatchType.Validate()).ToNot(HaveOccurred())

			noMatchMinKeys := queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					IDs:                  []string{ids[0]},
					MinNumberOfKeyValues: helpers.PointerOf(2),
					KeyValues: queryassociatedaction.SelectionKeyValues{
						"1": {
							Value:            datatypes.Int(1),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(noMatchMinKeys.Validate()).ToNot(HaveOccurred())

			foundValues := []string{}
			onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
				foundValues = append(foundValues, associatedKeyValues.Value().(string))
				return true
			}

			g.Expect(tree.QueryAction(&noMatchType, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(0))

			g.Expect(tree.QueryAction(&noMatchMinKeys, onFindPagination)).ToNot(HaveOccurred())
			g.Expect(len(foundValues)).To(Equal(0))
		})
	})
}

func TestAssociatedTree_Query_AND_Joins(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It intersects the AND query, with the ASsociatedSelection", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
			And: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"9", "11", "21", "23", "36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It can intersects all ids found from each of the queries", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			And: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"9", "11", "21", "23", "36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It does not find anything in the query if all intersections return 0 ids", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			And: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.LessThan,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.MinDataType, datatypes.T_uint),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})
}

func TestAssociatedTree_Query_OR_Joins(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It unions the OR query, with the AssociatedSelection", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: map[string]queryassociatedaction.ValueQuery{
					"1": {
						Value:            datatypes.Int(1),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
			Or: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"0", "1", "6", "7", "9", "10", "11", "12", "13", "14", "15", "17", "21", "22", "23", "24", "25", "26", "27", "28", "29", "33", "34", "35", "36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It unions multiple OR queries together", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Or: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"0", "1", "6", "7", "9", "10", "11", "12", "13", "14", "15", "17", "21", "22", "23", "24", "25", "26", "27", "28", "29", "33", "34", "35", "36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})

	t.Run("It ignores OR queries that return no IDs", func(t *testing.T) {
		_, tree := setupTestQueryTree(g)

		query := queryassociatedaction.AssociatedActionQuery{
			Or: []*queryassociatedaction.AssociatedActionQuery{
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
				{
					Selection: &queryassociatedaction.Selection{
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"2": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.LessThan,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		expectedValues := []string{"0", "1", "6", "7", "9", "10", "11", "12", "13", "14", "15", "17", "21", "22", "23", "24", "25", "26", "27", "28", "29", "33", "34", "35", "36", "38"}
		foundValues := []string{}
		onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
			foundValues = append(foundValues, associatedKeyValues.Value().(string))
			return true
		}

		g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
		g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
		for _, expectedValue := range expectedValues {
			g.Expect(foundValues).To(ContainElement(expectedValue))
		}
	})
}

func TestAssociated_Query_LargeJoinsAreFast(t *testing.T) {
	g := NewGomegaWithT(t)
	baicCreate := func() any { return true }
	testCounter := 10_000

	t.Run("It runs reasonably fast", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// 1. create 10k entries in the tree with up to 10 key values making up a single associated id
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			go func(numberOfKeys, tNum int) {
				defer wg.Done()
				keys := datatypes.KeyValues{}

				for i := 0; i <= numberOfKeys; i++ {
					keys[fmt.Sprintf("%d", tNum+i)] = datatypes.String(fmt.Sprintf("%d", tNum+i))
				}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, baicCreate)
				g.Expect(err).ToNot(HaveOccurred())
			}(i%10, i)
		}
		wg.Wait()

		// 2. generate a list of 10 key value pairs
		largeKeyValues := datatypes.KeyValues{}
		for i := 0; i < 10; i++ {
			largeKeyValues[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", i))
		}

		// 3. each key value is its own or join query
		query := &queryassociatedaction.AssociatedActionQuery{}
		for _, keyValues := range largeKeyValues.GenerateTagPairs() {
			associatedQuery := &queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: queryassociatedaction.SelectionKeyValues{},
				},
			}

			for key, value := range keyValues {
				associatedQuery.Selection.KeyValues[key] = queryassociatedaction.ValueQuery{Value: value, Comparison: v1common.Equals, TypeRestrictions: testmodels.NoTypeRestrictions(g)}
			}

			query.Or = append(query.Or, associatedQuery)
		}

		// 4. run the query
		counter := 0
		g.Expect(associatedTree.QueryAction(query, func(item AssociatedKeyValues) bool {
			counter++
			return true
		})).ToNot(HaveOccurred())

		g.Expect(counter).To(Equal(10))
	})
}

/*
	t.Run("Describe when using permutations", func(t *testing.T) {
		t.Run("Context PermutationsAll", func(t *testing.T) {
			t.Run("It respects the MinNumberOfPermutationKeyValues field", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						Permutations:                    v1common.PermutationsAll,
						MinNumberOfPermutationKeyValues: helpers.PointerOf(2),
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"3": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"9", "11", "12", "13", "14", "21", "23", "24", "25", "26", "27", "28", "29", "33", "34", "35", "36", "38"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("It respects the MaxNumberOfPermutationKeyValues field", func(t *testing.T) {
				// [[1,2], [1,3], [2,3]]
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						Permutations:                    v1common.PermutationsAll,
						MinNumberOfPermutationKeyValues: helpers.PointerOf(2),
						MaxNumberOfPermutationKeyValues: helpers.PointerOf(2),
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"3": {
								Value:            datatypes.Any(),
								Comparison:       v1common.NotEquals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "30", "31", "32", "36", "37", "38"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("Context when all queries are finding keys specifically", func(t *testing.T) {
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"3": {
									Value:            datatypes.Any(),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"3": {
									Value:            datatypes.Any(),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "5", "8", "9", "10", "11", "12", "13", "14", "15", "18", "19", "20", "21", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})

			t.Run("Context when queries contain relations for NOT EQUALS and OTHER queries", func(t *testing.T) {
				// This tests query miught seem a bit odd, but it proves out that stiff like:
				//		`datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}`
				// still show up in the query. This is because the KeyValues satisfy just the `"1"` permutation
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "2", "3", "4", "5", "6", "8", "9", "10", "11", "12", "13", "14", "16", "18", "19", "20", "21", "22", "23", "24", "25", "26", "30", "31", "32", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "16", "17", "18", "19", "20", "22", "23", "24", "25", "26", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})

			t.Run("Context when queries contain only NOT EQUALS", func(t *testing.T) {
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "12", "13", "14", "15", "16", "17", "18", "19", "20", "22", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsAll,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})
		})

		t.Run("Context PermutationsExact", func(t *testing.T) {
			t.Run("It respects the MinNumberOfPermutationKeyValues field", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						Permutations:                    v1common.PermutationsExact,
						MinNumberOfPermutationKeyValues: helpers.PointerOf(2),
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"3": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"9", "11", "12", "13", "14", "21", "23", "24", "25", "26", "27", "28", "29", "33", "34", "35"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("It respects the MaxNumberOfPermutationKeyValues field", func(t *testing.T) {
				_, tree := setupTestQueryTree(g)

				query := queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						Permutations:                    v1common.PermutationsExact,
						MinNumberOfPermutationKeyValues: helpers.PointerOf(1),
						MaxNumberOfPermutationKeyValues: helpers.PointerOf(2),
						KeyValues: map[string]queryassociatedaction.ValueQuery{
							"1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
							"3": {
								Value:            datatypes.Any(),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.NoTypeRestrictions(g),
							},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				expectedValues := []string{"0", "1", "2", "6", "7", "8", "9", "11", "12", "13", "14", "21", "23", "24", "25", "26", "27", "28", "29"}
				foundValues := []string{}
				onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
					foundValues = append(foundValues, associatedKeyValues.Value().(string))
					return true
				}

				g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
				g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
				for _, expectedValue := range expectedValues {
					g.Expect(foundValues).To(ContainElement(expectedValue))
				}
			})

			t.Run("Context when all queries are finding keys specifically", func(t *testing.T) {
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"3": {
									Value:            datatypes.Any(),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"3": {
									Value:            datatypes.Any(),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "5", "8", "9", "12", "13", "14", "27", "28", "29", "36"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})

			t.Run("Context when queries contain relations for NOT EQUALS and OTHER queries", func(t *testing.T) {
				// This tests query miught seem a bit odd, but it proves out that stiff like:
				//		`datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}`
				// still show up in the query. This is because the KeyValues satisfy just the `"1"` permutation
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "2", "3", "4", "5", "6", "8", "9", "10", "11", "12", "13", "14", "16", "18", "19", "20", "21", "22", "23", "24", "25", "26", "30", "31", "32", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.Equals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "16", "17", "18", "19", "20", "22", "23", "24", "25", "26", "30", "31", "32", "33", "34", "35", "36", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})

			t.Run("Context when queries contain only NOT EQUALS", func(t *testing.T) {
				t.Run("It can find all permutaions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.NoTypeRestrictions(g),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "12", "13", "14", "15", "16", "17", "18", "19", "20", "22", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})

				t.Run("It respects the TypeRestrictions", func(t *testing.T) {
					_, tree := setupTestQueryTree(g)

					query := queryassociatedaction.AssociatedActionQuery{
						Selection: &queryassociatedaction.Selection{
							Permutations: v1common.PermutationsExact,
							KeyValues: map[string]queryassociatedaction.ValueQuery{
								"1": {
									Value:            datatypes.Int(1),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
								"2": {
									Value:            datatypes.Int(2),
									Comparison:       v1common.NotEquals,
									TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
								},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					expectedValues := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "37", "38"}
					foundValues := []string{}
					onFindPagination := func(associatedKeyValues AssociatedKeyValues) bool {
						foundValues = append(foundValues, associatedKeyValues.Value().(string))
						return true
					}

					g.Expect(tree.QueryAction(&query, onFindPagination)).ToNot(HaveOccurred())
					g.Expect(len(foundValues)).To(Equal(len(expectedValues)))
					for _, expectedValue := range expectedValues {
						g.Expect(foundValues).To(ContainElement(expectedValue))
					}
				})
			})
		})
	})
*/
