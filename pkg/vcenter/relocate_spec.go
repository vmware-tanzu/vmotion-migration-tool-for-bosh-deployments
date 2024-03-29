/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"sort"
)

type RelocateSpec struct {
	DryRun bool

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

func (rs *RelocateSpec) WithDryRun(dryRun bool) *RelocateSpec {
	rs.DryRun = dryRun
	return rs
}

func (rs *RelocateSpec) Build(ctx context.Context) (*types.VirtualMachineRelocateSpec, error) {
	if rs.srcVM == nil {
		return nil, fmt.Errorf("must set a source VM first before calling build")
	}
	if rs.targetHost == nil {
		return nil, fmt.Errorf("must set a target host first before calling build")
	}
	if rs.vmTargetSpec == nil {
		return nil, fmt.Errorf("must set a target VM spec first before calling build")
	}

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

	// get the target cluster to build a fully qualified resource pool path
	targetCluster, err := destinationFinder.Cluster(ctx, rs.vmTargetSpec.Cluster)
	if err != nil {
		return nil, err
	}

	// get the (required) specified or default target resource pool
	rpPath := targetCluster.InventoryPath + "/Resources"
	if rs.vmTargetSpec.ResourcePool != "" {
		rpPath += "/" + rs.vmTargetSpec.ResourcePool
	}
	poolRef, err := destinationFinder.ResourcePoolRef(ctx, rpPath)
	if err != nil {
		return nil, err
	}

	// get the destination folder for the VM
	folderRef, err := destinationFinder.FolderRef(ctx, rs.vmTargetSpec.Folder)
	if err != nil {
		if rs.DryRun {
			// folder might not exist yet during dry run
			log.FromContext(ctx).Debugf("Could not find target VM folder %s, but continuing since dry-run is true",
				rs.vmTargetSpec.Folder)
		} else {
			return nil, fmt.Errorf(
				"could not find destination VM folder '%s', ensure the folder exists: %w",
				rs.vmTargetSpec.Folder, err)
		}
	}

	// map the VM disks to their datastores
	var diskMappings []types.VirtualMachineRelocateSpecDiskLocator
	for _, srcDisk := range rs.srcVM.Disks {
		targetDiskDatastore, ok := rs.vmTargetSpec.Datastores[srcDisk.Datastore]
		if !ok {
			return nil, fmt.Errorf("could not find target datastore for disk %d on source datastore %s",
				srcDisk.ID, srcDisk.Datastore)
		}
		targetDiskDatastoreRef, err := destinationFinder.DatastoreRef(ctx, targetDiskDatastore)
		if err != nil {
			return nil, err
		}
		diskMappings = append(diskMappings, types.VirtualMachineRelocateSpecDiskLocator{
			DiskId:    srcDisk.ID,
			Datastore: *targetDiskDatastoreRef,
		})
	}

	if len(diskMappings) == 0 {
		return nil, fmt.Errorf("found 0 disk mappings for VM %s", rs.srcVM.Name)
	}

	// ensure first device is first in list - below we use that as the default datastore for the VM
	sort.Slice(diskMappings, func(i, j int) bool {
		return diskMappings[i].DiskId < diskMappings[j].DiskId
	})

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

	// the ESXi host we're targeting
	hostRef := rs.targetHost.Reference()

	// now that we have all the details, create the migration spec
	spec := &types.VirtualMachineRelocateSpec{}
	spec.Host = &hostRef
	spec.Pool = poolRef
	spec.Folder = folderRef
	spec.Datastore = &diskMappings[0].Datastore
	spec.Disk = diskMappings
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
				Username: rs.destinationClient.UserName(),
				Password: rs.destinationClient.Password(),
			},
			SslThumbprint: targetThumbprint,
		}
	}

	return spec, nil
}
