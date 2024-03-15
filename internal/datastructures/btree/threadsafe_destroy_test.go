package btree

import (
	"fmt"
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestBTree_Destroy_ParamChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the key is not valid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Destroy(datatypes.EncapsulatedValue{}, nil)
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("It accepts a nil canDelete callback", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Destroy(datatypes.String("1"), nil)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestBTree_Destroy(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypeRestriction := v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}
	g.Expect(noTypeRestriction.Validate()).ToNot(HaveOccurred())

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 100; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("It returns an error if the key is already destroying", func(t *testing.T) {
		bTree := setupTree(g)

		counter := 0
		destroying := make(chan struct{})
		go func() {
			_ = bTree.Destroy(datatypes.Int(17), func(_ datatypes.EncapsulatedValue, item any) bool {
				if counter == 0 {
					destroying <- struct{}{}
					<-destroying
					counter++
				}
				return true
			})
		}()

		g.Eventually(destroying).Should(Receive())

		err := bTree.Destroy(datatypes.Int(17), nil)
		g.Expect(err).To(Equal(ErrorKeyDestroying))

		destroying <- struct{}{}
	})

	t.Run("It returns an error if the tree is already destroying", func(t *testing.T) {
		bTree := setupTree(g)

		counter := 0
		destroying := make(chan struct{})
		go func() {
			_ = bTree.DestroyAll(func(_ datatypes.EncapsulatedValue, item any) bool {
				if counter == 0 {
					destroying <- struct{}{}
					<-destroying
					counter++
				}
				return true
			})
		}()

		g.Eventually(destroying).Should(Receive())

		err := bTree.Destroy(datatypes.Int(17), nil)
		g.Expect(err).To(Equal(ErrorTreeDestroying))

		destroying <- struct{}{}
	})

	t.Run("It can delete the value in the btree with a nil canDelete", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Destroy(datatypes.Int(17), nil)
		g.Expect(err).ToNot(HaveOccurred())

		keys := []datatypes.EncapsulatedValue{}
		iterate := func(key datatypes.EncapsulatedValue, _ any) bool {
			keys = append(keys, key)
			return true
		}

		err = bTree.Find(datatypes.Any(), noTypeRestriction, iterate)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(keys)).To(Equal(99))
		g.Expect(keys).ToNot(ContainElement(datatypes.Int(17)))
	})

	t.Run("It can delete value in the btree with a canDelete set", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Destroy(datatypes.Int(17), func(_ datatypes.EncapsulatedValue, _ any) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())

		keys := []datatypes.EncapsulatedValue{}
		iterate := func(key datatypes.EncapsulatedValue, _ any) bool {
			keys = append(keys, key)
			return true
		}

		err = bTree.Find(datatypes.Any(), noTypeRestriction, iterate)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(keys)).To(Equal(99))
		g.Expect(keys).ToNot(ContainElement(datatypes.Int(17)))
	})

	t.Run("Context when canDelete returns false", func(t *testing.T) {
		t.Run("It does not delete the item", func(t *testing.T) {
			bTree := setupTree(g)

			err := bTree.Destroy(datatypes.Int(17), func(_ datatypes.EncapsulatedValue, _ any) bool { return false })
			g.Expect(err).ToNot(HaveOccurred())

			keys := []datatypes.EncapsulatedValue{}
			iterate := func(key datatypes.EncapsulatedValue, _ any) bool {
				keys = append(keys, key)
				return true
			}

			err = bTree.Find(datatypes.Any(), noTypeRestriction, iterate)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(keys)).To(Equal(100))
			g.Expect(keys).To(ContainElement(datatypes.Int(17)))
		})
	})
}

func TestBTree_DestroyAll_ParamChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It accepts a nil canDelete callback", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.DestroyAll(nil)
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestBTree_DestroyAll(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypeRestriction := v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}
	g.Expect(noTypeRestriction.Validate()).ToNot(HaveOccurred())

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 100; i++ {
			g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
		}

		return bTree
	}

	t.Run("It returns an error if the tree is already destroying", func(t *testing.T) {
		bTree := setupTree(g)

		counter := 0
		destroying := make(chan struct{})
		go func() {
			_ = bTree.DestroyAll(func(_ datatypes.EncapsulatedValue, item any) bool {
				if counter == 0 {
					destroying <- struct{}{}
					<-destroying
					counter++
				}
				return true
			})
		}()

		g.Eventually(destroying).Should(Receive())

		err := bTree.DestroyAll(nil)
		g.Expect(err).To(Equal(ErrorTreeDestroying))

		destroying <- struct{}{}
	})

	t.Run("It can delete all values in the btree with a nil canDelete", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.DestroyAll(nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.Empty()).To(BeTrue())
	})

	t.Run("It can delete all values in the btree with canDelete set", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.DestroyAll(func(_ datatypes.EncapsulatedValue, _ any) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.Empty()).To(BeTrue())
	})

	t.Run("Context when canDelete returns false", func(t *testing.T) {
		t.Run("It keeps the tree properly balanced", func(t *testing.T) {
			bTree := setupTree(g)

			// remove 25 items
			counter := 0
			canDelete := func(_ datatypes.EncapsulatedValue, item any) bool {
				if counter < 25 {
					counter++
					return true
				}

				return false
			}

			err := bTree.DestroyAll(canDelete)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(bTree.Empty()).To(BeFalse())
			validateThreadSafeTree(g, bTree.root)

			// ensure we still have the proper number of items in the tree
			iterateCounter := 0
			iterate := func(key datatypes.EncapsulatedValue, _ any) bool {
				iterateCounter++
				return true
			}

			err = bTree.Find(datatypes.Any(), noTypeRestriction, iterate)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(iterateCounter).To(Equal(75))
		})
	})
}
