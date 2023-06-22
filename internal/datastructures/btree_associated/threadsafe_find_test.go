package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestCompositeTree_Find_ParameterChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	keyValuePairs := datatypes.StringMap{"one": datatypes.String("1")}
	onFindNoOp := func(item any) {}

	t.Run("it returns an error if the 'keyValuePairs' are nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Find(nil, onFindNoOp)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("keyValuePairs requires a length of at least 1"))
	})

	t.Run("it returns an error if the 'keyValuePairs' are empty", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Find(datatypes.StringMap{}, onFindNoOp)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("keyValuePairs requires a length of at least 1"))
	})

	t.Run("it returns an error if the 'onFind' is nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Find(keyValuePairs, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("onFind cannot be nil"))
	})
}

func TestCompositeTree_Find(t *testing.T) {
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

	t.Run("it doesn't run onFind if all key value pairs don't match", func(t *testing.T) {
		associatedTree := setup()

		found := false
		onfind := func(item any) {
			found = true
		}

		err := associatedTree.Find(datatypes.StringMap{"1": datatypes.Int(1)}, onfind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})

	t.Run("it can find a single key value that is found in the tree", func(t *testing.T) {
		associatedTree := setup()

		found := false
		onfind := func(item any) {
			found = true
			g.Expect(item.(*JoinTreeTester).Value).To(Equal("1"))
		}

		err := associatedTree.Find(datatypes.StringMap{"1": datatypes.String("other")}, onfind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeTrue())
	})

	t.Run("it can find a multi key value pair that is found in the tree", func(t *testing.T) {
		associatedTree := setup()

		found := false
		onfind := func(item any) {
			found = true
			g.Expect(item.(*JoinTreeTester).Value).To(Equal("2"))
		}

		err := associatedTree.Find(datatypes.StringMap{"1": datatypes.String("other"), "2": datatypes.String("foo")}, onfind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeTrue())
	})

	t.Run("it does not run onFind if a key is not found", func(t *testing.T) {
		associatedTree := setup()

		found := false
		onfind := func(item any) {
			found = true
		}

		err := associatedTree.Find(datatypes.StringMap{"not found": datatypes.String("not found")}, onfind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})
}
