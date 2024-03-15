package btree

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func randomEncapsulatedValue(randomNum, index int) datatypes.EncapsulatedValue {
	selectedType := datatypes.GeneralDatatypesSlice[randomNum%len(datatypes.AllDatatypesSlice)]
	switch selectedType {
	case datatypes.T_uint, datatypes.T_uint8, datatypes.T_uint16, datatypes.T_uint32, datatypes.T_uint64:
		if uint(index) > math.MaxUint32 {
			selectedType = datatypes.T_uint64
		} else if uint(index) > math.MaxUint16 {
			selectedType = datatypes.T_uint32
		} else if uint(index) > math.MaxUint8 {
			selectedType = datatypes.T_uint16
		}
	case datatypes.T_int, datatypes.T_int8, datatypes.T_int16, datatypes.T_int32, datatypes.T_int64:
		if index > math.MaxInt32 {
			selectedType = datatypes.T_int64
		} else if index > math.MaxUint16 {
			selectedType = datatypes.T_int32
		} else if index > math.MaxInt8 {
			selectedType = datatypes.T_int16
		}
	}

	switch selectedType {
	case datatypes.T_uint8:
		return datatypes.Uint8(uint8(index))
	case datatypes.T_uint16:
		return datatypes.Uint16(uint16(index))
	case datatypes.T_uint32:
		return datatypes.Uint32(uint32(index))
	case datatypes.T_uint64:
		return datatypes.Uint64(uint64(index))
	case datatypes.T_uint:
		return datatypes.Uint8(uint8(index))
	case datatypes.T_int8:
		return datatypes.Int8(int8(index))
	case datatypes.T_int16:
		return datatypes.Int16(int16(index))
	case datatypes.T_int32:
		return datatypes.Int32(int32(index))
	case datatypes.T_int64:
		return datatypes.Int64(int64(index))
	case datatypes.T_int:
		return datatypes.Int(index)
	case datatypes.T_float32:
		return datatypes.Float32(float32(index))
	case datatypes.T_float64:
		return datatypes.Float64(float64(index))
	case datatypes.T_string:
		return datatypes.String(fmt.Sprintf("%d", index))
	default:
		panic(fmt.Errorf("unknown random number: %d", randomNum))
	}
}

func knownEncapsulatedValue(index int) datatypes.EncapsulatedValue {
	selectedType := datatypes.GeneralDatatypesSlice[index%(len(datatypes.AllDatatypesSlice)-1)]
	switch selectedType {
	case datatypes.T_uint, datatypes.T_uint8, datatypes.T_uint16, datatypes.T_uint32, datatypes.T_uint64:
		if uint(index) > math.MaxUint32 {
			selectedType = datatypes.T_uint64
		} else if uint(index) > math.MaxUint16 {
			selectedType = datatypes.T_uint32
		} else if uint(index) > math.MaxUint8 {
			selectedType = datatypes.T_uint16
		}
	case datatypes.T_int, datatypes.T_int8, datatypes.T_int16, datatypes.T_int32, datatypes.T_int64:
		if index > math.MaxInt32 {
			selectedType = datatypes.T_int64
		} else if index > math.MaxUint16 {
			selectedType = datatypes.T_int32
		} else if index > math.MaxInt8 {
			selectedType = datatypes.T_int16
		}
	}

	switch selectedType {
	case datatypes.T_uint8:
		return datatypes.Uint8(uint8(index))
	case datatypes.T_uint16:
		return datatypes.Uint16(uint16(index))
	case datatypes.T_uint32:
		return datatypes.Uint32(uint32(index))
	case datatypes.T_uint64:
		return datatypes.Uint64(uint64(index))
	case datatypes.T_uint:
		return datatypes.Uint8(uint8(index))
	case datatypes.T_int8:
		return datatypes.Int8(int8(index))
	case datatypes.T_int16:
		return datatypes.Int16(int16(index))
	case datatypes.T_int32:
		return datatypes.Int32(int32(index))
	case datatypes.T_int64:
		return datatypes.Int64(int64(index))
	case datatypes.T_int:
		return datatypes.Int(index)
	case datatypes.T_float32:
		return datatypes.Float32(float32(index))
	case datatypes.T_float64:
		return datatypes.Float64(float64(index))
	case datatypes.T_string:
		return datatypes.String(fmt.Sprintf("%d", index))
	default:
		panic(fmt.Errorf("unknown selected number: %d", selectedType))
	}
}

func TestBTree_Random_Create(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Parallel()

	iterateCount := 10_000

	t.Run("works for a tree nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

		// create a tree of different types
		for i := 0; i < iterateCount; i++ {
			num := iterateCount + i
			randomNum := randomGenerator.Intn(len(datatypes.GeneralDatatypesSlice))

			wg.Add(1)
			go func(randomN int, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(randomEncapsulatedValue(randomN, tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(randomNum, num)
		}

		// create the any tree
		g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())

		// create or find identical values
		for i := 0; i < iterateCount; i++ {
			for k := 0; k < 2; k++ {
				num := iterateCount + i
				key := datatypes.Int(iterateCount + i)

				wg.Add(1)
				go func(tKey datatypes.EncapsulatedValue, tNum int) {
					defer wg.Done()
					g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
				}(key, num)
			}
		}

		// create a full tree
		for i := 0; i < iterateCount; i++ {
			num := iterateCount*2 + i
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("works for a tree noOnFindTestdeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(3)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

		// create a tree of different types
		for i := 0; i < iterateCount; i++ {
			num := iterateCount + i
			randomNum := randomGenerator.Intn(len(datatypes.GeneralDatatypesSlice))

			wg.Add(1)
			go func(randomN int, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(randomEncapsulatedValue(randomN, tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(randomNum, num)
		}
		// create the any tree
		g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())

		// create or find identical values
		for i := 0; i < iterateCount; i++ {
			for k := 0; k < 2; k++ {
				num := iterateCount + i
				key := datatypes.Int(num)

				wg.Add(1)
				go func(tKey datatypes.EncapsulatedValue, tNum int) {
					defer wg.Done()
					g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
				}(key, num)
			}
		}

		// create a full tree
		for i := 0; i < iterateCount; i++ {
			num := iterateCount*2 + i
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("works for a tree nodeSize of 4", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(4)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

		// create a tree of different types
		for i := 0; i < iterateCount; i++ {
			num := iterateCount + i
			randomNum := randomGenerator.Intn(len(datatypes.GeneralDatatypesSlice))

			wg.Add(1)
			go func(randomN int, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(randomEncapsulatedValue(randomN, tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(randomNum, num)
		}

		// create the any tree
		g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())

		// create or find identical values
		for i := 0; i < iterateCount; i++ {
			for k := 0; k < 2; k++ {
				num := iterateCount + i
				key := datatypes.Int(num)

				wg.Add(1)
				go func(tKey datatypes.EncapsulatedValue, tNum int) {
					defer wg.Done()
					g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), OnFindTest)).ToNot(HaveOccurred())
				}(key, num)
			}
		}

		// create a full tree
		for i := 0; i < iterateCount; i++ {
			num := iterateCount*2 + i
			key := datatypes.Int(num)

			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(key, num)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})
}

func TestBTree_Random_Find(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Parallel()

	setup := func(g *GomegaWithT, order int) *threadSafeBTree {
		bTree, err := NewThreadSafe(order)
		g.Expect(err).ToNot(HaveOccurred())

		onFindNoOp := func(item any) {}

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			// create a tree if constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// create a tree of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)
		}

		// create the any tree
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())
		}()

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)

		return bTree
	}

	t.Run("it can find items in parallel with a nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 2)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))

			return true
		}

		for i := 0; i < 10_000; i++ {

			// find of constant keys
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(40_000))) // (10_000 + the any) * 2
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can find items in parallel with a nodeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 3)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))

			return true
		}

		for i := 0; i < 10_000; i++ {

			// find of constant keys
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(40_000))) // (10_000 + the any) * 2
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can find items in parallel with a nodeSize of 4", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 4)

		wg := new(sync.WaitGroup)
		counter := new(atomic.Int64)
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			g.Expect(item).To(BeAssignableToTypeOf(&BTreeTester{}))

			return true
		}

		for i := 0; i < 10_000; i++ {

			// find of constant keys
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(tKey, noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(datatypes.Int(i), i)

			// find of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, onFind)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()

		g.Expect(counter.Load()).To(Equal(int64(40_000))) // (10_000 + the any) * 2
		validateThreadSafeTree(g, bTree.root)
	})
}

func TestBTree_Random_Delete(t *testing.T) {
	g := NewGomegaWithT(t)
	t.Parallel()

	iterateCount := 10_000

	setup := func(g *GomegaWithT, order int) *threadSafeBTree {
		bTree, err := NewThreadSafe(order)
		g.Expect(err).ToNot(HaveOccurred())

		onFindNoOp := func(item any) {}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// create a tree if constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount+i), i)

			// create a tree of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)
		}

		// create the any tree
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())
		}()

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)

		return bTree
	}

	t.Run("it can delete items in parallel with a nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 2)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// delete of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// delete of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// delete the any key
		g.Expect(bTree.Delete(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})

	t.Run("it can delete items in parallel with a nodeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 3)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// delete of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// delete of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// delete the any key
		g.Expect(bTree.Delete(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})

	t.Run("it can delete items in parallel with a nodeSize of 4", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 4)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// delete of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Delete(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// delete of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// delete the any key
		g.Expect(bTree.Delete(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})
}

func TestBTree_Random_Destroy(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)
	iterateCount := 10_000

	setup := func(g *GomegaWithT, order int) *threadSafeBTree {
		bTree, err := NewThreadSafe(order)
		g.Expect(err).ToNot(HaveOccurred())

		onFindNoOp := func(item any) {}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// create a tree if constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue, tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(tKey, NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount+i), i)

			// create a tree of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)
		}

		// create the any tree
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Expect(bTree.Create(datatypes.Any(), NewBTreeTester("something"))).ToNot(HaveOccurred())
		}()

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)

		return bTree
	}

	t.Run("it can destroy items in parallel with a nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 2)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// destroy of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Destroy(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// destroy of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Destroy(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// destroy the any key
		g.Expect(bTree.Destroy(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})

	t.Run("it can destroy items in parallel with a nodeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 3)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// destroy of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Destroy(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// destroy of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Destroy(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// destroy the any key
		g.Expect(bTree.Destroy(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})

	t.Run("it can destroy items in parallel with a nodeSize of 4", func(t *testing.T) {
		t.Parallel()
		bTree := setup(g, 4)

		counter := new(atomic.Int64)
		delete := func(_ datatypes.EncapsulatedValue, item any) bool {
			counter.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < iterateCount; i++ {
			// destroy of constant types
			wg.Add(1)
			go func(tKey datatypes.EncapsulatedValue) {
				defer wg.Done()
				g.Expect(bTree.Destroy(tKey, delete)).ToNot(HaveOccurred())
			}(datatypes.Int(iterateCount + i))

			// destroy of different types
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Destroy(knownEncapsulatedValue(tNum), delete)).ToNot(HaveOccurred())
			}(i)
		}

		// destroy the any key
		g.Expect(bTree.Destroy(datatypes.Any(), delete)).ToNot(HaveOccurred())

		wg.Wait()
		g.Expect(bTree.Empty()).To(BeTrue())
		g.Expect(counter.Load()).To(Equal(int64(iterateCount*2 + 1)))
	})
}

func TestBTree_Random_AllActions(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	onFindNoOp := func(_ any) {}
	onFindPaginateTrue := func(_ datatypes.EncapsulatedValue, _ any) bool { return true }
	onFindPaginateFalse := func(_ datatypes.EncapsulatedValue, _ any) bool { return false }

	t.Run("it can run all actions with a nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(3)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 4", func(t *testing.T) {
		t.Parallel()

		bTree, err := NewThreadSafe(4)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).ToNot(HaveOccurred())
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).ToNot(HaveOccurred())
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})
}

func TestBTree_Random_AllActions_WithDestroyAll(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	onFindNoOp := func(item any) {}
	onFindPaginateTrue := func(_ datatypes.EncapsulatedValue, _ any) bool { return true }
	onFindPaginateFalse := func(_ datatypes.EncapsulatedValue, _ any) bool { return false }

	t.Run("it can run all actions with a nodeSize of 2", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			if i == 5_000 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					g.Expect(bTree.DestroyAll(nil)).ToNot(HaveOccurred())
				}()
			}
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 3", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(3)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			if i == 5_000 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					g.Expect(bTree.DestroyAll(nil)).ToNot(HaveOccurred())
				}()
			}
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})

	t.Run("it can run all actions with a nodeSize of 4", func(t *testing.T) {
		t.Parallel()
		bTree, err := NewThreadSafe(4)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("any"), onFindNoOp)).ToNot(HaveOccurred()) // create any upfront

		iterateCount := 10_000

		wg := new(sync.WaitGroup)
		var onFindPaginate func(key datatypes.EncapsulatedValue, item any) bool
		for i := 0; i < iterateCount; i++ {
			switch i % 3 {
			case 0:
				onFindPaginate = onFindPaginateTrue
			case 1:
				onFindPaginate = onFindPaginateFalse
			case 2:
				if time.Now().Unix()%2 == 0 {
					onFindPaginate = onFindPaginateTrue
				} else {
					onFindPaginate = onFindPaginateFalse
				}
			}

			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.CreateOrFind(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Create(knownEncapsulatedValue(tNum), NewBTreeTester(fmt.Sprintf("%d", tNum)))).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i + iterateCount)

			// find
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.Find(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find not equal
			wg.Add(1)
			go func(tNum int, callback func(_ datatypes.EncapsulatedValue, _ any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindNotEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find less than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindLessThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThan(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// find greater than or equal
			wg.Add(1)
			go func(tNum int, callback func(datatypes.EncapsulatedValue, any) bool) {
				defer wg.Done()
				g.Expect(bTree.FindGreaterThanOrEqual(knownEncapsulatedValue(tNum), noTypesRestriction, callback)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i, onFindPaginate)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				g.Expect(bTree.Delete(knownEncapsulatedValue(tNum), nil)).To(Or(BeNil(), Equal(ErrorTreeDestroying)))
			}(i)

			if i == 5_000 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					g.Expect(bTree.DestroyAll(nil)).ToNot(HaveOccurred())
				}()
			}
		}

		wg.Wait()
		validateThreadSafeTree(g, bTree.root)
	})
}
