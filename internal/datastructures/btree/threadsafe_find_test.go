package btree

import (
	"fmt"
	"testing"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers/testmodels"
	. "github.com/onsi/gomega"
)

func Test_BTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	intValues := []datatypes.EncapsulatedValue{}
	uintValues := []datatypes.EncapsulatedValue{}

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		intValues = []datatypes.EncapsulatedValue{}
		uintValues = []datatypes.EncapsulatedValue{}

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				intValues = append(intValues, datatypes.Int(i))
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
			} else {
				uintValues = append(uintValues, datatypes.Uint(uint(i)))
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
			}
		}

		return bTree
	}

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Find(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.Find(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorOnFindNil))
	})

	t.Run("It does not run the callback if no items are not found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		called := false
		onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
			called = true
			return false
		}

		g.Expect(bTree.Find(datatypes.String("no good"), noTypesRestriction, onIterate)).ToNot(HaveOccurred())
		g.Expect(called).To(BeFalse())
	})

	t.Run("Context when the key is not T_any", func(t *testing.T) {
		t.Run("It calls the callback when the key is found in the tree", func(t *testing.T) {
			bTree := setupTree(g)

			callCount := 0
			onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
				callCount++
				return true
			}

			bTree.Find(datatypes.Int(0), noTypesRestriction, onIterate)
			g.Expect(callCount).To(Equal(1))
		})

		t.Run("Context when there is a T_any saved in the tree", func(t *testing.T) {
			t.Run("It will call the T_any if they type restrictions allow it", func(t *testing.T) {
				bTree := setupTree(g)
				g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

				keys := []datatypes.EncapsulatedValue{}
				onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
					keys = append(keys, key)
					return true
				}

				g.Expect(bTree.Find(datatypes.Uint(1), noTypesRestriction, onIterate)).ToNot(HaveOccurred())
				g.Expect(len(keys)).To(Equal(2))
				g.Expect(keys).To(ContainElements(datatypes.Any(), datatypes.Uint(1)))
			})

			t.Run("It will call only T_any if type restrictions are for Any only", func(t *testing.T) {
				bTree := setupTree(g)
				g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

				keys := []datatypes.EncapsulatedValue{}
				onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
					keys = append(keys, key)
					return true
				}

				g.Expect(bTree.Find(datatypes.Uint(1), v1common.TypeRestrictions{MinDataType: datatypes.T_any, MaxDataType: datatypes.T_any}, onIterate)).ToNot(HaveOccurred())
				g.Expect(len(keys)).To(Equal(1))
				g.Expect(keys).To(ContainElements(datatypes.Any()))
			})

			t.Run("It will not call the T_any if type restrictions are selective", func(t *testing.T) {
				bTree := setupTree(g)
				g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

				keys := []datatypes.EncapsulatedValue{}
				onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
					keys = append(keys, key)
					return true
				}

				g.Expect(bTree.Find(datatypes.Uint(1), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.T_uint}, onIterate)).ToNot(HaveOccurred())
				g.Expect(len(keys)).To(Equal(1))
				g.Expect(keys).To(ContainElements(datatypes.Uint(1)))
			})
		})
	})

	t.Run("Context when the key is T_any", func(t *testing.T) {
		t.Run("It iterates over all types that match the restrictions", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

			foundAllValues := []datatypes.EncapsulatedValue{}
			foundIntValues := []datatypes.EncapsulatedValue{}
			foundUintValues := []datatypes.EncapsulatedValue{}

			g.Expect(bTree.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.T_int, MaxDataType: datatypes.T_int}, func(key datatypes.EncapsulatedValue, item any) bool {
				g.Expect(key.Type).To(Equal(datatypes.T_int))
				foundIntValues = append(foundIntValues, key)
				return true
			})).ToNot(HaveOccurred())

			g.Expect(bTree.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.T_uint, MaxDataType: datatypes.T_uint}, func(key datatypes.EncapsulatedValue, item any) bool {
				g.Expect(key.Type).To(Equal(datatypes.T_uint))
				foundUintValues = append(foundUintValues, key)
				return true
			})).ToNot(HaveOccurred())

			g.Expect(bTree.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.T_any}, func(key datatypes.EncapsulatedValue, item any) bool {
				foundAllValues = append(foundAllValues, key)
				return true
			})).ToNot(HaveOccurred())

			g.Expect(len(foundIntValues)).To(Equal(len(intValues)))
			g.Expect(foundIntValues).To(ContainElements(intValues))

			g.Expect(len(foundUintValues)).To(Equal(len(uintValues)))
			g.Expect(foundUintValues).To(ContainElements(uintValues))

			g.Expect(len(foundAllValues)).To(Equal(len(foundIntValues) + len(foundUintValues) + 1))
			g.Expect(foundAllValues).To(ContainElements(intValues))
			g.Expect(foundAllValues).To(ContainElements(uintValues))
			g.Expect(foundAllValues).To(ContainElements(datatypes.Any()))
		})

		t.Run("it breaks the iteration when the callback returns false", func(t *testing.T) {
			bTree := setupTree(g)
			g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

			keys := map[datatypes.EncapsulatedValue]struct{}{}
			seenValues := map[string]struct{}{}
			iterate := func(key datatypes.EncapsulatedValue, val any) bool {
				BTreeTester := val.(*BTreeTester)

				// check that each value is unique
				g.Expect(keys).ToNot(HaveKey(key))
				g.Expect(seenValues).ToNot(HaveKey(BTreeTester.Value))
				keys[key] = struct{}{}
				seenValues[BTreeTester.Value] = struct{}{}

				return len(seenValues) < 5
			}

			g.Expect(bTree.Find(datatypes.Any(), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, iterate)).ToNot(HaveOccurred())
			g.Expect(len(seenValues)).To(Equal(5))
		})

	})
}

func Test_BTree_FindNotEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
			} else {
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
			}
		}

		// create the any item
		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("It returns an error if the key is T_any", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindNotEqual(datatypes.Any(), noTypesRestriction, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindNotEqual(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindNotEqual(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})

	t.Run("It can the callback on all items found that are not the passed in key", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindNotEqual(datatypes.Uint(513), noTypesRestriction, onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(1024)) // account for 0-1024 + any, except for 513
		g.Expect(foundItems).ToNot(ContainElement("513"))
	})

	t.Run("It restricts the specific types when finding values", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindNotEqual(datatypes.Int(512), v1common.TypeRestrictions{MinDataType: datatypes.T_int8, MaxDataType: datatypes.T_int}, onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(511)) // only ints
		g.Expect(foundItems).ToNot(ContainElement("512"))
	})

	t.Run("It can find the T_any in the tree when restrictions are set properly", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items)
			return true
		}

		g.Expect(bTree.FindNotEqual(datatypes.String("no find"), v1common.TypeRestrictions{MinDataType: datatypes.T_any, MaxDataType: datatypes.T_any}, onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(1)) // all ints
		g.Expect(foundItems).ToNot(ContainElement("1024"))
	})

	t.Run("It breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items)
			return len(foundItems) < 5
		}

		bTree.FindNotEqual(datatypes.String("512"), noTypesRestriction, onIterate)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThan(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	var intValues []string
	var uintValues []string

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		intValues = []string{}
		uintValues = []string{}

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				intValues = append(intValues, fmt.Sprintf("%d", i))
			} else {
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				uintValues = append(uintValues, fmt.Sprintf("%d", i))
			}
		}

		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("It returns an error if the key is T_any", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThan(datatypes.Any(), noTypesRestriction, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThan(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThan(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})

	t.Run("It does not run the callback no items are found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindLessThan(datatypes.Uint(0), v1common.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.T_string}, onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("It can find all values within the specific restiction range", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindLessThan(datatypes.Int(1022), v1common.TypeRestrictions{MinDataType: datatypes.T_int, MaxDataType: datatypes.T_int}, onIterate)
		g.Expect(len(foundItems)).To(Equal(511)) // find all ints other than the last one
		g.Expect(foundItems).To(ContainElements(intValues[:len(intValues)-1]))
	})

	t.Run("It can find the T_any in the tree if the TypeRestrictions allows it", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindLessThan(datatypes.Int(1022), v1common.TypeRestrictions{MinDataType: datatypes.T_int16, MaxDataType: datatypes.T_any}, onIterate)
		g.Expect(len(foundItems)).To(Equal(512)) // find all ints other than the last one
		g.Expect(foundItems).To(ContainElements(append(intValues[:len(intValues)-1], "1024")))
	})

	t.Run("It breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThan(datatypes.Int(512), noTypesRestriction, onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindLessThanOrEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	var intValues []string
	var uintValues []string

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		intValues = []string{}
		uintValues = []string{}

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				intValues = append(intValues, fmt.Sprintf("%d", i))
			} else {
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				uintValues = append(uintValues, fmt.Sprintf("%d", i))
			}
		}

		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("It returns an error if the key is T_any", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanOrEqual(datatypes.Any(), noTypesRestriction, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanOrEqual(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindLessThanOrEqual(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})

	t.Run("It does not run the callback no items are found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindLessThanOrEqual(datatypes.Uint16(0), testmodels.TypeRestrictionsGeneral(g), onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("It can find all values within the specific restiction range", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindLessThanOrEqual(datatypes.Int(1022), v1common.TypeRestrictions{MinDataType: datatypes.T_int, MaxDataType: datatypes.T_int}, onIterate)
		g.Expect(len(foundItems)).To(Equal(512)) // find all ints
		g.Expect(foundItems).To(ContainElements(intValues))
	})

	t.Run("It can find the T_any in the tree if the TypeRestrictions allows it", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindLessThanOrEqual(datatypes.Int(1022), v1common.TypeRestrictions{MinDataType: datatypes.T_int16, MaxDataType: datatypes.T_any}, onIterate)
		g.Expect(len(foundItems)).To(Equal(513)) // find all ints and T_any
		g.Expect(foundItems).To(ContainElements(append(intValues, "1024")))
	})

	t.Run("It breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindLessThanOrEqual(datatypes.Int(512), noTypesRestriction, onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindGreaterThan(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	var intValues []string
	var uintValues []string

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		intValues = []string{}
		uintValues = []string{}

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				intValues = append(intValues, fmt.Sprintf("%d", i))
			} else {
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				uintValues = append(uintValues, fmt.Sprintf("%d", i))
			}
		}

		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("It returns an error if the key is T_any", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThan(datatypes.Any(), noTypesRestriction, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThan(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThan(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})

	t.Run("It does not run the callback no items are found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindGreaterThan(datatypes.String("no"), testmodels.TypeRestrictionsGeneral(g), onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("It can find all values within the specific restiction range", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindGreaterThan(datatypes.Int(2), v1common.TypeRestrictions{MinDataType: datatypes.T_int8, MaxDataType: datatypes.T_int}, onIterate)
		g.Expect(len(foundItems)).To(Equal(510)) // find all ints other than the first two
		g.Expect(foundItems).To(ContainElements(intValues[2:]))
	})

	t.Run("It can find the T_any in the tree if the TypeRestrictions allows it", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindGreaterThan(datatypes.Int(2), v1common.TypeRestrictions{MinDataType: datatypes.T_int16, MaxDataType: datatypes.T_any}, onIterate)
		g.Expect(len(foundItems)).To(Equal(511)) // find all ints other than the first two and T_any
		g.Expect(foundItems).To(ContainElements(append(intValues[2:], "1024")))
	})

	t.Run("It breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindGreaterThan(datatypes.Int(0), noTypesRestriction, onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}

func TestBTree_FindGreaterThanOrEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	noTypesRestriction := v1common.TypeRestrictions{MinDataType: datatypes.T_uint8, MaxDataType: datatypes.T_any}
	g.Expect(noTypesRestriction.Validate()).ToNot(HaveOccurred())

	var intValues []string
	var uintValues []string

	setupTree := func(g *GomegaWithT) *threadSafeBTree {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		intValues = []string{}
		uintValues = []string{}

		for i := 0; i < 1_024; i++ {
			if i%2 == 0 {
				g.Expect(bTree.CreateOrFind(datatypes.Int(i), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				intValues = append(intValues, fmt.Sprintf("%d", i))
			} else {
				g.Expect(bTree.CreateOrFind(datatypes.Uint(uint(i)), NewBTreeTester(fmt.Sprintf("%d", i)), OnFindTest)).ToNot(HaveOccurred())
				uintValues = append(uintValues, fmt.Sprintf("%d", i))
			}
		}

		g.Expect(bTree.CreateOrFind(datatypes.Any(), NewBTreeTester("1024"), OnFindTest)).ToNot(HaveOccurred())

		return bTree
	}

	t.Run("It returns an error if the key is T_any", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanOrEqual(datatypes.Any(), noTypesRestriction, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("key.Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})

	t.Run("It returns an error if type restriction is invalid", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanOrEqual(datatypes.String("ok"), v1common.TypeRestrictions{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("typeRestrictions.MinDataType: unknown value received '0'"))
	})

	t.Run("It returns an error if the onIterate callback is nil", func(t *testing.T) {
		bTree, err := NewThreadSafe(2)
		g.Expect(err).ToNot(HaveOccurred())

		err = bTree.FindGreaterThanOrEqual(datatypes.String("ok"), noTypesRestriction, nil)
		g.Expect(err).To(Equal(ErrorsOnIterateNil))
	})

	t.Run("It does not run the callback no items are found in the tree", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		g.Expect(bTree.FindGreaterThanOrEqual(datatypes.String("no"), testmodels.TypeRestrictionsGeneral(g), onIterate)).ToNot(HaveOccurred())
		g.Expect(len(foundItems)).To(Equal(0))
	})

	t.Run("It can find all values within the specific restiction range", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int(2), v1common.TypeRestrictions{MinDataType: datatypes.T_int8, MaxDataType: datatypes.T_int}, onIterate)
		g.Expect(len(foundItems)).To(Equal(511)) // find all ints other than the first one
		g.Expect(foundItems).To(ContainElements(intValues[1:]))
	})

	t.Run("It can find the T_any in the tree if the TypeRestrictions allows it", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []string{}
		onIterate := func(key datatypes.EncapsulatedValue, items any) bool {
			foundItems = append(foundItems, items.(*BTreeTester).Value)
			return true
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int(2), v1common.TypeRestrictions{MinDataType: datatypes.T_int16, MaxDataType: datatypes.T_any}, onIterate)
		g.Expect(len(foundItems)).To(Equal(512)) // find all ints other than the first one and T_any
		g.Expect(foundItems).To(ContainElements(append(intValues[1:], "1024")))
	})

	t.Run("It breaks the iteration when find callback returns false", func(t *testing.T) {
		bTree := setupTree(g)

		foundItems := []any{}
		onFind := func(key datatypes.EncapsulatedValue, item any) bool {
			foundItems = append(foundItems, item)
			return len(foundItems) < 5
		}

		bTree.FindGreaterThanOrEqual(datatypes.Int(0), noTypesRestriction, onFind)
		g.Expect(len(foundItems)).To(Equal(5))
	})
}
