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

func TestOneToManyTree_CreateOrFind(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if oneName is the empty string", func(t *testing.T) {
			tree := NewThreadSafe()

			id, err := tree.CreateOrFind("", datatypes.KeyValues{}, nil, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
			g.Expect(id).To(Equal(""))
		})

		t.Run("It returns an error if the KeyValues are bad", func(t *testing.T) {
			tree := NewThreadSafe()

			id, err := tree.CreateOrFind("one name", datatypes.KeyValues{}, nil, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("recieved no KeyValues, but requires a length of at least 1"))
			g.Expect(id).To(Equal(""))

		})

		t.Run("It returns an error if the onCreate is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			id, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, nil, nil)
			g.Expect(err).To(Equal(ErrorOnCreateNil))
			g.Expect(id).To(Equal(""))
		})

		t.Run("It returns an error if the onFind is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			id, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return true }, nil)
			g.Expect(err).To(Equal(ErrorOnFindNil))
			g.Expect(id).To(Equal(""))
		})
	})

	t.Run("It can create a new OneToMany tree", func(t *testing.T) {
		tree := NewThreadSafe()

		createID, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(createID).ToNot(Equal(""))

		// find the one relation
		var oneRelation *threadsafeValuesNode
		tree.oneKeys.Find(datatypes.String("one name"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
			oneRelation = item.(*threadsafeValuesNode)
			return true
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelatiosn = append(manyRelatiosn, item.Value().(OneToManyItem).Value())
			return true
		})
		g.Expect(manyRelatiosn).To(Equal([]any{"true"}))
	})

	t.Run("It allows for queries of the reeturned ID", func(t *testing.T) {
		tree := NewThreadSafe()

		createID, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(createID).ToNot(Equal(""))

		// query
		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				IDs: []string{createID},
			},
		}

		found := 0
		onIterate := func(item OneToManyItem) bool {
			found++
			g.Expect(item.OneID()).To(Equal("one name"))
			g.Expect(item.ManyID()).To(Equal(createID))
			g.Expect(item.ManyKeyValues()).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1)}))
			g.Expect(item.Value()).To(Equal("true"))
			return true
		}

		g.Expect(tree.QueryAction("one name", query, onIterate)).ToNot(HaveOccurred())
		g.Expect(found).To(Equal(1))
	})

	t.Run("It can create many items for a One relationship tree", func(t *testing.T) {
		tree := NewThreadSafe()

		id1, err := tree.CreateOrFind("one name", datatypes.KeyValues{"1": datatypes.Int(1)}, func() any { return "1" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := tree.CreateOrFind("one name", datatypes.KeyValues{"2": datatypes.Int(1), "something else": datatypes.Float32(3.7)}, func() any { return "2" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name", datatypes.KeyValues{"3": datatypes.Int(1)}, func() any { return "3" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name", datatypes.KeyValues{"4": datatypes.Int(1)}, func() any { return "4" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name", datatypes.KeyValues{"5": datatypes.Int(1), "other": datatypes.Int(3)}, func() any { return "5" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name", datatypes.KeyValues{"6": datatypes.Int(1)}, func() any { return "6" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id1).ToNot(And(Equal(""), Equal(id2)))

		// find the one relation
		var oneRelation *threadsafeValuesNode
		tree.oneKeys.Find(datatypes.String("one name"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
			oneRelation = item.(*threadsafeValuesNode)
			return true
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			manyRelatiosn = append(manyRelatiosn, associatedKeyValues.Value().(OneToManyItem).Value().(string))
			return true
		})
		g.Expect(len(manyRelatiosn)).To(Equal(6))
		g.Expect(manyRelatiosn).To(ContainElements([]any{"1", "2", "3", "4", "5", "6"}))
	})

	t.Run("It can create many One relationship tree", func(t *testing.T) {
		tree := NewThreadSafe()

		id1, err := tree.CreateOrFind("one name 1", datatypes.KeyValues{"1": datatypes.Int(1)}, func() any { return "1" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := tree.CreateOrFind("one name 2", datatypes.KeyValues{"2": datatypes.Int(1), "something else": datatypes.Float32(3.7)}, func() any { return "2" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name 3", datatypes.KeyValues{"3": datatypes.Int(1)}, func() any { return "3" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name 4", datatypes.KeyValues{"4": datatypes.Int(1)}, func() any { return "4" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name 5", datatypes.KeyValues{"5": datatypes.Int(1), "other": datatypes.Int(3)}, func() any { return "5" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())
		_, err = tree.CreateOrFind("one name 6", datatypes.KeyValues{"6": datatypes.Int(1)}, func() any { return "6" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id1).ToNot(And(Equal(""), Equal(id2)))

		// find the one relation
		var oneRelations []*threadsafeValuesNode
		tree.oneKeys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
			oneRelations = append(oneRelations, item.(*threadsafeValuesNode))
			return true
		})
		g.Expect(len(oneRelations)).To(Equal(6))

		// find the many relation
		for _, oneRelation := range oneRelations {
			var manyRelatiosn []any
			oneRelation.associaedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
				manyRelatiosn = append(manyRelatiosn, associatedKeyValues.Value().(OneToManyItem))
				return true
			})
			g.Expect(len(manyRelatiosn)).To(Equal(1))
		}
	})

	t.Run("It runs onFind if the key values already esist", func(t *testing.T) {
		tree := NewThreadSafe()

		id1, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).ToNot(HaveOccurred())

		found := false
		id2, err := tree.CreateOrFind("one name", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { panic("nothing to return") }, func(item OneToManyItem) { found = true })
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(id1).To(Equal(id2))
		g.Expect(found).To(BeTrue())
	})

	t.Run("It returns an error if the many relation is being destroyed", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			_, err := tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return true }, func(oneToManyItem OneToManyItem) { panic("should not find") })
			g.Expect(err).ToNot(HaveOccurred())
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

		id, err := tree.CreateOrFind("one name", datatypes.KeyValues{"another": datatypes.Int(100)}, func() any { panic("dont return") }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).To(Equal(ErrorOneIDDestroying))
		g.Expect(id).To(Equal(""))

		close(deleting)
	})

	t.Run("It returns an error if the one relation is being destroyed", func(t *testing.T) {
		// TODO
	})
}
