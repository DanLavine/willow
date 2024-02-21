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
	"github.com/DanLavine/willow/pkg/models/api"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func setupClient(g *GomegaWithT, url string) lockerclient.LockerClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:             url,
		ContentEncoding: api.ContentTypeJSON,
		CAFile:          filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}
	g.Expect(cfg.Validate()).ToNot(HaveOccurred())

	lockerClient, err := lockerclient.NewLockClient(cfg)
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

		lockerClient := setupClient(g, testConstruct.ServerURL)

		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lock, err := lockerClient.ObtainLock(ctx, lockRequest, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		g.Expect(lock.Release()).ToNot(HaveOccurred())
	})

	t.Run("It can aquire multiple locks with multipe KeyValues", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := setupClient(g, testConstruct.ServerURL)

		lockRequest1 := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		lockRequest2 := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key3": datatypes.String("key two"),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lock1, err := lockerClient.ObtainLock(ctx, lockRequest1, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock1).ToNot(BeNil())

		lock2, err := lockerClient.ObtainLock(ctx, lockRequest2, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock2).ToNot(BeNil())

		g.Expect(lock1.Release()).ToNot(HaveOccurred())
		g.Expect(lock2.Release()).ToNot(HaveOccurred())
	})

	t.Run("It blocks a second requst for the same lock tags", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		// client 1
		lockerClient1 := setupClient(g, testConstruct.ServerURL)

		// client 2
		lockerClient2 := setupClient(g, testConstruct.ServerURL)

		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 1st lock goes fine
		lock, err := lockerClient1.ObtainLock(ctx, lockRequest, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// 2nd lock blocks until the first lock is released
		done := make(chan struct{})
		go func() {
			defer close(done)
			lock, err = lockerClient2.ObtainLock(ctx, lockRequest, nil)
		}()

		g.Consistently(done).ShouldNot(BeClosed())

		// release the lock
		g.Expect(lock.Release()).ToNot(HaveOccurred()) // 1st lock release
		g.Eventually(done).Should(BeClosed())
		g.Expect(lock.Release()).ToNot(HaveOccurred()) // 2nd lock release
	})

	t.Run("It can release the clients when the server is shutting down", func(t *testing.T) {
		testConstruct.StartLocker(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// client 1
		lockerClient1 := setupClient(g, testConstruct.ServerURL)

		// client 2
		lockerClient2 := setupClient(g, testConstruct.ServerURL)

		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		// 1st lock goes fine
		lock, err := lockerClient1.ObtainLock(ctx, lockRequest, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// 2nd lock blocks until the first lock is released
		done := make(chan struct{})
		var lock2 lockerclient.Lock
		var err2 error
		go func() {
			defer close(done)
			lock2, err2 = lockerClient2.ObtainLock(ctx, lockRequest, nil)
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

		lockerClient := setupClient(g, testConstruct.ServerURL)

		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
			LockTimeout: 100 * time.Millisecond,
		}

		lock, err := lockerClient.ObtainLock(ctx, lockRequest, nil)
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
		lockerClient := setupClient(g, testConstruct.ServerURL)

		// setup the first lock
		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lock, err := lockerClient.ObtainLock(ctx, lockRequest, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// setup the second lock
		lockRequest = &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key3": datatypes.String("key three"),
			},
		}
		lock, err = lockerClient.ObtainLock(ctx, lockRequest, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).ToNot(BeNil())

		// list all the locks
		query := v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())
		data, err := query.EncodeJSON()
		g.Expect(err).ToNot(HaveOccurred())

		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(data))
		listLocks.Header.Add("Content-Type", "application/json")
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := httpsClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.Locks{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks)).To(Equal(2))

		if reflect.DeepEqual(locks[0].KeyValues.SortedKeys(), []string{"key1", "key2"}) {
			g.Expect(locks[1].KeyValues.SortedKeys()).To(Equal([]string{"key1", "key3"}))
		} else {
			g.Expect(locks[1].KeyValues.SortedKeys()).To(Equal([]string{"key1", "key2"}))
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

		lockRequest := &v1locker.LockCreateRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := new(sync.WaitGroup)
		for i := 0; i < 300; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// create a new client
				lockerClient := setupClient(g, testConstruct.ServerURL)

				// create the lock
				lock, err := lockerClient.ObtainLock(ctx, lockRequest, nil)
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
		query := v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())
		data, err := query.EncodeJSON()
		g.Expect(err).ToNot(HaveOccurred())

		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(data))
		listLocks.Header.Add("Content-Type", "application/json")
		g.Expect(err).ToNot(HaveOccurred())

		manualClient := testConstruct.ServerClient
		resp, err := manualClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.Locks{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks)).To(Equal(0))
	})

	t.Run("It can request many differnt lock combinations many times", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := new(sync.WaitGroup)
		for i := 0; i < 300; i++ {
			lockRequest := &v1locker.LockCreateRequest{
				KeyValues: datatypes.KeyValues{
					"key1": datatypes.String(fmt.Sprintf("%d", i%5)),
					"key2": datatypes.String(fmt.Sprintf("%d", i%17)),
				},
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				// setup new client
				lockerClient := setupClient(g, testConstruct.ServerURL)

				// create the lock
				lock, err := lockerClient.ObtainLock(ctx, lockRequest, nil)
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
		query := v1common.AssociatedQuery{
			AssociatedKeyValues: datatypes.AssociatedKeyValuesQuery{},
		}
		g.Expect(query.Validate()).ToNot(HaveOccurred())
		data, err := query.EncodeJSON()
		g.Expect(err).ToNot(HaveOccurred())

		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locks", testConstruct.ServerURL), bytes.NewBuffer(data))
		listLocks.Header.Add("Content-Type", "application/json")
		g.Expect(err).ToNot(HaveOccurred())

		manualClient := testConstruct.ServerClient
		resp, err := manualClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.Locks{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks)).To(Equal(0))
	})

}
