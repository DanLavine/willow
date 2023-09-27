package locker

import (
	"context"
	"sync"
	"testing"
	"time"

	btreeassociatedfakes "github.com/DanLavine/willow/internal/datastructures/btree_associated/btree_associated_fakes"

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

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), defaultKeyValues())
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
				disconnectCallback := generalLocker.ObtainLocks(context.Background(), defaultKeyValues())
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

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), defaultKeyValues())
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

	t.Run("Context when cancel is closed with creates", func(t *testing.T) {
		t.Run("It exists early and does not try to lock all entries", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()
			fakeTree := btreeassociatedfakes.NewMockBTreeAssociated(mockController)
			fakeTree.EXPECT().CreateOrFind(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			fakeTree.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			generalLocker := NewGeneralLocker(fakeTree)
			disconnectCallback := generalLocker.ObtainLocks(ctx, defaultKeyValues())
			g.Expect(disconnectCallback).To(BeNil())
		})

		t.Run("Create cleans up any newly created locks", func(t *testing.T) {
			generalLocker := NewGeneralLocker(nil)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			disconnectCallback := generalLocker.ObtainLocks(ctx, defaultKeyValues())
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

func TestGeneralLocker_FreeLocks(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It deletes all possible key value pairs", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		_ = generalLocker.ObtainLocks(context.Background(), defaultKeyValues())
		generalLocker.FreeLocks(defaultKeyValues())

		counter := 0
		lockCounter := func(_ any) bool {
			counter++
			return true
		}

		g.Expect(generalLocker.locks.Query(query.Select{}, lockCounter)).ToNot(HaveOccurred())
		g.Expect(counter).To(Equal(0))
	})
}

func TestGeneralLocker_LockWaitingLogic(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It blocks any requests for a shared key untill the original locks are released", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)

		disconnectCallback := generalLocker.ObtainLocks(context.Background(), defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		locked := make(chan struct{})
		go func() {
			generalLocker.ObtainLocks(context.Background(), overlapOneKeyValue())
			close(locked)
		}()

		g.Consistently(locked).ShouldNot(Receive())

		generalLocker.FreeLocks(defaultKeyValues())
		g.Eventually(locked).Should(BeClosed())
	})

	t.Run("It only unlocks one waiting client", func(t *testing.T) {
		generalLocker := NewGeneralLocker(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		disconnectCallback := generalLocker.ObtainLocks(ctx, defaultKeyValues())
		g.Expect(disconnectCallback).ToNot(BeNil())

		locked := make(chan struct{})
		for i := 0; i < 10; i++ {
			go func() {
				generalLocker.ObtainLocks(ctx, overlapOneKeyValue())
				select {
				case locked <- struct{}{}:
				case <-ctx.Done():
				}
			}()
		}

		g.Consistently(locked).ShouldNot(Receive())

		generalLocker.FreeLocks(defaultKeyValues())
		g.Eventually(locked).Should(Receive())
		g.Consistently(locked).ShouldNot(Receive())
	})

}
