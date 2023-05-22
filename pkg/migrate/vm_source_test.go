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
							ResourcePool: "RP5",
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

func TestConfigToBoshClient(t *testing.T) {
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

func TestConfigSourceAZToMultipleClusters(t *testing.T) {
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

func TestConfigSourceFromBosh(t *testing.T) {
	c := baseSourceConfig()
	c.AdditionalVMs = nil

	src := migrate.NewVMSourceFromConfig(c)

	// stub out a fake bosh client result
	b := &migratefakes.FakeBoshClient{}
	b.VMsAndStemcellsReturns([]bosh.VM{
		{
			Name: "bosh-vm1",
			AZ:   "az1",
		},
		{
			Name: "bosh-vm2",
			AZ:   "az1",
		},
		{
			Name: "bosh-vm3",
			AZ:   "az2",
		},
	}, nil)
	src.BoshClient = b

	vms, err := src.VMsToMigrate(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 3)

	vm := vms[0]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "bosh-vm1", vm.Name)
	require.Len(t, vms[0].Clusters, 2)
	require.Equal(t, "Cluster1", vm.Clusters[0])
	require.Equal(t, "Cluster2", vm.Clusters[1])

	vm = vms[1]
	require.Equal(t, "az1", vm.AZ)
	require.Equal(t, "bosh-vm2", vm.Name)
	require.Len(t, vm.Clusters, 2)
	require.Equal(t, "Cluster1", vm.Clusters[0])
	require.Equal(t, "Cluster2", vm.Clusters[1])

	vm = vms[2]
	require.Equal(t, "az2", vm.AZ)
	require.Equal(t, "bosh-vm3", vm.Name)
	require.Len(t, vm.Clusters, 1)
	require.Equal(t, "Cluster3", vm.Clusters[0])
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
