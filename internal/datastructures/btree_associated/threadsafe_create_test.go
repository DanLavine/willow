package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_CreateOrFind_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.StringMap{"1": datatypes.Int(1)}
	onCreate := func() any { return true }
	onFind := func(item any) {}

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateOrFind(nil, onCreate, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateOrFind(keys, nil, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be nil"))
	})

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateOrFind(keys, onCreate, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
	})
}

func TestAssociatedTree_CreateOrFind_SingleItemTest(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValues := datatypes.StringMap{"1": datatypes.String("one")}
	noOpOnCreate := func() any { return true }
	noOpOnFind := func(item any) {}

	t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return true
		}

		err := associatedTree.CreateOrFind(keyValues, onCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
	})

	t.Run("it calls onFind if the item already exists", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		findCalled := false
		onFind := func(item any) {
			jtt := item.(*JoinTreeTester)
			g.Expect(jtt.Value).To(Equal("other"))
			findCalled = true
		}

		// create item
		err := associatedTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		// find item
		err = associatedTree.CreateOrFind(keyValues, noOpOnCreate, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(findCalled).To(BeTrue())
	})
}

func TestAssociatedTree_CreateOrFind_MultipleItemTest(t *testing.T) {
	g := NewGomegaWithT(t)

	testKeyValues := datatypes.StringMap{
		"1": datatypes.String("one"),
		"2": datatypes.String("two"),
		"3": datatypes.String("three"),
		"4": datatypes.String("four"),
	}
	noOpOnCreate := func() any { return true }
	noOpOnFind := func(item any) {}

	setup := func() *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()
		g.Expect(associatedTree.CreateOrFind(testKeyValues, NewJoinTreeTester("setup data"), noOpOnFind)).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("it can create creates a value with multiple key value pairs", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		onCreate := func() any {
			called = true
			return true
		}

		err := associatedTree.CreateOrFind(testKeyValues, onCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
	})

	t.Run("it calls onFind if the item already exists", func(t *testing.T) {
		associatedTree := setup()

		called := false
		onFind := func(item any) {
			jtt := item.(*JoinTreeTester)
			g.Expect(jtt.Value).To(Equal("setup data"))
			called = true
		}

		err := associatedTree.CreateOrFind(testKeyValues, noOpOnCreate, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())
	})

	t.Run("it can create new values if the keyValues are different", func(t *testing.T) {
		associatedTree := setup()
		keyValuePairs1 := datatypes.StringMap{
			"1": datatypes.String("one"),
		}
		keyValuePairs2 := datatypes.StringMap{
			"1": datatypes.String("one"),
			"2": datatypes.String("two"),
		}
		keyValuePairs3 := datatypes.StringMap{
			"1": datatypes.String("one"),
			"2": datatypes.String("two"),
			"3": datatypes.String("three"),
			"5": datatypes.String("five"),
		}

		counter := 0
		onCreate := func() any {
			counter++
			return true
		}

		err := associatedTree.CreateOrFind(keyValuePairs1, onCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(1))

		err = associatedTree.CreateOrFind(keyValuePairs2, onCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(2))

		err = associatedTree.CreateOrFind(keyValuePairs3, onCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(3))
	})

	t.Run("it records multiple ids for a key value pair if there are multiple groupings", func(t *testing.T) {
		associatedTree := setup()

		keyValuePairs1 := datatypes.StringMap{
			"a": datatypes.String("a"),
			"b": datatypes.String("b"),
			"c": datatypes.String("c"),
		}
		keyValuePairs2 := datatypes.StringMap{
			"b": datatypes.String("b"),
			"c": datatypes.String("c"),
			"d": datatypes.String("d"),
		}

		err := associatedTree.CreateOrFind(keyValuePairs1, noOpOnCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())

		err = associatedTree.CreateOrFind(keyValuePairs2, noOpOnCreate, OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())

		calledFindIDHolders := false
		onFindIDHolders := func(item any) {
			idHolder := item.(*idHolder)
			g.Expect(len(idHolder.IDs)).To(Equal(2))
			calledFindIDHolders = true
		}

		calledFindValues := false
		onFindValues := func(item any) {
			keyValues := item.(*keyValues)
			keyValues.values.Find(datatypes.String("b"), onFindIDHolders)
			calledFindValues = true
		}

		calledFindKeys := false
		onFindKeys := func(item any) {
			keyValues := item.(*keyValues)
			keyValues.values.Find(datatypes.String("b"), onFindValues)
			calledFindKeys = true
		}

		g.Expect(associatedTree.groupedKeyValueAssociation.Find(datatypes.Int(3), onFindKeys)).ToNot(HaveOccurred())
		g.Expect(calledFindKeys).To(BeTrue())
		g.Expect(calledFindValues).To(BeTrue())
		g.Expect(calledFindIDHolders).To(BeTrue())
	})
}

func TestAssociatedTree_CreateOrFind_CreateReturnsNil(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValuePairs := datatypes.StringMap{
		"1": datatypes.String("one"),
		"2": datatypes.String("two"),
		"3": datatypes.String("three"),
		"4": datatypes.String("four"),
	}
	noOpOnFind := func(item any) {}

	setup := func() *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		err := associatedTree.CreateOrFind(keyValuePairs, NewJoinTreeTester("setup data"), noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		return associatedTree
	}

	t.Run("it cleans up extra key value pairs that might have been created", func(t *testing.T) {
		associatedTree := setup()
		newKeyValues := datatypes.StringMap{
			"1": datatypes.String("other"),
			"2": datatypes.String("foo"),
			"3": datatypes.String("three"),
			"5": datatypes.String("this shouldn't be there"),
		}

		called := false
		failedCreate := func() any {
			called = true
			return nil
		}

		err := associatedTree.CreateOrFind(newKeyValues, failedCreate, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(called).To(BeTrue())

		totalfound := 0

		checkAssociatedValues := func(key string) func(item any) {
			return func(item any) {
				shouldFind := false
				onShouldFind := func(item any) {
					totalfound++
					shouldFind = true
				}

				shouldNotFind := false
				onShouldNotFind := func(item any) {
					shouldNotFind = true
				}

				keyValues := item.(*keyValues)
				switch key {
				case "1":
					g.Expect(keyValues.values.Find(datatypes.String("one"), onShouldFind)).ToNot(HaveOccurred())
					g.Expect(keyValues.values.Find(datatypes.String("other"), onShouldNotFind)).ToNot(HaveOccurred())
				case "2":
					g.Expect(keyValues.values.Find(datatypes.String("two"), onShouldFind)).ToNot(HaveOccurred())
					g.Expect(keyValues.values.Find(datatypes.String("foo"), onShouldNotFind)).ToNot(HaveOccurred())
				case "3":
					g.Expect(keyValues.values.Find(datatypes.String("three"), onShouldFind)).ToNot(HaveOccurred())
				case "4":
					g.Expect(keyValues.values.Find(datatypes.String("four"), onShouldFind)).ToNot(HaveOccurred())
				}

				g.Expect(shouldFind).To(BeTrue(), key)
				g.Expect(shouldNotFind).To(BeFalse(), key)
			}
		}

		checkAssociatedKeys := func(item any) {
			keyValues := item.(*keyValues)
			g.Expect(keyValues.values.Find(datatypes.String("1"), checkAssociatedValues("1"))).ToNot(HaveOccurred())
			g.Expect(keyValues.values.Find(datatypes.String("2"), checkAssociatedValues("2"))).ToNot(HaveOccurred())
			g.Expect(keyValues.values.Find(datatypes.String("3"), checkAssociatedValues("3"))).ToNot(HaveOccurred())
			g.Expect(keyValues.values.Find(datatypes.String("4"), checkAssociatedValues("4"))).ToNot(HaveOccurred())

			found := false
			shoudNotFind := func(item any) {
				found = true
			}
			g.Expect(keyValues.values.Find(datatypes.String("5"), shoudNotFind)).ToNot(HaveOccurred())
			g.Expect(found).To(BeFalse())
		}

		g.Expect(associatedTree.groupedKeyValueAssociation.Find(datatypes.Int(4), checkAssociatedKeys)).ToNot(HaveOccurred())
		g.Expect(totalfound).To(Equal(4))
	})
}
