// Code generated by counterfeiter. DO NOT EDIT.
package queuesfakes

import (
	"context"
	"sync"

	"github.com/DanLavine/willow/internal/brokers/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type FakeManagedQueue struct {
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
	ExecuteStub        func(context.Context) error
	executeMutex       sync.RWMutex
	executeArgsForCall []struct {
		arg1 context.Context
	}
	executeReturns struct {
		result1 error
	}
	executeReturnsOnCall map[int]struct {
		result1 error
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
	ReadersStub        func(*zap.Logger, *v1.ReaderSelect) ([]<-chan func() *v1.DequeueItemResponse, *v1.Error)
	readersMutex       sync.RWMutex
	readersArgsForCall []struct {
		arg1 *zap.Logger
		arg2 *v1.ReaderSelect
	}
	readersReturns struct {
		result1 []<-chan func() *v1.DequeueItemResponse
		result2 *v1.Error
	}
	readersReturnsOnCall map[int]struct {
		result1 []<-chan func() *v1.DequeueItemResponse
		result2 *v1.Error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeManagedQueue) Enqueue(arg1 *zap.Logger, arg2 *v1.EnqueueItemRequest) *v1.Error {
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

func (fake *FakeManagedQueue) EnqueueCallCount() int {
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	return len(fake.enqueueArgsForCall)
}

func (fake *FakeManagedQueue) EnqueueCalls(stub func(*zap.Logger, *v1.EnqueueItemRequest) *v1.Error) {
	fake.enqueueMutex.Lock()
	defer fake.enqueueMutex.Unlock()
	fake.EnqueueStub = stub
}

func (fake *FakeManagedQueue) EnqueueArgsForCall(i int) (*zap.Logger, *v1.EnqueueItemRequest) {
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	argsForCall := fake.enqueueArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeManagedQueue) EnqueueReturns(result1 *v1.Error) {
	fake.enqueueMutex.Lock()
	defer fake.enqueueMutex.Unlock()
	fake.EnqueueStub = nil
	fake.enqueueReturns = struct {
		result1 *v1.Error
	}{result1}
}

func (fake *FakeManagedQueue) EnqueueReturnsOnCall(i int, result1 *v1.Error) {
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

func (fake *FakeManagedQueue) Execute(arg1 context.Context) error {
	fake.executeMutex.Lock()
	ret, specificReturn := fake.executeReturnsOnCall[len(fake.executeArgsForCall)]
	fake.executeArgsForCall = append(fake.executeArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.ExecuteStub
	fakeReturns := fake.executeReturns
	fake.recordInvocation("Execute", []interface{}{arg1})
	fake.executeMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeManagedQueue) ExecuteCallCount() int {
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
	return len(fake.executeArgsForCall)
}

func (fake *FakeManagedQueue) ExecuteCalls(stub func(context.Context) error) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = stub
}

func (fake *FakeManagedQueue) ExecuteArgsForCall(i int) context.Context {
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
	argsForCall := fake.executeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeManagedQueue) ExecuteReturns(result1 error) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = nil
	fake.executeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManagedQueue) ExecuteReturnsOnCall(i int, result1 error) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = nil
	if fake.executeReturnsOnCall == nil {
		fake.executeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.executeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManagedQueue) Metrics() *v1.QueueMetricsResponse {
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

func (fake *FakeManagedQueue) MetricsCallCount() int {
	fake.metricsMutex.RLock()
	defer fake.metricsMutex.RUnlock()
	return len(fake.metricsArgsForCall)
}

func (fake *FakeManagedQueue) MetricsCalls(stub func() *v1.QueueMetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = stub
}

func (fake *FakeManagedQueue) MetricsReturns(result1 *v1.QueueMetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = nil
	fake.metricsReturns = struct {
		result1 *v1.QueueMetricsResponse
	}{result1}
}

func (fake *FakeManagedQueue) MetricsReturnsOnCall(i int, result1 *v1.QueueMetricsResponse) {
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

func (fake *FakeManagedQueue) Readers(arg1 *zap.Logger, arg2 *v1.ReaderSelect) ([]<-chan func() *v1.DequeueItemResponse, *v1.Error) {
	fake.readersMutex.Lock()
	ret, specificReturn := fake.readersReturnsOnCall[len(fake.readersArgsForCall)]
	fake.readersArgsForCall = append(fake.readersArgsForCall, struct {
		arg1 *zap.Logger
		arg2 *v1.ReaderSelect
	}{arg1, arg2})
	stub := fake.ReadersStub
	fakeReturns := fake.readersReturns
	fake.recordInvocation("Readers", []interface{}{arg1, arg2})
	fake.readersMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeManagedQueue) ReadersCallCount() int {
	fake.readersMutex.RLock()
	defer fake.readersMutex.RUnlock()
	return len(fake.readersArgsForCall)
}

func (fake *FakeManagedQueue) ReadersCalls(stub func(*zap.Logger, *v1.ReaderSelect) ([]<-chan func() *v1.DequeueItemResponse, *v1.Error)) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = stub
}

func (fake *FakeManagedQueue) ReadersArgsForCall(i int) (*zap.Logger, *v1.ReaderSelect) {
	fake.readersMutex.RLock()
	defer fake.readersMutex.RUnlock()
	argsForCall := fake.readersArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeManagedQueue) ReadersReturns(result1 []<-chan func() *v1.DequeueItemResponse, result2 *v1.Error) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = nil
	fake.readersReturns = struct {
		result1 []<-chan func() *v1.DequeueItemResponse
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeManagedQueue) ReadersReturnsOnCall(i int, result1 []<-chan func() *v1.DequeueItemResponse, result2 *v1.Error) {
	fake.readersMutex.Lock()
	defer fake.readersMutex.Unlock()
	fake.ReadersStub = nil
	if fake.readersReturnsOnCall == nil {
		fake.readersReturnsOnCall = make(map[int]struct {
			result1 []<-chan func() *v1.DequeueItemResponse
			result2 *v1.Error
		})
	}
	fake.readersReturnsOnCall[i] = struct {
		result1 []<-chan func() *v1.DequeueItemResponse
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeManagedQueue) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.enqueueMutex.RLock()
	defer fake.enqueueMutex.RUnlock()
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
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

func (fake *FakeManagedQueue) recordInvocation(key string, args []interface{}) {
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

var _ queues.ManagedQueue = new(FakeManagedQueue)
