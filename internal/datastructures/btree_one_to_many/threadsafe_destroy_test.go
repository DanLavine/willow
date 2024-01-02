package btreeonetomany

import (
	"fmt"
	"testing"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_DestroyOne(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.DestroyOne("", nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})
	})

	t.Run("It removes all values for the One relation if the callback is nil", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOne("one name", nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the One relation was destroyed
		found := false
		onFind := func(_ datatypes.EncapsulatedValue, _ any) bool {
			found = true
			return false
		}

		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})

	t.Run("It removes all values for the One relation if the callback returns true", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOne("one name", func(item OneToManyItem) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the One relation was destroyed
		found := false
		onFind := func(_ datatypes.EncapsulatedValue, _ any) bool {
			found = true
			return false
		}

		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})

	t.Run("It stops proccessing the One relation destroy if canDelete returns false", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		counter := 0
		canDelete := func(item OneToManyItem) bool {
			if counter >= 25 {
				return false
			}

			counter++
			return true
		}
		err := tree.DestroyOne("one name", canDelete)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the proper number of many relations exist
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(25))
	})

	t.Run("Context when a destroy is in progress", func(t *testing.T) {
		t.Run("It returns an error if the same One Relation is used", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			for i := 0; i < 50; i++ {
				g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
			}

			deleting := make(chan struct{})
			go func() {
				counter := 0
				canDelete := func(item OneToManyItem) bool {
					if counter == 0 {
						deleting <- struct{}{}
						<-deleting
						counter++
					}
					return true
				}
				_ = tree.DestroyOne("one name", canDelete)
			}()

			g.Eventually(deleting).Should(Receive())

			err := tree.DestroyOne("one name", nil)
			g.Expect(err).To(Equal(ErrorOneIDDestroying))

			deleting <- struct{}{}
		})

		t.Run("It can destry another key in parallel ", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			for i := 0; i < 50; i++ {
				g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
				g.Expect(tree.CreateWithID("two name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
			}

			deleting := make(chan struct{})
			go func() {
				defer close(deleting)

				counter := 0
				canDelete := func(item OneToManyItem) bool {
					if counter == 0 {
						deleting <- struct{}{}
						<-deleting
						counter++
					}
					return true
				}
				_ = tree.DestroyOne("one name", canDelete)
			}()

			g.Eventually(deleting).Should(Receive())
			go func() {
				deleting <- struct{}{}
			}()

			err := tree.DestroyOne("two name", nil)
			g.Expect(err).ToNot(HaveOccurred())

			g.Eventually(deleting).Should(BeClosed())

			// ensure the One relations were destroyed
			found := false
			onFind := func(_ datatypes.EncapsulatedValue, _ any) bool {
				found = true
				return false
			}

			g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
			g.Expect(found).To(BeFalse())
		})
	})
}

func TestOneToManyTree_DestroyOneOfManyByID(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.DestroyOneOfManyByID("", "some id", nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the manyID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.DestroyOneOfManyByID("one iid", "", nil)
			g.Expect(err).To(Equal(ErrorManyIDEmpty))
		})
	})

	t.Run("It perfroms a no-op if the OneID is not found", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOneOfManyByID("not found", "assocID13", nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("It perfroms a no-op if the ManyID is not found", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOneOfManyByID("one name", "not found", nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("It can delete the ManyID if canDelete is nil", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOneOfManyByID("one name", "assocID13", nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(49))
	})

	t.Run("It can delete the ManyID if canDelete returns true", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOneOfManyByID("one name", "assocID13", func(item OneToManyItem) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(49))
	})

	t.Run("It does not delete the ManyID if canDelete returns false", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
		}

		err := tree.DestroyOneOfManyByID("one name", "assocID13", func(item OneToManyItem) bool { return false })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Iterate(onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("Context when the ManyID is already destroying", func(t *testing.T) {
		t.Run("It returns a proper error", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			for i := 0; i < 50; i++ {
				g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
			}

			// call delete
			deleting := make(chan struct{})
			go func() {
				counter := 0
				canDelete := func(item OneToManyItem) bool {
					if counter == 0 {
						deleting <- struct{}{}
						<-deleting
						counter++
					}
					return true
				}
				_ = tree.DestroyOneOfManyByID("one name", "assocID17", canDelete)
			}()

			g.Eventually(deleting).Should(Receive())

			err := tree.DestroyOneOfManyByID("one name", "assocID17", nil)
			g.Expect(err).To(Equal(ErrorManyIDDestroying))

			close(deleting)
		})
	})

	t.Run("Context when the tree is already destroying", func(t *testing.T) {
		t.Run("It returns a proper error", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			for i := 0; i < 50; i++ {
				g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true })).ToNot(HaveOccurred())
			}

			// call delete
			deleting := make(chan struct{})
			go func() {
				counter := 0
				canDelete := func(item OneToManyItem) bool {
					if counter == 0 {
						deleting <- struct{}{}
						<-deleting
						counter++
					}
					return true
				}
				_ = tree.DestroyOne("one name", canDelete)
			}()

			g.Eventually(deleting).Should(Receive())

			err := tree.DestroyOneOfManyByID("one name", "assicID17", nil)
			g.Expect(err).To(Equal(ErrorOneIDDestroying))

			close(deleting)
		})
	})
}
