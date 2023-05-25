package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/internal/datastructures/btree_associated/testhelpers"
	. "github.com/onsi/gomega"
)

func TestAssociatedTree_CreateOrFind(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := New()

		item, err := associatedTree.CreateOrFind(nil, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("it returns an error with empty keyValues", func(t *testing.T) {
		associatedTree := New()

		item, err := associatedTree.CreateOrFind(datatypes.StringMap{}, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := New()

		item, err := associatedTree.CreateOrFind(datatypes.StringMap{"1": "one"}, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("single item test", func(t *testing.T) {
		keyValues := datatypes.StringMap{
			"1": "one",
		}
		t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
			associatedTree := New()

			item, err := associatedTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			associatedTreeTester := item.(*JoinTreeTester)
			g.Expect(associatedTreeTester.Value).To(Equal("other"))
		})

		t.Run("it calls onFind if the item already exists", func(t *testing.T) {
			associatedTree := New()

			// create item
			item, err := associatedTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			associatedTreeTester := item.(*JoinTreeTester)
			g.Expect(associatedTreeTester.Value).To(Equal("other"))
			g.Expect(associatedTreeTester.OnFindCount).To(Equal(0))

			// find item
			item, err = associatedTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())

			associatedTreeTester = item.(*JoinTreeTester)
			g.Expect(associatedTreeTester.Value).To(Equal("other"))
			g.Expect(associatedTreeTester.OnFindCount).To(Equal(1))
		})
	})

	t.Run("multiple item test", func(t *testing.T) {
		testKeyValues := datatypes.StringMap{
			"1": "one",
			"2": "two",
			"3": "three",
			"4": "four",
		}

		setup := func() *associatedTree {
			associatedTree := New()

			item, err := associatedTree.CreateOrFind(testKeyValues, NewJoinTreeTester("setup data"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			return associatedTree
		}

		t.Run("it can create creates a value with multiple key value pairs", func(t *testing.T) {
			associatedTree := New()

			item, err := associatedTree.CreateOrFind(testKeyValues, NewJoinTreeTester("setup data"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())
		})

		t.Run("it calls onFind if the item already exists", func(t *testing.T) {
			associatedTree := setup()

			item, err := associatedTree.CreateOrFind(testKeyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())

			associatedTreeTester := item.(*JoinTreeTester)
			g.Expect(associatedTreeTester.Value).To(Equal("setup data"))
			g.Expect(associatedTreeTester.OnFindCount).To(Equal(1))
		})

		t.Run("it can create new values if the keyValues are different", func(t *testing.T) {
			associatedTree := setup()
			keyValuePairs1 := datatypes.StringMap{
				"1": "one",
			}
			keyValuePairs2 := datatypes.StringMap{
				"1": "one",
				"2": "two",
			}
			keyValuePairs3 := datatypes.StringMap{
				"1": "one",
				"2": "two",
				"3": "three",
				"5": "five",
			}

			item1, err := associatedTree.CreateOrFind(keyValuePairs1, NewJoinTreeTester("first"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			associatedTreeTester1 := item1.(*JoinTreeTester)
			g.Expect(associatedTreeTester1.Value).To(Equal("first"))

			item2, err := associatedTree.CreateOrFind(keyValuePairs2, NewJoinTreeTester("second"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			associatedTreeTester2 := item2.(*JoinTreeTester)
			g.Expect(associatedTreeTester2.Value).To(Equal("second"))

			item3, err := associatedTree.CreateOrFind(keyValuePairs3, NewJoinTreeTester("third"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			associatedTreeTester3 := item3.(*JoinTreeTester)
			g.Expect(associatedTreeTester3.Value).To(Equal("third"))
		})

		t.Run("it records multiple ids for a key value pair if there are multiple groupings", func(t *testing.T) {
			associatedTree := setup()
			keyValuePairs1 := datatypes.StringMap{
				"a": "a",
				"b": "b",
				"c": "c",
			}
			keyValuePairs2 := datatypes.StringMap{
				"b": "b",
				"c": "c",
				"d": "d",
			}

			item1, err := associatedTree.CreateOrFind(keyValuePairs1, NewJoinTreeTester("first"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			associatedTreeTester1 := item1.(*JoinTreeTester)
			g.Expect(associatedTreeTester1.Value).To(Equal("first"))

			item2, err := associatedTree.CreateOrFind(keyValuePairs2, NewJoinTreeTester("second"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			associatedTreeTester2 := item2.(*JoinTreeTester)
			g.Expect(associatedTreeTester2.Value).To(Equal("second"))

			associatedKeyValues := associatedTree.groupedKeyValueAssociation.Find(datatypes.Int(3), nil).(*keyValues)
			valuesForKey := associatedKeyValues.values.Find(datatypes.String("b"), nil).(*keyValues)
			idHolder := valuesForKey.values.Find(datatypes.String("b"), nil).(*idHolder)

			g.Expect(len(idHolder.IDs)).To(Equal(2))
		})
	})

	t.Run("when creating an item fails", func(t *testing.T) {
		keyValuePairs := datatypes.StringMap{
			"1": "one",
			"2": "two",
			"3": "three",
			"4": "four",
		}

		setup := func() *associatedTree {
			associatedTree := New()

			item, err := associatedTree.CreateOrFind(keyValuePairs, NewJoinTreeTester("setup data"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			return associatedTree
		}

		t.Run("it cleans up extra key value pairs that might have been created", func(t *testing.T) {
			associatedTree := setup()
			newKeyValues := datatypes.StringMap{
				"1": "other",
				"2": "foo",
				"3": "three",
				"5": "this shouldn't be there",
			}

			item, err := associatedTree.CreateOrFind(newKeyValues, NewJoinTreeTesterWithError, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(item).To(BeNil())

			compositeKeyValues := associatedTree.groupedKeyValueAssociation.Find(datatypes.Int(4), nil).(*keyValues)

			// "1" values
			oneValues := compositeKeyValues.values.Find(datatypes.String("1"), nil).(*keyValues)
			g.Expect(oneValues.values.Find(datatypes.String("one"), nil)).ToNot(BeNil())
			g.Expect(oneValues.values.Find(datatypes.String("other"), nil)).To(BeNil())

			// "2" values
			twoValues := compositeKeyValues.values.Find(datatypes.String("2"), nil).(*keyValues)
			g.Expect(twoValues.values.Find(datatypes.String("two"), nil)).ToNot(BeNil())
			g.Expect(twoValues.values.Find(datatypes.String("foo"), nil)).To(BeNil())

			// "3" values should still exist
			threeValues := compositeKeyValues.values.Find(datatypes.String("3"), nil).(*keyValues)
			g.Expect(threeValues.values.Find(datatypes.String("three"), nil)).ToNot(BeNil())

			// "4" values should still exist
			fourValues := compositeKeyValues.values.Find(datatypes.String("4"), nil).(*keyValues)
			g.Expect(fourValues.values.Find(datatypes.String("four"), nil)).ToNot(BeNil())

			// "5" values should be completely removed
			g.Expect(compositeKeyValues.values.Find(datatypes.String("5"), nil)).To(BeNil())
		})
	})
}
