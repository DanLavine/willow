package btreeassociated

import (
	"fmt"
	"testing"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_DestroyByAssociatedID(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnCreate := func() any { return "find me" }

	setupTree := func(g *GomegaWithT) ([]string, *threadsafeAssociatedTree) {
		associatedTree := NewThreadSafe()
		ids := []string{}

		// generate a key with a few different lengths
		for i := 0; i < 100; i++ {
			var keys datatypes.KeyValues
			if i%2 == 0 {
				keys = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.String(fmt.Sprintf("val%d", i)), fmt.Sprintf("key%d", i+1): datatypes.String(fmt.Sprintf("val%d", i+1))}
			} else {
				keys = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.String(fmt.Sprintf("val%d", i))}
			}
			id, err := associatedTree.Create(keys, noOpOnCreate)
			g.Expect(err).ToNot(HaveOccurred())
			ids = append(ids, id)
		}

		return ids, associatedTree
	}

	t.Run("It returns an error if the key is already being destroyed", func(t *testing.T) {
		ids, associatedTree := setupTree(g)

		destroyingChan := make(chan struct{})
		go func() {
			_ = associatedTree.DestroyByAssociatedID(ids[16], func(item AssociatedKeyValues) bool {
				destroyingChan <- struct{}{}
				<-destroyingChan
				return true
			})
		}()

		g.Eventually(destroyingChan).Should(Receive())

		err := associatedTree.DestroyByAssociatedID(ids[16], nil)
		g.Expect(err).To(Equal(ErrorTreeItemDestroying))
		close(destroyingChan)
	})

	t.Run("It returns an error if the tree is already being destroyed", func(t *testing.T) {
		ids, associatedTree := setupTree(g)

		destroyingChan := make(chan struct{})
		go func() {
			counter := 0
			_ = associatedTree.DestroyTree(func(item AssociatedKeyValues) bool {
				if counter == 0 {
					destroyingChan <- struct{}{}
					<-destroyingChan
					counter++
				}
				return true
			})
		}()

		g.Eventually(destroyingChan).Should(Receive())

		err := associatedTree.DestroyByAssociatedID(ids[16], nil)
		g.Expect(err).To(Equal(ErrorTreeDestroying))
		close(destroyingChan)
	})

	t.Run("It deletes the associatedID everything if the callback is nil", func(t *testing.T) {
		ids, associatedTree := setupTree(g)

		err := associatedTree.DestroyByAssociatedID(ids[16], nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the ID is missing
		associatedIDs := []string{}
		queryAll := func(associatedTree AssociatedKeyValues) bool {
			associatedIDs = append(associatedIDs, associatedTree.AssociatedID())
			return true
		}
		g.Expect(associatedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, queryAll)).ToNot(HaveOccurred())
		g.Expect(len(associatedIDs)).To(Equal(99))
		g.Expect(associatedIDs).ToNot(ContainElement(ids[16]))

		// ensure that the idnodes are in a valid state
		identifiers := map[string][]int{}
		associatedTree.keys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
			valueNode := item.(*threadsafeValuesNode)

			valueNode.values.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
				idNode := item.(*threadsafeIDNode)

				for index, ids := range idNode.ids {
					// for each id
					for _, id := range ids {
						// store the ID
						if _, ok := identifiers[id]; !ok {
							identifiers[id] = []int{0, 0}
						}

						// increase how many indexes there are for the ID
						identifiers[id][index]++
					}
				}

				return true
			})

			return true
		})

		// ensure that the IDs show up the proper number of times
		keysOne, keysTwo := []string{}, []string{}
		for associatedIDs, counters := range identifiers {
			if counters[0] != 0 {
				keysOne = append(keysOne, associatedIDs)
			}
			if counters[1] != 0 {
				keysTwo = append(keysTwo, associatedIDs)
			}
		}

		g.Expect(keysOne).ToNot(ContainElement(ids[16]))
		g.Expect(keysTwo).ToNot(ContainElement(ids[16]))
		g.Expect(keysTwo).ToNot(ContainElements(keysOne))
		g.Expect(len(keysOne) + len(keysTwo)).To(Equal(99))
	})

	t.Run("It deletes the key value pair if the onDelete callback returns true", func(t *testing.T) {
		ids, associatedTree := setupTree(g)

		err := associatedTree.DestroyByAssociatedID(ids[16], func(item AssociatedKeyValues) bool { return true })
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the ID is missing
		associatedIDs := []string{}
		queryAll := func(associatedTree AssociatedKeyValues) bool {
			associatedIDs = append(associatedIDs, associatedTree.AssociatedID())
			return true
		}
		g.Expect(associatedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, queryAll)).ToNot(HaveOccurred())
		g.Expect(len(associatedIDs)).To(Equal(99))
		g.Expect(associatedIDs).ToNot(ContainElement(ids[16]))

		// ensure that the idnodes are in a valid state
		identifiers := map[string][]int{}
		associatedTree.keys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
			valueNode := item.(*threadsafeValuesNode)

			valueNode.values.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
				idNode := item.(*threadsafeIDNode)

				for index, ids := range idNode.ids {
					// for each id
					for _, id := range ids {
						// store the ID
						if _, ok := identifiers[id]; !ok {
							identifiers[id] = []int{0, 0}
						}

						// increase how many indexes there are for the ID
						identifiers[id][index]++
					}
				}

				return true
			})

			return true
		})

		// ensure that the IDs show up the proper number of times
		keysOne, keysTwo := []string{}, []string{}
		for associatedIDs, counters := range identifiers {
			if counters[0] != 0 {
				keysOne = append(keysOne, associatedIDs)
			}
			if counters[1] != 0 {
				keysTwo = append(keysTwo, associatedIDs)
			}
		}

		g.Expect(keysOne).ToNot(ContainElement(ids[16]))
		g.Expect(keysTwo).ToNot(ContainElement(ids[16]))
		g.Expect(keysTwo).ToNot(ContainElements(keysOne))
		g.Expect(len(keysOne) + len(keysTwo)).To(Equal(99))
	})

	t.Run("Context when onDelete callback returns false", func(t *testing.T) {
		t.Run("It does not destroy the item and preserves the ID tree", func(t *testing.T) {
			ids, associatedTree := setupTree(g)

			err := associatedTree.DestroyByAssociatedID(ids[17], func(item AssociatedKeyValues) bool { return false })
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the ID is missing
			associatedIDs := []string{}
			queryAll := func(associatedTree AssociatedKeyValues) bool {
				associatedIDs = append(associatedIDs, associatedTree.AssociatedID())
				return true
			}
			g.Expect(associatedTree.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, queryAll)).ToNot(HaveOccurred())
			g.Expect(len(associatedIDs)).To(Equal(100))
			g.Expect(associatedIDs).To(ContainElement(ids[17]))

			// ensure that the idnodes are in a valid state
			identifiers := map[string][]int{}
			associatedTree.keys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
				valueNode := item.(*threadsafeValuesNode)

				valueNode.values.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
					idNode := item.(*threadsafeIDNode)

					for index, ids := range idNode.ids {
						// for each id
						for _, id := range ids {
							// store the ID
							if _, ok := identifiers[id]; !ok {
								identifiers[id] = []int{0, 0}
							}

							// increase how many indexes there are for the ID
							identifiers[id][index]++
						}
					}

					return true
				})

				return true
			})

			// ensure that the IDs show up the proper number of times
			keysOne, keysTwo := []string{}, []string{}
			for associatedIDs, counters := range identifiers {
				if counters[0] != 0 {
					keysOne = append(keysOne, associatedIDs)
				}
				if counters[1] != 0 {
					keysTwo = append(keysTwo, associatedIDs)
				}
			}

			g.Expect(keysOne).To(ContainElement(ids[17]))
			g.Expect(keysTwo).ToNot(ContainElement(ids[17]))
			g.Expect(keysTwo).ToNot(ContainElements(keysOne))
			g.Expect(len(keysOne) + len(keysTwo)).To(Equal(100))
		})
	})
}

func TestAssociatedTree_DestroyTree(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnCreate := func() any { return "find me" }

	setupTree := func(g *GomegaWithT) *threadsafeAssociatedTree {
		associatedTree := NewThreadSafe()

		// generate a key with a few different types
		for i := 0; i < 100; i++ {
			var keys datatypes.KeyValues
			if i%2 == 0 {
				keys = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.String(fmt.Sprintf("val%d", i)), fmt.Sprintf("key%d", i+1): datatypes.String(fmt.Sprintf("val%d", i+1))}
			} else {
				keys = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.String(fmt.Sprintf("val%d", i))}
			}
			_, err := associatedTree.Create(keys, noOpOnCreate)
			g.Expect(err).ToNot(HaveOccurred())
		}

		return associatedTree
	}

	t.Run("It returns an error if the tree desstroy is already in progress", func(t *testing.T) {
		associatedTree := setupTree(g)

		counter := 0
		destroying := make(chan struct{})
		go func() {
			_ = associatedTree.DestroyTree(func(item AssociatedKeyValues) bool {
				if counter == 0 {
					destroying <- struct{}{}
					<-destroying
					counter++
				}
				return true
			})
		}()

		g.Eventually(destroying).Should(Receive())

		err := associatedTree.DestroyTree(nil)
		g.Expect(err).To(Equal(ErrorTreeDestroying))

		destroying <- struct{}{}
	})

	t.Run("It deletes everything if the callback is nil", func(t *testing.T) {
		associatedTree := setupTree(g)

		err := associatedTree.DestroyTree(nil)
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
		g.Expect(associatedTree.associatedIDs.Empty()).To(BeTrue())
	})

	t.Run("It deletes the key value pair if the onDelete callback returns true", func(t *testing.T) {
		associatedTree := setupTree(g)

		count := 0
		associatedTree.DestroyTree(func(item AssociatedKeyValues) bool {
			count++
			return true
		})

		g.Expect(count).To(Equal(100))
		g.Expect(associatedTree.keys.Empty()).To(BeTrue())
		g.Expect(associatedTree.associatedIDs.Empty()).To(BeTrue())
	})

	t.Run("Context when onDelete callback returns false", func(t *testing.T) {
		t.Run("It stop processing on the deleted items and preserves the tree", func(t *testing.T) {
			associatedTree := setupTree(g)

			count := 0
			associatedTree.DestroyTree(func(item AssociatedKeyValues) bool {
				if count < 25 {
					count++
					return true
				}
				return false
			})

			g.Expect(count).To(Equal(25))
			g.Expect(associatedTree.keys.Empty()).ToNot(BeTrue())
			g.Expect(associatedTree.associatedIDs.Empty()).ToNot(BeTrue())

			// ensure that the idnodes are in a valid state
			identifiers := map[string][]int{}
			associatedTree.keys.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
				valueNode := item.(*threadsafeValuesNode)

				valueNode.values.Find(datatypes.Any(), testmodels.NoTypeRestrictions(g), func(_ datatypes.EncapsulatedValue, item any) bool {
					idNode := item.(*threadsafeIDNode)

					for index, ids := range idNode.ids {
						// for each id
						for _, id := range ids {
							// store the ID
							if _, ok := identifiers[id]; !ok {
								identifiers[id] = []int{0, 0}
							}

							// increase how many indexes there are for the ID
							identifiers[id][index]++
						}
					}

					return true
				})

				return true
			})

			// ensure that the IDs show up the proper number of times
			keysOne, keysTwo := []string{}, []string{}
			for associatedIDs, counters := range identifiers {
				if counters[0] != 0 {
					keysOne = append(keysOne, associatedIDs)
				}
				if counters[1] != 0 {
					keysTwo = append(keysTwo, associatedIDs)
				}
			}

			g.Expect(keysTwo).ToNot(ContainElements(keysOne))
			g.Expect(len(keysOne) + len(keysTwo)).To(Equal(75))
		})
	})
}
