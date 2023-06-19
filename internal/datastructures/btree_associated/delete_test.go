package btreeassociated

import (
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/btree_associated/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestCompositeTree_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

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

	t.Run("it does not run the delete function if the item is not found", func(t *testing.T) {
		associatedTree := setup()

		called := false
		canDelete := func(item any) bool {
			called = true
			return false
		}

		associatedTree.Delete(datatypes.StringMap{"not": "found"}, canDelete)
		g.Expect(called).To(BeFalse())
	})

	t.Run("it does not run the delete function if can delete returns false", func(t *testing.T) {
		associatedTree := New()
		_, err := associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), nil)
		g.Expect(err).ToNot(HaveOccurred())

		deleteCalled := false
		canDelete := func(item any) bool {
			deleteCalled = true
			return false
		}
		associatedTree.Delete(keyValues1, canDelete)
		g.Expect(deleteCalled).To(BeTrue())

		g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeFalse())
		g.Expect(associatedTree.Get(keyValues1, nil)).ToNot(BeNil())
	})

	t.Run("deletingthe only item in a tree", func(t *testing.T) {
		t.Run("it can delete an item with a single key value pair", func(t *testing.T) {
			associatedTree := New()
			_, err := associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), nil)
			g.Expect(err).ToNot(HaveOccurred())

			deleteCalled := false
			canDelete := func(item any) bool {
				deleteCalled = true
				return true
			}
			associatedTree.Delete(keyValues1, canDelete)
			g.Expect(deleteCalled).To(BeTrue())

			g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeTrue())
		})

		t.Run("it can delete an item with multiple key valu pairs", func(t *testing.T) {
			associatedTree := New()
			_, err := associatedTree.CreateOrFind(keyValues2, NewJoinTreeTester("2"), nil)
			g.Expect(err).ToNot(HaveOccurred())

			deleteCalled := false
			canDelete := func(item any) bool {
				deleteCalled = true
				return true
			}
			associatedTree.Delete(keyValues2, canDelete)
			g.Expect(deleteCalled).To(BeTrue())

			g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeTrue())
		})
	})

	t.Run("deletingthe one of many items in a tree", func(t *testing.T) {
		t.Run("it keeps the grouped key values if there are more than one", func(t *testing.T) {
			associatedTree := New()
			_, err := associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("a"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues1a, NewJoinTreeTester("b"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues2, NewJoinTreeTester("c"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = associatedTree.CreateOrFind(keyValues2a, NewJoinTreeTester("d"), nil)
			g.Expect(err).ToNot(HaveOccurred())

			calledCount := 0
			canDelete := func(item any) bool {
				calledCount++
				return true
			}
			associatedTree.Delete(keyValues1, canDelete)
			associatedTree.Delete(keyValues2, canDelete)
			g.Expect(calledCount).To(Equal(2))

			var item *JoinTreeTester
			getItem := func(foundItem any) {
				item = foundItem.(*JoinTreeTester)
			}

			_, err = associatedTree.Get(keyValues1a, getItem)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Value).To(Equal("b"))

			_, err = associatedTree.Get(keyValues2a, getItem)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Value).To(Equal("d"))
		})
	})
}
