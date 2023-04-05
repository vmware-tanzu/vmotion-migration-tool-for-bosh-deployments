/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"testing"
)

func VPXTest(f func(context.Context, *govmomi.Client)) {
	model := simulator.VPX()
	defer model.Remove()
	model.Pool = 1

	simulator.Test(func(ctx context.Context, vimClient *vim25.Client) {
		c := &govmomi.Client{
			Client:         vimClient,
			SessionManager: session.NewManager(vimClient),
		}
		f(ctx, c)
	}, model)
}

func findSimulatorObject(kind, name string) mo.Entity {
	for _, o := range simulator.Map.All(kind) {
		if o.Entity().Name == name {
			return o
		}
	}
	return nil
}

func TestFindVMInCluster(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		c := vcenter.NewFromGovmomiClient(client, "DC0")
		vm, err := c.FindVMInClusters(ctx, "az1", "DC0_C0_RP1_VM0", []string{"DC0_C0"})
		require.NoError(t, err)
		require.Equal(t, "DC0_C0_RP1_VM0", vm.Name)

		// ensure the name is populated using name and not inventory path
		vm, err = c.FindVMInClusters(ctx, "az1", "/DC0/vm/DC0_C0_RP1_VM0", []string{"DC0_C0"})
		require.NoError(t, err)
		require.Equal(t, "DC0_C0_RP1_VM0", vm.Name)

		_, err = c.FindVMInClusters(ctx, "az1", "DC0_C0_RP1_VM0", []string{"DC0_C1"})
		require.Error(t, err)

		_, err = c.FindVMInClusters(ctx, "az1", "DC0_C1_RP1_VM0", []string{"DC0_C0", "DC0_C1"})
		require.Error(t, err)
	})
}
