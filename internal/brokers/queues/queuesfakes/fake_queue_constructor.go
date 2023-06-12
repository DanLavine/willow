// Code generated by counterfeiter. DO NOT EDIT.
package queuesfakes

import (
	"sync"

	"github.com/DanLavine/willow/internal/brokers/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type FakeQueueConstructor struct {
	NewQueueStub        func(*v1.Create) (queues.ManagedQueue, *v1.Error)
	newQueueMutex       sync.RWMutex
	newQueueArgsForCall []struct {
		arg1 *v1.Create
	}
	newQueueReturns struct {
		result1 queues.ManagedQueue
		result2 *v1.Error
	}
	newQueueReturnsOnCall map[int]struct {
		result1 queues.ManagedQueue
		result2 *v1.Error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeQueueConstructor) NewQueue(arg1 *v1.Create) (queues.ManagedQueue, *v1.Error) {
	fake.newQueueMutex.Lock()
	ret, specificReturn := fake.newQueueReturnsOnCall[len(fake.newQueueArgsForCall)]
	fake.newQueueArgsForCall = append(fake.newQueueArgsForCall, struct {
		arg1 *v1.Create
	}{arg1})
	stub := fake.NewQueueStub
	fakeReturns := fake.newQueueReturns
	fake.recordInvocation("NewQueue", []interface{}{arg1})
	fake.newQueueMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeQueueConstructor) NewQueueCallCount() int {
	fake.newQueueMutex.RLock()
	defer fake.newQueueMutex.RUnlock()
	return len(fake.newQueueArgsForCall)
}

func (fake *FakeQueueConstructor) NewQueueCalls(stub func(*v1.Create) (queues.ManagedQueue, *v1.Error)) {
	fake.newQueueMutex.Lock()
	defer fake.newQueueMutex.Unlock()
	fake.NewQueueStub = stub
}

func (fake *FakeQueueConstructor) NewQueueArgsForCall(i int) *v1.Create {
	fake.newQueueMutex.RLock()
	defer fake.newQueueMutex.RUnlock()
	argsForCall := fake.newQueueArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeQueueConstructor) NewQueueReturns(result1 queues.ManagedQueue, result2 *v1.Error) {
	fake.newQueueMutex.Lock()
	defer fake.newQueueMutex.Unlock()
	fake.NewQueueStub = nil
	fake.newQueueReturns = struct {
		result1 queues.ManagedQueue
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeQueueConstructor) NewQueueReturnsOnCall(i int, result1 queues.ManagedQueue, result2 *v1.Error) {
	fake.newQueueMutex.Lock()
	defer fake.newQueueMutex.Unlock()
	fake.NewQueueStub = nil
	if fake.newQueueReturnsOnCall == nil {
		fake.newQueueReturnsOnCall = make(map[int]struct {
			result1 queues.ManagedQueue
			result2 *v1.Error
		})
	}
	fake.newQueueReturnsOnCall[i] = struct {
		result1 queues.ManagedQueue
		result2 *v1.Error
	}{result1, result2}
}

func (fake *FakeQueueConstructor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newQueueMutex.RLock()
	defer fake.newQueueMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeQueueConstructor) recordInvocation(key string, args []interface{}) {
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

var _ queues.QueueConstructor = new(FakeQueueConstructor)
