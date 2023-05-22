package compositetree

import (
	"fmt"
	"testing"

	. "github.com/DanLavine/willow/internal/datastructures/composite_tree/testhelpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestCompositeTree_FindInclusive(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("when the query is empty", func(t *testing.T) {
		t.Run("returns nothing on an empty tree", func(t *testing.T) {
			compositeTree := New()

			queryResults := compositeTree.FindInclusive(nil, nil)
			g.Expect(queryResults).To(BeNil())
		})

		t.Run("it returns the empty key value pair if it exists", func(t *testing.T) {
			compositeTree := New()

			item, err := compositeTree.CreateOrFind(nil, NewJoinTreeTester("empty"), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item).ToNot(BeNil())

			queryResults := compositeTree.FindInclusive(nil, nil)
			g.Expect(len(queryResults)).To(Equal(1))

			compositeTreeTesters := item.([]*JoinTreeTester)
			g.Expect(compositeTreeTesters[0].Value).To(Equal("other"))
			g.Expect(compositeTreeTesters[0].OnFindCount).To(Equal(0))
		})
	})

	t.Run("when the try is populated", func(t *testing.T) {
		setup := func(g *GomegaWithT) *compositeTree {
			compositeTree := New()

			// create 100 keys for the index table of 1
			for i := 0; i < 10; i++ {
				for j := 0; j < 10; j++ {
					keyValues := map[datatypes.String]datatypes.String{
						datatypes.String(fmt.Sprintf("%d", j)): datatypes.String(fmt.Sprintf("%d", i)),
					}

					item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester(fmt.Sprintf("%d%d", i, j)), nil)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(item).ToNot(BeNil())
				}
			}

			// create 100 keys for the index table of 2
			for i := 0; i < 10; i++ {
				for j := 0; j < 10; j++ {
					// create 10 value for each index
					keyValues := map[datatypes.String]datatypes.String{
						datatypes.String(fmt.Sprintf("%d", j)): datatypes.String(fmt.Sprintf("%d", i)),
						datatypes.String(fmt.Sprintf("%d", i)): datatypes.String(fmt.Sprintf("%d", j)),
					}

					item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester(fmt.Sprintf("%d%d", i, j)), nil)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(item).ToNot(BeNil())
				}
			}

			// create 100 keys for the index table of 3
			for i := 0; i < 10; i++ {
				for j := 0; j < 10; j++ {
					// create 10 value for each index
					keyValues := map[datatypes.String]datatypes.String{
						datatypes.String(fmt.Sprintf("a%d", j)): datatypes.String(fmt.Sprintf("%d", i)),
						datatypes.String(fmt.Sprintf("b%d", i)): datatypes.String(fmt.Sprintf("%d", j)),
						datatypes.String(fmt.Sprintf("c%d", i)): datatypes.String(fmt.Sprintf("%d", j)),
					}

					item, err := compositeTree.CreateOrFind(keyValues, NewJoinTreeTester(fmt.Sprintf("%d%d", i, j)), nil)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(item).ToNot(BeNil())
				}
			}

			return compositeTree
		}

		t.Run("returns everything in the tree", func(t *testing.T) {
			compositeTree := setup(g)

			queryResults := compositeTree.FindInclusive(nil, nil)
			g.Expect(len(queryResults)).To(Equal(300))
		})
	})
}
