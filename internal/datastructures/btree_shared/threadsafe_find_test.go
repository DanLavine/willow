package btreeshared

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func TestAssociatedTree_Find_ParamCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.StringMap{"1": datatypes.Int(1)}
	onFind := func(item any) {}

	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Find(nil, onFind)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
	})

	t.Run("it returns an error with nil onFind", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		err := associatedTree.Find(keys, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
	})
}

func TestAssociatedTree_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	keys := datatypes.StringMap{"1": datatypes.Int(1)}
	noOpOnFind := func(item any) {}

	t.Run("it does not run the callback when the value cannot be found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		found := false
		onFind := func(item any) {
			found = true
		}

		err := associatedTree.Find(keys, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(found).To(BeFalse())
	})

	t.Run("it fails fast if a key value index is not found", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := datatypes.StringMap{"1": datatypes.String("1")}
		keyValues2 := datatypes.StringMap{"1": datatypes.String("1"), "2": datatypes.Float32(3.4)}

		// create a single key value pair
		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())

		// this should break fast in the code since nothing has 2 indexes
		g.Expect(associatedTree.Find(keyValues2, noOpOnFind)).ToNot(HaveOccurred())
	})

	t.Run("it runs the callback for only key value pairs who match exacly", func(t *testing.T) {
		associatedTree := NewThreadSafe()

		keyValues1 := datatypes.StringMap{"1": datatypes.Int(1)}
		keyValues2 := datatypes.StringMap{"1": datatypes.String("1")}
		keyValues3 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Float32(3.4)}
		keyValues4 := datatypes.StringMap{"1": datatypes.Int(1), "2": datatypes.Float64(3.4)}

		keys := []string{}
		onFind := func(key string) func(item any) {
			return func(item any) {
				switch key {
				case "1":
					keys = append(keys, "1")
					g.Expect(item).To(Equal("1"))
				case "2":
					keys = append(keys, "2")
					g.Expect(item).To(Equal("2"))
				case "3":
					keys = append(keys, "3")
					g.Expect(item).To(Equal("3"))
				case "4":
					keys = append(keys, "4")
					g.Expect(item).To(Equal("4"))
				default:
					g.Fail("Unexpected key")
				}
			}
		}

		g.Expect(associatedTree.CreateOrFind(keyValues1, func() any { return "1" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues2, func() any { return "2" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues3, func() any { return "3" }, noOpOnFind)).ToNot(HaveOccurred())
		g.Expect(associatedTree.CreateOrFind(keyValues4, func() any { return "4" }, noOpOnFind)).ToNot(HaveOccurred())

		g.Expect(associatedTree.Find(keyValues1, onFind("1"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.Find(keyValues2, onFind("2"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.Find(keyValues3, onFind("3"))).ToNot(HaveOccurred())
		g.Expect(associatedTree.Find(keyValues4, onFind("4"))).ToNot(HaveOccurred())
		g.Expect(keys).To(ContainElements("1", "2", "3", "4"))
	})
}
