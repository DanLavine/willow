// Code generated by counterfeiter. DO NOT EDIT.
package queuesfakes

import (
	"sync"

	"github.com/DanLavine/willow/internal/v1/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type FakeQueueManager struct {
	CreateStub        func(*zap.Logger, *v1.Create) *v1.Error
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		arg1 *zap.Logger
		arg2 *v1.Create
	}
	createReturns struct {
		result1 *v1.Error
	}
	createReturnsOnCall map[int]struct {
		result1 *v1.Error
	}
	FindStub        func(*zap.Logger, v1.String) (queues.Queue, *v1.Error)
	findMutex       sync.RWMutex
	findArgsForCall []struct {
		arg1 *zap.Logger
		arg2 v1.String
	}
	findReturns struct {
		result1 queues.Queue
		result2 *v1.Error
	}
	findReturnsOnCall map[int]struct {
		result1 queues.Queue
		result2 *v1.Error
	}
	MetricsStub        func() *v1.MetricsResponse
	metricsMutex       sync.RWMutex
	metricsArgsForCall []struct {
	}
	metricsReturns struct {
		result1 *v1.MetricsResponse
	}
	metricsReturnsOnCall map[int]struct {
		result1 *v1.MetricsResponse
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeQueueManager) Create(arg1 *zap.Logger, arg2 *v1.Create) *v1.Error {
	fake.createMutex.Lock()
	ret, specificReturn := fake.createReturnsOnCall[len(fake.createArgsForCall)]
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		arg1 *zap.Logger
		arg2 *v1.Create
	}{arg1, arg2})
	stub := fake.CreateStub
	fakeReturns := fake.createReturns
	fake.recordInvocation("Create", []interface{}{arg1, arg2})
	fake.createMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeQueueManager) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeQueueManager) CreateCalls(stub func(*zap.Logger, *v1.Create) *v1.Error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = stub
}

func (fake *FakeQueueManager) CreateArgsForCall(i int) (*zap.Logger, *v1.Create) {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	argsForCall := fake.createArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeQueueManager) CreateReturns(result1 *v1.Error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 *v1.Error
	}{result1}
}

func (fake *FakeQueueManager) CreateReturnsOnCall(i int, result1 *v1.Error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	if fake.createReturnsOnCall == nil {
		fake.createReturnsOnCall = make(map[int]struct {
			result1 *v1.Error
		})
	}
	fake.createReturnsOnCall[i] = struct {
		result1 *v1.Error
	}{result1}
}

func (fake *FakeQueueManager) Find(arg1 *zap.Logger, arg2 v1.String) (queues.Queue, *v1.Error) {
	fake.findMutex.Lock()
	ret, specificReturn := fake.findReturnsOnCall[len(fake.findArgsForCall)]
	fake.findArgsForCall = append(fake.findArgsForCall, struct {
		arg1 *zap.Logger
		arg2 v1.String
	}{arg1, arg2})
	stub := fake.FindStub
	fakeReturns := fake.findReturns
	fake.recordInvocation("Find", []interface{}{arg1, arg2})
	fake.findMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeQueueManager) FindCallCount() int {
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
	return len(fake.findArgsForCall)
}

func (fake *FakeQueueManager) FindCalls(stub func(*zap.Logger, v1.String) (queues.Queue, *v1.Error)) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = stub
}

func (fake *FakeQueueManager) FindArgsForCall(i int) (*zap.Logger, v1.String) {
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
	argsForCall := fake.findArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeQueueManager) FindReturns(result1 queues.Queue, result2 *v1.Error) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = nil
	fake.findReturns = struct {
		result1 queues.Queue
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeQueueManager) FindReturnsOnCall(i int, result1 queues.Queue, result2 *v1.Error) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = nil
	if fake.findReturnsOnCall == nil {
		fake.findReturnsOnCall = make(map[int]struct {
			result1 queues.Queue
			result2 *v1.Error
		})
	}
	fake.findReturnsOnCall[i] = struct {
		result1 queues.Queue
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeQueueManager) Metrics() *v1.MetricsResponse {
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

func (fake *FakeQueueManager) MetricsCallCount() int {
	fake.metricsMutex.RLock()
	defer fake.metricsMutex.RUnlock()
	return len(fake.metricsArgsForCall)
}

func (fake *FakeQueueManager) MetricsCalls(stub func() *v1.MetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = stub
}

func (fake *FakeQueueManager) MetricsReturns(result1 *v1.MetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = nil
	fake.metricsReturns = struct {
		result1 *v1.MetricsResponse
	}{result1}
}

func (fake *FakeQueueManager) MetricsReturnsOnCall(i int, result1 *v1.MetricsResponse) {
	fake.metricsMutex.Lock()
	defer fake.metricsMutex.Unlock()
	fake.MetricsStub = nil
	if fake.metricsReturnsOnCall == nil {
		fake.metricsReturnsOnCall = make(map[int]struct {
			result1 *v1.MetricsResponse
		})
	}
	fake.metricsReturnsOnCall[i] = struct {
		result1 *v1.MetricsResponse
	}{result1}
}

func (fake *FakeQueueManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
	fake.metricsMutex.RLock()
	defer fake.metricsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeQueueManager) recordInvocation(key string, args []interface{}) {
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

var _ queues.QueueManager = new(FakeQueueManager)
