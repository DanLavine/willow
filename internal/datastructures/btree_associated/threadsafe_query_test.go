package btreeassociated

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Query_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	validSelection := datatypes.AssociatedKeyValuesQuery{}
	invalidSelection := datatypes.AssociatedKeyValuesQuery{KeyValueSelection: &datatypes.KeyValueSelection{}}
	onFindPagination := func(value *AssociatedKeyValues) bool { return true }

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
		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.String("2")})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.String("2"), "3": datatypes.Float64(3.4)})
		keyValues5 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)})
		keyValues6 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(4), "2": datatypes.Float32(3.4)})
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
			onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
				foundItem = associatedKeyValues
				return true
			}

			idString := datatypes.String(ids[0])
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItem).ToNot(BeNil())
			g.Expect(foundItem.KeyValues().RetrieveStringDataType()["1"]).To(Equal(datatypes.Int(1)))
			g.Expect(foundItem.AssociatedID()).To(Equal(ids[0]))
		})

		t.Run("It adds the id if the Limits are below the selected values", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItem *AssociatedKeyValues
			onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
				foundItem = associatedKeyValues
				return true
			}

			keyLimits := 32
			idString := datatypes.String(ids[3])
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
					},
					Limits: &datatypes.KeyLimits{
						NumberOfKeys: &keyLimits,
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItem).ToNot(BeNil())
			g.Expect(foundItem.AssociatedID()).To(Equal(ids[3]))
		})

		t.Run("It does not add the id if the Limits have to many values", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItem *AssociatedKeyValues
			onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
				foundItem = associatedKeyValues
				return true
			}

			keyLimits := 1
			idString := datatypes.String(ids[3])
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
					},
					Limits: &datatypes.KeyLimits{
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
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItem = associatedKeyValues
					return true
				}

				idString := datatypes.String(ids[3])
				oneString := datatypes.String("1")
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
							"1":              datatypes.Value{Value: &oneString, ValueComparison: datatypes.EqualsPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err := associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues().RetrieveStringDataType()["1"]).To(Equal(datatypes.String("1")))
				g.Expect(foundItem.KeyValues().RetrieveStringDataType()["2"]).To(Equal(datatypes.String("2")))
				g.Expect(foundItem.KeyValues().RetrieveStringDataType()["3"]).To(Equal(datatypes.Float64(3.4)))
				g.Expect(foundItem.AssociatedID()).To(Equal(ids[3]))
			})

			t.Run("It an add the _associated_id second", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := ConverDatatypesKeyValues(datatypes.KeyValues{"a": int8_4})
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItem = associatedKeyValues
					return true
				}

				idString := datatypes.String(newID)
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"a":              datatypes.Value{Value: &int8_4, ValueComparison: datatypes.EqualsPtr()},
							"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err = associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues().RetrieveStringDataType()["a"]).To(Equal(datatypes.Int8(4)))
				g.Expect(foundItem.AssociatedID()).To(Equal(newID))
			})

			t.Run("It an add the _associated_id second if it is under the Limits", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := ConverDatatypesKeyValues(datatypes.KeyValues{"a": int8_4})
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItem = associatedKeyValues
					return true
				}

				keyLimits := 32
				idString := datatypes.String(newID)
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"a":              datatypes.Value{Value: &int8_4, ValueComparison: datatypes.EqualsPtr()},
							"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
						},
						Limits: &datatypes.KeyLimits{
							NumberOfKeys: &keyLimits,
						},
					},
				}
				g.Expect(query.Validate()).ToNot(HaveOccurred())

				err = associatedTree.Query(query, onFindPagination)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(foundItem).ToNot(BeNil())
				g.Expect(foundItem.KeyValues().RetrieveStringDataType()["a"]).To(Equal(datatypes.Int8(4)))
				g.Expect(foundItem.AssociatedID()).To(Equal(newID))
			})

			t.Run("It does not add the _associated_id second if it is over the Limits", func(t *testing.T) {
				associatedTree := setupQuery(g)

				int8_4 := datatypes.Int8(4)
				newKeyValues := ConverDatatypesKeyValues(datatypes.KeyValues{"a": int8_4, "b": int8_4})
				newID, err := associatedTree.CreateOrFind(newKeyValues, func() any { return "1" }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())

				var foundItem *AssociatedKeyValues
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItem = associatedKeyValues
					return true
				}

				keyLimits := 1
				idString := datatypes.String(newID)
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"a":              datatypes.Value{Value: &int8_4, ValueComparison: datatypes.EqualsPtr()},
							"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
						},
						Limits: &datatypes.KeyLimits{
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
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItem = associatedKeyValues
					return true
				}

				int8_4 := datatypes.Int8(4)
				idString := datatypes.String(ids[2])
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"3":              datatypes.Value{Value: &int8_4, ValueComparison: datatypes.EqualsPtr()},
							"_associated_id": datatypes.Value{Value: &idString, ValueComparison: datatypes.EqualsPtr()},
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
			onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
				foundItems = append(foundItems, associatedKeyValues.Value())
				return true
			}

			query := datatypes.AssociatedKeyValuesQuery{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			err := associatedTree.Query(query, onFindPagination)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5", "6"))
		})

		t.Run("It can break the query quickly if the pagination callback returns false", func(t *testing.T) {
			associatedTree := setupQuery(g)

			var foundItems []any
			onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
				foundItems = append(foundItems, associatedKeyValues.Value())
				return false
			}

			query := datatypes.AssociatedKeyValuesQuery{}
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return false
					}

					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1": {Exists: &existsTrue},
								},
								Limits: &datatypes.KeyLimits{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1": {Exists: &existsFalse},
								},
								Limits: &datatypes.KeyLimits{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
						keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
						keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)})
						keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.Int16(1), "3": datatypes.Float64(3.4)})
						associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues3, func() any { return "2" }, noOpOnFind)

						var foundItems []any
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"2": {Exists: &existsTrue, ExistsType: &datatypes.T_string},
								},
								Limits: &datatypes.KeyLimits{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1": {Exists: &existsFalse, ExistsType: &datatypes.T_int8},
								},
								Limits: &datatypes.KeyLimits{
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
						keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("1")})
						keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("2"), "3": datatypes.Float32(3.2)})
						keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int8(1), "2": datatypes.String("3")})
						associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
						associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)

						var foundItems []any
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
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
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItems = append(foundItems, associatedKeyValues.Value())
					return true
				}

				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()},
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						stringOne := datatypes.String("1")
						stringTwo := datatypes.String("2")
						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1": {Value: &stringOne, ValueComparison: datatypes.EqualsPtr()},
									"2": {Value: &stringTwo, ValueComparison: datatypes.EqualsPtr()},
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						intOne := datatypes.Int(1)
						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"0": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}, //runs first, and should invalidate a query
									"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()},
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						intOne := datatypes.Int(1)
						query := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1":  {Value: &intOne, ValueComparison: datatypes.EqualsPtr()},
									"32": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}, //runs second, and should invalidate a query
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // filter out first key
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"32": {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // should be a no op
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						// no-op query
						intOne := datatypes.Int(1)
						query1 := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1":  {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // filter out first key
									"32": {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // should be a no op
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
						query2 := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"1": {Value: &strOne, ValueComparison: datatypes.NotEqualsPtr()},
									"2": {Value: &strTwo, ValueComparison: datatypes.NotEqualsPtr()},
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
						onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
							foundItems = append(foundItems, associatedKeyValues.Value())
							return true
						}

						// no-op query
						intOne := datatypes.Int(1)
						query1 := datatypes.AssociatedKeyValuesQuery{
							KeyValueSelection: &datatypes.KeyValueSelection{
								KeyValues: map[string]datatypes.Value{
									"-32": {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // should be a no op
									"1":   {Value: &intOne, ValueComparison: datatypes.NotEqualsPtr()}, // filter out first key
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intTwo := datatypes.Int(2)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intTwo, ValueComparison: datatypes.LessThanPtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intTwo := datatypes.Int(2)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intTwo, ValueComparison: datatypes.LessThanMatchTypePtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intOne, ValueComparison: datatypes.LessThanOrEqualPtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					intOne := datatypes.Int(1)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &intOne, ValueComparison: datatypes.LessThanOrEqualMatchTypePtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &int8Four, ValueComparison: datatypes.GreaterThanPtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					int8Three := datatypes.Int8(3)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &int8Three, ValueComparison: datatypes.GreaterThanMatchTypePtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &int8Four, ValueComparison: datatypes.GreaterThanOrEqualPtr()},
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
					onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
						foundItems = append(foundItems, associatedKeyValues.Value())
						return true
					}

					int8Four := datatypes.Int8(4)
					query := datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"1": {Value: &int8Four, ValueComparison: datatypes.GreaterThanOrEqualMatchTypePtr()},
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
				onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
					foundItems = append(foundItems, associatedKeyValues.Value())
					return true
				}

				int8One := datatypes.Int8(1)
				stringFloat := datatypes.String("3.3")
				query := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"3": {Value: &int8One, ValueComparison: datatypes.GreaterThanOrEqualPtr()},
							"4": {Value: &stringFloat, ValueComparison: datatypes.LessThanPtr()},
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
		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2)})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.Int(3)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues5 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.String("2")})
		keyValues6 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.String("3")})
		keyValues7 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)})
		keyValues8 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")})
		keyValues9 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)})
		keyValues10 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")})
		keyValues11 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues12 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")})
		keyValues13 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues14 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")})

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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}}},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}},
			And: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{
					And: []datatypes.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		int73 := datatypes.Int(73)
		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{
					And: []datatypes.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &int73, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2)})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.Int(3)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues5 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.String("2")})
		keyValues6 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.String("3")})
		keyValues7 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)})
		keyValues8 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")})
		keyValues9 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)})
		keyValues10 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")})
		keyValues11 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues12 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")})
		keyValues13 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues14 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")})

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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			Or: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}}},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}},
			Or: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		query := datatypes.AssociatedKeyValuesQuery{
			Or: []datatypes.AssociatedKeyValuesQuery{
				{And: []datatypes.AssociatedKeyValuesQuery{{And: []datatypes.AssociatedKeyValuesQuery{{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}}}}}}},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
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
		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2)})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.Int(3)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues5 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.String("2")})
		keyValues6 := ConverDatatypesKeyValues(datatypes.KeyValues{"3": datatypes.String("3")})
		keyValues7 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2)})
		keyValues8 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")})
		keyValues9 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.Int(3)})
		keyValues10 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "3": datatypes.String("3")})
		keyValues11 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues12 := ConverDatatypesKeyValues(datatypes.KeyValues{"2": datatypes.Int(2), "3": datatypes.String("3")})
		keyValues13 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Int(2), "3": datatypes.Int(3)})
		keyValues14 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2"), "3": datatypes.String("3")})

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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		intThree := datatypes.Int(3)
		query := datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}}}},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
			},
			Or: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"3": {Value: &intThree, ValueComparison: datatypes.EqualsPtr()}}}},
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
		onFindPagination := func(associatedKeyValues *AssociatedKeyValues) bool {
			foundItems = append(foundItems, associatedKeyValues.Value())
			return true
		}

		int73 := datatypes.Int(73)
		intOne := datatypes.Int(1)
		intTwo := datatypes.Int(2)
		intThree := datatypes.Int(3)
		query := datatypes.AssociatedKeyValuesQuery{
			And: []datatypes.AssociatedKeyValuesQuery{
				{
					And: []datatypes.AssociatedKeyValuesQuery{
						{ // do the same query 2x. shouldn't matter
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &int73, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
						{
							And: []datatypes.AssociatedKeyValuesQuery{
								{
									KeyValueSelection: &datatypes.KeyValueSelection{
										KeyValues: map[string]datatypes.Value{"1": {Value: &intOne, ValueComparison: datatypes.EqualsPtr()}},
									},
								},
							},
						},
					},
				},
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"2": {Value: &intTwo, ValueComparison: datatypes.EqualsPtr()}}}},
			},
			Or: []datatypes.AssociatedKeyValuesQuery{
				{KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{"3": {Value: &intThree, ValueComparison: datatypes.EqualsPtr()}}}},
			},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		err := associatedTree.Query(query, onFindPagination)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(4))
		g.Expect(foundItems).To(ContainElements("3", "9", "11", "13"))
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
				keys := KeyValues{}

				for i := 0; i <= numberOfKeys; i++ {
					keys[datatypes.String(fmt.Sprintf("%d", tNum))] = datatypes.String(fmt.Sprintf("%d", tNum))
				}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, baicCreate)
				g.Expect(err).ToNot(HaveOccurred())
			}(i%10, i)
		}
		wg.Wait()

		// 2. generate a massive list of key value pairs
		largeKeyValues := datatypes.KeyValues{}
		for i := 0; i < 6; i++ {
			largeKeyValues[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", i))
		}

		fmt.Println(len(largeKeyValues.GenerateTagPairs()))
		g.Fail("boo")

		// 3. generate a massive query to run
		query := datatypes.AssociatedKeyValuesQuery{}
		for _, keyValues := range largeKeyValues.GenerateTagPairs() {
			associatedQuery := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{}},
			}

			for key, value := range keyValues {
				associatedQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &value, ValueComparison: datatypes.EqualsPtr()}
			}

			query.Or = append(query.Or, associatedQuery)
		}

		// 4. run the query
		//counter := 0
		//g.Expect(associatedTree.Query(query, func(item *AssociatedKeyValues) bool {
		//	counter++
		//	return true
		//})).ToNot(HaveOccurred())
		//
		//g.Expect(counter).To(Equal(1000))
	})
}
