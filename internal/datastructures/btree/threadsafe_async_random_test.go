package btree

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestBTree_Random_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	iterateCount := 10_000

	t.Run("works for a tree nodeSize of 2", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < iterateCount; i++ {
			num := randomGenerator.Intn(iterateCount)
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("works for a tree nodeSize of 3", func(t *testing.T) {
		bTree, err := NewThreadSafe(3)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("works for a tree nodeSize of 4", func(t *testing.T) {
		bTree, err := NewThreadSafe(4)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			num := randomGenerator.Intn(10_000)
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})
}

func TestBTree_Random_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	setup := func(g *GomegaWithT, order int) *threadSafeBTree {
		bTree, err := NewThreadSafe(order)
		g.Expect(err).ToNot(HaveOccurred())

		onFindNoOp := func(item any) {}

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)

		return bTree
	}

	t.Run("it can find items in parallel with a nodeSize of 2", func(t *testing.T) {
		bTree := setup(g, 2)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				onFind := func(item any) {
					counter.Add(1)
					g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))
				}

				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(10_000)))
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can find items in parallel with a nodeSize of 3", func(t *testing.T) {
		bTree := setup(g, 3)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				onFind := func(item any) {
					counter.Add(1)
					g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))
				}

				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(10_000)))
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can find items in parallel with a nodeSize of 4", func(t *testing.T) {
		bTree := setup(g, 4)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				onFind := func(item any) {
					counter.Add(1)
					g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))
				}

				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(10_000)))
		validateThreadSafeTree(g, bTree.root)
	})
}

func TestBTree_Random_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	iterateCount := 10_000

	setup := func(g *GomegaWithT, order int) *threadSafeBTree {
		bTree, err := NewThreadSafe(order)
		g.Expect(err).ToNot(HaveOccurred())

		onFindNoOp := func(item any) {}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)

		return bTree
	}

	t.Run("it can delete items in parallel with a nodeSize of 2", func(t *testing.T) {
		bTree := setup(g, 2)
		bTree.root.print("")

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		if !bTree.Empty() {
			bTree.root.print("")
		}
		g.Expect(bTree.Empty()).To(BeTrue())
	})

	t.Run("it can delete items in parallel with a nodeSize of 3", func(t *testing.T) {
		bTree := setup(g, 3)

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
	})

	t.Run("it can delete items in parallel with a nodeSize of 4", func(t *testing.T) {
		bTree := setup(g, 4)

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
	})
}

func TestBTree_Random_AllActions(t *testing.T) {
	g := NewGomegaWithT(t)

	onFindNoOp := func(item any) {}

	t.Run("it can run all actions with a nodeSize of 2", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			// add
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// delete
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 3", func(t *testing.T) {
		bTree, err := NewThreadSafe(3)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			// add
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// delete
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 4", func(t *testing.T) {
		bTree, err := NewThreadSafe(4)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			// add
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// delete
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedData, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, nil)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})
}
