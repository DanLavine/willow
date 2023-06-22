package btreeshared

//import (
//	"testing"
//
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//
//	. "github.com/onsi/gomega"
//)
//
//func TestSharedTree_CreateOrFind_ParamCheck(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	keys := datatypes.StringMap{"1": datatypes.Int(1)}
//	onCreate := func() any { return true }
//	onFind := func(item any) {}
//
//	t.Run("it returns an error with nil keyValues", func(t *testing.T) {
//		sharedTree := NewThreadSafe()
//
//		err := sharedTree.CreateOrFind(nil, onCreate, onFind)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(ContainSubstring("keyValuePairs cannot be empty"))
//	})
//
//	t.Run("it returns an error with nil onCreate", func(t *testing.T) {
//		sharedTree := NewThreadSafe()
//
//		err := sharedTree.CreateOrFind(keys, nil, onFind)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(ContainSubstring("onCreate cannot be nil"))
//	})
//
//	t.Run("it returns an error with nil onFind", func(t *testing.T) {
//		sharedTree := NewThreadSafe()
//
//		err := sharedTree.CreateOrFind(keys, onCreate, nil)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(ContainSubstring("onFind cannot be nil"))
//	})
//}
//
//func TestSharedTree_CreateOrFind_SingleKeyValue(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	keyValues := datatypes.StringMap{"1": datatypes.String("one")}
//	//noOpOnCreate := func() any { return true }
//	noOpOnFind := func(item any) {}
//
//	t.Run("it creates a value if it doesn't already exist", func(t *testing.T) {
//		associatedTree := NewThreadSafe()
//
//		called := false
//		onCreate := func() any {
//			called = true
//			return true
//		}
//
//		err := associatedTree.CreateOrFind(keyValues, onCreate, noOpOnFind)
//		g.Expect(err).ToNot(HaveOccurred())
//		g.Expect(called).To(BeTrue())
//	})
//}
