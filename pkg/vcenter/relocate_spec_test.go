/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"testing"
)

func TestBuildRelocateSpec(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		c := vcenter.NewFromGovmomiClient(client, "DC0")

		finder := vcenter.NewFinder("DC0", client)
		hosts, err := finder.HostsInCluster(ctx, "DC0_C0")
		require.NoError(t, err)

		vm, err := c.FindVMInClusters(ctx, "az1", "DC0_C0_RP1_VM0", []string{"DC0_C0"})
		require.NoError(t, err)

		// since we only have one vcenter everything maps to the same as the source
		ts := &vcenter.TargetSpec{
			Name:         "DC0_C0_RP1_VM0",
			Datacenter:   "DC0",
			Cluster:      "DC0_C0",
			ResourcePool: "DC0_C0_RP1",
			Folder:       "/DC0/vm",
			Networks: map[string]string{
				"DC0_DVPG0": "DC0_DVPG0",
			},
			Datastores: map[string]string{
				"LocalDS_0": "LocalDS_0",
			},
		}

		_, err = vcenter.NewRelocateSpec(c, c).WithTargetSpec(ts).WithTargetHost(hosts[0]).Build(ctx)
		require.ErrorContains(t, err, "must set a source VM first before calling build")

		_, err = vcenter.NewRelocateSpec(c, c).WithSourceVM(vm).WithTargetHost(hosts[0]).Build(ctx)
		require.ErrorContains(t, err, "must set a target VM spec first before calling build")

		_, err = vcenter.NewRelocateSpec(c, c).WithSourceVM(vm).WithTargetSpec(ts).Build(ctx)
		require.ErrorContains(t, err, "must set a target host first before calling build")

		rs := vcenter.NewRelocateSpec(c, c).WithSourceVM(vm).WithTargetSpec(ts).WithTargetHost(hosts[0])
		spec, err := rs.Build(ctx)
		require.NoError(t, err)
		require.NotNil(t, spec)

		// the target datastore
		require.NotNil(t, spec.Datastore)
		require.Contains(t, spec.Datastore.Value, "datastore-")

		// optional resource pool mapping
		require.NotNil(t, spec.Pool)
		require.Contains(t, spec.Pool.Value, "resgroup-")

		// target host mapping to support spreading out the migration load
		require.NotNil(t, spec.Host)
		require.Contains(t, spec.Host.Value, "host-")

		// each disk must be mapped to specific datastores
		require.Len(t, spec.Disk, 1)
		require.Equal(t, spec.Disk[0].DiskId, int32(204))
		require.Equal(t, spec.Disk[0].Datastore.Value, spec.Datastore.Value)

		// eth0 needs to be re-connected to the target switch
		require.Len(t, spec.DeviceChange, 1)
		eth := spec.DeviceChange[0].GetVirtualDeviceConfigSpec().Device.GetVirtualDevice()
		require.NotNil(t, eth)

		require.Equal(t, "ethernet-0", eth.DeviceInfo.GetDescription().Label)
		require.Equal(t, "DVSwitch: fea97929-4b2d-5972-b146-930c6d0b4014", eth.DeviceInfo.GetDescription().Summary)

		ethBacking := eth.Backing.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
		require.Equal(t, "fea97929-4b2d-5972-b146-930c6d0b4014", ethBacking.Port.SwitchUuid)
		require.Contains(t, ethBacking.Port.PortgroupKey, "dvportgroup-")

		// target the same folder as the source structure
		require.NotNil(t, spec.Folder)
		require.Contains(t, spec.Folder.Value, "group-")
	})
}

func TestBuildRelocateSpecNoResourcePool(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		c := vcenter.NewFromGovmomiClient(client, "DC0")

		finder := vcenter.NewFinder("DC0", client)
		hosts, err := finder.HostsInCluster(ctx, "DC0_C0")
		require.NoError(t, err)

		vm, err := c.FindVMInClusters(ctx, "az1", "DC0_C0_RP1_VM0", []string{"DC0_C0"})
		require.NoError(t, err)

		// since we only have one vcenter everything maps to the same as the source
		ts := &vcenter.TargetSpec{
			Name:       "DC0_C0_RP1_VM0",
			Datacenter: "DC0",
			Cluster:    "DC0_C0",
			Folder:     "/DC0/vm",
			Networks: map[string]string{
				"DC0_DVPG0": "DC0_DVPG0",
			},
			Datastores: map[string]string{
				"LocalDS_0": "LocalDS_0",
			},
		}

		rs := vcenter.NewRelocateSpec(c, c).WithSourceVM(vm).WithTargetSpec(ts).WithTargetHost(hosts[0])
		spec, err := rs.Build(ctx)
		require.NoError(t, err)
		require.NotNil(t, spec)

		// default resource pool mapping
		require.NotNil(t, spec.Pool)
		require.Contains(t, spec.Pool.Value, "resgroup-")
	})
}
