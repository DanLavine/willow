package lockmanager

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanLavine/goasync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func defaultLockCreateRequest() *v1locker.LockCreateRequest {
	return &v1locker.LockCreateRequest{
		KeyValues: datatypes.KeyValues{
			"1": datatypes.String("one"),
			"2": datatypes.Int(2),
			"3": datatypes.Uint64(3),
		},
		Timeout: 15 * time.Second,
	}
}

// the keys "1" overlap
func overlapOneKeyValue() *v1locker.LockCreateRequest {
	return &v1locker.LockCreateRequest{
		KeyValues: datatypes.KeyValues{
			"1": datatypes.String("one"),
			"4": datatypes.Int(2),
			"5": datatypes.Uint64(3),
		},
		Timeout: 15 * time.Second,
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

		lockResp := generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())
		g.Expect(lockResp.SessionID).ToNot(Equal(""))
		g.Expect(lockResp.Timeout).To(Equal(15 * time.Second)) // default values

		counter := 0
		lockCounter := func(_ *btreeassociated.AssociatedKeyValues) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(datatypes.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
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

				lockResp := generalLocker.ObtainLock(ctx, defaultLockCreateRequest())
				g.Expect(lockResp).ToNot(BeNil())

				counter := 0
				lockCounter := func(_ *btreeassociated.AssociatedKeyValues) bool {
					counter++
					return true
				}

				g.Expect(generalLocker.locks.Query(datatypes.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
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

				lockResp := generalLocker.ObtainLock(clientCtx, defaultLockCreateRequest())
				g.Expect(lockResp).ToNot(BeNil())

				counter := 0
				lockCounter := func(_ *btreeassociated.AssociatedKeyValues) bool {
					counter++
					return true
				}

				g.Expect(generalLocker.locks.Query(datatypes.AssociatedKeyValuesQuery{}, lockCounter)).ToNot(HaveOccurred())
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

			lockResp := generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
			g.Expect(lockResp).ToNot(BeNil())

			locked := make(chan struct{})
			for i := 0; i < 10; i++ {
				go func() {
					_ = generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
					select {
					case locked <- struct{}{}:
					case <-ctx.Done():
					}
				}()
			}

			g.Consistently(locked).ShouldNot(Receive())

			// perform an unlock
			generalLocker.ReleaseLock(lockResp.SessionID)

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
					lockResp := generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
					generalLocker.ReleaseLock(lockResp.SessionID)
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

				testReq := &v1locker.LockCreateRequest{
					KeyValues: datatypes.KeyValues{
						fmt.Sprintf("%d", i%5): datatypes.String("doesn't matter 1"),
						fmt.Sprintf("%d", i%6): datatypes.String("doesn't matter 2"),
						fmt.Sprintf("%d", i%7): datatypes.String("doesn't matter 3"),
					},
					Timeout: 15 * time.Second,
				}
				g.Expect(testReq.Validate()).ToNot(HaveOccurred())

				go func(req *v1locker.LockCreateRequest) {
					defer wg.Done()
					lockResp := generalLocker.ObtainLock(context.Background(), req)
					generalLocker.ReleaseLock(lockResp.SessionID)
				}(testReq)
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
						_ = generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
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
						_ = generalLocker.ObtainLock(ctx, defaultLockCreateRequest())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = generalLocker.ObtainLock(ctx, defaultLockCreateRequest())
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

	t.Run("It properly queries all locks that were created", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		go func() {
			taskManager.Run(ctx)
		}()

		generalLocker := NewGeneralLocker(nil)
		taskManager.AddExecuteTask("teset", generalLocker)

		lock1 := generalLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lock1).ToNot(BeNil())
		lock2 := generalLocker.ObtainLock(context.Background(), overlapOneKeyValue())
		g.Expect(lock2).ToNot(BeNil())

		query := datatypes.AssociatedKeyValuesQuery{}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		locks := generalLocker.LocksQuery(&v1.AssociatedQuery{AssociatedKeyValues: query})
		g.Expect(len(locks)).To(Equal(2))

		if locks[0].SessionID == lock2.SessionID {
			g.Expect(locks[0].SessionID).To(Equal(lock2.SessionID))
			g.Expect(locks[0].LocksHeldOrWaiting).To(Equal(1))
			g.Expect(locks[0].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[0].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[0].TimeTillExipre).ToNot(Equal(0))

			g.Expect(locks[1].SessionID).To(Equal(lock1.SessionID))
			g.Expect(locks[1].LocksHeldOrWaiting).To(Equal(1))
			g.Expect(locks[1].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[1].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[1].TimeTillExipre).ToNot(Equal(0))
		} else {
			g.Expect(locks[1].SessionID).To(Equal(lock2.SessionID))
			g.Expect(locks[1].LocksHeldOrWaiting).To(Equal(1))
			g.Expect(locks[1].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[1].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[1].TimeTillExipre).ToNot(Equal(0))

			g.Expect(locks[0].SessionID).To(Equal(lock1.SessionID))
			g.Expect(locks[0].LocksHeldOrWaiting).To(Equal(1))
			g.Expect(locks[0].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[0].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[0].TimeTillExipre).ToNot(Equal(0))
		}
	})
}

func TestGeneralLocker_Heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	query := datatypes.AssociatedKeyValuesQuery{}
	g.Expect(query.Validate()).ToNot(HaveOccurred())
	generalQuery := &v1.AssociatedQuery{AssociatedKeyValues: query}

	t.Run("It returns an error if the session ID does not exist", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		go func() {
			taskManager.Run(ctx)
		}()

		generalLocker := NewGeneralLocker(nil)
		taskManager.AddExecuteTask("teset", generalLocker)

		err := generalLocker.Heartbeat("bad id")
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("SessionID could not be found"))
	})

	t.Run("It keeps the lock around as long as the heartbeats are received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		go func() {
			taskManager.Run(ctx)
		}()

		generalLocker := NewGeneralLocker(nil)
		taskManager.AddExecuteTask("teset", generalLocker)

		lockRequest := defaultLockCreateRequest()
		lockRequest.Timeout = 100 * time.Millisecond

		lockResp := generalLocker.ObtainLock(context.Background(), lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		// this timer is longer than the heartbeat timeout
		for i := 0; i < 3; i++ {
			time.Sleep(60 * time.Millisecond)
			g.Expect(generalLocker.Heartbeat(lockResp.SessionID)).To(BeNil())
		}

		locks := generalLocker.LocksQuery(generalQuery)
		g.Expect(len(locks)).To(Equal(1))
	})

	t.Run("It removes the lock if no heartbeats are received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// start the task manager with the locker
		generalLocker := NewGeneralLocker(nil)
		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		taskManager.AddExecuteTask("teset", generalLocker)
		go func() {
			taskManager.Run(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.Timeout = 100 * time.Millisecond

		lockResp := generalLocker.ObtainLock(context.Background(), lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		g.Eventually(func() int {
			return len(generalLocker.LocksQuery(generalQuery))
		}).Should(Equal(0))
	})

	t.Run("Context when multiple clients are waiting for the same lock", func(t *testing.T) {
		t.Run("It allows a new client to obtain the lock", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			go func() {
				taskManager.Run(ctx)
			}()

			generalLocker := NewGeneralLocker(nil)
			taskManager.AddExecuteTask("teset", generalLocker)

			lockRequest := defaultLockCreateRequest()
			lockRequest.Timeout = 100 * time.Millisecond

			lockResp := generalLocker.ObtainLock(context.Background(), lockRequest)
			g.Expect(lockResp).ToNot(BeNil())

			sessionIDChan := make(chan *v1locker.LockCreateResponse)
			go func() {
				sessionIDChan <- generalLocker.ObtainLock(context.Background(), lockRequest)
			}()

			// ensure that the channel eventually recieves for the next request
			g.Eventually(sessionIDChan).Should(Receive(Not(BeNil())))

			// the locks should still be listed
			locks := generalLocker.LocksQuery(generalQuery)
			g.Expect(len(locks)).To(Equal(1))
		})
	})
}
