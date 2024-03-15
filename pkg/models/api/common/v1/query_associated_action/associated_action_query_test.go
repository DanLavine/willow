package queryassociatedaction

import (
	"encoding/json"
	"testing"

	"github.com/DanLavine/willow/internal/helpers"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_AssociatedActionQuery_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It allows an empty AssociatedActionQuery as a Select All operation", func(t *testing.T) {
		associatedQuery := AssociatedActionQuery{}
		g.Expect(associatedQuery.Validate()).ToNot(HaveOccurred())
	})

	t.Run("It reports errors nicely", func(t *testing.T) {
		associatedQuery := &AssociatedActionQuery{
			Or: []*AssociatedActionQuery{
				{
					And: []*AssociatedActionQuery{
						{
							Selection: &Selection{
								KeyValues: map[string]ValueQuery{
									"Key1": ValueQuery{
										Value:      datatypes.String("valid value"),
										Comparison: v1.Equals,
										TypeRestrictions: v1.TypeRestrictions{
											MinDataType: 0,
											MaxDataType: 0,
										},
									},
									"key2": ValueQuery{
										Value:      datatypes.String("valid value"),
										Comparison: v1.Equals,
										TypeRestrictions: v1.TypeRestrictions{
											MinDataType: datatypes.MinDataType,
											MaxDataType: datatypes.MaxDataType,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		err := associatedQuery.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(`Or[0].And[0].Selection.KeyValues[Key1].TypeRestrictions.MinDataType: unknown value received '0'`))
	})
}

func Test_AssociatedActionQuery_Encode(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Describe JSON", func(t *testing.T) {
		t.Run("Context Unmarshal", func(t *testing.T) {
			t.Run("It decodes an empty request as select all", func(t *testing.T) {
				associatedQuery := AssociatedActionQuery{}

				err := json.Unmarshal([]byte(`{}`), &associatedQuery)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(associatedQuery.Validate()).ToNot(HaveOccurred())
			})

			t.Run("It properly decodes a full request", func(t *testing.T) {
				associatedQuery := &AssociatedActionQuery{}

				reqData := []byte(`
{
	"Selection": {
		"MinNumberOfKeyValues": 2,
		"MaxNumberOfKeyValues": 5,
		"KeyValues": {
			"key1":{
				"Value": {
					"Type": 2,
					"Data": "3"
				},
				"Comparison": "=",
				"TypeRestrictions": {
					"MinDataType": 2,
					"MaxDataType": 1024
				}
			}
		}
	},
	"Or": [
		{
			"Selection":{
				"IDs": [
					"one"
				]
			}
		}
	]
}`)

				err := json.Unmarshal(reqData, associatedQuery)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(associatedQuery.Validate()).ToNot(HaveOccurred())

				g.Expect(associatedQuery.Selection.KeyValues["key1"].Value).To(Equal(datatypes.Uint16(3)))
				g.Expect(associatedQuery.Selection.KeyValues["key1"].Comparison).To(Equal(v1.Equals))
				g.Expect(associatedQuery.Selection.KeyValues["key1"].TypeRestrictions).To(Equal(v1.TypeRestrictions{MinDataType: datatypes.T_uint16, MaxDataType: datatypes.T_any}))

				// check Or
				g.Expect(associatedQuery.Or[0].Selection.IDs).To(Equal([]string{"one"}))
			})
		})

		t.Run("Context Marshal", func(t *testing.T) {
			t.Run("It can properly encoded query", func(t *testing.T) {
				associatedQuery := &AssociatedActionQuery{
					Selection: &Selection{
						KeyValues: map[string]ValueQuery{
							"key1": {
								Value:      datatypes.Uint16(16),
								Comparison: v1.Equals,
								TypeRestrictions: v1.TypeRestrictions{
									MinDataType: datatypes.MinDataType,
									MaxDataType: datatypes.MaxDataType,
								},
							},
						},
					},
					Or: []*AssociatedActionQuery{
						{
							Selection: &Selection{
								IDs:                  []string{"one"},
								MinNumberOfKeyValues: helpers.PointerOf(2),
								MaxNumberOfKeyValues: helpers.PointerOf(3),
							},
						},
					},
				}

				g.Expect(associatedQuery.Validate()).ToNot(HaveOccurred())
				data, err := json.MarshalIndent(&associatedQuery, "", "	")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(string(data)).To(ContainSubstring(`{
	"Selection": {
		"KeyValues": {
			"key1": {
				"Value": {
					"Type": 2,
					"Data": "16"
				},
				"Comparison": "=",
				"TypeRestrictions": {
					"MinDataType": 1,
					"MaxDataType": 1024
				}
			}
		}
	},
	"Or": [
		{
			"Selection": {
				"IDs": [
					"one"
				],
				"MinNumberOfKeyValues": 2,
				"MaxNumberOfKeyValues": 3
			}
		}
	]
}`))
			})
		})
	})
}
