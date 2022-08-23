/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type RelocateSpec struct {
	sourceClient      *Client
	destinationClient *Client

	srcVM        *VM
	vmTargetSpec *TargetSpec
	targetHost   *object.HostSystem
}

func NewRelocateSpec(sourceClient *Client, destinationClient *Client) *RelocateSpec {
	return &RelocateSpec{
		sourceClient:      sourceClient,
		destinationClient: destinationClient,
	}
}

func (rs *RelocateSpec) WithSourceVM(vm *VM) *RelocateSpec {
	rs.srcVM = vm
	return rs
}

func (rs *RelocateSpec) WithTargetHost(host *object.HostSystem) *RelocateSpec {
	rs.targetHost = host
	return rs
}

func (rs *RelocateSpec) WithTargetSpec(vmTargetSpec *TargetSpec) *RelocateSpec {
	rs.vmTargetSpec = vmTargetSpec
	return rs
}

func (rs *RelocateSpec) Build(ctx context.Context) (*types.VirtualMachineRelocateSpec, error) {
	if rs.srcVM == nil {
		return nil, fmt.Errorf("must set a source VM first before callling build")
	}
	if rs.targetHost == nil {
		return nil, fmt.Errorf("must set a target host first before calling build")
	}
	if rs.vmTargetSpec == nil {
		return nil, fmt.Errorf("must set a target VM spec first before callling build")
	}

	hostRef := rs.targetHost.Reference()

	sourceClient, err := rs.sourceClient.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return nil, err
	}

	destinationClient, err := rs.destinationClient.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return nil, err
	}

	sourceFinder := NewFinder(rs.srcVM.Datacenter, sourceClient)
	destinationFinder := NewFinder(rs.vmTargetSpec.Datacenter, destinationClient)

	poolRef, err := destinationFinder.ResourcePoolFromSpecRef(ctx, *rs.vmTargetSpec)
	if err != nil {
		return nil, err
	}

	dsRef, err := destinationFinder.DatastoreRef(ctx, rs.vmTargetSpec.Datastore)
	if err != nil {
		return nil, err
	}

	adapterUpdater := NewAdapterUpdater(destinationFinder)
	var devicesToChange []types.BaseVirtualDeviceConfigSpec
	for sourceNetName, targetNetName := range rs.vmTargetSpec.Networks {
		srcNetworkAdapter, err := sourceFinder.Adapter(ctx, rs.srcVM.Name, sourceNetName)
		if err != nil {
			return nil, err
		}

		updatedAdapter, err := adapterUpdater.TargetNewNetwork(ctx, srcNetworkAdapter, targetNetName)
		if err != nil {
			return nil, err
		}

		deviceToChange := types.VirtualDeviceConfigSpec{}
		deviceToChange.Operation = types.VirtualDeviceConfigSpecOperationEdit
		deviceToChange.Device = updatedAdapter.Device()
		devicesToChange = append(devicesToChange, &deviceToChange)
	}

	spec := &types.VirtualMachineRelocateSpec{}
	spec.Host = &hostRef
	spec.Pool = poolRef
	spec.Datastore = dsRef
	spec.DeviceChange = devicesToChange

	// if source and target vcenter are different
	if rs.sourceClient.URL().String() != rs.destinationClient.URL().String() {
		targetThumbprint, err := rs.destinationClient.thumbprint(ctx)
		if err != nil {
			return nil, err
		}

		spec.Service = &types.ServiceLocator{
			Url:          rs.destinationClient.URL().String(),
			InstanceUuid: destinationClient.ServiceContent.About.InstanceUuid,
			Credential: &types.ServiceLocatorNamePassword{
				Username: rs.destinationClient.Username,
				Password: rs.destinationClient.Password,
			},
			SslThumbprint: targetThumbprint,
		}
	}

	return spec, nil
}
