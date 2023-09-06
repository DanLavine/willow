package btreeassociated

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"

	. "github.com/onsi/gomega"
)

// validation I believe boils down to each reference in the ID node is unique to the number of key values they are associated with
// also at each level, a key should only show up the same nummber of times. I.E. a KeyValue identifier of 2, will nly have a guid
// '1111-1111-xxx' 2 times. Similarly a KeyValue identifier of 3 will have '2222-2222-xxx' 3 times. Also the identifiers in 2 will never
// show up in the 3 identifier slots and vice versa
func validateThreadSafeTree(g *GomegaWithT, associatedTree *threadsafeAssociatedTree) {
	identifiers := []map[string]int{}

	// find all the ids and count them
	associatedTree.keys.Iterate(func(item any) bool {
		valueNode := item.(*threadsafeValuesNode)

		valueNode.values.Iterate(func(item any) bool {
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
	associatedTree.keys.Iterate(func(item any) bool {
		valueNode := item.(*threadsafeValuesNode)

		valueNode.values.Iterate(func(item any) bool {
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
		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		for i := 0; i < 10_000; i++ {
			wg.Add(1)
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := datatypes.StringMap{}
				repeatKeys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
						repeatKeys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", i))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
						repeatKeys[fmt.Sprintf("%d", i)] = datatypes.Int(i)
					}
				}

				g.Expect(associatedTree.CreateOrFind(keys, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())
				g.Expect(associatedTree.CreateOrFind(repeatKeys, noOpOnCreate, noOpOnFind)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
		validateThreadSafeTree(g, associatedTree)
	})
}

func TestAssociated_Random_Find(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}

	t.Run("It is threadsafe when finding many entries asynchronously", func(t *testing.T) {
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
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)).ToNot(HaveOccurred())
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
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				findCounter := 0
				g.Expect(associatedTree.Find(keys, func(item any) {
					findCounter++
					g.Expect(item).To(Equal(tNum))
				})).ToNot(HaveOccurred())

				g.Expect(findCounter).To(Equal(1), fmt.Sprintf("Index %d has an invalid find counter", tNum))
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
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)).ToNot(HaveOccurred())
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
				selection := query.Select{}
				for i := 0; i < modInt; i++ {
					whereSelection := query.Select{}
					switch tNum % 2 {
					case 0:
						strValue := datatypes.String(fmt.Sprintf("%d", tNum))
						whereSelection.Where = &query.Query{KeyValues: map[string]query.Value{fmt.Sprintf("%d", i): {Value: &strValue, ValueComparison: query.EqualsPtr()}}}
						selection.And = append(selection.And, whereSelection)
					default:
						intValue := datatypes.Int(tNum)
						whereSelection.Where = &query.Query{KeyValues: map[string]query.Value{fmt.Sprintf("%d", i): {Value: &intValue, ValueComparison: query.EqualsPtr()}}}
						selection.Or = append(selection.Or, whereSelection)
					}
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				findCounter := 0
				g.Expect(associatedTree.Query(selection, func(item any) bool {
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

	t.Run("It is threadsafe when deleting many entries asynchronously", func(t *testing.T) {
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
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)).ToNot(HaveOccurred())
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
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.Delete(keys, nil)).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		g.Expect(associatedTree.ids.Empty()).To(BeTrue())
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
	})
}

func TestAssociated_Random_AllActions(t *testing.T) {
	g := NewGomegaWithT(t)
	noOpOnFind := func(item any) {}

	t.Run("It is threadsafe when performing all operations in parallel", func(t *testing.T) {
		associatedTree := NewThreadSafe()
		wg := new(sync.WaitGroup)

		for i := 0; i < 10_000; i++ {
			wg.Add(4)

			// create
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.CreateOrFind(keys, func() any { return tNum }, noOpOnFind)).ToNot(HaveOccurred())
			}(i)

			// delete
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.Delete(keys, nil)).ToNot(HaveOccurred())
			}(i)

			// find
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				keys := datatypes.StringMap{}
				for i := 0; i < modInt; i++ {
					switch tNum % 2 {
					case 0:
						keys[fmt.Sprintf("%d", i)] = datatypes.String(fmt.Sprintf("%d", tNum))
					default:
						keys[fmt.Sprintf("%d", i)] = datatypes.Int(tNum)
					}
				}

				g.Expect(associatedTree.Find(keys, func(item any) { g.Expect(item).To(Equal(tNum)) })).ToNot(HaveOccurred())
			}(i)

			// query
			go func(tNum int) {
				defer wg.Done()

				modInt := tNum % 5
				modInt++

				// generate a key with a few different types
				selection := query.Select{}
				for i := 0; i < modInt; i++ {
					whereSelection := query.Select{}
					switch tNum % 2 {
					case 0:
						strValue := datatypes.String(fmt.Sprintf("%d", tNum))
						whereSelection.Where = &query.Query{KeyValues: map[string]query.Value{fmt.Sprintf("%d", i): {Value: &strValue, ValueComparison: query.EqualsPtr()}}}
						selection.And = append(selection.And, whereSelection)
					default:
						intValue := datatypes.Int(tNum)
						whereSelection.Where = &query.Query{KeyValues: map[string]query.Value{fmt.Sprintf("%d", i): {Value: &intValue, ValueComparison: query.EqualsPtr()}}}
						selection.Or = append(selection.Or, whereSelection)
					}
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())
				g.Expect(associatedTree.Query(selection, func(item any) bool {
					return true
				})).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		validateThreadSafeTreeWithoutKeyLenght(g, associatedTree)
	})
}
