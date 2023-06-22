package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the callback is nil", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Iterate(nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("callback is nil"))
	})

	t.Run("it doesn't run the callback when the tree is empty", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		called := false
		callback := func(value any) {
			called = true
		}

		g.Expect(associatedTree.Iterate(callback)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("it runs the callback for each key value pair", func(t *testing.T) {
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

		associatedTree := NewThreadSafe()

		g.Expect(associatedTree.CreateOrFind(keyValuePairs1, NewJoinTreeTester("first"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValuePairs2, NewJoinTreeTester("second"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValuePairs3, NewJoinTreeTester("third"), OnFindTest)).ToNot(HaveOccurred())

		values := []string{}
		iterateCallback := func(value any) {
			values = append(values, value.(*JoinTreeTester).Value)
		}

		g.Expect(associatedTree.Iterate(iterateCallback)).ToNot(HaveOccurred())
		g.Expect(values).To(ContainElements([]string{"first", "second", "third"}))
	})
}
