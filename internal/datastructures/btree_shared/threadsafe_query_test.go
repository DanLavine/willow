package btreeshared

import (
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

	setup := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// create a number of entries in the associated tree
		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"2": datatypes.String("2")}
		keyValues3 := datatypes.StringMap{"1": datatypes.Int8(1), "2": datatypes.Float32(3.4)}
		keyValues4 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.Int16(1), "3": datatypes.Float64(3.4)}
		keyValues5 := datatypes.StringMap{"3": datatypes.Int32(1), "4": datatypes.Float64(3.4)}
		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues5, func() any { return "5" }, noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("Describe Select ALL query", func(t *testing.T) {
		t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
			associatedTree := setup(g)

			var foundItems []any
			onFindSelection := func(items []any) {
				foundItems = items
			}

			err := associatedTree.Query(query.Select{}, onFindSelection)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(foundItems).To(ContainElements("1", "2", "3", "4", "5"))
		})
	})

	t.Run("Describe Select WHERE query", func(t *testing.T) {
		t.Run("Context when Exists == True", func(t *testing.T) {
			t.Run("It returns any value saved in the associated tree", func(t *testing.T) {
				associatedTree := setup(g)

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
				g.Expect(len(foundItems)).To(Equal(3))
				g.Expect(foundItems).To(ContainElements("1", "3", "4"))
			})

			t.Run("Context when also setting LIMITS", func(t *testing.T) {
				t.Run("It respects the max key legth", func(t *testing.T) {
					associatedTree := setup(g)

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
					associatedTree := setup(g)

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
					g.Expect(len(foundItems)).To(Equal(2))
					g.Expect(foundItems).To(ContainElements("3", "4"))
				})

				t.Run("It doesn't run the callback if no values are found", func(t *testing.T) {
					associatedTree := setup(g)

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
				associatedTree := setup(g)

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
					associatedTree := setup(g)

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
					associatedTree := setup(g)

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
					g.Expect(len(foundItems)).To(Equal(1))
					g.Expect(foundItems).To(ContainElements("5"))
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

		t.Run("Context with a combination of Exists == True and Exists == False", func(t *testing.T) {
			t.Run("It returns the proper values saved in the associated tree", func(t *testing.T) {
				associatedTree := setup(g)

				var foundItems []any
				onFindSelection := func(items []any) {
					foundItems = items
				}

				query := query.Select{
					Where: &query.Query{
						KeyValues: map[string]query.Value{
							"1": {Exists: &existsTrue},
							"2": {Exists: &existsTrue},
							"3": {Exists: &existsFalse},
						},
					},
				}

				err := associatedTree.Query(query, onFindSelection)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(len(foundItems)).To(Equal(1))
				g.Expect(foundItems).To(ContainElements("3"))
			})
		})
	})

	//	t.Run("Context when Value is properly set", func(t *testing.T) {
	//		t.Run("Context when Value Equals", func(t *testing.T) {
	//			t.Run("It returns any items that contain the exact key value pairs", func(t *testing.T) {
	//				associatedTree := setup(g)
	//
	//				var foundItems []any
	//				onFindSelection := func(items []any) {
	//					foundItems = items
	//				}
	//
	//				intOne := datatypes.Int(1)
	//				query := query.Select{
	//					Where: &query.Query{
	//						KeyValues: map[string]query.Value{
	//							"1": {Value: &intOne, ValueComparison: query.Equals()},
	//						},
	//					},
	//				}
	//
	//				err := associatedTree.Query(query, onFindSelection)
	//				g.Expect(err).ToNot(HaveOccurred())
	//				g.Expect(len(foundItems)).To(Equal(1))
	//				g.Expect(foundItems).To(ContainElements("1"))
	//			})
	//		})
	//	})
}
