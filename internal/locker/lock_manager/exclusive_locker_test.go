package lockmanager

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/testhelpers"

	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	. "github.com/onsi/gomega"
)

func defaultLockCreateRequest() *v1locker.Lock {
	return &v1locker.Lock{
		Spec: &v1locker.LockSpec{
			DBDeifinition: &dbdefinition.TypedKeyValues{
				"1": datatypes.String("one"),
				"2": datatypes.Int(2),
				"3": datatypes.Uint64(3),
			},
			Timeout: helpers.PointerOf(15 * time.Second),
		},
	}
}

// the keys "1" overlap
func overlapOneKeyValue() *v1locker.Lock {
	return &v1locker.Lock{
		Spec: &v1locker.LockSpec{
			DBDeifinition: &dbdefinition.TypedKeyValues{
				"1": datatypes.String("one"),
				"4": datatypes.Int(2),
				"5": datatypes.Uint64(3),
			},
			Timeout: helpers.PointerOf(15 * time.Second),
		},
	}
}

func TestExclusiveLocker_ObtainLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It creates a single lock for the provided key values", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockResp := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())
		g.Expect(*lockResp.Spec.Timeout).To(Equal(15 * time.Second)) // default values
		g.Expect(*lockResp.Spec.DBDeifinition).To(Equal(dbdefinition.TypedKeyValues{
			"1": datatypes.String("one"),
			"2": datatypes.Int(2),
			"3": datatypes.Uint64(3),
		}))
		g.Expect(lockResp.State.SessionID).ToNot(Equal(""))
		g.Expect(lockResp.State.LockID).ToNot(Equal(""))
		g.Expect(lockResp.State.LocksHeldOrWaiting).To(Equal(uint64(1)))
		g.Expect(lockResp.State.TimeTillExipre).To(Equal(15 * time.Second))
	})

	t.Run("Context when creating a lock", func(t *testing.T) {
		t.Run("Context after the server context is closed", func(t *testing.T) {
			t.Run("It returns nil without creating the lock", func(t *testing.T) {
				ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
				cancel()

				exclusiveLocker := NewExclusiveLocker()
				g.Expect(exclusiveLocker.Execute(ctx)).ToNot(HaveOccurred())

				lockResp := exclusiveLocker.ObtainLock(testhelpers.NewContextWithMiddlewareSetup(), defaultLockCreateRequest())
				g.Expect(lockResp).To(BeNil())
			})
		})

		t.Run("Context when client context is closed", func(t *testing.T) {
			t.Run("It returns nil and cleans up any potential locks", func(t *testing.T) {
				ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
				defer cancel()

				done := make(chan struct{})
				exclusiveLocker := NewExclusiveLocker()
				go func() {
					defer close(done)
					exclusiveLocker.Execute(ctx)
				}()

				clientCtx, clientCancel := testhelpers.NewCancelContextWithMiddlewareSetup()
				clientCancel()

				lockResp := exclusiveLocker.ObtainLock(clientCtx, defaultLockCreateRequest())
				g.Expect(lockResp).To(BeNil())
			})
		})
	})

	t.Run("Context when waiting for a lock", func(t *testing.T) {
		t.Run("It only unlocks one waiting client", func(t *testing.T) {
			ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
			defer cancel()

			exclusiveLocker := NewExclusiveLocker()
			go func() {
				exclusiveLocker.Execute(ctx)
			}()

			lockResp := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
			g.Expect(lockResp).ToNot(BeNil())

			locked := make(chan struct{})
			for i := 0; i < 10; i++ {
				go func() {
					_ = exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
					select {
					case locked <- struct{}{}:
					case <-ctx.Done():
					}
				}()
			}

			g.Consistently(locked).ShouldNot(Receive())

			// perform an unlock
			claim := &v1locker.LockClaim{
				SessionID: lockResp.State.SessionID,
			}
			g.Expect(claim.Validate()).ToNot(HaveOccurred())
			g.Expect(exclusiveLocker.Release(testhelpers.NewContextWithMiddlewareSetup(), lockResp.State.LockID, claim)).ToNot(HaveOccurred())

			g.Eventually(locked).Should(Receive())
			g.Consistently(locked).ShouldNot(Receive())
		})

		t.Run("It can make many concurent requests to the same lock", func(t *testing.T) {
			ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
			defer cancel()

			exclusiveLocker := NewExclusiveLocker()
			go func() {
				exclusiveLocker.Execute(ctx)
			}()

			wg := new(sync.WaitGroup)
			for i := 0; i < 10_000; i++ {
				wg.Add(1)

				go func() {
					defer wg.Done()
					lockResp := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
					claim := &v1locker.LockClaim{
						SessionID: lockResp.State.SessionID,
					}
					g.Expect(claim.Validate()).ToNot(HaveOccurred())
					exclusiveLocker.Release(ctx, lockResp.State.LockID, claim)
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
			ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
			defer cancel()

			exclusiveLocker := NewExclusiveLocker()
			go func() {
				exclusiveLocker.Execute(ctx)
			}()

			wg := new(sync.WaitGroup)
			for i := 0; i < 10_000; i++ {
				wg.Add(1)

				testReq := &v1locker.Lock{
					Spec: &v1locker.LockSpec{
						DBDeifinition: &dbdefinition.TypedKeyValues{
							fmt.Sprintf("%d", i%5): datatypes.String("doesn't matter 1"),
							fmt.Sprintf("%d", i%6): datatypes.String("doesn't matter 2"),
							fmt.Sprintf("%d", i%7): datatypes.String("doesn't matter 3"),
						},
						Timeout: helpers.PointerOf(15 * time.Second),
					},
				}
				g.Expect(testReq.ValidateSpecOnly()).ToNot(HaveOccurred())

				go func(req *v1locker.Lock) {
					defer wg.Done()
					lockResp := exclusiveLocker.ObtainLock(ctx, req)
					claim := &v1locker.LockClaim{
						SessionID: lockResp.State.SessionID,
					}
					g.Expect(claim.Validate()).ToNot(HaveOccurred())
					exclusiveLocker.Release(ctx, lockResp.State.LockID, claim)
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
				ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()

				exclusiveLocker := NewExclusiveLocker()
				go func() {
					exclusiveLocker.Execute(ctx)
				}()

				wg := new(sync.WaitGroup)
				for i := 0; i < 10; i++ {
					if i == 0 {
						_ = exclusiveLocker.ObtainLock(testhelpers.NewContextWithMiddlewareSetup(), defaultLockCreateRequest())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = exclusiveLocker.ObtainLock(testhelpers.NewContextWithMiddlewareSetup(), defaultLockCreateRequest())
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

		t.Run("Context if a single client is closed", func(t *testing.T) {
			t.Run("It frees up any waiting clients", func(t *testing.T) {
				ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()

				exclusiveLocker := NewExclusiveLocker()
				go func() {
					exclusiveLocker.Execute(ctx)
				}()

				wg := new(sync.WaitGroup)
				for i := 0; i < 10; i++ {
					if i == 0 {
						_ = exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
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

func TestExclusiveLocker_LocksQuery(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It properly queries all locks that were created", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lock1 := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
		g.Expect(lock1).ToNot(BeNil())
		lock2 := exclusiveLocker.ObtainLock(ctx, overlapOneKeyValue())
		g.Expect(lock2).ToNot(BeNil())

		query := &queryassociatedaction.AssociatedActionQuery{}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		locks := exclusiveLocker.LocksQuery(ctx, query)
		g.Expect(len(locks)).To(Equal(2))

		if locks[0].State.SessionID == lock2.State.SessionID {
			g.Expect(locks[0].State.SessionID).To(Equal(lock2.State.SessionID))
			g.Expect(locks[0].State.LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[0].Spec.DBDeifinition).To(Equal(&dbdefinition.TypedKeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[0].Spec.Timeout).To(Equal(helpers.PointerOf(15 * time.Second)))
			g.Expect(locks[0].State.TimeTillExipre).ToNot(Equal(15 * time.Second))

			g.Expect(locks[1].Spec.DBDeifinition).To(Equal(&dbdefinition.TypedKeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[1].Spec.Timeout).To(Equal(helpers.PointerOf(15 * time.Second)))
			g.Expect(locks[1].State.SessionID).To(Equal(lock1.State.SessionID))
			g.Expect(locks[1].State.LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[1].State.TimeTillExipre).ToNot(Equal(0))
		} else {
			g.Expect(locks[1].State.SessionID).To(Equal(lock2.State.SessionID))
			g.Expect(locks[1].State.LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[1].Spec.DBDeifinition).To(Equal(&dbdefinition.TypedKeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[1].Spec.Timeout).To(Equal(helpers.PointerOf(15 * time.Second)))
			g.Expect(locks[1].State.TimeTillExipre).ToNot(Equal(15 * time.Second))

			g.Expect(locks[0].Spec.DBDeifinition).To(Equal(&dbdefinition.TypedKeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[0].Spec.Timeout).To(Equal(helpers.PointerOf(15 * time.Second)))
			g.Expect(locks[0].State.SessionID).To(Equal(lock1.State.SessionID))
			g.Expect(locks[0].State.LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[0].State.TimeTillExipre).ToNot(Equal(0))
		}
	})
}

func TestExclusiveLocker_Heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	query := &queryassociatedaction.AssociatedActionQuery{}
	g.Expect(query.Validate()).ToNot(HaveOccurred())

	t.Run("It returns an error if the LockID does not exist", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat(ctx, "bad id", claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("LockID could not be found"))
	})

	t.Run("It returns an error if the claim is invalid", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()
		lockResp := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat(ctx, lockResp.State.LockID, claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("SessionID for the claim is invalid"))
	})

	t.Run("It keeps the lock around as long as the heartbeats are received", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.Spec.Timeout = helpers.PointerOf(100 * time.Millisecond)

		lockResp := exclusiveLocker.ObtainLock(ctx, lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		// this timer is longer than the heartbeat timeout
		for i := 0; i < 3; i++ {
			time.Sleep(60 * time.Millisecond)
			claim := &v1locker.LockClaim{
				SessionID: lockResp.State.SessionID,
			}
			g.Expect(claim.Validate()).ToNot(HaveOccurred())
			g.Expect(exclusiveLocker.Heartbeat(ctx, lockResp.State.LockID, claim)).To(BeNil())
		}

		locks := exclusiveLocker.LocksQuery(ctx, query)
		g.Expect(len(locks)).To(Equal(1))
	})

	t.Run("It removes the lock if no heartbeats are received and there are no other clients waiting", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		// start the task manager with the locker
		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.Spec.Timeout = helpers.PointerOf(100 * time.Millisecond)

		lockResp := exclusiveLocker.ObtainLock(ctx, lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		g.Eventually(func() int {
			return len(exclusiveLocker.LocksQuery(ctx, query))
		}).Should(Equal(0))
	})

	t.Run("It allows another client to claim the lock if no heartbeats are recieved", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		// start the task manager with the locker
		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.Spec.Timeout = helpers.PointerOf(100 * time.Millisecond)

		locked := make(chan struct{})
		for i := 0; i < 5; i++ {
			go func() {
				_ = exclusiveLocker.ObtainLock(ctx, lockRequest)
				select {
				case locked <- struct{}{}:
				case <-ctx.Done():
				}
			}()
		}

		for i := 0; i < 5; i++ {
			g.Eventually(locked).Should(Receive())
		}

		g.Eventually(func() int {
			return len(exclusiveLocker.LocksQuery(ctx, query))
		}).Should(Equal(0))
	})
}

func TestExclusiveLocker_Release(t *testing.T) {
	g := NewGomegaWithT(t)

	query := &queryassociatedaction.AssociatedActionQuery{}
	g.Expect(query.Validate()).ToNot(HaveOccurred())

	t.Run("It returns an error if the LockID does not exist", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Release(ctx, "bad id", claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("LockID could not be found"))
	})

	t.Run("It returns an error if the claim is invalid", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockResp := exclusiveLocker.ObtainLock(ctx, defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat(ctx, lockResp.State.LockID, claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("SessionID for the claim is invalid"))
	})
}

func TestExclusiveLocker_async(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It properly processes the client records for all timeout and release options", func(t *testing.T) {
		ctx, cancel := testhelpers.NewCancelContextWithMiddlewareSetup()
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		defaultLock := defaultLockCreateRequest()
		defaultLock.Spec.Timeout = helpers.PointerOf(time.Second)

		wg := new(sync.WaitGroup)
		for i := 0; i < 1000; i++ {
			wg.Add(1)

			go func(timeout int) {
				defer wg.Done()

				deadlineCtx, _ := context.WithDeadline(ctx, time.Now().Add(time.Duration((timeout%10)+1)*200*time.Millisecond))
				lock := exclusiveLocker.ObtainLock(deadlineCtx, defaultLock)

				// call with an invalid session ID for release
				if timeout%5 == 0 {
					if lock != nil {
						g.Expect(exclusiveLocker.Release(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: "bad id"})).To(HaveOccurred())
					}
				}

				// call with an invalid session ID for heartbeat
				if timeout%4 == 0 {
					if lock != nil {
						g.Expect(exclusiveLocker.Heartbeat(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: "bad id"})).To(HaveOccurred())
					}
				}

				// call with a good session ID for heartbeat
				if timeout%8 == 0 {
					if lock != nil {
						g.Expect(exclusiveLocker.Heartbeat(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: lock.State.SessionID})).ToNot(HaveOccurred())
					}
				}

				// successful release
				if timeout%7 == 0 {
					if lock != nil {
						g.Expect(exclusiveLocker.Release(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: lock.State.SessionID})).ToNot(HaveOccurred())
					}
				}

				// attempt to hit bad release before next claim
				if timeout%3 == 0 {
					if lock != nil {
						g.Expect(exclusiveLocker.Release(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: lock.State.SessionID})).ToNot(HaveOccurred())
						g.Expect(exclusiveLocker.Release(ctx, lock.State.LockID, &v1locker.LockClaim{SessionID: lock.State.SessionID})).To(HaveOccurred())
					}
				}
			}(i)
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
			// nothing to do here, everything closed properly
		}

		g.Eventually(func() int {
			locks := exclusiveLocker.LocksQuery(ctx, &queryassociatedaction.AssociatedActionQuery{})
			return len(locks)
		}, 3*time.Second).Should(Equal(0)) // race detection can be a bit slow on the check here. so allow for 3 seconds
	})
}
