package btreeassociated

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

var onFindNoOp = func(item any) {}

func validateThreadSafeValues(g *GomegaWithT, onFind datastructures.OnFind) func(item any) {
	return func(item any) {
		valuesForKey := item.(*keyValues)
		g.Expect(valuesForKey.values.Iterate(onFind)).ToNot(HaveOccurred())
	}
}

func validateThreadSafeKeys(g *GomegaWithT, onFind datastructures.OnFind) func(item any) {
	return func(item any) {
		keyValues := item.(*keyValues)
		g.Expect(keyValues.values.Iterate(validateThreadSafeValues(g, onFind))).ToNot(HaveOccurred())
	}
}

func validateThreadSafeAssociatedTree(g *GomegaWithT, tree *threadsafeAssociatedTree, onFind datastructures.OnFind) {
	if tree.groupedKeyValueAssociation.Empty() {
		return
	}

	g.Expect(tree.groupedKeyValueAssociation.Iterate(validateThreadSafeKeys(g, onFind))).ToNot(HaveOccurred())
}

func validateThreadSafeAssociatedIDs(g *GomegaWithT, tree *threadsafeAssociatedTree, onFind datastructures.OnFind) {
	tree.idTree.Iterate(onFind)
}

func createAsync(g *GomegaWithT, tree *threadsafeAssociatedTree) {
	wg := new(sync.WaitGroup)
	for i := 0; i < 10_000; i++ {
		wg.Add(1)
		go func(tNum int) {
			var keyValues datatypes.StringMap

			switch tNum % 4 {
			case 0:
				keyValues = datatypes.StringMap{fmt.Sprintf("%d", tNum): datatypes.Int(tNum)}
			case 1:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
				}
			case 2:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
				}
			case 3:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
					fmt.Sprintf("%d", tNum+3): datatypes.Int(tNum + 3),
				}
			}

			defer wg.Done()
			g.Expect(tree.CreateOrFind(keyValues, NewJoinTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
		}(i)
	}

	wg.Wait()
}

func createRandomAsync(g *GomegaWithT, tree *threadsafeAssociatedTree, onCreate datastructures.OnCreate, onFind datastructures.OnFind) {
	wg := new(sync.WaitGroup)
	for i := 0; i < 10_000; i++ {
		wg.Add(1)

		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		go func(tNum int) {
			var keyValues datatypes.StringMap

			switch tNum % 4 {
			case 0:
				keyValues = datatypes.StringMap{fmt.Sprintf("%d", tNum): datatypes.Int(tNum)}
			case 1:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
				}
			case 2:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
				}
			case 3:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
					fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
					fmt.Sprintf("%d", tNum+3): datatypes.Int(tNum + 3),
				}
			}

			defer wg.Done()
			g.Expect(tree.CreateOrFind(keyValues, onCreate, onFind)).ToNot(HaveOccurred())
		}(randomGenerator.Intn(10_000))
	}

	wg.Wait()
}

func TestThreadSafeAssociatedTree_Random_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("can create any number of keyValuePairs in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		idCounter := new(atomic.Int64)
		onFindID := func(item any) {
			idCounter.Add(1)
		}

		keyValueCounter := new(atomic.Int64)
		onFindKeyValue := func(item any) {
			keyValueCounter.Add(1)
		}

		createAsync(g, associatedTree)

		validateThreadSafeAssociatedIDs(g, associatedTree, onFindID)
		g.Expect(idCounter.Load()).To(Equal(int64(10_000))) // NOTE: this is for all actaul saved values

		validateThreadSafeAssociatedTree(g, associatedTree, onFindKeyValue)
		g.Expect(keyValueCounter.Load()).To(Equal(int64(25_000))) // NOTE: this is for all the key value groupings
	})

	t.Run("can create any number of random key value pairs in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		createCounter := new(atomic.Int64)
		onCreate := func() any {
			createCounter.Add(1)
			return true
		}

		onFindCreate := func(item any) {}

		createRandomAsync(g, associatedTree, onCreate, onFindCreate)

		idCounter := new(atomic.Int64)
		onFindID := func(item any) {
			idCounter.Add(1)
		}

		validateThreadSafeAssociatedIDs(g, associatedTree, onFindID)
		g.Expect(idCounter.Load()).To(Equal(createCounter.Load()))
	})
}

func TestThreadSafeAssociatedTree_Random_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("can find any number of keys in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		createAsync(g, associatedTree)

		foundIDs := new(atomic.Int64)
		onFind := func(item any) {
			foundIDs.Add(1)
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				var keyValues datatypes.StringMap
				switch tNum % 4 {
				case 0:
					keyValues = datatypes.StringMap{fmt.Sprintf("%d", tNum): datatypes.Int(tNum)}
				case 1:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					}
				case 2:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
					}
				case 3:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
						fmt.Sprintf("%d", tNum+3): datatypes.Int(tNum + 3),
					}
				}

				defer wg.Done()
				g.Expect(associatedTree.Find(keyValues, onFind)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		g.Expect(foundIDs.Load()).To(Equal(int64(10_000)))
	})

	t.Run("can find any number of random keys in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		createAsync(g, associatedTree)

		foundIDs := new(atomic.Int64)
		onFind := func(item any) {
			foundIDs.Add(1)
		}

		wg := new(sync.WaitGroup)
		randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				var keyValues datatypes.StringMap
				switch tNum % 4 {
				case 0:
					keyValues = datatypes.StringMap{fmt.Sprintf("%d", tNum): datatypes.Int(tNum)}
				case 1:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					}
				case 2:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
					}
				case 3:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
						fmt.Sprintf("%d", tNum+3): datatypes.Int(tNum + 3),
					}
				}

				defer wg.Done()
				g.Expect(associatedTree.Find(keyValues, onFind)).ToNot(HaveOccurred())
			}(randomGenerator.Intn(10_000))
		}

		wg.Wait()
		g.Expect(foundIDs.Load()).To(Equal(int64(10_000)))
	})
}

func TestThreadSafeAssociatedTree_Random_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("can delete any number of keys in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		createAsync(g, associatedTree)

		deleteCount := new(atomic.Int64)
		canDelete := func(item any) bool {
			deleteCount.Add(1)
			return true
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				var keyValues datatypes.StringMap
				switch tNum % 4 {
				case 0:
					keyValues = datatypes.StringMap{fmt.Sprintf("%d", tNum): datatypes.Int(tNum)}
				case 1:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
					}
				case 2:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
					}
				case 3:
					keyValues = datatypes.StringMap{
						fmt.Sprintf("%d", tNum):   datatypes.Int(tNum),
						fmt.Sprintf("%d", tNum+1): datatypes.Int(tNum + 1),
						fmt.Sprintf("%d", tNum+2): datatypes.Int(tNum + 2),
						fmt.Sprintf("%d", tNum+3): datatypes.Int(tNum + 3),
					}
				}

				defer wg.Done()
				g.Expect(associatedTree.Delete(keyValues, canDelete)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		g.Expect(deleteCount.Load()).To(Equal(int64(10_000)))
		g.Expect(associatedTree.groupedKeyValueAssociation.Empty()).To(BeTrue())
	})
}

func TestThreadSafeAssociatedTree_Random_AllActions(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("can create, delete and find any number of keys at random", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		createAsync(g, associatedTree)

		wg := new(sync.WaitGroup)
		for i := 0; i < 10_000; i++ {
			var keyValues datatypes.StringMap
			switch i % 4 {
			case 0:
				keyValues = datatypes.StringMap{fmt.Sprintf("%d", i): datatypes.Int(i)}
			case 1:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", i):   datatypes.Int(i),
					fmt.Sprintf("%d", i+1): datatypes.Int(i + 1),
				}
			case 2:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", i):   datatypes.Int(i),
					fmt.Sprintf("%d", i+1): datatypes.Int(i + 1),
					fmt.Sprintf("%d", i+2): datatypes.Int(i + 2),
				}
			case 3:
				keyValues = datatypes.StringMap{
					fmt.Sprintf("%d", i):   datatypes.Int(i),
					fmt.Sprintf("%d", i+1): datatypes.Int(i + 1),
					fmt.Sprintf("%d", i+2): datatypes.Int(i + 2),
					fmt.Sprintf("%d", i+3): datatypes.Int(i + 3),
				}
			}

			// add
			wg.Add(1)
			go func(tKeyValues datatypes.StringMap, tNum int) {
				defer wg.Done()
				g.Expect(associatedTree.CreateOrFind(tKeyValues, NewJoinTreeTester(fmt.Sprintf("%d", tNum)), onFindNoOp)).ToNot(HaveOccurred())
			}(keyValues, i)

			// find
			wg.Add(1)
			go func(tKeyValues datatypes.StringMap, tNum int) {
				defer wg.Done()
				g.Expect(associatedTree.Find(tKeyValues, onFindNoOp)).ToNot(HaveOccurred())
			}(keyValues, i)

			// delete
			wg.Add(1)
			go func(tKeyValues datatypes.StringMap, tNum int) {
				defer wg.Done()
				g.Expect(associatedTree.Delete(tKeyValues, func(item any) bool { return true })).ToNot(HaveOccurred())
			}(keyValues, i)
		}

		wg.Wait()
	})
}
