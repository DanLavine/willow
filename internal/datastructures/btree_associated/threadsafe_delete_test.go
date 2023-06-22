package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestCompositeTree_Delete_ParameterChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if keyValuePairs doesn't have a len of 1", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Delete(nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("keyValuePairs is nil"))
	})
}

func TestCompositeTree_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	onFindNoOp := func(item any) {}

	keyValues1 := datatypes.StringMap{
		"1": datatypes.String("other"),
	}
	keyValues1a := datatypes.StringMap{
		"1": datatypes.String("a"),
	}
	keyValues2 := datatypes.StringMap{
		"1": datatypes.String("other"),
		"2": datatypes.String("foo"),
	}
	keyValues2a := datatypes.StringMap{
		"1": datatypes.String("other"),
		"2": datatypes.String("a"),
	}

	setup := func() *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		g.Expect(associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), onFindNoOp)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues1a, NewJoinTreeTester("1a"), onFindNoOp)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, NewJoinTreeTester("2"), onFindNoOp)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2a, NewJoinTreeTester("2a"), onFindNoOp)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("it does not run the delete function if the item is not found", func(t *testing.T) {
		associatedTree := setup()

		called := false
		canDelete := func(item any) bool {
			called = true
			return false
		}

		g.Expect(associatedTree.Delete(datatypes.StringMap{"not": datatypes.Float64(2.0)}, canDelete)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it does not run the delete function if can delete returns false", func(t *testing.T) {
		associatedTree := setup()

		deleteCalled := false
		canDelete := func(item any) bool {
			deleteCalled = true
			return false
		}

		g.Expect(associatedTree.Delete(keyValues1, canDelete)).ToNot(HaveOccurred())
		g.Expect(deleteCalled).To(BeTrue())

		g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeFalse())

		found := false
		onFind := func(item any) {
			jtt := item.(*JoinTreeTester)
			g.Expect(jtt.Value).To(Equal("1"))
			found = true
		}
		g.Expect(associatedTree.Find(keyValues1, onFind)).ToNot(HaveOccurred())
		g.Expect(found).To(BeTrue())
	})

	t.Run("context: deleting the only item in a tree", func(t *testing.T) {
		t.Run("it can delete a single key value and set the tree to empty", func(t *testing.T) {
			associatedTree := NewThreadSafe()
			g.Expect(associatedTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), onFindNoOp)).ToNot(HaveOccurred())

			deleteCalled := false
			canDelete := func(item any) bool {
				deleteCalled = true
				return true
			}

			g.Expect(associatedTree.Delete(keyValues1, canDelete)).ToNot(HaveOccurred())
			g.Expect(deleteCalled).To(BeTrue())

			g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeTrue())
		})

		t.Run("it can delete an item with multiple key value pairs and set the tree to empty", func(t *testing.T) {
			associatedTree := NewThreadSafe()
			g.Expect(associatedTree.CreateOrFind(keyValues2, NewJoinTreeTester("2"), onFindNoOp)).ToNot(HaveOccurred())

			deleteCalled := false
			canDelete := func(item any) bool {
				deleteCalled = true
				return true
			}

			g.Expect(associatedTree.Delete(keyValues2, canDelete)).ToNot(HaveOccurred())
			g.Expect(deleteCalled).To(BeTrue())

			g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeTrue())
		})
	})

	t.Run("deletingthe one of many items in a tree", func(t *testing.T) {
		t.Run("it keeps the grouped key values if there are more than one", func(t *testing.T) {
			associatedTree := setup()

			calledCount := 0
			canDelete := func(item any) bool {
				calledCount++
				return true
			}

			g.Expect(associatedTree.Delete(keyValues1, canDelete)).ToNot(HaveOccurred())
			g.Expect(associatedTree.Delete(keyValues2, canDelete)).ToNot(HaveOccurred())
			g.Expect(calledCount).To(Equal(2))

			var item *JoinTreeTester
			getItem := func(foundItem any) {
				item = foundItem.(*JoinTreeTester)
			}

			g.Expect(associatedTree.Find(keyValues1a, getItem)).ToNot(HaveOccurred())
			g.Expect(item.Value).To(Equal("1a"))

			g.Expect(associatedTree.Find(keyValues2a, getItem)).ToNot(HaveOccurred())
			g.Expect(item.Value).To(Equal("2a"))
		})
	})
}
