package overrides

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/internal/datastructures/btree_one_to_many/btreeonetomanyfakes"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers"
	"github.com/DanLavine/willow/testhelpers/testmodels"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func Test_OverrideClientLocal_CreateOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It can create a new override", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		override := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override.ValidateSpecOnly()).ToNot(HaveOccurred())

		_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an error if the override KeyValues are already taken", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		override1 := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override1.ValidateSpecOnly()).ToNot(HaveOccurred())

		override2 := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override1.ValidateSpecOnly()).ToNot(HaveOccurred())

		_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override1)
		g.Expect(err).ToNot(HaveOccurred())

		_, err = overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override2)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("override KeyValues already exists"))
	})

	t.Run("It returns an error if the override is currently being destroyed", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().CreateOrFind(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", btreeonetomany.ErrorManyIDDestroying).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)

		override := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override.ValidateSpecOnly()).ToNot(HaveOccurred())

		_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("override is being destroy"))
	})

	t.Run("It returns an internal server error for everything else", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().CreateOrFind(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", fmt.Errorf("nope")).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)

		override := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override.ValidateSpecOnly()).ToNot(HaveOccurred())

		_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("Internal Server Error"))
	})
}

func Test_OverrideClientLocal_GetOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns a not found error if the override doesn't exist", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		override, err := overrideClient.GetOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(override).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("Override 'test override' not found"))
	})

	t.Run("It returns the override if found", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		override := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(override.ValidateSpecOnly()).ToNot(HaveOccurred())
		overrideID, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", override)
		g.Expect(err).ToNot(HaveOccurred())

		override, err = overrideClient.GetOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideID)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(override).ToNot(BeNil())
		g.Expect(override.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Int(1)}))
		g.Expect(*override.Spec.Properties.Limit).To(Equal(int64(3)))
	})

	t.Run("It returns an internal server error if the tree returns an error", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().QueryAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("nope")).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)

		override, err := overrideClient.GetOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(err).To(Equal(errors.InternalServerError))
		g.Expect(override).To(BeNil())
	})
}

func Test_OverrideClientLocal_UpdateOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It returns an error if the override was not found", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		updateReq := &v1limiter.OverrideProperties{
			Limit: helpers.PointerOf[int64](17),
		}
		g.Expect(updateReq.Validate()).ToNot(HaveOccurred())

		err := overrideClient.UpdateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override", updateReq)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("Override 'test override' not found"))
	})

	t.Run("It can update the found override", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		overrideReq := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(overrideReq.ValidateSpecOnly()).ToNot(HaveOccurred())
		overrideID, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideReq)
		g.Expect(err).ToNot(HaveOccurred())

		updateReq := &v1limiter.OverrideProperties{
			Limit: helpers.PointerOf[int64](17),
		}
		g.Expect(updateReq.Validate()).ToNot(HaveOccurred())

		err = overrideClient.UpdateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideID, updateReq)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure override was updated
		override, err := overrideClient.GetOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideID)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(override).ToNot(BeNil())
		g.Expect(override.Spec.DBDefinition.GroupByKeyValues).To(Equal(datatypes.KeyValues{"key1": datatypes.Int(1)}))
		g.Expect(*override.Spec.Properties.Limit).To(Equal(int64(17)))
	})

	t.Run("It returns an internal server error if the tree returns an error", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().QueryAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("nope")).Times(1)

		updateReq := &v1limiter.OverrideProperties{
			Limit: helpers.PointerOf[int64](17),
		}
		g.Expect(updateReq.Validate()).ToNot(HaveOccurred())

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
		err := overrideClient.UpdateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override", updateReq)
		g.Expect(err).To(Equal(errors.InternalServerError))
	})
}

func Test_OverrideClientLocal_QueryOverrides(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	setupTree := func(g *GomegaWithT) *overridesClientLocal {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		for i := 0; i < 50; i++ {
			overrideReq := &v1limiter.Override{
				Spec: &v1limiter.OverrideSpec{
					DBDefinition: &v1limiter.OverrideDBDefinition{},
					Properties: &v1limiter.OverrideProperties{
						Limit: helpers.PointerOf[int64](3),
					},
				},
			}

			if i%2 == 0 {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i)}
			} else {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}
			}

			g.Expect(overrideReq.ValidateSpecOnly()).ToNot(HaveOccurred())
			_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideReq)
			g.Expect(err).ToNot(HaveOccurred())
		}

		return overrideClient
	}

	t.Run("It returns empty overrides if there are no found items", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		query := &queryassociatedaction.AssociatedActionQuery{} // match all
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(overrides).ToNot(BeNil())
		g.Expect(len(overrides)).To(Equal(0))
	})

	t.Run("Context MatchAll", func(t *testing.T) {
		t.Run("It reutnrs all overrides", func(t *testing.T) {
			overrideClient := setupTree(g)

			query := &queryassociatedaction.AssociatedActionQuery{} // match all
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(overrides)).To(Equal(50))
		})

		t.Run("It returns an internal server error if the tree returns an error", func(t *testing.T) {
			mockController := gomock.NewController(t)
			mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
			defer mockController.Finish()

			mockBTreeOneToMany.EXPECT().QueryAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("nope")).Times(1)

			query := &queryassociatedaction.AssociatedActionQuery{} // match all
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
			overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
			g.Expect(err).To(Equal(errors.InternalServerError))
			g.Expect(len(overrides)).To(Equal(0))
		})
	})

	t.Run("Context Query", func(t *testing.T) {
		t.Run("It reutnrs any overrides for the query", func(t *testing.T) {
			overrideClient := setupTree(g)

			query := &queryassociatedaction.AssociatedActionQuery{
				Selection: &queryassociatedaction.Selection{
					KeyValues: queryassociatedaction.SelectionKeyValues{
						"key0": {
							Value:            datatypes.Int(0),
							Comparison:       v1common.Equals,
							TypeRestrictions: testmodels.NoTypeRestrictions(g),
						},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(overrides)).To(Equal(1)) // pairs of groups{ {0} }
		})

		t.Run("It returns an internal server error if the tree returns an error", func(t *testing.T) {
			mockController := gomock.NewController(t)
			mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
			defer mockController.Finish()

			mockBTreeOneToMany.EXPECT().QueryAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("nope")).Times(1)

			query := &queryassociatedaction.AssociatedActionQuery{}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
			overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
			g.Expect(err).To(Equal(errors.InternalServerError))
			g.Expect(len(overrides)).To(Equal(0))
		})
	})
}

func Test_OverrideClientLocal_DestroyOverride(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It runs a no-op if the override canno be found", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		err := overrideClient.DestroyOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It destroys an override", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		overrideReq := &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: datatypes.KeyValues{
						"key1": datatypes.Int(1),
					},
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf[int64](3),
				},
			},
		}
		g.Expect(overrideReq.ValidateSpecOnly()).ToNot(HaveOccurred())
		overrideID, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideReq)
		g.Expect(err).ToNot(HaveOccurred())

		err = overrideClient.DestroyOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideID)
		g.Expect(err).ToNot(HaveOccurred())

		override, err := overrideClient.GetOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(err).To(HaveOccurred())
		g.Expect(override).To(BeNil())
	})

	t.Run("It returns nil if the override is currently being destroyed", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().DestroyOneOfManyByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(btreeonetomany.ErrorManyIDDestroying).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
		err := overrideClient.DestroyOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns an internal server error for everything else", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().DestroyOneOfManyByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("deleting")).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
		err := overrideClient.DestroyOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", "test override")
		g.Expect(err).To(Equal(errors.InternalServerError))
	})
}

func Test_OverrideClientLocal_DestroyOverrides(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("It runs a no-op if the override relation cannot be found", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		err := overrideClient.DestroyOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It destroys an all overrides associated with the rule name", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		for i := 0; i < 50; i++ {
			overrideReq := &v1limiter.Override{
				Spec: &v1limiter.OverrideSpec{
					DBDefinition: &v1limiter.OverrideDBDefinition{
						GroupByKeyValues: datatypes.KeyValues{
							"key1": datatypes.Int(1),
						},
					},
					Properties: &v1limiter.OverrideProperties{
						Limit: helpers.PointerOf[int64](3),
					},
				},
			}

			if i%2 == 0 {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i)}
			} else {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}
			}

			g.Expect(overrideReq.ValidateSpecOnly()).ToNot(HaveOccurred())
			_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideReq)
			g.Expect(err).ToNot(HaveOccurred())
		}

		err := overrideClient.DestroyOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule")
		g.Expect(err).ToNot(HaveOccurred())

		// check all overrides are deleted
		query := &queryassociatedaction.AssociatedActionQuery{} // match all
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		overrides, err := overrideClient.QueryOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule", query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(0))
	})

	t.Run("It returns an internal server error for everything else", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().DestroyOne(gomock.Any(), gomock.Any()).Return(btreeonetomany.ErrorManyIDDestroying).Times(1)

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
		err := overrideClient.DestroyOverrides(testhelpers.NewContextWithMiddlewareSetup(), "test rule")
		g.Expect(err).To(Equal(errors.InternalServerError))
	})
}

func Test_OverrideClientLocal_FindOverrideLimits(t *testing.T) {
	g := NewGomegaWithT(t)

	constructor, err := NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	setupTree := func(g *GomegaWithT) *overridesClientLocal {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		for i := 0; i < 50; i++ {
			overrideReq := &v1limiter.Override{
				Spec: &v1limiter.OverrideSpec{
					DBDefinition: &v1limiter.OverrideDBDefinition{
						GroupByKeyValues: datatypes.KeyValues{
							"key1": datatypes.Int(1),
						},
					},
					Properties: &v1limiter.OverrideProperties{
						Limit: helpers.PointerOf[int64](int64(i)),
					},
				},
			}

			if i%2 == 0 {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i)}
			} else {
				overrideReq.Spec.DBDefinition.GroupByKeyValues = datatypes.KeyValues{fmt.Sprintf("key%d", i): datatypes.Int(i), fmt.Sprintf("key%d", i+1): datatypes.Int(i + 1)}
			}

			g.Expect(overrideReq.ValidateSpecOnly()).ToNot(HaveOccurred())
			_, err := overrideClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test rule", overrideReq)
			g.Expect(err).ToNot(HaveOccurred())
		}

		return overrideClient
	}

	t.Run("It returns empty overrides if there are no found items", func(t *testing.T) {
		overrideClient := NewDefaultOverridesClientLocal(constructor)

		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		overrides, err := overrideClient.FindOverrideLimits(testhelpers.NewContextWithMiddlewareSetup(), "test rule", keyValues)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(overrides).ToNot(BeNil())
		g.Expect(len(overrides)).To(Equal(0))
	})

	t.Run("It reutnrs all overrides that match the permutations", func(t *testing.T) {
		overrideClient := setupTree(g)

		keyValues := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
			"key3": datatypes.Int(3),
			"key4": datatypes.Int(4),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		overrides, err := overrideClient.FindOverrideLimits(testhelpers.NewContextWithMiddlewareSetup(), "test rule", keyValues)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(4)) // pairs of groups{ {2}, {4}, {1, 2}, {3, 4} }
	})

	t.Run("It stops processing overrides if the limit is 0 for a found override", func(t *testing.T) {
		overrideClient := setupTree(g)

		keyValues := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		overrides, err := overrideClient.FindOverrideLimits(testhelpers.NewContextWithMiddlewareSetup(), "test rule", keyValues)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(1)) // pairs of groups{ {0} }
		g.Expect(*overrides[0].Spec.Properties.Limit).To(Equal(int64(0)))
	})

	t.Run("It returns an internal server error if the tree returns an error", func(t *testing.T) {
		mockController := gomock.NewController(t)
		mockBTreeOneToMany := btreeonetomanyfakes.NewMockBTreeOneToMany(mockController)
		defer mockController.Finish()

		mockBTreeOneToMany.EXPECT().MatchAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("nope")).Times(1)

		keyValues := &datatypes.KeyValues{
			"key0":  datatypes.Int(0),
			"key17": datatypes.Int(17),
			"key18": datatypes.Int(18),
		}
		g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())

		overrideClient := NewOverridesClientLocal(mockBTreeOneToMany, constructor)
		overrides, err := overrideClient.FindOverrideLimits(testhelpers.NewContextWithMiddlewareSetup(), "test rule", *keyValues)
		g.Expect(err).To(Equal(errors.InternalServerError))
		g.Expect(len(overrides)).To(Equal(0))
	})
}
