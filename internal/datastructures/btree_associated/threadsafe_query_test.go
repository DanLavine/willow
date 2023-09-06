package btreeassociated

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
	onFindPagination := func(items any) bool { return true }

	t.Run("it returns an error if the select query is invalid", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Query(invalidSelection, onFindPagination)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("requires KeyValues or Limits parameters"))
	})

	t.Run("it returns an error with nil onFindPagination", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Query(validSelection, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFindPagination cannot be nil"))
	})
}

func TestAssociatedTree_Query_Basic(t *testing.T) {
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
			onFindPagination := func(item any) bool {
				foundItems = append(foundItems, item)
				return true
			}

			query := query.Select{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))
		})

		t.Run("It can break the query quickly if the pagination callback returns false", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItems []any
			onFindPagination := func(item any) bool {
				foundItems = append(foundItems, item)
				return false
			}

			query := query.Select{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(foundItems)).To(Equal(1))
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
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(4))
					g.Expect(foundItems).To(ContainElements("1", "3", "4", "6"))
				})

				t.Run("It can break the finds early when pagination returns false", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return false
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
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
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("1"))
					})
				})

				t.Run("Context when there are multiple TRUE statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue},
									"2": {Exists: &existsTrue},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(3))
						g.Expect(foundItems).To(ContainElements("3", "4", "6"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue},
									"4": {Exists: &existsTrue},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})

			t.Run("Context when Exists == False", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsFalse},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(2))
					g.Expect(foundItems).To(ContainElements("2", "5"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
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
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("2"))
					})
				})

				t.Run("Context when there are multiple false statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse},
									"2": {Exists: &existsFalse},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
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
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
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
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth where the types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
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
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("2"))
					})
				})

				t.Run("Context when there are multiple TRUE statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree where the types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_float32},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(2))
						g.Expect(foundItems).To(ContainElements("3", "6"))
					})

					t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsTrue, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_float64},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})
				})
			})

			t.Run("Context when Exists == False", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree where the types match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(4))
					g.Expect(foundItems).To(ContainElements("1", "2", "4", "5"))
				})

				t.Run("Context when also setting LIMITS", func(t *testing.T) {
					t.Run("It respects the max key legth", func(t *testing.T) {
						associatedTree := setupQuery(g)

						one := 1
						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
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
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(2))
						g.Expect(foundItems).To(ContainElements("1", "2"))
					})
				})

				t.Run("Context when there are multiple false statements", func(t *testing.T) {
					t.Run("It joins them with AND returns any values in the associated tree where types match", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsFalse, ExistsType: &datatypes.T_float32},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
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
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
									"2": {Exists: &existsFalse, ExistsType: &datatypes.T_string},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
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
				onFindPagination := func(item any) bool {
					foundItems = append(foundItems, item)
					return true
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
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(2))
				g.Expect(foundItems).To(ContainElements("3", "6"))
			})
		})
	})

	t.Run("Desribe Select WHERE.Value query", func(t *testing.T) {
		t.Run("Context when the MatchType is not set", func(t *testing.T) {
			t.Run("Context when ValueCompare is =", func(t *testing.T) {
				t.Run("It matches only the items that contain the exact key values", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.EqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})

				t.Run("Context whene there are multiple key values", func(t *testing.T) {
					t.Run("It matches only the items that contain all the exact key values", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						stringOne := datatypes.String("1")
						stringTwo := datatypes.String("2")
						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Value: &stringOne, ValueComparison: query.EqualsPtr()},
									"2": {Value: &stringTwo, ValueComparison: query.EqualsPtr()},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(1))
						g.Expect(foundItems).To(ContainElements("4"))
					})
				})
			})

			t.Run("Context when ValueCompare is !=", func(t *testing.T) {
				t.Run("It matches any items that do not contain the key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.NotEqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(5))
					g.Expect(foundItems).To(ContainElements("2", "3", "4", "5", "6"))
				})

				t.Run("Context whene there are multiple key values", func(t *testing.T) {
					t.Run("It matches all items that contain none of the key values", func(t *testing.T) {
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
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item)
							return true
						}

						intOne := datatypes.Int(1)
						stringTwo := datatypes.String("2")
						query := query.Select{
							Where: &query.Query{
								KeyValues: map[string]query.Value{
									"1": {Value: &intOne, ValueComparison: query.NotEqualsPtr()},
									"2": {Value: &stringTwo, ValueComparison: query.NotEqualsPtr()},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(3))
						g.Expect(foundItems).To(ContainElements("1", "2", "3"))
					})
				})
			})

			t.Run("Context when ValueCompare is <", func(t *testing.T) {
				t.Run("It matches only the items that are less than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intTwo := datatypes.Int(2)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intTwo, ValueComparison: query.LessThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "3", "6"))
				})
			})

			t.Run("Context when ValueCompare is <=", func(t *testing.T) {
				t.Run("It matches only the items that are less than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.LessThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "3", "6"))
				})
			})

			t.Run("Context when ValueCompare is >", func(t *testing.T) {
				t.Run("It matches only the items that are greater than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(2))
					g.Expect(foundItems).To(ContainElements("1", "4"))
				})
			})

			t.Run("Context when ValueCompare is >=", func(t *testing.T) {
				t.Run("It matches only the items that are greater than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(3))
					g.Expect(foundItems).To(ContainElements("1", "4", "6"))
				})
			})

			t.Run("It matches all provided key value comparisons when there are multiple", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItems []any
				onFindPagination := func(item any) bool {
					foundItems = append(foundItems, item)
					return true
				}

				int8One := datatypes.Int8(1)
				stringFloat := datatypes.String("3.3")
				query := query.Select{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"3": {Value: &int8One, ValueComparison: query.GreaterThanOrEqualPtr()},
							"4": {Value: &stringFloat, ValueComparison: query.LessThanPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(1))
				g.Expect(foundItems).To(ContainElements("5"))
			})
		})

		t.Run("Context when a MatchType is set", func(t *testing.T) {
			typeMatchTrue := true

			t.Run("Context when ValueCompare is !=", func(t *testing.T) {
				t.Run("It matches any items that do not contain the key value for the selected type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.NotEqualsPtr(), ValueTypeMatch: &typeMatchTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("3"))
				})
			})

			t.Run("Context when ValueCompare is <", func(t *testing.T) {
				t.Run("It matches only the items that are less than the provided key value and share a type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intTwo := datatypes.Int(2)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intTwo, ValueComparison: query.LessThanPtr(), ValueTypeMatch: &typeMatchTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})
			})

			t.Run("Context when ValueCompare is <=", func(t *testing.T) {
				t.Run("It matches only the items that are less than or equal to the provided key value when the types match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					intOne := datatypes.Int(1)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.LessThanOrEqualPtr(), ValueTypeMatch: &typeMatchTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("1"))
				})
			})

			t.Run("Context when ValueCompare is >", func(t *testing.T) {
				t.Run("It matches only the items that are greater than the provided key value for the matched type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					int8Three := datatypes.Int8(3)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Three, ValueComparison: query.GreaterThanPtr(), ValueTypeMatch: &typeMatchTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("6"))
				})
			})

			t.Run("Context when ValueCompare is >=", func(t *testing.T) {
				t.Run("It matches only the items that are greater than or equal to the provided key value, when keys match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item)
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.Select{
						Where: &query.Query{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThanOrEqualPtr(), ValueTypeMatch: &typeMatchTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("6"))
				})
			})

			t.Run("It matches all provided key value comparisons when there are multiple", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItems []any
				onFindPagination := func(item any) bool {
					foundItems = append(foundItems, item)
					return true
				}

				int8Two := datatypes.Int8(2)
				float32_5_4 := datatypes.Float32(5.4)
				query := query.Select{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"1": {Value: &int8Two, ValueComparison: query.GreaterThanOrEqualPtr(), ValueTypeMatch: &typeMatchTrue},
							"2": {Value: &float32_5_4, ValueComparison: query.LessThanPtr(), ValueTypeMatch: &typeMatchTrue},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(1))
				g.Expect(foundItems).To(ContainElements("6"))
			})
		})
	})
}

func TestAssociatedTree_Query_JoinAND(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"2": datatypes.Int(2)}
		keyValues3 := datatypes.StringMap{"3": datatypes.Int(3)}
		keyValues4 := datatypes.StringMap{"1": datatypes.String("1")}
		keyValues5 := datatypes.StringMap{"2": datatypes.String("2")}
		keyValues6 := datatypes.StringMap{"3": datatypes.String("3")}
		keyValues7 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("It runs the callback when values are intersected together at the same level", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			And: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(2))
		g.Expect(foundItems).To(ContainElements("7", "13"))
	})

	t.Run("It runs the callback when values are intersected together with the Where clause", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}},
			And: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(2))
		g.Expect(foundItems).To(ContainElements("7", "13"))
	})

	t.Run("It runs the callback when values are intersected together at multiple levels", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			And: []query.Select{
				{And: []query.Select{{And: []query.Select{{Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}}}}}},
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(2))
		g.Expect(foundItems).To(ContainElements("7", "13"))
	})
}

func TestAssociatedTree_Query_JoinOR(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"2": datatypes.Int(2)}
		keyValues3 := datatypes.StringMap{"3": datatypes.Int(3)}
		keyValues4 := datatypes.StringMap{"1": datatypes.String("1")}
		keyValues5 := datatypes.StringMap{"2": datatypes.String("2")}
		keyValues6 := datatypes.StringMap{"3": datatypes.String("3")}
		keyValues7 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("It runs the callback when values are unioned together at the same level", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			Or: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		fmt.Println("found items:", foundItems)
		g.Expect(len(foundItems)).To(Equal(10))
		g.Expect(foundItems).To(ContainElements("1", "2", "7", "8", "9", "10", "11", "12", "13", "14"))
	})

	t.Run("It runs the callback when values are unioned together with the Where clause", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}},
			Or: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(10))
		g.Expect(foundItems).To(ContainElements("1", "2", "7", "8", "9", "10", "11", "12", "13", "14"))
	})

	t.Run("It runs the callback when values are intersected together at multiple levels", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.Select{
			Or: []query.Select{
				{And: []query.Select{{And: []query.Select{{Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}}}}}},
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(10))
		g.Expect(foundItems).To(ContainElements("1", "2", "7", "8", "9", "10", "11", "12", "13", "14"))
	})
}

func TestAssociatedTree_Query_AND_OR_joins(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"2": datatypes.Int(2)}
		keyValues3 := datatypes.StringMap{"3": datatypes.Int(3)}
		keyValues4 := datatypes.StringMap{"1": datatypes.String("1")}
		keyValues5 := datatypes.StringMap{"2": datatypes.String("2")}
		keyValues6 := datatypes.StringMap{"3": datatypes.String("3")}
		keyValues7 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.StringMap{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("It joins all values together accordingly", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item)
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		intThree := datatypes.Int(3)
		query := query.Select{
			And: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{Where: &query.Query{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
			Or: []query.Select{
				{Where: &query.Query{KeyValues: map[string]query.Value{"3": {Value: &intThree, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(5))
		g.Expect(foundItems).To(ContainElements("3", "7", "9", "11", "13"))
	})
}
