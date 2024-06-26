package btreeassociated

import (
	"testing"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Create_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	onCreate := func() any { return true }

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.Create(nil, onCreate)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("recieved no KeyValues, but requires a length of at least 1"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.Create(keys, nil)
		g.Expect(err).To(Equal(ErrorOnCreateNil))
		g.Expect(id).To(Equal(""))
	})
}

func TestAssociatedTree_Create_FailedToCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5)}
	noOpOnCreate := func() any { return "find me" }

	t.Run("It returns an error if the KeyValues already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(4)}

		id, err := associatedTree.Create(keyValues, noOpOnCreate)
		g.Expect(id).ToNot(Equal(""))
		g.Expect(err).To(BeNil())

		id, err = associatedTree.Create(keyValues, noOpOnCreate)
		g.Expect(id).To(Equal(""))
		g.Expect(err).To(Equal(ErrorKeyValuesAlreadyExists))

		g.Expect(associatedTree.keys.Empty()).To(BeFalse())
	})

	t.Run("It cleans up any possible values that were if onCreate returns nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return nil
		}

		id, err := associatedTree.Create(keyValues, onCreate)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
		g.Expect(id).To(Equal(""))
	})

	t.Run("It does not remove any IDs that might already exist when onCreate returns nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		goodKeyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(4)}
		goodKeyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5), "3": datatypes.String("three")}
		badKeyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(4)} // This tests the break on processing shrinks

		badCreate := func() any {
			return nil
		}

		_, _ = associatedTree.Create(goodKeyValues1, noOpOnCreate)
		_, _ = associatedTree.Create(goodKeyValues2, noOpOnCreate)
		_, _ = associatedTree.Create(badKeyValues, badCreate)
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())

		foundCounter := 0
		onFind := func(key string) func(_ datatypes.EncapsulatedValue, item any) bool {
			return func(_ datatypes.EncapsulatedValue, item any) bool {
				foundCounter++

				valuesNode := item.(*threadsafeValuesNode)
				g.Expect(valuesNode.values.Empty()).To(BeFalse())

				called := 0
				switch key {
				case "1":
					valuesNode.values.Find(datatypes.String("one"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return false
					})

					g.Expect(called).To(Equal(1))
				case "2":
					// this shouldn't be found
					shouldNotFind := false
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						shouldNotFind = true
						return false
					})
					g.Expect(shouldNotFind).To(BeFalse())

					valuesNode.values.Find(datatypes.Int(5), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return false
					})

					g.Expect(called).To(Equal(1))
				case "3":
					valuesNode.values.Find(datatypes.String("three"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return false
					})

					g.Expect(called).To(Equal(1))
				case "4":
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(2))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))

						return false
					})

					g.Expect(called).To(Equal(1))
				}

				return true
			}
		}
		g.Expect(associatedTree.keys.Find(datatypes.String("1"), testmodels.NoTypeRestrictions(g), onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("2"), testmodels.NoTypeRestrictions(g), onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("3"), testmodels.NoTypeRestrictions(g), onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("4"), testmodels.NoTypeRestrictions(g), onFind("4"))).ToNot(HaveOccurred())
		g.Expect(foundCounter).To(Equal(4))
	})
}

func TestAssociatedTree_Create_SingleKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one")}
	noOpOnCreate := func() any { return "found me" }

	t.Run("It creates a value if it doesn't already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return true
		}

		id, err := associatedTree.Create(keyValues, onCreate)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
		g.Expect(id).ToNot(Equal(""))
	})

	t.Run("It properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one")}
		keyValues2 := datatypes.KeyValues{"1": datatypes.Int(5)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.String("three")}

		createCount := 0
		onCreate := func() any {
			createCount++
			return true
		}

		id0, err := associatedTree.Create(keyValues0, onCreate)
		g.Expect(err).ToNot(HaveOccurred())
		id1, err := associatedTree.Create(keyValues1, onCreate)
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := associatedTree.Create(keyValues2, onCreate)
		g.Expect(err).ToNot(HaveOccurred())
		id3, err := associatedTree.Create(keyValues3, onCreate)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id0).ToNot(And(Equal(id1), Equal(id2), Equal(id3)))
		g.Expect(id1).ToNot(And(Equal(id2), Equal(id3)))
		g.Expect(id2).ToNot(Equal(id3))
		g.Expect(createCount).To(Equal(4))
	})

	t.Run("It returns the ID for the saved item in the tree", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		var foundItem any
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItem = item
			return true
		}

		id, err := associatedTree.Create(keyValues, noOpOnCreate)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the value is whats actually saved as the tree's id
		foundItem = nil
		err = associatedTree.associatedIDs.Find(datatypes.String(id), testmodels.NoTypeRestrictions(g), onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundItem).ToNot(BeNil())
	})
}

func TestAssociatedTree_Create_MultiKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float64(3.0)}
		keyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
		keyValues3 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(5.0)}
		keyValues4 := datatypes.KeyValues{"2": datatypes.String("two"), "3": datatypes.Float32(3.0)}

		createCount := 0
		onCreate := func() any {
			createCount++
			return createCount
		}

		associatedTree.Create(keyValues0, onCreate)
		associatedTree.Create(keyValues1, onCreate)
		associatedTree.Create(keyValues2, onCreate)
		associatedTree.Create(keyValues3, onCreate)
		associatedTree.Create(keyValues4, onCreate)
		g.Expect(createCount).To(Equal(5))
	})
}

func TestAssociatedTree_CreateWithID_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	onCreate := func() any { return true }

	t.Run("it returns an error if associatedID is empty", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateWithID("", keys, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorAssociatedIDEmpty))
	})

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateWithID("something", nil, onCreate)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("recieved no KeyValues, but requires a length of at least 1"))
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateWithID("something", keys, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorOnCreateNil))
	})
}

func TestAssociatedTree_CreateWithID_FailedToCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnCreate := func() any { return "find me" }

	t.Run("It returns an error if the KeyValues already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(4)}

		err := associatedTree.CreateWithID("key1", keyValues, func() any { return "value1" })
		g.Expect(err).To(BeNil())

		err = associatedTree.CreateWithID("key2", keyValues, func() any { return "value2" })
		g.Expect(err).To(Equal(ErrorKeyValuesAlreadyExists))

		// ensure the first value still exists
		foundValue := ""
		err = associatedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{Selection: &queryassociatedaction.Selection{IDs: []string{"key1"}}}, func(associatedKeyValues AssociatedKeyValues) bool {
			foundValue = associatedKeyValues.Value().(string)
			return true
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundValue).To(Equal("value1"))
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())
	})

	t.Run("It returns an error if the associatedID already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateWithID("key", datatypes.KeyValues{"1": datatypes.String("one")}, noOpOnCreate)
		g.Expect(err).To(BeNil())

		err = associatedTree.CreateWithID("key", datatypes.KeyValues{"4": datatypes.Int(4)}, noOpOnCreate)
		g.Expect(err).To(Equal(ErrorAssociatedIDAlreadyExists))

		// ensure the proper value still exists
		found := false
		err = associatedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{Selection: &queryassociatedaction.Selection{IDs: []string{"key"}}}, func(associatedKeyValues AssociatedKeyValues) bool {
			found = true
			return true
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())
	})

	t.Run("It does not remove any IDs that might already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		goodKeyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(4)}
		goodKeyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5), "3": datatypes.String("three")}
		badKeyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(4)} // This tests the break on processing shrinks

		badCreate := func() any {
			return nil
		}

		_, _ = associatedTree.Create(goodKeyValues1, noOpOnCreate)
		_, _ = associatedTree.Create(goodKeyValues2, noOpOnCreate)
		_, _ = associatedTree.Create(badKeyValues, badCreate)
		g.Expect(associatedTree.keys.Empty()).To(BeFalse())

		foundCounter := 0
		onFind := func(key string) func(_ datatypes.EncapsulatedValue, item any) bool {
			return func(_ datatypes.EncapsulatedValue, item any) bool {
				foundCounter++

				valuesNode := item.(*threadsafeValuesNode)
				g.Expect(valuesNode.values.Empty()).To(BeFalse())

				called := 0
				switch key {
				case "1":
					valuesNode.values.Find(datatypes.String("one"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "2":
					// this shouldn't be found
					shouldNotFind := false
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						shouldNotFind = true
						return true
					})
					g.Expect(shouldNotFind).To(BeFalse())

					valuesNode.values.Find(datatypes.Int(5), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "3":
					valuesNode.values.Find(datatypes.String("three"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "4":
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(2))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				}

				return true
			}
		}
		g.Expect(associatedTree.keys.Find(datatypes.String("1"), testmodels.NoTypeRestrictions(g), onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("2"), testmodels.NoTypeRestrictions(g), onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("3"), testmodels.NoTypeRestrictions(g), onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("4"), testmodels.NoTypeRestrictions(g), onFind("4"))).ToNot(HaveOccurred())
		g.Expect(foundCounter).To(Equal(4))
	})
}

func TestAssociatedTree_CreateWithID_SingleKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one")}
	noOpOnCreate := func() any { return "found me" }

	t.Run("it properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one")}
		keyValues2 := datatypes.KeyValues{"1": datatypes.Int(5)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.String("three")}

		createCount := 0
		onCreate := func() any {
			createCount++
			return true
		}

		g.Expect(associatedTree.CreateWithID("key0", keyValues0, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key1", keyValues1, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key2", keyValues2, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key3", keyValues3, onCreate)).ToNot(HaveOccurred())

		g.Expect(createCount).To(Equal(4))
	})

	t.Run("It can find the item with the id used for creation", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		var foundItem any
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItem = item
			return true
		}

		err := associatedTree.CreateWithID("key1", keyValues, noOpOnCreate)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the value is whats actually saved as the tree's id
		err = associatedTree.associatedIDs.Find(datatypes.String("key1"), testmodels.NoTypeRestrictions(g), onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundItem).ToNot(BeNil())
	})
}

func TestAssociatedTree_CreateWithID_MultiKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float64(3.0)}
		keyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
		keyValues3 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(5.0)}
		keyValues4 := datatypes.KeyValues{"2": datatypes.String("two"), "3": datatypes.Float32(3.0)}

		createCount := 0
		onCreate := func() any {
			createCount++
			return createCount
		}

		g.Expect(associatedTree.CreateWithID("key0", keyValues0, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key1", keyValues1, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key2", keyValues2, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key3", keyValues3, onCreate)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateWithID("key4", keyValues4, onCreate)).ToNot(HaveOccurred())
		g.Expect(createCount).To(Equal(5))

		found := false
		onFind := func(item AssociatedKeyValues) {
			found = true
			g.Expect(item.Value()).To(Equal(3))
		}

		associatedTree.CreateOrFind(keyValues2, onCreate, onFind)
		g.Expect(found).To(BeTrue())
	})
}

func TestAssociatedTree_CreateOrFind_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.KeyValues{"1": datatypes.Int(1)}
	onCreate := func() any { return true }
	onFind := func(item AssociatedKeyValues) {}

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(nil, onCreate, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("recieved no KeyValues, but requires a length of at least 1"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(keys, nil, onFind)
		g.Expect(err).To(Equal(ErrorOnCreateNil))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.CreateOrFind(keys, onCreate, nil)
		g.Expect(err).To(Equal(ErrorOnFindNil))
		g.Expect(id).To(Equal(""))
	})
}

func TestAssociatedTree_CreateOrFind_FailedToCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(5)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item AssociatedKeyValues) {}

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
		onFind := func(key string) func(_ datatypes.EncapsulatedValue, item any) bool {
			return func(_ datatypes.EncapsulatedValue, item any) bool {
				foundCounter++

				valuesNode := item.(*threadsafeValuesNode)
				g.Expect(valuesNode.values.Empty()).To(BeFalse())

				called := 0
				switch key {
				case "1":
					valuesNode.values.Find(datatypes.String("one"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "2":
					// this shouldn't be found
					shouldNotFind := false
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						shouldNotFind = true
						return true
					})
					g.Expect(shouldNotFind).To(BeFalse())

					valuesNode.values.Find(datatypes.Int(5), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "3":
					valuesNode.values.Find(datatypes.String("three"), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(3))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(0))
						g.Expect(len(idNode.ids[2])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				case "4":
					valuesNode.values.Find(datatypes.Int(4), testmodels.NoTypeRestrictions(g), func(key datatypes.EncapsulatedValue, item any) bool {
						called++
						idNode := item.(*threadsafeIDNode)
						g.Expect(len(idNode.ids)).To(Equal(2))
						g.Expect(len(idNode.ids[0])).To(Equal(0))
						g.Expect(len(idNode.ids[1])).To(Equal(1))

						return true
					})

					g.Expect(called).To(Equal(1))
				}

				return true
			}
		}
		g.Expect(associatedTree.keys.Find(datatypes.String("1"), testmodels.NoTypeRestrictions(g), onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("2"), testmodels.NoTypeRestrictions(g), onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("3"), testmodels.NoTypeRestrictions(g), onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.keys.Find(datatypes.String("4"), testmodels.NoTypeRestrictions(g), onFind("4"))).ToNot(HaveOccurred())
		g.Expect(foundCounter).To(Equal(4))
	})
}

func TestAssociatedTree_CreateOrFind_SingleKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one")}
	noOpOnCreate := func() any { return "found me" }
	noOpOnFind := func(item AssociatedKeyValues) {}

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
		onFind := func(item AssociatedKeyValues) {
			called = true
			g.Expect(item.Value()).To(Equal("found me"))
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

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one")}
		keyValues2 := datatypes.KeyValues{"1": datatypes.Int(5)}
		keyValues3 := datatypes.KeyValues{"3": datatypes.String("three")}

		createCount := 0
		onCreate := func() any {
			createCount++
			return true
		}

		id0, err := associatedTree.CreateOrFind(keyValues0, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		id1, err := associatedTree.CreateOrFind(keyValues1, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := associatedTree.CreateOrFind(keyValues2, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		id3, err := associatedTree.CreateOrFind(keyValues3, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id0).ToNot(And(Equal(id3), Equal(id2), Equal(id3)))
		g.Expect(id1).ToNot(And(Equal(id2), Equal(id3)))
		g.Expect(id2).ToNot(Equal(id3))
		g.Expect(createCount).To(Equal(4))
	})

	t.Run("It returns the ID for the saved item in the tree", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		var foundItem any
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItem = item
			return true
		}

		id, err := associatedTree.CreateOrFind(keyValues, noOpOnCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure that the value is whats actually saved as the tree's id
		foundItem = nil
		err = associatedTree.associatedIDs.Find(datatypes.String(id), testmodels.NoTypeRestrictions(g), onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundItem).ToNot(BeNil())
	})
}

func TestAssociatedTree_CreateOrFind_MultiKeyValue(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item AssociatedKeyValues) {}

	t.Run("it finds an item that already exists", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onFind := func(item AssociatedKeyValues) {
			called = true
			g.Expect(item.Value()).To(Equal("find me"))
		}

		associatedTree.CreateOrFind(keyValues, noOpOnCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues, noOpOnCreate, onFind)
		g.Expect(called).To(BeTrue())
	})

	t.Run("it properly saves multiple key values", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues0 := datatypes.KeyValues{"1": datatypes.Any(), "2": datatypes.Any()}
		keyValues1 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float64(3.0)}
		keyValues2 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(3.0)}
		keyValues3 := datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Float32(5.0)}
		keyValues4 := datatypes.KeyValues{"2": datatypes.String("two"), "3": datatypes.Float32(3.0)}

		createCount := 0
		onCreate := func() any {
			createCount++
			return createCount
		}

		associatedTree.CreateOrFind(keyValues0, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues1, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues2, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues3, onCreate, noOpOnFind)
		associatedTree.CreateOrFind(keyValues4, onCreate, noOpOnFind)
		g.Expect(createCount).To(Equal(5))

		found := false
		onFind := func(item AssociatedKeyValues) {
			found = true
			g.Expect(item.Value()).To(Equal(3))
		}

		associatedTree.CreateOrFind(keyValues2, onCreate, onFind)
		g.Expect(found).To(BeTrue())
	})
}
