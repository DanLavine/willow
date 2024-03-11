package lockerclient

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	. "github.com/onsi/gomega"
)

func TestLockerClient_New(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the config is not valid", func(t *testing.T) {
		lockerClient, err := NewLockClient(&clients.Config{})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("client's Config.URL cannot be empty"))
		g.Expect(lockerClient).To(BeNil())
	})
}

func TestLockerClient_ObtainLock(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It does not return an error if the request context is closed", func(t *testing.T) {
		// setup client
		cfg := &clients.Config{URL: "http://127.0.0.1:8080"}

		lockerClient, err := NewLockClient(cfg)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())

		// obtain the lock
		lockContext, lockCancel := context.WithCancel(context.Background())
		lockCancel()

		lock, err := lockerClient.ObtainLock(lockContext, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lock).To(BeNil())
	})

	t.Run("It returns an error if the client cannot make the request to the remote locker service", func(t *testing.T) {
		// setup client
		cfg := &clients.Config{URL: "bad url"}

		lockerClient, err := NewLockClient(cfg)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(lockerClient).ToNot(BeNil())

		// obtain the lock 1st time
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lock, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("unable to make request to locker service"))
		g.Expect(lock).To(BeNil())
	})

	t.Run("Context when the response is http.StatusOK", func(t *testing.T) {
		t.Run("It can obtain a lock for the provided key values", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks", func(w http.ResponseWriter, r *http.Request) {
				createLockResponse := &v1locker.LockCreateResponse{
					SessionID:   "24",
					LockTimeout: time.Second,
				}
				_, err := api.EncodeAndSendHttpResponse(http.Header{}, w, http.StatusOK, createLockResponse)
				g.Expect(err).ToNot(HaveOccurred())
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockClient(cfg)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// obtain the lock
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			lock, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock).ToNot(BeNil())
			defer lock.Release(nil)
		})

		t.Run("It returns the original lock obtained on a 2nd call", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			createCounter := 0
			mux.HandleFunc("/v1/locks", func(w http.ResponseWriter, r *http.Request) {
				createCounter++
				createLockResponse := &v1locker.LockCreateResponse{
					SessionID:   "24",
					LockTimeout: time.Second,
				}
				_, err := api.EncodeAndSendHttpResponse(http.Header{}, w, http.StatusOK, createLockResponse)
				g.Expect(err).ToNot(HaveOccurred())
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockClient(cfg)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// obtain the lock 1st time
			lock1, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock1).ToNot(BeNil())

			// obtain the lock 2nd time
			lock2, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lock2).ToNot(BeNil())

			// ensure server was only called 1 time
			g.Expect(lock1).To(Equal(lock2))
			g.Expect(createCounter).To(Equal(1))
			defer lock1.Release(nil)
		})

		t.Run("It returns an error if the response body cannot be read", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("content-length", "5")
				w.Header().Add("Content-Type", "application/json")

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`this has a bad body`))
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockClient(cfg)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// try to obtain the lock
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			lock, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to read http response body"))
			g.Expect(lock).To(BeNil())
		})

		t.Run("It returns an error if the response body cannot be parsed", func(t *testing.T) {
			// setup server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/locks", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`this has a bad json`))
			})
			server := setupServerHttp(mux)
			defer server.Close()

			// setup client
			cfg := &clients.Config{URL: server.URL}

			lockerClient, err := NewLockClient(cfg)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(lockerClient).ToNot(BeNil())

			// try to obtain the lock
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			lock, err := lockerClient.ObtainLock(ctx, &v1locker.LockCreateRequest{KeyValues: datatypes.KeyValues{"one": datatypes.Float32(3.4)}}, nil, nil)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to decode response"))
			g.Expect(lock).To(BeNil())
		})
	})
}
