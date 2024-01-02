package counters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/pkg/clients/locker_client/lockerclientfakes"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("1"),
			},
			Counters: 1,
		}

		err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a limit reached error if any matched rule has a limit of 0", func(t *testing.T) {
		countersClientLocal, rulesClient := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i := 0; i < 5; i++ {
			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    fmt.Sprintf("test%d", i),
				GroupBy: []string{fmt.Sprintf("key%d", i)},
				Limit:   int64(i),
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}

		err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(ContainSubstring("Limit has already been reached for rule 'test0'"))
	})

	t.Run("It returns a limit reached error if any matched overrides have a limit of 0", func(t *testing.T) {
		countersClientLocal, rulesClient := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// single instance rule group by
		createRequest := &v1limiter.RuleCreateRequest{
			Name:    "test1",
			GroupBy: []string{"key1"},
			Limit:   15,
		}
		g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
		g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

		// create 5 overrides
		for k := 2; k < 7; k++ {
			// create number of overrides
			overrideRequest := v1limiter.Override{
				Name: fmt.Sprintf("override%d", k),
				KeyValues: datatypes.KeyValues{
					"key1":                  datatypes.Int(1),
					fmt.Sprintf("key%d", k): datatypes.Int(k),
				},
				Limit: int64(k - 2),
			}
			g.Expect(overrideRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateOverride(zap.NewNop(), "test1", &overrideRequest)).ToNot(HaveOccurred())
		}

		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.Int(1),
				"key2": datatypes.Int(2),
				"key3": datatypes.Int(3),
				"key4": datatypes.Int(4),
			},
			Counters: 1,
		}

		err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Message).To(Equal("Limit has already been reached for rule 'test1'"))
	})

	t.Run("Describe obtaining lock failures", func(t *testing.T) {
		t.Run("It locks and releases each key value pair when trying to increment a counter if a rule is found", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			//setup rules
			for i := 0; i < 5; i++ {
				// single instance rule group by
				createRequest := &v1limiter.RuleCreateRequest{
					Name:    fmt.Sprintf("test%d", i),
					GroupBy: []string{fmt.Sprintf("key%d", i)},
					Limit:   int64(i),
				}
				g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}

			mockController := gomock.NewController(t)
			defer mockController.Finish()

			fakeLock := lockerclientfakes.NewMockLocker(mockController)
			fakeLock.EXPECT().Done().Times(2)
			fakeLock.EXPECT().Release().Times(2)

			fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
			fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Locker, error) {
				return fakeLock, nil
			}).Times(2)

			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("Context when obtaining the lock fails", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				countersClientLocal, rulesClient := setupLocalClient(g)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.RuleCreateRequest{
						Name:    fmt.Sprintf("test%d", i),
						GroupBy: []string{fmt.Sprintf("key%d", i)},
						Limit:   int64(i),
					}
					g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
					g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					KeyValues: datatypes.KeyValues{
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
					Counters: 1,
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				fakeLock := lockerclientfakes.NewMockLocker(mockController)
				fakeLock.EXPECT().Release().Times(1)
				fakeLock.EXPECT().Done().MaxTimes(1)

				obtainCount := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Locker, error) {
					if obtainCount == 0 {
						obtainCount++
						return fakeLock, nil
					} else {
						return nil, fmt.Errorf("failed to obtain 2nd lock")
					}
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testLgger := zap.New(testZapCore)

				err := countersClientLocal.IncrementCounters(testLgger, ctx, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("failed to obtain a lock from the locker service"))
			})
		})

		t.Run("Context when a lock is lost that was already obtained", func(t *testing.T) {
			t.Run("It returns an error and releases any locks currently held", func(t *testing.T) {
				countersClientLocal, rulesClient := setupLocalClient(g)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				for i := 0; i < 5; i++ {
					// single instance rule group by
					createRequest := &v1limiter.RuleCreateRequest{
						Name:    fmt.Sprintf("test%d", i),
						GroupBy: []string{fmt.Sprintf("key%d", i)},
						Limit:   int64(i),
					}
					g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
					g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
				}

				counter := &v1limiter.Counter{
					KeyValues: datatypes.KeyValues{
						"key1": datatypes.String("1"),
						"key2": datatypes.String("2"),
					},
					Counters: 1,
				}

				mockController := gomock.NewController(t)
				defer mockController.Finish()

				donechan := make(chan struct{})
				close(donechan)

				fakeLock := lockerclientfakes.NewMockLocker(mockController)
				fakeLock.EXPECT().Release().Times(2)
				fakeLock.EXPECT().Done().DoAndReturn(func() <-chan struct{} {
					return donechan
				}).MaxTimes(2)

				count := 0
				fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
				fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Locker, error) {
					if count == 0 {
						count++
					} else {
						time.Sleep(100 * time.Millisecond)
					}

					return fakeLock, nil
				}).Times(2)

				// observe the proper error message in the logs
				testZapCore, testLogs := observer.New(zap.InfoLevel)
				testLgger := zap.New(testZapCore)

				err := countersClientLocal.IncrementCounters(testLgger, ctx, fakeLocker, counter)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(Equal(errors.InternalServerError))
				g.Expect(len(testLogs.All())).To(Equal(1))
				g.Expect(testLogs.All()[0].Message).To(ContainSubstring("a lock was released unexpedily"))
			})
		})
	})

	t.Run("Context rule limits", func(t *testing.T) {
		mockController := gomock.NewController(t)

		fakeLock := lockerclientfakes.NewMockLocker(mockController)
		fakeLock.EXPECT().Release().AnyTimes()
		fakeLock.EXPECT().Done().AnyTimes()

		fakeLocker := lockerclientfakes.NewMockLockerClient(mockController)
		fakeLocker.EXPECT().ObtainLock(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, lockRequest *v1locker.LockCreateRequest) (lockerclient.Locker, error) {
			return fakeLock, nil
		}).AnyTimes()

		t.Run("It adds the counter if no rules have reached their limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// counter shuold be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			found := false
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				found = true
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue())
		})

		t.Run("It respects the unlimited rules", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   -1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 341,
			}

			// counter shuold be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was added
			found := false
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				found = true
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue())
		})

		t.Run("It can update the limit for counters that are below all the rules", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   5,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("1"),
					"key1": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// counter shuold be added
			g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())
			g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)).ToNot(HaveOccurred())

			// ensure the counter was added
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				count = item.Value().(Counter).Load()
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(4)))
		})

		t.Run("It returns an error if the counter >= the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				count = item.Value().(Counter).Load()
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if the counter >= the limit with any combination of different counters", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// single instance rule group by
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			counter1 := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}
			counter2 := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter1)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter2)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				count = item.Value().(Counter).Load()
			}

			counterErr := countersClientLocal.counters.Find(counter1.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))

			count = 0 // reset the counter
			counterErr = countersClientLocal.counters.Find(counter2.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(0)))
		})

		t.Run("It returns an error if any rule has hit the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// restrictive rule
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// non restrictive rules
			for i := 1; i < 5; i++ {
				createRequest := &v1limiter.RuleCreateRequest{
					Name:    fmt.Sprintf("test%d", i),
					GroupBy: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("key%d", i+1)},
					Limit:   int64(i + 10),
				}
				g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
				g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())
			}

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
					"key3": datatypes.String("3"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should have an error
			err = countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Limit has already been reached for rule"))

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				count = item.Value().(Counter).Load()
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
			g.Expect(counterErr).ToNot(HaveOccurred())
			g.Expect(count).To(Equal(int64(1)))
		})

		t.Run("It returns an error if any rule's overrides hit the limit", func(t *testing.T) {
			countersClientLocal, rulesClient := setupLocalClient(g)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// restrictive rule
			createRequest := &v1limiter.RuleCreateRequest{
				Name:    "test0",
				GroupBy: []string{"key0", "key1"},
				Limit:   1,
			}
			g.Expect(createRequest.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateRule(zap.NewNop(), createRequest)).ToNot(HaveOccurred())

			// set override to allow for more values
			override := &v1limiter.Override{
				Name: "override1",
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
				},
				Limit: 5,
			}
			g.Expect(override.Validate()).ToNot(HaveOccurred())
			g.Expect(rulesClient.CreateOverride(zap.NewNop(), "test0", override)).ToNot(HaveOccurred())

			counter := &v1limiter.Counter{
				KeyValues: datatypes.KeyValues{
					"key0": datatypes.String("0"),
					"key1": datatypes.String("1"),
					"key2": datatypes.String("2"),
				},
				Counters: 1,
			}

			// first counter should be added
			err := countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// second counter should be added as well
			err = countersClientLocal.IncrementCounters(zap.NewNop(), ctx, fakeLocker, counter)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the counter was only ever incremented 1 time
			count := int64(0)
			onFind := func(item btreeassociated.AssociatedKeyValues) {
				count = item.Value().(Counter).Load()
			}

			counterErr := countersClientLocal.counters.Find(counter.KeyValues, onFind)
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
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("1"),
			},
			Counters: 1,
		}

		err := countersClientLocal.DecrementCounters(zap.NewNop(), counter)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It decrements a counter when the count is above 1", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item btreeassociated.AssociatedKeyValues) {
			count = item.Value().(Counter).Load()
		}

		err := countersClientLocal.counters.Find(counter.KeyValues, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(4)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: -1,
		}
		err = countersClientLocal.DecrementCounters(zap.NewNop(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		err = countersClientLocal.counters.Find(counter.KeyValues, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(3)))
	})

	t.Run("It removes a counter when the count is at 0", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// increment the counters 4 times
		counter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: 1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// ensure the counter value is correct
		count := int64(0)
		onFind := func(item btreeassociated.AssociatedKeyValues) {
			count = item.Value().(Counter).Load()
		}

		err := countersClientLocal.counters.Find(counter.KeyValues, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(1)))

		// run a decrement count
		decrementCounter := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.String("0"),
				"key1": datatypes.String("1"),
				"key2": datatypes.String("2"),
			},
			Counters: -1,
		}
		err = countersClientLocal.DecrementCounters(zap.NewNop(), decrementCounter)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counter was decremented correctly
		count = int64(0)
		err = countersClientLocal.counters.Find(counter.KeyValues, onFind)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(count).To(Equal(int64(0)))
	})
}

func TestRulesManager_QueryCounters(t *testing.T) {
	g := NewGomegaWithT(t)

	exists := true

	t.Run("It returns empty list if there are no counters that match the query", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"not found": datatypes.Value{
							Exists: &exists,
						},
					},
				},
			},
		}

		countersResponse, err := countersClientLocal.QueryCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It returns a list of counters that match the query", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create a number of various counters
		keyValuesOne := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
			"key2": datatypes.String("2"),
		}
		counter1 := &v1limiter.Counter{
			KeyValues: keyValuesOne,
			Counters:  1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter1)).ToNot(HaveOccurred())

		keyValuesTwo := datatypes.KeyValues{
			"key0": datatypes.Int(0),
			"key1": datatypes.Int(1),
			"key2": datatypes.Int(2),
		}
		counter2 := &v1limiter.Counter{
			KeyValues: keyValuesTwo,
			Counters:  1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter2)).ToNot(HaveOccurred())

		counter3 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
			},
			Counters: 1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter3)).ToNot(HaveOccurred())

		counter4 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("0"),
			},
			Counters: 1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter4)).ToNot(HaveOccurred())

		counter5 := &v1limiter.Counter{
			KeyValues: datatypes.KeyValues{
				"key0": datatypes.Int(0),
				"key1": datatypes.String("0"),
			},
			Counters: 1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter5)).ToNot(HaveOccurred())

		// run the query
		int0 := datatypes.Int(0)
		int1 := datatypes.Int(1)
		int2 := datatypes.Int(2)

		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"key0": datatypes.Value{ // values 1
							Exists:     &exists,
							ExistsType: &datatypes.T_string,
						},
					},
				},
				Or: []datatypes.AssociatedKeyValuesQuery{ // values 2
					datatypes.AssociatedKeyValuesQuery{
						KeyValueSelection: &datatypes.KeyValueSelection{
							KeyValues: map[string]datatypes.Value{
								"key0": datatypes.Value{Value: &int0, ValueComparison: datatypes.EqualsPtr()},
								"key1": datatypes.Value{Value: &int1, ValueComparison: datatypes.EqualsPtr()},
								"key2": datatypes.Value{Value: &int2, ValueComparison: datatypes.EqualsPtr()},
							},
						},
					},
				},
			},
		}

		resp1 := &v1limiter.Counter{
			KeyValues: keyValuesOne,
			Counters:  3,
		}
		resp2 := &v1limiter.Counter{
			KeyValues: keyValuesTwo,
			Counters:  1,
		}

		countersResponse, err := countersClientLocal.QueryCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(2))
		g.Expect(countersResponse).To(ConsistOf(resp1, resp2))
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestRulesManager_SetCounter(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It sets the value for the particualr key values regardless of the value already stored", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		// set the counters
		kvs := datatypes.KeyValues{
			"key1": datatypes.Int(1),
			"key2": datatypes.Float64(3.4),
		}
		countersSet := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  87,
		}

		err := countersClientLocal.SetCounter(zap.NewNop(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}, // select all
		}

		countersResponse, err := countersClientLocal.QueryCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(1))
		g.Expect(countersResponse[0].KeyValues).To(Equal(kvs))
		g.Expect(countersResponse[0].Counters).To(Equal(int64(87)))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It removes the counters if set to 0", func(t *testing.T) {
		countersClientLocal, _ := setupLocalClient(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create initial counters through increment
		kvs := datatypes.KeyValues{
			"key0": datatypes.String("0"),
			"key1": datatypes.String("1"),
		}
		counter := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  1,
		}
		g.Expect(countersClientLocal.IncrementCounters(zap.NewNop(), ctx, nil, counter)).ToNot(HaveOccurred())

		// set the counters to 0
		countersSet := &v1limiter.Counter{
			KeyValues: kvs,
			Counters:  0,
		}

		err := countersClientLocal.SetCounter(zap.NewNop(), countersSet)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure they are set properly
		query := &v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{}, // select all
		}

		countersResponse, err := countersClientLocal.QueryCounters(zap.NewNop(), query)
		g.Expect(len(countersResponse)).To(Equal(0))
		g.Expect(err).ToNot(HaveOccurred())
	})
}
