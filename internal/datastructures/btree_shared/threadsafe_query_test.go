package btreeshared

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Query_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	validSelection := query.Select{}
	invalidSelection := query.Select{Where: &query.Query{}}
	onFindSelection := func(items []any) {}

	t.Run("it returns an error if the select query is invalid", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Query(invalidSelection, onFindSelection)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Requires KeyValues or Limits parameters"))
	})

	t.Run("it returns an error with nil onFindSelection", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Query(validSelection, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFindSelection cannot be nil"))
	})
}

func TestAssociatedTree_Query(t *testing.T) {
	g := NewGomegaWithT(t)

	existsTrue := true
	existsFalse := false
	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"2": datatypes.String("2")}
		keyValues3 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
		keyValues4 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)}
		keyValues5 := datatypes.StringMap{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)}
		keyValues6 := datatypes.StringMap{"1": datatypes.Int8(4), "2": datatypes.Float32(3.4)}
		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("Describe Select ALL query", func(t *testing.T) {
		t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItems []any
			onFindSelection := func(items []any) {
				foundItems = items
			}

			err := associatedTree.Query(query.Select{}, onFindSelection)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))
		})
	})

	t.Run("Describe Select WHERE.Exists query", func(t *testing.T) {
		t.Run("Context when ExistsType is nil", func(t *testing.T) {
			setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
				associatedTree := NewThreadSafe()

				// create a number of entries in the associated tree
				keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
				keyValues2 := datatypes.StringMap{"2": datatypes.String("2")}
				keyValues3 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
				keyValues4 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)}
				keyValues5 := datatypes.StringMap{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)}
				keyValues6 := datatypes.StringMap{"1": datatypes.Int8(4), "2": datatypes.Float32(3.4)}
				g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())

				return associatedTree
			}

			t.Run("Context when Exists == True", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsTrue},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(4))
					g.Expect(foundItems).To(ContainElements("1", "3", "4", "6"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue},
								},
								Limits: &query.KeyLimits{
									NumberOfKeys: &one,
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("1"))
					})
				})

				t.Run("Context when there are multiple TRUE statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue},
									"2": {Exists: &existsTrue},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(3))
						g.Expect(foundItems).To(ContainElements("3", "4", "6"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue},
									"4": {Exists: &existsTrue},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})

			t.Run("Context when Exists == False", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsFalse},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(2))
					g.Expect(foundItems).To(ContainElements("2", "5"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse},
								},
								Limits: &query.KeyLimits{
									NumberOfKeys: &one,
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("2"))
					})
				})

				t.Run("Context when there are multiple false statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse},
									"2": {Exists: &existsFalse},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(3))
						g.Expect(foundItems).To(ContainElements("1", "2", "5"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := NewThreadSafe()

						// create a number of entries in the associated tree
						keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
						keyValues2 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
						keyValues3 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.Int16(1), "3": datatypes.Float64(3.4)}
						g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})
		})

		t.Run("Context when ExistsType is set", func(t *testing.T) {
			setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
				associatedTree := NewThreadSafe()

				// create a number of entries in the associated tree
				keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
				keyValues2 := datatypes.StringMap{"2": datatypes.String("2")}
				keyValues3 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
				keyValues4 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)}
				keyValues5 := datatypes.StringMap{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)}
				keyValues6 := datatypes.StringMap{"1": datatypes.Int8(4), "2": datatypes.Float32(3.4)}
				g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())

				return associatedTree
			}

			t.Run("Context when Exists == True", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree where the types match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth where the types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_string},
								},
								Limits: &query.KeyLimits{
									NumberOfKeys: &one,
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("2"))
					})
				})

				t.Run("Context when there are multiple TRUE statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree where the types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_float32},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(2))
						g.Expect(foundItems).To(ContainElements("3", "6"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_float64},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})

			t.Run("Context when Exists == False", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree where the types match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(4))
					g.Expect(foundItems).To(ContainElements("1", "2", "4", "5"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
								},
								Limits: &query.KeyLimits{
									NumberOfKeys: &one,
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(2))
						g.Expect(foundItems).To(ContainElements("1", "2"))
					})
				})

				t.Run("Context when there are multiple false statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree where types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsFalse, ExistsType: &datatypes.T_float32},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(4))
						g.Expect(foundItems).To(ContainElements("1", "2", "4", "5"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := NewThreadSafe()

						// create a number of entries in the associated tree
						keyValues1 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.String("1")}
						keyValues2 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.String("2"), "3": datatypes.Float32(3.2)}
						keyValues3 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.String("3")}
						g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsFalse, ExistsType: &datatypes.T_string},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})
		})

		t.Run("Context with a combination of Exists == True and Exists == False", func(t *testing.T) {
			t.Run("It returns the proper values saved in the associated tree", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItems []any
				onFindSelection := func(items []any) {
					foundItems = items
				}

				query := query.Select{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int8},
							"2": {Exists: &existsTrue},
							"3": {Exists: &existsFalse},
						},
					},
				}

				err := associatedTree.Query(query, onFindSelection)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(2))
				g.Expect(foundItems).To(ContainElements("3", "6"))
			})
		})
	})

	t.Run("Desribe Select WHERE.Value query", func(t *testing.T) {
		t.Run("Context when the MatchType is not set or false", func(t *testing.T) {
			t.Run("Context when ValueCompare is =", func(t *testing.T) {
				t.Run("It matches only the itmes that contain the exact key values", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.Equals()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})

				t.Run("Context whene there are multiple key values", func(t *testing.T) {
					t.Run("It matches only the itmes that contain all the exact key values", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						stringOne := datatypes.String("1")
						stringTwo := datatypes.String("2")
						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Value: &stringOne, ValueComparison: query.Equals()},
									"2": {Value: &stringTwo, ValueComparison: query.Equals()},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("4"))
					})
				})
			})

			t.Run("Context when ValueCompare is !=", func(t *testing.T) {
				t.Run("It matches any itmes that do not contain the key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.NotEquals()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(5))
					g.Expect(foundItems).To(ContainElements("2", "3", "4", "5", "6"))
				})

				t.Run("Context whene there are multiple key values", func(t *testing.T) {
					t.Run("It matches all itmes that contain none of the key values", func(t *testing.T) {
						associatedTree := NewThreadSafe()

						// create a number of entries in the associated tree
						keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
						keyValues2 := datatypes.StringMap{"2": datatypes.String("2")}
						keyValues3 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Float32(3.4)}
						keyValues4 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2")}
						keyValues5 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)}
						g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
						g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())

						var foundItems []any
						onFindSelection := func(items []any) {
							foundItems = items
						}

						intOne := datatypes.Int(1)
						stringTwo := datatypes.String("2")
						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Value: &intOne, ValueComparison: query.NotEquals()},
									"2": {Value: &stringTwo, ValueComparison: query.NotEquals()},
								},
							},
						}

						err := associatedTree.Query(query, onFindSelection)
						g.Expect(err).ToNot(HaveOccurred())
						fmt.Println("foundItem:", foundItems)
						g.Expect(len(foundItems)).To(Equal(3))
						g.Expect(foundItems).To(ContainElements("1", "2", "3"))
					})
				})
			})

			t.Run("Context when ValueCompare is <", func(t *testing.T) {
				t.Run("It matches only the itmes that are less than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					intTwo := datatypes.Int(2)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intTwo, ValueComparison: query.LessThan()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "3", "6"))
				})
			})

			t.Run("Context when ValueCompare is <=", func(t *testing.T) {
				t.Run("It matches only the itmes that are less than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.LessThanOrEqual()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "3", "6"))
				})
			})

			t.Run("Context when ValueCompare is >", func(t *testing.T) {
				t.Run("It matches only the itmes that are greater than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThan()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(2))
					g.Expect(foundItems).To(ContainElements("1", "4"))
				})
			})

			t.Run("Context when ValueCompare is >=", func(t *testing.T) {
				t.Run("It matches only the itmes that are greater than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindSelection := func(items []any) {
						foundItems = items
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThanOrEqual()},
							},
						},
					}

					err := associatedTree.Query(query, onFindSelection)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "4", "6"))
				})
			})

			t.Run("It matches all provided key value comparisons when there are multiple", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItems []any
				onFindSelection := func(items []any) {
					foundItems = items
				}

				int8One := datatypes.Int8(1)
				stringFloat := datatypes.String("3.3")
				query := query.Select{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"3": {Value: &int8One, ValueComparison: query.GreaterThanOrEqual()},
							"4": {Value: &stringFloat, ValueComparison: query.LessThan()},
						},
					},
				}

				err := associatedTree.Query(query, onFindSelection)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(1))
				g.Expect(foundItems).To(ContainElements("5"))
			})
		})
	})
}
