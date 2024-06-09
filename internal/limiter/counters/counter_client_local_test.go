package counters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/clients/locker_client/lockerclientfakes"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers"
	"github.com/DanLavine/willow/testhelpers/testmodels"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	. "github.com/onsi/gomega"
)

func setupLocalClient(g *GomegaWithT) (*counterClientLocal, rules.RuleClient) {
	overridesConstructor, err := overrides.NewOverrideConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	rulesConstructor, err := rules.NewRuleConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	countersConstructor, err := NewCountersConstructor("memory")
	g.Expect(err).ToNot(HaveOccurred())

	overridesClient := overrides.NewDefaultOverridesClientLocal(overridesConstructor)
	rulesClient := rules.NewLocalRulesClient(rulesConstructor, overridesClient)

	return NewCountersClientLocal(countersConstructor, rulesClient), rulesClient
}

func TestRulesManager_IncrementCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It increments the counters without obtaining any locks or checking other counters if there are no Rules", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.String("1"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}

		err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a limit reached error if any matched rule has a limit of 0", func(t *testing.T) {
		countersClientLocal, rulesClient := setupLocalClient(g)

		for i := 0; i < 5; i++ {
			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf(fmt.Sprintf("test%d", i)),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							fmt.Sprintf("key%d", i): datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](int64(i)),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}

		err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(ContainSubstring("Limit has already been reached for rule 'test0'"))
	})

	t.Run("It returns a limit reached error if any matched overrides have a limit of 0", func(t *testing.T) {
		countersClientLocal, rulesClient := setupLocalClient(g)

		// single instance rule group by
		createRequest := &v1limiter.Rule{
			Spec: &v1limiter.RuleSpec{
				DBDefinition: &v1limiter.RuleDBDefinition{
					Name: helpers.PointerOf("test1"),
					GroupByKeyValues: dbdefinition.AnyKeyValues{
						"key1": datatypes.Any(),
					},
				},
				Properties: &v1limiter.RuleProperties{
					Limit: helpers.PointerOf[int64](15),
				},
			},
		}
		g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
		g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

		// create 5 overrides
		for k := 2; k < 7; k++ {
			// create number of overrides
			overrideRequest := v1limiter.Override{
				Spec: &v1limiter.OverrideSpec{
					DBDefinition: &v1limiter.OverrideDBDefinition{
						Name: helpers.PointerOf(fmt.Sprintf("override%d", k)),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key1":                  datatypes.Int(1),
							fmt.Sprintf("key%d", k): datatypes.Int(k),
						},
					},
					Properties: &v1limiter.OverrideProperties{
						Limit: helpers.PointerOf[int64](int64(k - 2)),
					},
				},
			}
			g.Expect(overrideRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test1", &overrideRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.Int(1),
						"key2": datatypes.Int(2),
						"key3": datatypes.Int(3),
						"key4": datatypes.Int(4)},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}

		err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("Limit has already been reached for rule 'test1'"))
	})

	t.Run("Describe obtaining lock failures", func(t *testing.T) {
		t.Run("It locks and releases each key value pair when trying to increment a counter if a rule is found", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			//setup rules
			for i := 0; i < 5; i++ {
				// single instance rule group by
				createRequest := &v1limiter.Rule{
					Spec: &v1limiter.RuleSpec{
						DBDefinition: &v1limiter.RuleDBDefinition{
							Name: helpers.PointerOf(fmt.Sprintf("test%d", i)),
							GroupByKeyValues: dbdefinition.AnyKeyValues{
								fmt.Sprintf("key%d", i): datatypes.Any(),
							},
						},
						Properties: &v1limiter.RuleProperties{
							Limit: helpers.PointerOf[int64](int64(i)),
						},
					},
				}
				g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
				g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key1": datatypes.String("1"),
							"key2": datatypes.String("2"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			mockController := gomock.NewController(t)
			defer mockController.Finish()

			fakeLock := lockerclientfakes.NewMockLock(mockController)
			fakeLock.EXPECT().Done().Times(2)
			fakeLock.EXPECT().Release(gomock.Any()).Times(2)

			fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
			fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.Lock, errlog func(keyValue datatypes.KeyValues, err error)) (lockerclient.Lock, error) {
				return fakeLock, nil
			}).Times(2)

			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("Context when obtaining the lock fails", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				countersClientLocal, rulesClient := setupLocalClient(g)

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.Rule{
						Spec: &v1limiter.RuleSpec{
							DBDefinition: &v1limiter.RuleDBDefinition{
								Name: helpers.PointerOf(fmt.Sprintf("test%d", i)),
								GroupByKeyValues: dbdefinition.AnyKeyValues{
									fmt.Sprintf("key%d", i): datatypes.Any(),
								},
							},
							Properties: &v1limiter.RuleProperties{
								Limit: helpers.PointerOf[int64](int64(i)),
							},
						},
					}
					g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
					g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					Spec: &v1limiter.CounterSpec{
						DBDefinition: &v1limiter.CounterDBDefinition{
							KeyValues: dbdefinition.TypedKeyValues{
								"key1": datatypes.String("1"),
								"key2": datatypes.String("2"),
							},
						},
						Properties: &v1limiter.CounteProperties{
							Counters: helpers.PointerOf[int64](1),
						},
					},
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				fakeLock := lockerclientfakes.NewMockLock(mockController)
				fakeLock.EXPECT().Release(gomock.Any()).Times(1)
				fakeLock.EXPECT().Done().MaxTimes(1)

				obtainCount := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.Lock, errlog func(keyValue datatypes.KeyValues, err error)) (lockerclient.Lock, error) {
					if obtainCount == 0 {
						obtainCount++
						return fakeLock, nil
					} else {
						return nil, fmt.Errorf("failed to obtain 2nd lock")
					}
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testContext := context.WithValue(context.Background(), middleware.LoggerCtxKey, zap.New(testZapCore))

				err := countersClientLocal.IncrementCounters(testContext, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("failed to obtain a lock from the locker service"))
			})
		})

		t.Run("Context when a lock is lost that was already obtained", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				countersClientLocal, rulesClient := setupLocalClient(g)

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.Rule{
						Spec: &v1limiter.RuleSpec{
							DBDefinition: &v1limiter.RuleDBDefinition{
								Name: helpers.PointerOf(fmt.Sprintf("test%d", i)),
								GroupByKeyValues: dbdefinition.AnyKeyValues{
									fmt.Sprintf("key%d", i): datatypes.Any(),
								},
							},
							Properties: &v1limiter.RuleProperties{
								Limit: helpers.PointerOf[int64](int64(i)),
							},
						},
					}
					g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
					g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					Spec: &v1limiter.CounterSpec{
						DBDefinition: &v1limiter.CounterDBDefinition{
							KeyValues: dbdefinition.TypedKeyValues{
								"key1": datatypes.String("1"),
								"key2": datatypes.String("2"),
							},
						},
						Properties: &v1limiter.CounteProperties{
							Counters: helpers.PointerOf[int64](1),
						},
					},
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				donechan := make(chan struct{})
				close(donechan)

				fakeLock := lockerclientfakes.NewMockLock(mockController)
				fakeLock.EXPECT().Release(gomock.Any()).Times(2)
				fakeLock.EXPECT().Done().DoAndReturn(func() <-chan struct{} {
					return donechan
				}).MaxTimes(2)

				count := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.Lock, errlog func(keyValue datatypes.KeyValues, err error)) (lockerclient.Lock, error) {
					if count == 0 {
						count++
					} else {
						time.Sleep(100 * time.Millisecond)
					}

					return fakeLock, nil
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testContext := context.WithValue(context.Background(), middleware.LoggerCtxKey, zap.New(testZapCore))

				err := countersClientLocal.IncrementCounters(testContext, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("a lock was released unexpedily"))
			})
		})
	})

	t.Run("Context rule limits", func(t *testing.T) {
		mockController := gomock.NewController(t)

		fakeLock := lockerclientfakes.NewMockLock(mockController)
		fakeLock.EXPECT().Release(gomock.Any()).AnyTimes()
		fakeLock.EXPECT().Done().AnyTimes()

		fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
		fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.Lock, errlog func(keyValue datatypes.KeyValues, err error)) (lockerclient.Lock, error) {
			return fakeLock, nil
		}).AnyTimes()

		t.Run("It adds the counter if no rules have reached their limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](5),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("1"),
							"key1": datatypes.String("2"),
							"key3": datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// counter shuold be added
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			counterCount := 0
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				counterCount++
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(counterCount).To(Equal(1))
		})

		t.Run("It respects the unlimited rules", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](-1),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("1"),
							"key1": datatypes.String("2"),
							"key3": datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// counter shuold be added
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			counters := 0
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				counters++
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(counters).To(Equal(1))
		})

		t.Run("It can update the limit for counters that are below all the rules", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](5),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("1"),
							"key1": datatypes.String("2"),
							"key3": datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// counter shuold be added
			g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)).ToNot(HaveOccurred())

			// ensure the counter was added
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				count = item.Value().(Counter).Load()
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(4)))
		})

		t.Run("It returns an error if the counter >= the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](1),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				count = item.Value().(Counter).Load()
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if the counter >= the limit with any combination of different counters", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// single instance rule group by
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](1),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			counter1 := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
							"key2": datatypes.String("2"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}
			counter2 := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
							"key3": datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter1)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter2)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				count = item.Value().(Counter).Load()
				return true
			}

			query1 := queryassociatedaction.KeyValuesToExactAssociatedActionQuery(counter1.Spec.DBDefinition.KeyValues.ToKeyValues())
			counterErr := countersClientLocal.counters.QueryAction(query1, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))

			count = 0 // reset the counter
			query2 := queryassociatedaction.KeyValuesToExactAssociatedActionQuery(counter2.Spec.DBDefinition.KeyValues.ToKeyValues())
			counterErr = countersClientLocal.counters.QueryAction(query2, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(0)))
		})

		t.Run("It returns an error if any rule has hit the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// restrictive rule
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](1),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			// non restrictive rules
			for i := 1; i < 5; i++ {
				createRequest := &v1limiter.Rule{
					Spec: &v1limiter.RuleSpec{
						DBDefinition: &v1limiter.RuleDBDefinition{
							Name: helpers.PointerOf(fmt.Sprintf("test%d", i)),
							GroupByKeyValues: dbdefinition.AnyKeyValues{
								fmt.Sprintf("key%d", i):    datatypes.Any(),
								fmt.Sprintf("keyd%d", i+1): datatypes.Any(),
							},
						},
						Properties: &v1limiter.RuleProperties{
							Limit: helpers.PointerOf[int64](int64(i + 10)),
						},
					},
				}
				g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
				g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
							"key2": datatypes.String("2"),
							"key3": datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// first counter should be added
			fmt.Println("incrementing the counters now")
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule 'test0'"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				count = item.Value().(Counter).Load()
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if any rule's overrides hit the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			// restrictive rule
			createRequest := &v1limiter.Rule{
				Spec: &v1limiter.RuleSpec{
					DBDefinition: &v1limiter.RuleDBDefinition{
						Name: helpers.PointerOf("test0"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.Any(),
							"key1": datatypes.Any(),
						},
					},
					Properties: &v1limiter.RuleProperties{
						Limit: helpers.PointerOf[int64](1),
					},
				},
			}
			g.Expect(createRequest.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(testhelpers.NewContextWithMiddlewareSetup(), createRequest)).ToNot(HaveOccurred())

			// set override to allow for more values
			override := &v1limiter.Override{
				Spec: &v1limiter.OverrideSpec{
					DBDefinition: &v1limiter.OverrideDBDefinition{
						Name: helpers.PointerOf("override1"),
						GroupByKeyValues: dbdefinition.AnyKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
						},
					},
					Properties: &v1limiter.OverrideProperties{
						Limit: helpers.PointerOf[int64](5),
					},
				},
			}
			g.Expect(override.ValidateSpecOnly()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateOverride(testhelpers.NewContextWithMiddlewareSetup(), "test0", override)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"key0": datatypes.String("0"),
							"key1": datatypes.String("1"),
							"key2": datatypes.String("2"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should be added as well
			err = countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) bool {
				count = item.Value().(Counter).Load()
				return true
			}

			counterErr := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(2)))
		})
	})
}

func TestRulesManager_DecrementCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op if there are no counters to decrement", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.String("1"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}

		err := countersClientLocal.DecrementCounters(testhelpers.NewContextWithMiddlewareSetup(), counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It decrements a counter when the count is above 1", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item btreeassociated.AssociatedKeyValues) bool {
			count = item.Value().(Counter).Load()
			return true
		}

		err := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(4)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](-1),
				},
			},
		}
		err = countersClientLocal.DecrementCounters(testhelpers.NewContextWithMiddlewareSetup(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		err = countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(3)))
	})

	t.Run("It removes a counter when the count is at 0", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item btreeassociated.AssociatedKeyValues) bool {
			count = item.Value().(Counter).Load()
			return true
		}

		err := countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(1)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](-1),
				},
			},
		}
		err = countersClientLocal.DecrementCounters(testhelpers.NewContextWithMiddlewareSetup(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		err = countersClientLocal.counters.QueryAction(&queryassociatedaction.AssociatedActionQuery{}, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(0)))
	})
}

func TestRulesManager_QueryCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns empty list if there are no counters that match the query", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"not found": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.NoTypeRestrictions(g),
					},
				},
			},
		}

		countersResponse, err := countersClientLocal.QueryCounters(testhelpers.NewContextWithMiddlewareSetup(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a list of counters that match the query", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// create a number of various counters
		counter1 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter1)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter1)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter1)).ToNot(HaveOccurred())

		counter2 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.Int(0),
						"key1": datatypes.Int(1),
						"key2": datatypes.Int(2),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter2)).ToNot(HaveOccurred())

		counter3 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.Int(0),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter3)).ToNot(HaveOccurred())

		counter4 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.String("0"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter4)).ToNot(HaveOccurred())

		counter5 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.Int(0),
						"key1": datatypes.String("0"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter5)).ToNot(HaveOccurred())

		// run the query
		query := &queryassociatedaction.AssociatedActionQuery{
			Selection: &queryassociatedaction.Selection{
				KeyValues: queryassociatedaction.SelectionKeyValues{
					"key0": {
						Value:            datatypes.Any(),
						Comparison:       v1common.Equals,
						TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_string, datatypes.T_string),
					},
				},
			},
			Or: []*queryassociatedaction.AssociatedActionQuery{
				&queryassociatedaction.AssociatedActionQuery{
					Selection: &queryassociatedaction.Selection{
						KeyValues: queryassociatedaction.SelectionKeyValues{
							"key0": {
								Value:            datatypes.Int(0),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
							"key1": {
								Value:            datatypes.Int(1),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
							"key2": {
								Value:            datatypes.Int(2),
								Comparison:       v1common.Equals,
								TypeRestrictions: testmodels.TypeRestrictions(g, datatypes.T_int, datatypes.T_int),
							},
						},
					},
				},
			},
		}

		resp1 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](3),
				},
			},
			State: &v1limiter.CounterState{
				Deleting: false,
			},
		}
		resp2 := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.Int(0),
						"key1": datatypes.Int(1),
						"key2": datatypes.Int(2),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
			State: &v1limiter.CounterState{
				Deleting: false,
			},
		}

		countersResponse, err := countersClientLocal.QueryCounters(testhelpers.NewContextWithMiddlewareSetup(), query)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(countersResponse)).To(Equal(2))
		g.Expect(countersResponse).To(ContainElement(resp1))
		g.Expect(countersResponse).To(ContainElement(resp2))
	})
}

func TestRulesManager_SetCounter(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It sets the value for the particualr key values regardless of the value already stored", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// set the counters
		kvs := dbdefinition.TypedKeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Float64(3.4),
		}
		countersSet := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key1": datatypes.Int(1),
						"key2": datatypes.Float64(3.4),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](87),
				},
			},
		}

		err := countersClientLocal.SetCounter(testhelpers.NewContextWithMiddlewareSetup(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &queryassociatedaction.AssociatedActionQuery{} // select all

		countersResponse, err := countersClientLocal.QueryCounters(testhelpers.NewContextWithMiddlewareSetup(), query)
		g.Expect(len(countersResponse)).To(Equal(1))
		g.Expect(countersResponse[0].Spec.DBDefinition.KeyValues).To(Equal(kvs))
		g.Expect(*countersResponse[0].Spec.Properties.Counters).To(Equal(int64(87)))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It removes the counters if set to 0", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// create initial counters through increment
		counter := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](1),
				},
			},
		}
		g.Expect(countersClientLocal.IncrementCounters(testhelpers.NewContextWithMiddlewareSetup(), nil, counter)).ToNot(HaveOccurred())

		// set the counters to 0
		countersSet := &v1limiter.Counter{
			Spec: &v1limiter.CounterSpec{
				DBDefinition: &v1limiter.CounterDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"key0": datatypes.String("0"),
						"key1": datatypes.String("1"),
					},
				},
				Properties: &v1limiter.CounteProperties{
					Counters: helpers.PointerOf[int64](0),
				},
			},
		}

		err := countersClientLocal.SetCounter(testhelpers.NewContextWithMiddlewareSetup(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &queryassociatedaction.AssociatedActionQuery{} // select all

		countersResponse, err := countersClientLocal.QueryCounters(testhelpers.NewContextWithMiddlewareSetup(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})
}
