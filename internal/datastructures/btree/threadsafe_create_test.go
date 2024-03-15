package btree

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestBTree_Create_ParameterChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	onCreate := func() any { return "ok" }

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Create(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, onCreate)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if onCreate is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Create(Key1, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate callback cannot be nil"))
	})
}

func TestBTree_Create_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key2, NewBTreeTester("2"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("adding the same key multiple times returns an error", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())

		err = bTree.Create(Key1, NewBTreeTester("2"))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorKeyAlreadyExists))

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
	})

	t.Run("inserts the values in the proper nodeSize, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key2, NewBTreeTester("2"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("it splits the node when adding a left pivot value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key2, NewBTreeTester("2"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key3, NewBTreeTester("3"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})

	t.Run("it splits the node when adding a middle pivot value value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key3, NewBTreeTester("3"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key2, NewBTreeTester("2"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})

	t.Run("it splits the node when adding a right pivot value value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key2, NewBTreeTester("2"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key1, NewBTreeTester("1"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key3, NewBTreeTester("3"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})
}

func TestBTree_Create_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key10, NewBTreeTester("10"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key20, NewBTreeTester("20"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key30, NewBTreeTester("30"))).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("it returns an error if the left child tree already exists", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Create(Key10, NewBTreeTester("5"))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorKeyAlreadyExists))
	})

	t.Run("it returns an error if the right child tree already exists", func(t *testing.T) {
		bTree := setupTree(g)

		err := bTree.Create(Key30, NewBTreeTester("5"))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorKeyAlreadyExists))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  5,10   30
	 */
	t.Run("can add a new entry on the leftChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key5, NewBTreeTester("5"))).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key5))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("5"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(leftChild.keyValues[1].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10     25,30
	 */
	t.Run("can add a new entry on the rightChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key25, NewBTreeTester("25"))).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key25))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("25"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10,15   30
	 */
	t.Run("can add a new entry on the leftChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key15, NewBTreeTester("15"))).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(leftChild.keyValues[1].key).To(Equal(Key15))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("15"))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,35
	 */
	t.Run("can add a new entry on the rightChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key35, NewBTreeTester("35"))).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(Key35))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("35"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,T_any
	 */
	t.Run("can add a special T_any node always to the rightmost position", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("5"))).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(1))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(datatypes.Any()))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("5"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("it returns an error if the T_any node already exists", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("5"))).ToNot(HaveOccurred())

		err := bTree.Create(datatypes.Any(), NewBTreeTester("5"))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(ErrorKeyAlreadyExists))
	})
}

func TestBTree_Create_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("leftChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.Create(Key30, NewBTreeTester("30"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key10, NewBTreeTester("10"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key20, NewBTreeTester("20"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key15, NewBTreeTester("15"))).ToNot(HaveOccurred())

			return bTree
		}

		// generate a tree of
		/*
		 *      10,20
		 *    /   |   \
		 *   8    15  30
		 */
		t.Run("adding the leftChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key8, NewBTreeTester("8"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key10))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key8))
			g.Expect(child1.keyValues[0].value.(*BTreeTester).Value).To(Equal("8"))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key12, NewBTreeTester("12"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key12))
			g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("12"))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key17, NewBTreeTester("17"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key15))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key17))
			g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("17"))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.Create(Key30, NewBTreeTester("20"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key10, NewBTreeTester("30"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key20, NewBTreeTester("10"))).ToNot(HaveOccurred())
			g.Expect(bTree.Create(Key35, NewBTreeTester("35"))).ToNot(HaveOccurred())

			return bTree
		}

		// generate a tree of
		/*
		 *      20,30
		 *    /   |  \
		 *   10   25  35
		 */
		t.Run("adding the rightChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key25, NewBTreeTester("25"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key30))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key25))
			g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("25"))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key32, NewBTreeTester("32"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key32))
			g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("32"))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.Create(Key37, NewBTreeTester("37"))).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key35))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key37))
			g.Expect(child3.keyValues[0].value.(*BTreeTester).Value).To(Equal("37"))
		})
	})
}

func TestBTree_Create_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.Create(Key10, NewBTreeTester("10"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key20, NewBTreeTester("20"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key30, NewBTreeTester("30"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key40, NewBTreeTester("40"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key50, NewBTreeTester("50"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key5, NewBTreeTester("5"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key25, NewBTreeTester("25"))).ToNot(HaveOccurred())
		g.Expect(bTree.Create(Key45, NewBTreeTester("45"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key40))

		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		g.Expect(child1.keyValues[1].key).To(Equal(Key10))

		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.keyValues[0].key).To(Equal(Key25))
		g.Expect(child2.keyValues[1].key).To(Equal(Key30))

		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.keyValues[0].key).To(Equal(Key45))
		g.Expect(child3.keyValues[1].key).To(Equal(Key50))

		return bTree
	}

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        5          40
	 *      /  \      /      \
	 *     0   10   25,30  45,50
	 */
	t.Run("adding the leftChild[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key0, NewBTreeTester("0"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key5))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key0))
		g.Expect(gchild1.keyValues[0].value.(*BTreeTester).Value).To(Equal("0"))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        7          40
	 *      /  \      /      \
	 *     5   10   25,30  45,50
	 */
	t.Run("adding the leftChild[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key7, NewBTreeTester("7"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key7))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        7          40
	 *      /  \      /      \
	 *     5   10   25,30  45,50
	 */
	t.Run("adding the leftChild[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key12, NewBTreeTester("12"))).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key10))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key12))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             25
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  22   30   45,50
	 */
	t.Run("adding the middleChild[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key22, NewBTreeTester("22"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key22))
		g.Expect(gchild2.keyValues[0].value.(*BTreeTester).Value).To(Equal("22"))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             27
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  25   30   45,50
	 */
	t.Run("adding the middleChild[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key27, NewBTreeTester("27"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key27))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             30
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  25   32   45,50
	 */
	t.Run("adding the middleChild[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key32, NewBTreeTester("32"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key30))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key32))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         45
	 *      /    \      /   \
	 *    5,10  25,30  42   50
	 */
	t.Run("adding the right[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key42, NewBTreeTester("42"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key45))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key42))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         47
	 *      /    \      /   \
	 *    5,10  25,30  45   50
	 */
	t.Run("adding the right[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key47, NewBTreeTester("47"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key47))
		g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("47"))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         50
	 *      /    \      /   \
	 *    5,10  25,30  45   60
	 */
	t.Run("adding the right[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.Create(Key60, NewBTreeTester("460"))).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key50))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key60))
	})
}

func TestBTree_CreateOrFind_ParameterChecks(t *testing.T) {
	g := NewGomegaWithT(t)

	onFind := func(value any) {}
	onCreate := func() any { return "ok" }

	t.Run("it returns an error if the key is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.CreateOrFind(datatypes.EncapsulatedValue{Type: -1, Data: "bad"}, onCreate, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key is invalid:"))
	})

	t.Run("it returns an error if onCreate is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.CreateOrFind(Key1, nil, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onCreate callback cannot be nil"))
	})

	t.Run("it returns an error if oniFind is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.CreateOrFind(Key1, onCreate, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind callback cannot be nil"))
	})
}

func TestBTree_CreateOrFind_SingleNode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a new tree with proper size limits", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("adding the same value multiple times uses onFind for each subsequent call", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(1)))
	})

	t.Run("inserts the values in the proper nodeSize, no matter how they were added to the tree", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key1))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("2"))
		g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("it splits the node when adding a left pivot value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})

	t.Run("it splits the node when adding a pivot value value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})

	t.Run("it splits the node when adding a right pivot value value", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key2, NewBTreeTester("2"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key1, NewBTreeTester("1"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key3, NewBTreeTester("3"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key2))
		g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("2"))

		// left child
		leftchild := bTree.root.children[0]
		g.Expect(leftchild.numberOfValues).To(Equal(1))
		g.Expect(leftchild.keyValues[0].key).To(Equal(Key1))
		g.Expect(leftchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("1"))
		// right child
		rightchild := bTree.root.children[1]
		g.Expect(rightchild.numberOfValues).To(Equal(1))
		g.Expect(rightchild.keyValues[0].key).To(Equal(Key3))
		g.Expect(rightchild.keyValues[0].value.(*BTreeTester).Value).To(Equal("3"))
	})
}

func TestBTree_CreateOrFind_Tree_SimpleOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	// generate a tree of
	/*
	 *     20
	 *   /    \
	 *  10    30
	 */
	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("10"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key20, NewBTreeTester("20"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("30"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("it updates the value on a left child tree if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(1))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(1)))
	})

	t.Run("it updates the value on a right child tree if it already exists", func(t *testing.T) {
		bTree := setupTree(g)

		g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(1))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(1)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  5,10   30
	 */
	t.Run("can add a new entry on the leftChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key5, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key5))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("5"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(leftChild.keyValues[1].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10     25,30
	 */
	t.Run("can add a new entry on the rightChild[0]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key25, NewBTreeTester("25"), OnFindTest)).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key25))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("25"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10,15   30
	 */
	t.Run("can add a new entry on the leftChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key15, NewBTreeTester("15"), OnFindTest)).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(2))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(leftChild.keyValues[1].key).To(Equal(Key15))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("15"))
		g.Expect(leftChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,35
	 */
	t.Run("can add a new entry on the rightChild[1]", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key35, NewBTreeTester("35"), OnFindTest)).ToNot(HaveOccurred())

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(Key35))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("35"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	// generate a tree of
	/*
	 *      20
	 *    /    \
	 *  10    30,T_any
	 */
	t.Run("can add a special T_any node always to the rightmost position", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())

		leftChild := bTree.root.children[0]
		g.Expect(leftChild.numberOfValues).To(Equal(1))
		g.Expect(leftChild.keyValues[0].key).To(Equal(Key10))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("10"))
		g.Expect(leftChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))

		rightChild := bTree.root.children[1]
		g.Expect(rightChild.numberOfValues).To(Equal(2))
		g.Expect(rightChild.keyValues[0].key).To(Equal(Key30))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).Value).To(Equal("30"))
		g.Expect(rightChild.keyValues[0].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
		g.Expect(rightChild.keyValues[1].key).To(Equal(datatypes.Any()))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).Value).To(Equal("5"))
		g.Expect(rightChild.keyValues[1].value.(*BTreeTester).OnFindCount()).To(Equal(int64(0)))
	})

	t.Run("it runs the on Find operation if T_any already exists", func(t *testing.T) {
		var itemCreated *BTreeTester
		bTree := setupTree(g)

		g.Expect(bTree.CreateOrFind(datatypes.Any(), func() any {
			itemCreated = NewBTreeTester("5")().(*BTreeTester)
			return itemCreated
		}, OnFindTest)).ToNot(HaveOccurred())
		g.Expect(itemCreated.onFindCount).To(Equal(int64(0)))

		err := bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("5"), OnFindTest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(itemCreated.onFindCount).To(Equal(int64(1)))
	})
}

func TestBTree_CreateOrFind_Tree_SimplePromoteOperations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("leftChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 * 10,15   30
		 */
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("30"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("10"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key20, NewBTreeTester("20"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key15, NewBTreeTester("15"), OnFindTest)).ToNot(HaveOccurred())

			return bTree
		}

		// generate a tree of
		/*
		 *      10,20
		 *    /   |   \
		 *   8    15  30
		 */
		t.Run("adding the leftChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key8, NewBTreeTester("8"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key10))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key8))
			g.Expect(child1.keyValues[0].value.(*BTreeTester).Value).To(Equal("8"))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      12,20
		 *    /   |   \
		 *   10   15  30
		 */
		t.Run("adding the leftChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key12, NewBTreeTester("12"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key12))
			g.Expect(bTree.root.keyValues[0].value.(*BTreeTester).Value).To(Equal("12"))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key15))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})

		// generate a tree of
		/*
		 *      15,20
		 *    /   |   \
		 *   10   17  30
		 */
		t.Run("adding the leftChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key17, NewBTreeTester("17"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key15))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key20))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key17))
			g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("17"))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key30))
		})
	})

	t.Run("rightChild promotions", func(t *testing.T) {
		// all tests in this section start with a base tree like so
		/*
		 *      20
		 *    /    \
		 *   10   30,35
		 */
		setupTree := func(g *GomegaWithT) *threadSafeBTree {
			bTree, err := NewThreadSafe(2)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("20"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("30"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key20, NewBTreeTester("10"), OnFindTest)).ToNot(HaveOccurred())
			g.Expect(bTree.CreateOrFind(Key35, NewBTreeTester("35"), OnFindTest)).ToNot(HaveOccurred())

			return bTree
		}

		// generate a tree of
		/*
		 *      20,30
		 *    /   |  \
		 *   10   25  35
		 */
		t.Run("adding the rightChild[0] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key25, NewBTreeTester("25"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key30))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key25))
			g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("25"))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,32
		 *    /   |   \
		 *   10   30  35
		 */
		t.Run("adding the rightChild[1] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key32, NewBTreeTester("32"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key32))
			g.Expect(bTree.root.keyValues[1].value.(*BTreeTester).Value).To(Equal("32"))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key35))
		})

		// generate a tree of
		/*
		 *      20,35
		 *    /   |   \
		 *   10   30  37
		 */
		t.Run("adding the rightChild[2] node splits properly", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(Key37, NewBTreeTester("37"), OnFindTest)).ToNot(HaveOccurred())

			g.Expect(bTree.root.numberOfValues).To(Equal(2))
			g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
			g.Expect(bTree.root.keyValues[1].key).To(Equal(Key35))

			child1 := bTree.root.children[0]
			g.Expect(child1.numberOfValues).To(Equal(1))
			g.Expect(child1.keyValues[0].key).To(Equal(Key10))

			child2 := bTree.root.children[1]
			g.Expect(child2.numberOfValues).To(Equal(1))
			g.Expect(child2.keyValues[0].key).To(Equal(Key30))

			child3 := bTree.root.children[2]
			g.Expect(child3.numberOfValues).To(Equal(1))
			g.Expect(child3.keyValues[0].key).To(Equal(Key37))
			g.Expect(child3.keyValues[0].value.(*BTreeTester).Value).To(Equal("37"))
		})
	})
}

func TestBTree_CreateOrFind_Tree_NewRootNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// all tests in this section start with a base tree like so
	/*
	 *        20,40
	 *    /     |     \
	 *  5,10  25,30  45,50
	 */
	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(bTree.CreateOrFind(Key10, NewBTreeTester("10"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key20, NewBTreeTester("20"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key30, NewBTreeTester("30"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key40, NewBTreeTester("40"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key50, NewBTreeTester("50"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key5, NewBTreeTester("5"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key25, NewBTreeTester("25"), OnFindTest)).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(Key45, NewBTreeTester("45"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(2))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))
		g.Expect(bTree.root.keyValues[1].key).To(Equal(Key40))

		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(2))
		g.Expect(child1.keyValues[0].key).To(Equal(Key5))
		g.Expect(child1.keyValues[1].key).To(Equal(Key10))

		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(2))
		g.Expect(child2.keyValues[0].key).To(Equal(Key25))
		g.Expect(child2.keyValues[1].key).To(Equal(Key30))

		child3 := bTree.root.children[2]
		g.Expect(child3.numberOfValues).To(Equal(2))
		g.Expect(child3.keyValues[0].key).To(Equal(Key45))
		g.Expect(child3.keyValues[1].key).To(Equal(Key50))

		return bTree
	}

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        5          40
	 *      /  \      /      \
	 *     0   10   25,30  45,50
	 */
	t.Run("adding the leftChild[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key0, NewBTreeTester("0"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key5))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key0))
		g.Expect(gchild1.keyValues[0].value.(*BTreeTester).Value).To(Equal("0"))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        7          40
	 *      /  \      /      \
	 *     5   10   25,30  45,50
	 */
	t.Run("adding the leftChild[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key7, NewBTreeTester("7"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key7))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key10))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             20
	 *         /         \
	 *        7          40
	 *      /  \      /      \
	 *     5   10   25,30  45,50
	 */
	t.Run("adding the leftChild[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key12, NewBTreeTester("12"), OnFindTest)).ToNot(HaveOccurred())

		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key20))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key10))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key12))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             25
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  22   30   45,50
	 */
	t.Run("adding the middleChild[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key22, NewBTreeTester("22"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key25))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key22))
		g.Expect(gchild2.keyValues[0].value.(*BTreeTester).Value).To(Equal("22"))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             27
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  25   30   45,50
	 */
	t.Run("adding the middleChild[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key27, NewBTreeTester("27"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key27))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key30))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *             30
	 *         /        \
	 *        20        40
	 *      /  \      /    \
	 *    5,10  25   32   45,50
	 */
	t.Run("adding the middleChild[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key32, NewBTreeTester("32"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key30))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key40))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key32))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key45))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         45
	 *      /    \      /   \
	 *    5,10  25,30  42   50
	 */
	t.Run("adding the right[0] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key42, NewBTreeTester("42"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key45))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key42))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         47
	 *      /    \      /   \
	 *    5,10  25,30  45   50
	 */
	t.Run("adding the right[1] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key47, NewBTreeTester("47"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key47))
		g.Expect(child2.keyValues[0].value.(*BTreeTester).Value).To(Equal("47"))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key50))
	})

	// generate a tree of
	/*
	 *              40
	 *         /         \
	 *        20         50
	 *      /    \      /   \
	 *    5,10  25,30  45   60
	 */
	t.Run("adding the right[2] promotes properly", func(t *testing.T) {
		bTree := setupTree(g)
		g.Expect(bTree.CreateOrFind(Key60, NewBTreeTester("460"), OnFindTest)).ToNot(HaveOccurred())

		// root
		g.Expect(bTree.root.numberOfValues).To(Equal(1))
		g.Expect(bTree.root.keyValues[0].key).To(Equal(Key40))

		// left sub stree
		child1 := bTree.root.children[0]
		g.Expect(child1.numberOfValues).To(Equal(1))
		g.Expect(child1.keyValues[0].key).To(Equal(Key20))

		gchild1 := child1.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(2))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key5))
		g.Expect(gchild1.keyValues[1].key).To(Equal(Key10))

		gchild2 := child1.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(2))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key25))
		g.Expect(gchild2.keyValues[1].key).To(Equal(Key30))

		// right sub tree
		child2 := bTree.root.children[1]
		g.Expect(child2.numberOfValues).To(Equal(1))
		g.Expect(child2.keyValues[0].key).To(Equal(Key50))

		gchild1 = child2.children[0]
		g.Expect(gchild1.numberOfValues).To(Equal(1))
		g.Expect(gchild1.keyValues[0].key).To(Equal(Key45))

		gchild2 = child2.children[1]
		g.Expect(gchild2.numberOfValues).To(Equal(1))
		g.Expect(gchild2.keyValues[0].key).To(Equal(Key60))
	})
}
