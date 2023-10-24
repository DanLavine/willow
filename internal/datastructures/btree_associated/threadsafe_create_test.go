package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_CreateOrFind_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	onCreate := func() any { return true }
	onFind := func(item any) {}

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(nil, onCreate, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error if the keyValuePairs contains '_associated_id' key", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(datatypes.KeyValues{"_associated_id": datatypes.Int(1)}, onCreate, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("reserved keyword '_associated_id' cannot be used in the keyValuePairs"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(keys, nil, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be nil"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(keys, onCreate, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
		g.Expect(id).To(Equal(""))
	})
}

func TestAssociatedTree_CreateOrFind_FailedToCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item any) {}

	t.Run("it cleans up any possible values that were created to store a new ID", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return nil
		}

		id, err := associatedTree.CreateOrFind(keyValues, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
		g.Expect(id).ToNot(Equal(""))
	})

	t.Run("it does not remove any IDs that might already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		goodKeyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(4)}
		goodKeyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5), "3": datatypes.String("three")}
		badKeyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(4)} // This tests the break on processing shrinks

		badCreate := func() any {
			return nil
		}

		_, _ = associatedTree.CreateOrFind(goodKeyValues1, noOpOnCreate, noOpOnFind)
		_, _ = associatedTree.CreateOrFind(goodKeyValues2, noOpOnCreate, noOpOnFind)
		_, _ = associatedTree.CreateOrFind(badKeyValues, badCreate, noOpOnFind)
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())

		foundCounter := 0
		onFind := func(key string) func(item any) {
			return func(item any) {
				foundCounter++

				valuesNode := item.(*threadsafeValuesNode)
				g.Expect(valuesNode.values.Empty()).To(BeFalse())

				called := 0
				switch key {
				case "1":
					valuesNode.values.Find(datatypes.String("one"), func(item any) {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))
						g.Expect(len(idNode.ids[2])).To(Equal(1))
					})

					g.Expect(called).To(Equal(1))
				case "2":
					// this shouldn't be found
					shouldNotFind := false
					valuesNode.values.Find(datatypes.Int(4), func(item any) {
						shouldNotFind = true
					})
					g.Expect(shouldNotFind).To(BeFalse())

					valuesNode.values.Find(datatypes.Int(5), func(item any) {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))
					})

					g.Expect(called).To(Equal(1))
				case "3":
					valuesNode.values.Find(datatypes.String("three"), func(item any) {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))
					})

					g.Expect(called).To(Equal(1))
				case "4":
					valuesNode.values.Find(datatypes.Int(4), func(item any) {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(2))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))
					})

					g.Expect(called).To(Equal(1))
				}
			}
		}
		g.Expect(associatedTree.keys.Find(datatypes.String("1"), onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("2"), onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("3"), onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("4"), onFind("4"))).ToNot(HaveOccurred())
		g.Expect(foundCounter).To(Equal(4))
	})
}

func TestAssociatedTree_CreateOrFind_SingleKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one")}
	noOpOnCreate := func() any { return "found me" }
	noOpOnFind := func(item any) {}

	t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return true
		}

		id, err := associatedTree.CreateOrFind(keyValues, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(id).ToNot(Equal(""))
	})

	t.Run("it finds an item that already exists", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onFind := func(item any) {
			called = true
			g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("found me"))
		}

		id1, err := associatedTree.CreateOrFind(keyValues, noOpOnCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		id2, err := associatedTree.CreateOrFind(keyValues, noOpOnCreate, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(id1).To(Equal(id2))
	})

	t.Run("it properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one")}
		keyValues2 := datatypes.KeyValues{"1": datatypes.Int(5)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.String("three")}

		createCount := 0
		onCreate := func() any {
			createCount++
			return true
		}

		id1, err := associatedTree.CreateOrFind(keyValues1, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := associatedTree.CreateOrFind(keyValues2, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		id3, err := associatedTree.CreateOrFind(keyValues3, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id1).ToNot(And(Equal(id2), Equal(id3)))
		g.Expect(id2).ToNot(Equal(id3))
		g.Expect(createCount).To(Equal(3))
	})

	t.Run("It returns the ID for the saved item in the tree", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		var foundItem any
		onFind := func(item any) {
			foundItem = item
		}

		id, err := associatedTree.CreateOrFind(keyValues, noOpOnCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the value is whats actually saved as the tree's id
		foundItem = nil
		err = associatedTree.ids.Find(datatypes.String(id), onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundItem).ToNot(BeNil())
	})
}

func TestAssociatedTree_CreateOrFind_MultiKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item any) {}

	t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return true
		}

		id, err := associatedTree.CreateOrFind(keyValues, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(id).ToNot(Equal(""))
	})

	t.Run("it finds an item that already exists", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onFind := func(item any) {
			called = true
			g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("find me"))
		}

		associatedTree.CreateOrFind(keyValues, noOpOnCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues, noOpOnCreate, onFind)
		g.Expect(called).To(BeTrue())
	})

	t.Run("it properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float64(3.0)}
		keyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
		keyValues3 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(5.0)}
		keyValues4 := datatypes.KeyValues{"2": datatypes.String("two"), "3": datatypes.Float32(3.0)}

		createCount := 0
		onCreate := func() any {
			createCount++
			return createCount
		}

		associatedTree.CreateOrFind(keyValues1, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues2, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues3, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues4, onCreate, noOpOnFind)
		g.Expect(createCount).To(Equal(4))

		found := false
		onFind := func(item any) {
			found = true
			g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(2))
		}

		associatedTree.CreateOrFind(keyValues2, onCreate, onFind)
		g.Expect(found).To(BeTrue())
	})
}
