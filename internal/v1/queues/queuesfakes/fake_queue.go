// Code generated by counterfeiter. DO NOT EDIT.
package queuesfakes

import (
	"sync"

	"github.com/DanLavine/willow/internal/v1/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type FakeQueue struct {
	EnqueueStub        func(*zap.Logger, *v1.EnqueueItemRequest) *v1.Error
	enqueueMutex       sync.RWMutex
	enqueueArgsForCall []struct {
		arg1 *zap.Logger
		arg2 *v1.EnqueueItemRequest
	}
	enqueueReturns struct {
		result1 *v1.Error
	}
	enqueueReturnsOnCall map[int]struct {
		result1 *v1.Error
	}
	MetricsStub        func() *v1.QueueMetricsResponse
	metricsMutex       sync.RWMutex
	metricsArgsForCall []struct {
	}
	metricsReturns struct {
		result1 *v1.QueueMetricsResponse
	}
	metricsReturnsOnCall map[int]struct {
		result1 *v1.QueueMetricsResponse
	}
	ReadersStub        func(*v1.MatchQuery) []<-chan func() *v1.DequeueItemResponse
	readersMutex       sync.RWMutex
	readersArgsForCall []struct {
		arg1 *v1.MatchQuery
	}
	readersReturns struct {
		result1 []<-chan func() *v1.DequeueItemResponse
	}
	readersReturnsOnCall map[int]struct {
		result1 []<-chan func() *v1.DequeueItemResponse
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeQueue) Enqueue(arg1 *zap.Logger, arg2 *v1.EnqueueItemRequest) *v1.Error {
	fake.enqueueMutex.Lock()
	ret, specificReturn := fake.enqueueReturnsOnCall[len(fake.enqueueArgsForCall)]
	fake.enqueueArgsForCall = append(fake.enqueueArgsForCall, struct {
		arg1 *zap.Logger
		arg2 *v1.EnqueueItemRequest
	}{arg1, arg2})
	stub := fake.EnqueueStub
	fakeReturns := fake.enqueueReturns
	fake.recordInvocation("Enqueue", []interface{}{arg1, arg2})
	fake.enqueueMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeQueue) EnqueueCallCount() int {
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	return len(fake.enqueueArgsForCall)
}

func (fake *FakeQueue) EnqueueCalls(stub func(*zap.Logger, *v1.EnqueueItemRequest) *v1.Error) {
	fake.enqueueMutex.Lock()
	defer fake.enqueueMutex.Unlock()
	fake.EnqueueStub = stub
}

func (fake *FakeQueue) EnqueueArgsForCall(i int) (*zap.Logger, *v1.EnqueueItemRequest) {
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	argsForCall := fake.enqueueArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeQueue) EnqueueReturns(result1 *v1.Error) {
	fake.enqueueMutex.Lock()
	defer fake.enqueueMutex.Unlock()
	fake.EnqueueStub = nil
	fake.enqueueReturns = struct {
		result1 *v1.Error
	}{result1}
}

func (fake *FakeQueue) EnqueueReturnsOnCall(i int, result1 *v1.Error) {
	fake.enqueueMutex.Lock()
	defer fake.enqueueMutex.Unlock()
	fake.EnqueueStub = nil
	if fake.enqueueReturnsOnCall == nil {
		fake.enqueueReturnsOnCall = make(map[int]struct {
			result1 *v1.Error
		})
	}
	fake.enqueueReturnsOnCall[i] = struct {
		result1 *v1.Error
	}{result1}
}

func (fake *FakeQueue) Metrics() *v1.QueueMetricsResponse {
	fake.metricsMutex.Lock()
	ret, specificReturn := fake.metricsReturnsOnCall[len(fake.metricsArgsForCall)]
	fake.metricsArgsForCall = append(fake.metricsArgsForCall, struct {
	}{})
	stub := fake.MetricsStub
	fakeReturns := fake.metricsReturns
	fake.recordInvocation("Metrics", []interface{}{})
	fake.metricsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeQueue) MetricsCallCount() int {
	fake.metricsMutex.RLock()
	defer fake.metricsMutex.RUnlock()
	return len(fake.metricsArgsForCall)
}

func (fake *FakeQueue) MetricsCalls(stub func() *v1.QueueMetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = stub
}

func (fake *FakeQueue) MetricsReturns(result1 *v1.QueueMetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = nil
	fake.metricsReturns = struct {
		result1 *v1.QueueMetricsResponse
	}{result1}
}

func (fake *FakeQueue) MetricsReturnsOnCall(i int, result1 *v1.QueueMetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = nil
	if fake.metricsReturnsOnCall == nil {
		fake.metricsReturnsOnCall = make(map[int]struct {
			result1 *v1.QueueMetricsResponse
		})
	}
	fake.metricsReturnsOnCall[i] = struct {
		result1 *v1.QueueMetricsResponse
	}{result1}
}

func (fake *FakeQueue) Readers(arg1 *v1.MatchQuery) []<-chan func() *v1.DequeueItemResponse {
	fake.readersMutex.Lock()
	ret, specificReturn := fake.readersReturnsOnCall[len(fake.readersArgsForCall)]
	fake.readersArgsForCall = append(fake.readersArgsForCall, struct {
		arg1 *v1.MatchQuery
	}{arg1})
	stub := fake.ReadersStub
	fakeReturns := fake.readersReturns
	fake.recordInvocation("Readers", []interface{}{arg1})
	fake.readersMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeQueue) ReadersCallCount() int {
	fake.readersMutex.RLock()
	defer fake.readersMutex.RUnlock()
	return len(fake.readersArgsForCall)
}

func (fake *FakeQueue) ReadersCalls(stub func(*v1.MatchQuery) []<-chan func() *v1.DequeueItemResponse) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = stub
}

func (fake *FakeQueue) ReadersArgsForCall(i int) *v1.MatchQuery {
	fake.readersMutex.RLock()
	defer fake.readersMutex.RUnlock()
	argsForCall := fake.readersArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeQueue) ReadersReturns(result1 []<-chan func() *v1.DequeueItemResponse) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = nil
	fake.readersReturns = struct {
		result1 []<-chan func() *v1.DequeueItemResponse
	}{result1}
}

func (fake *FakeQueue) ReadersReturnsOnCall(i int, result1 []<-chan func() *v1.DequeueItemResponse) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = nil
	if fake.readersReturnsOnCall == nil {
		fake.readersReturnsOnCall = make(map[int]struct {
			result1 []<-chan func() *v1.DequeueItemResponse
		})
	}
	fake.readersReturnsOnCall[i] = struct {
		result1 []<-chan func() *v1.DequeueItemResponse
	}{result1}
}

func (fake *FakeQueue) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	fake.metricsMutex.RLock()
	defer fake.metricsMutex.RUnlock()
	fake.readersMutex.RLock()
	defer fake.readersMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeQueue) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ queues.Queue = new(FakeQueue)