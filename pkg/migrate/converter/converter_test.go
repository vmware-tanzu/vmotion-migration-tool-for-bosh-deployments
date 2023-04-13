/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

var mappedTests = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
	err  string
}{
	{
		"Standard VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"VM in sub-folder",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm/sub1/sub2",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Folder:       "/DC/vm/sub1/sub2",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Multi-Disk VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
				{
					ID:        202,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Multi-Disk Multi-Datstore VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
				{
					ID:        202,
					Datastore: "sDS2",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS", "sDS2": "tDS2"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Datastore",
		&vcenter.VM{
			Name:         "VM3",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS-missing",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{},
		"could not find a target datastore for VM VM3 with source datastore sDS-missing: ensure you add a corresponding datastore mapping to the config file",
	},
	{
		"Default Resource Pool VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "Resources",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Network",
		&vcenter.VM{
			Name:         "VM3",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN-missing"},
		},
		&vcenter.TargetSpec{},
		"could not find a target network for VM VM3 attached to network sN-missing: ensure you add a corresponding network mapping to the config file",
	},
	{
		"Unmapped Resource Pool",
		&vcenter.VM{
			Name:         "VM12",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP-missing",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN"},
		},
		&vcenter.TargetSpec{},
		"could not find target compute, source: (az: 'az1', datacenter: 'DC', cluster: 'sC', resource pool: 'sRP-missing') ensure you add a corresponding compute mapping to the config file",
	},
	{
		"Stemcell (no network)",
		&vcenter.VM{
			Name:         "sc-someguid",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
		},
		&vcenter.TargetSpec{
			Name:         "sc-someguid",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{},
		}, "",
	},
}

func TestMappedConverter(t *testing.T) {
	net := converter.NewMappedNetwork(map[string]string{
		"sN":  "tN",
		"sN2": "tN2",
		"sN3": "tN3",
	})
	ds := converter.NewMappedDatastore(map[string]string{
		"sDS":  "tDS",
		"sDS2": "tDS2",
	})
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "DC",
		Name:         "az1",
		Cluster:      "sC",
		ResourcePool: "sRP",
	}, converter.AZ{
		Datacenter:   "DC",
		Name:         "az1",
		Cluster:      "tC",
		ResourcePool: "tRP",
	})
	cm.Add(converter.AZ{
		Datacenter: "DC",
		Name:       "az1",
		Cluster:    "sC",
	}, converter.AZ{
		Datacenter: "DC",
		Name:       "az1",
		Cluster:    "tC",
	})
	c := converter.New(net, ds, cm)
	for _, tt := range mappedTests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			if tt.err != "" {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.out, spec)
			}
		})
	}
}

var mappedTestsNoRP = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
	err  string
}{
	{
		"Default Resource Pool VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "Resources",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "DC",
			Cluster:      "tC",
			ResourcePool: "",
			Folder:       "/DC/vm",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Resource Pool",
		&vcenter.VM{
			Name:         "VM12",
			AZ:           "az1",
			Datacenter:   "DC",
			Cluster:      "sC",
			ResourcePool: "sRP-missing",
			Folder:       "/DC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN"},
		},
		&vcenter.TargetSpec{},
		"could not find target compute, source: (az: 'az1', datacenter: 'DC', cluster: 'sC', resource pool: 'sRP-missing') ensure you add a corresponding compute mapping to the config file",
	},
}

func TestMappedConverterNoResourcePools(t *testing.T) {
	net := converter.NewMappedNetwork(map[string]string{
		"sN":  "tN",
		"sN2": "tN2",
		"sN3": "tN3",
	})
	ds := converter.NewMappedDatastore(map[string]string{
		"sDS": "tDS",
	})
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter: "DC",
		Name:       "az1",
		Cluster:    "sC",
	}, converter.AZ{
		Datacenter: "DC",
		Name:       "az1",
		Cluster:    "tC",
	})
	c := converter.New(net, ds, cm)
	for _, tt := range mappedTestsNoRP {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			if tt.err != "" {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.out, spec)
			}
		})
	}
}

var mappedTestsCrossDC = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
	err  string
}{
	{
		"VM in different datacenter",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "sDC",
			Cluster:      "CL",
			ResourcePool: "RP",
			Folder:       "/sDC/vm",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "DS",
				},
			},
			Networks: []string{"NET"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "tDC",
			Cluster:      "CL",
			ResourcePool: "RP",
			Folder:       "/tDC/vm",
			Datastores:   map[string]string{"DS": "DS"},
			Networks:     map[string]string{"NET": "NET"},
		}, "",
	},
	{
		"VM in sub-folder in different datacenter",
		&vcenter.VM{
			Name:         "virtualMachine42",
			AZ:           "az1",
			Datacenter:   "sDC",
			Cluster:      "CL",
			ResourcePool: "RP",
			Folder:       "/sDC/vm/sub1/sub2",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "DS",
				},
			},
			Networks: []string{"NET"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "tDC",
			Cluster:      "CL",
			ResourcePool: "RP",
			Folder:       "/tDC/vm/sub1/sub2",
			Datastores:   map[string]string{"DS": "DS"},
			Networks:     map[string]string{"NET": "NET"},
		}, "",
	},
}

func TestMappedConverterCrossDatacenter(t *testing.T) {
	net := converter.NewMappedNetwork(map[string]string{
		"NET": "NET",
	})
	ds := converter.NewMappedDatastore(map[string]string{
		"DS": "DS",
	})
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "sDC",
		Name:         "az1",
		Cluster:      "CL",
		ResourcePool: "RP",
	}, converter.AZ{
		Datacenter:   "tDC",
		Name:         "az1",
		Cluster:      "CL",
		ResourcePool: "RP",
	})
	c := converter.New(net, ds, cm)
	for _, tt := range mappedTestsCrossDC {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			if tt.err != "" {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.out, spec)
			}
		})
	}
}
