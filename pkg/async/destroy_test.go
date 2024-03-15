package async

import (
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestDestroySyncer(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can run many read lock operations all at once", func(t *testing.T) {
		destroySync := NewDestroySync()
		startTime := time.Now()

		wg := new(sync.WaitGroup)
		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				g.Expect(destroySync.GuardOperation()).To(BeTrue())

				time.Sleep(100 * time.Millisecond)
				defer destroySync.ClearOperation()
			}()
		}

		wg.Wait()
		g.Expect(time.Since(startTime)).To(BeNumerically("<", time.Second))
	})

	t.Run("It requires GuardOperation calls to finish before WaitDestroy() returns", func(t *testing.T) {
		destroySync := NewDestroySync()

		// start 5 calls
		for i := 0; i < 5; i++ {
			g.Expect(destroySync.GuardOperation()).To(BeTrue())
		}

		canDestroy := make(chan bool)
		go func() {
			canDestroy <- destroySync.WaitDestroy()
		}()

		g.Consistently(canDestroy).ShouldNot(Receive())

		// stop 4 operation calls
		for i := 0; i < 4; i++ {
			destroySync.ClearOperation()
			g.Consistently(canDestroy).ShouldNot(Receive())
		}

		// stop the last operation call
		destroySync.ClearOperation()

		g.Eventually(canDestroy).Should(Receive(BeTrue()))
	})

	t.Run("It returns False on operations when a destroy is taking place", func(t *testing.T) {
		destroySync := NewDestroySync()
		g.Expect(destroySync.WaitDestroy()).To(BeTrue())

		// call to Guard
		g.Expect(destroySync.GuardOperation()).To(BeFalse())

		// 2nd call to WaitDestroy
		g.Expect(destroySync.WaitDestroy()).To(BeFalse())
	})

	t.Run("It clears the guard lock on ClearDestroy", func(t *testing.T) {
		destroySync := NewDestroySync()

		g.Expect(destroySync.WaitDestroy()).To(BeTrue())
		destroySync.ClearDestroy()

		g.Expect(destroySync.GuardOperation()).To(BeTrue())
	})
}
