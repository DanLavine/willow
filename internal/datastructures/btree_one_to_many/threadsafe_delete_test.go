package btreeonetomany

import (
	"fmt"
	"testing"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_DeleteOneOfManyByKeyValues(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.DeleteOneOfManyByKeyValues("", datatypes.KeyValues{"one": datatypes.Int(1)}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the associatedKeyValues is invalid", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.DeleteOneOfManyByKeyValues("one id", nil, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("recieved no KeyValues, but requires a length of at least 1"))
		})
	})

	t.Run("It perfroms a no-op if the OneID is not found", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := tree.DeleteOneOfManyByKeyValues("not found", datatypes.KeyValues{"assocID13": datatypes.Int(13)}, nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("It perfroms a no-op if the ManyKeyValues are not found", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"not found": datatypes.Float32(3.8)}, nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("It can delete the manyKeyValuse if canDelete is nil", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"13": datatypes.Int(13)}, nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(49))
	})

	t.Run("It can delete the manyKeyValues if canDelete returns true", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"13": datatypes.Int(13)}, func(item OneToManyItem) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(49))
	})

	t.Run("It does not delete the manyKeyValues if canDelete returns false", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"13": datatypes.Int(13)}, func(item OneToManyItem) bool { return false })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the tree is still valid
		manyRelations := 0
		onPaginate := func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelations++
			return true
		}
		onFind := func(_ datatypes.EncapsulatedValue, item any) bool {
			threadsafeValuesNode := item.(*threadsafeValuesNode)
			g.Expect(threadsafeValuesNode.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onPaginate)).ToNot(HaveOccurred())

			return true
		}
		g.Expect(tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), onFind)).ToNot(HaveOccurred())
		g.Expect(manyRelations).To(Equal(50))
	})

	t.Run("Context when the manyKeyValues is already destroying", func(t *testing.T) {
		t.Run("It returns a proper error", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			var id17 string
			for i := 0; i < 50; i++ {
				id, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
				g.Expect(err).ToNot(HaveOccurred())

				if i == 17 {
					id17 = id
				}
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
				_ = tree.DestroyOneOfManyByID("one name", id17, canDelete)
			}()

			g.Eventually(deleting).Should(Receive())

			err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"17": datatypes.Int(17)}, func(item OneToManyItem) bool { return true })
			g.Expect(err).To(Equal(ErrorManyKeyValuesDestroying))

			close(deleting)
		})
	})

	t.Run("Context when the tree is already destroying", func(t *testing.T) {
		t.Run("It returns a proper error", func(t *testing.T) {
			tree := NewThreadSafe()

			// create 50 items all under the same one tree
			for i := 0; i < 50; i++ {
				_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
				g.Expect(err).ToNot(HaveOccurred())
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

			err := tree.DeleteOneOfManyByKeyValues("one name", datatypes.KeyValues{"17": datatypes.Int(17)}, func(item OneToManyItem) bool { return true })
			g.Expect(err).To(Equal(ErrorOneIDDestroying))

			close(deleting)
		})
	})
}
