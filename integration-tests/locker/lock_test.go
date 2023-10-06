package locker_integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
	"golang.org/x/net/http2"
)

func Test_Lock(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can request a lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("It block a second requst for the same lock tags", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
			lockerClient.Do(r)
			close(done)
		}()

		g.Consistently(done).ShouldNot(BeClosed())
	})

	t.Run("It block a second requst if they share any of the tags from a previous request that has a lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest1 := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data1, err := json.Marshal(lockRequest1)
		g.Expect(err).ToNot(HaveOccurred())

		lockRequest2 := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data2, err := json.Marshal(lockRequest2)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data1))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data2))
			lockerClient.Do(r)
			close(done)
		}()

		g.Consistently(done).ShouldNot(BeClosed())
	})

	t.Run("It releases all locks if the original client disconnects", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}

		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// create a new client for the second request
		newClient := &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: lockerClient.Transport(),
			},
		}

		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
			newClient.Do(r)
			close(done)
		}()

		g.Consistently(done).ShouldNot(BeClosed())

		// close the first client
		lockerClient.Disconnect()

		// the second request should now process
		g.Eventually(done).Should(BeClosed())
	})
}

func TestLocker_Delete_API(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It releases all the locks on a free request for a new client to obtain the lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// setup the first lock
		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// start a second request that is blocked
		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
			lockerClient.Do(r)
			close(done)
		}()
		g.Consistently(done).ShouldNot(Receive())

		// free the first request
		releaseRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		newData, err := json.Marshal(releaseRequest)
		g.Expect(err).ToNot(HaveOccurred())

		releaseLock, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", testConstruct.LockerClient.Address()), bytes.NewBuffer(newData))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(releaseLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

		// ensure that the blocked client now has the lock
		g.Eventually(done).Should(BeClosed())
	})

	t.Run("It does not release a lock if the original client doesn't send the unlock requuest", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// setup the first lock
		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// start a second request that is blocked
		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
			lockerClient.Do(r)
			close(done)
		}()
		g.Consistently(done).ShouldNot(Receive())

		// free the first request
		releaseRequest := *&v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		newData, err := json.Marshal(releaseRequest)
		g.Expect(err).ToNot(HaveOccurred())

		releaseLock, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", lockerClient.Address()), bytes.NewBuffer(newData))
		g.Expect(err).ToNot(HaveOccurred())

		newClient := &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: lockerClient.Transport(),
			},
		}
		resp, err = newClient.Do(releaseLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

		// ensure that the blocked client is still blocked
		g.Consistently(done).ShouldNot(Receive())
		runtime.KeepAlive(lockerClient)
	})
}

func TestLocker_List_API(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It lists all locks currently held", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// setup the first lock
		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		// list all the locks
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locker/list", testConstruct.LockerClient.Address()), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.LockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(3))
	})
}

func TestLocker_Async_API_Threading_Checks(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can request the same lock many times", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		defer func() {
			fmt.Println("DSL start")
			fmt.Println(string(testConstruct.Session.Out.Contents()))
			fmt.Println(string(testConstruct.Session.Err.Contents()))
			fmt.Println("DSL end")
		}()

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		releaseRequest := v1locker.LockRequest{
			KeyValues: datatypes.StringMap{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		newData, err := json.Marshal(releaseRequest)
		g.Expect(err).ToNot(HaveOccurred())

		wg := new(sync.WaitGroup)
		for i := 0; i < 300; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// create the lock
				obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
				g.Expect(err).ToNot(HaveOccurred())

				resp, err := lockerClient.Do(obtainLock)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

				// release the lock
				releaseLock, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", testConstruct.LockerClient.Address()), bytes.NewBuffer(newData))
				g.Expect(err).ToNot(HaveOccurred())

				resp, err = lockerClient.Do(releaseLock)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			}()
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-time.After(10 * time.Second):
			g.Fail("filed")
		case <-done:
			// nothing to do here
		}

		// ensure all the locks are cleaned up
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locker/list", testConstruct.LockerClient.Address()), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.LockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(0))
	})
}
