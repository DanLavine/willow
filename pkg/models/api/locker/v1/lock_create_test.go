package v1

import (
	"encoding/json"
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_LockCreateRequest_Encoding(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Describe JSON", func(t *testing.T) {
		t.Run("Context Unmarshal", func(t *testing.T) {
			t.Run("It properly decodes a full request", func(t *testing.T) {
				lockCreateRequest := LockCreateRequest{}

				// NOTE: can still parse any, but validation should fail
				reqData := []byte(`
{
	"KeyValues":{
		"key1":{"Type":2,"Data":"3"},
		"any":{"Type":1024}
	},
	"LockTimeout": 0
}`)

				err := json.Unmarshal(reqData, &lockCreateRequest)
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(lockCreateRequest.KeyValues["key1"]).To(Equal(datatypes.Uint16(3)))
			})
		})
	})
}

func Test_LockCreateRequest_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if any key values have the 'Any' type", func(t *testing.T) {
		lockCreateRequest := LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"any_key": datatypes.Any(),
				"str_key": datatypes.String("something"),
			},
		}

		err := lockCreateRequest.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("KeyValues.[any_key].Type: invalid value '1024'. The required value must be with the data types [1:13] inclusively"))
	})
}
