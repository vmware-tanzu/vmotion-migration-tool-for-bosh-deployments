// Code generated by counterfeiter. DO NOT EDIT.
package migratefakes

import (
	"context"
	"sync"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type FakeVCenterClient struct {
	DatacenterStub        func() string
	datacenterMutex       sync.RWMutex
	datacenterArgsForCall []struct {
	}
	datacenterReturns struct {
		result1 string
	}
	datacenterReturnsOnCall map[int]struct {
		result1 string
	}
	FindVMInClustersStub        func(context.Context, string, string, []string) (*vcenter.VM, error)
	findVMInClustersMutex       sync.RWMutex
	findVMInClustersArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 []string
	}
	findVMInClustersReturns struct {
		result1 *vcenter.VM
		result2 error
	}
	findVMInClustersReturnsOnCall map[int]struct {
		result1 *vcenter.VM
		result2 error
	}
	HostNameStub        func() string
	hostNameMutex       sync.RWMutex
	hostNameArgsForCall []struct {
	}
	hostNameReturns struct {
		result1 string
	}
	hostNameReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeVCenterClient) Datacenter() string {
	fake.datacenterMutex.Lock()
	ret, specificReturn := fake.datacenterReturnsOnCall[len(fake.datacenterArgsForCall)]
	fake.datacenterArgsForCall = append(fake.datacenterArgsForCall, struct {
	}{})
	stub := fake.DatacenterStub
	fakeReturns := fake.datacenterReturns
	fake.recordInvocation("Datacenter", []interface{}{})
	fake.datacenterMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeVCenterClient) DatacenterCallCount() int {
	fake.datacenterMutex.RLock()
	defer fake.datacenterMutex.RUnlock()
	return len(fake.datacenterArgsForCall)
}

func (fake *FakeVCenterClient) DatacenterCalls(stub func() string) {
	fake.datacenterMutex.Lock()
	defer fake.datacenterMutex.Unlock()
	fake.DatacenterStub = stub
}

func (fake *FakeVCenterClient) DatacenterReturns(result1 string) {
	fake.datacenterMutex.Lock()
	defer fake.datacenterMutex.Unlock()
	fake.DatacenterStub = nil
	fake.datacenterReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeVCenterClient) DatacenterReturnsOnCall(i int, result1 string) {
	fake.datacenterMutex.Lock()
	defer fake.datacenterMutex.Unlock()
	fake.DatacenterStub = nil
	if fake.datacenterReturnsOnCall == nil {
		fake.datacenterReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.datacenterReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeVCenterClient) FindVMInClusters(arg1 context.Context, arg2 string, arg3 string, arg4 []string) (*vcenter.VM, error) {
	var arg4Copy []string
	if arg4 != nil {
		arg4Copy = make([]string, len(arg4))
		copy(arg4Copy, arg4)
	}
	fake.findVMInClustersMutex.Lock()
	ret, specificReturn := fake.findVMInClustersReturnsOnCall[len(fake.findVMInClustersArgsForCall)]
	fake.findVMInClustersArgsForCall = append(fake.findVMInClustersArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 []string
	}{arg1, arg2, arg3, arg4Copy})
	stub := fake.FindVMInClustersStub
	fakeReturns := fake.findVMInClustersReturns
	fake.recordInvocation("FindVMInClusters", []interface{}{arg1, arg2, arg3, arg4Copy})
	fake.findVMInClustersMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeVCenterClient) FindVMInClustersCallCount() int {
	fake.findVMInClustersMutex.RLock()
	defer fake.findVMInClustersMutex.RUnlock()
	return len(fake.findVMInClustersArgsForCall)
}

func (fake *FakeVCenterClient) FindVMInClustersCalls(stub func(context.Context, string, string, []string) (*vcenter.VM, error)) {
	fake.findVMInClustersMutex.Lock()
	defer fake.findVMInClustersMutex.Unlock()
	fake.FindVMInClustersStub = stub
}

func (fake *FakeVCenterClient) FindVMInClustersArgsForCall(i int) (context.Context, string, string, []string) {
	fake.findVMInClustersMutex.RLock()
	defer fake.findVMInClustersMutex.RUnlock()
	argsForCall := fake.findVMInClustersArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeVCenterClient) FindVMInClustersReturns(result1 *vcenter.VM, result2 error) {
	fake.findVMInClustersMutex.Lock()
	defer fake.findVMInClustersMutex.Unlock()
	fake.FindVMInClustersStub = nil
	fake.findVMInClustersReturns = struct {
		result1 *vcenter.VM
		result2 error
	}{result1, result2}
}

func (fake *FakeVCenterClient) FindVMInClustersReturnsOnCall(i int, result1 *vcenter.VM, result2 error) {
	fake.findVMInClustersMutex.Lock()
	defer fake.findVMInClustersMutex.Unlock()
	fake.FindVMInClustersStub = nil
	if fake.findVMInClustersReturnsOnCall == nil {
		fake.findVMInClustersReturnsOnCall = make(map[int]struct {
			result1 *vcenter.VM
			result2 error
		})
	}
	fake.findVMInClustersReturnsOnCall[i] = struct {
		result1 *vcenter.VM
		result2 error
	}{result1, result2}
}

func (fake *FakeVCenterClient) HostName() string {
	fake.hostNameMutex.Lock()
	ret, specificReturn := fake.hostNameReturnsOnCall[len(fake.hostNameArgsForCall)]
	fake.hostNameArgsForCall = append(fake.hostNameArgsForCall, struct {
	}{})
	stub := fake.HostNameStub
	fakeReturns := fake.hostNameReturns
	fake.recordInvocation("HostName", []interface{}{})
	fake.hostNameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeVCenterClient) HostNameCallCount() int {
	fake.hostNameMutex.RLock()
	defer fake.hostNameMutex.RUnlock()
	return len(fake.hostNameArgsForCall)
}

func (fake *FakeVCenterClient) HostNameCalls(stub func() string) {
	fake.hostNameMutex.Lock()
	defer fake.hostNameMutex.Unlock()
	fake.HostNameStub = stub
}

func (fake *FakeVCenterClient) HostNameReturns(result1 string) {
	fake.hostNameMutex.Lock()
	defer fake.hostNameMutex.Unlock()
	fake.HostNameStub = nil
	fake.hostNameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeVCenterClient) HostNameReturnsOnCall(i int, result1 string) {
	fake.hostNameMutex.Lock()
	defer fake.hostNameMutex.Unlock()
	fake.HostNameStub = nil
	if fake.hostNameReturnsOnCall == nil {
		fake.hostNameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.hostNameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeVCenterClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.datacenterMutex.RLock()
	defer fake.datacenterMutex.RUnlock()
	fake.findVMInClustersMutex.RLock()
	defer fake.findVMInClustersMutex.RUnlock()
	fake.hostNameMutex.RLock()
	defer fake.hostNameMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeVCenterClient) recordInvocation(key string, args []interface{}) {
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

var _ migrate.VCenterClient = new(FakeVCenterClient)
