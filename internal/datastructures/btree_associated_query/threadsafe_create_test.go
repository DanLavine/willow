package btreeassociatedquery

/*
import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

var exists = true

func TestAssociatedQueryTree_Create_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	onCreate := func() any { return true }

	t.Run("It returns an error if the query is not valid", func(t *testing.T) {
		associatedQueryTree := NewThreadSafe()

		badQuery := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"bad": datatypes.Value{},
				},
			},
		}

		id, err := associatedQueryTree.Create(badQuery, onCreate)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("KeyValueSelection.KeyValues[bad]: Requires an Exists or Value check"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("It returns an error if onCreate callback is nil", func(t *testing.T) {
		associatedQueryTree := NewThreadSafe()

		goodQuery := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{
					"bad": datatypes.Value{Exists: &exists},
				},
			},
		}

		id, err := associatedQueryTree.Create(goodQuery, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be nil"))
		g.Expect(id).To(Equal(""))
	})
}

func TestAssociatedQueryTree_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	onCreate := func() any { return true }

	t.Run("Describe adding a single valid query to the tree", func(t *testing.T) {
		t.Run("Context Exists == true", func(t *testing.T) {
			t.Run("It can properly match agains key values", func(t *testing.T) {
				associatedQueryTree := NewThreadSafe()

				goodQuery := datatypes.AssociatedKeyValuesQuery{
					KeyValueSelection: &datatypes.KeyValueSelection{
						KeyValues: map[string]datatypes.Value{
							"good": datatypes.Value{Exists: &exists},
						},
					},
				}

				id, err := associatedQueryTree.Create(goodQuery, onCreate)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("KeyValueSelection.KeyValues[bad]: Requires an Exists or Value check"))
				g.Expect(id).To(Equal(""))
			})
		})
	})
}
*/
