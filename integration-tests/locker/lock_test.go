package locker_integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func Test_Lock(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can request a lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
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
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	})

	t.Run("It blocks a second requst for the same lock tags", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
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
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", lockerClient.Address()), bytes.NewBuffer(data))
			lockerClient.Do(r)
			close(done)
		}()

		g.Consistently(done).ShouldNot(BeClosed())
	})
}

func TestLocker_Delete_API(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It releases the lock on a free request for a new client to obtain the lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// setup the first lock
		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
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
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		obtainLockResp := v1locker.DeleteLockRequest{}
		respData, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(json.Unmarshal(respData, &obtainLockResp)).ToNot(HaveOccurred())
		g.Expect(obtainLockResp.SessionID).ToNot(Equal(""))

		// start a second request that is blocked
		done := make(chan struct{})
		go func() {
			r, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
			lockerClient.Do(r)
			close(done)
		}()
		g.Consistently(done).ShouldNot(Receive())

		// free the first request
		newData, err := json.Marshal(obtainLockResp)
		g.Expect(err).ToNot(HaveOccurred())

		releaseLock, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", testConstruct.LockerClient.Address()), bytes.NewBuffer(newData))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(releaseLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

		// ensure that the blocked client now has the lock
		g.Eventually(done).Should(BeClosed())
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
		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
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
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		// setup the second lock
		lockRequest = v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key3": datatypes.String("key three"),
			},
		}
		data, err = json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err = http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		// list all the locks
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locker/list", testConstruct.LockerClient.Address()), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
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

func TestLocker_Heartbeat_API(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It returns an error when heartbeating a sessionID that does not exist", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		heartbeatRequest := v1locker.HeartbeatLocksRequst{
			SessionIDs: []string{"nope"},
		}

		data, err := json.Marshal(heartbeatRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/heartbeat", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	t.Run("It heartbeat a lock to keep the lock", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// create the lock
		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
			Timeout: time.Second,
		}

		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		obtainLockResp := v1locker.DeleteLockRequest{}
		respData, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(json.Unmarshal(respData, &obtainLockResp)).ToNot(HaveOccurred())
		g.Expect(obtainLockResp.SessionID).ToNot(Equal(""))

		// continueoously heartbeat for longer than the time expiration
		for i := 0; i < 10; i++ {
			heartbeatRequest := v1locker.HeartbeatLocksRequst{
				SessionIDs: []string{string(obtainLockResp.SessionID)},
			}

			data, err := json.Marshal(heartbeatRequest)
			g.Expect(err).ToNot(HaveOccurred())

			heartbeatReq, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/heartbeat", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = lockerClient.Do(heartbeatReq)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}

		// ensure the lock still exists
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locker/list", testConstruct.LockerClient.Address()), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(1))
	})

	t.Run("It releases a lock if the heartbeats do not happen often enough", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		// create the lock
		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
			Timeout: 100 * time.Millisecond,
		}

		data, err := json.Marshal(lockRequest)
		g.Expect(err).ToNot(HaveOccurred())

		obtainLock, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/create", testConstruct.LockerClient.Address()), bytes.NewBuffer(data))
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := lockerClient.Do(obtainLock)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		obtainLockResp := v1locker.DeleteLockRequest{}
		respData, err := io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(json.Unmarshal(respData, &obtainLockResp)).ToNot(HaveOccurred())
		g.Expect(obtainLockResp.SessionID).ToNot(Equal(""))

		// sleep to let the lock expire
		time.Sleep(200 * time.Millisecond)

		// ensure the lock still exists
		listLocks, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/locker/list", testConstruct.LockerClient.Address()), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err = lockerClient.Do(listLocks)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

		data, err = io.ReadAll(resp.Body)
		g.Expect(err).ToNot(HaveOccurred())

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(0))
	})
}

func TestLocker_Async_API_Threading_Checks(t *testing.T) {
	g := NewGomegaWithT(t)

	testConstruct := NewIntrgrationLockerTestConstruct(g)
	defer testConstruct.Cleanup(g)

	t.Run("It can request the same lock many times", func(t *testing.T) {
		testConstruct.StartLocker(g)
		defer testConstruct.Shutdown(g)

		lockerClient := testConstruct.LockerClient

		lockRequest := v1locker.CreateLockRequest{
			KeyValues: datatypes.KeyValues{
				"key1": datatypes.String("key one"),
				"key2": datatypes.String("key two"),
			},
		}
		data, err := json.Marshal(lockRequest)
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
				g.Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				obtainLockResp := v1locker.CreateLockResponse{}
				respData, err := io.ReadAll(resp.Body)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(json.Unmarshal(respData, &obtainLockResp)).ToNot(HaveOccurred())

				// release the lock
				releaseLock, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", testConstruct.LockerClient.Address()), bytes.NewBuffer((&v1locker.DeleteLockRequest{SessionID: obtainLockResp.SessionID}).ToBytes()))
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

		locks := v1locker.ListLockResponse{}
		g.Expect(json.Unmarshal(data, &locks)).ToNot(HaveOccurred())
		g.Expect(len(locks.Locks)).To(Equal(0))
	})
}
