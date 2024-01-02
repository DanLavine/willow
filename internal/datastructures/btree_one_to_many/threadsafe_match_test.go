package btreeonetomany

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_MatchPermutations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchPermutations("", datatypes.KeyValues{}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the KeyValues are empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchPermutations("oneID", datatypes.KeyValues{}, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("keyValues cannot be empty"))
		})

		t.Run("It returns an error if the KeyValues are invalid", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchPermutations("oneID", datatypes.KeyValues{
				"bad key": datatypes.EncapsulatedValue{Type: datatypes.T_int, Data: "nope"},
			}, nil)
			g.Expect(err.Error()).To(ContainSubstring("EncapsulatedValue has an int data type, but the Value is a: string"))
		})

		t.Run("It returns an error if onPagination is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.MatchPermutations("oneID", datatypes.KeyValues{"key": datatypes.Int(2)}, nil)
			g.Expect(err).To(Equal(ErrorOnIterateNil))
		})
	})

	t.Run("It can match all items in the One Relation", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		g.Expect(tree.CreateWithID("one name", "assoic id 1", datatypes.KeyValues{"key1": datatypes.Int(1)}, func() any { return "1" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "assoic id 2", datatypes.KeyValues{"key2": datatypes.Int(2)}, func() any { return "2" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "assoic id 3", datatypes.KeyValues{"key3": datatypes.Int(3)}, func() any { return "3" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "assoic id 4", datatypes.KeyValues{"key1": datatypes.Int(1), "key2": datatypes.Int(2)}, func() any { return "4" })).ToNot(HaveOccurred())
		g.Expect(tree.CreateWithID("one name", "assoic id 5", datatypes.KeyValues{"key4": datatypes.Int(4)}, func() any { return "5" })).ToNot(HaveOccurred())

		foundPagination := []string{}
		onPaginationQuery := func(paginationItem OneToManyItem) bool {
			foundPagination = append(foundPagination, paginationItem.Value().(string))
			return true
		}

		matchKeys := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
			"key3": datatypes.Int(3),
		}

		err := tree.MatchPermutations("one name", matchKeys, onPaginationQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(foundPagination)).To(Equal(4))
		g.Expect(foundPagination).To(ContainElements([]string{"1", "2", "3", "4"}))
	})

	// TODO: error handling checks, but thats going to be reworked so ignore for now
}
