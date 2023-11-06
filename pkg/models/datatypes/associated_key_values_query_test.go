package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_AssociatedKeyValuesQuery_Validate(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It accepts all empty values. This is a Query ALL", func(t *testing.T) {
		query := &AssociatedKeyValuesQuery{}

		err := query.Validate()
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Context when the WHERE clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			query := &AssociatedKeyValuesQuery{KeyValueSelection: &KeyValueSelection{Limits: &KeyLimits{}}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("KeyValueSelection.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the OR clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			query := &AssociatedKeyValuesQuery{Or: []AssociatedKeyValuesQuery{{KeyValueSelection: &KeyValueSelection{Limits: &KeyLimits{}}}}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("Or[0].KeyValueSelection.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Context when the AND clause is bad", func(t *testing.T) {
		t.Run("It propigates the error", func(t *testing.T) {
			query := &AssociatedKeyValuesQuery{And: []AssociatedKeyValuesQuery{{KeyValueSelection: &KeyValueSelection{Limits: &KeyLimits{}}}}}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].KeyValueSelection.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})

	t.Run("Describe multi level joins", func(t *testing.T) {
		t.Run("It reports errors nicely", func(t *testing.T) {
			numberOfKeys := 5

			query := &AssociatedKeyValuesQuery{
				And: []AssociatedKeyValuesQuery{
					{Or: []AssociatedKeyValuesQuery{
						{KeyValueSelection: &KeyValueSelection{Limits: &KeyLimits{NumberOfKeys: &numberOfKeys}}},
						{KeyValueSelection: &KeyValueSelection{Limits: &KeyLimits{}}},
					},
					},
				},
			}

			err := query.Validate()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(Equal("And[0].Or[1].KeyValueSelection.Limits.NumberOfKeys: Requires an int to be provided"))
		})
	})
}

func Test_AssociatedKeyValuesQuery_Parse(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can parse a valis JSON Where", func(t *testing.T) {
		query, err := ParseAssociatedKeyValuesQuery([]byte(`{"KeyValueSelection": {"KeyValues": {"value1":{"Exists":true}}}}`))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*(query.KeyValueSelection.KeyValues["value1"].Exists)).To(BeTrue())
	})
}

func Test_AssociatedKeyValuesQuery_MatchTags(t *testing.T) {
	g := NewGomegaWithT(t)

	isTrue := true
	isFalse := false

	t.Run("Context when only WHERE is set", func(t *testing.T) {
		t.Run("It returns false if the tags have more key values that the limits", func(t *testing.T) {
			numberOfKeys := 1
			query := &AssociatedKeyValuesQuery{
				KeyValueSelection: &KeyValueSelection{
					Limits: &KeyLimits{
						NumberOfKeys: &numberOfKeys,
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			tags := KeyValues{
				"one": String("1"),
				"two": Float64(2.0),
			}

			g.Expect(query.MatchTags(tags)).To(BeFalse())
		})

		t.Run("It joins all the clauses together with an AND", func(t *testing.T) {
			intOne := Int(1)
			twoString := String("2")
			query := &AssociatedKeyValuesQuery{
				KeyValueSelection: &KeyValueSelection{
					KeyValues: map[string]Value{
						"one": {Value: &intOne, ValueComparison: EqualsPtr()},
						"two": {Value: &twoString, ValueComparison: NotEqualsPtr()},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			valdTags := KeyValues{
				"one": Int(1),
				"two": String("3"),
			}

			invaldTags := KeyValues{
				"one": Int(1),
				"two": String("2"),
			}

			g.Expect(query.MatchTags(valdTags)).To(BeTrue())
			g.Expect(query.MatchTags(invaldTags)).To(BeFalse())
		})

		t.Run("It fails if a provided query value does not exist in the tags", func(t *testing.T) {
			intOne := Int(1)
			twoString := String("2")
			query := &AssociatedKeyValuesQuery{
				KeyValueSelection: &KeyValueSelection{
					KeyValues: map[string]Value{
						"one": {Value: &intOne, ValueComparison: EqualsPtr()},
						"two": {Value: &twoString, ValueComparison: NotEqualsPtr()},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			tags := KeyValues{
				"one": String("1"),
			}

			g.Expect(query.MatchTags(tags)).To(BeFalse())
		})

		t.Run("Context Exist checks", func(t *testing.T) {
			t.Run("Context Exists is true", func(t *testing.T) {
				t.Run("It returns true if the tags have the desired keys", func(t *testing.T) {
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Exists: &isTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are missing the keys", func(t *testing.T) {
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"two": {Exists: &isTrue},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("Context when setting the ExistsType check as well", func(t *testing.T) {
					t.Run("It returns true if the tags have the desired type", func(t *testing.T) {
						query := &AssociatedKeyValuesQuery{
							KeyValueSelection: &KeyValueSelection{
								KeyValues: map[string]Value{
									"one": {Exists: &isTrue, ExistsType: &T_string},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						tags := KeyValues{
							"one": String("1"),
						}

						g.Expect(query.MatchTags(tags)).To(BeTrue())
					})

					t.Run("It returns false if the tags have the desired type", func(t *testing.T) {
						query := &AssociatedKeyValuesQuery{
							KeyValueSelection: &KeyValueSelection{
								KeyValues: map[string]Value{
									"one": {Exists: &isTrue, ExistsType: &T_int},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						tags := KeyValues{
							"one": String("1"),
						}

						g.Expect(query.MatchTags(tags)).To(BeFalse())
					})
				})
			})

			t.Run("Context Exists is false", func(t *testing.T) {
				t.Run("It returns fals if the tags have the same keys", func(t *testing.T) {
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Exists: &isFalse},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns true if the tags are missing the keys", func(t *testing.T) {
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"two": {Exists: &isFalse},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("Context when setting the ExistsType check as well", func(t *testing.T) {
					t.Run("It returns false if the tags have the desired type", func(t *testing.T) {
						query := &AssociatedKeyValuesQuery{
							KeyValueSelection: &KeyValueSelection{
								KeyValues: map[string]Value{
									"one": {Exists: &isFalse, ExistsType: &T_string},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						tags := KeyValues{
							"one": String("1"),
						}

						g.Expect(query.MatchTags(tags)).To(BeFalse())
					})

					t.Run("It returns true if the tags key does not have the desirred type", func(t *testing.T) {
						query := &AssociatedKeyValuesQuery{
							KeyValueSelection: &KeyValueSelection{
								KeyValues: map[string]Value{
									"one": {Exists: &isFalse, ExistsType: &T_int},
								},
							},
						}
						g.Expect(query.Validate()).ToNot(HaveOccurred())

						tags := KeyValues{
							"one": String("1"),
						}

						g.Expect(query.MatchTags(tags)).To(BeTrue())
					})
				})
			})
		})

		t.Run("Context Value checks", func(t *testing.T) {
			t.Run("Context Equals", func(t *testing.T) {
				t.Run("It returns true if the tags have the exact key", func(t *testing.T) {
					oneValue := String("1")
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags don't have the exact key", func(t *testing.T) {
					oneValue := String("1")
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("2"),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context NotEquals", func(t *testing.T) {
				t.Run("It returns false if the tags have the exact key value", func(t *testing.T) {
					oneValue := String("1")
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: NotEqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("1"),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns true if the tags don't have the exact key value", func(t *testing.T) {
					oneValue := String("1")
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: NotEqualsPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": String("2"),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})
			})

			t.Run("Context LessThan", func(t *testing.T) {
				t.Run("It returns true if the tags are less than the query", func(t *testing.T) {
					fiftyValue := Int(50)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiftyValue, ValueComparison: LessThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater or equal to the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context LessThanMatchType", func(t *testing.T) {
				t.Run("It returns true if the tags are less than the query", func(t *testing.T) {
					fiftyValue := Int(50)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiftyValue, ValueComparison: LessThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(20),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater or equal to the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns false if the Values type do not match", func(t *testing.T) {
					// value type less than
					zeroValue := Int64(0)
					query1 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query1.Validate()).ToNot(HaveOccurred())

					// value type greater than
					floatZeroValue := Float32(0.0)
					query2 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &floatZeroValue, ValueComparison: LessThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query2.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query1.MatchTags(tags)).To(BeFalse())
					g.Expect(query2.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context LessThanOrEqual", func(t *testing.T) {
				t.Run("It returns true if the tags are less than or Equal to the query", func(t *testing.T) {
					oneValue := Int(1)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: LessThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater than the quert", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context LessThanOrEqualMatchType", func(t *testing.T) {
				t.Run("It returns true if the tags are less than or equalt than the query", func(t *testing.T) {
					fiftyValue := Int(50)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiftyValue, ValueComparison: LessThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(50),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the are greater or equal to the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns false if the Values type do not match", func(t *testing.T) {
					// value type less than
					zeroValue := Int64(0)
					query1 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: LessThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query1.Validate()).ToNot(HaveOccurred())

					// value type greater than
					floatZeroValue := Float32(0.0)
					query2 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &floatZeroValue, ValueComparison: LessThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query2.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query1.MatchTags(tags)).To(BeFalse())
					g.Expect(query2.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context GreaterThan", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less than the query", func(t *testing.T) {
					fiveValue := Int(5)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiveValue, ValueComparison: GreaterThanPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context GreaterThanMatchType", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(50),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less or equal to the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(0),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns false if the Values type do not match", func(t *testing.T) {
					// value type less than
					zeroValue := Int64(0)
					query1 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query1.Validate()).ToNot(HaveOccurred())

					// value type greater than
					floatZeroValue := Float32(0.0)
					query2 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &floatZeroValue, ValueComparison: GreaterThanMatchTypePtr()},
							},
						},
					}
					g.Expect(query2.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query1.MatchTags(tags)).To(BeFalse())
					g.Expect(query2.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context GreaterThanOrEqual", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than or equal to the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(0),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less than the query", func(t *testing.T) {
					fiveValue := Int(5)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiveValue, ValueComparison: GreaterThanOrEqualPtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(1),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})
			})

			t.Run("Context GreaterThanOrEqualMatchType", func(t *testing.T) {
				t.Run("It returns true if the tags are greater than or equal the query", func(t *testing.T) {
					fiftyValue := Int(50)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &fiftyValue, ValueComparison: GreaterThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(50),
					}

					g.Expect(query.MatchTags(tags)).To(BeTrue())
				})

				t.Run("It returns false if the tags are less than the query", func(t *testing.T) {
					zeroValue := Int(0)
					query := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(-1),
					}

					g.Expect(query.MatchTags(tags)).To(BeFalse())
				})

				t.Run("It returns false if the Values type do not match", func(t *testing.T) {
					// value type less than
					zeroValue := Int64(0)
					query1 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &zeroValue, ValueComparison: GreaterThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query1.Validate()).ToNot(HaveOccurred())

					// value type greater than
					floatZeroValue := Float32(0.0)
					query2 := &AssociatedKeyValuesQuery{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &floatZeroValue, ValueComparison: GreaterThanOrEqualMatchTypePtr()},
							},
						},
					}
					g.Expect(query2.Validate()).ToNot(HaveOccurred())

					tags := KeyValues{
						"one": Int(2),
					}

					g.Expect(query1.MatchTags(tags)).To(BeFalse())
					g.Expect(query2.MatchTags(tags)).To(BeFalse())
				})
			})

		})
	})

	t.Run("Context when only AND is set", func(t *testing.T) {
		t.Run("It returns true only if all checks pass", func(t *testing.T) {
			oneValue := Int(1)
			numberOfKeys := 1
			query := &AssociatedKeyValuesQuery{
				And: []AssociatedKeyValuesQuery{
					{
						KeyValueSelection: &KeyValueSelection{
							Limits: &KeyLimits{
								NumberOfKeys: &numberOfKeys,
							},
						},
					},
					{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			validTags := KeyValues{
				"one": Int(1),
			}

			invalidTags := KeyValues{
				"one": Int(1),
				"two": Float64(2.0),
			}

			g.Expect(query.MatchTags(validTags)).To(BeTrue())
			g.Expect(query.MatchTags(invalidTags)).To(BeFalse())
		})
	})

	t.Run("Context when only OR is set", func(t *testing.T) {
		t.Run("It returns true if any checks pass", func(t *testing.T) {
			oneValue := Int(1)
			numberOfKeys := 1
			query := &AssociatedKeyValuesQuery{
				Or: []AssociatedKeyValuesQuery{
					{
						KeyValueSelection: &KeyValueSelection{
							Limits: &KeyLimits{
								NumberOfKeys: &numberOfKeys,
							},
						},
					},
					{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			validTags1 := KeyValues{
				"one": Int(1),
			}

			validTags2 := KeyValues{
				"one": Int(1),
				"two": Float64(2.0),
			}

			g.Expect(query.MatchTags(validTags1)).To(BeTrue())
			g.Expect(query.MatchTags(validTags2)).To(BeTrue())
		})
	})

	t.Run("Context when WHERE and AND is set", func(t *testing.T) {
		t.Run("It returns true by performing a join on al the values", func(t *testing.T) {
			oneValue := Int(1)
			numberOfKeys := 1
			query := &AssociatedKeyValuesQuery{
				KeyValueSelection: &KeyValueSelection{
					Limits: &KeyLimits{
						NumberOfKeys: &numberOfKeys,
					},
				},
				And: []AssociatedKeyValuesQuery{
					{
						KeyValueSelection: &KeyValueSelection{
							KeyValues: map[string]Value{
								"one": {Value: &oneValue, ValueComparison: EqualsPtr()},
							},
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			validTags := KeyValues{
				"one": Int(1),
			}

			invalidTags := KeyValues{
				"one": Int(1),
				"two": Float64(2.0),
			}

			g.Expect(query.MatchTags(validTags)).To(BeTrue())
			g.Expect(query.MatchTags(invalidTags)).To(BeFalse())
		})
	})
}
