package queryassociatedaction

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func PointerOf[T any](value T) *T {
	return &value
}

func Test_Selection_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if all the fields are nil", func(t *testing.T) {
		selection := Selection{}

		err := selection.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("requires 'IDs', 'KeyValues', 'MinNumberOfKeyValues' or 'MaxNumberOfKeyValues' to be specified, but received nothing"))
	})

	t.Run("It returns an error if MaxNumberOfKeyValues is less than MinNumberOfKeyValues", func(t *testing.T) {
		selection := Selection{
			MinNumberOfKeyValues: PointerOf(12),
			MaxNumberOfKeyValues: PointerOf(2),
		}

		err := selection.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("MaxNumberOfKeyValues is less than MinNumberOfKeyValues"))

	})

	t.Run("It allows a query of just AssociatedIDs", func(t *testing.T) {
		selection := Selection{
			IDs: []string{"id_one"},
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())
	})

	t.Run("It allows a query of just KeyValues", func(t *testing.T) {
		selection := Selection{
			KeyValues: map[string]ValueQuery{
				"key1": {
					Value:      datatypes.Float64(6.4),
					Comparison: v1.Equals,
					TypeRestrictions: v1.TypeRestrictions{
						MinDataType: datatypes.MinDataType,
						MaxDataType: datatypes.MaxDataType,
					},
				},
			},
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())
	})

	t.Run("It allows a query of just MinNumberOfKeyValues", func(t *testing.T) {
		selection := Selection{
			MinNumberOfKeyValues: PointerOf(3),
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())
	})

	t.Run("It allows a query of just MaxNumberOfKeyValues", func(t *testing.T) {
		selection := Selection{
			MaxNumberOfKeyValues: PointerOf(3),
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())
	})

	t.Run("It allows a query of all values", func(t *testing.T) {
		selection := Selection{
			IDs: []string{"id_one"},
			KeyValues: map[string]ValueQuery{
				"key1": {
					Value:      datatypes.Float64(6.4),
					Comparison: v1.Equals,
					TypeRestrictions: v1.TypeRestrictions{
						MinDataType: datatypes.MinDataType,
						MaxDataType: datatypes.MaxDataType,
					},
				},
			},
			MinNumberOfKeyValues: PointerOf(2),
			MaxNumberOfKeyValues: PointerOf(5),
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())
	})
}

func Test_Selection_Keys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an empty list if there are no KeyValues", func(t *testing.T) {
		selection := &Selection{
			MinNumberOfKeyValues: PointerOf(3),
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keys := selection.Keys()
		g.Expect(keys).To(Equal([]string{}))
	})

	t.Run("It all Keys of the KeyValues", func(t *testing.T) {
		selection := &Selection{
			KeyValues: map[string]ValueQuery{
				"key1": {Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
				"key2": {Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
				"key3": {Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
			},
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keys := selection.Keys()
		g.Expect(keys).To(ContainElements("key1", "key2", "key3"))
	})
}

func Test_Selection_SortedKeys(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an empty list if there are no KeyValues", func(t *testing.T) {
		selection := &Selection{
			MinNumberOfKeyValues: PointerOf(3),
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keys := selection.SortedKeys()
		g.Expect(keys).To(Equal([]string{}))
	})

	t.Run("It all Keys of the KeyValues in a sorted order", func(t *testing.T) {
		selection := &Selection{
			KeyValues: map[string]ValueQuery{
				"key1": {Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
				"key2": ValueQuery{Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
				"key3": ValueQuery{Value: datatypes.Any(), Comparison: v1.Equals, TypeRestrictions: v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}},
			},
		}

		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keys := selection.SortedKeys()
		g.Expect(keys).To(Equal([]string{"key1", "key2", "key3"}))
	})
}

func Test_Selection_MatchKeyValues(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if the MinNumberOfKeyValues are not met", func(t *testing.T) {
		selection := &Selection{
			MinNumberOfKeyValues: PointerOf(3),
		}
		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keyValues := datatypes.KeyValues{
			"key1": datatypes.Any(),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		matched := selection.MatchKeyValues(keyValues)
		g.Expect(matched).To(BeFalse())
	})

	t.Run("It returns false if over the MaxNumberOfKeyValues", func(t *testing.T) {
		selection := &Selection{
			MaxNumberOfKeyValues: PointerOf(1),
		}
		g.Expect(selection.Validate()).ToNot(HaveOccurred())

		keyValues := datatypes.KeyValues{
			"key1": datatypes.Any(),
			"key2": datatypes.String("boo"),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		matched := selection.MatchKeyValues(keyValues)
		g.Expect(matched).To(BeFalse())
	})

	t.Run("Describe Equals query", func(t *testing.T) {
		t.Run("It returns false if the key is not in the match request", func(t *testing.T) {
			selection := &Selection{
				KeyValues: map[string]ValueQuery{
					"key1": {
						Value:      datatypes.String("str value"),
						Comparison: v1.Equals,
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.T_string,
							MaxDataType: datatypes.T_string,
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			matchKeyValues := datatypes.KeyValues{
				"key2": datatypes.Int(3),
			}
			g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
			g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())
		})

		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns an error if the key is outside the allowed values", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.String("str value"),
							Comparison: v1.Equals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_string,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				keyValuesGeneral := datatypes.KeyValues{
					"key1": datatypes.Int(3),
				}
				g.Expect(keyValuesGeneral.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(keyValuesGeneral)).To(BeFalse())

				keyValuesAny := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(keyValuesAny.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(keyValuesAny)).To(BeFalse())
			})

			t.Run("It accepts an exact value comparison", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.String("str value"),
							Comparison: v1.Equals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_string,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				keyValues := datatypes.KeyValues{
					"key1": datatypes.String("str value"),
				}
				g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				matched := selection.MatchKeyValues(keyValues)
				g.Expect(matched).To(BeTrue())
			})
		})

		t.Run("Context when the tree value is a T_any", func(t *testing.T) {
			t.Run("It returns an error if the key is outside the allowed values", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Any(),
							Comparison: v1.Equals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_string,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				keyValuesGeneral := datatypes.KeyValues{
					"key1": datatypes.Int(3),
				}
				g.Expect(keyValuesGeneral.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(keyValuesGeneral)).To(BeFalse())

				keyValuesAny := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(keyValuesAny.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(keyValuesAny)).To(BeFalse())
			})

			t.Run("It accepts an exact value comparison", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Any(),
							Comparison: v1.Equals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_string,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				keyValues := datatypes.KeyValues{
					"key1": datatypes.String("str value"),
				}
				g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				matched := selection.MatchKeyValues(keyValues)
				g.Expect(matched).To(BeTrue())
			})
		})
	})

	t.Run("Describe NotEquals query", func(t *testing.T) {
		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns false if they query and match request are the same", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int(3),
							Comparison: v1.NotEquals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_int,
								MaxDataType: datatypes.T_int,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int(3),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts any other general value", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.String("str value"),
							Comparison: v1.NotEquals,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.T_string,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				keyValues := datatypes.KeyValues{
					"key1": datatypes.String("boop"),
				}
				g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(keyValues)).To(BeTrue())
			})
		})

		t.Run("Context when using T_any", func(t *testing.T) {
			t.Run("Context when the query has T_any", func(t *testing.T) {
				t.Run("It can serch for just T_any specificly with strict TypeRestrictions", func(t *testing.T) {
					selection := &Selection{
						KeyValues: map[string]ValueQuery{
							"key1": {
								Value:      datatypes.Any(),
								Comparison: v1.NotEquals,
								TypeRestrictions: v1.TypeRestrictions{
									MinDataType: datatypes.T_any,
									MaxDataType: datatypes.T_any,
								},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					keyValues := datatypes.KeyValues{
						"key1": datatypes.String("boop"),
					}
					g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
					g.Expect(selection.MatchKeyValues(keyValues)).To(BeTrue())
				})

				// This is the test to ensure the key does not exist
				t.Run("It returns false if the matchQuery in the restricted range", func(t *testing.T) {
					selection := &Selection{
						KeyValues: map[string]ValueQuery{
							"key1": {
								Value:      datatypes.Any(),
								Comparison: v1.NotEquals,
								TypeRestrictions: v1.TypeRestrictions{
									MinDataType: datatypes.MinDataType,
									MaxDataType: datatypes.MaxDataType,
								},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					keyValues := datatypes.KeyValues{
						"key1": datatypes.String("boop"),
					}
					g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
					g.Expect(selection.MatchKeyValues(keyValues)).To(BeFalse())
				})
			})

			t.Run("Context when the match KeyValues has T_any", func(t *testing.T) {
				t.Run("It allows matchQuery T_any if the query TypeRestrictions are strict", func(t *testing.T) {
					selection := &Selection{
						KeyValues: map[string]ValueQuery{
							"key1": {
								Value:      datatypes.String("something"),
								Comparison: v1.NotEquals,
								TypeRestrictions: v1.TypeRestrictions{
									MinDataType: datatypes.T_string,
									MaxDataType: datatypes.T_string,
								},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					keyValues := datatypes.KeyValues{
						"key1": datatypes.Any(),
					}
					g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
					g.Expect(selection.MatchKeyValues(keyValues)).To(BeTrue())
				})

				// This is the test to ensure the key does not exist
				t.Run("It returns false if the matchQuery T_any is in the TypeRestrictions", func(t *testing.T) {
					selection := &Selection{
						KeyValues: map[string]ValueQuery{
							"key1": {
								Value:      datatypes.String("ok"),
								Comparison: v1.NotEquals,
								TypeRestrictions: v1.TypeRestrictions{
									MinDataType: datatypes.T_string,
									MaxDataType: datatypes.T_any,
								},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					keyValues := datatypes.KeyValues{
						"key1": datatypes.Any(),
					}
					g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
					g.Expect(selection.MatchKeyValues(keyValues)).To(BeFalse())
				})
			})

		})
	})

	t.Run("Describe LessThan query", func(t *testing.T) {
		t.Run("It returns false if matchKeyValues does not contain the query key", func(t *testing.T) {
			selection := &Selection{
				KeyValues: map[string]ValueQuery{
					"key1": {
						Value:      datatypes.Int16(8),
						Comparison: v1.LessThan,
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.MinDataType,
							MaxDataType: datatypes.MaxDataType,
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			matchKeyValues := datatypes.KeyValues{
				"key2": datatypes.Int(3),
			}
			g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
			g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())

		})

		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns false if the matchKeyValues is GreaterThan or equal to the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int16(8),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues when less than the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int8(12),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It allows matchKeyValues of T_any if type restrictions allow it", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_any,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It returns false matchKeyValues of T_any if type restrictions are strict", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})
		})

		t.Run("Context when the tree matchKeyValues are of T_any", func(t *testing.T) {
			t.Run("It returns false if they type restrictions are strict", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues if type restrictins allow it", func(t *testing.T) {
				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})
		})
	})

	t.Run("Describe LessThanOrEqual query", func(t *testing.T) {
		t.Run("It returns false if matchKeyValues does not contain the query key", func(t *testing.T) {
			selection := &Selection{
				KeyValues: map[string]ValueQuery{
					"key1": {
						Value:      datatypes.Int16(8),
						Comparison: v1.LessThanOrEqual,
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.MinDataType,
							MaxDataType: datatypes.MaxDataType,
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			matchKeyValues := datatypes.KeyValues{
				"key2": datatypes.Int(3),
			}
			g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
			g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())

		})

		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns false if the matchKeyValues is GreaterThan the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int16(9),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues when less than or equal to the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int8(8),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It allows matchKeyValues of T_any if type restrictions allow it", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_any,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It returns false matchKeyValues of T_any if type restrictions are strict", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})
		})

		t.Run("Context when the tree matchKeyValues are of T_any", func(t *testing.T) {
			t.Run("It returns false if they type restrictions are strict", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues if type restrictins allow it", func(t *testing.T) {
				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.LessThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})
		})
	})

	t.Run("Describe GreaterThan query", func(t *testing.T) {
		t.Run("It returns false if matchKeyValues does not contain the query key", func(t *testing.T) {
			selection := &Selection{
				KeyValues: map[string]ValueQuery{
					"key1": {
						Value:      datatypes.Int16(8),
						Comparison: v1.GreaterThan,
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.MinDataType,
							MaxDataType: datatypes.MaxDataType,
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			matchKeyValues := datatypes.KeyValues{
				"key2": datatypes.Int(3),
			}
			g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
			g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())

		})

		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns false if the matchKeyValues is LessThan or equal to the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int16(8),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues when greater than the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int(12),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It allows matchKeyValues of T_any if type restrictions allow it", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_any,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})

			t.Run("It returns false if matchKeyValues T_any is not allowed in the query restrictions", func(t *testing.T) {
				selection := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_string,
							},
						},
					},
				}
				g.Expect(selection.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})
		})

		t.Run("Context when the tree matchKeyValues are of T_any", func(t *testing.T) {
			t.Run("It returns false if they type restrictions are strict", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues if type restrictins allow it", func(t *testing.T) {
				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThan,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})
		})
	})

	t.Run("Describe GreaterThanOrEqual query", func(t *testing.T) {
		t.Run("It returns false if matchKeyValues does not contain the query key", func(t *testing.T) {
			selection := &Selection{
				KeyValues: map[string]ValueQuery{
					"key1": {
						Value:      datatypes.Int16(8),
						Comparison: v1.GreaterThanOrEqual,
						TypeRestrictions: v1.TypeRestrictions{
							MinDataType: datatypes.MinDataType,
							MaxDataType: datatypes.MaxDataType,
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			matchKeyValues := datatypes.KeyValues{
				"key2": datatypes.Int(3),
			}
			g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
			g.Expect(selection.MatchKeyValues(matchKeyValues)).To(BeFalse())

		})

		t.Run("Context when the tree value is a general type", func(t *testing.T) {
			t.Run("It returns false if the matchKeyValues is lessThan the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int16(7),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues when greater than or equal to the queryKey", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Int16(8),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})
		})

		t.Run("Context when the tree matchKeyValues are of T_any", func(t *testing.T) {
			t.Run("It returns false if they type restrictions are strict", func(t *testing.T) {
				selectionRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.T_int16,
							},
						},
					},
				}
				g.Expect(selectionRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionRestrictions.MatchKeyValues(matchKeyValues)).To(BeFalse())
			})

			t.Run("It accepts the matchKeyValues if type restrictins allow it", func(t *testing.T) {
				selectionNoRestrictions := &Selection{
					KeyValues: map[string]ValueQuery{
						"key1": {
							Value:      datatypes.Int16(8),
							Comparison: v1.GreaterThanOrEqual,
							TypeRestrictions: v1.TypeRestrictions{
								MinDataType: datatypes.MinDataType,
								MaxDataType: datatypes.MaxDataType,
							},
						},
					},
				}
				g.Expect(selectionNoRestrictions.Validate()).ToNot(HaveOccurred())

				matchKeyValues := datatypes.KeyValues{
					"key1": datatypes.Any(),
				}
				g.Expect(matchKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
				g.Expect(selectionNoRestrictions.MatchKeyValues(matchKeyValues)).To(BeTrue())
			})
		})
	})
}
