package datatypes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func MatchOnce(g *GomegaWithT, group []KeyValues, val KeyValues) {
	count := 0
	for _, groupVal := range group {
		if reflect.DeepEqual(groupVal, val) {
			count++
		}
	}

	_, _, line, _ := runtime.Caller(1)
	g.Expect(count).To(Equal(1), fmt.Sprintf("line: %d, vale: %v", line, val))
}

func TestKeyValues_Keys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map", func(t *testing.T) {
		KeyValues := KeyValues{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := KeyValues.Keys()
		g.Expect(keys).To(ContainElements([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestKeyValues_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns all keys for the map in a sorted order", func(t *testing.T) {
		KeyValues := KeyValues{
			"a": String("1"),
			"b": String("2"),
			"c": String("3"),
			"d": Int(4),
			"e": String("5"),
		}

		keys := KeyValues.SortedKeys()
		g.Expect(keys).To(Equal([]string{"a", "b", "c", "d", "e"}))
	})
}

func TestKeyValues_DataEncoding(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It encodes data properly through JSON", func(t *testing.T) {
		// data to encode
		keyValues := KeyValues{
			"uint8":   Uint8(1),
			"uint16":  Uint16(1),
			"uint32":  Uint32(1),
			"uint64":  Uint64(1),
			"uint":    Uint(1),
			"Int8":    Int8(1),
			"Int16":   Int16(1),
			"Int32":   Int32(1),
			"Int64":   Int64(1),
			"Int":     Int(1),
			"Float32": Float32(1),
			"Float64": Float64(1),
			"String":  String("1"),
			"Any":     Any(),
		}

		data, err := json.Marshal(keyValues)
		g.Expect(err).ToNot(HaveOccurred())

		// decode the data
		decodedKeyValues := KeyValues{}
		err = json.Unmarshal(data, &decodedKeyValues)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure bothe values match
		g.Expect(keyValues).To(Equal(decodedKeyValues))
	})

	t.Run("It returns an error when json Type is invalid", func(t *testing.T) {
		decodedKeyValues := KeyValues{}
		err := json.Unmarshal([]byte(`{"key1":{"other":13, "Data": "some str"}}`), &decodedKeyValues)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("'key1' has an invalid value: failed to decode JSON. Unknown type '0'"))
	})

	t.Run("It returns an error when json Data is invalid", func(t *testing.T) {
		decodedKeyValues := KeyValues{}
		err := json.Unmarshal([]byte(`{"key1":{"Type":12, "unnkown": "some str"}}`), &decodedKeyValues)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("'key1' has an invalid value: type 'float64' has nil Data. Requires a castable string to the provided type"))
	})
}

func TestKeyValues_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can validate a nil map", func(t *testing.T) {
		var kvs KeyValues

		err := kvs.Validate(MinDataType, MaxDataType)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("recieved no KeyValues, but requires a length of at least 1"))
	})
}
