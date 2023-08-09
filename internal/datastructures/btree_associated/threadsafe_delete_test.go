package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Delete_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the keyValuePairs is empty", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Delete(nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
	})
}

func TestAssociatedTree_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.StringMap{"1": datatypes.Int(1)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item any) {}

	onDeleteTrue := func(item any) bool { return true }
	onDeleteFalse := func(item any) bool { return false }

	t.Run("It deletes the key value pair if the onDelete callback is nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		g.Expect(associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())

		g.Expect(associatedTree.Delete(keys, nil)).ToNot(HaveOccurred())
		g.Expect(associatedTree.ids.Empty()).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It deletes the key value pair if the onDelete callback returns true", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		g.Expect(associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())

		g.Expect(associatedTree.Delete(keys, onDeleteTrue)).ToNot(HaveOccurred())
		g.Expect(associatedTree.ids.Empty()).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})

	t.Run("It does not the key value pair if the onDelete callback returns false", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		g.Expect(associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())

		g.Expect(associatedTree.Delete(keys, onDeleteFalse)).ToNot(HaveOccurred())
		g.Expect(associatedTree.ids.Empty()).To(BeFalse())
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())
	})

	t.Run("Context when there are multiple ids in an ID Node", func(t *testing.T) {
		t.Run("It truncates the ID Node to the smallest size", func(t *testing.T) {
			associatedTree := NewThreadSafe()

			keys1 := datatypes.StringMap{"1": datatypes.Int(1)}
			keys2 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.String("2")}

			g.Expect(associatedTree.CreateOrFind(keys1, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())
			g.Expect(associatedTree.CreateOrFind(keys2, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())

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
