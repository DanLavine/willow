package btreeonetomany

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_Query(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.Query("", datatypes.AssociatedKeyValuesQuery{}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the query is not valid", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.Query("oneID", datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{}, // cannot be mepty
					},
				},
			}, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("query error: KeyValueSelection.KeyValues[one]: Requires an Exists or Value check"))
		})

		t.Run("It returns an error if the onQueryPagination is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.Query("oneID", datatypes.AssociatedKeyValuesQuery{}, nil)
			g.Expect(err).To(Equal(ErrorOnIterateNil))
		})
	})

	t.Run("It can query an item in the One Relation", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			g.Expect(tree.CreateWithID("one name", fmt.Sprintf("assocID%d", i), datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return fmt.Sprintf("%d", i) })).ToNot(HaveOccurred())
		}

		foundPagination := []string{}
		onPaginationQuery := func(paginationItem OneToManyItem) bool {
			foundPagination = append(foundPagination, paginationItem.Value().(string))
			return true
		}

		trueCheck := true

		query := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"17": datatypes.Value{Exists: &trueCheck},
				},
			},
		}

		err := tree.Query("one name", query, onPaginationQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundPagination).To(Equal([]string{"17"}))
	})

	// TODO: error handling checks, but thats going to be reworked so ignore for now
}
