package compositetree

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/DanLavine/willow/internal/datastructures/composite_tree/testhelpers"
	. "github.com/onsi/gomega"
)

func TestCompositeTree_CreateOrFind(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		compositeTree := New()

		item, err := compositeTree.CreateOrFind(nil, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValues cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("it returns an error with empty keyValues", func(t *testing.T) {
		compositeTree := New()

		item, err := compositeTree.CreateOrFind(map[datatypes.String]datatypes.String{}, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValues cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
		compositeTree := New()

		item, err := compositeTree.CreateOrFind(map[datatypes.String]datatypes.String{"1": "one"}, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be empty"))
		g.Expect(item).To(BeNil())
	})

	t.Run("single item test", func(t *testing.T) {
		keyValues := map[datatypes.String]datatypes.String{
			"1": "one",
		}
		t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
			compositeTree := New()

			item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			compositeTreeTester := item.(*JoinTreeTester)
			g.Expect(compositeTreeTester.Value).To(Equal("other"))
		})

		t.Run("it calls onFind if the item already exists", func(t *testing.T) {
			compositeTree := New()

			// create item
			item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			compositeTreeTester := item.(*JoinTreeTester)
			g.Expect(compositeTreeTester.Value).To(Equal("other"))
			g.Expect(compositeTreeTester.OnFindCount).To(Equal(0))

			// find item
			item, err = compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())

			compositeTreeTester = item.(*JoinTreeTester)
			g.Expect(compositeTreeTester.Value).To(Equal("other"))
			g.Expect(compositeTreeTester.OnFindCount).To(Equal(1))
		})
	})

	t.Run("multiple item test", func(t *testing.T) {
		keyValues := map[datatypes.String]datatypes.String{
			"1": "one",
			"2": "two",
			"3": "three",
			"4": "four",
		}

		setup := func() *compositeTree {
			compositeTree := New()

			item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("setup data"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			return compositeTree
		}

		t.Run("it can create creates a value with multiple key value pairs", func(t *testing.T) {
			compositeTree := New()

			item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("setup data"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())
		})

		t.Run("it calls onFind if the item already exists", func(t *testing.T) {
			compositeTree := setup()

			item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester("other"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())

			compositeTreeTester := item.(*JoinTreeTester)
			g.Expect(compositeTreeTester.Value).To(Equal("setup data"))
			g.Expect(compositeTreeTester.OnFindCount).To(Equal(1))
		})

		t.Run("it can create new values if the keyValues are different", func(t *testing.T) {
			compositeTree := setup()
			keyValues1 := map[datatypes.String]datatypes.String{
				"1": "one",
			}
			keyValues2 := map[datatypes.String]datatypes.String{
				"1": "one",
				"2": "two",
			}
			keyValues3 := map[datatypes.String]datatypes.String{
				"1": "one",
				"2": "two",
				"3": "three",
				"5": "five",
			}

			item1, err := compositeTree.CreateOrFind(keyValues1, NewJoinTreeTester("first"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			compositeTreeTester1 := item1.(*JoinTreeTester)
			g.Expect(compositeTreeTester1.Value).To(Equal("first"))

			item2, err := compositeTree.CreateOrFind(keyValues2, NewJoinTreeTester("second"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			compositeTreeTester2 := item2.(*JoinTreeTester)
			g.Expect(compositeTreeTester2.Value).To(Equal("second"))

			item3, err := compositeTree.CreateOrFind(keyValues3, NewJoinTreeTester("third"), OnFindTest)
			g.Expect(err).ToNot(HaveOccurred())
			compositeTreeTester3 := item3.(*JoinTreeTester)
			g.Expect(compositeTreeTester3.Value).To(Equal("third"))
		})
	})
}
