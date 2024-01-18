package heartbeater

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_heartbeater_Execute(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if onTimeout is nil", func(t *testing.T) {
		heartbeat, err := New(time.Second, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("onTimeout cannot be nil"))
		g.Expect(heartbeat).To(BeNil())
	})

	t.Run("It stops processing when the stop operation is called without start", func(t *testing.T) {
		heartbeat, err := New(time.Second, func() { panic("should not shutdown") }, func() { panic("should not timeout") })
		g.Expect(err).ToNot(HaveOccurred())

		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = heartbeat.Execute(context.Background())
		}()

		stopped := heartbeat.Stop()
		g.Expect(stopped).To(BeTrue())

		g.Eventually(done).Should(BeClosed())
	})

	t.Run("Context when the start operation is called", func(t *testing.T) {
		t.Run("It cancels and runs the onShutdown callback when the Execute Context is closed", func(t *testing.T) {
			calledShutdown := false
			heartbeat, err := New(time.Second, func() { calledShutdown = true }, func() { panic("should not timeout") })
			g.Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = heartbeat.Execute(ctx)
			}()

			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())

			g.Eventually(done).Should(BeClosed())
			g.Expect(calledShutdown).To(BeTrue())
		})

		t.Run("It stops executing when Stop is called", func(t *testing.T) {
			heartbeat, err := New(time.Second, func() { panic("shutdown called") }, func() { panic("timeout should not be called") })
			g.Expect(err).ToNot(HaveOccurred())

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = heartbeat.Execute(context.Background())
			}()

			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())

			stopped := heartbeat.Stop()
			g.Expect(stopped).To(BeTrue())

			g.Eventually(done).Should(BeClosed())
		})

		t.Run("It returns an false on multiple calls to Start", func(t *testing.T) {
			heartbeat, err := New(time.Second, func() { panic("shutdown called") }, func() { panic("timeout should not be called") })
			g.Expect(err).ToNot(HaveOccurred())

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = heartbeat.Execute(context.Background())
			}()

			// start 1
			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())

			// start 2 with an error
			started = heartbeat.Start()
			g.Expect(started).To(BeFalse())

			// stop processing
			stopped := heartbeat.Stop()
			g.Expect(stopped).To(BeTrue())

			g.Eventually(done).Should(BeClosed())
		})

		t.Run("It can continue to proccess as long as heartbats are received", func(t *testing.T) {
			heartbeat, err := New(200*time.Millisecond, nil, func() { panic("should not timeout when heartbeating") })
			g.Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = heartbeat.Execute(ctx)
			}()

			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())

			for i := 0; i < 5; i++ {
				heartbeated := heartbeat.Heartbeat()
				g.Expect(heartbeated).To(BeTrue())
				time.Sleep(50 * time.Millisecond)
			}
		})

		t.Run("It returns false if heartbeating on a processes that already stopped", func(t *testing.T) {
			heartbeat, err := New(time.Second, func() { panic("should not shutdown") }, func() { panic("should not timeout") })
			g.Expect(err).ToNot(HaveOccurred())

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = heartbeat.Execute(context.Background())
			}()

			stopped := heartbeat.Stop()
			g.Expect(stopped).To(BeTrue())
			g.Eventually(done).Should(BeClosed())

			passedHeartbeat := heartbeat.Heartbeat()
			g.Expect(passedHeartbeat).To(BeFalse())
		})

		t.Run("It runs the onTimeout callback when timing out", func(t *testing.T) {
			timedOut := false
			heartbeat, err := New(time.Nanosecond, nil, func() { timedOut = true })
			g.Expect(err).ToNot(HaveOccurred())

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = heartbeat.Execute(context.Background())
			}()

			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())
			g.Eventually(done).Should(BeClosed())

			g.Expect(timedOut).To(BeTrue())
		})
	})
}

func Test_heartbeater_GetLastHeartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns currrent time if the process has not yet started heartbeating", func(t *testing.T) {
		heartbeat, err := New(time.Second, nil, func() {})
		g.Expect(err).ToNot(HaveOccurred())

		currentTime := time.Now()
		heartbeatTime := heartbeat.GetLastHeartbeat()
		g.Expect(heartbeatTime.Sub(currentTime).Seconds()).To(BeNumerically("<", time.Second))
	})

	t.Run("Context when heartbeating", func(t *testing.T) {
		t.Run("It returns the time of the last actual heartbeat", func(t *testing.T) {
			heartbeat, err := New(200*time.Millisecond, nil, func() {})
			g.Expect(err).ToNot(HaveOccurred())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				_ = heartbeat.Execute(ctx)
			}()

			started := heartbeat.Start()
			g.Expect(started).To(BeTrue())

			startTime := time.Now()
			for i := 0; i < 5; i++ {
				heartbeat.Heartbeat()
				time.Sleep(50 * time.Millisecond)
			}

			endTime := heartbeat.GetLastHeartbeat()
			g.Expect(endTime.After(startTime)).To(BeTrue())
			g.Expect(endTime.Sub(startTime)).To(BeNumerically(">", 200*time.Millisecond))
		})
	})
}
