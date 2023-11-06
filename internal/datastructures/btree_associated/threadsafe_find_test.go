package btreeassociated

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Find_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
	onFind := func(item any) {}

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.Find(nil, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
		g.Expect(id).To(Equal(""))
	})

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		id, err := associatedTree.Find(keys, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
		g.Expect(id).To(Equal(""))
	})
}

func TestAssociatedTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
	noOpOnFind := func(item any) {}

	t.Run("it does not run the callback when the value cannot be found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		found := false
		onFind := func(item any) {
			found = true
		}

		id, err := associatedTree.Find(keys, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
		g.Expect(id).To(Equal(""))
	})

	t.Run("it doesn't return an id if the key value pairs are not found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1"), "2": datatypes.Float32(3.4)})

		// create a single key value pair
		_, err := associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())

		// this should break fast in the code since nothing has 2 indexes
		id, err := associatedTree.Find(keyValues2, noOpOnFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(id).To(Equal(""))
	})

	t.Run("it runs the callback for only key value pairs who match exacly", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Float32(3.4)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Float64(3.4)})

		keys := []string{}
		onFind := func(key string) func(item any) {
			return func(item any) {
				switch key {
				case "1":
					keys = append(keys, "1")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("1"))
				case "2":
					keys = append(keys, "2")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("2"))
				case "3":
					keys = append(keys, "3")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("3"))
				case "4":
					keys = append(keys, "4")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("4"))
				default:
					g.Fail("Unexpected key")
				}
			}
		}

		_, _ = associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		_, _ = associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		_, _ = associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		_, _ = associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)

		id1, err := associatedTree.Find(keyValues1, onFind("1"))
		g.Expect(err).ToNot(HaveOccurred())
		id2, err := associatedTree.Find(keyValues2, onFind("2"))
		g.Expect(err).ToNot(HaveOccurred())
		id3, err := associatedTree.Find(keyValues3, onFind("3"))
		g.Expect(err).ToNot(HaveOccurred())
		id4, err := associatedTree.Find(keyValues4, onFind("4"))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(id1).ToNot(Or(Equal(id2), Equal(id3), Equal(id4)))
		g.Expect(id2).ToNot(Or(Equal(id3), Equal(id4)))
		g.Expect(id3).ToNot(Equal(id4))
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(keys).To(ContainElements("1", "2", "3", "4"))
	})
}

func TestAssociatedTree_FindByAssociatedID_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.FindByAssociatedID("something", nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
	})
}

func TestAssociatedTree_FindByAssociatedID(t *testing.T) {
	g := NewGomegaWithT(t)

	noOpOnFind := func(item any) {}

	t.Run("it does not run the callback when the value cannot be found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		found := false
		onFind := func(item any) {
			found = true
		}

		err := associatedTree.FindByAssociatedID("not found", onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})

	t.Run("it runs the callback when the value is found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1)})
		keyValues2 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.String("1")})
		keyValues3 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Float32(3.4)})
		keyValues4 := ConverDatatypesKeyValues(datatypes.KeyValues{"1": datatypes.Int(1), "2": datatypes.Float64(3.4)})

		keys := []string{}
		onFind := func(key string) func(item any) {
			return func(item any) {
				switch key {
				case "1":
					keys = append(keys, "1")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("1"))
				case "2":
					keys = append(keys, "2")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("2"))
				case "3":
					keys = append(keys, "3")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("3"))
				case "4":
					keys = append(keys, "4")
					g.Expect(item.(*AssociatedKeyValues).Value()).To(Equal("4"))
				default:
					g.Fail("Unexpected key")
				}
			}
		}

		id1, _ := associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)
		id2, _ := associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)
		id3, _ := associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)
		id4, _ := associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)

		g.Expect(associatedTree.FindByAssociatedID(id1, onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.FindByAssociatedID(id2, onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.FindByAssociatedID(id3, onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.FindByAssociatedID(id4, onFind("4"))).ToNot(HaveOccurred())

	})
}
