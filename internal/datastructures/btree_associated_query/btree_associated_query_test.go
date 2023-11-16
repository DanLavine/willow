package btreeassociatedquery

/*
import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestAssociatedqueryTree_convertAssociatedKeyValuesQuery(t *testing.T) {
	g := NewGomegaWithT(t)

	exists := true
	notExist := false

	t.Run("It can convert a single value lookup", func(t *testing.T) {
		t.Run("Context exists", func(t *testing.T) {
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key1": datatypes.Value{Exists: &exists},
					},
				},
			}

			insertableKeyValues := convertAssociatedKeyValuesQuery(query)
			g.Expect(len(insertableKeyValues)).To(Equal(1))
		})

		t.Run("Context not exists", func(t *testing.T) {
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key1": datatypes.Value{Exists: &notExist},
					},
				},
			}

			insertableKeyValues := convertAssociatedKeyValuesQuery(query)
			g.Expect(len(insertableKeyValues)).To(Equal(1))
		})
	})

}
*/
