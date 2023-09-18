package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/server/versions/v1limiter/v1limiterfakes"
	"github.com/DanLavine/willow/testhelpers/testclient"
	"github.com/DanLavine/willow/testhelpers/testconfig"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLimiterTCP_ShutdownBehavior(t *testing.T) {
	g := NewGomegaWithT(t)
	limiterConfig := testconfig.Limiter(g)

	setupLogger := func() (*zap.Logger, *observer.ObservedLogs) {
		zapCore, observer := observer.New(zap.DebugLevel)
		logger := zap.New(zapCore)

		return logger, observer
	}

	t.Run("It shuts down if there are no processing clients", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeLimitHandler := v1limiterfakes.NewMockLimitRuleHandler(mockController)

		logger, logObserver := setupLogger()
		server := NewLimiterTCP(logger, limiterConfig, fakeLimitHandler)

		// create a goasync task manager and add the server
		taskManager := goasync.NewTaskManager(goasync.StrictConfig())
		g.Expect(taskManager.AddTask("tcp server", server)).ToNot(HaveOccurred())

		// start running the goasync task mnager
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan []goasync.NamedError)
		go func() {
			errChan <- taskManager.Run(ctx)
		}()

		// ensure that the server is eventually running
		g.Consistently(errChan).ShouldNot(Receive())
		g.Eventually(logObserver.Len).Should(Equal(1))
		g.Expect(logObserver.All()[0].Message).To(Equal("Limiter TCP server running"))

		// stop running the server
		cancel()

		g.Eventually(errChan).Should(Receive(BeNil()))
	})

	t.Run("It waits for any in flight http requests", func(t *testing.T) {
		handlerChan := make(chan struct{})

		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeLimitHandler := v1limiterfakes.NewMockLimitRuleHandler(mockController)
		fakeLimitHandler.EXPECT().Create(gomock.Any(), gomock.Any()).Do(func(w http.ResponseWriter, r *http.Request) {
			handlerChan <- struct{}{}
			handlerChan <- struct{}{}

			w.WriteHeader(http.StatusOK)
		})

		logger, logObserver := setupLogger()
		server := NewLimiterTCP(logger, limiterConfig, fakeLimitHandler)

		// create a goasync task manager and add the server
		taskManager := goasync.NewTaskManager(goasync.StrictConfig())
		g.Expect(taskManager.AddTask("tcp server", server)).ToNot(HaveOccurred())

		// start running the goasync task mnager
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan []goasync.NamedError)
		go func() {
			errChan <- taskManager.Run(ctx)
		}()

		// ensure that the server is eventually running
		g.Consistently(errChan).ShouldNot(Receive())
		g.Eventually(logObserver.Len).Should(Equal(1))
		g.Expect(logObserver.All()[0].Message).To(Equal("Limiter TCP server running"))

		// make a request to the server
		client := testclient.Limiter(g)
		httpRequest, err := http.NewRequest("GET", "https://127.0.0.1:8080/v1/group_rules/create", nil)
		g.Expect(err).ToNot(HaveOccurred())

		respChan := make(chan *http.Response)
		go func() {
			serverResponse, err := client.Do(httpRequest)
			g.Expect(err).ToNot(HaveOccurred())
			respChan <- serverResponse
		}()

		// ensure that the http handler was called
		g.Eventually(handlerChan).Should(Receive(Equal(struct{}{})))

		// stop running the server
		cancel()

		// ensure that the server stays running
		g.Consistently(errChan).ShouldNot(Receive())

		// allow the http handler to stop running
		g.Eventually(handlerChan).Should(Receive(Equal(struct{}{})))

		// the server should now be shut down
		g.Eventually(errChan).Should(Receive(BeNil()))

		// check the server responded properly
		g.Eventually(func() int {
			serverResponse := <-respChan
			return serverResponse.StatusCode
		}).Should(Equal(http.StatusOK))
	})
}
