package locker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanLavine/goasync"

	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	. "github.com/onsi/gomega"
)

func defaultKeyValues() datatypes.StringMap {
	return datatypes.StringMap{
		"1": datatypes.String("one"),
		"2": datatypes.Int(2),
		"3": datatypes.Uint64(3),
	}
}

// the keys "1" overlap
func overlapOneKeyValue() datatypes.StringMap {
	return datatypes.StringMap{
		"1": datatypes.String("one"),
		"4": datatypes.Int(2),
		"5": datatypes.Uint64(3),
	}
}

func TestGeneralLocker_NewGeneralLocker(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It creates a tree if nil is passed in for the locks", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)
		g.Expect(generalLocker).ToNot(BeNil())
		g.Expect(generalLocker.locks).ToNot(BeNil())
	})
}

func TestGeneralLocker_ObtainLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It creates a single lock for the provided key values", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		go func() {
			taskManager.Run(ctx)
		}()

		generalLocker := NewGeneralLocker(nil)
		taskManager.AddExecuteTask("teset", generalLocker)

		lockID := generalLocker.ObtainLock(context.Background(), defaultKeyValues())
		g.Expect(lockID).ToNot(Equal(""))

		counter := 0
		lockCounter := func(_ any) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(query.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(1))
	})

	t.Run("Context when creating a lock", func(t *testing.T) {
		t.Run("Context when the server context is closed", func(t *testing.T) {
			t.Run("It always reports the sessionID", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())

				done := make(chan struct{})
				taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
				go func() {
					taskManager.Run(ctx)
					close(done)
				}()

				generalLocker := NewGeneralLocker(nil)
				taskManager.AddExecuteTask("teset", generalLocker)

				cancel()
				g.Eventually(done).Should(BeClosed())

				lockID := generalLocker.ObtainLock(ctx, defaultKeyValues())
				g.Expect(lockID).ToNot(Equal(""))

				counter := 0
				lockCounter := func(_ any) bool {
					counter++
					return true
				}

				g.Expect(generalLocker.locks.Query(query.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
				g.Expect(counter).To(Equal(1))
			})
		})

		t.Run("Context when client context is closed", func(t *testing.T) {
			t.Run("It always reports the sessionID", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				done := make(chan struct{})
				taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
				go func() {
					taskManager.Run(ctx)
					close(done)
				}()

				generalLocker := NewGeneralLocker(nil)
				taskManager.AddExecuteTask("teset", generalLocker)

				clientCtx, clientCancel := context.WithCancel(context.Background())
				clientCancel()

				lockID := generalLocker.ObtainLock(clientCtx, defaultKeyValues())
				g.Expect(lockID).ToNot(Equal(""))

				counter := 0
				lockCounter := func(_ any) bool {
					counter++
					return true
				}

				g.Expect(generalLocker.locks.Query(query.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
				g.Expect(counter).To(Equal(1))
			})
		})
	})

	t.Run("Context when waiting for a lock", func(t *testing.T) {
		t.Run("It only unlocks one waiting client", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			go func() {
				taskManager.Run(ctx)
			}()

			generalLocker := NewGeneralLocker(nil)
			taskManager.AddExecuteTask("teset", generalLocker)

			lockID := generalLocker.ObtainLock(context.Background(), defaultKeyValues())
			g.Expect(lockID).ToNot(Equal(""))

			locked := make(chan struct{})
			for i := 0; i < 10; i++ {
				go func() {
					_ = generalLocker.ObtainLock(context.Background(), defaultKeyValues())
					select {
					case locked <- struct{}{}:
					case <-ctx.Done():
					}
				}()
			}

			g.Consistently(locked).ShouldNot(Receive())

			// perform an unlock
			generalLocker.ReleaseLock(lockID)

			g.Eventually(locked).Should(Receive())
			g.Consistently(locked).ShouldNot(Receive())
		})

		t.Run("It can make many concurent requests to the same lock", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			go func() {
				taskManager.Run(ctx)
			}()

			generalLocker := NewGeneralLocker(nil)
			taskManager.AddExecuteTask("teset", generalLocker)

			wg := new(sync.WaitGroup)
			for i := 10_000; i < 0; i++ {
				wg.Add(1)

				go func() {
					defer wg.Done()
					lockID := generalLocker.ObtainLock(context.Background(), defaultKeyValues())
					generalLocker.ReleaseLock(lockID)
				}()
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-time.After(10 * time.Second):
				g.Fail("Failed to run async waits for same lock")
			case <-done:
				// nothing to do here, everything passed
			}
		})

		t.Run("It can make many concurent requests to multiple lock", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			go func() {
				taskManager.Run(ctx)
			}()

			generalLocker := NewGeneralLocker(nil)
			taskManager.AddExecuteTask("teset", generalLocker)

			wg := new(sync.WaitGroup)
			for i := 10_000; i < 0; i++ {
				wg.Add(1)

				testKeyValues := datatypes.StringMap{
					fmt.Sprintf("%d", i%5): datatypes.String("doesn't matter 1"),
					fmt.Sprintf("%d", i%6): datatypes.String("doesn't matter 2"),
					fmt.Sprintf("%d", i%7): datatypes.String("doesn't matter 3"),
				}

				go func(keyValues datatypes.StringMap) {
					defer wg.Done()
					lockID := generalLocker.ObtainLock(context.Background(), keyValues)
					generalLocker.ReleaseLock(lockID)
				}(testKeyValues)
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-time.After(10 * time.Second):
				g.Fail("Failed to run async waits for same lock")
			case <-done:
				// nothing to do here, everything passed
			}
		})

		t.Run("Context when the server is shutdown", func(t *testing.T) {
			t.Run("It frees up any waiting clients", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
				go func() {
					taskManager.Run(ctx)
				}()

				generalLocker := NewGeneralLocker(nil)
				taskManager.AddExecuteTask("teset", generalLocker)

				wg := new(sync.WaitGroup)
				for i := 10; i < 0; i++ {
					if i == 0 {
						_ = generalLocker.ObtainLock(context.Background(), defaultKeyValues())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = generalLocker.ObtainLock(context.Background(), defaultKeyValues())
						}()
					}
				}

				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()

				cancel()

				select {
				case <-time.After(10 * time.Second):
					g.Fail("Failed to run async waits for same lock")
				case <-done:
					// nothing to do here, everything closed properly
				}
			})
		})

		t.Run("Context if the client is closed", func(t *testing.T) {
			t.Run("It frees up any waiting clients", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
				go func() {
					taskManager.Run(context.Background())
				}()

				generalLocker := NewGeneralLocker(nil)
				taskManager.AddExecuteTask("teset", generalLocker)

				wg := new(sync.WaitGroup)
				for i := 10; i < 0; i++ {
					if i == 0 {
						_ = generalLocker.ObtainLock(ctx, defaultKeyValues())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = generalLocker.ObtainLock(ctx, defaultKeyValues())
						}()
					}
				}

				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()

				cancel()

				select {
				case <-time.After(10 * time.Second):
					g.Fail("Failed to run async waits for same lock")
				case <-done:
					// nothing to do here, everything closed properly
				}
			})
		})
	})
}

func TestGeneralLocker_ListLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It lists all locks that have been created", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		go func() {
			taskManager.Run(ctx)
		}()

		generalLocker := NewGeneralLocker(nil)
		taskManager.AddExecuteTask("teset", generalLocker)

		g.Expect(generalLocker.ObtainLock(context.Background(), defaultKeyValues())).ToNot(BeNil())
		g.Expect(generalLocker.ObtainLock(context.Background(), overlapOneKeyValue())).ToNot(BeNil())

		locks := generalLocker.ListLocks()
		g.Expect(len(locks)).To(Equal(2))

		g.Expect(locks).To(ContainElement(v1locker.Lock{LocksHeldOrWaiting: 1, KeyValues: datatypes.StringMap{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{LocksHeldOrWaiting: 1, KeyValues: datatypes.StringMap{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}}))
	})
}
