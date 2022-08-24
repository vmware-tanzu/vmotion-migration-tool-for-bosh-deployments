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
		"DC0_C0_RP0_VM0",
		"DC0_C0_RP0_VM1",
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

func TestResourcePool(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)

		t.Run("Find RP", func(t *testing.T) {
			rp, err := finder.ResourcePool(ctx, "/DC0/host/DC0_C0/Resources/DC0_C0_RP1")
			require.NoError(t, err)
			require.Equal(t, "DC0_C0_RP1", rp.Name())
		})

		t.Run("Find RP from spec", func(t *testing.T) {
			spec := vcenter.TargetSpec{
				ResourcePool: "DC0_C0_RP1",
				Datacenter:   "DC0",
				Cluster:      "DC0_C0",
			}
			rp, err := finder.ResourcePoolFromSpec(ctx, spec)
			require.NoError(t, err)
			require.Equal(t, "DC0_C0_RP1", rp.Name())

			rpRef, err := finder.ResourcePoolFromSpecRef(ctx, spec)
			require.NoError(t, err)
			require.Equal(t, rp.Reference(), *rpRef)
		})

		t.Run("Find RP from spec errors with DC mismatch", func(t *testing.T) {
			spec := vcenter.TargetSpec{
				ResourcePool: "DC0_C0_RP1",
				Datacenter:   "not-a-dc", //doesn't match finder DC
				Cluster:      "DC0_C0",
			}
			_, err := finder.ResourcePoolFromSpec(ctx, spec)
			require.Error(t, err)
			require.Equal(t, "mismatched resource pool datacenter, expected DC0 but got not-a-dc", err.Error())
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

func TestNetworks(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		finder := vcenter.NewFinder("DC0", client)
		vm, err := finder.VirtualMachine(ctx, "DC0_C0_RP0_VM0")
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
		adapter, err := finder.Adapter(ctx, "DC0_C0_RP0_VM0", "DC0_DVPG0")
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// if it's the right adapter it should be plugged into the expected network
		info := adapter.VirtualE1000.Backing.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
		require.Equal(t, net.DistributedVirtualPortgroup.Key, info.Port.PortgroupKey)
	})
}
