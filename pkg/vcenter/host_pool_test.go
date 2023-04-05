/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter_test

import (
	"context"
	"github.com/vmware/govmomi/simulator"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware/govmomi"
)

func TestLeaseAvailableHost(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		vcenterClient := vcenter.NewFromGovmomiClient(client, "DC0")
		azToVCenterMap := map[string]*vcenter.Client{
			"az1": vcenterClient,
		}
		vcenterPool := vcenter.NewPoolWithExternalClients(azToVCenterMap, azToVCenterMap)
		hpc := &vcenter.HostPoolConfig{
			AZs: map[string]vcenter.HostPoolAZ{"az1": {
				Clusters: []string{
					"DC0_C0",
				},
			}},
		}

		hostPool := vcenter.NewHostPool(vcenterPool, hpc)
		hostPool.MaxLeasePerHost = 2
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease a host and release it
		host, err := hostPool.LeaseAvailableHost(ctx, "az1")
		require.NoError(t, err)
		require.NotNil(t, host)
		require.Contains(t, host.Name(), "DC0_C0_H")
		hostPool.Release(ctx, host)

		// lease all 3 hosts twice
		for i := 0; i < 6; i++ {
			host, err = hostPool.LeaseAvailableHost(ctx, "az1")
			require.NoError(t, err)
			require.NotNil(t, host)
			require.Contains(t, host.Name(), "DC0_C0_H")
		}

		// try to get another lease, should fail and return nil
		host, err = hostPool.LeaseAvailableHost(ctx, "az1")
		require.NoError(t, err)
		require.Nil(t, host)
	})
}

func TestLeaseAvailableHostMin(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		vcenterClient := vcenter.NewFromGovmomiClient(client, "DC0")
		azToVCenterMap := map[string]*vcenter.Client{
			"az1": vcenterClient,
		}
		vcenterPool := vcenter.NewPoolWithExternalClients(azToVCenterMap, azToVCenterMap)
		hpc := &vcenter.HostPoolConfig{
			AZs: map[string]vcenter.HostPoolAZ{"az1": {
				Clusters: []string{
					"DC0_C0",
				},
			}},
		}

		hostPool := vcenter.NewHostPool(vcenterPool, hpc)
		hostPool.MaxLeasePerHost = 1
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease all 3 hosts once
		for i := 0; i < 3; i++ {
			host, err := hostPool.LeaseAvailableHost(ctx, "az1")
			require.NoError(t, err)
			require.NotNil(t, host)
			require.Contains(t, host.Name(), "DC0_C0_H")
		}

		// try to get another lease, should fail and return nil
		host, err := hostPool.LeaseAvailableHost(ctx, "az1")
		require.NoError(t, err)
		require.Nil(t, host)
	})
}

func TestLeaseAvailableHostInNonExistentAZ(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		vcenterClient := vcenter.NewFromGovmomiClient(client, "DC0")
		azToVCenterMap := map[string]*vcenter.Client{
			"az1": vcenterClient,
		}
		vcenterPool := vcenter.NewPoolWithExternalClients(azToVCenterMap, azToVCenterMap)
		hpc := &vcenter.HostPoolConfig{
			AZs: map[string]vcenter.HostPoolAZ{"az1": {
				Clusters: []string{
					"DC0_C0",
				},
			}},
		}

		hostPool := vcenter.NewHostPool(vcenterPool, hpc)
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease a host and release it
		_, err = hostPool.LeaseAvailableHost(ctx, "az2")
		require.Error(t, err)
	})
}

func TestLeaseAvailableHostSkipsHostsInMaintenanceMode(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		// put first host in maintenance mode
		h := findSimulatorObject("HostSystem", "DC0_C0_H1").(*simulator.HostSystem)
		h.Runtime.InMaintenanceMode = true

		vcenterClient := vcenter.NewFromGovmomiClient(client, "DC0")
		azToVCenterMap := map[string]*vcenter.Client{
			"az1": vcenterClient,
		}
		vcenterPool := vcenter.NewPoolWithExternalClients(azToVCenterMap, azToVCenterMap)
		hpc := &vcenter.HostPoolConfig{
			AZs: map[string]vcenter.HostPoolAZ{"az1": {
				Clusters: []string{
					"DC0_C0",
				},
			}},
		}

		hostPool := vcenter.NewHostPool(vcenterPool, hpc)
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease first two hosts
		for i := 0; i < 2; i++ {
			host, err := hostPool.LeaseAvailableHost(ctx, "az1")
			require.NoError(t, err)
			require.NotNil(t, host)
			require.NotContains(t, host.Name(), "DC0_C0_H1")
		}

		// third lease should fail since we only have two valid hosts
		host, err := hostPool.LeaseAvailableHost(ctx, "az1")
		require.NoError(t, err)
		require.Nil(t, host)
	})
}
