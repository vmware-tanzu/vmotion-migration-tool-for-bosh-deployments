/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/migratefakes"
	"testing"
)

func baseSourceConfig() config.Config {
	return config.Config{
		DryRun:         false,
		WorkerPoolSize: 1,
		NetworkMap: map[string]string{
			"Net1": "Net2",
		},
		DatastoreMap: map[string]string{
			"DS1": "DS2",
		},
		Compute: config.Compute{
			Source: []config.ComputeAZ{
				{
					Name: "az1",
					VCenter: &config.VCenter{
						Host:       "vcenter1.example.com",
						Username:   "admin1",
						Password:   "secret1",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster1",
							ResourcePool: "RP1",
						},
						{
							Name:         "Cluster2",
							ResourcePool: "RP2",
						},
					},
				},
				{
					Name: "az2",
					VCenter: &config.VCenter{
						Host:       "vcenter1.example.com",
						Username:   "admin1",
						Password:   "secret1",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster3",
							ResourcePool: "RP1",
						},
					},
				},
				{
					Name: "az3",
					VCenter: &config.VCenter{
						Host:       "vcenter1.example.com",
						Username:   "admin1",
						Password:   "secret1",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster4",
							ResourcePool: "RP1",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name: "az1",
					VCenter: &config.VCenter{
						Host:       "vcenter2.example.com",
						Username:   "admin2",
						Password:   "secret2",
						Datacenter: "DC2",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster4",
							ResourcePool: "RP4",
						},
					},
				},
				{
					Name: "az2",
					VCenter: &config.VCenter{
						Host:       "vcenter2.example.com",
						Username:   "admin2",
						Password:   "secret2",
						Datacenter: "DC2",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster5",
							ResourcePool: "RP1",
						},
					},
				},
				{
					Name: "az2",
					VCenter: &config.VCenter{
						Host:       "vcenter2.example.com",
						Username:   "admin2",
						Password:   "secret2",
						Datacenter: "DC2",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster6",
							ResourcePool: "RP1",
						},
					},
				},
			},
		},
		AdditionalVMs: map[string][]string{
			"az1": {
				"additional-vm1",
			},
		},
	}
}

func TestNewVMSourceFromConfigBoshClient(t *testing.T) {
	c := baseSourceConfig()
	src := migrate.NewVMSourceFromConfig(c)
	require.IsType(t, migrate.NullBoshClient{}, src.BoshClient)

	c.Bosh = &config.Bosh{
		Host:         "192.168.1.2",
		ClientID:     "admin",
		ClientSecret: "secret",
	}
	src = migrate.NewVMSourceFromConfig(c)
	b := src.BoshClient.(*bosh.Client)
	require.Equal(t, "192.168.1.2", b.Environment)
	require.Equal(t, "admin", b.ClientID)
	require.Equal(t, "secret", b.ClientSecret)
}

func TestVMsToMigrateWithMultipleClusters(t *testing.T) {
	c := baseSourceConfig()
	src := migrate.NewVMSourceFromConfig(c)
	vms, err := src.VMsToMigrate(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 1)
	require.Equal(t, "az1", vms[0].AZ)
	require.Equal(t, "additional-vm1", vms[0].Name)
	require.Len(t, vms[0].Clusters, 2)
	require.Equal(t, "Cluster1", vms[0].Clusters[0])
	require.Equal(t, "Cluster2", vms[0].Clusters[1])
}

func TestVMsToMigrateInterleavesVMsByAZ(t *testing.T) {
	c := baseSourceConfig()
	c.AdditionalVMs = nil

	src := migrate.NewVMSourceFromConfig(c)

	// stub out a fake bosh client result
	b := &migratefakes.FakeBoshClient{}
	b.VMsAndStemcellsReturns([]bosh.VM{
		{
			Name: "vm1az1",
			AZ:   "az1",
		},
		{
			Name: "vm2az1",
			AZ:   "az1",
		},
		{
			Name: "vm3az1",
			AZ:   "az1",
		},
		{
			Name: "vm4az1",
			AZ:   "az1",
		},
		{
			Name: "vm1az2",
			AZ:   "az2",
		},
		{
			Name: "vm2az2",
			AZ:   "az2",
		},
		{
			Name: "vm3az2",
			AZ:   "az2",
		},
		{
			Name: "vm1az3",
			AZ:   "az3",
		},
		{
			Name: "vm2az3",
			AZ:   "az3",
		},
		{
			Name: "vm3az3",
			AZ:   "az3",
		},
		{
			Name: "vm4az3",
			AZ:   "az3",
		},
	}, nil)
	src.BoshClient = b

	vms, err := src.VMsToMigrate(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 11)

	vm := vms[0]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "vm1az1", vm.Name)

	vm = vms[1]
	require.Equal(t, "az2", vm.AZ)
	require.Equal(t, "vm1az2", vm.Name)

	vm = vms[2]
	require.Equal(t, "az3", vm.AZ)
	require.Equal(t, "vm1az3", vm.Name)

	vm = vms[3]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "vm2az1", vm.Name)

	vm = vms[4]
	require.Equal(t, "az2", vm.AZ)
	require.Equal(t, "vm2az2", vm.Name)

	vm = vms[5]
	require.Equal(t, "az3", vm.AZ)
	require.Equal(t, "vm2az3", vm.Name)

	vm = vms[6]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "vm3az1", vm.Name)

	vm = vms[7]
	require.Equal(t, "az2", vm.AZ)
	require.Equal(t, "vm3az2", vm.Name)

	vm = vms[8]
	require.Equal(t, "az3", vm.AZ)
	require.Equal(t, "vm3az3", vm.Name)

	vm = vms[9]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "vm4az1", vm.Name)

	vm = vms[10]
	require.Equal(t, "az3", vm.AZ)
	require.Equal(t, "vm4az3", vm.Name)
}

func TestConfigSourceBoshError(t *testing.T) {
	c := baseSourceConfig()
	src := migrate.NewVMSourceFromConfig(c)

	// stub out a fake bosh client error
	b := &migratefakes.FakeBoshClient{}
	b.VMsAndStemcellsReturns(nil, errors.New("could not connect to BOSH"))
	src.BoshClient = b

	_, err := src.VMsToMigrate(context.Background())
	require.Error(t, err)
}
