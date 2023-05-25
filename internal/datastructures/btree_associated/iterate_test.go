package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/internal/datastructures/btree_associated/testhelpers"
	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it doesn't run the callback when the tree is empty", func(t *testing.T) {
		associatedTree := New()

		called := false
		callback := func(key datatypes.CompareType, value any) {
			called = true
		}

		associatedTree.Iterate(callback)
		g.Expect(called).To(BeFalse())
	})

	t.Run("it runs the callback for each key value pair", func(t *testing.T) {
		keyValuePairs1 := datatypes.StringMap{
			"1": "one",
		}
		keyValuePairs2 := datatypes.StringMap{
			"1": "one",
			"2": "two",
		}
		keyValuePairs3 := datatypes.StringMap{
			"1": "one",
			"2": "two",
			"3": "three",
			"5": "five",
		}

		associatedTree := New()

		_, err := associatedTree.CreateOrFind(keyValuePairs1, NewJoinTreeTester("first"), OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())

		_, err = associatedTree.CreateOrFind(keyValuePairs2, NewJoinTreeTester("second"), OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())

		_, err = associatedTree.CreateOrFind(keyValuePairs3, NewJoinTreeTester("third"), OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())

		values := []string{}
		iterateCallback := func(key datatypes.CompareType, value any) {
			values = append(values, value.(*JoinTreeTester).Value)
		}

		associatedTree.Iterate(iterateCallback)
		g.Expect(values).To(ContainElements([]string{"first", "second", "third"}))
	})
}
