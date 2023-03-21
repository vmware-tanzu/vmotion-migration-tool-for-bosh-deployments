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

var explicitTests = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
}{
	{
		"Standard VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN": "tN"},
		},
	},
}

func TestExplicitConverter(t *testing.T) {
	net := converter.NewEmptyMappedNetwork().Add("sN", "tN")
	ds := converter.NewEmptyMappedDatastore().Add("sDS", "tDS")
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC",
		ResourcePool: "sRP",
	}, converter.AZMapping{
		Cluster:      "tC",
		ResourcePool: "tRP",
	})
	c := converter.New(net, ds, cm, "tDC")
	for _, tt := range explicitTests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			require.NoError(t, err)
			require.Equal(t, tt.out, spec)
		})
	}
}

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
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
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
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Multi-Disk VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
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
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Multi-Disk Multi-Datstore VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
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
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastores:   map[string]string{"sDS": "tDS", "sDS2": "tDS2"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Datastore",
		&vcenter.VM{
			Name:         "VM3",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
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
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "Resources",
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
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Network",
		&vcenter.VM{
			Name:         "VM3",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
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
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP-missing",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN"},
		},
		&vcenter.TargetSpec{},
		"could not find target compute for VM in source cluster sC, resource pool sRP-missing: ensure you add a corresponding compute mapping to the config file",
	},
	{
		"Stemcell (no network)",
		&vcenter.VM{
			Name:         "sc-someguid",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
		},
		&vcenter.TargetSpec{
			Name:         "sc-someguid",
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
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
	cm.Add(converter.AZMapping{
		Cluster:      "sC",
		ResourcePool: "sRP",
	}, converter.AZMapping{
		Cluster:      "tC",
		ResourcePool: "tRP",
	})
	cm.Add(converter.AZMapping{
		Cluster: "sC",
	}, converter.AZMapping{
		Cluster: "tC",
	})
	c := converter.New(net, ds, cm, "tDC")
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
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "Resources",
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
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "",
			Datastores:   map[string]string{"sDS": "tDS"},
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Resource Pool",
		&vcenter.VM{
			Name:         "VM12",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP-missing",
			Disks: []vcenter.Disk{
				{
					ID:        201,
					Datastore: "sDS",
				},
			},
			Networks: []string{"sN"},
		},
		&vcenter.TargetSpec{},
		"could not find target compute for VM in source cluster sC, resource pool sRP-missing: ensure you add a corresponding compute mapping to the config file",
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
	cm.Add(converter.AZMapping{
		Cluster: "sC",
	}, converter.AZMapping{
		Cluster: "tC",
	})
	c := converter.New(net, ds, cm, "tDC")
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
