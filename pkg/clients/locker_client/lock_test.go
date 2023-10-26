package lockerclient

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func getFreePort(g *GomegaWithT) int {
	l, err := net.Listen("tcp", ":0")
	g.Expect(err).ToNot(HaveOccurred())
	l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func setupClientHttp() *http.Client {
	return &http.Client{}
}

func setupServerHttp(serverMux *http.ServeMux) *httptest.Server {
	return httptest.NewServer(serverMux)
}

func setupLock(server *httptest.Server) (*lock, *atomic.Int64, *error, *atomic.Int64, *error, *atomic.Int64, *error) {
	var lockLostCallbackError error
	lostLockCallbackCounter := new(atomic.Int64)
	lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
		lockLostCallbackError = err
		lostLockCallbackCounter.Add(1)
	}

	var heartbeatErrorCallbackError error
	heartbeatErrorCallbackCounter := new(atomic.Int64)
	heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
		heartbeatErrorCallbackError = err
		heartbeatErrorCallbackCounter.Add(1)
	}

	var deleteLockCallbackError error
	deleteLockCallbackCounter := new(atomic.Int64)
	deleteLockCallback := func(keyValues datatypes.KeyValues, err error) {
		deleteLockCallbackError = err
		deleteLockCallbackCounter.Add(1)
	}

	lock := &lock{
		client:                  server.Client(),
		url:                     server.URL,
		lockLostCallback:        lockLostCallback,
		heartbeatErrorCallback:  heartbeatErrorCallback,
		deleteLockErrorCallback: deleteLockCallback,
		keyValues:               datatypes.KeyValues{},
		done:                    make(chan struct{}),
		releaseChan:             make(chan struct{}),
		sessionID:               "someID",
		timeout:                 200 * time.Millisecond,
	}

	return lock, lostLockCallbackCounter, &lockLostCallbackError, heartbeatErrorCallbackCounter, &heartbeatErrorCallbackError, deleteLockCallbackCounter, &deleteLockCallbackError
}

func TestLock_Execute(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It constantly executes heartbeats for a lock", func(t *testing.T) {
		// setup the server
		serverCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			serverCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// create the lock and counters
		lock, lostCounter, _, hertbeatErrorCounter, _, deleteCounter, _ := setupLock(server)

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		g.Eventually(func() int64 { return serverCounter.Load() }).Should(BeNumerically(">=", int64(5)))
		g.Consistently(func() int64 { return lostCounter.Load() }).Should(Equal(int64(0)))
		g.Consistently(func() int64 { return hertbeatErrorCounter.Load() }).Should(Equal(int64(0)))
		g.Consistently(func() int64 { return deleteCounter.Load() }).Should(Equal(int64(0)))
	})

	t.Run("It releases the lock locally after failing to heartbeat for the duration of the configured timeout", func(t *testing.T) {
		// setup the server
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := setupServerHttp(mux)
		server.Close()

		// create the lock and counters
		lock, lostCounter, lostErr, hertbeatErrorCounter, heartbeatErr, deleteCounter, deleteErr := setupLock(server)

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		g.Eventually(func() int64 { return hertbeatErrorCounter.Load() }).Should(BeNumerically(">=", 3)) // with all the timers, this should eventually be 3 or 4
		g.Eventually(func() int64 { return lostCounter.Load() }).Should(Equal(int64(1)))
		g.Consistently(func() int64 { return deleteCounter.Load() }).Should(Equal(int64(0)))
		g.Expect((*lostErr).Error()).To(ContainSubstring("could not heartbeat successfuly since the timeout. Releasing the local lock since remote is unreachable"))
		g.Expect((*heartbeatErr).Error()).To(ContainSubstring("client closed unexpectedly when heartbeating"))
		g.Expect(*deleteErr).ToNot(HaveOccurred())
	})

	t.Run("It stops the heartbeating if the context is canceled", func(t *testing.T) {
		// setup the server
		deleteCounter := new(atomic.Int64)
		serverCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			serverCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/v1/locks/delete", func(w http.ResponseWriter, r *http.Request) {
			deleteCounter.Add(1)
			w.WriteHeader(http.StatusNoContent)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// create the lock and counters
		lock, lostCounter, lostErr, hertbeatErrorCounter, heartbeatErr, deleteCounter, deleteErr := setupLock(server)

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		// wait for requests to start
		g.Eventually(func() int64 { return serverCounter.Load() }).Should(BeNumerically(">=", 3))

		// cancel
		cancel()
		g.Eventually(lock.done).Should(BeClosed())
		g.Eventually(func() int64 { return deleteCounter.Load() }).Should(Equal(int64(1)))

		finalCounter := serverCounter.Load()
		g.Consistently(func() int64 { return serverCounter.Load() }).Should(Equal(finalCounter))

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(deleteCounter.Load()).To(Equal(int64(0)))
		g.Expect(*lostErr).To(BeNil())
		g.Expect(*heartbeatErr).To(BeNil())
		g.Expect(*deleteErr).To(BeNil())
	})

	t.Run("It stops heartbeating and delete the lock if releaseAndStopHeartbeat is called", func(t *testing.T) {
		// setup the server
		deleteApiCounter := new(atomic.Int64)
		serverCounter := new(atomic.Int64)
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
			serverCounter.Add(1)
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/v1/locks/delete", func(w http.ResponseWriter, r *http.Request) {
			deleteApiCounter.Add(1)
			w.WriteHeader(http.StatusNoContent)
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// create the lock and counters
		lock, lostCounter, lostErr, hertbeatErrorCounter, heartbeatErr, deleteCounter, deleteErr := setupLock(server)

		// start processing the lock's heartbeats
		asyncManager := goasync.NewTaskManager(goasync.StrictConfig())
		asyncManager.AddExecuteTask("test lock", lock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = asyncManager.Run(ctx)
		}()

		// wait fr requests to start processing
		g.Eventually(func() int64 { return serverCounter.Load() }).Should(BeNumerically(">=", 3))

		// release the lock
		lock.releaseAndStopHeartbeat()
		g.Eventually(lock.done).Should(BeClosed())
		g.Eventually(func() int64 { return deleteApiCounter.Load() }).Should(Equal(int64(1)))

		finalCounter := serverCounter.Load()
		g.Consistently(func() int64 { return serverCounter.Load() }).Should(Equal(finalCounter))

		g.Expect(lostCounter.Load()).To(Equal(int64(1)))
		g.Expect(hertbeatErrorCounter.Load()).To(Equal(int64(0)))
		g.Expect(deleteCounter.Load()).To(Equal(int64(0)))
		g.Expect(*lostErr).To(BeNil())
		g.Expect(*heartbeatErr).To(BeNil())
		g.Expect(*deleteErr).To(BeNil())
	})
}

func TestLock_heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It calls heartbeatErrorCallbak if the request cannot be made to the server", func(t *testing.T) {
		client := setupClientHttp()

		lostLockCallbackCounter := 0
		lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
			lostLockCallbackCounter++
		}

		var heartbeatErrorCallbackError error
		heartbeatErrorCallbackCounter := 0
		heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
			heartbeatErrorCallbackError = err
			heartbeatErrorCallbackCounter++
		}

		lock := &lock{
			client:                 client,
			url:                    fmt.Sprintf("http://127.0.0.1:%d", getFreePort(g)),
			lockLostCallback:       lockLostCallback,
			heartbeatErrorCallback: heartbeatErrorCallback,
			keyValues:              datatypes.KeyValues{},
			releaseChan:            make(chan struct{}),
			sessionID:              "someID",
			timeout:                time.Second,
		}

		lock.heartbeat()
		g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
		g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("dial tcp 127.0.0.1:"))
		g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("connect: connection refused"))
	})

	t.Run("Context when the response code is http.StatusOK", func(t *testing.T) {
		t.Run("It does not call any callbacks", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackError).To(BeNil())
		})
	})

	t.Run("Context when the response code is http.StatusBadRequest", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("internal error. client unable to read response body"))
		})

		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the heartbeatErrorCallback with the remote error", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				apiErr := &api.Error{Message: "this is the api error"}
				w.WriteHeader(http.StatusBadRequest)
				w.Write(apiErr.ToBytes())
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("this is the api error"))
		})

	})

	t.Run("Context when the response code is http.StatusConflict", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server crashes unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`this isn't json`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("internal error. client unable to read response body"))
		})

		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`this isn't json`))
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the lostLock callback and returns the original server error", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				heartbeatErrors := v1locker.HeartbeatLocksResponse{
					HeartbeatErrors: []v1locker.HeartbeatError{
						{Session: "someID", Error: "the session id does not exist"},
					},
				}

				w.WriteHeader(http.StatusConflict)
				w.Write(heartbeatErrors.ToBytes())
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			var lostLockCallbackErr error
			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackErr = err
				lostLockCallbackCounter++
			}

			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(0))
			g.Expect(lostLockCallbackErr).To(HaveOccurred())
			g.Expect(lostLockCallbackErr.Error()).To(Equal("the session id does not exist"))
		})
	})

	t.Run("Context when the response code is anything else", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`what should i do`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("received an unexpected status code: 500"))
		})
	})
}

func TestLock_release(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It calls heartbeatErrorCallbak if the request cannot be made to the server", func(t *testing.T) {
		client := setupClientHttp()

		lostLockCallbackCounter := 0
		lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
			lostLockCallbackCounter++
		}

		var heartbeatErrorCallbackError error
		heartbeatErrorCallbackCounter := 0
		heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
			heartbeatErrorCallbackError = err
			heartbeatErrorCallbackCounter++
		}

		lock := &lock{
			client:                 client,
			url:                    fmt.Sprintf("http://127.0.0.1:%d", getFreePort(g)),
			lockLostCallback:       lockLostCallback,
			heartbeatErrorCallback: heartbeatErrorCallback,
			keyValues:              datatypes.KeyValues{},
			releaseChan:            make(chan struct{}),
			sessionID:              "someID",
			timeout:                time.Second,
		}

		lock.heartbeat()
		g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
		g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("dial tcp 127.0.0.1:"))
		g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("connect: connection refused"))
	})

	t.Run("Context when the response code is http.StatusOK", func(t *testing.T) {
		t.Run("It does not call any callbacks", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackError).To(BeNil())
		})
	})

	t.Run("Context when the response code is http.StatusBadRequest", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("internal error. client unable to read response body"))
		})

		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`this isn't json`))
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the heartbeatErrorCallback with the remote error", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				apiErr := &api.Error{Message: "this is the api error"}
				w.WriteHeader(http.StatusBadRequest)
				w.Write(apiErr.ToBytes())
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("this is the api error"))
		})

	})

	t.Run("Context when the response code is http.StatusConflict", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server crashes unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`this isn't json`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("internal error. client unable to read response body"))
		})

		t.Run("It calls the heartbeatErrorCallback if the server response cannot be parsed", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`this isn't json`))
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("error paring server response body"))
		})

		t.Run("It calls the lostLock callback and returns the original server error", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				heartbeatErrors := v1locker.HeartbeatLocksResponse{
					HeartbeatErrors: []v1locker.HeartbeatError{
						{Session: "someID", Error: "the session id does not exist"},
					},
				}

				w.WriteHeader(http.StatusConflict)
				w.Write(heartbeatErrors.ToBytes())
			})

			client := setupClientHttp()
			server := setupServerHttp(mux)
			defer server.Close()

			var lostLockCallbackErr error
			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackErr = err
				lostLockCallbackCounter++
			}

			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(0))
			g.Expect(lostLockCallbackErr).To(HaveOccurred())
			g.Expect(lostLockCallbackErr.Error()).To(Equal("the session id does not exist"))
		})
	})

	t.Run("Context when the response code is anything else", func(t *testing.T) {
		t.Run("It calls the heartbeatErrorCallback the remote server response cannot be read unexpectedly", func(t *testing.T) {
			mux := http.NewServeMux()
			client := setupClientHttp()
			server := setupServerHttp(mux)

			mux.HandleFunc("/v1/locks/heartbeat", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`what should i do`))
			})

			lostLockCallbackCounter := 0
			lockLostCallback := func(keyValues datatypes.KeyValues, err error) {
				lostLockCallbackCounter++
			}

			var heartbeatErrorCallbackError error
			heartbeatErrorCallbackCounter := 0
			heartbeatErrorCallback := func(keyValues datatypes.KeyValues, err error) {
				heartbeatErrorCallbackError = err
				heartbeatErrorCallbackCounter++
			}

			lock := &lock{
				client:                 client,
				url:                    server.URL,
				lockLostCallback:       lockLostCallback,
				heartbeatErrorCallback: heartbeatErrorCallback,
				keyValues:              datatypes.KeyValues{},
				releaseChan:            make(chan struct{}),
				sessionID:              "someID",
				timeout:                time.Second,
			}

			lock.heartbeat()
			g.Expect(lostLockCallbackCounter).To(Equal(0))
			g.Expect(heartbeatErrorCallbackCounter).To(Equal(1))
			g.Expect(heartbeatErrorCallbackError).To(HaveOccurred())
			g.Expect(heartbeatErrorCallbackError.Error()).To(ContainSubstring("received an unexpected status code: 500"))
		})
	})
}
