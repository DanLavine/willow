package btreeassociated

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

// validation I believe boils down to each reference in the ID node is unique to the number of key values they are associated with
// also at each level, a key should only show up the same nummber of times. I.E. a KeyValue identifier of 2, will nly have a guid
// '1111-1111-xxx' 2 times. Similarly a KeyValue identifier of 3 will have '2222-2222-xxx' 3 times. Also the identifiers in 2 will never
// show up in the 3 identifier slots and vice versa
func validateThreadSafeTree(g *GomegaWithT, associatedTree *threadsafeAssociatedTree) {
	identifiers := []map[string]int{}

	// find all the ids and count them
	associatedTree.keys.Iterate(func(_ datatypes.EncapsulatedData, item any) bool {
		valueNode := item.(*threadsafeValuesNode)

		valueNode.values.Iterate(func(_ datatypes.EncapsulatedData, item any) bool {
			idNode := item.(*threadsafeIDNode)

			for index, ids := range idNode.ids {
				if len(identifiers) <= index {
					identifiers = append(identifiers, map[string]int{})
				}

				for _, id := range ids {
					if value, ok := identifiers[index][id]; ok {
						identifiers[index][id] = value + 1
					} else {
						identifiers[index][id] = 1
					}
				}
			}

			return true
		})

		return true
	})

	// ensure that the IDs show up the proper number of times
	keysOne, keysTwo, keysThree, keysFour, keysFive := []string{}, []string{}, []string{}, []string{}, []string{}
	for index, identifier := range identifiers {
		for key, value := range identifier {
			switch index {
			case 0:
				keysOne = append(keysOne, key)
				g.Expect(value).To(Equal(1))
			case 1:
				keysTwo = append(keysTwo, key)
				g.Expect(value).To(Equal(2))
			case 2:
				keysThree = append(keysThree, key)
				g.Expect(value).To(Equal(3))
			case 3:
				keysFour = append(keysFour, key)
				g.Expect(value).To(Equal(4))
			case 4:
				keysFive = append(keysFive, key)
				g.Expect(value).To(Equal(5))
			default:
				g.Fail("Index for key value pairs should only go up to 5")
			}
		}
	}

	g.Expect(len(keysOne)).To(BeNumerically(">=", 2000))
	g.Expect(len(keysTwo)).To(BeNumerically(">=", 2000))
	g.Expect(len(keysThree)).To(BeNumerically(">=", 2000))
	g.Expect(len(keysFour)).To(BeNumerically(">=", 2000))
	g.Expect(len(keysFive)).To(BeNumerically(">=", 2000))

	// vlidate that all keys for their index re unique
	g.Expect(keysTwo).ToNot(ContainElements(keysOne))
	g.Expect(keysThree).ToNot(ContainElements(keysOne))
	g.Expect(keysFour).ToNot(ContainElements(keysOne))
	g.Expect(keysFive).ToNot(ContainElements(keysOne))

	g.Expect(keysThree).ToNot(ContainElements(keysTwo))
	g.Expect(keysFour).ToNot(ContainElements(keysTwo))
	g.Expect(keysFive).ToNot(ContainElements(keysTwo))

	g.Expect(keysFour).ToNot(ContainElements(keysThree))
	g.Expect(keysFive).ToNot(ContainElements(keysThree))

	g.Expect(keysFive).ToNot(ContainElements(keysFour))
}

func validateThreadSafeTreeWithoutKeyLenght(g *GomegaWithT, associatedTree *threadsafeAssociatedTree) {
	identifiers := []map[string]int{}

	// find all the ids and count them
	associatedTree.keys.Iterate(func(_ datatypes.EncapsulatedData, item any) bool {
		valueNode := item.(*threadsafeValuesNode)

		valueNode.values.Iterate(func(_ datatypes.EncapsulatedData, item any) bool {
			idNode := item.(*threadsafeIDNode)

			for index, ids := range idNode.ids {
				if len(identifiers) <= index {
					identifiers = append(identifiers, map[string]int{})
				}

				for _, id := range ids {
					if value, ok := identifiers[index][id]; ok {
						identifiers[index][id] = value + 1
					} else {
						identifiers[index][id] = 1
					}
				}
			}

			return true
		})

		return true
	})

	// ensure that the IDs show up the proper number of times
	keysOne, keysTwo, keysThree, keysFour, keysFive := []string{}, []string{}, []string{}, []string{}, []string{}
	for index, identifier := range identifiers {
		for key, value := range identifier {
			switch index {
			case 0:
				keysOne = append(keysOne, key)
				g.Expect(value).To(Equal(1))
			case 1:
				keysTwo = append(keysTwo, key)
				g.Expect(value).To(Equal(2))
			case 2:
				keysThree = append(keysThree, key)
				g.Expect(value).To(Equal(3))
			case 3:
				keysFour = append(keysFour, key)
				g.Expect(value).To(Equal(4))
			case 4:
				keysFive = append(keysFive, key)
				g.Expect(value).To(Equal(5))
			default:
				g.Fail("Index for key value pairs should only go up to 5")
			}
		}
	}

	// vlidate that all keys for their index re unique
	g.Expect(keysTwo).ToNot(ContainElements(keysOne))
	g.Expect(keysThree).ToNot(ContainElements(keysOne))
	g.Expect(keysFour).ToNot(ContainElements(keysOne))
	g.Expect(keysFive).ToNot(ContainElements(keysOne))

	g.Expect(keysThree).ToNot(ContainElements(keysTwo))
	g.Expect(keysFour).ToNot(ContainElements(keysTwo))
	g.Expect(keysFive).ToNot(ContainElements(keysTwo))

	g.Expect(keysFour).ToNot(ContainElements(keysThree))
	g.Expect(keysFive).ToNot(ContainElements(keysThree))

	g.Expect(keysFive).ToNot(ContainElements(keysFour))
}

func TestAssociated_Random_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnCreate := func() any { return "find me" }
	noOpOnFind := func(item any) {}

	t.Run("It is threadsafe when adding many entries asynchronously", func(t *testing.T) {
		t.Parallel()
		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		testCounter := 10_000
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()
				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				repeatKeys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						repeatKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", i))
					default:
						repeatKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(i)
					}
				}

				_, err := associatedTree.CreateOrFind(repeatKeys, noOpOnCreate, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Create(keys, noOpOnCreate)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, noOpOnCreate)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		validateThreadSafeTree(g, associatedTree)
	})
}

func TestAssociated_Random_Find(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}

	testCounter := 10_000

	t.Run("It is threadsafe when finding many entries asynchronously", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// create 10k entries
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Create(keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// find 10k entries
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				findCounter := 0
				_, err := associatedTree.Find(keys, func(item any) {
					findCounter++
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(findCounter).To(Equal(1), fmt.Sprintf("Index %d has an invalid find counter", tNum))
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				findCounter := 0
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Find(keys, func(item any) {
					findCounter++
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(findCounter).To(Equal(1), fmt.Sprintf("Index %d has an invalid find counter", tNum))
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				findCounter := 0
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Find(keys, func(item any) {
					findCounter++
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(findCounter).To(Equal(1), fmt.Sprintf("Index %d has an invalid find counter", tNum))
			}(i)
		}
		wg.Wait()

		validateThreadSafeTree(g, associatedTree)
	})

	t.Run("It is threadsafe when finding and inserting many entries asynchronously", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// create 10k entries
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Create(keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			// find entries
			//// find create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.Find(keys, func(item any) {
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			//// find create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Find(keys, func(item any) {
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			//// find create with id
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Find(keys, func(item any) {
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum))
				})
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		validateThreadSafeTree(g, associatedTree)
	})
}

func TestAssociated_Random_Match(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}

	t.Run("It is threadsafe when querying many entries asynchronously", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// create 10k entries
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// match 10k entries
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keyValues := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keyValues[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keyValues[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				findCounter := 0
				g.Expect(associatedTree.MatchPermutations(keyValues, func(item *AssociatedKeyValues) bool {
					findCounter++
					return true
				})).ToNot(HaveOccurred())
				g.Expect(findCounter).To(BeNumerically(">=", 1), fmt.Sprintf("Index %d has an invalid match counter", tNum))
			}(i)
		}
		wg.Wait()

		validateThreadSafeTree(g, associatedTree)
	})
}

func TestAssociated_Random_Query(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}

	t.Run("It is threadsafe when querying many entries asynchronously", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// create 10k entries
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// find 10k entries
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				queryDB := datatypes.AssociatedKeyValuesQuery{}
				for i := 0; i < modInt; i++ {
					asociatedKeyValuesQuery := datatypes.AssociatedKeyValuesQuery{}
					switch tNum % 2 {
					case 0:
						strValue := datatypes.String(fmt.Sprintf("%d", tNum))
						asociatedKeyValuesQuery.KeyValueSelection = &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{fmt.Sprintf("%d", i): {Value: &strValue, ValueComparison: datatypes.EqualsPtr()}}}
						queryDB.And = append(queryDB.And, asociatedKeyValuesQuery)
					default:
						intValue := datatypes.Int(tNum)
						asociatedKeyValuesQuery.KeyValueSelection = &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{fmt.Sprintf("%d", i): {Value: &intValue, ValueComparison: datatypes.EqualsPtr()}}}
						queryDB.Or = append(queryDB.Or, asociatedKeyValuesQuery)
					}
				}
				g.Expect(queryDB.Validate()).ToNot(HaveOccurred())

				findCounter := 0
				g.Expect(associatedTree.Query(queryDB, func(item *AssociatedKeyValues) bool {
					findCounter++
					return true
				})).ToNot(HaveOccurred())
				g.Expect(findCounter).To(Equal(1), fmt.Sprintf("Index %d has an invalid find counter", tNum))
			}(i)
		}
		wg.Wait()

		validateThreadSafeTree(g, associatedTree)
	})
}

func TestAssociated_Random_Delete(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}
	testCounter := 10_000

	t.Run("It is threadsafe when deleting many entries asynchronously", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		// create 30k entries
		for i := 0; i < testCounter; i++ {
			wg.Add(1)
			// create or find
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			// create
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				_, err := associatedTree.Create(keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			// create with ID
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, func() any { return tNum })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// delte 10k entries
		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				createOrFindKeys := KeyValues{}
				createKeys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}
				createKeys[datatypes.String(fmt.Sprintf("%d", tNum+testCounter))] = datatypes.String(fmt.Sprintf("%d", tNum))

				g.Expect(associatedTree.Delete(createOrFindKeys, nil)).ToNot(HaveOccurred())
				g.Expect(associatedTree.Delete(createKeys, nil)).ToNot(HaveOccurred())
				g.Expect(associatedTree.DeleteByAssociatedID(fmt.Sprintf("%d", tNum), nil)).ToNot(HaveOccurred())

			}(i)
		}
		wg.Wait()

		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})
}

func TestAssociated_Random_AllActions(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}
	basicCreate := func() any { return true }
	testCounter := 10_000

	t.Run("It is threadsafe when performing all operations in parallel", func(t *testing.T) {
		t.Parallel()

		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		for i := 0; i < testCounter; i++ {
			// create or find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				_, err := associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			// create
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+testCounter)): datatypes.String(fmt.Sprintf("%d", tNum))}

				id, err := associatedTree.Create(keys, basicCreate)
				g.Expect(id).ToNot(Equal(""))
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			// create with id
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				// generate a key with a few different types
				keys := KeyValues{datatypes.String(fmt.Sprintf("%d", tNum+(2*testCounter))): datatypes.String(fmt.Sprintf("%d", tNum))}

				err := associatedTree.CreateWithID(fmt.Sprintf("%d", tNum), keys, basicCreate)
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			// delete
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				createOrFindKeys := KeyValues{}
				createKeys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				createKeys[datatypes.String(fmt.Sprintf("%d", tNum+testCounter))] = datatypes.String(fmt.Sprintf("%d", tNum))

				g.Expect(associatedTree.Delete(createOrFindKeys, nil)).ToNot(HaveOccurred())
				g.Expect(associatedTree.Delete(createKeys, nil)).ToNot(HaveOccurred())
				g.Expect(associatedTree.DeleteByAssociatedID(fmt.Sprintf("%d", tNum), nil)).ToNot(HaveOccurred())
			}(i)

			// find
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				createOrFindKeys := KeyValues{}
				createKeys := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						createOrFindKeys[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				createKeys[datatypes.String(fmt.Sprintf("%d", tNum+testCounter))] = datatypes.String(fmt.Sprintf("%d", tNum))

				_, err := associatedTree.Find(createOrFindKeys, func(item any) { g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal(tNum)) })
				g.Expect(err).ToNot(HaveOccurred())
				_, err = associatedTree.Find(createKeys, func(item any) { g.Expect(item.(*AssociatedKeyValues).Value()).To(BeTrue()) })
				g.Expect(err).ToNot(HaveOccurred())
				err = associatedTree.FindByAssociatedID(fmt.Sprintf("%d", tNum+(2*testCounter)), func(item any) { g.Expect(item.(*AssociatedKeyValues).Value()).To(BeTrue()) })
				g.Expect(err).ToNot(HaveOccurred())
			}(i)

			// match
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keyValues := KeyValues{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keyValues[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keyValues[datatypes.String(fmt.Sprintf("%d", i))] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.MatchPermutations(keyValues, func(item *AssociatedKeyValues) bool {
					for key, value := range keyValues {
						g.Expect(item.KeyValues()).To(HaveKeyWithValue(key, value))
					}

					return true
				})).ToNot(HaveOccurred())
			}(i)

			// query
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				queryDB := datatypes.AssociatedKeyValuesQuery{}
				for i := 0; i < modInt; i++ {
					asociatedKeyValuesQuery := datatypes.AssociatedKeyValuesQuery{}
					switch tNum % 2 {
					case 0:
						strValue := datatypes.String(fmt.Sprintf("%d", tNum))
						asociatedKeyValuesQuery.KeyValueSelection = &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{fmt.Sprintf("%d", i): {Value: &strValue, ValueComparison: datatypes.EqualsPtr()}}}
						queryDB.And = append(queryDB.And, asociatedKeyValuesQuery)
					default:
						intValue := datatypes.Int(tNum)
						asociatedKeyValuesQuery.KeyValueSelection = &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{fmt.Sprintf("%d", i): {Value: &intValue, ValueComparison: datatypes.EqualsPtr()}}}
						queryDB.Or = append(queryDB.Or, asociatedKeyValuesQuery)
					}
				}
				g.Expect(queryDB.Validate()).ToNot(HaveOccurred())
				g.Expect(associatedTree.Query(queryDB, func(item *AssociatedKeyValues) bool { return true })).ToNot(HaveOccurred())
			}(i)

		}
		wg.Wait()

		validateThreadSafeTreeWithoutKeyLenght(g, associatedTree)
	})
}
