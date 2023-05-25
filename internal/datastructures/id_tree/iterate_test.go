package idtree

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestIDGenerator_Iterate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it doesn't iterate if there are no values to iterate over", func(t *testing.T) {
		itemTracker := NewIDTree()

		called := false
		callback := func(key datatypes.CompareType, item any) {
			called = true
		}

		itemTracker.Iterate(callback)
		g.Expect(called).To(BeFalse())
	})

	t.Run("it iterates all values in the tree", func(t *testing.T) {
		/* generate a tree like so:
		height |						   ID
		h0 - 1 (first index in this row) |                1
		h1 - 2 (first index in this row) |        2					       3
		h2 - 4 (first index in this row) |		4			  6			   5				7
		h3 - 8 (first index in this row) |	8		12	10	14	 9	 13	  11  15
		*/
		itemTracker := NewIDTree()
		for i := 1; i <= 15; i++ {
			g.Expect(itemTracker.Add(fmt.Sprintf("%d", i))).To(Equal(uint64(i)))
		}

		expectedKeyValuePairs := map[datatypes.Uint64]string{
			1:  "1",
			2:  "2",
			3:  "3",
			4:  "4",
			5:  "5",
			6:  "6",
			7:  "7",
			8:  "8",
			9:  "9",
			10: "10",
			11: "11",
			12: "12",
			13: "13",
			14: "14",
			15: "15",
		}

		keyValuePairs := map[datatypes.Uint64]string{}
		callback := func(key datatypes.CompareType, item any) {
			keyValuePairs[key.(datatypes.Uint64)] = item.(string)
		}

		itemTracker.Iterate(callback)
		g.Expect(keyValuePairs).To(Equal(expectedKeyValuePairs))
	})
}
