package lockerclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	. "github.com/onsi/gomega"
)

func setupServerHttp(serverMux *http.ServeMux) *httptest.Server {
	return httptest.NewServer(serverMux)
}

func setupLock(server *httptest.Server) (*lock, *atomic.Int64, *atomic.Int64, *[]error) {
	var heartbeatErrorCallbackError []error
	heartbeatErrorCallbackCounter := new(atomic.Int64)
	heartbeatErrorCallback := func(err error) {
		heartbeatErrorCallbackError = append(heartbeatErrorCallbackError, err)
		heartbeatErrorCallbackCounter.Add(1)
	}

	lostLockCallbackCounter := new(atomic.Int64)
	lockLostCallback := func() {
		lostLockCallbackCounter.Add(1)
	}

	cfg := &clients.Config{URL: server.URL}
	client, err := clients.NewHTTPClient(cfg)
	if err != nil {
		panic(err)
	}

	lock := newLock("someID", 200*time.Millisecond, server.URL, client, api.ContentTypeJSON, heartbeatErrorCallback, lockLostCallback)

	return lock, lostLockCallbackCounter, heartbeatErrorCallbackCounter, &heartbeatErrorCallbackError
}

func TestLock_heartbeat_operations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It constantly executes heartbeats for a lock", func(t *testing.T) {
		// setup the server
		serverCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			serverCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// create the lock and counters
		lock, lostCounter, hertbeatErrorCounter, _ := setupLock(server)
		defer lock.Release()

		g.Consistently(func() int64 { return lostCounter.Load() }).Should(Equal(int64(0)))
		g.Consistently(func() int64 { return hertbeatErrorCounter.Load() }).Should(Equal(int64(0)))
		g.Eventually(func() int64 { return serverCounter.Load() }).Should(BeNumerically(">=", int64(5)))
	})

	t.Run("It releases the lock locally after failing to heartbeat for the duration of the configured timeout", func(t *testing.T) {
		// setup the server and shut it down immediately
		mux := http.NewServeMux()
		server := setupServerHttp(mux)
		server.Close()

		// create the lock and counters
		lock, lostCounter, hertbeatErrorCounter, heartbeatErr := setupLock(server)

		// ensure that the lock has been lost
		g.Eventually(func() int64 { return lostCounter.Load() }).Should(Equal(int64(1)))
		g.Expect(lock.Done()).To(BeClosed())

		// check the errors
		for i := int64(0); i < hertbeatErrorCounter.Load(); i++ {
			if i == hertbeatErrorCounter.Load()-1 {
				fmt.Println(heartbeatErr)
				g.Expect((*heartbeatErr)[i].Error()).To(ContainSubstring("could not heartbeat successfuly since the timeout. Releasing the local Lock since remote is unreachable"))
			} else {
				g.Expect((*heartbeatErr)[i].Error()).To(ContainSubstring("failed to heartbeat"))
			}
		}
	})

	t.Run("It stops heartbeating and delete the lock if Release is called", func(t *testing.T) {
		// setup the server
		heartbeatCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			heartbeatCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		deleteCounter := new(atomic.Int64)
		mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
			deleteCounter.Add(1)
			w.WriteHeader(http.StatusNoContent)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// create the lock and counters
		lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

		// wait for heartbeat requests to start processing
		g.Eventually(func() int64 { return heartbeatCounter.Load() }).Should(BeNumerically(">=", 3))

		// release the lock
		err := lock.Release()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock.Done()).To(BeClosed())
		g.Expect(deleteCounter.Load()).To(Equal(int64(1)))

		finalCounter := heartbeatCounter.Load()
		g.Consistently(func() int64 { return heartbeatCounter.Load() }).Should(Equal(finalCounter))

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(*heartbeatErrs).To(BeNil())
	})

	t.Run("It calls heartbeatErrorCallbak if the request cannot be made to the server", func(t *testing.T) {
		// create the lock and counters
		mux := http.NewServeMux()
		server := setupServerHttp(mux)
		server.Close()
		lock, _, hertbeatErrorCounter, heartbeatErrs := setupLock(server)
		defer lock.Release()

		g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
		g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("connect: connection refused"))
	})

	t.Run("Context when the response code is http.StatusOK", func(t *testing.T) {
		t.Run("It does not call any callbacks", func(t *testing.T) {
			heartbeatCounter := new(atomic.Int64)
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				heartbeatCounter.Add(1)
				w.WriteHeader(http.StatusOK)
			})
			server := setupServerHttp(mux)
			defer server.Close()

			lock, _, _, heartbeatErrs := setupLock(server)
			defer lock.Release()

			g.Eventually(heartbeatCounter.Load).Should(BeNumerically(">=", int64(1)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})

	t.Run("Context when the response code is http.StatusGone", func(t *testing.T) {
		t.Run("Context when the response body cannot be read", func(t *testing.T) {
			t.Run("It calls heartbeatErrorCallback if the response body cannot be read", func(t *testing.T) {
				mux := http.NewServeMux()
				mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("content-length", "5")
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusGone)
					w.Write([]byte(`this isn't json`))
				})
				server := setupServerHttp(mux)
				defer server.Close()

				_, _, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

				g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
				g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("failed to read http response body: unexpected EOF"))
			})
		})

		t.Run("Context when the response body cannot be parsed", func(t *testing.T) {
			t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
				mux := http.NewServeMux()
				mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusGone)
					w.Write([]byte(`this isn't json`))
				})
				server := setupServerHttp(mux)
				defer server.Close()

				lock, _, hertbeatErrorCounter, heartbeatErrs := setupLock(server)
				defer lock.Release()

				g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
				g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("failed to decode api error response"))
			})
		})

		t.Run("Context when then remote error can be parsed", func(t *testing.T) {
			t.Run("It calls the heartbeatErrorCallback with the remote error", func(t *testing.T) {
				mux := http.NewServeMux()
				mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
					apiErr := &errors.Error{Message: "this is the api error"}
					_, err := api.EncodeAndSendHttpResponse(http.Header{}, w, http.StatusGone, apiErr)
					g.Expect(err).ToNot(HaveOccurred())
				})
				server := setupServerHttp(mux)
				defer server.Close()

				lock, _, hertbeatErrorCounter, heartbeatErrs := setupLock(server)
				defer lock.Release()

				g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
				g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("this is the api error"))
			})
		})

		t.Run("It releases the lock", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`this isn't json`))
			})
			server := setupServerHttp(mux)
			defer server.Close()

			lock, _, hertbeatErrorCounter, _ := setupLock(server)
			defer lock.Release()

			// gone
			g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
			g.Eventually(lock.Done()).Should(BeClosed())
		})
	})

	t.Run("Context when the response code is anything else", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`what should i do`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, _, hertbeatErrorCounter, heartbeatErrs := setupLock(server)
			defer lock.Release()

			g.Eventually(hertbeatErrorCounter.Load).Should(BeNumerically(">=", int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("received an unexpected status code: 500"))
		})
	})
}

func TestLock_Release(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It stops the heartbeat operation if it is processing", func(t *testing.T) {
		// setup server
		heartbeatCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			heartbeatCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

		// release the lock
		err := lock.Release()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock.Done()).To(BeClosed())
		currentHeartbeats := heartbeatCounter.Load()

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(*heartbeatErrs).To(BeNil())
		g.Consistently(heartbeatCounter.Load()).Should(Equal(currentHeartbeats))
	})

	t.Run("Context when the response code is anything other than http.StatusGone", func(t *testing.T) {
		t.Run("It stops heartbeating and closes the Done channel", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// release the lock
			err := lock.Release()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("unexpected response code from the remote locker service '500'. Need to wait for the lock to expire "))
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})

	t.Run("It reurns an error if called multiple times", func(t *testing.T) {
		// setup server
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		lock, _, _, _ := setupLock(server)

		// release the lock
		err := lock.Release()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock.Done()).To(BeClosed())

		err = lock.Release()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("already released the lock"))
	})

	t.Run("It returns an error if called after the heartbeat operation has stopped", func(t *testing.T) {
		// setup server
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusGone)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		lock, _, _, _ := setupLock(server)

		// wait for lock to be released
		g.Eventually(lock.Done()).Should(BeClosed())

		// release the lock
		err := lock.Release()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("already released the lock"))

	})
}
