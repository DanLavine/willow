package btreeonetomany

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_CreateWithID(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if oneName is the empty string", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.CreateWithID("", "", datatypes.KeyValues{}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if associatedID is the empty string", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.CreateWithID("one name", "", datatypes.KeyValues{}, nil)
			g.Expect(err).To(Equal(ErrorManyIDEmpty))
		})

		t.Run("It returns an error if the KeyValues are empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.CreateWithID("one name", "associated id", datatypes.KeyValues{}, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("KeyValues cannot be empty"))
		})

		t.Run("It returns an error if the onCreate is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.CreateWithID("one name", "associated id", datatypes.KeyValues{"one": datatypes.Int(1)}, nil)
			g.Expect(err).To(Equal(ErrorOnCreateNil))
		})
	})

	t.Run("It can create a new OneToMany tree", func(t *testing.T) {
		tree := NewThreadSafe()

		err := tree.CreateWithID("one name", "associated id", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" })
		g.Expect(err).ToNot(HaveOccurred())

		// find the one relation
		var oneRelation *threadsafeValuesNode
		tree.oneKeys.Find(datatypes.String("one name"), func(item any) {
			oneRelation = item.(*threadsafeValuesNode)
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(item btreeassociated.AssociatedKeyValues) bool {
			manyRelatiosn = append(manyRelatiosn, item.Value().(OneToManyItem).Value())
			return true
		})
		g.Expect(manyRelatiosn).To(Equal([]any{"true"}))
	})

	t.Run("It can create many items for a One relationship tree", func(t *testing.T) {
		tree := NewThreadSafe()

		g.Expect(tree.CreateWithID("one name", "associated id 1", datatypes.KeyValues{"1": datatypes.Int(1)}, func() any { return "1" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "associated id 2", datatypes.KeyValues{"2": datatypes.Int(1), "something else": datatypes.Float32(3.7)}, func() any { return "2" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "associated id 3", datatypes.KeyValues{"3": datatypes.Int(1)}, func() any { return "3" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "associated id 4", datatypes.KeyValues{"4": datatypes.Int(1)}, func() any { return "4" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "associated id 5", datatypes.KeyValues{"5": datatypes.Int(1), "other": datatypes.Int(3)}, func() any { return "5" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "associated id 6", datatypes.KeyValues{"6": datatypes.Int(1)}, func() any { return "6" })).ToNot(HaveOccurred())

		// find the one relation
		var oneRelation *threadsafeValuesNode
		tree.oneKeys.Find(datatypes.String("one name"), func(item any) {
			oneRelation = item.(*threadsafeValuesNode)
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			manyRelatiosn = append(manyRelatiosn, associatedKeyValues.Value().(OneToManyItem).Value().(string))
			return true
		})
		g.Expect(len(manyRelatiosn)).To(Equal(6))
		g.Expect(manyRelatiosn).To(ContainElements([]any{"1", "2", "3", "4", "5", "6"}))
	})

	t.Run("It can create many One relationship tree", func(t *testing.T) {
		tree := NewThreadSafe()

		g.Expect(tree.CreateWithID("one name 1", "associated id", datatypes.KeyValues{"1": datatypes.Int(1)}, func() any { return "1" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name 2", "associated id", datatypes.KeyValues{"2": datatypes.Int(1), "something else": datatypes.Float32(3.7)}, func() any { return "2" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name 3", "associated id", datatypes.KeyValues{"3": datatypes.Int(1)}, func() any { return "3" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name 4", "associated id", datatypes.KeyValues{"4": datatypes.Int(1)}, func() any { return "4" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name 5", "associated id", datatypes.KeyValues{"5": datatypes.Int(1), "other": datatypes.Int(3)}, func() any { return "5" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name 6", "associated id", datatypes.KeyValues{"6": datatypes.Int(1)}, func() any { return "6" })).ToNot(HaveOccurred())

		// find the one relation
		var oneRelations []*threadsafeValuesNode
		tree.oneKeys.Iterate(func(key datatypes.EncapsulatedValue, item any) bool {
			oneRelations = append(oneRelations, item.(*threadsafeValuesNode))
			return true
		})
		g.Expect(len(oneRelations)).To(Equal(6))

		// find the many relation
		for _, oneRelation := range oneRelations {
			var manyRelatiosn []any
			oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
				manyRelatiosn = append(manyRelatiosn, associatedKeyValues.Value().(OneToManyItem))
				return true
			})
			g.Expect(len(manyRelatiosn)).To(Equal(1))
		}
	})

	t.Run("It returns an error if the manyID already exists", func(t *testing.T) {
		tree := NewThreadSafe()

		err := tree.CreateWithID("one name", "associated id", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" })
		g.Expect(err).ToNot(HaveOccurred())

		err = tree.CreateWithID("one name", "associated id", datatypes.KeyValues{"two": datatypes.Int(2)}, func() any { return "true" })
		g.Expect(err).To(Equal(ErrorManyIDAlreadyExists))
	})

	t.Run("It returns an error if the manyKeyValues exists", func(t *testing.T) {
		tree := NewThreadSafe()

		err := tree.CreateWithID("one name", "associated id", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" })
		g.Expect(err).ToNot(HaveOccurred())

		err = tree.CreateWithID("one name", "associated id 2", datatypes.KeyValues{"one": datatypes.Int(1)}, func() any { return "true" })
		g.Expect(err).To(Equal(ErrorManyKeyValuesAlreadyExist))
	})

	t.Run("It returns an error if the many relation is being destroyed", func(t *testing.T) {
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

		err := tree.CreateWithID("one name", "assoc id", datatypes.KeyValues{"another": datatypes.Int(100)}, func() any { return true })
		g.Expect(err).To(Equal(ErrorOneIDDestroying))

		close(deleting)
	})

	t.Run("It returns an error if the one relation is being destroyed", func(t *testing.T) {
		// TODO
	})
}

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
			g.Expect(err.Error()).To(ContainSubstring("KeyValues cannot be empty"))
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
		tree.oneKeys.Find(datatypes.String("one name"), func(item any) {
			oneRelation = item.(*threadsafeValuesNode)
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(item btreeassociated.AssociatedKeyValues) bool {
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
		strIDValue := datatypes.String(createID)
		query := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"_associated_id": datatypes.Value{Value: &strIDValue, ValueComparison: datatypes.EqualsPtr()},
				},
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

		g.Expect(tree.Query("one name", query, onIterate)).ToNot(HaveOccurred())
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
		tree.oneKeys.Find(datatypes.String("one name"), func(item any) {
			oneRelation = item.(*threadsafeValuesNode)
		})
		g.Expect(oneRelation).ToNot(BeNil())

		// find the many relation
		var manyRelatiosn []any
		oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
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
		tree.oneKeys.Iterate(func(key datatypes.EncapsulatedValue, item any) bool {
			oneRelations = append(oneRelations, item.(*threadsafeValuesNode))
			return true
		})
		g.Expect(len(oneRelations)).To(Equal(6))

		// find the many relation
		for _, oneRelation := range oneRelations {
			var manyRelatiosn []any
			oneRelation.associaedTree.Query(datatypes.AssociatedKeyValuesQuery{}, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
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

		id, err := tree.CreateOrFind("one name", datatypes.KeyValues{"another": datatypes.Int(100)}, func() any { panic("dont return") }, func(item OneToManyItem) { panic(item) })
		g.Expect(err).To(Equal(ErrorOneIDDestroying))
		g.Expect(id).To(Equal(""))

		close(deleting)
	})

	t.Run("It returns an error if the one relation is being destroyed", func(t *testing.T) {
		// TODO
	})
}
