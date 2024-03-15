package btree

import (
	"sync/atomic"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var (
	Key0   = datatypes.Int(0)
	Key1   = datatypes.Int(1)
	Key2   = datatypes.Int(2)
	Key3   = datatypes.Int(3)
	Key4   = datatypes.Int(4)
	Key5   = datatypes.Int(5)
	Key6   = datatypes.Int(6)
	Key7   = datatypes.Int(7)
	Key8   = datatypes.Int(8)
	Key9   = datatypes.Int(9)
	Key10  = datatypes.Int(10)
	Key12  = datatypes.Int(12)
	Key15  = datatypes.Int(15)
	Key17  = datatypes.Int(17)
	Key20  = datatypes.Int(20)
	Key22  = datatypes.Int(22)
	Key25  = datatypes.Int(25)
	Key27  = datatypes.Int(27)
	Key30  = datatypes.Int(30)
	Key32  = datatypes.Int(32)
	Key35  = datatypes.Int(35)
	Key37  = datatypes.Int(37)
	Key38  = datatypes.Int(38)
	Key40  = datatypes.Int(40)
	Key42  = datatypes.Int(42)
	Key45  = datatypes.Int(45)
	Key47  = datatypes.Int(47)
	Key50  = datatypes.Int(50)
	Key60  = datatypes.Int(60)
	Key70  = datatypes.Int(70)
	Key75  = datatypes.Int(75)
	Key78  = datatypes.Int(78)
	Key80  = datatypes.Int(80)
	Key90  = datatypes.Int(90)
	Key100 = datatypes.Int(100)
	Key110 = datatypes.Int(110)
	Key120 = datatypes.Int(120)
	Key130 = datatypes.Int(130)
)

type BTreeTester struct {
	onFindCount int64
	Value       string
}

func NewBTreeTester(Value string) func() any {
	return func() any {
		return &BTreeTester{
			onFindCount: 0,
			Value:       Value,
		}
	}
}

func OnFindTest(item any) {
	btt := item.(*BTreeTester)
	atomic.AddInt64(&btt.onFindCount, 1)
}

func OnIterateTest(key datatypes.EncapsulatedValue, item any) bool {
	btt := item.(*BTreeTester)
	atomic.AddInt64(&btt.onFindCount, 1)
	return true
}

func (btt *BTreeTester) OnFindCount() int64 {
	return atomic.LoadInt64(&btt.onFindCount)
}
