package lockerclient

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestLockerClient_New(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := &clients.Config{
		URL: "doesn't matter here",
	}

	t.Run("It returns an error if the context is nil, Background, TODO", func(t *testing.T) {
		lockerClient, err := NewLockerClient(nil, cfg, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("cannot use provided context. The context must be canceled by the caller to cleanup async resource management"))
		g.Expect(lockerClient).To(BeNil())

		lockerClient, err = NewLockerClient(context.Background(), cfg, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("cannot use provided context. The context must be canceled by the caller to cleanup async resource management"))
		g.Expect(lockerClient).To(BeNil())

		lockerClient, err = NewLockerClient(context.TODO(), cfg, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("cannot use provided context. The context must be canceled by the caller to cleanup async resource management"))
		g.Expect(lockerClient).To(BeNil())
	})

	t.Run("It returns an error if the config is not valid", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lockerClient, err := NewLockerClient(ctx, &clients.Config{}, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("client's Config.URL cannot be empty"))
		g.Expect(lockerClient).To(BeNil())
	})

	t.Run("It accepts a nil heartbeatErrorCallback", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lockerClient, err := NewLockerClient(ctx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())
		g.Consistently(lockerClient.Done()).ShouldNot(BeClosed())
	})

	t.Run("It closes done when the Context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		lockerClient, err := NewLockerClient(ctx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())
		g.Eventually(lockerClient.Done()).Should(BeClosed())
	})
}

func TestLockerClient_ObtainLock(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the locker client has been closed", func(t *testing.T) {
		// setup client
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		cfg := &clients.Config{URL: "http://127.0.0.1:8080"}

		lockerClient, err := NewLockerClient(ctx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())
		g.Eventually(lockerClient.Done()).Should(BeClosed())

		// try to obtain the lock 1st time
		lock, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("locker client has already been canceled and won't process heartbeats. Refusing to obtain the lock"))
		g.Expect(lock).To(BeNil())
	})

	t.Run("It does not return an error if the request context is closed", func(t *testing.T) {
		// setup client
		clientCtx, clienetCancel := context.WithCancel(context.Background())
		defer clienetCancel()

		cfg := &clients.Config{URL: "http://127.0.0.1:8080"}

		lockerClient, err := NewLockerClient(clientCtx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())

		// obtain the lock 1st time
		lockContext, lockCancel := context.WithCancel(context.Background())
		lockCancel()

		lock, err := lockerClient.ObtainLock(lockContext, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).To(BeNil())
	})

	t.Run("It returns an error if the client cannot make the request to the remote locker service", func(t *testing.T) {
		// setup client
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cfg := &clients.Config{URL: "bad url"}

		lockerClient, err := NewLockerClient(ctx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())

		// obtain the lock 1st time
		lock, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("unable to make request to locker service"))
		g.Expect(lock).To(BeNil())
	})

	t.Run("It returns an error if the locker client is closed while making a request", func(t *testing.T) {
		// setup server
		mux := http.NewServeMux()
		serverReceivedRequst := make(chan struct{})
		mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
			close(serverReceivedRequst)
			time.Sleep(time.Second)

			createLockResponse := &v1locker.CreateLockResponse{
				SessionID: "24",
				Timeout:   time.Second,
			}
			w.WriteHeader(http.StatusCreated)
			w.Write(createLockResponse.ToBytes())
		})
		server := setupServerHttp(mux)
		defer server.Close()

		// setup client
		clientCtx, clientCancel := context.WithCancel(context.Background())
		cfg := &clients.Config{URL: server.URL}

		lockerClient, err := NewLockerClient(clientCtx, cfg, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())

		// try to obtain the lock
		lockContext, lockCancel := context.WithCancel(context.Background())
		defer lockCancel()

		var lock Lock
		doneObtaineLock := make(chan struct{})
		go func() {
			defer close(doneObtaineLock)
			lock, err = lockerClient.ObtainLock(lockContext, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
		}()

		// close the client
		g.Eventually(serverReceivedRequst).Should(BeClosed())
		clientCancel()
		g.Eventually(doneObtaineLock).Should(BeClosed())

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("locker client has been canceled and won't process heartbeats. Refusing to obtain the lock"))
		g.Expect(lock).To(BeNil())
	})

	t.Run("Context when the response is http.StatusCreated", func(t *testing.T) {
		t.Run("It can obtain a lock for the provided key values", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
				createLockResponse := &v1locker.CreateLockResponse{
					SessionID: "24",
					Timeout:   time.Second,
				}
				w.WriteHeader(http.StatusCreated)
				w.Write(createLockResponse.ToBytes())
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockerClient(ctx, cfg, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// obtain the lock
			lock, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock).ToNot(BeNil())
		})

		t.Run("It returns the originally obtained lock on a 2nd call", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			createCounter := 0
			mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
				createCounter++
				createLockResponse := &v1locker.CreateLockResponse{
					SessionID: "24",
					Timeout:   time.Second,
				}
				w.WriteHeader(http.StatusCreated)
				w.Write(createLockResponse.ToBytes())
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockerClient(ctx, cfg, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// obtain the lock 1st time
			lock1, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock1).ToNot(BeNil())

			// obtain the lock 2nd time
			lock2, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock2).ToNot(BeNil())

			// ensure server was only called 1 time
			g.Expect(lock1).To(Equal(lock2))
			g.Expect(createCounter).To(Equal(1))
		})

		t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`this has a bad header`))
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockerClient(ctx, cfg, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// try to obtain the lock
			lock, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read response body"))
			g.Expect(lock).To(BeNil())
		})

		t.Run("It returns an error if the response body cannot be parsed", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`this has a bad json`))
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockerClient(ctx, cfg, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// try to obtain the lock
			lock, err := lockerClient.ObtainLock(ctx, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to parse server response"))
			g.Expect(lock).To(BeNil())
		})

		t.Run("It returns an error if the client is closed as soon as a lock is obtained", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locker/create", func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)

				createLockResponse := &v1locker.CreateLockResponse{
					SessionID: "24",
					Timeout:   time.Second,
				}
				w.WriteHeader(http.StatusCreated)
				w.Write(createLockResponse.ToBytes())
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			cfg := &clients.Config{URL: server.URL}

			clientCtx, clientCancel := context.WithCancel(context.Background())
			defer clientCancel()

			lockerClient, err := NewLockerClient(clientCtx, cfg, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// for using a custom async manager to similate this specific race condiition
			lockerClient.(*lockerclient).asyncManager = goasync.NewTaskManager(goasync.RelaxedConfig())
			stoppedContext, cancelNow := context.WithCancel(context.Background())
			cancelNow()
			lockerClient.(*lockerclient).asyncManager.Run(stoppedContext)

			// try to obtain the lock
			lockContext, lockCancel := context.WithCancel(context.Background())
			defer lockCancel()

			var lock Lock
			doneObtaineLock := make(chan struct{})
			go func() {
				defer close(doneObtaineLock)
				lock, err = lockerClient.ObtainLock(lockContext, datatypes.KeyValues{"one": datatypes.Float32(3.4)}, time.Second)
			}()

			// wait for the lock obtain to fail
			g.Eventually(doneObtaineLock).Should(BeClosed())

			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("client has been closed, will not obtain lock"))
			g.Expect(lock).To(BeNil())
		})

	})
}
