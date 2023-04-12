/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/migratefakes"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

func TestVMMigrator_MigrateVMToTarget(t *testing.T) {
	vmToMigrate := migrate.VM{
		Name:     "vm1",
		AZ:       "az1",
		Clusters: []string{"Cluster1"},
	}

	sourceClient := &migratefakes.FakeVCenterClient{}
	sourceClient.FindVMInClustersReturnsOnCall(0, &vcenter.VM{
		Name:         "vm1",
		AZ:           "az1",
		Datacenter:   "DC1",
		Cluster:      "Cluster1",
		Folder:       "/DC1/vm",
		ResourcePool: "RP1",
		Disks: []vcenter.Disk{
			{
				ID:        201,
				Datastore: "DS1",
			},
		},
		Networks: []string{"Net1"},
	}, nil)

	vmConverter := converter.New(
		converter.NewEmptyMappedNetwork().Add("Net1", "Net2"),
		converter.NewEmptyMappedDatastore().Add("DS1", "DS2"),
		converter.NewEmptyMappedCompute().Add(converter.AZ{
			Datacenter:   "DC1",
			Cluster:      "Cluster1",
			ResourcePool: "RP1",
			Name:         "az1",
		}, converter.AZ{
			Datacenter:   "DC2",
			Cluster:      "Cluster2",
			ResourcePool: "RP2",
			Name:         "az1",
		}))

	out := log.NewUpdatableStdout()
	vmRelocator := &migratefakes.FakeVMRelocator{}
	vmMigrator := migrate.NewVMMigrator(&vcenter.Pool{}, vmConverter, vmRelocator, out)

	err := vmMigrator.MigrateVMToTarget(context.Background(), sourceClient, vmToMigrate)
	require.NoError(t, err)

	_, srcVM, targetSpec := vmRelocator.RelocateVMArgsForCall(0)
	require.Equal(t, "vm1", srcVM.Name)
	require.Equal(t, "vm1", targetSpec.Name)
	require.Equal(t, "DC2", targetSpec.Datacenter)
	require.Equal(t, "RP2", targetSpec.ResourcePool)
	require.Equal(t, "Cluster2", targetSpec.Cluster)
	require.Equal(t, map[string]string{"DS1": "DS2"}, targetSpec.Datastores)
	require.Equal(t, map[string]string{"Net1": "Net2"}, targetSpec.Networks)
}

func TestVMMigrator_MigrateVMToTarget_VMNotFound(t *testing.T) {
	vmToMigrate := migrate.VM{
		Name:     "vm1",
		AZ:       "az1",
		Clusters: []string{"Cluster1"},
	}

	sourceClient := &migratefakes.FakeVCenterClient{}
	sourceClient.FindVMInClustersReturnsOnCall(0, nil, &vcenter.VMNotFoundError{})

	vmConverter := converter.New(
		converter.NewEmptyMappedNetwork().Add("Net1", "Net2"),
		converter.NewEmptyMappedDatastore().Add("DS1", "DS2"),
		converter.NewEmptyMappedCompute().Add(converter.AZ{
			Datacenter:   "DC1",
			Cluster:      "Cluster1",
			ResourcePool: "RP1",
			Name:         "az1",
		}, converter.AZ{
			Datacenter:   "DC2",
			Cluster:      "Cluster2",
			ResourcePool: "RP2",
			Name:         "az1",
		}))

	out := log.NewBufferedStdout()
	vmRelocator := &migratefakes.FakeVMRelocator{}
	vmMigrator := migrate.NewVMMigrator(&vcenter.Pool{}, vmConverter, vmRelocator, out)

	err := vmMigrator.MigrateVMToTarget(context.Background(), sourceClient, vmToMigrate)
	require.NoError(t, err)
	require.Equal(t, 0, vmRelocator.RelocateVMCallCount())

	require.Contains(t, out.String(), "not found in source vCenter, skipping")
}
