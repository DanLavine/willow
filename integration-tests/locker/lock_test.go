package locker_integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func setupClient(g *GomegaWithT, ctx context.Context, url string) lockerclient.LockerClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:           url,
		CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	lockerClient, err := lockerclient.NewLockerClient(ctx, cfg, nil)
	g.Expect(err).ToNot(HaveOccurred())

	return lockerClient
}

func Test_Lock(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can aquire a lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		lock, err := lockerClient.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		g.Expect(lock.Release()).ToNot(HaveOccurred())
	})

	t.Run("It can aquire multiple locks with multipe KeyValues", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest1 := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		lockRequest2 := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key3": datatypes.String("key two"),
			},
		}

		lock1, err := lockerClient.ObtainLock(ctx, lockRequest1)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock1).ToNot(BeNil())

		lock2, err := lockerClient.ObtainLock(ctx, lockRequest2)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock2).ToNot(BeNil())

		g.Expect(lock1.Release()).ToNot(HaveOccurred())
		g.Expect(lock2.Release()).ToNot(HaveOccurred())
	})

	t.Run("It blocks a second requst for the same lock tags", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		// client 1
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient1 := setupClient(g, ctx, testConstruct.ServerURL)

		// client 2
		lockerClient2 := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		// 1st lock goes fine
		lock, err := lockerClient1.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// 2nd lock blocks until the first lock is released
		done := make(chan struct{})
		go func() {
			defer close(done)
			lock, err = lockerClient2.ObtainLock(ctx, lockRequest)
		}()

		g.Consistently(done).ShouldNot(BeClosed())

		// release the lock
		g.Expect(lock.Release()).ToNot(HaveOccurred())
		g.Eventually(done).Should(BeClosed())
		g.Expect(lock.Release()).ToNot(HaveOccurred())
	})

	t.Run("It can release the clients when the server is shutting down", func(t *testing.T) {
		testConstruct.StartLocker(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// client 1
		lockerClient1 := setupClient(g, ctx, testConstruct.ServerURL)

		// client 2
		lockerClient2 := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		// 1st lock goes fine
		lock, err := lockerClient1.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// 2nd lock blocks until the first lock is released
		done := make(chan struct{})
		var lock2 lockerclient.Lock
		var err2 error
		go func() {
			defer close(done)
			lock2, err2 = lockerClient2.ObtainLock(ctx, lockRequest)
		}()

		g.Consistently(done, time.Second).ShouldNot(BeClosed())

		// release the lock
		testConstruct.Shutdown(g)
		g.Eventually(lock.Done).Should(BeClosed()) // 1st lock should be released
		g.Eventually(done).Should(BeClosed())      // 2nd lock attempt has failed
		g.Expect(lock2).To(BeNil())                // 2nd lock is nil
		g.Expect(err2).To(HaveOccurred())
	})

	t.Run("It keeps a lock by heartbeating", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
			Timeout: 100 * time.Millisecond,
		}

		lock, err := lockerClient.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		g.Consistently(lock.Done(), time.Second).ShouldNot(BeClosed())

		g.Expect(lock.Release()).ToNot(HaveOccurred())
		g.Expect(lock.Done()).To(BeClosed())
	})
}

// This would really be an admin API, and I don't think the client should have this implemented?
// if it was a "list", it would just list the locks that the client currently holds
func TestLocker_List_API(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It lists all locks currently held", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		httpsClient := testConstruct.ServerClient

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		// setup the first lock
		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		lock, err := lockerClient.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// setup the second lock
		lockRequest = v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key3": datatypes.String("key three"),
			},
		}
		lock, err = lockerClient.ObtainLock(ctx, lockRequest)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// list all the locks
		query := v1common.GeneralAssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(query.ToBytes()))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := httpsClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(2))

		if reflect.DeepEqual(locks.Locks[0].KeyValues.SoretedKeys(), []string{"key1", "key2"}) {
			g.Expect(locks.Locks[1].KeyValues.SoretedKeys()).To(Equal([]string{"key1", "key3"}))
		} else {
			g.Expect(locks.Locks[1].KeyValues.SoretedKeys()).To(Equal([]string{"key1", "key2"}))
		}
	})
}

func TestLocker_Async_API_Threading_Checks(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can request the same lock many times", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		wg := new(sync.WaitGroup)
		for i := 0; i < 300; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// create the lock
				lock, err := lockerClient.ObtainLock(ctx, lockRequest)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(lock).ToNot(BeNil())

				// release the lock
				g.Expect(lock.Release()).ToNot(HaveOccurred())
			}()
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-time.After(10 * time.Second):
			g.Fail("failed")
		case <-done:
			// nothing to do here
		}

		// ensure all the locks are cleaned up
		query := v1common.GeneralAssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(query.ToBytes()))
		g.Expect(err).ToNot(HaveOccurred())

		manualClient := testConstruct.ServerClient
		resp, err := manualClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(0))
	})

	t.Run("It can request many differnt lock combinations many times", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lockerClient := setupClient(g, ctx, testConstruct.ServerURL)

		wg := new(sync.WaitGroup)
		for i := 0; i < 300; i++ {
			lockRequest := v1locker.CreateLockRequest{
				KeyValues: datatypes.KeyValues{
					"key1": datatypes.String(fmt.Sprintf("%d", i%5)),
					"key2": datatypes.String(fmt.Sprintf("%d", i%17)),
				},
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				// create the lock
				lock, err := lockerClient.ObtainLock(ctx, lockRequest)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(lock).ToNot(BeNil())

				// release the lock
				g.Expect(lock.Release()).ToNot(HaveOccurred())
			}()
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-time.After(10 * time.Second):
			g.Fail("failed")
		case <-done:
			// nothing to do here
		}

		// ensure all the locks are cleaned up
		query := v1common.GeneralAssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(query.ToBytes()))
		g.Expect(err).ToNot(HaveOccurred())

		manualClient := testConstruct.ServerClient
		resp, err := manualClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(0))
	})

}
