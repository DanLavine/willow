package btreeassociated

// TODO revisit this
//
//import (
//	"testing"
//
//	. "github.com/DanLavine/willow/internal/datastructures/composite_tree/testhelpers"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//	v1 "github.com/DanLavine/willow/pkg/models/v1"
//	. "github.com/onsi/gomega"
//)
//
//func TestCompositeTree_FindInclusive(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("when the tree is empty", func(t *testing.T) {
//		t.Run("it returns nothing, no matter the query", func(t *testing.T) {
//			compositeTree := New()
//
//			items := compositeTree.FindInclusive(nil, nil)
//			g.Expect(items).To(BeNil())
//		})
//	})
//
//	t.Run("context when the tree is populated", func(t *testing.T) {
//		keyValues1 := map[datatypes.String]datatypes.String{
//			"1": "other",
//		}
//		keyValues1a := map[datatypes.String]datatypes.String{
//			"1": "a",
//		}
//		keyValues2 := map[datatypes.String]datatypes.String{
//			"1": "other",
//			"2": "foo",
//		}
//		keyValues2a := map[datatypes.String]datatypes.String{
//			"1": "other",
//			"2": "a",
//		}
//
//		setup := func() *compositeTree {
//			compositeTree := New()
//
//			_, err := compositeTree.CreateOrFind(keyValues1, NewJoinTreeTester("1"), nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = compositeTree.CreateOrFind(keyValues1a, NewJoinTreeTester("1a"), nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = compositeTree.CreateOrFind(keyValues2, NewJoinTreeTester("2"), nil)
//			g.Expect(err).ToNot(HaveOccurred())
//			_, err = compositeTree.CreateOrFind(keyValues2a, NewJoinTreeTester("2a"), nil)
//			g.Expect(err).ToNot(HaveOccurred())
//
//			return compositeTree
//		}
//
//		t.Run("it returns everything, if the query is empty", func(t *testing.T) {
//			compositeTree := setup()
//
//			items := compositeTree.FindInclusive(nil, nil)
//			g.Expect(len(items)).To(Equal(4))
//			g.Expect(items[0]).To(BeAssignableToTypeOf(&JoinTreeTester{}))
//			g.Expect(items[1]).To(BeAssignableToTypeOf(&JoinTreeTester{}))
//			g.Expect(items[2]).To(BeAssignableToTypeOf(&JoinTreeTester{}))
//			g.Expect(items[3]).To(BeAssignableToTypeOf(&JoinTreeTester{}))
//
//			g.Expect(items[0].(*JoinTreeTester).Value).To(Equal("1"))
//			g.Expect(items[1].(*JoinTreeTester).Value).To(Equal("1a"))
//			g.Expect(items[2].(*JoinTreeTester).Value).To(Equal("2a")) // like this because of how we iterate through id tree
//			g.Expect(items[3].(*JoinTreeTester).Value).To(Equal("2"))
//		})
//
//		t.Run("context, when using an exact where", func(t *testing.T) {
//			t.Run("it returns nothing if the exact match cannot be found", func(t *testing.T) {
//				compositeTree := setup()
//
//				query := &v1.InclusiveWhere{ExactWhere: v1.KeyValues{"not found": "nope"}}
//				g.Expect(query.Validate()).ToNot(HaveOccurred())
//
//				items := compositeTree.FindInclusive(query, nil)
//				g.Expect(items).To(BeNil())
//			})
//
//			t.Run("it only returns the keys that match exactly", func(t *testing.T) {
//				compositeTree := setup()
//
//				query := &v1.InclusiveWhere{ExactWhere: v1.KeyValues{"1": "other"}}
//				g.Expect(query.Validate()).ToNot(HaveOccurred())
//
//				items := compositeTree.FindInclusive(query, nil)
//				g.Expect(len(items)).To(Equal(1))
//				g.Expect(items[0].(*JoinTreeTester).Value).To(Equal("1"))
//			})
//
//			t.Run("when using multiple key values", func(t *testing.T) {
//				t.Run("it returns nothing if any key values don't match", func(t *testing.T) {
//					compositeTree := setup()
//
//					query := &v1.InclusiveWhere{ExactWhere: v1.KeyValues{"1": "other", "2": "not found"}}
//					g.Expect(query.Validate()).ToNot(HaveOccurred())
//
//					items := compositeTree.FindInclusive(query, nil)
//					g.Expect(items).To(BeNil())
//				})
//
//				t.Run("it only returns the keys that match exactly", func(t *testing.T) {
//					compositeTree := setup()
//
//					query := &v1.InclusiveWhere{ExactWhere: v1.KeyValues{"1": "other", "2": "a"}}
//					g.Expect(query.Validate()).ToNot(HaveOccurred())
//
//					items := compositeTree.FindInclusive(query, nil)
//					g.Expect(len(items)).To(Equal(1))
//					g.Expect(items[0].(*JoinTreeTester).Value).To(Equal("2a"))
//				})
//			})
//		})
//	})
//}
