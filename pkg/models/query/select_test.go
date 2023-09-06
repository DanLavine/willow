package query

import (
	"testing"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_Select_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It accepts all empty values. This is a Select ALL", func(t *testing.T) {
		selection := &Select{}

		err := selection.Validate()
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Context when the WHERE clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{Where: &Query{Limits: &KeyLimits{}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the OR clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{Or: []Select{{Where: &Query{Limits: &KeyLimits{}}}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Or[0].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the AND clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			selection := &Select{And: []Select{{Where: &Query{Limits: &KeyLimits{}}}}}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Describe multi level joins", func(t *testing.T) {
		t.Run("It reports errors nicely", func(t *testing.T) {
			numberOfKeys := 5

			selection := &Select{
				And: []Select{
					{Or: []Select{
						{Where: &Query{Limits: &KeyLimits{NumberOfKeys: &numberOfKeys}}},
						{Where: &Query{Limits: &KeyLimits{}}},
					},
					},
				},
			}

			err := selection.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].Or[1].Where.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})
}

func Test_Select_Parse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can parse a JSON select", func(t *testing.T) {
		selection, err := ParseSelect([]byte(`{"Where": {"KeyValues": {"value1":{"Exists":true}}}}`))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*(selection.Where.KeyValues["value1"].Exists)).To(BeTrue())
	})
}

func Test_Select_MatchTags(t *testing.T) {
	g := NewGomegaWithT(t)

	isTrue := true
	isFalse := false

	t.Run("Context when only WHERE is set", func(t *testing.T) {
		t.Run("It returns false if the tags have more key values that the limits", func(t *testing.T) {
			numberOfKeys := 1
			selection := &Select{
				Where: &Query{
					Limits: &KeyLimits{
						NumberOfKeys: &numberOfKeys,
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			tags := datatypes.StringMap{
				"one": datatypes.String("1"),
				"two": datatypes.Float64(2.0),
			}

			g.Expect(selection.MatchTags(tags)).To(BeFalse())
		})

		t.Run("It joins all the clauses together with an AND", func(t *testing.T) {
			intOne := datatypes.Int(1)
			twoString := datatypes.String("2")
			selection := &Select{
				Where: &Query{
					KeyValues: map[string]Value{
						"one": {Value: &intOne, ValueComparison: EqualsPtr()},
						"two": {Value: &twoString, ValueComparison: NotEqualsPtr()},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			valdTags := datatypes.StringMap{
				"one": datatypes.Int(1),
				"two": datatypes.String("3"),
			}

			invaldTags := datatypes.StringMap{
				"one": datatypes.Int(1),
				"two": datatypes.String("2"),
			}

			g.Expect(selection.MatchTags(valdTags)).To(BeTrue())
			g.Expect(selection.MatchTags(invaldTags)).To(BeFalse())
		})

		t.Run("It fails if a provided query value does not exist in the tags", func(t *testing.T) {
			intOne := datatypes.Int(1)
			twoString := datatypes.String("2")
			selection := &Select{
				Where: &Query{
					KeyValues: map[string]Value{
						"one": {Value: &intOne, ValueComparison: EqualsPtr()},
						"two": {Value: &twoString, ValueComparison: NotEqualsPtr()},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			tags := datatypes.StringMap{
				"one": datatypes.String("1"),
			}

			g.Expect(selection.MatchTags(tags)).To(BeFalse())
		})

		t.Run("Context Exist checks", func(t *testing.T) {
			t.Run("Context Exists is true", func(t *testing.T) {
				t.Run("It returns true if the tags have the desired keys", func(t *testing.T) {
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Exists: &isTrue},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are missing the keys", func(t *testing.T) {
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"two": {Exists: &isTrue},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when setting the ExistsType check as well", func(t *testing.T) {
					t.Run("It returns true if the tags have the desired type", func(t *testing.T) {
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Exists: &isTrue, ExistsType: &datatypes.T_string},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.String("1"),
						}

						g.Expect(selection.MatchTags(tags)).To(BeTrue())
					})

					t.Run("It returns false if the tags have the desired type", func(t *testing.T) {
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Exists: &isTrue, ExistsType: &datatypes.T_int},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.String("1"),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context Exists is false", func(t *testing.T) {
				t.Run("It returns fals if the tags have the same keys", func(t *testing.T) {
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Exists: &isFalse},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns true if the tags are missing the keys", func(t *testing.T) {
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"two": {Exists: &isFalse},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("Context when setting the ExistsType check as well", func(t *testing.T) {
					t.Run("It returns false if the tags have the desired type", func(t *testing.T) {
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Exists: &isFalse, ExistsType: &datatypes.T_string},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.String("1"),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})

					t.Run("It returns true if the tags key does not have the desirred type", func(t *testing.T) {
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Exists: &isFalse, ExistsType: &datatypes.T_int},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.String("1"),
						}

						g.Expect(selection.MatchTags(tags)).To(BeTrue())
					})
				})
			})
		})

		t.Run("Context Value checks", func(t *testing.T) {
			t.Run("Context Equals", func(t *testing.T) {
				t.Run("It returns true if the tags have the exact key", func(t *testing.T) {
					oneValue := datatypes.String("1")
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags don't have the exact key", func(t *testing.T) {
					oneValue := datatypes.String("1")
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("2"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context NotEquals", func(t *testing.T) {
				t.Run("It returns false if the tags have the exact key", func(t *testing.T) {
					oneValue := datatypes.String("1")
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: NotEqualsPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("1"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns true if the tags don't have the exact key", func(t *testing.T) {
					oneValue := datatypes.String("1")
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: NotEqualsPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.String("2"),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("Context ValueTypeMatch check is true", func(t *testing.T) {
					t.Run("It ensures that the value has the same type as the Value", func(t *testing.T) {
						oneValue := datatypes.String("1")
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Value: &oneValue, ValueTypeMatch: &isTrue, ValueComparison: NotEqualsPtr()},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.Int(1),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context LessThan", func(t *testing.T) {
				t.Run("It returns true if the tags are less than the query", func(t *testing.T) {
					fiftyValue := datatypes.Int(50)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &fiftyValue, ValueComparison: LessThanPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater or equal to the query", func(t *testing.T) {
					zeroValue := datatypes.Int(0)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(2),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when ValueTypeMatch is true", func(t *testing.T) {
					t.Run("It ensures that the value must still match the provided Value", func(t *testing.T) {
						oneValue := datatypes.String("1")
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Value: &oneValue, ValueTypeMatch: &isTrue, ValueComparison: LessThanPtr()},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.Int(1),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context LessThanOrEqual", func(t *testing.T) {
				t.Run("It returns true if the tags are less than or Equal to the query", func(t *testing.T) {
					oneValue := datatypes.Int(1)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: LessThanOrEqualPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater than the quert", func(t *testing.T) {
					zeroValue := datatypes.Int(0)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanOrEqualPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when ValueTypeMatch is true", func(t *testing.T) {
					t.Run("It ensures that the value must still match the provided Value", func(t *testing.T) {
						oneValue := datatypes.String("1")
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Value: &oneValue, ValueTypeMatch: &isTrue, ValueComparison: LessThanOrEqualPtr()},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.Int(1),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context GreaterThan", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than the query", func(t *testing.T) {
					zeroValue := datatypes.Int(0)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less than the query", func(t *testing.T) {
					fiveValue := datatypes.Int(5)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &fiveValue, ValueComparison: GreaterThanPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when ValueTypeMatch is true", func(t *testing.T) {
					t.Run("It ensures that the value must still match the provided Value", func(t *testing.T) {
						oneValue := datatypes.Int(1)
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Value: &oneValue, ValueTypeMatch: &isTrue, ValueComparison: GreaterThanPtr()},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.Int64(3),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context GreaterThanOrEqual", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than or equal to the query", func(t *testing.T) {
					zeroValue := datatypes.Int(0)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanOrEqualPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(0),
					}

					g.Expect(selection.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less than the query", func(t *testing.T) {
					fiveValue := datatypes.Int(5)
					selection := &Select{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &fiveValue, ValueComparison: GreaterThanOrEqualPtr()},
							},
						},
					}
					g.Expect(selection.Validate()).ToNot(HaveOccurred())

					tags := datatypes.StringMap{
						"one": datatypes.Int(1),
					}

					g.Expect(selection.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when ValueTypeMatch is true", func(t *testing.T) {
					t.Run("It ensures that the value must still match the provided Value", func(t *testing.T) {
						oneValue := datatypes.Int(1)
						selection := &Select{
							Where: &Query{
								KeyValues: map[string]Value{
									"one": {Value: &oneValue, ValueTypeMatch: &isTrue, ValueComparison: GreaterThanOrEqualPtr()},
								},
							},
						}
						g.Expect(selection.Validate()).ToNot(HaveOccurred())

						tags := datatypes.StringMap{
							"one": datatypes.Int64(3),
						}

						g.Expect(selection.MatchTags(tags)).To(BeFalse())
					})
				})
			})
		})
	})

	t.Run("Context when only AND is set", func(t *testing.T) {
		t.Run("It returns true only if all checks pass", func(t *testing.T) {
			oneValue := datatypes.Int(1)
			numberOfKeys := 1
			selection := &Select{
				And: []Select{
					{
						Where: &Query{
							Limits: &KeyLimits{
								NumberOfKeys: &numberOfKeys,
							},
						},
					},
					{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			validTags := datatypes.StringMap{
				"one": datatypes.Int(1),
			}

			invalidTags := datatypes.StringMap{
				"one": datatypes.Int(1),
				"two": datatypes.Float64(2.0),
			}

			g.Expect(selection.MatchTags(validTags)).To(BeTrue())
			g.Expect(selection.MatchTags(invalidTags)).To(BeFalse())
		})
	})

	t.Run("Context when only OR is set", func(t *testing.T) {
		t.Run("It returns true if any checks pass", func(t *testing.T) {
			oneValue := datatypes.Int(1)
			numberOfKeys := 1
			selection := &Select{
				Or: []Select{
					{
						Where: &Query{
							Limits: &KeyLimits{
								NumberOfKeys: &numberOfKeys,
							},
						},
					},
					{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			validTags1 := datatypes.StringMap{
				"one": datatypes.Int(1),
			}

			validTags2 := datatypes.StringMap{
				"one": datatypes.Int(1),
				"two": datatypes.Float64(2.0),
			}

			g.Expect(selection.MatchTags(validTags1)).To(BeTrue())
			g.Expect(selection.MatchTags(validTags2)).To(BeTrue())
		})
	})

	t.Run("Context when WHERE and AND is set", func(t *testing.T) {
		t.Run("It returns true by performing a join on al the values", func(t *testing.T) {
			oneValue := datatypes.Int(1)
			numberOfKeys := 1
			selection := &Select{
				Where: &Query{
					Limits: &KeyLimits{
						NumberOfKeys: &numberOfKeys,
					},
				},
				And: []Select{
					{
						Where: &Query{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(selection.Validate()).ToNot(HaveOccurred())

			validTags := datatypes.StringMap{
				"one": datatypes.Int(1),
			}

			invalidTags := datatypes.StringMap{
				"one": datatypes.Int(1),
				"two": datatypes.Float64(2.0),
			}

			g.Expect(selection.MatchTags(validTags)).To(BeTrue())
			g.Expect(selection.MatchTags(invalidTags)).To(BeFalse())
		})
	})
}
