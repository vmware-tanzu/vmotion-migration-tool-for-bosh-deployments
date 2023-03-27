/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

//func TestMigrateFoundation(t *testing.T) {
//	boshClient := &migratefakes.FakeBoshClient{}
//	boshClient.VMsAndStemcellsReturns([]string{
//		"sc-1",
//		"vm1",
//		"vm2",
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
//	dstVCenter := &migratefakes.FakeVCenterClient{}
//
//	vmConverter := converter.New(
//		converter.NewEmptyMappedNetwork().Add("Net1", "Net2"),
//		converter.NewEmptyMappedDatastore().Add("DS1", "DS2"),
//		converter.NewEmptyMappedCompute().Add(converter.AZ{
//			Cluster: "Cluster1",
//		}, converter.AZ{
//			Cluster: "Cluster2",
//		}),
//		"DC2")
//
//	vmRelocator := &migratefakes.FakeVMRelocator{}
//
//	out := log.NewUpdatableStdout()
//	vmMigrator := migrate.NewVMMigrator(srcVCenter, dstVCenter, vmConverter, vmRelocator, out)
//
//	migrator := migrate.NewFoundationMigrator("DC1", boshClient, vmMigrator, out)
//	migrator.AdditionalVMs = []string{
//		"additional-vm1",
//	}
//	err := migrator.Migrate(context.Background())
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
//func TestBuildAZMappingFromConfig(t *testing.T) {
//
//}
