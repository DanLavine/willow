package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/internal/datastructures/btree_associated/testhelpers"
	. "github.com/onsi/gomega"
)

func TestCompositeTree_Get(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the 'keyValuePairs' are nil", func(t *testing.T) {
		associatedTree := New()

		item, err := associatedTree.Get(nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("keyValuePairs requires a length of at least 1"))
		g.Expect(item).To(BeNil())
	})

	t.Run("it returns an error if the 'keyValuePairs' are empty", func(t *testing.T) {
		associatedTree := New()

		item, err := associatedTree.Get(datatypes.StringMap{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("keyValuePairs requires a length of at least 1"))
		g.Expect(item).To(BeNil())
	})

	t.Run("when the tree is empty", func(t *testing.T) {
		t.Run("it returns nothing, no matter the query", func(t *testing.T) {
			associatedTree := New()

			item, err := associatedTree.Get(datatypes.StringMap{"one": "1"}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).To(BeNil())
		})
	})

	t.Run("context when the tree is populated", func(t *testing.T) {
		keyValues1 := datatypes.StringMap{
			"1": "other",
		}
		keyValues1a := datatypes.StringMap{
			"1": "a",
		}
		keyValues2 := datatypes.StringMap{
			"1": "other",
			"2": "foo",
		}
		keyValues2a := datatypes.StringMap{
			"1": "other",
			"2": "a",
		}

		setup := func() *associatedTree {
			associatedTree := New()

			_, err := associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues1a, NewJoinTreeTester("1a"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues2, NewJoinTreeTester("2"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues2a, NewJoinTreeTester("2a"), nil)
			g.Expect(err).ToNot(HaveOccurred())

			return associatedTree
		}

		t.Run("it returns a single key value that is found in the tree", func(t *testing.T) {
			associatedTree := setup()

			item, err := associatedTree.Get(datatypes.StringMap{"1": "other"}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())
			g.Expect(item.(*JoinTreeTester).Value).To(Equal("1"))
		})

		t.Run("it returns a multi key value pair that is found in the tree", func(t *testing.T) {
			associatedTree := setup()

			item, err := associatedTree.Get(datatypes.StringMap{"1": "other", "2": "foo"}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())
			g.Expect(item.(*JoinTreeTester).Value).To(Equal("2"))
		})

		t.Run("it runs the 'onFind' callback if an item is found", func(t *testing.T) {
			associatedTree := setup()

			onFind := func(item any) {
				joinTreeTester := item.(*JoinTreeTester)
				joinTreeTester.OnFindCount++
			}

			item, err := associatedTree.Get(datatypes.StringMap{"1": "other", "2": "foo"}, onFind)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())
			g.Expect(item.(*JoinTreeTester).Value).To(Equal("2"))
			g.Expect(item.(*JoinTreeTester).OnFindCount).To(Equal(1))
		})

		t.Run("it returns nothing if a key is not found", func(t *testing.T) {
			associatedTree := setup()

			item, err := associatedTree.Get(datatypes.StringMap{"not found": "not found"}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).To(BeNil())
		})

		t.Run("it returns nothing if a key is found, but no value", func(t *testing.T) {
			associatedTree := setup()

			item, err := associatedTree.Get(datatypes.StringMap{"1": "not found"}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).To(BeNil())
		})
	})
}
