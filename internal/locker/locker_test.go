package locker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	btreeassociatedfakes "github.com/DanLavine/willow/internal/datastructures/btree_associated/btree_associated_fakes"

	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
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
	})
}

func TestGeneralLocker_ObtainLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It creates a lock for all possible key value combinations", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		counter := 0
		lockCounter := func(_ any) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(7)) // 1, 2, 3, 12, 13, 23, 123
	})

	t.Run("It always grabs locks for a collection of tags in the same sorted order to not deadlock on any keys", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		wg := new(sync.WaitGroup)
		for i := 0; i < 1_000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				disconnectCallback := generalLocker.ObtainLocks(context.Background(), context.Background(), defaultKeyValues())
				disconnectCallback()
			}()
		}

		done := make(chan struct{})

		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			g.Fail("OObtaining locks are stuck")
		}

		counter := 0
		lockCounter := func(_ any) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(0))
	})

	t.Run("It returns a disconnect callback that can be used to release all locks", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())
		disconnectCallback()

		counter := 0
		lockCounter := func(_ any) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(0))
	})

	t.Run("Context when server context is closed", func(t *testing.T) {
		t.Run("It exists early and does not try to lock all entries", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()
			fakeTree := btreeassociatedfakes.NewMockBTreeAssociated(mockController)
			fakeTree.EXPECT().CreateOrFind(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			fakeTree.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			generalLocker := NewGeneralLocker(fakeTree)
			disconnectCallback := generalLocker.ObtainLocks(ctx, context.Background(), defaultKeyValues())
			g.Expect(disconnectCallback).To(BeNil())
		})

		t.Run("It cleans up any newly created locks", func(t *testing.T) {
			generalLocker := NewGeneralLocker(nil)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			disconnectCallback := generalLocker.ObtainLocks(ctx, context.Background(), defaultKeyValues())
			g.Expect(disconnectCallback).To(BeNil())

			counter := 0
			lockCounter := func(_ any) bool {
				counter++
				return true
			}

			g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
			g.Expect(counter).To(Equal(0))
		})
	})

	t.Run("Context when client context is closed", func(t *testing.T) {
		t.Run("It exists early and does not try to lock all entries", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()
			fakeTree := btreeassociatedfakes.NewMockBTreeAssociated(mockController)
			fakeTree.EXPECT().CreateOrFind(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			fakeTree.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			generalLocker := NewGeneralLocker(fakeTree)
			disconnectCallback := generalLocker.ObtainLocks(context.Background(), ctx, defaultKeyValues())
			g.Expect(disconnectCallback).To(BeNil())
		})

		t.Run("It cleans up any newly created locks", func(t *testing.T) {
			generalLocker := NewGeneralLocker(nil)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			disconnectCallback := generalLocker.ObtainLocks(context.Background(), ctx, defaultKeyValues())
			g.Expect(disconnectCallback).To(BeNil())

			counter := 0
			lockCounter := func(_ any) bool {
				counter++
				return true
			}

			g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
			g.Expect(counter).To(Equal(0))
		})
	})
}

func TestGeneralLocker_LockWaitingLogic(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It blocks any requests for a shared key untill the original locks are released", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		locked := make(chan struct{})
		go func() {
			generalLocker.ObtainLocks(context.Background(), context.Background(), overlapOneKeyValue())
			close(locked)
		}()

		g.Consistently(locked).ShouldNot(Receive())

		disconnectCallback()
		g.Eventually(locked).Should(BeClosed())
	})

	t.Run("It only unlocks one waiting client", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		disconnectCallback := generalLocker.ObtainLocks(ctx, context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		locked := make(chan struct{})
		for i := 0; i < 10; i++ {
			go func() {
				generalLocker.ObtainLocks(ctx, context.Background(), overlapOneKeyValue())
				select {
				case locked <- struct{}{}:
				case <-ctx.Done():
				}
			}()
		}

		g.Consistently(locked).ShouldNot(Receive())

		disconnectCallback()
		g.Eventually(locked).Should(Receive())
		g.Consistently(locked).ShouldNot(Receive())
	})

	t.Run("It can make many concurent requests to the same lock", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := new(sync.WaitGroup)
		for i := 10_000; i < 0; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				disconnectCallback := generalLocker.ObtainLocks(ctx, context.Background(), defaultKeyValues())
				disconnectCallback()
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
		generalLocker := NewGeneralLocker(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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
				disconnectCallback := generalLocker.ObtainLocks(ctx, context.Background(), keyValues)
				disconnectCallback()
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
}

func TestGeneralLocker_ListLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It lists all locks that have been created", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		locks := generalLocker.ListLocks()
		g.Expect(len(locks)).To(Equal(7))

		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"1": datatypes.String("one")}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"2": datatypes.Int(2)}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"3": datatypes.Uint64(3)}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"1": datatypes.String("one"), "2": datatypes.Int(2)}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"1": datatypes.String("one"), "3": datatypes.Uint64(3)}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"2": datatypes.Int(2), "3": datatypes.Uint64(3)}, GeneratedFromKeyValues: defaultKeyValues()}))
		g.Expect(locks).To(ContainElement(v1locker.Lock{KeyValues: datatypes.StringMap{"1": datatypes.String("one"), "2": datatypes.Int(2), "3": datatypes.Uint64(3)}, GeneratedFromKeyValues: defaultKeyValues()}))
	})
}
