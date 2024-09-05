package btreeonetomany

import (
	"fmt"
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	. "github.com/onsi/gomega"
)

func TestOneToManyTree_QueryAction(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context parameters", func(t *testing.T) {
		t.Run("It returns an error if the oneID is empty", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.QueryAction("", &queryassociatedaction.AssociatedActionQuery{}, nil)
			g.Expect(err).To(Equal(ErrorOneIDEmpty))
		})

		t.Run("It returns an error if the query is not valid", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.QueryAction("oneID", &queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: queryassociatedaction.SelectionKeyValues{
						"one": queryassociatedaction.ValueQuery{},
					},
				},
			}, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Selection.KeyValues[one].Comparison: unknown value ''"))
		})

		t.Run("It returns an error if the onQueryPagination is nil", func(t *testing.T) {
			tree := NewThreadSafe()

			err := tree.QueryAction("oneID", &queryassociatedaction.AssociatedActionQuery{}, nil)
			g.Expect(err).To(Equal(ErrorOnIterateNil))
		})
	})

	t.Run("It can query an item in the One Relation", func(t *testing.T) {
		tree := NewThreadSafe()

		// create 50 items all under the same one tree
		for i := 0; i < 50; i++ {
			tree.CreateOrFind("one name", datatypes.KeyValues{fmt.Sprintf("%d", i): datatypes.Int(i)}, func() any { return fmt.Sprintf("%d", i) }, func(oneToManyItem OneToManyItem) { panic("no") })
		}

		foundPagination := []string{}
		onPaginationQuery := func(paginationItem OneToManyItem) bool {
			foundPagination = append(foundPagination, paginationItem.Value().(string))
			return true
		}

		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"17": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}

		err := tree.QueryAction("one name", query, onPaginationQuery)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(foundPagination).To(Equal([]string{"17"}))
	})

	// TODO: error handling checks, but thats going to be reworked so ignore for now
}
