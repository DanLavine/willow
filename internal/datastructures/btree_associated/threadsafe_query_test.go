package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Query_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	validSelection := query.AssociatedKeyValuesQuery{}
	invalidSelection := query.AssociatedKeyValuesQuery{KeyValueSelection: &query.KeyValueSelection{}}
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

	ids := make([]string, 6)
	existsTrue := true
	existsFalse := false
	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.KeyValues{"1": datatypes.Int(1)}
		keyValues2 := datatypes.KeyValues{"2": datatypes.String("2")}
		keyValues3 := datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
		keyValues4 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)}
		keyValues5 := datatypes.KeyValues{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)}
		keyValues6 := datatypes.KeyValues{"1": datatypes.Int8(4), "2": datatypes.Float32(3.4)}
		id, err := associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		ids[0] = id
		ids[1], err = associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		ids[2], err = associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		ids[3], err = associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		ids[4], err = associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		ids[5], err = associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("Describe _associated_id special key query", func(t *testing.T) {
		t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItem *AssociatedKeyValues
			onFindPagination := func(item any) bool {
				foundItem = item.(*AssociatedKeyValues)
				return true
			}

			idString := datatypes.String(ids[0])
			query := query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItem).ToNot(BeNil())
			g.Expect(foundItem.KeyValues()["1"]).To(Equal(datatypes.Int(1)))
			g.Expect(foundItem.KeyValues()["_associated_id"]).To(Equal(idString))
		})

		t.Run("It adds the id if the Limits are below the selected values", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItem *AssociatedKeyValues
			onFindPagination := func(item any) bool {
				foundItem = item.(*AssociatedKeyValues)
				return true
			}

			keyLimits := 32
			idString := datatypes.String(ids[3])
			query := query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
					},
					Limits: &query.KeyLimits{
						NumberOfKeys: &keyLimits,
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItem).ToNot(BeNil())
			g.Expect(foundItem.KeyValues()["_associated_id"]).To(Equal(idString))
		})

		t.Run("It does not add the id if the Limits have to many values", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItem *AssociatedKeyValues
			onFindPagination := func(item any) bool {
				foundItem = item.(*AssociatedKeyValues)
				return true
			}

			keyLimits := 1
			idString := datatypes.String(ids[3])
			query := query.AssociatedKeyValuesQuery{
				KeyValueSelection: &query.KeyValueSelection{
					KeyValues: map[string]query.Value{
						"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
					},
					Limits: &query.KeyLimits{
						NumberOfKeys: &keyLimits,
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItem).To(BeNil())
		})

		t.Run("Context when there are additional keys to check", func(t *testing.T) {
			t.Run("It an add the _associated_id first", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItem *AssociatedKeyValues
				onFindPagination := func(item any) bool {
					foundItem = item.(*AssociatedKeyValues)
					return true
				}

				idString := datatypes.String(ids[3])
				oneString := datatypes.String("1")
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
						KeyValues: map[string]query.Value{
							"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
							"1":              query.Value{Value: &oneString, ValueComparison: query.EqualsPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues()["1"]).To(Equal(datatypes.String("1")))
				g.Expect(foundItem.KeyValues()["2"]).To(Equal(datatypes.String("2")))
				g.Expect(foundItem.KeyValues()["3"]).To(Equal(datatypes.Float64(3.4)))
				g.Expect(foundItem.KeyValues()["_associated_id"]).To(Equal(idString))
			})

			t.Run("It an add the _associated_id second", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := datatypes.KeyValues{"a": int8_4}
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(item any) bool {
					foundItem = item.(*AssociatedKeyValues)
					return true
				}

				idString := datatypes.String(newID)
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
						KeyValues: map[string]query.Value{
							"a":              query.Value{Value: &int8_4, ValueComparison: query.EqualsPtr()},
							"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err = associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues()["a"]).To(Equal(datatypes.Int8(4)))
				g.Expect(foundItem.KeyValues()["_associated_id"]).To(Equal(idString))
			})

			t.Run("It an add the _associated_id second if it is under the Limits", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := datatypes.KeyValues{"a": int8_4}
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(item any) bool {
					foundItem = item.(*AssociatedKeyValues)
					return true
				}

				keyLimits := 32
				idString := datatypes.String(newID)
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
						KeyValues: map[string]query.Value{
							"a":              query.Value{Value: &int8_4, ValueComparison: query.EqualsPtr()},
							"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
						},
						Limits: &query.KeyLimits{
							NumberOfKeys: &keyLimits,
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err = associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues()["a"]).To(Equal(datatypes.Int8(4)))
				g.Expect(foundItem.KeyValues()["_associated_id"]).To(Equal(idString))
			})

			t.Run("It does not add the _associated_id second if it is over the Limits", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := datatypes.KeyValues{"a": int8_4, "b": int8_4}
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(item any) bool {
					foundItem = item.(*AssociatedKeyValues)
					return true
				}

				keyLimits := 1
				idString := datatypes.String(newID)
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
						KeyValues: map[string]query.Value{
							"a":              query.Value{Value: &int8_4, ValueComparison: query.EqualsPtr()},
							"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
						},
						Limits: &query.KeyLimits{
							NumberOfKeys: &keyLimits,
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err = associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).To(BeNil())
			})

			t.Run("It filters out the item if the key values pairs do not match", func(t *testing.T) {
				associatedTree := setupQuery(g)

				var foundItem *AssociatedKeyValues
				onFindPagination := func(item any) bool {
					foundItem = item.(*AssociatedKeyValues)
					return true
				}

				int8_4 := datatypes.Int8(4)
				idString := datatypes.String(ids[2])
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
						KeyValues: map[string]query.Value{
							"3":              query.Value{Value: &int8_4, ValueComparison: query.EqualsPtr()},
							"_associated_id": query.Value{Value: &idString, ValueComparison: query.EqualsPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).To(BeNil())
			})
		})
	})

	t.Run("Describe Select ALL query", func(t *testing.T) {
		t.Run("It returns the value saved in the associated tree when found", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItems []any
			onFindPagination := func(item any) bool {
				foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
				return true
			}

			query := query.AssociatedKeyValuesQuery{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))
		})

		t.Run("It can break the query quickly if the pagination callback returns false", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItems []any
			onFindPagination := func(item any) bool {
				foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
				return false
			}

			query := query.AssociatedKeyValuesQuery{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(foundItems)).To(Equal(1))
		})
	})

	t.Run("Describe Select WHERE.Exists query", func(t *testing.T) {
		t.Run("Context when ExistsType is nil", func(t *testing.T) {
			t.Run("Context when Exists == True", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return false
					}

					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
						keyValues1 := datatypes.KeyValues{"1": datatypes.Int(1)}
						keyValues2 := datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
						keyValues3 := datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.Int16(1), "3": datatypes.Float64(3.4)}
						associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues3, func() any { return "2" }, noOpOnFind)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
			t.Run("Context when Exists == True", func(t *testing.T) {
				t.Run("It returns any value saved in the associated tree where the types match", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
						keyValues1 := datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("1")}
						keyValues2 := datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("2"), "3": datatypes.Float32(3.2)}
						keyValues3 := datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("3")}
						associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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
					foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
					return true
				}

				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
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
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						stringOne := datatypes.String("1")
						stringTwo := datatypes.String("2")
						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
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

					t.Run("It filters out values when desired keys that not found in the first part of a query", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						intOne := datatypes.Int(1)
						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"0": {Value: &intOne, ValueComparison: query.EqualsPtr()}, //runs first, and should invalidate a query
									"1": {Value: &intOne, ValueComparison: query.EqualsPtr()},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(0))
					})

					t.Run("It filters out values when desired keys that not found after the first lookup in a query", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						intOne := datatypes.Int(1)
						query := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"1":  {Value: &intOne, ValueComparison: query.EqualsPtr()},
									"32": {Value: &intOne, ValueComparison: query.EqualsPtr()}, //runs second, and should invalidate a query
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

			t.Run("Context when ValueCompare is !=", func(t *testing.T) {
				t.Run("It removes an items that match all undesiredable key values", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // filter out first key
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(5))
					g.Expect(foundItems).To(ContainElements("2", "3", "4", "5", "6"))
				})

				t.Run("It matches any items that do not contain the key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"32": {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // should be a no op
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					err := associatedTree.Query(query, onFindPagination)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(len(foundItems)).To(Equal(6))
					g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))
				})

				t.Run("Context whene there are multiple key values", func(t *testing.T) {
					t.Run("It only removes tree items that match all exact not operations together", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						// no-op query
						intOne := datatypes.Int(1)
						query1 := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"1":  {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // filter out first key
									"32": {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // should be a no op
								},
							},
						}
						g.Expect(query1.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query1, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(6))
						g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))

						// removal query
						strOne := datatypes.String("1")
						strTwo := datatypes.String("2")
						query2 := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"1": {Value: &strOne, ValueComparison: query.NotEqualsPtr()},
									"2": {Value: &strTwo, ValueComparison: query.NotEqualsPtr()},
								},
							},
						}
						g.Expect(query2.Validate()).ToNot(HaveOccurred())

						foundItems = []any{} // clear the found items
						err = associatedTree.Query(query2, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(5))
						g.Expect(foundItems).To(ContainElements("1", "2", "3", "5", "6"))
					})

					t.Run("It can run removing tree items that match all exact not operations together in any search order", func(t *testing.T) {
						associatedTree := setupQuery(g)

						var foundItems []any
						onFindPagination := func(item any) bool {
							foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
							return true
						}

						// no-op query
						intOne := datatypes.Int(1)
						query1 := query.AssociatedKeyValuesQuery{
							KeyValueSelection: &query.KeyValueSelection{
								KeyValues: map[string]query.Value{
									"-32": {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // should be a no op
									"1":   {Value: &intOne, ValueComparison: query.NotEqualsPtr()}, // filter out first key
								},
							},
						}
						g.Expect(query1.Validate()).ToNot(HaveOccurred())

						err := associatedTree.Query(query1, onFindPagination)
						g.Expect(err).ToNot(HaveOccurred())
						g.Expect(len(foundItems)).To(Equal(6))
					})
				})
			})

			t.Run("Context when ValueCompare is <", func(t *testing.T) {
				t.Run("It matches only the items that are less than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intTwo := datatypes.Int(2)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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

			t.Run("Context when ValueCompare is < MATCH for type enforcement", func(t *testing.T) {
				t.Run("It matches only the items that are less than the provided key and value's type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intTwo := datatypes.Int(2)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"1": {Value: &intTwo, ValueComparison: query.LessThanMatchTypePtr()},
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
				t.Run("It matches only the items that are less than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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

			t.Run("Context when ValueCompare is <= MATCH for type enforcement", func(t *testing.T) {
				t.Run("It matches only the items that are less than or equal to the provided key and value's type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"1": {Value: &intOne, ValueComparison: query.LessThanOrEqualMatchTypePtr()},
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
				t.Run("It matches only the items that are greater than the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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

			t.Run("Context when ValueCompare is > MATCH for type enforcement", func(t *testing.T) {
				t.Run("It matches only the items that are greater than the provided key and value's type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					int8Three := datatypes.Int8(3)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Three, ValueComparison: query.GreaterThanMatchTypePtr()},
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
				t.Run("It matches only the items that are greater than or equal to the provided key value", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
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

			t.Run("Context when ValueCompare is >= MATCH for type enforcement", func(t *testing.T) {
				t.Run("It matches only the items that are greater than or equal to the provided key and value's type", func(t *testing.T) {
					associatedTree := setupQuery(g)

					var foundItems []any
					onFindPagination := func(item any) bool {
						foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := query.AssociatedKeyValuesQuery{
						KeyValueSelection: &query.KeyValueSelection{
							KeyValues: map[string]query.Value{
								"1": {Value: &int8Four, ValueComparison: query.GreaterThanOrEqualMatchTypePtr()},
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
					foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
					return true
				}

				int8One := datatypes.Int8(1)
				stringFloat := datatypes.String("3.3")
				query := query.AssociatedKeyValuesQuery{
					KeyValueSelection: &query.KeyValueSelection{
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
	})
}

func TestAssociatedTree_Query_JoinAND(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.KeyValues{"1": datatypes.Int(1)}
		keyValues2 := datatypes.KeyValues{"2": datatypes.Int(2)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.Int(3)}
		keyValues4 := datatypes.KeyValues{"1": datatypes.String("1")}
		keyValues5 := datatypes.KeyValues{"2": datatypes.String("2")}
		keyValues6 := datatypes.KeyValues{"3": datatypes.String("3")}
		keyValues7 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)

		return associatedTree
	}

	t.Run("It runs the callback when values are intersected together at the same level", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			And: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
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
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}},
			And: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
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
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			And: []query.AssociatedKeyValuesQuery{
				{
					And: []query.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(2))
		g.Expect(foundItems).To(ContainElements("7", "13"))
	})

	t.Run("It stops the query if no ids are found", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		int73 := datatypes.Int(73)
		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			And: []query.AssociatedKeyValuesQuery{
				{
					And: []query.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &int73, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(0))
	})
}

func TestAssociatedTree_Query_JoinOR(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	setupQuery := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.KeyValues{"1": datatypes.Int(1)}
		keyValues2 := datatypes.KeyValues{"2": datatypes.Int(2)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.Int(3)}
		keyValues4 := datatypes.KeyValues{"1": datatypes.String("1")}
		keyValues5 := datatypes.KeyValues{"2": datatypes.String("2")}
		keyValues6 := datatypes.KeyValues{"3": datatypes.String("3")}
		keyValues7 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)

		return associatedTree
	}

	t.Run("It runs the callback when values are unioned together at the same level", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			Or: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(10))
		g.Expect(foundItems).To(ContainElements("1", "2", "7", "8", "9", "10", "11", "12", "13", "14"))
	})

	t.Run("It runs the callback when values are unioned together with the Where clause", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}},
			Or: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
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
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := query.AssociatedKeyValuesQuery{
			Or: []query.AssociatedKeyValuesQuery{
				{And: []query.AssociatedKeyValuesQuery{{And: []query.AssociatedKeyValuesQuery{{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}}}}}},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
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
		keyValues1 := datatypes.KeyValues{"1": datatypes.Int(1)}
		keyValues2 := datatypes.KeyValues{"2": datatypes.Int(2)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.Int(3)}
		keyValues4 := datatypes.KeyValues{"1": datatypes.String("1")}
		keyValues5 := datatypes.KeyValues{"2": datatypes.String("2")}
		keyValues6 := datatypes.KeyValues{"3": datatypes.String("3")}
		keyValues7 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)}
		keyValues8 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}
		keyValues9 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)}
		keyValues10 := datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")}
		keyValues11 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues12 := datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")}
		keyValues13 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)}
		keyValues14 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")}

		associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues6, func() any { return "6" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues7, func() any { return "7" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues8, func() any { return "8" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues9, func() any { return "9" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues10, func() any { return "10" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues11, func() any { return "11" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues12, func() any { return "12" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues13, func() any { return "13" }, noOpOnFind)
		associatedTree.CreateOrFind(keyValues14, func() any { return "14" }, noOpOnFind)

		return associatedTree
	}

	t.Run("It joins all values together accordingly", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		intThree := datatypes.Int(3)
		query := query.AssociatedKeyValuesQuery{
			And: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}}}},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
			Or: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"3": {Value: &intThree, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		//g.Expect(len(foundItems)).To(Equal(5))
		g.Expect(foundItems).To(ContainElements("3", "7", "9", "11", "13"))
	})

	t.Run("It still querys the ORs properly if AND finds nothing", func(t *testing.T) {
		associatedTree := setupQuery(g)

		var foundItems []any
		onFindPagination := func(item any) bool {
			foundItems = append(foundItems, item.(*AssociatedKeyValues).Value())
			return true
		}

		int73 := datatypes.Int(73)
		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		intThree := datatypes.Int(3)
		query := query.AssociatedKeyValuesQuery{
			And: []query.AssociatedKeyValuesQuery{
				{
					And: []query.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &int73, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []query.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &query.KeyValueSelection{
										KeyValues: map[string]query.Value{"1": {Value: &intOne, ValueComparison: query.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"2": {Value: &intTwo, ValueComparison: query.EqualsPtr()}}}},
			},
			Or: []query.AssociatedKeyValuesQuery{
				{KeyValueSelection: &query.KeyValueSelection{KeyValues: map[string]query.Value{"3": {Value: &intThree, ValueComparison: query.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(4))
		g.Expect(foundItems).To(ContainElements("3", "9", "11", "13"))
	})
}
