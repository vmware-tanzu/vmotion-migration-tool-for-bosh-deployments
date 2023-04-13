/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"
)

func TestFindHostsInCluster(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)
		hosts, err := finder.HostsInCluster(context.Background(), "DC0_C0")
		require.NoError(t, err)
		require.Len(t, hosts, 3)
		require.Equal(t, "DC0_C0_H0", hosts[0].Name())
		require.Equal(t, "DC0_C0_H1", hosts[1].Name())
		require.Equal(t, "DC0_C0_H2", hosts[2].Name())
	})
}

func TestVirtualMachine(t *testing.T) {
	expectedVMs := []string{
		"DC0_C0_RP1_VM0",
		"DC0_C0_RP1_VM1",
		"DC0_H0_VM0",
		"DC0_H0_VM0",
	}

	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)
		for _, expectedVM := range expectedVMs {
			vm, err := finder.VirtualMachine(ctx, expectedVM)
			require.NoError(t, err)
			require.Equal(t, expectedVM, vm.Name())
		}

		_, err := finder.VirtualMachine(ctx, "does-not-exist-VM")
		require.Error(t, err)
	})
}

func TestCluster(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		t.Run("Find cluster", func(t *testing.T) {
			cluster, err := finder.Cluster(ctx, "DC0_C0")
			require.NoError(t, err)
			require.Equal(t, "DC0_C0", cluster.Name())
			require.Equal(t, "/DC0/host/DC0_C0", cluster.InventoryPath)
		})

		t.Run("Non-existent cluster", func(t *testing.T) {
			_, err := finder.Cluster(ctx, "not-a-cluster")
			require.Error(t, err)
			require.Contains(t, err.Error(), "cluster 'not-a-cluster' not found")
		})
	})
}

func TestVMClusterName(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		t.Run("Find cluster", func(t *testing.T) {
			vm0, err := finder.VirtualMachine(ctx, "DC0_C0_RP1_VM0")
			require.NoError(t, err)
			cluster, err := finder.VMClusterName(ctx, vm0)
			require.NoError(t, err)
			require.Equal(t, "DC0_C0", cluster)

			vm1, err := finder.VirtualMachine(ctx, "DC0_C0_RP1_VM1")
			require.NoError(t, err)
			cluster, err = finder.VMClusterName(ctx, vm1)
			require.NoError(t, err)
			require.Equal(t, "DC0_C0", cluster)
		})

		t.Run("Non-existent cluster", func(t *testing.T) {
			vm0, err := finder.VirtualMachine(ctx, "DC0_H0_VM0")
			require.NoError(t, err)
			_, err = finder.VMClusterName(ctx, vm0)
			require.Error(t, err)
			require.Contains(t, err.Error(), "found unsupported compute type ComputeResource")
		})
	})
}

func TestResourcePool(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		t.Run("Find RP", func(t *testing.T) {
			rp, err := finder.ResourcePool(ctx, "/DC0/host/DC0_C0/Resources/DC0_C0_RP1")
			require.NoError(t, err)
			require.Equal(t, "DC0_C0_RP1", rp.Name())
		})

		t.Run("Non-existent RP", func(t *testing.T) {
			_, err := finder.ResourcePool(ctx, "does-not-exist-RP")
			require.Error(t, err)
		})
	})
}

func TestDatastore(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		t.Run("Find datastore", func(t *testing.T) {
			ds, err := finder.Datastore(ctx, "LocalDS_0")
			require.NoError(t, err)
			require.Equal(t, "LocalDS_0", ds.Name())

			dsRef, err := finder.DatastoreRef(ctx, "LocalDS_0")
			require.NoError(t, err)
			require.Equal(t, ds.Reference(), *dsRef)
		})

		t.Run("Non-existent datastore", func(t *testing.T) {
			_, err := finder.Datastore(ctx, "NFSDatastore_01")
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to find datastore NFSDatastore_01:")
		})
	})
}

func TestDisks(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		vm, err := finder.VirtualMachine(ctx, "DC0_C0_RP1_VM0")
		require.NoError(t, err)
		disks, err := finder.Disks(ctx, vm)
		require.NoError(t, err)
		require.Len(t, disks, 1)
		require.Equal(t, "LocalDS_0", disks[0].Datastore)
		require.NotEqual(t, int32(0), disks[0].ID)
	})
}

func TestNetworks(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)
		vm, err := finder.VirtualMachine(ctx, "DC0_C0_RP1_VM0")
		require.NoError(t, err)

		nets, err := finder.Networks(ctx, vm)
		require.NoError(t, err)
		require.Len(t, nets, 1)
		require.Equal(t, "DC0_DVPG0", nets[0])
	})
}

func TestAdapterBackingInfo(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)
		info, err := finder.AdapterBackingInfo(ctx, "DC0_DVPG0")
		require.NoError(t, err)
		require.NotNil(t, info)
	})
}

func TestAdapter(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		// grab the expected network from the simulator map
		net := findSimulatorObject("DistributedVirtualPortgroup", "DC0_DVPG0").(*simulator.DistributedVirtualPortgroup)

		// get the adapter
		finder := vcenter.NewFinder("DC0", client)
		adapter, err := finder.Adapter(ctx, "DC0_C0_RP1_VM0", "DC0_DVPG0")
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// if it's the right adapter it should be plugged into the expected network
		info := adapter.VirtualE1000.Backing.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
		require.Equal(t, net.DistributedVirtualPortgroup.Key, info.Port.PortgroupKey)
	})
}

func TestFolder(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		// get the adapter
		finder := vcenter.NewFinder("DC0", client)
		folder, err := finder.Folder(ctx, "/DC0/vm")
		require.NoError(t, err)
		require.NotNil(t, folder)
		require.Equal(t, "/DC0/vm", folder.InventoryPath)
		require.Equal(t, "vm", folder.Name())
	})
}
