/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"testing"
)

func TestNewFoundationMigratorFromConfig(t *testing.T) {
	c := config.Config{
		Bosh: &config.Bosh{
			Host:         "192.168.1.2",
			ClientID:     "admin",
			ClientSecret: "secret",
		},
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
						Host:       "vcenter.example.com",
						Username:   "admin",
						Password:   "secret",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster1",
							ResourcePool: "",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name: "az1",
					VCenter: &config.VCenter{
						Host:       "vcenter.example.com",
						Username:   "admin",
						Password:   "secret",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster2",
							ResourcePool: "",
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
	_, err := migrate.NewFoundationMigratorFromConfig(context.Background(), c)
	require.NoError(t, err)

	// TODO write assertions to validate the config was read/used properly to construct object graph
}

//
//func TestMigrateFoundation(t *testing.T) {
//	boshClient := &migratefakes.FakeBoshClient{}
//	boshClient.VMsAndStemcellsReturns([]bosh.VM{
//		{
//			Name: "sc-1",
//			AZ: "az1",
//		},
//		{
//			Name: "vm1",
//			AZ: "az1",
//		},
//		{
//			Name: "vm2",
//			AZ: "az1",
//		},
//	}, nil)
//	srcVCenter := &migratefakes.FakeVCenterClient{}
//	srcVCenter.FindVMReturnsOnCall(0, &vcenter.VM{
//		Name:         "sc-1",
//		AZ:           "az1",
//		Datacenter:   "DC1",
//		Cluster:      "Cluster1",
//		ResourcePool: "RP1",
//		Disks: []vcenter.Disk{
//			{
//				ID:        201,
//				Datastore: "DS1",
//			},
//		},
//	}, nil)
//	srcVCenter.FindVMReturnsOnCall(1, &vcenter.VM{
//		Name:         "vm1",
//		AZ:           "az1",
//		Datacenter:   "DC",
//		Cluster:      "Cluster1",
//		ResourcePool: "RP1",
//		Disks: []vcenter.Disk{
//			{
//				ID:        201,
//				Datastore: "DS1",
//			},
//		},
//		Networks: []string{"Net1"},
//	}, nil)
//	srcVCenter.FindVMReturnsOnCall(2, &vcenter.VM{
//		Name:         "vm2",
//		AZ:           "az1",
//		Datacenter:   "DC1",
//		Cluster:      "Cluster1",
//		ResourcePool: "RP1",
//		Disks: []vcenter.Disk{
//			{
//				ID:        201,
//				Datastore: "DS1",
//			},
//		},
//		Networks: []string{"Net1"},
//	}, nil)
//	srcVCenter.FindVMReturnsOnCall(3, &vcenter.VM{
//		Name:         "additional-vm1",
//		AZ:           "az1",
//		Datacenter:   "DC1",
//		Cluster:      "Cluster1",
//		ResourcePool: "RP1",
//		Disks: []vcenter.Disk{
//			{
//				ID:        201,
//				Datastore: "DS1",
//			},
//		},
//		Networks: []string{"Net1"},
//	}, nil)
//
//	sMap := map[string]*vcenter.Client{
//		"az1": srcVCenter,
//	}
//	clientPool := vcenter.NewPoolWithExternalClients(map["az1"])
//	vmRelocator := &migratefakes.FakeVMRelocator{}
//	//c := config.Config{
//	//	Bosh: &config.Bosh{
//	//		Host:         "192.168.1.2",
//	//		ClientID:     "admin",
//	//		ClientSecret: "secret",
//	//	},
//	//	DryRun:         false,
//	//	WorkerPoolSize: 1,
//	//	NetworkMap: map[string]string{
//	//		"Net1": "Net2",
//	//	},
//	//	DatastoreMap:   map[string]string{
//	//		"DS1": "DS2",
//	//	},
//	//	Compute: config.Compute{
//	//		Source: nil,
//	//		Target: nil,
//	//	},
//	//	AdditionalVMs: map[string][]string{
//	//		"az1": []string{
//	//			"additional-vm1",
//	//		},
//	//	},
//	//}
//	migrator, err := migrate.NewFoundationMigrator()
//	require.NoError(t, err)
//
//	err = migrator.Migrate(context.Background())
//	require.NoError(t, err)
//
//	_, srcTemplate, targetSpec := vmRelocator.RelocateVMArgsForCall(0)
//	require.Equal(t, "sc-1", srcTemplate.Name)
//	require.Equal(t, "sc-1", targetSpec.Name)
//	require.Equal(t, "DC2", targetSpec.Datacenter)
//	require.Equal(t, "RP2", targetSpec.ResourcePool)
//	require.Equal(t, "Cluster2", targetSpec.Cluster)
//	require.Equal(t, map[string]string{"DS1": "DS2"}, targetSpec.Datastores)
//	require.Equal(t, map[string]string{}, targetSpec.Networks)
//
//	_, srcVM, targetSpec := vmRelocator.RelocateVMArgsForCall(1)
//	require.Equal(t, "vm1", srcVM.Name)
//	require.Equal(t, "vm1", targetSpec.Name)
//	require.Equal(t, "DC2", targetSpec.Datacenter)
//	require.Equal(t, "RP2", targetSpec.ResourcePool)
//	require.Equal(t, "Cluster2", targetSpec.Cluster)
//	require.Equal(t, map[string]string{"DS1": "DS2"}, targetSpec.Datastores)
//	require.Equal(t, map[string]string{"Net1": "Net2"}, targetSpec.Networks)
//
//	_, srcVM, targetSpec = vmRelocator.RelocateVMArgsForCall(2)
//	require.Equal(t, "vm2", srcVM.Name)
//	require.Equal(t, "vm2", targetSpec.Name)
//	require.Equal(t, "DC2", targetSpec.Datacenter)
//	require.Equal(t, "RP2", targetSpec.ResourcePool)
//	require.Equal(t, "Cluster2", targetSpec.Cluster)
//	require.Equal(t, map[string]string{"DS1": "DS2"}, targetSpec.Datastores)
//	require.Equal(t, map[string]string{"Net1": "Net2"}, targetSpec.Networks)
//
//	_, srcVM, targetSpec = vmRelocator.RelocateVMArgsForCall(3)
//	require.Equal(t, "additional-vm1", srcVM.Name)
//	require.Equal(t, "additional-vm1", targetSpec.Name)
//	require.Equal(t, "DC2", targetSpec.Datacenter)
//	require.Equal(t, "RP2", targetSpec.ResourcePool)
//	require.Equal(t, "Cluster2", targetSpec.Cluster)
//	require.Equal(t, map[string]string{"DS1": "DS2"}, targetSpec.Datastores)
//	require.Equal(t, map[string]string{"Net1": "Net2"}, targetSpec.Networks)
//}
//
