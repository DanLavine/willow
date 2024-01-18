package memory

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_item_CreateHeartbeater(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can create a new heartbeater process if one does not yet exist", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())
		g.Expect(item.heartbeatProcess).To(Equal(heartbeater))
	})

	t.Run("It panics if the heartbeater is already set", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		g.Expect(func() { item.CreateHeartbeater(func() {}, func() {}) }).To(Panic())
	})

	t.Run("It panics if onTimeout is nil", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)
		g.Expect(func() { item.CreateHeartbeater(func() {}, nil) }).To(Panic())
	})

	t.Run("It accepts a nil onShutdown callback", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)
		g.Expect(func() { item.CreateHeartbeater(nil, func() {}) }).ToNot(Panic())
	})
}

func Test_item_UnsetHeartbeater(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op if the heartbeater is not set", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		item.UnsetHeartbeater()
		g.Expect(item.heartbeatProcess).To(BeNil())
	})

	t.Run("It unsets a set heartbeater", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())
		g.Expect(item.heartbeatProcess).To(Equal(heartbeater))

		item.UnsetHeartbeater()
		g.Expect(item.heartbeatProcess).To(BeNil())
	})
}

func Test_item_StartHeartbeater(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if there is no heartbeater process", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		started := item.StartHeartbeater()
		g.Expect(started).To(BeFalse())
	})

	t.Run("It can start a heartbeater process", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		started := item.StartHeartbeater()
		g.Expect(started).To(BeTrue())
	})

	t.Run("It rerturns false for each of the N+ calls to Start", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		g.Expect(item.StartHeartbeater()).To(BeTrue())

		g.Expect(item.StartHeartbeater()).To(BeFalse())
		g.Expect(item.StartHeartbeater()).To(BeFalse())
		g.Expect(item.StartHeartbeater()).To(BeFalse())
		g.Expect(item.StartHeartbeater()).To(BeFalse())
		g.Expect(item.StartHeartbeater()).To(BeFalse())
	})
}

func Test_item_StopHeartbeater(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if there is no heartbeater process", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		stopped := item.StopHeartbeater()
		g.Expect(stopped).To(BeFalse())
	})

	t.Run("It can stop a running heartbeater process", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		started := item.StartHeartbeater()
		g.Expect(started).To(BeTrue())

		stopped := item.StopHeartbeater()
		g.Expect(stopped).To(BeTrue())
	})

	t.Run("It rerturns false for each of the N+ calls to Stop", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		g.Expect(item.StartHeartbeater()).To(BeTrue())

		g.Expect(item.StopHeartbeater()).To(BeTrue())

		g.Expect(item.StopHeartbeater()).To(BeFalse())
		g.Expect(item.StopHeartbeater()).To(BeFalse())
		g.Expect(item.StopHeartbeater()).To(BeFalse())
		g.Expect(item.StopHeartbeater()).To(BeFalse())
		g.Expect(item.StopHeartbeater()).To(BeFalse())
	})
}

func Test_item_Heartbeater(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns false if there is no heartbeater process", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeated := item.Heartbeat()
		g.Expect(heartbeated).To(BeFalse())
	})

	t.Run("It prevets a heartbeat process from timing out", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		heartbeater := item.CreateHeartbeater(func() {}, func() {})
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		started := item.StartHeartbeater()
		g.Expect(started).To(BeTrue())

		for i := 0; i < 5; i++ {
			time.Sleep(200 * time.Millisecond)
			g.Expect(item.Heartbeat()).To(BeTrue())
		}

		stopped := item.StopHeartbeater()
		g.Expect(stopped).To(BeTrue())
	})

	t.Run("It allows the heartbeat to timeout if not processed in time", func(t *testing.T) {
		item := newItem([]byte(`data`), true, 3, "front", time.Second)

		timedOut := new(atomic.Bool)
		heartbeater := item.CreateHeartbeater(func() {}, func() { timedOut.Store(true) })
		g.Expect(heartbeater).ToNot(BeNil())

		// run the heartbeater like an async task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = heartbeater.Execute(ctx)
		}()

		started := item.StartHeartbeater()
		g.Expect(started).To(BeTrue())

		g.Eventually(timedOut.Load, 2*time.Second).Should(BeTrue())

		g.Expect(item.Heartbeat()).To(BeFalse())
		g.Expect(item.StartHeartbeater()).To(BeFalse())
		g.Expect(item.StopHeartbeater()).To(BeFalse())
	})
}
