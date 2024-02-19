package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Delete_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the keyValues is empty", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Delete(nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("KeyValues cannot be empty"))
	})

	t.Run("it returns an error if the keyValues has _associated_id", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Delete(datatypes.KeyValues{ReservedID: datatypes.Int(1)}, nil)
		g.Expect(err).To(Equal(ErrorKeyValuesHasAssociatedID))
	})
}

func TestAssociatedTree_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item AssociatedKeyValues) {}

	onDeleteTrue := func(item AssociatedKeyValues) bool { return true }
	onDeleteFalse := func(item AssociatedKeyValues) bool { return false }

	t.Run("It deletes the key value pair if the onDelete callback is nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		_, _ = associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.Delete(keys, nil)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It deletes the key value pair if the onDelete callback returns true", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		_, _ = associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.Delete(keys, onDeleteTrue)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It does not the key value pair if the onDelete callback returns false", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		_, _ = associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.Delete(keys, onDeleteFalse)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())
	})

	t.Run("Context when there are multiple ids in an ID Node", func(t *testing.T) {
		t.Run("It truncates the ID Node to the smallest size", func(t *testing.T) {
			associatedTree := NewThreadSafe()

			keys1 := datatypes.KeyValues{"1": datatypes.Int(1)}
			keys2 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}

			_, _ = associatedTree.CreateOrFind(keys1, noOpOnCreate, noOpOnFind)
			_, _ = associatedTree.CreateOrFind(keys2, noOpOnCreate, noOpOnFind)

			// check before the delete
			idNodeCalled := false
			idNodeCallback := func(item any) {
				idNodeCalled = true

				idNode := item.(*threadsafeIDNode)
				g.Expect(len(idNode.ids)).To(Equal(2))
				g.Expect(len(idNode.ids[0])).To(Equal(1))
				g.Expect(len(idNode.ids[1])).To(Equal(1))
			}

			associatedTree.keys.Find(datatypes.String("1"), func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Find(datatypes.Int(1), idNodeCallback)
			})
			g.Expect(idNodeCalled).To(BeTrue())

			g.Expect(associatedTree.Delete(keys2, onDeleteTrue)).ToNot(HaveOccurred())

			// check after the delete
			idNodeCalled = false
			idNodeCallbackAfterDelete := func(item any) {
				idNodeCalled = true

				idNode := item.(*threadsafeIDNode)
				g.Expect(len(idNode.ids)).To(Equal(1))
				g.Expect(len(idNode.ids[0])).To(Equal(1))
			}

			associatedTree.keys.Find(datatypes.String("1"), func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Find(datatypes.Int(1), idNodeCallbackAfterDelete)
			})
			g.Expect(idNodeCalled).To(BeTrue())
		})
	})
}

func TestAssociatedTree_DeleteByAssociatedID(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item AssociatedKeyValues) {}

	onDeleteTrue := func(item AssociatedKeyValues) bool { return true }
	onDeleteFalse := func(item AssociatedKeyValues) bool { return false }

	t.Run("It deletes the key value pair if the onDelete callback is nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		id, _ := associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.DeleteByAssociatedID(id, nil)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It deletes the key value pair if the onDelete callback returns true", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		id, _ := associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.DeleteByAssociatedID(id, onDeleteTrue)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It does not delete the key value pair if the onDelete callback returns false", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		id, _ := associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)

		g.Expect(associatedTree.DeleteByAssociatedID(id, onDeleteFalse)).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())

		// ensure find stille works
		found := false
		onFind := func(item AssociatedKeyValues) {
			found = true
		}
		g.Expect(associatedTree.FindByAssociatedID(id, onFind)).ToNot(HaveOccurred())
		g.Expect(found).To(BeTrue())
	})

	t.Run("Context when there are multiple ids in an ID Node", func(t *testing.T) {
		t.Run("It truncates the ID Node to the smallest size", func(t *testing.T) {
			associatedTree := NewThreadSafe()

			keys1 := datatypes.KeyValues{"1": datatypes.Int(1)}
			keys2 := datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.String("2")}

			_, _ = associatedTree.CreateOrFind(keys1, noOpOnCreate, noOpOnFind)
			id2, _ := associatedTree.CreateOrFind(keys2, noOpOnCreate, noOpOnFind)

			// check before the delete
			idNodeCalled := false
			idNodeCallback := func(item any) {
				idNodeCalled = true

				idNode := item.(*threadsafeIDNode)
				g.Expect(len(idNode.ids)).To(Equal(2))
				g.Expect(len(idNode.ids[0])).To(Equal(1))
				g.Expect(len(idNode.ids[1])).To(Equal(1))
			}

			associatedTree.keys.Find(datatypes.String("1"), func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Find(datatypes.Int(1), idNodeCallback)
			})
			g.Expect(idNodeCalled).To(BeTrue())

			g.Expect(associatedTree.DeleteByAssociatedID(id2, onDeleteTrue)).ToNot(HaveOccurred())

			// check after the delete
			idNodeCalled = false
			idNodeCallbackAfterDelete := func(item any) {
				idNodeCalled = true

				idNode := item.(*threadsafeIDNode)
				g.Expect(len(idNode.ids)).To(Equal(1))
				g.Expect(len(idNode.ids[0])).To(Equal(1))
			}

			associatedTree.keys.Find(datatypes.String("1"), func(item any) {
				valuesNode := item.(*threadsafeValuesNode)
				valuesNode.values.Find(datatypes.Int(1), idNodeCallbackAfterDelete)
			})
			g.Expect(idNodeCalled).To(BeTrue())
		})
	})
}
