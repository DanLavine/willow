package lockerclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	. "github.com/onsi/gomega"
)

func ErrorToBytes(e *errors.Error) []byte {
	data, _ := json.Marshal(e)
	return data
}

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

	lock := newLock("someID", 200*time.Millisecond, server.Client(), server.URL, heartbeatErrorCallback, lockLostCallback)

	return lock, lostLockCallbackCounter, heartbeatErrorCallbackCounter, &heartbeatErrorCallbackError
}

func TestLock_Execute(t *testing.T) {
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

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = asyncManager.Run(ctx)
		}()

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

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.StrictConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		done := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			defer close(done)
			_ = asyncManager.Run(ctx)
		}()

		// ensure that the lock has been lost
		g.Eventually(func() int64 { return lostCounter.Load() }).Should(Equal(int64(1)))
		g.Eventually(done).Should(BeClosed())
		g.Expect(lock.Done()).To(BeClosed())

		// check the errors
		for i := int64(0); i < hertbeatErrorCounter.Load(); i++ {
			if i == hertbeatErrorCounter.Load()-1 {
				g.Expect((*heartbeatErr)[i].Error()).To(ContainSubstring("could not heartbeat successfuly since the timeout. Releasing the local Lock since remote is unreachable"))
			} else {
				g.Expect((*heartbeatErr)[i].Error()).To(ContainSubstring("client closed unexpectedly when heartbeating"))
			}
		}
	})

	t.Run("It stops the heartbeating if the context is canceled", func(t *testing.T) {
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

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		// wait for requests to start
		g.Eventually(func() int64 { return heartbeatCounter.Load() }).Should(BeNumerically(">=", 3))

		// cancel
		cancel()
		g.Eventually(lock.Done()).Should(BeClosed())
		g.Eventually(func() int64 { return deleteCounter.Load() }).Should(Equal(int64(1))) // ensure delete api was called

		finalCounter := heartbeatCounter.Load()
		g.Consistently(func() int64 { return heartbeatCounter.Load() }).Should(Equal(finalCounter))

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(*heartbeatErrs).To(BeNil())
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

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		// wait fr requests to start processing
		g.Eventually(func() int64 { return heartbeatCounter.Load() }).Should(BeNumerically(">=", 3))

		// release the lock
		lock.Release()
		g.Eventually(lock.Done()).Should(BeClosed())
		g.Expect(deleteCounter.Load()).To(Equal(int64(1)))

		finalCounter := heartbeatCounter.Load()
		g.Consistently(func() int64 { return heartbeatCounter.Load() }).Should(Equal(finalCounter))

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(*heartbeatErrs).To(BeNil())
	})
}

func TestLock_heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It calls heartbeatErrorCallbak if the request cannot be made to the server", func(t *testing.T) {
		// create the lock and counters
		mux := http.NewServeMux()
		server := setupServerHttp(mux)
		server.Close()
		lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

		lock.heartbeat()
		g.Expect(lostCounter.Load()).To(Equal(int64(0)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
		g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("connect: connection refused"))
	})

	t.Run("Context when the response code is http.StatusOK", func(t *testing.T) {
		t.Run("It does not call any callbacks", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})

	t.Run("Context when the response code is http.StatusBadRequest or http.StatusConflict", func(t *testing.T) {
		t.Run("It calls heartbeatErrorCallback if the response body canot be read", func(t *testing.T) {
			counter := 0
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				switch counter {
				case 0:
					w.Header().Add("content-length", "5")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`this isn't json`))
					counter++
				default:
					w.Header().Add("content-length", "5")
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(`this isn't json`))
				}
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// bad request
			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("internal error. client unable to read response body"))

			// conflict
			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(2)))
			g.Expect((*heartbeatErrs)[1].Error()).To(ContainSubstring("internal error. client unable to read response body"))
		})
	})

	t.Run("Context when the response code is http.StatusBadRequest", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the heartbeatErrorCallback with the remote error", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				apiErr := &errors.Error{Message: "this is the api error"}
				w.WriteHeader(http.StatusBadRequest)
				w.Write(ErrorToBytes(apiErr))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("this is the api error"))
		})
	})

	t.Run("Context when the response code is http.StatusConflict", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`this isn't json`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the lostLock callback and returns the original server error when the response is parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				heartbeatError := &v1locker.HeartbeatError{
					Session: "someID",
					Error:   "the session id does not exist",
				}

				w.WriteHeader(http.StatusConflict)
				w.Write(heartbeatError.ToBytes())
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("the session id does not exist"))
		})
	})

	t.Run("Context when the response code is anything else", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`what should i do`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			lock.heartbeat()
			g.Expect(lostCounter.Load()).To(Equal(int64(0)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(1)))
			g.Expect((*heartbeatErrs)[0].Error()).To(ContainSubstring("received an unexpected status code: 500"))
		})
	})
}

func TestLock_release(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context when the response code is http.StatusNoContent", func(t *testing.T) {
		t.Run("It stops heartbeating and closes the Done channel", func(t *testing.T) {
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
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// setup async task manager
			asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			asyncManager.AddExecuteTask("test lock", lock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = asyncManager.Run(ctx)
			}()

			// release the lock
			err := lock.Release()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})

	t.Run("Context when the response code is http.StatusBadRequest", func(t *testing.T) {
		t.Run("It reports the api error on a bad request", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
				apiErr := &errors.Error{Message: "bad request budy"}
				w.WriteHeader(http.StatusBadRequest)
				w.Write(ErrorToBytes(apiErr))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// setup async task manager
			asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			asyncManager.AddExecuteTask("test lock", lock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = asyncManager.Run(ctx)
			}()

			// release the lock
			err := lock.Release()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("bad request budy"))
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})

		t.Run("It returns an error if the response cannot be read", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// setup async task manager
			asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			asyncManager.AddExecuteTask("test lock", lock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = asyncManager.Run(ctx)
			}()

			// release the lock
			err := lock.Release()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("internal error. client unable to read response body"))
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})

		t.Run("It returns an error if the response body cannot be parsed", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// setup async task manager
			asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			asyncManager.AddExecuteTask("test lock", lock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = asyncManager.Run(ctx)
			}()

			// release the lock
			err := lock.Release()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("error paring server response body"))
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})

	t.Run("Context when the response code is anything else", func(t *testing.T) {
		t.Run("It stops heartbeating and closes the Done channel", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/someID/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			mux.HandleFunc("/v1/locks/someID", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			server := setupServerHttp(mux)
			defer server.Close()
			lock, lostCounter, hertbeatErrorCounter, heartbeatErrs := setupLock(server)

			// setup async task manager
			asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
			asyncManager.AddExecuteTask("test lock", lock)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = asyncManager.Run(ctx)
			}()

			// release the lock
			err := lock.Release()
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("unexpected response code from the remote locker service. Need to wait for the lock to expire "))
			g.Expect(lock.Done()).To(BeClosed())

			g.Expect(lostCounter.Load()).To(Equal(int64(1)))
			g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
			g.Expect(*heartbeatErrs).To(BeNil())
		})
	})
}
