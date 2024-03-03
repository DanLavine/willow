package lockmanager

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	. "github.com/onsi/gomega"
)

func defaultLockCreateRequest() *v1locker.LockCreateRequest {
	return &v1locker.LockCreateRequest{
		KeyValues: datatypes.KeyValues{
			"1": datatypes.String("one"),
			"2": datatypes.Int(2),
			"3": datatypes.Uint64(3),
		},
		LockTimeout: 15 * time.Second,
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
		LockTimeout: 15 * time.Second,
	}
}

func TestExclusiveLocker_ObtainLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It creates a single lock for the provided key values", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockResp := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())
		g.Expect(lockResp.SessionID).ToNot(Equal(""))
		g.Expect(lockResp.LockTimeout).To(Equal(15 * time.Second)) // default values
	})

	t.Run("Context when creating a lock", func(t *testing.T) {
		t.Run("Context after the server context is closed", func(t *testing.T) {
			t.Run("It reeturns nil without creating the lock", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				exclusiveLocker := NewExclusiveLocker()
				g.Expect(exclusiveLocker.Execute(ctx)).ToNot(HaveOccurred())

				lockResp := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
				g.Expect(lockResp).To(BeNil())
			})
		})

		t.Run("Context when client context is closed", func(t *testing.T) {
			t.Run("It returns nil and cleans up any potential locks", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				done := make(chan struct{})
				exclusiveLocker := NewExclusiveLocker()
				go func() {
					defer close(done)
					exclusiveLocker.Execute(ctx)
				}()

				clientCtx, clientCancel := context.WithCancel(context.Background())
				clientCancel()

				lockResp := exclusiveLocker.ObtainLock(clientCtx, defaultLockCreateRequest())
				g.Expect(lockResp).To(BeNil())
			})
		})
	})

	t.Run("Context when waiting for a lock", func(t *testing.T) {
		t.Run("It only unlocks one waiting client", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
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
				SessionID: lockResp.SessionID,
			}
			g.Expect(claim.Validate()).ToNot(HaveOccurred())
			g.Expect(exclusiveLocker.Release(lockResp.LockID, claim)).ToNot(HaveOccurred())

			g.Eventually(locked).Should(Receive())
			g.Consistently(locked).ShouldNot(Receive())
		})

		t.Run("It can make many concurent requests to the same lock", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			exclusiveLocker := NewExclusiveLocker()
			go func() {
				exclusiveLocker.Execute(ctx)
			}()

			wg := new(sync.WaitGroup)
			for i := 10_000; i < 0; i++ {
				wg.Add(1)

				go func() {
					defer wg.Done()
					lockResp := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
					claim := &v1locker.LockClaim{
						SessionID: lockResp.SessionID,
					}
					g.Expect(claim.Validate()).ToNot(HaveOccurred())
					exclusiveLocker.Release(lockResp.LockID, claim)
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

			exclusiveLocker := NewExclusiveLocker()
			go func() {
				exclusiveLocker.Execute(ctx)
			}()

			wg := new(sync.WaitGroup)
			for i := 10_000; i < 0; i++ {
				wg.Add(1)

				testReq := &v1locker.LockCreateRequest{
					KeyValues: datatypes.KeyValues{
						fmt.Sprintf("%d", i%5): datatypes.String("doesn't matter 1"),
						fmt.Sprintf("%d", i%6): datatypes.String("doesn't matter 2"),
						fmt.Sprintf("%d", i%7): datatypes.String("doesn't matter 3"),
					},
					LockTimeout: 15 * time.Second,
				}
				g.Expect(testReq.Validate()).ToNot(HaveOccurred())

				go func(req *v1locker.LockCreateRequest) {
					defer wg.Done()
					lockResp := exclusiveLocker.ObtainLock(context.Background(), req)
					claim := &v1locker.LockClaim{
						SessionID: lockResp.SessionID,
					}
					g.Expect(claim.Validate()).ToNot(HaveOccurred())
					exclusiveLocker.Release(lockResp.LockID, claim)
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

				exclusiveLocker := NewExclusiveLocker()
				go func() {
					exclusiveLocker.Execute(ctx)
				}()

				wg := new(sync.WaitGroup)
				for i := 10; i < 0; i++ {
					if i == 0 {
						_ = exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
					} else {
						wg.Add(1)

						go func() {
							defer wg.Done()
							_ = exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
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

				exclusiveLocker := NewExclusiveLocker()
				go func() {
					exclusiveLocker.Execute(ctx)
				}()

				wg := new(sync.WaitGroup)
				for i := 10; i < 0; i++ {
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lock1 := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lock1).ToNot(BeNil())
		lock2 := exclusiveLocker.ObtainLock(context.Background(), overlapOneKeyValue())
		g.Expect(lock2).ToNot(BeNil())

		query := datatypes.AssociatedKeyValuesQuery{}
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		locks := exclusiveLocker.LocksQuery(&v1.AssociatedQuery{AssociatedKeyValues: query})
		g.Expect(len(locks)).To(Equal(2))

		if locks[0].SessionID == lock2.SessionID {
			g.Expect(locks[0].SessionID).To(Equal(lock2.SessionID))
			g.Expect(locks[0].LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[0].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[0].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[0].TimeTillExipre).ToNot(Equal(0))

			g.Expect(locks[1].SessionID).To(Equal(lock1.SessionID))
			g.Expect(locks[1].LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[1].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[1].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[1].TimeTillExipre).ToNot(Equal(0))
		} else {
			g.Expect(locks[1].SessionID).To(Equal(lock2.SessionID))
			g.Expect(locks[1].LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[1].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "4": datatypes.Int(2), "5": datatypes.Uint64(3)}))
			g.Expect(locks[1].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[1].TimeTillExipre).ToNot(Equal(0))

			g.Expect(locks[0].SessionID).To(Equal(lock1.SessionID))
			g.Expect(locks[0].LocksHeldOrWaiting).To(Equal(uint64(1)))
			g.Expect(locks[0].KeyValues).To(Equal(datatypes.KeyValues{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}))
			g.Expect(locks[0].Timeout).To(Equal(15 * time.Second))
			g.Expect(locks[0].TimeTillExipre).ToNot(Equal(0))
		}
	})
}

func TestExclusiveLocker_Heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	query := datatypes.AssociatedKeyValuesQuery{}
	g.Expect(query.Validate()).ToNot(HaveOccurred())
	generalQuery := &v1.AssociatedQuery{AssociatedKeyValues: query}

	t.Run("It returns an error if the LockID does not exist", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat("bad id", claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("LockID could not be found"))
	})

	t.Run("It returns an error if the claim is invalid", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()
		lockResp := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat(lockResp.LockID, claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("SessionID for the claim is invalid"))
	})

	t.Run("It keeps the lock around as long as the heartbeats are received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.LockTimeout = 100 * time.Millisecond

		lockResp := exclusiveLocker.ObtainLock(context.Background(), lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		// this timer is longer than the heartbeat timeout
		for i := 0; i < 3; i++ {
			time.Sleep(60 * time.Millisecond)
			claim := &v1locker.LockClaim{
				SessionID: lockResp.SessionID,
			}
			g.Expect(claim.Validate()).ToNot(HaveOccurred())
			g.Expect(exclusiveLocker.Heartbeat(lockResp.LockID, claim)).To(BeNil())
		}

		locks := exclusiveLocker.LocksQuery(generalQuery)
		g.Expect(len(locks)).To(Equal(1))
	})

	t.Run("It removes the lock if no heartbeats are received", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// start the task manager with the locker
		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockRequest := defaultLockCreateRequest()
		lockRequest.LockTimeout = 100 * time.Millisecond

		lockResp := exclusiveLocker.ObtainLock(context.Background(), lockRequest)
		g.Expect(lockResp).ToNot(BeNil())

		g.Eventually(func() int {
			return len(exclusiveLocker.LocksQuery(generalQuery))
		}).Should(Equal(0))
	})
}

func TestExclusiveLocker_Release(t *testing.T) {
	g := NewGomegaWithT(t)

	query := datatypes.AssociatedKeyValuesQuery{}
	g.Expect(query.Validate()).ToNot(HaveOccurred())

	t.Run("It returns an error if the LockID does not exist", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Release("bad id", claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("LockID could not be found"))
	})

	t.Run("It returns an error if the claim is invalid", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exclusiveLocker := NewExclusiveLocker()
		go func() {
			exclusiveLocker.Execute(ctx)
		}()

		lockResp := exclusiveLocker.ObtainLock(context.Background(), defaultLockCreateRequest())
		g.Expect(lockResp).ToNot(BeNil())

		claim := &v1locker.LockClaim{
			SessionID: "nope",
		}
		g.Expect(claim.Validate()).ToNot(HaveOccurred())

		err := exclusiveLocker.Heartbeat(lockResp.LockID, claim)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("SessionID for the claim is invalid"))
	})
}
